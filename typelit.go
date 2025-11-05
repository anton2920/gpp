package main

import (
	"fmt"
	"go/token"
	"strconv"
)

type TypeLit interface {
	String() string
}

type Array struct {
	Size    int
	Element Type
}

type Float struct {
	Bitsize int
}

type FuncArg struct {
	Name string
	Type Type
}

type Func struct {
	Args   []FuncArg
	Values []FuncArg
}

type Int struct {
	Bitsize  int
	Unsigned bool
}

type Interface struct {
	Functions []Func
}

type Pointer struct {
	BaseType Type
}

type Slice struct {
	Element Type
}

type String struct{}

type StructField struct {
	Comment

	Name string
	Type Type
	Tag  string
}

type Struct struct {
	Fields []StructField
}

var (
	_ = TypeLit(new(Float))
	_ = TypeLit(new(Int))
	_ = TypeLit(new(Interface))
	_ = TypeLit(new(Pointer))
	_ = TypeLit(new(Slice))
	_ = TypeLit(new(String))
	_ = TypeLit(new(Struct))
)

func (a *Array) String() string {
	return fmt.Sprintf("[%d]%s", a.Size, a.Element.String())
}

func (i *Int) String() string {
	var prefix, suffix string
	if i.Unsigned {
		prefix = "u"
	}
	if i.Bitsize > 0 {
		suffix = strconv.Itoa(i.Bitsize)
	}
	return fmt.Sprintf("%sint%s", prefix, suffix)
}

func (i *Interface) String() string {
	return "interface"
}

func (f *Float) String() string {
	return fmt.Sprintf("float%d", f.Bitsize)
}

func (p *Pointer) String() string {
	return fmt.Sprintf("*%s", p.BaseType)
}

func (s *Slice) String() string {
	return fmt.Sprintf("[]%s", s.Element.String())
}

func (s *String) String() string {
	return "string"
}

func (s *Struct) String() string {
	return "struct"
}

func (p *Parser) Array(a *Array) bool {
	if p.Token(token.LBRACK) {
		if p.IntLit(&a.Size) {
			if p.Token(token.RBRACK) {
				if p.Type(&a.Element) {
					return true
				}
			}
		}
	}
	return false
}

func (p *Parser) Float(f *Float) bool {
	var ident string

	if p.Ident(&ident) {
		switch ident {
		default:
			return false
		case "float32":
			f.Bitsize = 32
		case "float64":
			f.Bitsize = 64
		}
		return true
	}

	return false
}

func (p *Parser) Int(i *Int) bool {
	var ident string

	if p.Ident(&ident) {
		switch ident {
		default:
			return false
		case "int":
		case "uint":
			i.Unsigned = true
		case "int8":
			i.Bitsize = 8
		case "uint8":
			i.Bitsize = 8
			i.Unsigned = true
		case "int16":
			i.Bitsize = 16
		case "uint16":
			i.Bitsize = 16
			i.Unsigned = true
		case "int32":
			i.Bitsize = 32
		case "uint32":
			i.Bitsize = 32
			i.Unsigned = true
		case "int64":
			i.Bitsize = 64
		case "uint64":
			i.Bitsize = 64
			i.Unsigned = true
		}
		return true
	}

	return false
}

func (p *Parser) Pointer(ptr *Pointer) bool {
	if p.Token(token.MUL) {
		if p.Type(&ptr.BaseType) {
			return true
		}
	}
	return false
}

func (p *Parser) Slice(s *Slice) bool {
	if p.Token(token.LBRACK) {
		if p.Token(token.RBRACK) {
			if p.Type(&s.Element) {
				return true
			}
		}
	}
	return false
}

func (p *Parser) String(s *String) bool {
	var ident string

	if p.Ident(&ident) {
		if ident == "string" {
			return true
		}
	}

	return false
}

func (p *Parser) StructFields(fs *[]StructField) bool {
	var comment Comment
	p.Comment(&comment)
	p.Error = nil

	/* Option 1: IdentList Type. */
	pos := p.Position
	var idents []string
	if p.IdentList(&idents) {
		var t Type
		if p.Type(&t) {
			var tag string
			if p.Curr().GoToken == token.STRING {
				p.StringLit(&tag)
			}

			if comment == nil {
				p.Comment(&comment)
				p.Error = nil
			}

			for i := 0; i < len(idents); i++ {
				*fs = append(*fs, StructField{Comment: comment, Name: idents[i], Type: t, Tag: tag})
			}

			p.Token(token.SEMICOLON)
			p.Error = nil
			return true
		}
	}

	p.Position = pos
	p.Error = nil

	/* Option 2: Type. */
	var t Type
	if p.Type(&t) {
		var tag string
		if p.Curr().GoToken == token.STRING {
			p.StringLit(&tag)
		}

		if comment == nil {
			p.Comment(&comment)
			p.Error = nil
		}

		*fs = append(*fs, StructField{Comment: comment, Type: t, Tag: tag})

		p.Token(token.SEMICOLON)
		p.Error = nil
		return true
	}

	return false
}

func (p *Parser) Struct(s *Struct) bool {
	if p.Token(token.STRUCT) {
		if p.Token(token.LBRACE) {
			for p.Curr().GoToken != token.RBRACE {
				p.Error = nil
				if !p.StructFields(&s.Fields) {
					return false
				}
			}
			//fmt.Printf("%#v\n", s)
			return true
		}
	}
	return false
}

func (p *Parser) TypeLit(tl *TypeLit) bool {
	pos := p.Position

	switch p.Curr().GoToken {
	case token.IDENT:
		switch p.Curr().Literal {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64":
			i := new(Int)
			if p.Int(i) {
				*tl = i
				return true
			}
			return false
		case "float32", "float64":
			f := new(Float)
			if p.Float(f) {
				*tl = f
				return true
			}
			return false
		case "string":
			s := new(String)
			if p.String(s) {
				*tl = s
				return true
			}
			return false
		}
	case token.LBRACK:
		switch p.Next().GoToken {
		case token.INT:
			p.Position = pos
			a := new(Array)
			if p.Array(a) {
				*tl = a
				return true
			}
			return false
		case token.RBRACK:
			p.Position = pos
			s := new(Slice)
			if p.Slice(s) {
				*tl = s
				return true
			}
			return false
		}
	case token.MUL:
		ptr := new(Pointer)
		if p.Pointer(ptr) {
			*tl = ptr
			return true
		}
		return false
	case token.STRUCT:
		s := new(Struct)
		if p.Struct(s) {
			*tl = s
			return true
		}
		return false
	}

	return false
}
