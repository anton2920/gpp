package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingWireDeserialize struct {
	Level int
}

func (g GeneratorEncodingWireDeserialize) Imports() []string {
	return []string{GOFA + "encoding/wire"}
}

func (g GeneratorEncodingWireDeserialize) Func(specName string, varName string) string {
	return fmt.Sprintf("Deserialize%sWire(d *wire.Deserializer, %s *%s) bool", specName, varName, specName)
}

func (g GeneratorEncodingWireDeserialize) Return() string {
	return "d.Error == nil"
}

func (g GeneratorEncodingWireDeserialize) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	tabs := r.Tabs

	if len(t.Package) > 0 {
		r.AddImport(t.Package)
		r.String(t.Package)
		r.Tabs = 0
		r.Rune('.')
	}

	r.Printf("Deserialize%sWire(d, &%s)", t.Name, varName)
	r.Tabs = tabs
}

func (g GeneratorEncodingWireDeserialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool) {
	var amp string
	if !pointer {
		amp = "&"
	}

	litName := lit.String()
	if (len(castName) == 0) || (litName == castName) {
		r.Printf("d.%c%s(%s%s)", unicode.ToUpper(rune(litName[0])), litName[1:], amp, varName)
	} else {
		r.AddImport("unsafe")
		r.Printf("d.%c%s((*%s)(unsafe.Pointer(%s%s)))", unicode.ToUpper(rune(litName[0])), litName[1:], castName, amp, varName)
	}
}

func (g GeneratorEncodingWireDeserialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorEncodingWireDeserialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	GenerateStructField(g, r, p, field, lit, specName, fieldName, LiteralName(lit), varName, field.Comments)
}

func (g GeneratorEncodingWireDeserialize) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorEncodingWireDeserialize) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
	const elementPrefix = "e"

	element := fmt.Sprintf("%s%d", elementPrefix, g.Level)
	g.Level++
	{
		r.Line("{")
		r.Tabs++
		{
			if a.Size == 0 {
				r.Printf("var n %s", EncodingWireSliceLengthType)
				r.Printf("d.%c%s(&n)", unicode.ToUpper(rune(EncodingWireSliceLengthType[0])), EncodingWireSliceLengthType[1:])
				r.Printf("%s = make([]%s, n)", varName, a.Element.String())
			}
			r.Printf("for i := 0; i < len(%s); i++ {", varName)
			r.Tabs++
			{
				r.Printf("var %s %s", element, a.Element.String())
				GenerateSliceElement(g, r, p, &a.Element, specName, element, comments)
				r.Printf("%s[i] = %s", varName, element)

			}
			r.Tabs--
			r.Line("}")
		}
		r.Tabs--
		r.Line("}")
	}
	g.Level--
}

func (g GeneratorEncodingWireDeserialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	a := Array{Element: s.Element}
	g.Array(r, p, &a, specName, varName, comments)
}

func (g GeneratorEncodingWireDeserialize) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
	const value = "value"
	const typ = "typ"

	r.Printf("var %s %s", typ, EncodingWireUnionKindType)
	r.Printf("d.%c%s(&%s)", unicode.ToUpper(rune(EncodingWireUnionKindType[0])), EncodingWireUnionKindType[1:], typ)
	r.Printf("switch typ {")
	{
		for i, t := range u.Types {
			var amp string
			if t[0] == '*' {
				t = t[1:]
				amp = "&"
			}

			r.Printf("case %d:", i)
			r.Tabs++
			{
				r.Printf("var %s %s", value, t)
				r.Printf("Deserialize%sWire(d, &%s)", t, value)
				r.Printf("*%s = %s%s", varName, amp, value)
			}
			r.Tabs--
		}
	}
	r.Line("}")
}
