package main

import (
	"fmt"
	"go/scanner"
	"go/token"
)

type File struct {
	Name    string
	Package string

	Imports Imports
	Specs   []TypeSpec
}

func (p *Parser) File(f *token.File, file *File) bool {
	src, err := ReadEntireFile(f.Name())
	if err != nil {
		p.Error = fmt.Errorf("failed to read file %q: %v", f.Name(), err)
		return false
	}
	file.Name = f.Name()

	p.Scanner.Init(f, src, nil, scanner.ScanComments)
	p.Lexer.Tokens = p.Lexer.Tokens[:0]
	p.Lexer.Position = 0

	done := false
	for !done {
		switch p.Curr().GoToken {
		case token.PACKAGE:
			if !p.Package(&file.Package) {
				p.Error = fmt.Errorf("failed to parse package declaration: %v", p.Error)
				return false
			}
		case token.IMPORT:
			if !p.Imports(&file.Imports) {
				p.Error = fmt.Errorf("failed to parse imports: %v", p.Error)
				return false
			}
			continue
		case token.COMMENT, token.TYPE:
			/* If not type assertion (i.e. variable.(type)). */
			if p.Prev().GoToken != token.LPAREN {
				var specs []TypeSpec
				if !p.TypeDecl(&specs) {
					p.Error = fmt.Errorf("failed to parse type declarations: %v", p.Error)
					return false
				}
				file.Specs = append(file.Specs, specs...)
				continue
			}
		case token.EOF:
			done = true
		}
		p.Next()
	}

	return true
}
