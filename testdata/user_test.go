package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

var testUser = User{FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}

func TestJSONSerialize(t *testing.T) {
	now := time.Now().Unix()
	tests := [...]User{
		testUser,
		User{FirstName: `Quote"quote`, LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: now},
	}

	for _, test := range tests {
		expected, err := json.Marshal(test)
		if err != nil {
			t.Fatalf("Failed to marshal user into JSON: %v", err)
		}

		actual := make([]byte, 0, 1024)
		JSONSerializeUser(&actual, &test)

		if bytes.Compare(expected, actual) != 0 {
			t.Errorf("Expected '%s', got '%s'", expected, actual)
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(testUser)
		if err != nil {
			b.Errorf("Failed to marshal user into JSON: %v", err)
		}
	}
}

func BenchmarkJSONEncoder(b *testing.B) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := enc.Encode(testUser); err != nil {
			b.Errorf("Failed to marshal user into JSON: %v", err)
		}
	}
}

func BenchmarkJSONSerialize(b *testing.B) {
	buffer := make([]byte, 0, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer = buffer[:0]
		JSONSerializeUser(&buffer, &testUser)
	}
}
