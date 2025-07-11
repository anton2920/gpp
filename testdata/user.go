package main

import (
	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/strings"
)

//gofa:gpp json
type User struct {
	// database.RecordHeader
	ID    database.ID
	Flags uint32

	FirstName string
	LastName  string
	Email     string
	Password  string `json:"-"`
	CreatedOn int64
}

func JSONSerializeInt32(buffer *[]byte, x int32) {
	tmp := make([]byte, 30)
	n := slices.PutInt(tmp, int(x))
	*buffer = append(*buffer, tmp[:n]...)
}

func JSONSerializeUint32(buffer *[]byte, x uint32) {
	tmp := make([]byte, 30)
	n := slices.PutInt(tmp, int(x))
	*buffer = append(*buffer, tmp[:n]...)
}

func JSONSerializeInt64(buffer *[]byte, x int64) {
	tmp := make([]byte, 30)
	n := slices.PutInt(tmp, int(x))
	*buffer = append(*buffer, tmp[:n]...)
}

func JSONSerializeString(buffer *[]byte, s string) {
	*buffer = append(*buffer, `"`...)
	for {
		quote := strings.FindChar(s, '"')
		if quote == -1 {
			*buffer = append(*buffer, s...)
			break
		}
		*buffer = append(*buffer, s[:quote]...)
		*buffer = append(*buffer, `\"`...)
		if quote == len(s)-1 {
			break
		}
		s = s[quote+1:]
	}
	*buffer = append(*buffer, `"`...)
}

func JSONSerializeUser(buffer *[]byte, user *User) {
	*buffer = append(*buffer, `{`...)
	*buffer = append(*buffer, `"ID":`...)
	JSONSerializeInt32(buffer, int32(user.ID))
	*buffer = append(*buffer, `,`...)
	*buffer = append(*buffer, `"Flags":`...)
	JSONSerializeUint32(buffer, user.Flags)
	*buffer = append(*buffer, `,`...)
	*buffer = append(*buffer, `"FirstName":`...)
	JSONSerializeString(buffer, user.FirstName)
	*buffer = append(*buffer, `,`...)
	*buffer = append(*buffer, `"LastName":`...)
	JSONSerializeString(buffer, user.LastName)
	*buffer = append(*buffer, `,`...)
	*buffer = append(*buffer, `"Email":`...)
	JSONSerializeString(buffer, user.Email)
	*buffer = append(*buffer, `,`...)
	*buffer = append(*buffer, `"CreatedOn":`...)
	JSONSerializeInt64(buffer, user.CreatedOn)
	*buffer = append(*buffer, `}`...)
}

//gofa:gpp json
type CompanyMember struct {
	database.RecordHeader

	// Company
	User

	RoleIDs []database.ID
}
