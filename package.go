package main

import (
	"go/token"
)

type Package struct {
	ImportPath string
	Files      []File
}

func (p *Parser) Package(pack *string) bool {
	if p.Token(token.PACKAGE) {
		if p.Ident(pack) {
			return true
		}
	}
	return false
}
