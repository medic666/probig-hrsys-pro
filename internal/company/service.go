package company

import (
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

func (s *Service) List(pageNum, pageSize int, name, creditCode string) ([]map[string]interface{}, int64, error) {
	var list []Company
	var total int64
	db := s.DB.Model(&Company{})
	if name != "" {
		db = db.Where("name like ?", "%"+name+"%")
	}
	if creditCode != "" {
		db = db.Where("credit_code like ?", "%"+creditCode+"%")
	}
	db.Count(&total)
	if err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	var result []map[string]interface{}
	for _, c := range list {
		result = append(result, map[string]interface{}{
			"id":            c.ID,
			"name":          c.Name,
			"credit_code":   c.CreditCode,
			"address":       c.Address,
			"contact_phone": c.ContactPhone,
			"bank_name":     c.BankName,
			"bank_account":  database.Decrypt(c.BankAccount),
			"created_at":    c.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return result, total, nil
}

func (s *Service) GetByID(id uint) (map[string]interface{}, error) {
	var c Company
	if err := s.DB.First(&c, id).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":            c.ID,
		"name":          c.Name,
		"credit_code":   c.CreditCode,
		"address":       c.Address,
		"contact_phone": c.ContactPhone,
		"bank_name":     c.BankName,
		"bank_account":  database.Decrypt(c.BankAccount),
	}, nil
}

func (s *Service) Create(req map[string]interface{}) (uint, error) {
	c := Company{
		Name:         getStr(req, "name"),
		CreditCode:   getStr(req, "credit_code"),
		Address:      getStr(req, "address"),
		ContactPhone: getStr(req, "contact_phone"),
		BankName:     getStr(req, "bank_name"),
		BankAccount:  database.Encrypt(getStr(req, "bank_account")),
	}
	if err := s.DB.Create(&c).Error; err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (s *Service) Update(id uint, req map[string]interface{}) error {
	updates := map[string]interface{}{}
	for _, k := range []string{"name", "credit_code", "address", "contact_phone", "bank_name"} {
		if v := getStr(req, k); v != "" {
			updates[k] = v
		}
	}
	if v := getStr(req, "bank_account"); v != "" {
		updates["bank_account"] = database.Encrypt(v)
	}
	return s.DB.Model(&Company{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Service) Delete(id uint) error {
	return s.DB.Delete(&Company{}, id).Error
}

func (s *Service) Restore(id uint) error {
	return s.DB.Unscoped().Model(&Company{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (s *Service) GetDeletedList(pageNum, pageSize int) ([]Company, int64, error) {
	var list []Company
	var total int64
	db := s.DB.Unscoped().Where("deleted_at is not null")
	db.Model(&Company{}).Count(&total)
	err := db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Order("deleted_at desc").Find(&list).Error
	return list, total, err
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		s, _ := v.(string)
		return s
	}
	return ""
}
