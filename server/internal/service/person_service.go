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

// buildPersonUpdates 基础字段差异更新 map：仅返回发生变化字段（未变化零操作零审计）
func buildPersonUpdates(existing *model.Person, p *model.Person) map[string]interface{} {
	updates := map[string]interface{}{}
	if p.Name != existing.Name { updates["name"] = p.Name }
	if p.IDCard != existing.IDCard { updates["id_card"] = p.IDCard }
	if p.Gender != existing.Gender { updates["gender"] = p.Gender }
	if !sameDate(p.Birthday, existing.Birthday) { updates["birthday"] = p.Birthday }
	if p.Nation != existing.Nation { updates["nation"] = p.Nation }
	if p.NativePlace != existing.NativePlace { updates["native_place"] = p.NativePlace }
	if p.Address != existing.Address { updates["address"] = p.Address }
	if p.PoliticalStatus != existing.PoliticalStatus { updates["political_status"] = p.PoliticalStatus }
	if p.MaritalStatus != existing.MaritalStatus { updates["marital_status"] = p.MaritalStatus }
	if p.Alias != existing.Alias { updates["alias"] = p.Alias }
	return updates
}

func sameDate(a, b *utils.DateOnly) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
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

	updates := buildPersonUpdates(&existing, p)
	if len(updates) == 0 {
		return nil
	}
	return dao.DBFromContext(ctx).Model(&existing).Updates(updates).Error
}

// PersonProfile 人员聚合档案：基础字段 + 四类子表（提交时按主键同步，未变化零审计）
type PersonProfile struct {
	model.Person
	Phones            []model.PersonPhone            `json:"phones"`
	Emails            []model.PersonEmail            `json:"emails"`
	BankCards         []model.PersonBankCard         `json:"bank_cards"`
	EmergencyContacts []model.PersonEmergencyContact `json:"emergency_contacts"`
}

// UpdatePersonProfile 聚合更新人员档案（事务）：基础字段 + 四类子表同步（UPSERT）
func UpdatePersonProfile(ctx context.Context, id uint, req *PersonProfile) error {
	if req.Name == "" {
		return errors.New("姓名不能为空")
	}
	return utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		var existing model.Person
		if err := tx.First(&existing, id).Error; err != nil {
			return errors.New("人员不存在")
		}
		if req.IDCard != "" && req.IDCard != existing.IDCard {
			var count int64
			tx.Model(&model.Person{}).Where("id_card = ? AND id != ?", req.IDCard, id).Count(&count)
			if count > 0 {
				return errors.New("身份证号已存在")
			}
		}
		updates := buildPersonUpdates(&existing, &req.Person)
		if len(updates) > 0 {
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}

		syncPhones := func(tx *gorm.DB) error {
			return SyncChildRecords(tx, "person_id", id, req.Phones,
				func(p model.PersonPhone) uint { return p.ID },
				func(a, b model.PersonPhone) bool { return a.PhoneType == b.PhoneType && a.Phone == b.Phone },
				func(p *model.PersonPhone) { p.PersonID = id })
		}
		syncEmails := func(tx *gorm.DB) error {
			return SyncChildRecords(tx, "person_id", id, req.Emails,
				func(e model.PersonEmail) uint { return e.ID },
				func(a, b model.PersonEmail) bool { return a.EmailType == b.EmailType && a.Email == b.Email },
				func(e *model.PersonEmail) { e.PersonID = id })
		}
		syncBankCards := func(tx *gorm.DB) error {
			return SyncChildRecords(tx, "person_id", id, req.BankCards,
				func(b model.PersonBankCard) uint { return b.ID },
				func(a, b model.PersonBankCard) bool {
					return a.BankName == b.BankName && a.AccountNumber == b.AccountNumber && a.AccountHolder == b.AccountHolder
				},
				func(b *model.PersonBankCard) { b.PersonID = id })
		}
		syncContacts := func(tx *gorm.DB) error {
			return SyncChildRecords(tx, "person_id", id, req.EmergencyContacts,
				func(c model.PersonEmergencyContact) uint { return c.ID },
				func(a, b model.PersonEmergencyContact) bool {
					return a.ContactName == b.ContactName && a.ContactPhone == b.ContactPhone && a.Sort == b.Sort
				},
				func(c *model.PersonEmergencyContact) { c.PersonID = id })
		}
		for _, fn := range []func(*gorm.DB) error{syncPhones, syncEmails, syncBankCards, syncContacts} {
			if err := fn(tx); err != nil {
				return err
			}
		}
		return nil
	})
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

// PersonCard 人员卡片：基本信息 + 当前职务快照（公司/部门/职位/在职状态）
type PersonCard struct {
	ID          uint            `json:"id"`
	PersonID    uint            `json:"person_id"`
	Name        string          `json:"name"`
	CompanyID   uint            `json:"company_id"`
	CompanyName string          `json:"company_name"`
	Department  string          `json:"department"`
	Position    string          `json:"position"`
	IsActive    bool            `json:"is_active"`
	EntryDate   *utils.DateOnly `json:"entry_date"`
	LeaveDate   *utils.DateOnly `json:"leave_date"`
}

// GetPersonCards 人员卡片列表：以当前职务快照段（9999-12-31 结束）关联公司/部门/职位/在职状态；
// 无快照段者（未入职）EntryDate 为空、IsActive 为 false
func GetPersonCards() ([]PersonCard, error) {
	var cards []PersonCard
	err := dao.DB.Table("persons").
		Select(`persons.id, persons.id AS person_id, persons.name,
			s.company_id, c.name AS company_name, s.department, s.position,
			s.is_active, s.entry_date, s.leave_date`).
		Joins(`LEFT JOIN position_snapshots s ON s.person_id = persons.id
			AND s.effective_end_date = ?`, realFarFuture).
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
