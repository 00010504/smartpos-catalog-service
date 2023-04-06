package models

type Brand struct {
	Id        string `db:"id"`
	Name      string `db:"name"`
	CompanyId string `db:"company_id"`
}
