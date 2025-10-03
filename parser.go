package main

import (
	"fmt"
	"go/token"
	"strconv"
)

type Parser struct {
	Lexer

	Error error
}

func (p *Parser) Token(expectedTok token.Token) bool {
	if p.Error != nil {
		return false
	}

	if expectedTok != token.COMMENT {
		for p.Curr().GoToken == token.COMMENT {
			p.Next()
		}
	}

	tok := p.Curr()
	if tok.GoToken == expectedTok {
		p.Next()
		return true
	}

	p.Error = fmt.Errorf("%s:%d:%d: expected %q, got %q (%q)", tok.Position.Filename, tok.Position.Line, tok.Position.Column, expectedTok, tok.GoToken, tok.Literal)
	return false
}

func (p *Parser) Ident(ident *string) bool {
	if p.Token(token.IDENT) {
		*ident = p.Prev().Literal
		return true
	}
	return false
}

func (p *Parser) IdentList(idents *[]string) bool {
	var ident string

	for p.Ident(&ident) {
		*idents = append(*idents, ident)
		if !p.Token(token.COMMA) {
			p.Error = nil
			return true
		}
	}

	return len(*idents) != 0
}

func (p *Parser) IntLit(n *int) bool {
	if p.Token(token.INT) {
		var err error
		*n, err = strconv.Atoi(p.Prev().Literal)
		if err != nil {
			p.Error = fmt.Errorf("failed to parse int value: %v", err)
		}
		return err == nil
	}
	return false
}

func (p *Parser) StringLit(s *string) bool {
	if p.Token(token.STRING) {
		*s = p.Prev().Literal
		if ((*s)[0] == '"') || ((*s)[0] == '`') {
			*s = (*s)[1:]
		}
		if ((*s)[len(*s)-1] == '"') || ((*s)[len(*s)-1] == '`') {
			*s = (*s)[:len(*s)-1]
		}
		return true
	}
	return false
}
