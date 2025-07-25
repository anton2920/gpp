package main

import "go/token"

type Import struct {
	QualifiedName string
	Path          string
}

type Imports []Import

func (is Imports) Len() int               { return len(is) }
func (is Imports) Less(i int, j int) bool { return is[i].Path < is[j].Path }
func (is Imports) Swap(i int, j int)      { is[i], is[j] = is[j], is[i] }

func ParseImport(l *Lexer, i *Import) bool {
	ParseIdent(l, &i.QualifiedName)
	l.Error = nil

	return ParseStringLit(l, &i.Path)
}

func ParseImports(l *Lexer, is *Imports) bool {
	if ParseToken(l, token.IMPORT) {
		if ParseToken(l, token.LPAREN) {
			for l.Peek().GoToken != token.RPAREN {
				var i Import
				if !ParseImport(l, &i) {
					return false
				}
				if !ParseToken(l, token.SEMICOLON) {
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
			return false
		}
	}
	return false
}
