package main

import (
	"go/token"
)

func (p *Parser) Package(pack *string) bool {
	if p.Token(token.PACKAGE) {
		if p.Ident(pack) {
			return true
		}
	}
	return false
}
