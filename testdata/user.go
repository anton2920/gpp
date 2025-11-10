package main

import (
	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/l10n"
)

type UserType int32

const (
	UserTypeNone = UserType(iota)
	UserTypeAdmin
	UserTypeRegular
	UserTypeGuest
	UserTypeCount
)

//gpp:generate: encoding(wire)
type User struct {
	RecordHeader database.RecordHeader //gpp:fill: nop

	FirstName string //gpp:verify: MinLength=1, MaxLength=45, Func=NameValid
	LastName  string //gpp:verify: MinLength=1, MaxLength=45, Func=NameValid
	Email     string //gpp:verify: MinLength=1, MaxLength=128, Func=EmailValid
	Password  string `json:"-"` //gpp:verify: MinLength=5, MaxLength=45

	// UserType //gpp:fill: enum

	CreatedAt int64 //gpp:fill: nop
}

func NameValid(l l10n.Language, name string) error {
	return nil
}

func EmailValid(l l10n.Language, email string) error {
	return nil
}

//gpp:generate: fill(values), verify, encoding(json,wire)
