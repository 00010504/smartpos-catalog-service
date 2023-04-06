package models

type Company struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	DeletedAt string `json:"deleted_at"`
	CreatedBy string `json:"created_by"`
}

type Shop struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Company Company `json:"company"`
	// CompanyId string  `json:"company_id"`
}
