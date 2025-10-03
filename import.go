package main

import (
	"go/token"
	"path/filepath"

	"github.com/anton2920/gofa/strings"
)

type Import struct {
	QualifiedName    string
	Path             string
	WithoutQualifier bool
}

type Imports []Import

func (is Imports) Len() int { return len(is) }
func (is Imports) Less(i int, j int) bool {
	if (strings.FindChar(is[i].Path, '/') == -1) && (strings.FindChar(is[j].Path, '/') != -1) {
		return true
	} else if (strings.FindChar(is[i].Path, '/') != -1) && (strings.FindChar(is[j].Path, '/') == -1) {
		return false
	}
	return is[i].Path < is[j].Path
}
func (is Imports) Swap(i int, j int) { is[i], is[j] = is[j], is[i] }

func (is Imports) PackagePath(pkg string) string {
	for _, i := range is {
		if i.QualifiedName == pkg {
			pkg = filepath.Base(i.Path)
			break
		}
	}
	return pkg
}

func (p *Parser) Import(i *Import) bool {
	p.Ident(&i.QualifiedName)
	p.Error = nil

	if p.Token(token.PERIOD) {
		i.WithoutQualifier = true
	}
	p.Error = nil

	return p.StringLit(&i.Path)
}

func (p *Parser) Imports(is *Imports) bool {
	if p.Token(token.IMPORT) {
		if p.Token(token.LPAREN) {
			for p.Curr().GoToken != token.RPAREN {
				var i Import
				if !p.Import(&i) {
					return false
				}
				if !p.Token(token.SEMICOLON) {
					return false
				}
				*is = append(*is, i)
			}
			return true
		}
		p.Error = nil

		var i Import
		if p.Import(&i) {
			*is = append(*is, i)
			return true
		}
	}
	return false
}
