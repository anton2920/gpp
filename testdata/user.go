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

type CompanyMember struct {
	database.RecordHeader

	// Company
	User

	RoleIDs []database.ID
}
