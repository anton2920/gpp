package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingJSONSerialize struct{}

func (g GeneratorEncodingJSONSerialize) Imports() []string {
	return []string{GOFA + "encoding/json"}
}

func (g GeneratorEncodingJSONSerialize) Func(specName string, varName string) string {
	return fmt.Sprintf("Serialize%sJSON(s *json.Serializer, %s *%s)", specName, varName, specName)
}

func (g GeneratorEncodingJSONSerialize) Return() string {
	return ""
}

func (g GeneratorEncodingJSONSerialize) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, _ []Comment, pointer bool) {
	r.Printf("%sSerialize%sJSON(s, &%s)", t.PackagePrefix(), t.Name, varName)
}

func (g GeneratorEncodingJSONSerialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, _ string, castName string, varName string, _ []Comment, pointer bool) {
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

func (g GeneratorEncodingJSONSerialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	r.Line("s.ObjectBegin()")
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
	r.Line("s.ObjectEnd()")
}

func (g GeneratorEncodingJSONSerialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName, varName string) {
	r.Printf("s.Key(`%s`)", fieldName)
	GenerateStructField(g, r, p, field, lit, specName, fieldName, LiteralName(lit), varName, field.Comments)
}

func (g GeneratorEncodingJSONSerialize) StructFieldSkip(field *StructField) bool {
	return JSONStructFieldSkip(field)
}

func (g GeneratorEncodingJSONSerialize) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
}

func (g GeneratorEncodingJSONSerialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	const element = "element"

	r.Line("s.ArrayBegin()")
	r.Printf("for _, %s := range %s {", element, varName)
	r.Tabs++
	{
		GenerateSliceElement(g, r, p, &s.Element, specName, element, comments)
	}
	r.Tabs--
	r.Line("}")
	r.Line("s.ArrayEnd()")
}

func (g GeneratorEncodingJSONSerialize) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
}
