package main

import "github.com/anton2920/gofa/database"

type UserType int32

const (
	UserTypeNone = UserType(iota)
	UserTypeAdmin
	UserTypeRegular
	UserTypeGuest
	UserTypeCount
)

//gpp:generate: fill(values), verify, encoding(json)
type User struct {
	database.RecordHeader //gpp:fill: nop

	FirstName string //gpp:verify: Func={NameValid(l, ?, MinUserNameLen, MaxUserNameLen)}
	LastName  string //gpp:verify: Func={NameValid(l, ?, MinUserNameLen, MaxUserNameLen)}
	Email     string //gpp:verify: MinLength=1, MaxLength=128, Func=EmailValid
	Password  string `json:"-"` //gpp:verify: MinLength=5, MaxLength=45

	UserType //gpp:fill: enum

	CreatedAt int64 //gpp:fill: nop
}
