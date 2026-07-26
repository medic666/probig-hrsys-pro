package person

import (
	"time"

	"probig/internal/pkg/database"

	"gorm.io/gorm"
)

func CreatePerson(tx *gorm.DB, person *database.Person) error {
	return tx.Create(person).Error
}

func UpdatePerson(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.Person{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePerson(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.Person{}, id).Error
}

func RestorePerson(tx *gorm.DB, id uint) error {
	return tx.Unscoped().Model(&database.Person{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func GetPersonByID(id uint) (*database.Person, error) {
	var person database.Person
	err := database.DB.Where("id = ?", id).First(&person).Error
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func IDCardExists(idCard string, excludeID uint) (bool, error) {
	var count int64
	query := database.DB.Model(&database.Person{}).Where("id_card = ?", idCard)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

type ListPersonsFilter struct {
	Name            string
	IDCard          string
	AttendanceGroup string
	EmploymentStatus string
}

type PersonListRow struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	IDCard           string    `json:"id_card"`
	Gender           int8      `json:"gender"`
	Aliaa            string    `json:"alias"`
	AttendanceGroup  string    `json:"attendance_group"`
	EmploymentStatus string    `json:"employment_status"`
	CreatedAt       time.Time `json:"created_at"`
}

func ListPersons(pageNum, pageSize int, filter ListPersonsFilter) ([]PersonListRow, int64, error) {
	var results []PersonListRow
	var total int64

	baseQuery := database.DB.Table("person").
		Select(`person.id, person.name, person.id_card, person.gender, person.alias, person.created_at,
			COALESCE(latest_ps.attendance_group, '') as attendance_group,
			CASE
				WHEN latest_ps.effective_end_date = '9999-12-31' AND latest_ps.leave_date IS NULL THEN '在职'
				WHEN latest_ps.effective_end_date = '9999-12-31' THEN '离职'
				ELSE '无记录'
			END as employment_status`).
		Joins(`LEFT JOIN position_snapshots latest_ps ON latest_ps.person_id = person.id
			AND latest_ps.id = (
				SELECT ps2.id FROM position_snapshots ps2
				WHERE ps2.person_id = person.id AND ps2.deleted_at IS NULL
				ORDER BY ps2.effective_start_date DESC LIMIT 1
			)
			AND latest_ps.deleted_at IS NULL`).
		Where("person.deleted_at IS NULL")

	if filter.Name != "" {
		baseQuery = baseQuery.Where("person.name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.IDCard != "" {
		baseQuery = baseQuery.Where("person.id_card LIKE ?", "%"+filter.IDCard+"%")
	}

	if filter.AttendanceGroup != "" || filter.EmploymentStatus != "" {
		subQuery := baseQuery

		if filter.AttendanceGroup != "" {
			baseQuery = baseQuery.Where("latest_ps.attendance_group = ?", filter.AttendanceGroup)
		}
		if filter.EmploymentStatus != "" {
			_ = subQuery
			if filter.EmploymentStatus == "在职" {
				baseQuery = baseQuery.Where("latest_ps.effective_end_date = '9999-12-31' AND latest_ps.leave_date IS NULL")
			} else if filter.EmploymentStatus == "离职" {
				baseQuery = baseQuery.Where("latest_ps.effective_end_date = '9999-12-31' AND latest_ps.leave_date IS NOT NULL")
			}
		}
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Order("person.id DESC").Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func ListTrashPersons(pageNum, pageSize int) ([]database.Person, int64, error) {
	var persons []database.Person
	var total int64

	query := database.DB.Unscoped().Where("deleted_at IS NOT NULL")

	if err := query.Model(&database.Person{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&persons).Error; err != nil {
		return nil, 0, err
	}

	return persons, total, nil
}

func CascadeDeletePerson(tx *gorm.DB, personID uint) error {
	personTables := []struct {
		table string
		field string
	}{
		{"person_phone", "person_id"},
		{"person_email", "person_id"},
		{"person_bank_card", "person_id"},
		{"position_event", "person_id"},
		{"attendance_event", "person_id"},
		{"leave_account_event", "person_id"},
		{"salary_event", "person_id"},
	}

	for _, t := range personTables {
		if err := tx.Table(t.table).Where(t.field+" = ?", personID).Update("deleted_at", gorm.Expr("datetime('now')")).Error; err != nil {
			return err
		}
	}

	if err := tx.Table("file_relation").Where("target_type = ? AND target_id = ?", "person", personID).Update("deleted_at", gorm.Expr("datetime('now')")).Error; err != nil {
		return err
	}

	return tx.Delete(&database.Person{}, personID).Error
}

func CascadeRestorePerson(tx *gorm.DB, personID uint) error {
	personTables := []struct {
		table string
		field string
	}{
		{"person_phone", "person_id"},
		{"person_email", "person_id"},
		{"person_bank_card", "person_id"},
		{"position_event", "person_id"},
		{"attendance_event", "person_id"},
		{"leave_account_event", "person_id"},
		{"salary_event", "person_id"},
	}

	for _, t := range personTables {
		if err := tx.Table(t.table).Where(t.field+" = ?", personID).Update("deleted_at", nil).Error; err != nil {
			return err
		}
	}

	if err := tx.Table("file_relation").Where("target_type = ? AND target_id = ?", "person", personID).Update("deleted_at", nil).Error; err != nil {
		return err
	}

	return tx.Unscoped().Model(&database.Person{}).Where("id = ?", personID).Update("deleted_at", nil).Error
}

func GetPhonesByPersonID(personID uint) ([]database.PersonPhone, error) {
	var phones []database.PersonPhone
	err := database.DB.Where("person_id = ?", personID).Find(&phones).Error
	return phones, err
}

func GetPhoneByID(id uint) (*database.PersonPhone, error) {
	var phone database.PersonPhone
	err := database.DB.Where("id = ?", id).First(&phone).Error
	if err != nil {
		return nil, err
	}
	return &phone, nil
}

func CreatePhone(tx *gorm.DB, phone *database.PersonPhone) error {
	return tx.Create(phone).Error
}

func UpdatePhone(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.PersonPhone{}).Where("id = ?", id).Updates(updates).Error
}

func DeletePhone(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.PersonPhone{}, id).Error
}

func GetEmailsByPersonID(personID uint) ([]database.PersonEmail, error) {
	var emails []database.PersonEmail
	err := database.DB.Where("person_id = ?", personID).Find(&emails).Error
	return emails, err
}

func GetEmailByID(id uint) (*database.PersonEmail, error) {
	var email database.PersonEmail
	err := database.DB.Where("id = ?", id).First(&email).Error
	if err != nil {
		return nil, err
	}
	return &email, nil
}

func CreateEmail(tx *gorm.DB, email *database.PersonEmail) error {
	return tx.Create(email).Error
}

func UpdateEmail(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.PersonEmail{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteEmail(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.PersonEmail{}, id).Error
}

func GetBankCardsByPersonID(personID uint) ([]database.PersonBankCard, error) {
	var cards []database.PersonBankCard
	err := database.DB.Where("person_id = ?", personID).Find(&cards).Error
	return cards, err
}

func GetBankCardByID(id uint) (*database.PersonBankCard, error) {
	var card database.PersonBankCard
	err := database.DB.Where("id = ?", id).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func CreateBankCard(tx *gorm.DB, card *database.PersonBankCard) error {
	return tx.Create(card).Error
}

func UpdateBankCard(tx *gorm.DB, id uint, updates map[string]interface{}) error {
	return tx.Model(&database.PersonBankCard{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteBankCard(tx *gorm.DB, id uint) error {
	return tx.Delete(&database.PersonBankCard{}, id).Error
}
