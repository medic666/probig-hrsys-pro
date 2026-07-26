package person

import (
	"fmt"
	"time"
	"probig/internal/pkg/config"
	"probig/internal/pkg/database"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService() *Service {
	return &Service{DB: config.DB}
}

func (s *Service) List(pageNum, pageSize int, name, idCard, attendanceGroup, status string) ([]map[string]interface{}, int64, error) {
	var persons []Person
	var total int64
	db := s.DB.Model(&Person{})
	if name != "" {
		db = db.Where("name like ?", "%"+name+"%")
	}
	if idCard != "" {
		encrypted := database.Encrypt(idCard)
		db = db.Where("id_card = ?", encrypted)
	}

	if attendanceGroup != "" || status != "" {
		subQuery := s.DB.Table("position_snapshot AS ps").
			Select("1").
			Where("ps.person_id = persons.id").
			Where("ps.effective_start_date <= ? AND ps.effective_end_date >= ?",
				time.Now().Format("2006-01-02"), time.Now().Format("2006-01-02"))
		if attendanceGroup != "" {
			subQuery = subQuery.Where("ps.attendance_group = ?", attendanceGroup)
		}
		db = db.Where("EXISTS (?)", subQuery)
	}

	db.Count(&total)
	if err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&persons).Error; err != nil {
		return nil, 0, err
	}

	var result []map[string]interface{}
	for _, p := range persons {
		item := map[string]interface{}{
			"id":               p.ID,
			"name":             p.Name,
			"id_card":          database.Decrypt(p.IDCard),
			"gender":           p.Gender,
			"birthday":         formatDate(p.Birthday),
			"nation":           p.Nation,
			"native_place":     p.NativePlace,
			"address":          p.Address,
			"political_status": p.PoliticalStatus,
			"marital_status":   p.MaritalStatus,
			"alias":            p.Alias,
			"created_at":       p.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		result = append(result, item)
	}
	return result, total, nil
}

func (s *Service) GetByID(id uint) (map[string]interface{}, error) {
	var p Person
	if err := s.DB.First(&p, id).Error; err != nil {
		return nil, err
	}

	p.IDCard = database.Decrypt(p.IDCard)

	phones := s.GetPhones(id)
	emails := s.GetEmails(id)
	cards := s.GetBankCards(id)

	return map[string]interface{}{
		"id":               p.ID,
		"name":             p.Name,
		"id_card":          p.IDCard,
		"gender":           p.Gender,
		"birthday":         formatDate(p.Birthday),
		"nation":           p.Nation,
		"native_place":     p.NativePlace,
		"address":          p.Address,
		"political_status": p.PoliticalStatus,
		"marital_status":   p.MaritalStatus,
		"alias":            p.Alias,
		"phones":           phones,
		"emails":           emails,
		"bank_cards":       cards,
		"created_at":       p.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *Service) Create(req map[string]interface{}) (uint, error) {
	person := Person{
		Name:            getString(req, "name"),
		IDCard:          database.Encrypt(getString(req, "id_card")),
		Gender:          getInt(req, "gender"),
		Nation:          getString(req, "nation"),
		NativePlace:     getString(req, "native_place"),
		Address:         getString(req, "address"),
		PoliticalStatus: getString(req, "political_status"),
		MaritalStatus:   getInt(req, "marital_status"),
		Alias:           getString(req, "alias"),
	}
	if b := getString(req, "birthday"); b != "" {
		t := parseDate(b)
		person.Birthday = &t
	}

	if err := s.DB.Create(&person).Error; err != nil {
		return 0, fmt.Errorf("创建失败，身份证号可能已存在")
	}
	return person.ID, nil
}

func (s *Service) Update(id uint, req map[string]interface{}) error {
	updates := map[string]interface{}{}
	if v, ok := req["name"]; ok && v.(string) != "" {
		updates["name"] = v
	}
	if v, ok := req["id_card"]; ok && v.(string) != "" {
		updates["id_card"] = database.Encrypt(v.(string))
	}
	if v, ok := req["gender"]; ok {
		updates["gender"] = v
	}
	if v, ok := req["nation"]; ok {
		updates["nation"] = v
	}
	if v, ok := req["native_place"]; ok {
		updates["native_place"] = v
	}
	if v, ok := req["address"]; ok {
		updates["address"] = v
	}
	if v, ok := req["political_status"]; ok {
		updates["political_status"] = v
	}
	if v, ok := req["marital_status"]; ok {
		updates["marital_status"] = v
	}
	if v, ok := req["alias"]; ok {
		updates["alias"] = v
	}
	if v, ok := req["birthday"]; ok && v.(string) != "" {
		t := parseDate(v.(string))
		updates["birthday"] = &t
	}
	return s.DB.Model(&Person{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) Delete(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
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

func (s *Service) Restore(id uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&Person{}).Where("id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		tx.Unscoped().Model(&PersonPhone{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&PersonEmail{}).Where("person_id = ?", id).Update("deleted_at", nil)
		tx.Unscoped().Model(&PersonBankCard{}).Where("person_id = ?", id).Update("deleted_at", nil)
		return nil
	})
}

func (s *Service) GetDeletedList(pageNum, pageSize int) ([]Person, int64, error) {
	var list []Person
	var total int64
	db := s.DB.Unscoped().Where("deleted_at is not null")
	db.Model(&Person{}).Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("deleted_at desc").Find(&list).Error
	return list, total, err
}

func (s *Service) AddPhone(personID uint, phone string) error {
	return s.DB.Create(&PersonPhone{PersonID: personID, Phone: database.Encrypt(phone)}).Error
}

func (s *Service) UpdatePhone(id uint, phone string) error {
	return s.DB.Model(&PersonPhone{}).Where("id = ?", id).Update("phone", database.Encrypt(phone)).Error
}

func (s *Service) DeletePhone(id uint) error {
	return s.DB.Delete(&PersonPhone{}, id).Error
}

func (s *Service) GetPhones(personID uint) []PersonPhone {
	var phones []PersonPhone
	s.DB.Where("person_id = ?", personID).Find(&phones)
	for i := range phones {
		phones[i].Phone = database.Decrypt(phones[i].Phone)
	}
	return phones
}

func (s *Service) AddEmail(personID uint, email string) error {
	return s.DB.Create(&PersonEmail{PersonID: personID, Email: database.Encrypt(email)}).Error
}

func (s *Service) UpdateEmail(id uint, email string) error {
	return s.DB.Model(&PersonEmail{}).Where("id = ?", id).Update("email", database.Encrypt(email)).Error
}

func (s *Service) DeleteEmail(id uint) error {
	return s.DB.Delete(&PersonEmail{}, id).Error
}

func (s *Service) GetEmails(personID uint) []PersonEmail {
	var emails []PersonEmail
	s.DB.Where("person_id = ?", personID).Find(&emails)
	for i := range emails {
		emails[i].Email = database.Decrypt(emails[i].Email)
	}
	return emails
}

func (s *Service) AddBankCard(personID uint, cardNo, bankName string) error {
	return s.DB.Create(&PersonBankCard{PersonID: personID, CardNo: database.Encrypt(cardNo), BankName: bankName}).Error
}

func (s *Service) UpdateBankCard(id uint, cardNo, bankName string) error {
	return s.DB.Model(&PersonBankCard{}).Where("id = ?", id).Updates(map[string]interface{}{
		"card_no":   database.Encrypt(cardNo),
		"bank_name": bankName,
	}).Error
}

func (s *Service) DeleteBankCard(id uint) error {
	return s.DB.Delete(&PersonBankCard{}, id).Error
}

func (s *Service) GetBankCards(personID uint) []PersonBankCard {
	var cards []PersonBankCard
	s.DB.Where("person_id = ?", personID).Find(&cards)
	for i := range cards {
		cards[i].CardNo = database.Decrypt(cards[i].CardNo)
	}
	return cards
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return 0
}

func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}
