package person

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"probig/internal/pkg/audit"
)

type Service struct {
	dao *DAO
}

var globalService *Service

func NewService(db *gorm.DB, encryptKey string) *Service {
	svc := &Service{dao: NewDAO(db, encryptKey)}
	globalService = svc
	return svc
}

func GetService() *Service {
	return globalService
}

func (s *Service) CreatePerson(person *Person) error {
	existing, err := s.dao.GetPersonByIDCard(person.IDCard)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil && existing.ID != 0 {
		return errors.New("身份证号已存在")
	}

	if err := s.dao.CreatePerson(person); err != nil {
		return err
	}

	afterJSON, _ := json.Marshal(person)
	audit.GlobalAuditService.Log(0, "", "person", person.ID, "create", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) GetPersonByID(id uint) (*Person, error) {
	return s.dao.GetPersonByID(id)
}

func (s *Service) ListPersons(page, pageSize int, name, idCard string) ([]Person, int64, error) {
	return s.dao.ListPersons(page, pageSize, name, idCard)
}

func (s *Service) UpdatePerson(person *Person) error {
	existing, err := s.dao.GetPersonByIDCard(person.IDCard)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil && existing.ID != 0 && existing.ID != person.ID {
		return errors.New("身份证号已存在")
	}

	before, _ := s.dao.GetPersonByID(person.ID)
	beforeJSON, _ := json.Marshal(before)

	if err := s.dao.UpdatePerson(person); err != nil {
		return err
	}

	afterJSON, _ := json.Marshal(person)
	audit.GlobalAuditService.Log(0, "", "person", person.ID, "update", string(beforeJSON), string(afterJSON), "", "")
	return nil
}

func (s *Service) DeletePerson(id uint) error {
	before, _ := s.dao.GetPersonByID(id)
	beforeJSON, _ := json.Marshal(before)

	if err := s.dao.DeletePerson(id); err != nil {
		return err
	}

	audit.GlobalAuditService.Log(0, "", "person", id, "delete", string(beforeJSON), "", "", "")
	return nil
}

func (s *Service) RestorePerson(id uint) error {
	if err := s.dao.RestorePerson(id); err != nil {
		return err
	}
	audit.GlobalAuditService.Log(0, "", "person", id, "restore", "", "", "", "")
	return nil
}

func (s *Service) GetPersonByIDCard(idCard string) (*Person, error) {
	return s.dao.GetPersonByIDCard(idCard)
}

func (s *Service) GetAllPersonsSimple() ([]Person, error) {
	return s.dao.GetAllPersonsSimple()
}

func (s *Service) CreatePhone(phone *PersonPhone) error {
	if err := s.dao.CreatePhone(phone); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(phone)
	audit.GlobalAuditService.Log(0, "", "person_phone", phone.ID, "create", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) UpdatePhone(phone *PersonPhone) error {
	if err := s.dao.UpdatePhone(phone); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(phone)
	audit.GlobalAuditService.Log(0, "", "person_phone", phone.ID, "update", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) DeletePhone(id uint) error {
	if err := s.dao.DeletePhone(id); err != nil {
		return err
	}
	audit.GlobalAuditService.Log(0, "", "person_phone", id, "delete", "", "", "", "")
	return nil
}

func (s *Service) CreateEmail(email *PersonEmail) error {
	if err := s.dao.CreateEmail(email); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(email)
	audit.GlobalAuditService.Log(0, "", "person_email", email.ID, "create", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) UpdateEmail(email *PersonEmail) error {
	if err := s.dao.UpdateEmail(email); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(email)
	audit.GlobalAuditService.Log(0, "", "person_email", email.ID, "update", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) DeleteEmail(id uint) error {
	if err := s.dao.DeleteEmail(id); err != nil {
		return err
	}
	audit.GlobalAuditService.Log(0, "", "person_email", id, "delete", "", "", "", "")
	return nil
}

func (s *Service) CreateBankCard(card *PersonBankCard) error {
	if err := s.dao.CreateBankCard(card); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(card)
	audit.GlobalAuditService.Log(0, "", "person_bank_card", card.ID, "create", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) UpdateBankCard(card *PersonBankCard) error {
	if err := s.dao.UpdateBankCard(card); err != nil {
		return err
	}
	afterJSON, _ := json.Marshal(card)
	audit.GlobalAuditService.Log(0, "", "person_bank_card", card.ID, "update", "", string(afterJSON), "", "")
	return nil
}

func (s *Service) DeleteBankCard(id uint) error {
	if err := s.dao.DeleteBankCard(id); err != nil {
		return err
	}
	audit.GlobalAuditService.Log(0, "", "person_bank_card", id, "delete", "", "", "", "")
	return nil
}
