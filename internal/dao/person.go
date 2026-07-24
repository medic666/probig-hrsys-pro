package dao

import (
	"probig/internal/models"
	"gorm.io/gorm"
)

func GetPersonList(page, pageSize int, keyword string) ([]models.Person, int64, error) {
	var list []models.Person
	var total int64
	q := DB().Model(&models.Person{})
	if keyword != "" {
		q = q.Where("name LIKE ? OR alias LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Preload("Phones").Preload("Emails").Preload("BankCards").
		Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	for i := range list {
		list[i].IDCardPlain = decrypt(list[i].IDCard)
		for j := range list[i].Phones {
			list[i].Phones[j].PhonePlain = decrypt(list[i].Phones[j].Phone)
		}
		for j := range list[i].Emails {
			list[i].Emails[j].EmailPlain = decrypt(list[i].Emails[j].Email)
		}
		for j := range list[i].BankCards {
			list[i].BankCards[j].BankCardPlain = decrypt(list[i].BankCards[j].BankCard)
		}
	}
	return list, total, nil
}

func GetPersonByID(id uint) (*models.Person, error) {
	var p models.Person
	if err := DB().Preload("Phones").Preload("Emails").Preload("BankCards").
		First(&p, id).Error; err != nil {
		return nil, err
	}
	p.IDCardPlain = decrypt(p.IDCard)
	for i := range p.Phones {
		p.Phones[i].PhonePlain = decrypt(p.Phones[i].Phone)
	}
	for i := range p.Emails {
		p.Emails[i].EmailPlain = decrypt(p.Emails[i].Email)
	}
	for i := range p.BankCards {
		p.BankCards[i].BankCardPlain = decrypt(p.BankCards[i].BankCard)
	}
	return &p, nil
}

func GetPersonByIDCard(idCard string) (*models.Person, error) {
	var p models.Person
	if err := DB().Where("id_card = ?", encrypt(idCard)).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func CreatePerson(p *models.Person) error {
	p.IDCard = encrypt(p.IDCardPlain)
	for i := range p.Phones {
		p.Phones[i].Phone = encrypt(p.Phones[i].PhonePlain)
	}
	for i := range p.Emails {
		p.Emails[i].Email = encrypt(p.Emails[i].EmailPlain)
	}
	for i := range p.BankCards {
		p.BankCards[i].BankCard = encrypt(p.BankCards[i].BankCardPlain)
	}
	return DB().Create(p).Error
}

func UpdatePerson(p *models.Person) error {
	p.IDCard = encrypt(p.IDCardPlain)
	for i := range p.Phones {
		p.Phones[i].Phone = encrypt(p.Phones[i].PhonePlain)
	}
	for i := range p.Emails {
		p.Emails[i].Email = encrypt(p.Emails[i].EmailPlain)
	}
	for i := range p.BankCards {
		p.BankCards[i].BankCard = encrypt(p.BankCards[i].BankCardPlain)
	}
	tx := DB().Begin()
	if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(p).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func SoftDeletePerson(id uint) error {
	tx := DB().Begin()
	if err := tx.Where("person_id = ?", id).Delete(&models.PersonPhone{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("person_id = ?", id).Delete(&models.PersonEmail{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("person_id = ?", id).Delete(&models.PersonBankCard{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("person_id = ?", id).Delete(&models.PositionEvent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("person_id = ?", id).Delete(&models.AttendanceEvent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("person_id = ?", id).Delete(&models.SalaryEvent{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("target_type = ? AND target_id = ?", "person", id).Delete(&models.FileRelation{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.Person{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func RestorePerson(id uint) error {
	tx := DB().Begin()
	if err := tx.Unscoped().Model(&models.Person{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.PersonPhone{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.PersonEmail{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.PersonBankCard{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.PositionEvent{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.AttendanceEvent{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.SalaryEvent{}).Where("person_id = ?", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Model(&models.FileRelation{}).Where("target_type = ? AND target_id = ?", "person", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
