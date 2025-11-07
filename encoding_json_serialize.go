package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingJSONSerialize struct{}

func (g GeneratorEncodingJSONSerialize) NOP([]Comment) bool {
	return false
}

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

func (g GeneratorEncodingJSONSerialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, castName string, varName string, _ []Comment, pointer bool) {
	var star string
	if pointer {
		star = "*"
	}

	litName := lit.String()
	if (len(castName) == 0) || (castName == litName) {
		r.Printf("s.%c%s(%s%s)", unicode.ToUpper(rune(litName[0])), litName[1:], star, varName)
	} else {
		r.Printf("s.%c%s(%s(%s%s))", unicode.ToUpper(rune(litName[0])), litName[1:], castName, star, varName)
	}
}

func (g GeneratorEncodingJSONSerialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, _ []Comment) {
	r.Line("s.ObjectBegin()")
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil, nil)
	r.Line("s.ObjectEnd()")
}

func (g GeneratorEncodingJSONSerialize) StructFieldBegin(r *Result, p *Parser, fieldName string, specName string, varName string, _ []Comment) {
	r.Printf("s.Key(`%s`)", fieldName)
}

func (g GeneratorEncodingJSONSerialize) SkipField(field *StructField) bool {
	return field.Tag == `json:"-"`
}

func (g GeneratorEncodingJSONSerialize) StructFieldEnd(r *Result, p *Parser, fieldName string, specName string, varName string, _ []Comment) {
}

func (g GeneratorEncodingJSONSerialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, _ []Comment) {
	r.Line("s.ArrayBegin()")
	GenerateSliceElement(g, r, p, &s.Element, specName, varName, nil)
	r.Line("s.ArrayEnd()")
}

func (g GeneratorEncodingJSONSerialize) SliceElementBegin() {}
func (g GeneratorEncodingJSONSerialize) SliceElementEnd()   {}
