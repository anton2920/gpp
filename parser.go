package main

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/anton2920/gofa/strings"
)

type Parser struct {
	Lexer

	Packages           map[string]Package
	ReferencedPackages map[string]struct{}

	Error error
}

func NewParser(fs *token.FileSet) Parser {
	var p Parser

	p.FileSet = fs
	p.Packages = make(map[string]Package)
	p.ReferencedPackages = make(map[string]struct{})

	return p
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

func (p *Parser) FindTypeLit(is Imports, typePkg string, typeName string) ForeignTypeLit {
	pkgName := is.PackageName(typePkg)

	pkg := p.Packages[pkgName]
	for _, file := range pkg.Files {
		for _, spec := range file.Specs {
			if typeName == spec.Name {
				if spec.Type.Literal == nil {
					p.FindTypeLit(is, strings.Or(spec.Type.Package, pkgName), spec.Type.Name)
				}
				return ForeignTypeLit{ImportPath: pkg.ImportPath, TypeLit: spec.Type.Literal}
			}
		}
	}

	return ForeignTypeLit{}
}

func (p *Parser) FindFunction(is Imports, funcPkg string, funcName string) *Func {
	pkgName := is.PackageName(funcPkg)

	for _, file := range p.Packages[pkgName].Files {
		for i := 0; i < len(file.Funcs); i++ {
			fn := &file.Funcs[i]
			if fn.Name == funcName {
				return fn
			}
		}
	}

	return nil
}
