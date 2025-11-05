package main

import (
	"bytes"
	"go/token"
)

type Type struct {
	Package string
	Name    string
	Args    []Type
	Literal TypeLit
}

type TypeParam struct {
	Name       string
	Constraint Type
}

type TypeSpec struct {
	Comment

	Name   string
	Params []TypeParam
	Type   Type
	Alias  bool
}

func (t *Type) String() string {
	if t.Literal != nil {
		return t.Literal.String()
	}

	var buf bytes.Buffer

	if len(t.Package) > 0 {
		buf.WriteString(t.Package)
		buf.WriteRune('.')
	}
	buf.WriteString(t.Name)

	if len(t.Args) > 0 {
		buf.WriteRune('[')
		for i := 0; i < len(t.Args); i++ {
			if i > 0 {
				buf.WriteRune(',')
			}
			buf.WriteString(t.Args[i].String())
		}
		buf.WriteRune(']')
	}

	return buf.String()
}

func (p *Parser) Type(t *Type) bool {
	if p.TypeLit(&t.Literal) {
		return true
	}
	p.Error = nil

	var ident string
	if p.Ident(&ident) {
		if p.Token(token.PERIOD) {
			t.Package = ident
			ReferencedPackages[t.Package] = struct{}{}

			if p.Ident(&t.Name) {
				if p.Curr().GoToken == token.LBRACK {
					p.TypeArgs(&t.Args)
				}
				return true
			}
		}
		p.Error = nil
		t.Name = ident

		if p.Curr().GoToken == token.LBRACK {
			p.TypeArgs(&t.Args)
		}
		return true
	}

	return false
}

func (p *Parser) TypeArgs(ts *[]Type) bool {
	if p.Token(token.LBRACK) {
		for p.Curr().GoToken != token.RBRACK {
			var t Type
			if !p.Type(&t) {
				return false
			}
			*ts = append(*ts, t)
		}
		return true
	}
	return false
}

func (p *Parser) TypeParams(tparams *[]TypeParam) bool {
	if p.Token(token.LBRACK) {
		var idents []string
		if p.IdentList(&idents) {
			var t Type
			if p.Type(&t) {
				if p.Token(token.RBRACK) {
					for i := 0; i < len(idents); i++ {
						*tparams = append(*tparams, TypeParam{Name: idents[i], Constraint: t})
					}
					return true
				}
			}
		}
	}
	return false
}

func (p *Parser) TypeSpec(ts *TypeSpec) bool {
	var comment Comment
	p.Comment(&comment)
	p.Error = nil

	if p.Ident(&ts.Name) {
		p.TypeParams(&ts.Params)
		p.Error = nil

		if p.Token(token.ASSIGN) {
			ts.Alias = true
		}
		p.Error = nil

		if p.Type(&ts.Type) {
			ts.Comment = comment
			return true
		}
	}
	return false
}

func (p *Parser) TypeDecl(tss *[]TypeSpec) bool {
	/* Comment before 'type': apply to all type specs. */
	/* NOTE(anton2920): because this function is called when 'p.Curr()' is 'token.TYPE', we need to backtrack to check for possible comment. */
	p.Position--
	var comment Comment
	if !p.Comment(&comment) {
		p.Error = nil
		p.Next()
	}

	if p.Token(token.TYPE) {
		if p.Token(token.LPAREN) {
			for p.Curr().GoToken != token.RPAREN {
				var ts TypeSpec
				if !p.TypeSpec(&ts) {
					return false
				}
				if !p.Token(token.SEMICOLON) {
					return false
				}
				if ts.Comment == nil {
					ts.Comment = comment
				}
				*tss = append(*tss, ts)
			}
			return true
		}
		p.Error = nil

		var ts TypeSpec
		if p.TypeSpec(&ts) {
			if ts.Comment == nil {
				ts.Comment = comment
			}
			*tss = append(*tss, ts)
			return true
		}
	}
	return false
}
