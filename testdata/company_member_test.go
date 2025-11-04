package main

import (
	"bytes"
	stdjson "encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/encoding/json"
)

var (
	testCompanyMember     = CompanyMember{RecordHeader: database.RecordHeader{ID: 123, Flags: 456}, User: User{RecordHeader: database.RecordHeader{ID: 789, Flags: 12}, FirstName: "FirstName", LastName: "LastName", Email: "member@example.com", Password: "qwerty", CreatedOn: time.Now().Unix()}, RoleIDs: []database.ID{1, 2, 3, 4, 8}}
	testCompanyMemberJSON = []byte(`{"ID":123,"Flags":456,"User":{"ID":789,"Flags":12,"FirstName":"FirstName","LastName":"LastName","Email":"member@example.com","CreatedOn":1753724928},"RoleIDs":[1,2,3,4,8]}`)
)

func testCompanyMemberCompare(self *CompanyMember, other *CompanyMember) bool {
	return reflect.DeepEqual(self, other)
}

func TestJSONPutCompanyMemberJSON(t *testing.T) {
	var s json.Serializer
	s.Buffer = make([]byte, 512)

	tests := [...]CompanyMember{
		testCompanyMember,
	}

	for _, test := range tests {
		expected, err := stdjson.Marshal(test)
		if err != nil {
			t.Fatalf("Failed to marshal member into JSON: %v", err)
		}

		s.Reset()
		SerializeCompanyMemberJSON(&s, &test)

		if bytes.Compare(expected, s.Bytes()) != 0 {
			t.Errorf("Expected '%s', got '%s'", expected, s.Bytes())
		}
	}
}

/*
func TestJSONGetCompanyMemberJSON(t *testing.T) {
	var d json.Deserializer

	tests := [...][]byte{
		testCompanyMemberJSON,
	}

	for _, test := range tests {
		var expected CompanyMember
		var actual CompanyMember

		if err := stdjson.Unmarshal(test, &expected); err != nil {
			t.Fatalf("Failed to unmarshal JSON into member: %v", err)
		}

		d.Init(test)
		GetCompanyMemberJSON(&d, &actual)
		if d.Error != nil {
			t.Errorf("Failed to get member from JSON: %v", d.Error)
		}

		if !testCompanyMemberCompare(&expected, &actual) {
			t.Errorf("Expected '%#v', got '%#v'", expected, actual)
		}
	}
}
*/

func BenchmarkMarshalCompanyMemberJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := stdjson.Marshal(testCompanyMember)
		if err != nil {
			b.Errorf("Failed to marshal member into JSON: %v", err)
		}
	}
}

func BenchmarkEncoderCompanyMemberJSON(b *testing.B) {
	var buf bytes.Buffer
	enc := stdjson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := enc.Encode(testCompanyMember); err != nil {
			b.Errorf("Failed to marshal member into JSON: %v", err)
		}
	}
}

func BenchmarkPutCompanyMemberJSON(b *testing.B) {
	var s json.Serializer
	s.Buffer = make([]byte, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Reset()
		SerializeCompanyMemberJSON(&s, &testCompanyMember)
	}
}

func BenchmarkUnmarshalCompanyMemberJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var member CompanyMember
		if err := stdjson.Unmarshal(testCompanyMemberJSON, &member); err != nil {
			b.Errorf("Failed to unmarshal JSON into member: %v", err)
		}
	}
}

/*
func BenchmarkGetCompanyMemberJSON(b *testing.B) {
	var d json.Deserializer
	d.Init(testCompanyMemberJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var member CompanyMember

		d.Reset()
		if !GetCompanyMemberJSON(&d, &member) {
			b.Errorf("Failed to get member from JSON: %v", d.Error)
		}
	}
}
*/
