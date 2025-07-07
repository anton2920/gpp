package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"

	"github.com/anton2920/gofa/log"
)

func ReadEntireFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}

	st, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stat: %v", err)
	}

	contents := make([]byte, int(st.Size()))
	if _, err := io.ReadFull(f, contents); err != nil {
		return nil, fmt.Errorf("failed to read all data from file: %v", err)
	}

	return contents, nil
}

func ParseToken(l *Lexer, tok token.Token) bool {
	if l.Error != nil {
		return false
	}

	if l.Next() == tok {
		return true
	}

	l.Error = fmt.Errorf("%s:%d:%d: expected %q, got %q (%q)", l.Position.Filename, l.Position.Line, l.Position.Column, tok, l.Token, l.Literal)
	return false
}

func ParseIdent(l *Lexer, ident *string) bool {
	if ParseToken(l, token.IDENT) {
		*ident = l.Literal
		return true
	}
	return false
}

func ParseString(l *Lexer, s *string) bool {
	if ParseToken(l, token.STRING) {
		*s = l.Literal
		return true
	}
	return false
}

func main() {
	const filename = "testdata/user.go"

	var l Lexer

	src, err := ReadEntireFile(filename)
	if err != nil {
		log.Fatalf("Failed to read entire file: %v", err)
	}

	l.FileSet = token.NewFileSet()
	file := l.FileSet.AddFile(filename, l.FileSet.Base(), len(src))
	l.Scanner.Init(file, src, nil, scanner.ScanComments)

	var buf bytes.Buffer

	done := false
	for !done {
		switch l.Next() {
		case token.EOF:
			done = true
		case token.COMMENT:
			var comment Comment
			if ParseGofaComment(&l, &comment) {
				var structure Struct
				if ParseStruct(&l, &structure) {
					for t := TargetNone + 1; t < TargetCount; t++ {
						if comment.Targets[t] {
							Target2Generator[t](&buf, &structure)
						}
					}
				}
			}
		}
	}

	if l.Error == nil {
		fmt.Printf("Generated: %q\n", buf.String())
	} else {
		fmt.Printf("Error = %v, error count = %d\n", l.Error, l.Scanner.ErrorCount)
	}
}
