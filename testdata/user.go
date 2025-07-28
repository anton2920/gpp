package main

import "github.com/anton2920/gofa/database"

//gpp:generate json
type User struct {
	database.RecordHeader

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}
