package main

import (
	"go/token"
)

func ParsePackage(l *Lexer, p *string) bool {
	if ParseToken(l, token.PACKAGE) {
		if ParseIdent(l, p) {
			return true
		}
	}
	return false
}
