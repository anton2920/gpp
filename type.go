package main

import (
	"bytes"
	"go/token"

	"github.com/anton2920/gofa/strings"
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
	Comments []Comment
	Name     string
	Params   []TypeParam
	Type     Type
	Alias    bool
}

func Plural(name string) string {
	if strings.EndsWith(name, "y") {
		return name[:len(name)-1] + "ies"
	} else if strings.EndsWith(name, "s") {
		return name + "es"
	}
	return name + "s"
}

func Singular(name string) string {
	if strings.EndsWith(name, "ies") {
		return name[:len(name)-len("ies")] + "y"
	} else if strings.EndsWith(name, "s") {
		return name[:len(name)-1]
	}
	return name
}

func (t *Type) PackagePrefix() string {
	if len(t.Package) > 0 {
		return t.Package + "."
	}
	return ""
}

func (t Type) String() string {
	if litName := LiteralName(t.Literal); len(litName) > 0 {
		return litName
	}

	var buf bytes.Buffer

	buf.WriteString(t.PackagePrefix())
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
			p.ReferencedPackages[t.Package] = struct{}{}

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
	var comments []Comment
	p.Comments(&comments)
	p.Error = nil

	if p.Ident(&ts.Name) {
		pos := p.Position
		if !p.TypeParams(&ts.Params) {
			p.Position = pos
			p.Error = nil
		}

		if p.Token(token.ASSIGN) {
			ts.Alias = true
		}
		p.Error = nil

		if p.Type(&ts.Type) {
			ts.Comments = comments
			if _, ok := ts.Type.Literal.(Interface); ok {
				for _, comment := range comments {
					if uc, ok := comment.(UnionComment); ok {
						ts.Type.Literal = Union{Types: uc.Types}
						break
					}
				}
			}
			return true
		}
	}
	return false
}

func (p *Parser) TypeDecl(tss *[]TypeSpec) bool {
	var comments []Comment
	p.Comments(&comments)
	p.Error = nil

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
				ts.Comments = AppendComments(comments, ts.Comments)
				*tss = append(*tss, ts)
			}
			p.Next()
			return true
		}
		p.Error = nil

		var ts TypeSpec
		if p.TypeSpec(&ts) {
			ts.Comments = AppendComments(comments, ts.Comments)
			*tss = append(*tss, ts)
			return true
		}
	}
	return false
}
