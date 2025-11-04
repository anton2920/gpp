package main

import (
	"fmt"
	"time"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/encoding/json"
)

func main() {
	{
		var s json.Serializer
		user := User{FirstName: `Quote"quote`, LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedAt: time.Now().Unix()}

		s.Buffer = make([]byte, 1024)
		SerializeUserJSON(&s, &user)

		fmt.Printf("Result 1: %s\n", s.Bytes())
	}
	/*
		{
			var d json.Deserializer
			var user User

			d.Buffer = []byte(`{"ID":0,"Flags":0,"FirstName":"FirstName","LastName":"LastName","Email":"user@example.com","CreatedOn":1753717395}`)
			DeserializeUserJSON(&d, &user)

			if d.Error == nil {
				fmt.Printf("Result 2: %v\n", user)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to get user from JSON: %v\n", d.Error)
			}
		}
	*/
	{
		var s json.Serializer
		member := CompanyMember{RecordHeader: database.RecordHeader{ID: 123, Flags: 456}, User: User{RecordHeader: database.RecordHeader{ID: 789, Flags: 12}, FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedAt: time.Now().Unix()}, RoleIDs: []database.ID{1, 2, 3, 4, 8}}

		s.Buffer = make([]byte, 1024)
		SerializeCompanyMemberJSON(&s, &member)

		fmt.Printf("Result 3: %s\n", s.Bytes())
	}
	/*
		{
			var d json.Deserializer
			var member CompanyMember

			d.Init([]byte(`{"ID":123,"Flags":456,"User":{"ID":789,"Flags":12,"FirstName":"FirstName","LastName":"LastName","Email":"user@example.com","CreatedOn":1753724928},"RoleIDs":[1,2,3,4,8]}`))
			DeserializeCompanyMemberJSON(&d, &member)

			if d.Error == nil {
				fmt.Printf("Result 4: %v\n", member)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to get company member from JSON: %v\n", d.Error)
			}
		}
	*/
}
