package main

import (
	"bytes"
	stdjson "encoding/json"
	"testing"
	"time"

	"github.com/anton2920/gofa/encoding/json"
)

var testUser = User{FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}

func TestJSONSerialize(t *testing.T) {
	var s json.Serializer
	s.Buf = make([]byte, 512)

	now := time.Now().Unix()
	tests := [...]User{
		testUser,
		{FirstName: `Quote"quote`, LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: now},
	}

	for _, test := range tests {
		expected, err := stdjson.Marshal(test)
		if err != nil {
			t.Fatalf("Failed to marshal user into JSON: %v", err)
		}

		s.Reset()
		PutUserJSON(&s, &test)

		if bytes.Compare(expected, s.Bytes()) != 0 {
			t.Errorf("Expected '%s', got '%s'", expected, s.String())
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := stdjson.Marshal(testUser)
		if err != nil {
			b.Errorf("Failed to marshal user into JSON: %v", err)
		}
	}
}

func BenchmarkJSONEncoder(b *testing.B) {
	var buf bytes.Buffer
	enc := stdjson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := enc.Encode(testUser); err != nil {
			b.Errorf("Failed to marshal user into JSON: %v", err)
		}
	}
}

func BenchmarkPutJSON(b *testing.B) {
	var s json.Serializer
	s.Buf = make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		PutUserJSON(&s, &testUser)
	}
}
