package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"strconv"

	"github.com/anton2920/gofa/log"
)

const (
	DefaultCap = 16
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

func ParseToken(l *Lexer, expectedTok token.Token) bool {
	if l.Error != nil {
		return false
	}

	if expectedTok != token.COMMENT {
		for l.Peek().GoToken == token.COMMENT {
			l.Next()
		}
	}

	tok := l.Peek()
	if tok.GoToken == expectedTok {
		l.Next()
		return true
	}

	l.Error = fmt.Errorf("%s:%d:%d: expected %q, got %q (%q)", tok.Position.Filename, tok.Position.Line, tok.Position.Column, expectedTok, tok.GoToken, tok.Literal)
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

func ParseIdent(l *Lexer, ident *string) bool {
	if ParseToken(l, token.IDENT) {
		*ident = l.Prev().Literal
		return true
	}
	return false
}

func ParseInt(l *Lexer, n *int) bool {
	if ParseToken(l, token.INT) {
		var err error
		*n, err = strconv.Atoi(l.Prev().Literal)
		if err != nil {
			l.Error = fmt.Errorf("failed to parse int: %v", err)
		}
		return err == nil
	}
	return false
}

func ParseString(l *Lexer, s *string) bool {
	if ParseToken(l, token.STRING) {
		*s = l.Prev().Literal
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
		switch l.Peek().GoToken {
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
				continue
			}
		}
		l.Next()
	}

	if l.Error == nil {
		fmt.Printf("Generated: %q\n", buf.String())
	} else {
		fmt.Printf("Error = %v, error count = %d\n", l.Error, l.Scanner.ErrorCount)
	}
}
