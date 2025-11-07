package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct{}

func (g GeneratorFillValues) NOP(comments []Comment) bool {
	for _, comment := range comments {
		if fc, ok := comment.(FillComment); ok {
			return fc.NOP
		}
	}
	return false
}

func (g GeneratorFillValues) Imports() []string {
	return []string{GOFA + "net/url"}
}

func (g GeneratorFillValues) Func(specName string, varName string) string {
	return fmt.Sprintf("Fill%sFromValues(vs url.Values, %s *%s)", specName, varName, specName)
}

func (g GeneratorFillValues) Return() string {
	return ""
}

func (g GeneratorFillValues) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	tabs := r.Tabs

	if len(t.Package) > 0 {
		r.AddImport(t.Package)
		r.String(t.Package)
		r.Tabs = 0
		r.Rune('.')
	}

	r.Printf("Fill%sFromValues(vs, &%s)", t.Name, varName)
	r.Tabs = tabs
}

func (g GeneratorFillValues) Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool) {
	var star string
	if pointer {
		star = "*"
	}

	var fc FillComment
	for _, comment := range comments {
		if c, ok := comment.(FillComment); ok {
			fc.Enum = fc.Enum || c.Enum
			strings.Replace(&fc.Func, c.Func)
		}
	}

	switch lit := lit.(type) {
	case *Int, *Float:
		litName := lit.String()
		if (len(castName) == 0) || (litName == castName) {
			r.Printf(`%s%s, _ = vs.Get%c%s("%s")`, star, varName, unicode.ToUpper(rune(litName[0])), litName[1:], fieldName)
		} else {
			r.Line("{")
			r.Tabs++
			r.Printf(`tmp, _ := vs.Get%c%s("%s")`, unicode.ToUpper(rune(litName[0])), litName[1:], fieldName)
			if !fc.Enum {
				r.Printf("%s%s = %s(tmp)", star, varName, castName)
			} else {
				r.AddImport(GOFA + "ints")
				r.Printf("%s%s = %s(ints.Clamp(int(%s), int(%sNone+1), int(%sCount)))", star, varName, castName, varName, castName, castName)
			}
			r.Tabs--
			r.Line("}")
		}
	case *String:
		if len(fc.Func) == 0 {
			r.Printf(`%s%s = vs.Get("%s")`, star, varName, fieldName)
		} else {
			r.Printf(`%s%s, _ = %s(vs.Get("%s"))`, star, varName, fc.Func, fieldName)
		}
	}
}

func (g GeneratorFillValues) Slice(*Result, *Parser, *Slice, string, string, []Comment) {}

func (g GeneratorFillValues) Struct(r *Result, p *Parser, s *Struct, specName string, varName string) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorFillValues) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	if lit != nil {
		GenerateTypeLit(g, r, p, lit, specName, fieldName, field.Type.Name, varName, field.Comments, false)
	} else if field.Type.Name == "" {
		GenerateTypeLit(g, r, p, field.Type.Literal, specName, fieldName, "", varName, field.Comments, false)
	} else {
		GenerateType(g, r, p, &field.Type, specName, varName, field.Comments, false)
	}
}
