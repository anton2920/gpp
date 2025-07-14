package main

import (
	"fmt"
	"time"

	"github.com/anton2920/gofa/encoding/json"
)

func main() {
	var s json.Serializer

	user := User{FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}

	s.Buf = make([]byte, 64*1024)
	PutUserJSON(&s, &user)

	fmt.Printf("Result: %s\n", s)
}
