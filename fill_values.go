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

	var comment FillComment
	for _, comment := range ts.Comments {
		if fc, ok := comment.(FillComment); ok {
			comment = fc
		}
	}
	g.GenerateType(r, p, "", name, &ts.Type, ts.Name, true, comment)

	r.Tabs--
	r.Line("}")
}

func (g GeneratorFillValues) GenerateType(r *Result, p *Parser, key string, name string, t *Type, castName string, alreadyPointer bool, fc FillComment) {
	if fc.NOP {
		return
	}

	if t.Literal != nil {
		g.GenerateTypeLit(r, p, key, name, t.Literal, castName, alreadyPointer, fc)
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

func (g GeneratorFillValues) GenerateTypeLit(r *Result, p *Parser, key string, name string, lit TypeLit, castName string, alreadyPointer bool, fc FillComment) {
	if fc.NOP {
		return
	}

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
			if !fc.Enum {
				r.Printf("%s%s = %s(tmp)", star, name, castName)
			} else {
				r.AddImport(GOFA + "ints")
				r.Printf("%s%s = %s(ints.Clamp(int(%s), int(%sNone+1), int(%sCount)))", star, name, castName, name, castName, castName)
			}
			r.Tabs--
			r.Line("}")
		}
	case *String:
		if len(fc.Func) == 0 {
			r.Printf(`%s%s = vs.Get("%s")`, star, name, key)
		} else {
			r.Printf(`%s%s, _ = %s(vs.Get("%s"))`, star, name, fc.Func, key)
		}
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
	for _, field := range fields {
		var comment FillComment

		for _, c := range field.Comments {
			if fc, ok := c.(FillComment); ok {
				comment = fc
			}
		}

		key := strings.Or(field.Name, field.Type.Name)
		fieldName := fmt.Sprintf("%s.%s", name, key)

		if len(field.Type.Name) > 0 {
			lit := p.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					if !comment.NOP {
						g.GenerateStructFields(r, p, fieldName, s.Fields)
					}
					continue
				}
			} else if lit != nil {
				g.GenerateTypeLit(r, p, key, fieldName, lit, field.Type.Name, false, comment)
				// fmt.Printf("For field %q: comment %#v\n", fieldName, comment)
				continue
			}
		}

		g.GenerateType(r, p, key, fieldName, &field.Type, "", false, comment)
	}
}

func (g GeneratorFillValues) GenerateSlice(r *Result, p *Parser, name string, s *Slice) {
	panic("TODO")
}
