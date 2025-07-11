package main

import "github.com/anton2920/gofa/database"

//gofa:gpp json
type User[T any] struct {
	database.RecordHeader
	//ID    database.ID
	//Flags uint32

	FirstName string `json:omitempty`
	LastName  string
	Email     string
	Password  string
	CreatedOn int64
}

//gofa:gpp json
type CompanyMember struct {
	database.RecordHeader

	Company
	User

	RoleIDs []database.ID
}
