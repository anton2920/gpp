package main

import "github.com/anton2920/gofa/database"

//gpp:generate json
type CompanyMember struct {
	database.RecordHeader

	// Company
	User User

	RoleIDs []database.ID
}
