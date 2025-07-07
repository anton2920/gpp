package main

import "go/token"

type StructField struct {
	Name string
	Type Type
	Tag  string
}

type Struct struct {
	Name   string
	Fields []StructField
}

func ParseStructField(l *Lexer, f *StructField) bool {
	var ret = true

	ret = ret && ParseIdent(l, &f.Name)
	ret = ret && ParseType(l, &f.Type)
	ret = ret && ParseString(l, &f.Tag)

	return ret
}

func ParseStruct(l *Lexer, s *Struct) bool {
	if ParseToken(l, token.TYPE) {
		if ParseIdent(l, &s.Name) {
			if ParseToken(l, token.STRUCT) {
				if ParseToken(l, token.LBRACE) {
					done := false
					for !done {
						var field StructField
						if !ParseStructField(l, &field) {
							return false
						}
						s.Fields = append(s.Fields, field)

						if l.Next() == token.RBRACE {
							break
						}
					}
					return true
				}
			}
		}
	}

	return false
}
