package service

import (
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func GetPersonList(pageNum, pageSize int, name, idCard string, personID string) ([]model.Person, int64, error) {
	tx := dao.DB.Model(&model.Person{}).Preload("Phones").Preload("Emails").Preload("BankCards")
	if personID != "" {
		tx = tx.Where("id = ?", personID)
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if idCard != "" {
		tx = tx.Where("id_card LIKE ?", "%"+idCard+"%")
	}

	var total int64
	tx.Count(&total)

	var list []model.Person
	offset := (pageNum - 1) * pageSize
	if err := tx.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func GetPersonByID(id uint) (*model.Person, error) {
	var person model.Person
	if err := dao.DB.Preload("Phones").Preload("Emails").Preload("BankCards").First(&person, id).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func CreatePerson(p *model.Person) error {
	if p.IDCard != "" {
		var count int64
		dao.DB.Model(&model.Person{}).Where("id_card = ?", p.IDCard).Count(&count)
		if count > 0 {
			return errors.New("身份证号已存在")
		}
	}
	return dao.DB.Create(p).Error
}

func UpdatePerson(id uint, p *model.Person) error {
	var existing model.Person
	if err := dao.DB.First(&existing, id).Error; err != nil {
		return errors.New("人员不存在")
	}

	if p.IDCard != "" && p.IDCard != existing.IDCard {
		var count int64
		dao.DB.Model(&model.Person{}).Where("id_card = ? AND id != ?", p.IDCard, id).Count(&count)
		if count > 0 {
			return errors.New("身份证号已存在")
		}
	}

	updates := map[string]interface{}{
		"name":             p.Name,
		"id_card":          p.IDCard,
		"gender":           p.Gender,
		"birthday":         p.Birthday,
		"nation":           p.Nation,
		"native_place":     p.NativePlace,
		"address":          p.Address,
		"political_status": p.PoliticalStatus,
		"marital_status":   p.MaritalStatus,
		"alias":            p.Alias,
	}
	return dao.DB.Model(&existing).Updates(updates).Error
}

func DeletePerson(id uint) error {
	var person model.Person
	if err := dao.DB.First(&person, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Delete(&person).Error; err != nil {
			return err
		}
		tx.Where("person_id = ?", id).Delete(&model.PersonPhone{})
		tx.Where("person_id = ?", id).Delete(&model.PersonEmail{})
		tx.Where("person_id = ?", id).Delete(&model.PersonBankCard{})
		tx.Where("person_id = ?", id).Delete(&model.PositionEvent{})
		tx.Where("person_id = ?", id).Delete(&model.AttendanceEvent{})
		tx.Where("person_id = ?", id).Delete(&model.AnnualLeaveAccountEvent{})
		tx.Where("person_id = ?", id).Delete(&model.SalaryEvent{})
		tx.Where("target_type = ? AND target_id = ?", "person", id).Delete(&model.FileRelation{})
		return nil
	})
}

func RestorePerson(id uint) error {
	var person model.Person
	if err := dao.DB.Unscoped().First(&person, id).Error; err != nil {
		return err
	}
	if person.IDCard != "" {
		var count int64
		dao.DB.Model(&model.Person{}).Where("id_card = ? AND id != ?", person.IDCard, id).Count(&count)
		if count > 0 {
			return errors.New("身份证号已被占用，无法恢复")
		}
	}
	return utils.WithTransaction(dao.DB, func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&person).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&model.PersonPhone{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PersonEmail{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PersonBankCard{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PositionEvent{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.AttendanceEvent{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.AnnualLeaveAccountEvent{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.SalaryEvent{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.FileRelation{}).Where("target_type = ? AND target_id = ?", "person", id).Update("deleted_at", nil)
		return nil
	})
}

func GetDeletedPersons(pageNum, pageSize int) ([]model.Person, int64, error) {
	var list []model.Person
	var total int64
	tx := dao.DB.Unscoped().Model(&model.Person{}).Where("deleted_at IS NOT NULL")
	tx.Count(&total)
	offset := (pageNum - 1) * pageSize
	tx.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&list)
	return list, total, nil
}

func GetAllPersons() ([]model.Person, error) {
	var list []model.Person
	if err := dao.DB.Order("name").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func AddPersonPhone(personID uint, phone, phoneType string) error {
	if phoneType == "" {
		phoneType = "mobile"
	}
	p := model.PersonPhone{PersonID: personID, Phone: phone, PhoneType: phoneType}
	return dao.DB.Create(&p).Error
}

func UpdatePersonPhone(id uint, phone, phoneType string) error {
	updates := map[string]interface{}{"phone": phone}
	if phoneType != "" {
		updates["phone_type"] = phoneType
	}
	return dao.DB.Model(&model.PersonPhone{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonPhone(id uint) error {
	return dao.DB.Delete(&model.PersonPhone{}, id).Error
}

func AddPersonEmail(personID uint, email, emailType string) error {
	if emailType == "" {
		emailType = "personal"
	}
	e := model.PersonEmail{PersonID: personID, Email: email, EmailType: emailType}
	return dao.DB.Create(&e).Error
}

func UpdatePersonEmail(id uint, email, emailType string) error {
	updates := map[string]interface{}{"email": email}
	if emailType != "" {
		updates["email_type"] = emailType
	}
	return dao.DB.Model(&model.PersonEmail{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonEmail(id uint) error {
	return dao.DB.Delete(&model.PersonEmail{}, id).Error
}

func AddPersonBankCard(personID uint, bankName, accountNumber, accountHolder string) error {
	c := model.PersonBankCard{PersonID: personID, BankName: bankName, AccountNumber: accountNumber, AccountHolder: accountHolder}
	return dao.DB.Create(&c).Error
}

func UpdatePersonBankCard(id uint, bankName, accountNumber, accountHolder string) error {
	updates := map[string]interface{}{"bank_name": bankName, "account_number": accountNumber, "account_holder": accountHolder}
	return dao.DB.Model(&model.PersonBankCard{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonBankCard(id uint) error {
	return dao.DB.Delete(&model.PersonBankCard{}, id).Error
}
