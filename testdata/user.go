package main

import "github.com/anton2920/gofa/database"

//gpp:generate: fill(values), verify, encoding(json)
type User struct {
	database.RecordHeader //gpp:fill: nop

	FirstName string //gpp:verify: Func=NameValid
	LastName  string //gpp:verify: Func=NameValid
	Email     string //gpp:verify: MinLength=1, MaxLength=128, Func=EmailValid
	Password  string `json:"-"` //gpp:verify: MinLength=5, MaxLength=45

	CreatedAt int64 //gpp:fill: nop
}
