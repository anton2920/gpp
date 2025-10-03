package main

import (
	"fmt"
	"go/token"
	"strconv"
)

func ParseToken(l *Lexer, expectedTok token.Token) bool {
	if l.Error != nil {
		return false
	}

	if expectedTok != token.COMMENT {
		for l.Curr().GoToken == token.COMMENT {
			l.Next()
		}
	}

	tok := l.Curr()
	if tok.GoToken == expectedTok {
		l.Next()
		return true
	}

	l.Error = fmt.Errorf("%s:%d:%d: expected %q, got %q (%q)", tok.Position.Filename, tok.Position.Line, tok.Position.Column, expectedTok, tok.GoToken, tok.Literal)
	return false
}

func ParseIdent(l *Lexer, ident *string) bool {
	if ParseToken(l, token.IDENT) {
		*ident = l.Prev().Literal
		return true
	}
	return false
}

func ParseIdentList(l *Lexer, idents *[]string) bool {
	var ident string

	for ParseIdent(l, &ident) {
		*idents = append(*idents, ident)
		if !ParseToken(l, token.COMMA) {
			l.Error = nil
			return true
		}
	}

	return len(*idents) != 0
}

func ParseIntLit(l *Lexer, n *int) bool {
	if ParseToken(l, token.INT) {
		var err error
		*n, err = strconv.Atoi(l.Prev().Literal)
		if err != nil {
			l.Error = fmt.Errorf("failed to parse int value: %v", err)
		}
		return err == nil
	}
	return false
}

func ParseStringLit(l *Lexer, s *string) bool {
	if ParseToken(l, token.STRING) {
		*s = l.Prev().Literal
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
