package main

import "github.com/anton2920/gofa/database"

//gofa:generate serial, json, db(psql), db(gofa)
type User struct {
	FirstName string `json:omitempty`
	LastName  string
	Email     string
	Password  string
	CreatedOn int64
}
