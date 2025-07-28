package main

import (
	"bytes"
	"go/token"

	"github.com/anton2920/gofa/go/lexer"
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
	Comment *Comment
	Name    string
	Params  []TypeParam
	Type    Type
	Alias   bool
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

func ParseType(l *lexer.Lexer, t *Type) bool {
	if ParseTypeLit(l, &t.Literal) {
		return true
	}
	l.Error = nil

	var ident string
	if l.ParseIdent(&ident) {
		if l.ParseToken(token.PERIOD) {
			t.Package = ident
			ReferencedPackages[t.Package] = struct{}{}

			if l.ParseIdent(&t.Name) {
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

func ParseTypeArgs(l *lexer.Lexer, ts *[]Type) bool {
	if l.ParseToken(token.LBRACK) {
		for l.Curr().GoToken != token.RBRACK {
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

func ParseTypeParams(l *lexer.Lexer, tparams *[]TypeParam) bool {
	if l.ParseToken(token.LBRACK) {
		var idents []string
		if l.ParseIdentList(&idents) {
			var t Type
			if ParseType(l, &t) {
				if l.ParseToken(token.RBRACK) {
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

func ParseTypeSpec(l *lexer.Lexer, ts *TypeSpec) bool {
	if l.ParseIdent(&ts.Name) {
		ParseTypeParams(l, &ts.Params)
		l.Error = nil

		if l.ParseToken(token.ASSIGN) {
			ts.Alias = true
		}
		l.Error = nil

		if ParseType(l, &ts.Type) {
			return true
		}
	}
	return false
}

func ParseTypeDecl(l *lexer.Lexer, tss *[]TypeSpec) bool {
	if l.ParseToken(token.TYPE) {
		if l.ParseToken(token.LPAREN) {
			for l.Curr().GoToken != token.RPAREN {
				var ts TypeSpec
				if !ParseTypeSpec(l, &ts) {
					return false
				}
				if !l.ParseToken(token.SEMICOLON) {
					return false
				}
				*tss = append(*tss, ts)
			}
			return true
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
