package main

import "github.com/anton2920/gofa/database"

//gpp:generate: encoding(json)
type CompanyMember struct {
	RecordHeader database.RecordHeader

	// Company
	User

	RoleIDs []database.ID

	CreatedAt int64
}
