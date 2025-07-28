package main

import (
	"go/token"

	"github.com/anton2920/gofa/go/lexer"
)

func ParsePackage(l *lexer.Lexer, p *string) bool {
	if l.ParseToken(token.PACKAGE) {
		if l.ParseIdent(p) {
			return true
		}
	}
	return false
}
