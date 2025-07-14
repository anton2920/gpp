package main

import (
	"fmt"
	"go/scanner"
	"go/token"
)

type Token struct {
	Position token.Position
	GoToken  token.Token
	Literal  string
}

type Lexer struct {
	Scanner scanner.Scanner
	FileSet *token.FileSet

	Tokens   []Token
	Position int

	Error error
}

func (t *Token) String() string {
	return fmt.Sprintf("%s\t%s\t%q\n", t.Position, t.GoToken, t.Literal)
}

func (l *Lexer) Next() Token {
	tok := l.Peek()
	l.Position++
	return tok
}

func (l *Lexer) Peek() Token {
	if l.Position == len(l.Tokens) {
		pos, tok, lit := l.Scanner.Scan()
		l.Tokens = append(l.Tokens, Token{Position: l.FileSet.Position(pos), GoToken: tok, Literal: lit})
		// debug.Printf("[lexer]: %s", l.Tokens[l.Position])
	}
	return l.Tokens[l.Position]
}

func (l *Lexer) Prev() Token {
	if l.Position == 0 {
		panic("no previous token")
	}
	return l.Tokens[l.Position-1]
}
