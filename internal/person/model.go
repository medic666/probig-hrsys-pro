package person

type CreatePersonRequest struct {
	Name            string `json:"name" binding:"required"`
	IDCard          string `json:"id_card"`
	Gender          *int8  `json:"gender"`
	Birthday        string `json:"birthday"`
	Nation          string `json:"nation"`
	NativePlace     string `json:"native_place"`
	Address         string `json:"address"`
	PoliticalStatus string `json:"political_status"`
	MaritalStatus   *int8  `json:"marital_status"`
	Alias           string `json:"alias"`
}

type UpdatePersonRequest struct {
	Name            string `json:"name"`
	IDCard          string `json:"id_card"`
	Gender          *int8  `json:"gender"`
	Birthday        string `json:"birthday"`
	Nation          string `json:"nation"`
	NativePlace     string `json:"native_place"`
	Address         string `json:"address"`
	PoliticalStatus string `json:"political_status"`
	MaritalStatus   *int8  `json:"marital_status"`
	Alias           string `json:"alias"`
}

type CreatePhoneRequest struct {
	Phone     string `json:"phone" binding:"required"`
	PhoneType string `json:"phone_type"`
}

type UpdatePhoneRequest struct {
	Phone     string `json:"phone"`
	PhoneType string `json:"phone_type"`
}

type CreateEmailRequest struct {
	Email     string `json:"email" binding:"required"`
	EmailType string `json:"email_type"`
}

type UpdateEmailRequest struct {
	Email     string `json:"email"`
	EmailType string `json:"email_type"`
}

type CreateBankCardRequest struct {
	BankName string `json:"bank_name" binding:"required"`
	CardNo   string `json:"card_no" binding:"required"`
}

type UpdateBankCardRequest struct {
	BankName string `json:"bank_name"`
	CardNo   string `json:"card_no"`
}

type PersonListItem struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	IDCard           string `json:"id_card"`
	Gender           int8   `json:"gender"`
	Aliaa            string `json:"alias"`
	AttendanceGroup  string `json:"attendance_group"`
	EmploymentStatus string `json:"employment_status"`
	CreatedAt        string `json:"created_at"`
}

type PersonDetailResponse struct {
	Person    interface{} `json:"person"`
	Phones    interface{} `json:"phones"`
	Emails    interface{} `json:"emails"`
	BankCards interface{} `json:"bank_cards"`
}
