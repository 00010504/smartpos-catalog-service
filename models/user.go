package models

import "database/sql"

type User struct {
	Id          string `json:"id"`
	UserTypeId  string `json:"user_type_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Image       string `json:"image"`
}

type NullShortUser struct {
	FirstName sql.NullString
	LastName  sql.NullString
	ID        sql.NullString
	Image     sql.NullString
}
