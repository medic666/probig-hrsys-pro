package person

import (
	"probig/internal/pkg/encrypt"

	"gorm.io/gorm"
)

type DAO struct {
	db         *gorm.DB
	encryptKey string
}

func NewDAO(db *gorm.DB, encryptKey string) *DAO {
	return &DAO{db: db, encryptKey: encryptKey}
}

func (d *DAO) encrypt(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	return encrypt.Encrypt(s, d.encryptKey)
}

func (d *DAO) decrypt(s string) (string, error) {
	if s == "" {
		return s, nil
	}
	return encrypt.Decrypt(s, d.encryptKey)
}

func (d *DAO) encryptPerson(p *Person) error {
	var err error
	p.IDCard, err = d.encrypt(p.IDCard)
	return err
}

func (d *DAO) decryptPerson(p *Person) error {
	var err error
	p.IDCard, err = d.decrypt(p.IDCard)
	return err
}

func (d *DAO) encryptPhone(p *PersonPhone) error {
	var err error
	p.Phone, err = d.encrypt(p.Phone)
	return err
}

func (d *DAO) decryptPhone(p *PersonPhone) error {
	var err error
	p.Phone, err = d.decrypt(p.Phone)
	return err
}

func (d *DAO) encryptEmail(p *PersonEmail) error {
	var err error
	p.Email, err = d.encrypt(p.Email)
	return err
}

func (d *DAO) decryptEmail(p *PersonEmail) error {
	var err error
	p.Email, err = d.decrypt(p.Email)
	return err
}

func (d *DAO) encryptBankCard(p *PersonBankCard) error {
	var err error
	p.BankCard, err = d.encrypt(p.BankCard)
	return err
}

func (d *DAO) decryptBankCard(p *PersonBankCard) error {
	var err error
	p.BankCard, err = d.decrypt(p.BankCard)
	return err
}

func (d *DAO) decryptPersonFull(p *Person) error {
	if err := d.decryptPerson(p); err != nil {
		return err
	}
	for i := range p.Phones {
		if err := d.decryptPhone(&p.Phones[i]); err != nil {
			return err
		}
	}
	for i := range p.Emails {
		if err := d.decryptEmail(&p.Emails[i]); err != nil {
			return err
		}
	}
	for i := range p.BankCards {
		if err := d.decryptBankCard(&p.BankCards[i]); err != nil {
			return err
		}
	}
	return nil
}

func (d *DAO) CreatePerson(person *Person) error {
	if err := d.encryptPerson(person); err != nil {
		return err
	}
	for i := range person.Phones {
		if err := d.encryptPhone(&person.Phones[i]); err != nil {
			return err
		}
	}
	for i := range person.Emails {
		if err := d.encryptEmail(&person.Emails[i]); err != nil {
			return err
		}
	}
	for i := range person.BankCards {
		if err := d.encryptBankCard(&person.BankCards[i]); err != nil {
			return err
		}
	}
	if err := d.db.Create(person).Error; err != nil {
		return err
	}
	return d.decryptPersonFull(person)
}

func (d *DAO) GetPersonByID(id uint) (*Person, error) {
	var person Person
	err := d.db.
		Preload("Phones", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Emails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("BankCards", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		First(&person, id).Error
	if err != nil {
		return nil, err
	}
	if err := d.decryptPersonFull(&person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (d *DAO) ListPersons(page, pageSize int, name, idCard string) ([]Person, int64, error) {
	var persons []Person
	var total int64
	query := d.db.Model(&Person{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if idCard != "" {
		encryptedIDCard, err := d.encrypt(idCard)
		if err != nil {
			return nil, 0, err
		}
		query = query.Where("id_card = ?", encryptedIDCard)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.
		Preload("Phones", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Emails", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("BankCards", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Offset(offset).Limit(pageSize).Order("id DESC").
		Find(&persons).Error
	if err != nil {
		return nil, 0, err
	}
	for i := range persons {
		if err := d.decryptPersonFull(&persons[i]); err != nil {
			return nil, 0, err
		}
	}
	return persons, total, nil
}

func (d *DAO) UpdatePerson(person *Person) error {
	if err := d.encryptPerson(person); err != nil {
		return err
	}
	for i := range person.Phones {
		if err := d.encryptPhone(&person.Phones[i]); err != nil {
			return err
		}
	}
	for i := range person.Emails {
		if err := d.encryptEmail(&person.Emails[i]); err != nil {
			return err
		}
	}
	for i := range person.BankCards {
		if err := d.encryptBankCard(&person.BankCards[i]); err != nil {
			return err
		}
	}
	if err := d.db.Session(&gorm.Session{FullSaveAssociations: true}).Model(person).Select("*").Updates(person).Error; err != nil {
		return err
	}
	return d.decryptPersonFull(person)
}

func (d *DAO) DeletePerson(id uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Person{}, id).Error; err != nil {
			return err
		}
		if err := tx.Where("person_id = ?", id).Delete(&PersonPhone{}).Error; err != nil {
			return err
		}
		if err := tx.Where("person_id = ?", id).Delete(&PersonEmail{}).Error; err != nil {
			return err
		}
		if err := tx.Where("person_id = ?", id).Delete(&PersonBankCard{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DAO) RestorePerson(id uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&Person{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&PersonPhone{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&PersonEmail{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&PersonBankCard{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
}

func (d *DAO) GetPersonByIDCard(idCard string) (*Person, error) {
	encryptedIDCard, err := d.encrypt(idCard)
	if err != nil {
		return nil, err
	}
	var person Person
	err = d.db.Where("id_card = ?", encryptedIDCard).First(&person).Error
	if err != nil {
		return nil, err
	}
	if err := d.decryptPerson(&person); err != nil {
		return nil, err
	}
	return &person, nil
}

func (d *DAO) GetAllPersonsSimple() ([]Person, error) {
	var persons []Person
	err := d.db.Select("id", "name").Find(&persons).Error
	return persons, err
}

func (d *DAO) CreatePhone(phone *PersonPhone) error {
	if err := d.encryptPhone(phone); err != nil {
		return err
	}
	if err := d.db.Create(phone).Error; err != nil {
		return err
	}
	return d.decryptPhone(phone)
}

func (d *DAO) UpdatePhone(phone *PersonPhone) error {
	if err := d.encryptPhone(phone); err != nil {
		return err
	}
	return d.db.Save(phone).Error
}

func (d *DAO) DeletePhone(id uint) error {
	return d.db.Delete(&PersonPhone{}, id).Error
}

func (d *DAO) CreateEmail(email *PersonEmail) error {
	if err := d.encryptEmail(email); err != nil {
		return err
	}
	if err := d.db.Create(email).Error; err != nil {
		return err
	}
	return d.decryptEmail(email)
}

func (d *DAO) UpdateEmail(email *PersonEmail) error {
	if err := d.encryptEmail(email); err != nil {
		return err
	}
	return d.db.Save(email).Error
}

func (d *DAO) DeleteEmail(id uint) error {
	return d.db.Delete(&PersonEmail{}, id).Error
}

func (d *DAO) CreateBankCard(card *PersonBankCard) error {
	if err := d.encryptBankCard(card); err != nil {
		return err
	}
	if err := d.db.Create(card).Error; err != nil {
		return err
	}
	return d.decryptBankCard(card)
}

func (d *DAO) UpdateBankCard(card *PersonBankCard) error {
	if err := d.encryptBankCard(card); err != nil {
		return err
	}
	return d.db.Save(card).Error
}

func (d *DAO) DeleteBankCard(id uint) error {
	return d.db.Delete(&PersonBankCard{}, id).Error
}
