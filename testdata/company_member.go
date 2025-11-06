package main

import "github.com/anton2920/gofa/database"

//gpp:generate: fill, encoding(json)
type CompanyMember struct {
	database.RecordHeader //gpp:fill: nop

	// Company
	User //gpp:fill: nop
	
	FirstName string
	LastName  string
	Position  string

	RoleIDs []database.ID //gpp:fill: nop

	CreatedAt int64 //gpp:fill: nop
}
