package main

import (
	"bytes"
	"go/token"
)

type Type struct {
	Package string
	Name    string
	Kind    TypeKind
}

type TypeArg struct {
	Name string
	Type Type
}

type TypeKind int

const (
	TypeKindNone = TypeKind(iota)
	TypeKindInt
	TypeKindInt32
	TypeKindInt64
	TypeKindPointer
	TypeKindString
	TypeKindSlice
	TypeKindUnknown
)

var TypeKind2Name = [...]string{
	TypeKindInt:     "int",
	TypeKindInt32:   "int32",
	TypeKindInt64:   "int64",
	TypeKindPointer: "*T",
	TypeKindString:  "string",
	TypeKindSlice:   "[]T",
}

func (t *Type) String() string {
	var buf bytes.Buffer

	switch t.Kind {
	case TypeKindPointer:
		buf.WriteRune('*')
	case TypeKindSlice:
		buf.WriteString(`[]`)
	}
	if len(t.Package) > 0 {
		buf.WriteString(t.Package)
		buf.WriteRune('.')
	}
	buf.WriteString(t.Name)

	return buf.String()
}

func ParseType(l *Lexer, t *Type) bool {
	var pointer bool
	var name string

	if l.Error != nil {
		return false
	}

	if ParseToken(l, token.LBRACK) {
		var n int
		ParseInt(l, &n)
		l.Error = nil

		if !ParseToken(l, token.RBRACK) {
			return false
		}

		if ParseType(l, t) {
			t.Kind = TypeKindSlice
			return true
		}
		return false
	}

	if ParseToken(l, token.MUL) {
		pointer = true
	}
	l.Error = nil

	if ParseIdent(l, &name) {
		t.Name = name
		if ParseToken(l, token.PERIOD) {
			if !ParseIdent(l, &name) {
				return false
			}
			t.Package = t.Name
			t.Name = name
		}
		l.Error = nil

		if pointer {
			t.Kind = TypeKindPointer
		} else {
			switch t.Name {
			default:
				/* TODO(anton2920): expand list of known types. */
				t.Kind = TypeKindUnknown
			case "int", "uint":
				t.Kind = TypeKindInt
			case "int32", "uint32", "ID":
				t.Kind = TypeKindInt32
			case "int64", "uint64":
				t.Kind = TypeKindInt64
			case "string":
				t.Kind = TypeKindString
			}
		}
	}

	return t.Kind != TypeKindNone
}
