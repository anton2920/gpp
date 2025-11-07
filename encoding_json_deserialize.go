package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingJSONDeserialize struct{}

func (g GeneratorEncodingJSONDeserialize) NOP([]Comment) bool {
	return false
}

func (g GeneratorEncodingJSONDeserialize) Imports() []string {
	return []string{GOFA + "encoding/json"}
}

func (g GeneratorEncodingJSONDeserialize) Func(specName string, varName string) string {
	return fmt.Sprintf("Deserialize%sJSON(s *json.Deserializer, %s *%s) bool", specName, varName, specName)
}

func (g GeneratorEncodingJSONDeserialize) Return() string {
	return "d.Error == nil"
}

func (g GeneratorEncodingJSONDeserialize) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, _ []Comment, pointer bool) {
	tabs := r.Tabs

	if len(t.Package) > 0 {
		r.AddImport(t.Package)
		r.String(t.Package)
		r.Tabs = 0
		r.Rune('.')
	}

	r.Printf("Deserialize%sJSON(s, &%s)", t.Name, varName)
	r.Tabs = tabs
}

func (g GeneratorEncodingJSONDeserialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, castName string, varName string, _ []Comment, pointer bool) {
	var amp string
	if !pointer {
		amp = "&"
	}

	litName := lit.String()
	if (len(castName) == 0) || (castName == litName) {
		r.Printf("d.%c%s(%s%s)", unicode.ToUpper(rune(litName[0])), litName[1:], amp, varName)
	} else {
		r.AddImport("unsafe")
		r.Printf("d.%c%s((*%s)(unsafe.Pointer(%s%s))", unicode.ToUpper(rune(litName[0])), litName[1:], castName, amp, varName)
	}
}

func (g GeneratorEncodingJSONDeserialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, _ []Comment) {
	r.Line("var key string")
	r.Line("d.ObjectBegin()")
	r.Line("for d.Key(&key) {")
	r.Tabs++
	{
		r.Line("switch key {")
		GenerateStructFields(g, r, p, s.Fields, specName, varName, nil, nil)
		r.Line("}")
	}
	r.Tabs--
	r.Line("d.ObjectEnd()")
}

func (g GeneratorEncodingJSONDeserialize) StructFieldBegin(r *Result, p *Parser, fieldName string, specName string, varName string, _ []Comment) {
	r.Printf("case \"%s\":", fieldName)
	r.Tabs++
}

func (g GeneratorEncodingJSONDeserialize) SkipField(field *StructField) bool {
	return field.Tag == `json:"-"`
}

func (g GeneratorEncodingJSONDeserialize) StructFieldEnd(r *Result, p *Parser, fieldName string, specName string, varName string, _ []Comment) {
	r.Tabs--
}

func (g GeneratorEncodingJSONDeserialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, _ []Comment) {
	r.Line("s.ArrayBegin()")
	GenerateSliceElement(g, r, p, &s.Element, specName, varName, nil)
	r.Line("s.ArrayEnd()")
}

func (g GeneratorEncodingJSONDeserialize) SliceElementBegin() {}
func (g GeneratorEncodingJSONDeserialize) SliceElementEnd()   {}
