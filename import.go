package main

import (
	"go/token"

	"github.com/anton2920/gofa/go/lexer"
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

func ParseImport(l *lexer.Lexer, i *Import) bool {
	l.ParseIdent(&i.QualifiedName)
	l.Error = nil

	if l.ParseToken(token.PERIOD) {
		i.WithoutQualifier = true
	}
	l.Error = nil

	return l.ParseStringLit(&i.Path)
}

func ParseImports(l *lexer.Lexer, is *Imports) bool {
	if l.ParseToken(token.IMPORT) {
		if l.ParseToken(token.LPAREN) {
			for l.Curr().GoToken != token.RPAREN {
				var i Import
				if !ParseImport(l, &i) {
					return false
				}
				if !l.ParseToken(token.SEMICOLON) {
					return false
				}
				*is = append(*is, i)
			}
			return true
		}
		l.Error = nil

		var i Import
		if ParseImport(l, &i) {
			*is = append(*is, i)
			return true
		}
	}
	return false
}
