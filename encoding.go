package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type Encoding struct {
	*Parser

	Index int
}

var Letters = []byte{'i', 'j', 'k', 'l', 'm', 'n'}

func (e *Encoding) SerializeSlice(r *Result, name string, s *Slice) {
	letter := Letters[e.Index]
	e.Index++

	r.Line("s.ArrayBegin()")
	r.Printfln("for %c := 0; %c < len(%s); %c++ {", letter, letter, name, letter)
	r.Tabs++
	for {
		if len(s.Element.Name) != 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(s.Element.Package, r.Package), s.Element.Name)
			if _, ok := lit.(*Struct); !ok {
				e.SerializeTypeLit(r, fmt.Sprintf("%s(%s[%c])", lit, name, letter), lit)
				break
			}
		}
		e.SerializeType(r, fmt.Sprintf("%s[%c]", name, letter), &s.Element)
		break
	}
	r.Tabs--
	r.Line("}")
	r.Line("s.ArrayEnd()")

	e.Index--
}

func JSONPrivate(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}

func (e *Encoding) SerializeStructFields(r *Result, name string, fields []StructField) {
	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (JSONPrivate(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (JSONPrivate(field.Type.Literal.String()[0])) {
				continue
			}
		}

		if len(field.Type.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					e.SerializeStructFields(r, name+"."+field.Type.Name, s.Fields)
					continue
				}
			} else if lit != nil {
				r.Printfln("s.Key(`%s`)", field.Type.Name)
				e.SerializeTypeLit(r, fmt.Sprintf("%s(%s.%s)", lit, name, strings.Or(field.Name, field.Type.Name)), lit)
				continue
			}
		}

		r.Printfln("s.Key(`%s`)", strings.Or(field.Name, field.Type.Name))
		e.SerializeType(r, name+"."+strings.Or(field.Name, field.Type.Name), &field.Type)
	}
}

func (e *Encoding) SerializeStruct(r *Result, name string, s *Struct) {
	r.Line("s.ObjectBegin()")
	e.SerializeStructFields(r, name, s.Fields)
	r.Line("s.ObjectEnd()")
}

func (e *Encoding) SerializeTypeLit(r *Result, name string, lit TypeLit) {
	switch lit := lit.(type) {
	case *Int, *Float, *String:
		s := lit.String()
		r.Printfln("s.%c%s(%s)", unicode.ToUpper(rune(s[0])), s[1:], name)
	case *Slice:
		e.SerializeSlice(r, name, lit)
	case *Struct:
		e.SerializeStruct(r, name, lit)
	}
}

func (e *Encoding) SerializeType(r *Result, name string, t *Type) {
	if t.Literal != nil {
		e.SerializeTypeLit(r, name, t.Literal)
	} else {
		tabs := r.Tabs

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.String(t.Package)
			r.Tabs = 0
			r.Rune('.')
		}

		r.Printfln("%sJSON(s, &%s)", t.Name, name)
		r.Tabs = tabs
	}
}

func (e *Encoding) Serialize(r *Result, ts *TypeSpec) {
	r.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	r.Printfln("func Serialize%sJSON(s *e.Serializer, %s *%s) {", ts.Name, name, ts.Name)
	r.Tabs++

	e.SerializeType(r, name, &ts.Type)

	r.Tabs--
	r.Line("}")
}

func (e *Encoding) DeserializeSlice(r *Result, name string, s *Slice) {
	var elementType string

	letter := Letters[e.Index]
	e.Index++

	if len(s.Element.Name) == 0 {
		elementType = s.Element.Literal.String()
	} else {
		if len(s.Element.Package) > 0 {
			r.AddImport(s.Element.Package)
			elementType = s.Element.Package + "."
		}
		elementType += s.Element.Name
	}

	r.Line("var n int")
	r.Line("d.SliceBegin(&n)")
	r.Printfln("%s = make([]%s, n)", name, elementType)
	r.Printfln("for %c := 0; %c < len(%s); %c++ {", letter, letter, name, letter)
	r.Tabs++
	for {
		if len(s.Element.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(s.Element.Package, r.Package), s.Element.Name)
			if _, ok := lit.(*Struct); !ok {
				e.DeserializeTypeLit(r, fmt.Sprintf("(*%s)(unsafe.Pointer(&%s[%c]))", lit, name, letter), lit, true)
				break
			}
		}
		e.DeserializeType(r, fmt.Sprintf("%s[%c]", name, letter), &s.Element)
		break
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.SliceEnd()")

	e.Index--
}

func (e *Encoding) DeserializeStructFields(r *Result, name string, fields []StructField) {
	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (JSONPrivate(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (JSONPrivate(field.Type.Literal.String()[0])) {
				continue
			}
		}

		if len(field.Type.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					e.DeserializeStructFields(r, name+"."+field.Type.Name, s.Fields)
					continue
				}
			} else if lit != nil {
				r.Printfln("case \"%s\":", field.Type.Name)
				r.Tabs++
				e.DeserializeTypeLit(r, fmt.Sprintf("(*%s)(unsafe.Pointer(&%s.%s))", lit, name, strings.Or(field.Name, field.Type.Name)), lit, true)
				r.Tabs--
				continue
			}
		}

		r.Printfln("case \"%s\":", strings.Or(field.Name, field.Type.Name))
		r.Tabs++
		e.DeserializeType(r, name+"."+strings.Or(field.Name, field.Type.Name), &field.Type)
		r.Tabs--
	}
}

func (e *Encoding) DeserializeStruct(r *Result, name string, s *Struct) {
	r.Line("d.ObjectBegin()")
	r.Line("for d.Key(&key) {")
	r.Tabs++
	{
		r.Line("switch key {")
		e.DeserializeStructFields(r, name, s.Fields)
		r.Line("}")
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.ObjectEnd()")
}

func (e *Encoding) DeserializeTypeLit(r *Result, name string, lit TypeLit, alreadyPointer bool) {
	var amp string
	if !alreadyPointer {
		amp = "&"
	}

	switch lit := lit.(type) {
	case *Int, *Float, *String:
		s := lit.String()
		r.Printfln("d.%c%s(%s%s)", unicode.ToUpper(rune(s[0])), s[1:], amp, name)
	case *Slice:
		e.DeserializeSlice(r, name, lit)
	case *Struct:
		e.DeserializeStruct(r, name, lit)
	}
}

func (e *Encoding) DeserializeType(r *Result, name string, t *Type) {
	if t.Literal != nil {
		e.DeserializeTypeLit(r, name, t.Literal, false)
	} else {
		tabs := r.Tabs

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.Line(t.Package)
			r.Tabs = 0
			r.Rune('.')
		}

		r.Printfln("%sJSON(d, &%s)", t.Name, name)
		r.Tabs = tabs
	}
}

func (e *Encoding) Deserialize(r *Result, ts *TypeSpec) {
	r.AddImport("unsafe")
	r.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	r.Printfln("func Deserialize%sJSON(d *e.Deserializer, %s *%s) bool {", ts.Name, name, ts.Name)
	r.Tabs++

	r.Line("var key string\n")
	e.DeserializeType(r, name, &ts.Type)
	r.WithoutTabs().Rune('\n')
	r.Line("return d.Error == nil")

	r.Tabs--
	r.Line("}")
}

func GeneratorsEncodingAll() []Generator {
	return []Generator{GeneratorJSON{}, GeneratorWire{}}
}
