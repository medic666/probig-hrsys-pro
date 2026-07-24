package services

import (
	"errors"
	"time"

	"probig/database"
	"probig/models"

	"gorm.io/gorm"
)

func ListPersons(query string, offset, limit int) ([]models.Person, int64, error) {
	var persons []models.Person
	var total int64
	db := database.DB.Model(&models.Person{})
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ? OR alias LIKE ?", like, like)
	}
	db.Count(&total)
	err := db.Preload("Phones").Preload("Emails").Preload("BankCards").
		Offset(offset).Limit(limit).Order("id DESC").Find(&persons).Error
	return persons, total, err
}

func GetPerson(id uint) (*models.Person, error) {
	var person models.Person
	err := database.DB.Preload("Phones").Preload("Emails").Preload("BankCards").First(&person, id).Error
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func CreatePerson(person *models.Person) error {
	if person.IdCard != "" {
		var existing models.Person
		if err := database.DB.Where("id_card = ?", person.IdCard).First(&existing).Error; err == nil {
			return errors.New("身份证号已存在")
		}
	}
	return database.DB.Create(person).Error
}

func UpdatePerson(id uint, updates map[string]interface{}) error {
	if card, ok := updates["id_card"].(string); ok && card != "" {
		var existing models.Person
		if err := database.DB.Where("id_card = ? AND id != ?", card, id).First(&existing).Error; err == nil {
			return errors.New("身份证号已存在")
		}
	}
	return database.DB.Model(&models.Person{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePerson(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Person{}, id).Error; err != nil {
			return err
		}
		tx.Where("person_id = ?", id).Delete(&models.PersonPhone{})
		tx.Where("person_id = ?", id).Delete(&models.PersonEmail{})
		tx.Where("person_id = ?", id).Delete(&models.PersonBankCard{})
		tx.Where("person_id = ?", id).Delete(&models.PositionEvent{})
		tx.Where("person_id = ?", id).Delete(&models.AttendanceEvent{})
		tx.Where("person_id = ?", id).Delete(&models.SalaryEvent{})
		tx.Where("person_id = ?", id).Delete(&models.PositionSnapshot{})
		tx.Where("person_id = ?", id).Delete(&models.AttendanceSummary{})
		tx.Where("person_id = ?", id).Delete(&models.SalarySummary{})
		tx.Where("target_type = 'person' AND target_id = ?", id).Delete(&models.FileRelation{})
		return nil
	})
}

func RestorePerson(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&models.Person{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Where("person_id = ?", id).Model(&models.PersonPhone{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.PersonEmail{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.PersonBankCard{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.PositionEvent{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.AttendanceEvent{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.SalaryEvent{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.PositionSnapshot{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.AttendanceSummary{}).Update("deleted_at", nil)
		tx.Unscoped().Where("person_id = ?", id).Model(&models.SalarySummary{}).Update("deleted_at", nil)
		tx.Unscoped().Where("target_type = 'person' AND target_id = ?", id).Model(&models.FileRelation{}).Update("deleted_at", nil)
		return nil
	})
}

func ListDeletedPersons(query string, offset, limit int) ([]models.Person, int64, error) {
	var persons []models.Person
	var total int64
	db := database.DB.Unscoped().Model(&models.Person{}).Where("deleted_at IS NOT NULL")
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("name LIKE ? OR alias LIKE ?", like, like)
	}
	db.Count(&total)
	err := db.Offset(offset).Limit(limit).Order("deleted_at DESC").Find(&persons).Error
	return persons, total, err
}

func AddPersonPhone(personID uint, phone *models.PersonPhone) error {
	phone.PersonID = personID
	phone.CreatedAt = time.Now()
	phone.UpdatedAt = time.Now()
	return database.DB.Create(phone).Error
}

func UpdatePersonPhone(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.PersonPhone{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonPhone(id uint) error {
	return database.DB.Delete(&models.PersonPhone{}, id).Error
}

func AddPersonEmail(personID uint, email *models.PersonEmail) error {
	email.PersonID = personID
	email.CreatedAt = time.Now()
	email.UpdatedAt = time.Now()
	return database.DB.Create(email).Error
}

func UpdatePersonEmail(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.PersonEmail{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonEmail(id uint) error {
	return database.DB.Delete(&models.PersonEmail{}, id).Error
}

func AddPersonBankCard(personID uint, card *models.PersonBankCard) error {
	card.PersonID = personID
	card.CreatedAt = time.Now()
	card.UpdatedAt = time.Now()
	return database.DB.Create(card).Error
}

func UpdatePersonBankCard(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&models.PersonBankCard{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonBankCard(id uint) error {
	return database.DB.Delete(&models.PersonBankCard{}, id).Error
}
