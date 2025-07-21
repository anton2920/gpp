package main

import (
	"fmt"

	"github.com/anton2920/gofa/encoding/json"
)

func main() {
	var s json.Serializer

	s.Buffer = make([]byte, 1024)

	s.PutObjectBegin()
	s.PutKey("Test")
	s.PutString("test")
	s.PutKey("Array")
	s.PutArrayBegin()
	for i := 0; i < 5; i++ {
		s.PutInt(i)
	}
	s.PutArrayEnd()
	s.PutObjectEnd()

	fmt.Printf("Result: %s\n", s)
}
