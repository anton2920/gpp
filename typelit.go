package main

import (
	"fmt"
	"go/token"
	"strconv"

	"github.com/anton2920/gofa/go/lexer"
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

func ParseArray(l *lexer.Lexer, a *Array) bool {
	if l.ParseToken(token.LBRACK) {
		if l.ParseIntLit(&a.Size) {
			if l.ParseToken(token.RBRACK) {
				if ParseType(l, &a.Element) {
					return true
				}
			}
		}
	}
	return false
}

func ParseFloat(l *lexer.Lexer, f *Float) bool {
	var ident string

	if l.ParseIdent(&ident) {
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

func ParseInt(l *lexer.Lexer, i *Int) bool {
	var ident string

	if l.ParseIdent(&ident) {
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

func ParsePointer(l *lexer.Lexer, p *Pointer) bool {
	if l.ParseToken(token.MUL) {
		if ParseType(l, &p.BaseType) {
			return true
		}
	}
	return false
}

func ParseSlice(l *lexer.Lexer, s *Slice) bool {
	if l.ParseToken(token.LBRACK) {
		if l.ParseToken(token.RBRACK) {
			if ParseType(l, &s.Element) {
				return true
			}
		}
	}
	return false
}

func ParseString(l *lexer.Lexer, s *String) bool {
	var ident string

	if l.ParseIdent(&ident) {
		if ident == "string" {
			return true
		}
	}

	return false
}

func ParseStructFields(l *lexer.Lexer, fs *[]StructField) bool {
	pos := l.Position

	/* Option 1: IdentList Type. */
	var idents []string
	if l.ParseIdentList(&idents) {
		var t Type
		if ParseType(l, &t) {
			var tag string
			l.ParseStringLit(&tag)
			l.Error = nil

			for i := 0; i < len(idents); i++ {
				*fs = append(*fs, StructField{Name: idents[i], Type: t, Tag: tag})
			}

			l.ParseToken(token.SEMICOLON)
			l.Error = nil
			return true
		}
	}

	l.Position = pos
	l.Error = nil

	/* Option 2: Type. */
	var t Type
	if ParseType(l, &t) {
		var tag string
		l.ParseStringLit(&tag)
		l.Error = nil

		*fs = append(*fs, StructField{Type: t, Tag: tag})

		l.ParseToken(token.SEMICOLON)
		l.Error = nil
		return true
	}

	return false
}

func ParseStruct(l *lexer.Lexer, s *Struct) bool {
	if l.ParseToken(token.STRUCT) {
		if l.ParseToken(token.LBRACE) {
			for !l.ParseToken(token.RBRACE) {
				l.Error = nil
				if !ParseStructFields(l, &s.Fields) {
					return false
				}
			}
			//fmt.Printf("%#v\n", s)
			return true
		}
	}
	return false
}

func ParseTypeLit(l *lexer.Lexer, tl *TypeLit) bool {
	pos := l.Position

	tok := l.Curr()
	switch tok.GoToken {
	case token.IDENT:
		switch tok.Literal {
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64":
			i := new(Int)
			if ParseInt(l, i) {
				*tl = i
				return true
			}
			return false
		case "float32", "float64":
			f := new(Float)
			if ParseFloat(l, f) {
				*tl = f
				return true
			}
			return false
		case "string":
			s := new(String)
			if ParseString(l, s) {
				*tl = s
				return true
			}
			return false
		}
	case token.LBRACK:
		switch l.Next().GoToken {
		case token.INT:
			l.Position = pos
			a := new(Array)
			if ParseArray(l, a) {
				*tl = a
				return true
			}
			return false
		case token.RBRACK:
			l.Position = pos
			s := new(Slice)
			if ParseSlice(l, s) {
				*tl = s
				return true
			}
			return false
		}
	case token.MUL:
		p := new(Pointer)
		if ParsePointer(l, p) {
			*tl = p
			return true
		}
		return false
	case token.STRUCT:
		s := new(Struct)
		if ParseStruct(l, s) {
			*tl = s
			return true
		}
		return false
	}

	return false
}
