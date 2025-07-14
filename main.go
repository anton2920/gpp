package main

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/anton2920/gofa/strings"
)

const GOFA = "github.com/anton2920/gofa/"

const Stdin = "<stdin>"

func ReadEntireStdin() ([]byte, error) {
	var buf bytes.Buffer

	if _, err := io.Copy(&buf, os.Stdin); err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %v", err)
	}

	return buf.Bytes(), nil
}

func ReadEntireFile(filename string) ([]byte, error) {
	if filename == Stdin {
		return ReadEntireStdin()
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %v", filename, err)
	}
	defer f.Close()

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

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: gpp [file ...]")
	os.Exit(1)
}

func Errorf(format string, args ...interface{}) {
	if format[len(format)-1] != '\n' {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}

func main() {
	var files []string

	fileSet := token.NewFileSet()
	sources := make(map[int][]byte)

	paths := os.Args[1:]
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		st, err := os.Stat(path)
		if err != nil {
			Fatalf("Failed to stat %q: %v", path, err)
		}

		if !st.IsDir() {
			files = append(files, path)
		} else {
			dir, err := os.Open(path)
			if err != nil {
				Fatalf("Failed to open directory %q: %v", path, err)
			}
			names, err := dir.Readdirnames(-1)
			if err != nil {
				Fatalf("Failed to read names of directory %q entries: %v", path, err)
			}
			for j := 0; j < len(names); j++ {
				name := names[j]
				if strings.EndsWith(name, ".go") {
					files = append(files, filepath.Join(path, name))
				}
			}
			dir.Close()
		}
	}

	if len(files) == 0 {
		files = append(files, Stdin)
	}
	for i := 0; i < len(files); i++ {
		file := files[i]

		src, err := ReadEntireFile(file)
		if err != nil {
			Fatalf("Failed to read entire file %q: %v", file, err)
		}

		base := fileSet.Base()
		fileSet.AddFile(file, base, len(src))
		sources[base] = src
	}

	fileSet.Iterate(func(f *token.File) bool {
		var g Generator
		var l Lexer

		l.FileSet = fileSet
		l.Scanner.Init(f, sources[f.Base()], nil, scanner.ScanComments)

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
						if l.Error != nil {
							Errorf("Failed to parse structure: %v", l.Error)
							return false
						}

						for i := 0; i < len(comment.Formats); i++ {
							format := comment.Formats[i]
							format.Generate(&g, &structure)
						}
					}
					continue
				}
				l.Error = nil
			case token.PACKAGE:
				if !ParsePackage(&l, &g.Package) {
					Errorf("Failed to parse package: %v", l.Error)
					return false
				}
			}
			l.Next()
		}

		if g.ShouldDump() {
			name := GeneratedName(f.Name())
			file, err := os.Create(name)
			if err != nil {
				Errorf("Failed to create generated file %q: %v", name, err)
			}
			defer file.Close()

			g.Dump(file)

			fmt.Println(f.Name())
		}
		return true
	})
}
