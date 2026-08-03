package service

import (
	"context"
	"errors"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/utils"

	"gorm.io/gorm"
)

// PersonListRow 人员列表行：基础信息 + 当前职务快照信息（公司/部门/职位/在职状态/入职离职日期）
type PersonListRow struct {
	model.Person
	CompanyID   uint            `json:"company_id"`
	CompanyName string          `json:"company_name"`
	Department  string          `json:"department"`
	Position    string          `json:"position"`
	IsActive    *bool           `json:"is_active"`
	EntryDate   *utils.DateOnly `json:"entry_date"`
	LeaveDate   *utils.DateOnly `json:"leave_date"`
}

// PersonListQuery 人员列表查询（列表与导出共用，筛选参数单一来源）
type PersonListQuery struct {
	PageNum    int
	PageSize   int
	Name       string
	IDCard     string
	PersonID   string
	CompanyID  uint
	Department string
	Status     string
}

func GetPersonList(q PersonListQuery) ([]PersonListRow, int64, error) {
	tx := dao.DB.Model(&model.Person{}).Preload("Phones").Preload("Emails").Preload("BankCards").Preload("EmergencyContacts")
	if q.PersonID != "" {
		tx = tx.Where("id = ?", q.PersonID)
	}
	if q.Name != "" {
		tx = tx.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.IDCard != "" {
		tx = tx.Where("id_card LIKE ?", "%"+q.IDCard+"%")
	}

	// 当前快照段过滤：公司/部门/在职状态（未入职 = 无快照段）
	if q.CompanyID > 0 || q.Department != "" || q.Status != "" {
		snapTx := dao.DB.Model(&model.PositionSnapshot{}).
			Select("person_id").
			Where("effective_end_date = ?", realFarFuture)
		if q.CompanyID > 0 {
			snapTx = snapTx.Where("company_id = ?", q.CompanyID)
		}
		if q.Department != "" {
			snapTx = snapTx.Where("department LIKE ?", "%"+q.Department+"%")
		}
		if q.Status == "active" {
			snapTx = snapTx.Where("is_active = true")
		}
		if q.Status == "left" {
			snapTx = snapTx.Where("is_active = false")
		}
		if q.Status == "not_entered" {
			tx = tx.Where("id NOT IN (?)", snapTx)
		} else {
			tx = tx.Where("id IN (?)", snapTx)
		}
	}

	var total int64
	tx.Count(&total)

	var persons []model.Person
	offset := (q.PageNum - 1) * q.PageSize
	if err := tx.Offset(offset).Limit(q.PageSize).Order("id DESC").Find(&persons).Error; err != nil {
		return nil, 0, err
	}

	ids := make([]uint, len(persons))
	for i, p := range persons {
		ids[i] = p.ID
	}

	snapMap := make(map[uint]PersonListRow, len(ids))
	if len(ids) > 0 {
		var rows []struct {
			PersonID    uint
			CompanyID   uint
			CompanyName string
			Department  string
			Position    string
			IsActive    bool
			EntryDate   *utils.DateOnly
			LeaveDate   *utils.DateOnly
		}
		dao.DB.Table("position_snapshots s").
			Select("s.person_id, s.company_id, s.department, s.position, s.is_active, s.entry_date, s.leave_date, c.name AS company_name").
			Joins("LEFT JOIN companies c ON c.id = s.company_id").
			Where("s.person_id IN ? AND s.effective_end_date = ?", ids, realFarFuture).
			Scan(&rows)
		for _, r := range rows {
			active := r.IsActive
			snapMap[r.PersonID] = PersonListRow{
				CompanyID:   r.CompanyID,
				CompanyName: r.CompanyName,
				Department:  r.Department,
				Position:    r.Position,
				IsActive:    &active,
				EntryDate:   r.EntryDate,
				LeaveDate:   r.LeaveDate,
			}
		}
	}

	list := make([]PersonListRow, len(persons))
	for i, p := range persons {
		row := snapMap[p.ID]
		row.Person = p
		list[i] = row
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

// PersonProfile 人员聚合档案：基础字段 + 四类子表（提交时按主键同步，未变化零审计）
type PersonProfile struct {
	model.Person
	Phones            []model.PersonPhone            `json:"phones"`
	Emails            []model.PersonEmail            `json:"emails"`
	BankCards         []model.PersonBankCard         `json:"bank_cards"`
	EmergencyContacts []model.PersonEmergencyContact `json:"emergency_contacts"`
}

func syncPersonChildren(tx *gorm.DB, id uint, req *PersonProfile) error {
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
}

// UpsertPersonProfile 人员档案统一 upsert（新增=编辑同一入口）：
// req.ID == 0 → 事务内创建人员并同步四类子表；req.ID > 0 → 更新基础字段 + 同步子表
func UpsertPersonProfile(ctx context.Context, req *PersonProfile) (*model.Person, error) {
	if req.Name == "" {
		return nil, errors.New("姓名不能为空")
	}
	var person model.Person
	err := utils.WithTransaction(dao.DBFromContext(ctx), func(tx *gorm.DB) error {
		if req.ID == 0 {
			person = req.Person
			person.ID = 0
			if person.IDCard != "" {
				var count int64
				tx.Model(&model.Person{}).Where("id_card = ?", person.IDCard).Count(&count)
				if count > 0 {
					return errors.New("身份证号已存在")
				}
			}
			if err := tx.Create(&person).Error; err != nil {
				return err
			}
			req.ID = person.ID
		} else {
			var existing model.Person
			if err := tx.First(&existing, req.ID).Error; err != nil {
				return errors.New("人员不存在")
			}
			if req.IDCard != "" && req.IDCard != existing.IDCard {
				var count int64
				tx.Model(&model.Person{}).Where("id_card = ? AND id != ?", req.IDCard, req.ID).Count(&count)
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
			person = existing
		}
		return syncPersonChildren(tx, req.ID, req)
	})
	if err != nil {
		return nil, err
	}
	return &person, nil
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

// PersonOption 人员选项：基础信息 + 当前快照段域字段（公司/考勤组/在职状态）。
// 供人员选择组件（全体/公司/考勤组/在职状态多域筛选）与 NameSelect 共用，
// 域字段随当前快照段（9999-12-31 结束）JOIN 公司表一次取得。
type PersonOption struct {
	ID              uint            `json:"id"`
	Name            string          `json:"name"`
	CompanyID       uint            `json:"company_id"`
	CompanyName     string          `json:"company_name"`
	AttendanceGroup string          `json:"attendance_group"`
	IsActive        bool            `json:"is_active"`
	EntryDate       *utils.DateOnly `json:"entry_date"`
	LeaveDate       *utils.DateOnly `json:"leave_date"`
}

func GetAllPersons() ([]PersonOption, error) {
	var list []PersonOption
	err := dao.DB.Table("persons").
		Select(`persons.id, persons.name,
			s.company_id, c.name AS company_name, s.attendance_group,
			COALESCE(s.is_active, false) AS is_active, s.entry_date, s.leave_date`).
		Joins(`LEFT JOIN position_snapshots s ON s.person_id = persons.id
			AND s.effective_end_date = ?`, realFarFuture).
		Joins("LEFT JOIN companies c ON c.id = s.company_id").
		Where("persons.deleted_at IS NULL").
		Order("persons.name").
		Scan(&list).Error
	return list, err
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
