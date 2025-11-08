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
	tabs := r.Tabs

	if len(t.Package) > 0 {
		r.AddImport(t.Package)
		r.String(t.Package)
		r.Tabs = 0
		r.Rune('.')
	}

	r.Printf("Serialize%sJSON(s, &%s)", t.Name, varName)
	r.Tabs = tabs
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

func (g GeneratorEncodingJSONSerialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string) {
	r.Line("s.ObjectBegin()")
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
	r.Line("s.ObjectEnd()")
}

func (g GeneratorEncodingJSONSerialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName, varName string) {
	if field.Tag == `json:"-"` {
		return
	}

	r.Printf("s.Key(`%s`)", fieldName)
	if lit != nil {
		GenerateTypeLit(g, r, p, lit, specName, fieldName, lit.String(), varName, field.Comments, false)
	} else {
		GenerateType(g, r, p, &field.Type, specName, varName, field.Comments, false)
	}
}

func (g GeneratorEncodingJSONSerialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	r.Line("s.ArrayBegin()")
	r.Printf("for _, element := range %s {", varName)
	r.Tabs++
	{
		GenerateSliceElement(g, r, p, &s.Element, specName, varName, comments)
	}
	r.Tabs--
	r.Line("}")
	r.Line("s.ArrayEnd()")
}
