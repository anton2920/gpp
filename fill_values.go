package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct{}

func (g GeneratorFillValues) Generate(r *Result, p *Parser, ts *TypeSpec) {
	r.AddImport(GOFA + "net/url")
	name := VariableName(ts.Name, false)

	r.Printf("\nfunc Fill%sFromValues(vs url.Values, %s *%s) {", ts.Name, name, ts.Name)
	r.Tabs++

	g.GenerateType(r, p, "", name, &ts.Type, ts.Name, true)

	r.Tabs--
	r.Line("}")
}

func (g GeneratorFillValues) GenerateType(r *Result, p *Parser, key string, name string, t *Type, castName string, alreadyPointer bool) {
	if t.Literal != nil {
		g.GenerateTypeLit(r, p, key, name, t.Literal, castName, alreadyPointer)
	} else {
		tabs := r.Tabs

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.String(t.Package)
			r.Tabs = 0
			r.Rune('.')
		}

		r.Printf("Fill%sFromValues(vs, &%s)", t.Name, name)
		r.Tabs = tabs
	}
}

func (g GeneratorFillValues) GenerateTypeLit(r *Result, p *Parser, key string, name string, lit TypeLit, castName string, alreadyPointer bool) {
	var star string
	if alreadyPointer {
		star = "*"
	}

	switch lit := lit.(type) {
	case *Int, *Float:
		s := lit.String()
		if (len(castName) == 0) || (s == castName) {
			r.Printf(`%s%s, _ = vs.Get%c%s("%s")`, star, name, unicode.ToUpper(rune(s[0])), s[1:], key)
		} else {
			r.Line("{")
			r.Tabs++
			r.Printf(`tmp, _ := vs.Get%c%s("%s")`, unicode.ToUpper(rune(s[0])), s[1:], key)
			r.Printf("%s%s = %s(tmp)", star, name, castName)
			r.Tabs--
			r.Line("}")
		}
	case *String:
		r.Printf(`%s%s = vs.Get("%s")`, star, name, key)
	case *Slice:
		g.GenerateSlice(r, p, name, lit)
	case *Struct:
		g.GenerateStruct(r, p, name, lit)
	}
}

func (g GeneratorFillValues) GenerateStruct(r *Result, p *Parser, name string, s *Struct) {
	g.GenerateStructFields(r, p, name, s.Fields)
}

func (g GeneratorFillValues) GenerateStructFields(r *Result, p *Parser, name string, fields []StructField) {
	for i := 0; i < len(fields); i++ {
		field := &fields[i]
		if _, ok := field.Comment.(FillCommentNOP); ok {
			continue
		}

		key := strings.Or(field.Name, field.Type.Name)
		g.GenerateType(r, p, key, fmt.Sprintf("%s.%s", name, key), &field.Type, "", false)
	}
}

func (g GeneratorFillValues) GenerateSlice(r *Result, p *Parser, name string, s *Slice) {
	panic("TODO")
}
