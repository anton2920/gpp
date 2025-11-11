package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingWireSerialize struct {
	Level int
}

func (g GeneratorEncodingWireSerialize) Imports() []string {
	return []string{GOFA + "encoding/wire"}
}

func (g GeneratorEncodingWireSerialize) Func(specName string, varName string) string {
	return fmt.Sprintf("Serialize%sWire(s *wire.Serializer, %s *%s)", specName, varName, specName)
}

func (g GeneratorEncodingWireSerialize) Body(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment) {
	GenerateType(g, r, p, t, specName, "", LiteralName(t.Literal), varName, comments, true)
}

func (g GeneratorEncodingWireSerialize) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	r.Printf("%sSerialize%sWire(s, &%s)", t.PackagePrefix(), t.Name, varName)
}

func (g GeneratorEncodingWireSerialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool) {
	var star string
	if pointer {
		star = "*"
	}

	litName := lit.String()
	if len(castName) == 0 {
		r.Printf("s.%c%s(%s%s)", unicode.ToUpper(rune(litName[0])), litName[1:], star, varName)
	} else {
		r.Printf("s.%c%s(%s(%s%s))", unicode.ToUpper(rune(litName[0])), litName[1:], castName, star, varName)
	}
}

func (g GeneratorEncodingWireSerialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorEncodingWireSerialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	GenerateStructField(g, r, p, field, lit, specName, fieldName, LiteralName(lit), varName, field.Comments)
}

func (g GeneratorEncodingWireSerialize) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorEncodingWireSerialize) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
	i := fmt.Sprintf("i%d", g.Level)
	g.Level++

	if a.Size == 0 {
		r.Printf("s.%c%s(%s(len(%s)))", unicode.ToUpper(rune(EncodingWireSliceLengthType[0])), EncodingWireSliceLengthType[1:], EncodingWireSliceLengthType, varName)
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

func (g GeneratorEncodingWireSerialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	a := Array{Element: s.Element}
	g.Array(r, p, &a, specName, varName, comments)
}

func (g GeneratorEncodingWireSerialize) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
	r.Printf("switch %s := (*%s).(type) {", varName, varName)
	{
		for i, t := range u.Types {
			var amp string
			if t[0] != '*' {
				amp = "&"
			}

			r.Printf("case %s:", t)
			r.Tabs++
			{
				r.Printf("s.%c%s(%d)", unicode.ToUpper(rune(EncodingWireUnionKindType[0])), EncodingWireUnionKindType[1:], i)
				r.Printf("Serialize%sWire(s, %s%s)", t[1:], amp, varName)
			}
			r.Tabs--
		}
	}
	r.Line("}")
}
