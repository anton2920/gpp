package main

import (
	"github.com/anton2920/gofa/database"
)

type Int int

//gpp:generate json
type User struct {
	database.RecordHeader

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}

//gpp:generate json
type CompanyMember struct {
	database.RecordHeader

	// Company
	User

	RoleIDs []database.ID
}
