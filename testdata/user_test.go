package main

import (
	"bytes"
	stdjson "encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/anton2920/gofa/encoding/json"
	_ "github.com/anton2920/gofa/time"
)

var (
	testUser     = User{FirstName: "FirstName", LastName: "LastName", Email: "user@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}
	testUserJSON = []byte(`{"ID":0,"Flags":0,"FirstName":"FirstName","LastName":"LastName","Email":"user@example.com","CreatedOn":1753717395}`)
)

func testUserCompare(self *User, other *User) bool {
	return reflect.DeepEqual(self, other)
}

func TestPutUserJSON(t *testing.T) {
	var s json.Serializer
	s.Buffer = make([]byte, 512)

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

func TestGetUserJSON(t *testing.T) {
	var d json.Deserializer

	tests := [...][]byte{
		testUserJSON,
		[]byte("{\"ID\":0,\"Flags\":0,\"FirstName\":\"Quote\\\"quote\",\"LastName\":\"LastName\",\"Email\":\"user@example.com\",\"CreatedOn\":1753718826}"),
	}

	for _, test := range tests {
		var expected User
		var actual User

		if err := stdjson.Unmarshal(test, &expected); err != nil {
			t.Fatalf("Failed to unmarshal JSON into user: %v", err)
		}

		d.Init(test)
		GetUserJSON(&d, &actual)
		if d.Error != nil {
			t.Errorf("Failed to get user from JSON: %v", d.Error)
		}

		if !testUserCompare(&expected, &actual) {
			t.Errorf("Expected '%#v', got '%#v'", expected, actual)
		}
	}
}

func BenchmarkMarshalUserJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := stdjson.Marshal(testUser)
		if err != nil {
			b.Errorf("Failed to marshal user into JSON: %v", err)
		}
	}
}

func BenchmarkEncodeUserJSON(b *testing.B) {
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

func BenchmarkPutUserJSON(b *testing.B) {
	var s json.Serializer
	s.Buffer = make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		PutUserJSON(&s, &testUser)
	}
}

func BenchmarkUnmarshalUserJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var user User
		if err := stdjson.Unmarshal(testUserJSON, &user); err != nil {
			b.Errorf("Failed to unmarshal JSON into user: %v", err)
		}
	}
}

func BenchmarkGetUserJSON(b *testing.B) {
	var d json.Deserializer
	d.Init(testUserJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user User

		d.Reset()
		if !GetUserJSON(&d, &user) {
			b.Errorf("Failed to get user from JSON: %v", d.Error)
		}
	}
}
