package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct{}

func (g GeneratorFillValues) Imports() []string {
	return []string{GOFA + "net/url"}
}

func (g GeneratorFillValues) Decl(t *Type, specName string, varName string) string {
	return fmt.Sprintf("Fill%sFromValues(vs url.Values, %s %s)", specName, varName, t.String())
}

func (g GeneratorFillValues) Body(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	GenerateType(g, r, p, t, specName, "", LiteralName(t.Literal), varName, comments, pointer)
}

func (g GeneratorFillValues) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	r.Printf("%sFill%sFromValues(vs, &%s)", t.PackagePrefix(), t.Name, varName)
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
			const tmp = "tmp"

			r.Line("{")
			r.Tabs++
			r.Printf(`%s, _ := vs.Get%c%s("%s")`, tmp, unicode.ToUpper(rune(litName[0])), litName[1:], fieldName)
			if !fc.Enum {
				r.Printf("%s%s = %s(%s)", star, varName, castName, tmp)
			} else {
				r.AddImport(GOFA + "ints")
				r.Printf("%s%s = %s(ints.Clamp(int(%s), 1, int(%sCount)))", star, varName, castName, tmp, castName)
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

func (g GeneratorFillValues) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorFillValues) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	GenerateStructField(g, r, p, field, lit, specName, fieldName, field.Type.Name, varName, field.Comments)
}

func (g GeneratorFillValues) StructFieldSkip(field *StructField) bool {
	for _, comment := range field.Comments {
		if fc, ok := comment.(FillComment); ok {
			if fc.NOP {
				return true
			}
		}
	}
	return false
}

func (g GeneratorFillValues) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
}

func (g GeneratorFillValues) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
}

func (g GeneratorFillValues) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
}
