package main

import (
	"bytes"
	"go/token"
)

type Type struct {
	Package string
	Name    string
	Args    []Type
	Literal TypeLit
}

type TypeParam struct {
	Name       string
	Constraint Type
}

type TypeSpec struct {
	Name   string
	Params []TypeParam
	Type   Type
	Alias  bool
}

func (t *Type) String() string {
	if t.Literal != nil {
		return t.Literal.String()
	}

	var buf bytes.Buffer

	if len(t.Package) > 0 {
		buf.WriteString(t.Package)
		buf.WriteRune('.')
	}
	buf.WriteString(t.Name)

	if len(t.Args) > 0 {
		buf.WriteRune('[')
		for i := 0; i < len(t.Args); i++ {
			if i > 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(t.Args[i].String())
		}
		buf.WriteRune(']')
	}

	return buf.String()
}

func ParseType(l *Lexer, t *Type) bool {
	if ParseTypeLit(l, &t.Literal) {
		return true
	}
	l.Error = nil

	var ident string
	if ParseIdent(l, &ident) {
		if ParseToken(l, token.PERIOD) {
			t.Package = ident
			if ParseIdent(l, &t.Name) {
				ParseTypeArgs(l, &t.Args)
				l.Error = nil
				return true
			}
		}
		l.Error = nil
		t.Name = ident

		ParseTypeArgs(l, &t.Args)
		l.Error = nil
		return true
	}

	return false
}

func ParseTypeArgs(l *Lexer, ts *[]Type) bool {
	if ParseToken(l, token.LBRACK) {
		for l.Peek().GoToken != token.RBRACK {
			var t Type
			if !ParseType(l, &t) {
				return false
			}
			*ts = append(*ts, t)
		}
		return true
	}
	return false
}

func ParseTypeParams(l *Lexer, tparams *[]TypeParam) bool {
	if ParseToken(l, token.LBRACK) {
		var idents []string
		if ParseIdentList(l, &idents) {
			var t Type
			if ParseType(l, &t) {
				if ParseToken(l, token.RBRACK) {
					for i := 0; i < len(idents); i++ {
						*tparams = append(*tparams, TypeParam{Name: idents[i], Constraint: t})
					}
					return true
				}
			}
		}
	}
	return false
}

func ParseTypeSpec(l *Lexer, ts *TypeSpec) bool {
	if ParseIdent(l, &ts.Name) {
		ParseTypeParams(l, &ts.Params)
		l.Error = nil

		if ParseToken(l, token.ASSIGN) {
			ts.Alias = true
		}
		l.Error = nil

		if ParseType(l, &ts.Type) {
			return true
		}
	}
	return false
}

func ParseTypeDecl(l *Lexer, tss *[]TypeSpec) bool {
	if ParseToken(l, token.TYPE) {
		if ParseToken(l, token.LPAREN) {
			for l.Peek().GoToken != token.RPAREN {
				var ts TypeSpec
				if !ParseTypeSpec(l, &ts) {
					return false
				}
				*tss = append(*tss, ts)
			}
		}
		l.Error = nil

		var ts TypeSpec
		if ParseTypeSpec(l, &ts) {
			*tss = append(*tss, ts)
			return true
		}
	}

	return false
}
