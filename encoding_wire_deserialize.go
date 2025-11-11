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

func (g GeneratorEncodingWireDeserialize) Decl(t *Type, specName string, varName string) string {
	return fmt.Sprintf("Deserialize%sWire(d *wire.Deserializer, %s %s)", specName, varName, t.String())
}

func (g GeneratorEncodingWireDeserialize) Body(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	GenerateType(g, r, p, t, specName, "", specName, varName, comments, pointer)
}

func (g GeneratorEncodingWireDeserialize) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	r.Printf("%sDeserialize%sWire(d, &%s)", t.PackagePrefix(), t.Name, varName)
}

func (g GeneratorEncodingWireDeserialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool) {
	var star string
	if pointer {
		star = "*"
	}

	litName := lit.String()
	if len(castName) == 0 {
		r.Printf("%s%s = d.%c%s()", star, varName, unicode.ToUpper(rune(litName[0])), litName[1:])
	} else {
		r.Printf("%s%s = %s(d.%c%s())", star, varName, castName, unicode.ToUpper(rune(litName[0])), litName[1:])
	}
}

func (g GeneratorEncodingWireDeserialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorEncodingWireDeserialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	r.AddImport(field.Type.Package)
	GenerateStructField(g, r, p, field, lit, specName, fieldName, field.Type.String(), varName, field.Comments)
}

func (g GeneratorEncodingWireDeserialize) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorEncodingWireDeserialize) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
	i := fmt.Sprintf("i%d", g.Level)
	g.Level++

	if a.Size == 0 {
		r.Printf("%s = make([]%s, d.%c%s())", varName, a.Element.String(), unicode.ToUpper(rune(EncodingWireSliceLengthType[0])), EncodingWireSliceLengthType[1:])
	}
	r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, varName, i)
	r.Tabs++
	{
		GenerateSliceElement(g, r, p, &a.Element, specName, fmt.Sprintf("%s[%s]", varName, i), comments)
	}
	r.Tabs--
	r.Line("}")

	g.Level--
}

func (g GeneratorEncodingWireDeserialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	a := Array{Element: s.Element}
	g.Array(r, p, &a, specName, varName, comments)
}

func (g GeneratorEncodingWireDeserialize) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
	const value = "value"

	r.Printf("switch d.%c%s() {", unicode.ToUpper(rune(EncodingWireUnionKindType[0])), EncodingWireUnionKindType[1:])
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
