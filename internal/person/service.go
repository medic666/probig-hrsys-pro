package person

import (
	"errors"
	"time"

	"probig/internal/pkg/database"
	"probig/internal/pkg/utils"

	"gorm.io/gorm"
)

type Service struct{}

var DefaultService = &Service{}

func (s *Service) CreatePerson(req *CreatePersonRequest) (*database.Person, error) {
	if req.Name == "" {
		return nil, errors.New("人员姓名不能为空")
	}

	if req.IDCard != "" {
		exists, err := IDCardExists(req.IDCard, 0)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("身份证号已存在")
		}
	}

	person := &database.Person{
		Name:            req.Name,
		IDCard:          req.IDCard,
		Nation:          req.Nation,
		NativePlace:     req.NativePlace,
		Address:         req.Address,
		PoliticalStatus: req.PoliticalStatus,
		Alias:           req.Alias,
	}

	if req.Gender != nil {
		person.Gender = *req.Gender
	}
	if req.MaritalStatus != nil {
		person.MaritalStatus = *req.MaritalStatus
	}
	if req.Birthday != "" {
		t, err := time.Parse(utils.DateFormat, req.Birthday)
		if err == nil {
			person.Birthday = &t
		}
	}

	if err := CreatePerson(nil, person); err != nil {
		return nil, err
	}
	return person, nil
}

func (s *Service) UpdatePerson(id uint, req *UpdatePersonRequest) error {
	person, err := GetPersonByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.IDCard != "" && req.IDCard != person.IDCard {
		exists, err := IDCardExists(req.IDCard, id)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("身份证号已存在")
		}
		updates["id_card"] = req.IDCard
	}
	if req.Gender != nil {
		updates["gender"] = *req.Gender
	}
	if req.Birthday != "" {
		t, err := time.Parse(utils.DateFormat, req.Birthday)
		if err == nil {
			updates["birthday"] = t
		}
	}
	if req.Nation != "" {
		updates["nation"] = req.Nation
	}
	if req.NativePlace != "" {
		updates["native_place"] = req.NativePlace
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.PoliticalStatus != "" {
		updates["political_status"] = req.PoliticalStatus
	}
	if req.MaritalStatus != nil {
		updates["marital_status"] = *req.MaritalStatus
	}
	if req.Alias != "" {
		updates["alias"] = req.Alias
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdatePerson(nil, id, updates)
}

func (s *Service) DeletePerson(id uint) error {
	_, err := GetPersonByID(id)
	if err != nil {
		return err
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return CascadeDeletePerson(tx, id)
	})
}

func (s *Service) RestorePerson(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return CascadeRestorePerson(tx, id)
	})
}

func (s *Service) GetPersonDetail(id uint) (*PersonDetailResponse, error) {
	person, err := GetPersonByID(id)
	if err != nil {
		return nil, err
	}

	phones, err := GetPhonesByPersonID(id)
	if err != nil {
		return nil, err
	}

	emails, err := GetEmailsByPersonID(id)
	if err != nil {
		return nil, err
	}

	bankCards, err := GetBankCardsByPersonID(id)
	if err != nil {
		return nil, err
	}

	return &PersonDetailResponse{
		Person:    person,
		Phones:    phones,
		Emails:    emails,
		BankCards: bankCards,
	}, nil
}

func (s *Service) ListPersons(pageNum, pageSize int, name, idCard, attendanceGroup, employmentStatus string) ([]PersonListRow, int64, error) {
	return ListPersons(pageNum, pageSize, ListPersonsFilter{
		Name:             name,
		IDCard:           idCard,
		AttendanceGroup:  attendanceGroup,
		EmploymentStatus: employmentStatus,
	})
}

func (s *Service) ListTrashPersons(pageNum, pageSize int) ([]database.Person, int64, error) {
	return ListTrashPersons(pageNum, pageSize)
}

func (s *Service) CreatePhone(personID uint, req *CreatePhoneRequest) (*database.PersonPhone, error) {
	if req.Phone == "" {
		return nil, errors.New("电话号码不能为空")
	}

	phone := &database.PersonPhone{
		PersonID:  personID,
		Phone:     req.Phone,
		PhoneType: req.PhoneType,
	}

	if err := CreatePhone(nil, phone); err != nil {
		return nil, err
	}
	return phone, nil
}

func (s *Service) UpdatePhone(id uint, req *UpdatePhoneRequest) error {
	_, err := GetPhoneByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.PhoneType != "" {
		updates["phone_type"] = req.PhoneType
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdatePhone(nil, id, updates)
}

func (s *Service) DeletePhone(id uint) error {
	_, err := GetPhoneByID(id)
	if err != nil {
		return err
	}
	return DeletePhone(nil, id)
}

func (s *Service) CreateEmail(personID uint, req *CreateEmailRequest) (*database.PersonEmail, error) {
	if req.Email == "" {
		return nil, errors.New("邮箱不能为空")
	}

	email := &database.PersonEmail{
		PersonID:  personID,
		Email:     req.Email,
		EmailType: req.EmailType,
	}

	if err := CreateEmail(nil, email); err != nil {
		return nil, err
	}
	return email, nil
}

func (s *Service) UpdateEmail(id uint, req *UpdateEmailRequest) error {
	_, err := GetEmailByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.EmailType != "" {
		updates["email_type"] = req.EmailType
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdateEmail(nil, id, updates)
}

func (s *Service) DeleteEmail(id uint) error {
	_, err := GetEmailByID(id)
	if err != nil {
		return err
	}
	return DeleteEmail(nil, id)
}

func (s *Service) CreateBankCard(personID uint, req *CreateBankCardRequest) (*database.PersonBankCard, error) {
	if req.BankName == "" {
		return nil, errors.New("银行名称不能为空")
	}
	if req.CardNo == "" {
		return nil, errors.New("银行卡号不能为空")
	}

	card := &database.PersonBankCard{
		PersonID: personID,
		BankName: req.BankName,
		CardNo:   req.CardNo,
	}

	if err := CreateBankCard(nil, card); err != nil {
		return nil, err
	}
	return card, nil
}

func (s *Service) UpdateBankCard(id uint, req *UpdateBankCardRequest) error {
	_, err := GetBankCardByID(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if req.BankName != "" {
		updates["bank_name"] = req.BankName
	}
	if req.CardNo != "" {
		updates["card_no"] = req.CardNo
	}

	if len(updates) == 0 {
		return nil
	}

	return UpdateBankCard(nil, id, updates)
}

func (s *Service) DeleteBankCard(id uint) error {
	_, err := GetBankCardByID(id)
	if err != nil {
		return err
	}
	return DeleteBankCard(nil, id)
}
