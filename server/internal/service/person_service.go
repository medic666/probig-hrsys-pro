package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

func GetPersonList(pageNum, pageSize int, name, idCard string, personID string) ([]model.Person, int64, error) {
	tx := dao.DB.Model(&model.Person{}).Preload("Phones").Preload("Emails").Preload("BankCards").Preload("EmergencyContacts")
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
	if err := dao.DB.Preload("Phones").Preload("Emails").Preload("BankCards").Preload("EmergencyContacts").First(&person, id).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func CreatePerson(ctx context.Context, p *model.Person) error {
	if p.IDCard != "" {
		var count int64
		dao.DB.Model(&model.Person{}).Where("id_card = ?", p.IDCard).Count(&count)
		if count > 0 {
			return errors.New("身份证号已存在")
		}
	}
	return dao.DBFromContext(ctx).Create(p).Error
}

func UpdatePerson(ctx context.Context, id uint, p *model.Person) error {
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
	return dao.DBFromContext(ctx).Model(&existing).Updates(updates).Error
}

func DeletePerson(ctx context.Context, id uint) error {
	var person model.Person
	if err := dao.DB.First(&person, id).Error; err != nil {
		return err
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Delete(&person).Error; err != nil {
			return err
		}
		tx.Where("person_id = ?", id).Delete(&model.PersonPhone{})
		tx.Where("person_id = ?", id).Delete(&model.PersonEmail{})
		tx.Where("person_id = ?", id).Delete(&model.PersonBankCard{})
		tx.Where("person_id = ?", id).Delete(&model.PersonEmergencyContact{})
		tx.Where("person_id = ?", id).Delete(&model.PositionEvent{})
		tx.Where("person_id = ?", id).Delete(&model.AttendanceDaily{})
		tx.Where("person_id = ?", id).Delete(&model.AnnualLeaveAccountEvent{})
		tx.Where("person_id = ?", id).Delete(&model.SalaryEvent{})
		tx.Where("target_type = ? AND target_id = ?", "person", id).Delete(&model.FileRelation{})
		return nil
	})
}

func RestorePerson(ctx context.Context, id uint) error {
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
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&person).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&model.PersonPhone{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PersonEmail{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PersonBankCard{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PersonEmergencyContact{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.PositionEvent{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&model.AttendanceDaily{}).Where("person_id = ?", id).Update("deleted_at", nil)
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

// PersonCard 人员卡片：基本信息 + 当前职务快照（公司/部门/职位）
type PersonCard struct {
	ID          uint   `json:"id"`
	PersonID    uint   `json:"person_id"`
	Name        string `json:"name"`
	CompanyID   uint   `json:"company_id"`
	CompanyName string `json:"company_name"`
	Department  string `json:"department"`
	Position    string `json:"position"`
}

// GetPersonCards 人员卡片列表：以当前职务快照（9999-12-31 结束段）关联公司/部门/职位
func GetPersonCards() ([]PersonCard, error) {
	var cards []PersonCard
	err := dao.DB.Table("persons").
		Select(`persons.id, persons.id AS person_id, persons.name,
			s.company_id, c.name AS company_name, s.department, s.position`).
		Joins(`LEFT JOIN position_snapshots s ON s.person_id = persons.id
			AND s.effective_end_date = ? AND s.is_active = 1`, realFarFuture).
		Joins("LEFT JOIN companies c ON c.id = s.company_id").
		Where("persons.deleted_at IS NULL").
		Order("persons.name").
		Scan(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func AddPersonPhone(ctx context.Context, personID uint, phone, phoneType string) error {
	if phoneType == "" {
		phoneType = "mobile"
	}
	p := model.PersonPhone{PersonID: personID, Phone: phone, PhoneType: phoneType}
	return dao.DBFromContext(ctx).Create(&p).Error
}

func UpdatePersonPhone(ctx context.Context, id uint, phone, phoneType string) error {
	updates := map[string]interface{}{"phone": phone}
	if phoneType != "" {
		updates["phone_type"] = phoneType
	}
	return dao.DBFromContext(ctx).Model(&model.PersonPhone{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonPhone(ctx context.Context, id uint) error {
	return dao.DBFromContext(ctx).Delete(&model.PersonPhone{}, id).Error
}

func AddPersonEmail(ctx context.Context, personID uint, email, emailType string) error {
	if emailType == "" {
		emailType = "personal"
	}
	e := model.PersonEmail{PersonID: personID, Email: email, EmailType: emailType}
	return dao.DB.Create(&e).Error
}

func UpdatePersonEmail(ctx context.Context, id uint, email, emailType string) error {
	updates := map[string]interface{}{"email": email}
	if emailType != "" {
		updates["email_type"] = emailType
	}
	return dao.DBFromContext(ctx).Model(&model.PersonEmail{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonEmail(ctx context.Context, id uint) error {
	return dao.DBFromContext(ctx).Delete(&model.PersonEmail{}, id).Error
}

func AddPersonBankCard(ctx context.Context, personID uint, bankName, accountNumber, accountHolder string) error {
	c := model.PersonBankCard{PersonID: personID, BankName: bankName, AccountNumber: accountNumber, AccountHolder: accountHolder}
	return dao.DB.Create(&c).Error
}

func UpdatePersonBankCard(ctx context.Context, id uint, bankName, accountNumber, accountHolder string) error {
	updates := map[string]interface{}{"bank_name": bankName, "account_number": accountNumber, "account_holder": accountHolder}
	return dao.DBFromContext(ctx).Model(&model.PersonBankCard{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonBankCard(ctx context.Context, id uint) error {
	return dao.DBFromContext(ctx).Delete(&model.PersonBankCard{}, id).Error
}

func AddPersonEmergencyContact(ctx context.Context, personID uint, contactName, contactPhone string, sort int) error {
	if sort == 0 {
		sort = 1
	}
	c := model.PersonEmergencyContact{PersonID: personID, ContactName: contactName, ContactPhone: contactPhone, Sort: sort}
	return dao.DBFromContext(ctx).Create(&c).Error
}

func UpdatePersonEmergencyContact(ctx context.Context, id uint, contactName, contactPhone string, sort int) error {
	updates := map[string]interface{}{"contact_name": contactName, "contact_phone": contactPhone}
	if sort > 0 {
		updates["sort"] = sort
	}
	return dao.DBFromContext(ctx).Model(&model.PersonEmergencyContact{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePersonEmergencyContact(ctx context.Context, id uint) error {
	return dao.DBFromContext(ctx).Delete(&model.PersonEmergencyContact{}, id).Error
}

// PersonName 单条人员姓名查询（审计/导出等一次性场景）
func PersonName(personID uint) string {
	var name string
	dao.DB.Table("persons").Select("name").Where("id = ?", personID).Scan(&name)
	return name
}

// PersonNameMap 一次 IN 查询返回 id→name 映射，替代列表循环逐行查库（消除 N+1）
func PersonNameMap(personIDs []uint) map[uint]string {
	m := make(map[uint]string)
	if len(personIDs) == 0 {
		return m
	}
	var rows []struct {
		ID   uint
		Name string
	}
	dao.DB.Table("persons").Where("id IN ?", personIDs).Scan(&rows)
	for _, r := range rows {
		m[r.ID] = r.Name
	}
	return m
}
