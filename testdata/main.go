package main

import (
	"fmt"
	"time"

	"github.com/anton2920/gofa/encoding/json"
)

type RecordHeader struct {
	ID    int32
	Flags uint32
}

func main() {
	var s json.Serializer

	user := User{FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}

	s.Buffer = make([]byte, 1024)
	PutUserJSON(&s, &user)

	fmt.Printf("Result: %s\n", s)
}
