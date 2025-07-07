package main

import (
	"go/scanner"
	"go/token"

	"github.com/anton2920/gofa/debug"
)

type Lexer struct {
	Scanner scanner.Scanner
	FileSet *token.FileSet

	Position token.Position
	Token    token.Token
	Literal  string

	Error error
}

func (l *Lexer) Next() token.Token {
	var pos token.Pos

	pos, l.Token, l.Literal = l.Scanner.Scan()
	l.Position = l.FileSet.Position(pos)

	debug.Printf("%s\t%s\t%q\n", l.Position, l.Token, l.Literal)

	return l.Token
}
