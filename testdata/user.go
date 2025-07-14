package main

import (
	"github.com/anton2920/gofa/bits"
	"github.com/anton2920/gofa/database"
)

//gofa:gpp json
type User struct {
	// database.RecordHeader
	ID    database.ID
	Flags bits.Flags

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}

//gofa:gpp json
type CompanyMember struct {
	database.RecordHeader

	// Company
	User

	RoleIDs []database.ID
}
