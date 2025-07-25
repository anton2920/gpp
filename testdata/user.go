package main

import (
	"github.com/anton2920/gofa/database"
)

type Test int

//gofa:gpp json
type User struct {
	RecordHeader
	//ID    database.ID
	//Flags bits.Flags

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}

//gofa:gpp jso
type CompanyMember struct {
	database.RecordHeader

	// Company
	User

	RoleIDs []database.ID
}
