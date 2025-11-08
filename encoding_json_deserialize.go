package main

import (
	"fmt"
	"unicode"
)

type GeneratorEncodingJSONDeserialize struct{}

func (g GeneratorEncodingJSONDeserialize) Imports() []string {
	return []string{GOFA + "encoding/json"}
}

func (g GeneratorEncodingJSONDeserialize) Func(specName string, varName string) string {
	return fmt.Sprintf("Deserialize%sJSON(d *json.Deserializer, %s *%s) bool", specName, varName, specName)
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

	r.Printf("Deserialize%sJSON(d, &%s)", t.Name, varName)
	r.Tabs = tabs
}

func (g GeneratorEncodingJSONDeserialize) Primitive(r *Result, p *Parser, lit TypeLit, specName string, _ string, castName string, varName string, _ []Comment, pointer bool) {
	var amp string
	if !pointer {
		amp = "&"
	}

	litName := lit.String()
	if len(castName) == 0 {
		r.Printf("d.%c%s(%s%s)", unicode.ToUpper(rune(litName[0])), litName[1:], amp, varName)
	} else {
		r.AddImport("unsafe")
		r.Printf("d.%c%s((*%s)(unsafe.Pointer(%s%s)))", unicode.ToUpper(rune(litName[0])), litName[1:], castName, amp, varName)
	}
}

func (g GeneratorEncodingJSONDeserialize) Struct(r *Result, p *Parser, s *Struct, specName string, varName string) {
	r.Line("var key string")
	r.Line("d.ObjectBegin()")
	r.Line("for d.Key(&key) {")
	r.Tabs++
	{
		r.Line("switch key {")
		GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
		r.Line("}")
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.ObjectEnd()")
}

func (g GeneratorEncodingJSONDeserialize) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	if field.Tag == `json:"-"` {
		return
	}

	r.Printf("case \"%s\":", fieldName)
	r.Tabs++
	{
		GenerateStructField(g, r, p, field, lit, specName, fieldName, LiteralName(lit), varName, field.Comments)
	}
	r.Tabs--
}

func (g GeneratorEncodingJSONDeserialize) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
	r.Line("d.ArrayBegin()")
	r.Line("for d.Next() {")
	r.Tabs++
	{
		const element = "element"
		r.Printf("var %s %s", element, s.Element.String())
		GenerateSliceElement(g, r, p, &s.Element, specName, "element", comments)
		r.Printf("%s = append(%s, %s)", varName, varName, element)
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.ArrayEnd()")
}
