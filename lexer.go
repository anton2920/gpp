package main

import (
	"go/scanner"
	"go/token"
)

type Token struct {
	token.Position

	GoToken token.Token
	Literal string
}

type Lexer struct {
	scanner.Scanner
	*token.FileSet

	Tokens   []Token
	Position int
}

func (l *Lexer) Curr() Token {
	if l.Position == len(l.Tokens) {
		pos, tok, lit := l.Scanner.Scan()
		l.Tokens = append(l.Tokens, Token{Position: l.FileSet.Position(pos), GoToken: tok, Literal: lit})
		//debug.Printf("[lexer]: %s\t%s\t%q", l.Tokens[l.Position].Position, l.Tokens[l.Position].GoToken, l.Tokens[l.Position].Literal)
	}
	return l.Tokens[l.Position]
}

func (l *Lexer) Next() Token {
	l.Position++
	tok := l.Curr()
	return tok
}

func (l *Lexer) Prev() Token {
	if l.Position == 0 {
		panic("no previous token")
	}
	return l.Tokens[l.Position-1]
}
