package main

import (
	"fmt"
	"go/token"
	"strconv"
)

type TypeLit interface {
	/* NOTE(anton2920): dummy function, so not all 'fmt.Stringer's are 'TypeLit's. */
	TypeLit()
	String() string
}

type Array struct {
	Size    string
	Element Type
}

type Bool struct{}

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

type Map struct {
	KeyType   Type
	ValueType Type
}

type Pointer struct {
	BaseType Type
}

type Slice struct {
	Element Type
}

type Union struct {
	Types []string
}

type String struct{}

type StructField struct {
	Comments []Comment
	Name     string
	Type     Type
	Tag      string
}

type Struct struct {
	Fields []StructField
}

var (
	_ = TypeLit(Array{})
	_ = TypeLit(Bool{})
	_ = TypeLit(Float{})
	_ = TypeLit(Int{})
	_ = TypeLit(Interface{})
	_ = TypeLit(Map{})
	_ = TypeLit(Pointer{})
	_ = TypeLit(Slice{})
	_ = TypeLit(String{})
	_ = TypeLit(Struct{})
)

func IsArray(lit TypeLit) bool {
	_, ok := lit.(Array)
	return ok
}

func IsPrimitive(lit TypeLit) bool {
	switch lit.(type) {
	case Int, Float, String, Pointer:
		return true
	default:
		return false
	}
}

func IsSlice(lit TypeLit) bool {
	_, ok := lit.(Slice)
	return ok
}

func IsStruct(lit TypeLit) bool {
	_, ok := lit.(Struct)
	return ok
}

func LiteralName(lit TypeLit) string {
	var name string
	if lit != nil {
		name = lit.String()
	}
	return name
}

func (Array) TypeLit()     {}
func (Bool) TypeLit()      {}
func (Float) TypeLit()     {}
func (Int) TypeLit()       {}
func (Interface) TypeLit() {}
func (Map) TypeLit()       {}
func (Pointer) TypeLit()   {}
func (Slice) TypeLit()     {}
func (String) TypeLit()    {}
func (Struct) TypeLit()    {}
func (Union) TypeLit()     {}

func (a Array) String() string {
	return fmt.Sprintf("[%s]%s", a.Size, a.Element.String())
}

func (b Bool) String() string {
	return "bool"
}

func (f Float) String() string {
	return fmt.Sprintf("float%d", f.Bitsize)
}

func (i Int) String() string {
	var prefix, suffix string
	if i.Unsigned {
		prefix = "u"
	}
	if i.Bitsize > 0 {
		suffix = strconv.Itoa(i.Bitsize)
	}
	return fmt.Sprintf("%sint%s", prefix, suffix)
}

func (i Interface) String() string {
	return "interface"
}

func (m Map) String() string {
	return fmt.Sprintf("map[%s]%s", m.KeyType, m.ValueType)
}

func (p Pointer) String() string {
	return fmt.Sprintf("*%s", p.BaseType)
}

func (s Slice) String() string {
	return fmt.Sprintf("[]%s", s.Element)
}

func (s String) String() string {
	return "string"
}

func (s Struct) String() string {
	return "struct"
}

func (u Union) String() string {
	return "union"
}

func (p *Parser) Array(a *Array) bool {
	if p.Token(token.LBRACK) {
		var size int
		if p.IntLit(&size) {
			a.Size = strconv.Itoa(size)
		} else {
			p.Error = nil
			if !p.Ident(&a.Size) {
				return false
			}
		}
		if p.Token(token.RBRACK) {
			if p.Type(&a.Element) {
				return true
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

func (p *Parser) Interface(i *Interface) bool {
	if p.Token(token.INTERFACE) {
		if p.Token(token.LBRACE) {
			for p.Curr().GoToken != token.RBRACE {
				p.Next()
			}
			//fmt.Printf("%#v\n", i)
			p.Next()
			return true
		}
	}
	return false
}

func (p *Parser) Map(m *Map) bool {
	if p.Token(token.MAP) {
		if p.Token(token.LBRACK) {
			if p.Type(&m.KeyType) {
				if p.Token(token.RBRACK) {
					if p.Type(&m.ValueType) {
						return true
					}
				}
			}
		}
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
	var comments []Comment
	p.Comments(&comments)
	p.Error = nil

	/* Option 1: IdentList Type. */
	pos := p.Position
	var idents []string
	if p.IdentList(&idents) {
		var t Type
		if p.Type(&t) {
			var tag string
			p.StringLit(&tag)
			p.Error = nil

			if comments == nil {
				p.Comments(&comments)
				p.Error = nil
			}

			for i := 0; i < len(idents); i++ {
				*fs = append(*fs, StructField{Comments: comments, Name: idents[i], Type: t, Tag: tag})
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
		p.StringLit(&tag)
		p.Error = nil

		if comments == nil {
			p.Comments(&comments)
			p.Error = nil
		}

		*fs = append(*fs, StructField{Comments: comments, Type: t, Tag: tag})

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
			p.Next()
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
		case "bool":
			*tl = Bool{}
			p.Next()
			return true
		case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64":
			var i Int
			if p.Int(&i) {
				*tl = i
				return true
			}
			return false
		case "float32", "float64":
			var f Float
			if p.Float(&f) {
				*tl = f
				return true
			}
			return false
		case "string":
			var s String
			if p.String(&s) {
				*tl = s
				return true
			}
			return false
		}
	case token.INTERFACE:
		var i Interface
		if p.Interface(&i) {
			*tl = i
			return true
		}
		return false
	case token.LBRACK:
		switch p.Next().GoToken {
		case token.INT, token.IDENT:
			p.Position = pos
			var a Array
			if p.Array(&a) {
				*tl = a
				return true
			}
			return false
		case token.RBRACK:
			p.Position = pos
			var s Slice
			if p.Slice(&s) {
				*tl = s
				return true
			}
			return false
		}
	case token.MAP:
		var m Map
		if p.Map(&m) {
			*tl = m
			return true
		}
		return false
	case token.MUL:
		var ptr Pointer
		if p.Pointer(&ptr) {
			*tl = ptr
			return true
		}
		return false
	case token.STRUCT:
		var s Struct
		if p.Struct(&s) {
			*tl = s
			return true
		}
		return false
	}

	return false
}
