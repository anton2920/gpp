package main

import "github.com/anton2920/gofa/database"

//gpp:generate: encoding(json)
type User struct {
	RecordHeader database.RecordHeader

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}
