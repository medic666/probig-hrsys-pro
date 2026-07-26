package company

type CreateCompanyRequest struct {
	Name         string `json:"name" binding:"required"`
	CreditCode   string `json:"credit_code"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
	BankName     string `json:"bank_name"`
	BankAccount  string `json:"bank_account"`
}

type UpdateCompanyRequest struct {
	Name         string `json:"name"`
	CreditCode   string `json:"credit_code"`
	Address      string `json:"address"`
	ContactPhone string `json:"contact_phone"`
	BankName     string `json:"bank_name"`
	BankAccount  string `json:"bank_account"`
}
