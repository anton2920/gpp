package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type KeySet map[string]struct{}

type Generator interface {
	Imports() []string
	Func(specName string, varName string) string
	Return() string

	NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool)
	Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool)

	Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment)
	StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string)

	Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment)
	Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment)

	Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment)
}

func Private(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}

func Generate(g Generator, r *Result, p *Parser, ts *TypeSpec) {
	r.AddImports(g.Imports())
	varName := VariableName(ts.Name, false)

	r.Printf("\nfunc %s {", g.Func(ts.Name, varName))
	r.Tabs++

	GenerateType(g, r, p, &ts.Type, ts.Name, varName, ts.Comments, true)

	if ret := g.Return(); len(ret) > 0 {
		r.Printf("return %s", ret)
	}

	r.Tabs--
	r.Line("}")
}

func GenerateType(g Generator, r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, varPointer bool) {
	if t.Literal != nil {
		GenerateTypeLit(g, r, p, t.Literal, specName, "", t.Literal.String(), varName, comments, varPointer)
	} else {
		if len(t.Package) > 0 {
			r.AddImport(t.Package)
		}
		g.NamedType(r, p, t, specName, varName, comments, varPointer)
	}
}

func GenerateTypeLit(g Generator, r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, varPointer bool) {
	switch lit := lit.(type) {
	case *Int, *Float, *String:
		g.Primitive(r, p, lit, specName, fieldName, castName, varName, comments, varPointer)
	case *Array:
		g.Array(r, p, lit, specName, varName, comments)
	case *Slice:
		g.Slice(r, p, lit, specName, varName, comments)
	case *Struct:
		g.Struct(r, p, lit, specName, varName, comments)
	case *Union:
		g.Union(r, p, lit, specName, varName, comments)
	}
}

func SkipField(field *StructField) bool {
	if len(field.Name) == 0 {
		/* struct { myType } */
		if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
			return true
		}
		/* struct { int } */
		if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
			return true
		}
	}
	return false
}

func GenerateStructField(g Generator, r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment) {
	if lit != nil {
		GenerateTypeLit(g, r, p, lit, specName, fieldName, castName, varName, field.Comments, false)
	} else if field.Type.Name == "" {
		GenerateTypeLit(g, r, p, field.Type.Literal, specName, fieldName, "", varName, field.Comments, false)
	} else {
		GenerateType(g, r, p, &field.Type, specName, varName, field.Comments, false)
	}
}

func GenerateStructFields(g Generator, r *Result, p *Parser, fields []StructField, specName string, varName string, forbiddenFields KeySet) {
	currentFields := make(KeySet)
	for _, field := range fields {
		if SkipField(&field) {
			continue
		}
		fieldName := strings.Or(field.Name, field.Type.Name)
		currentFields[fieldName] = struct{}{}
	}

	for _, field := range fields {
		if SkipField(&field) {
			continue
		}

		fieldName := strings.Or(field.Name, field.Type.Name)
		name := fmt.Sprintf("%s.%s", varName, fieldName)

		var lit TypeLit
		if (len(field.Name) == 0) && (len(field.Type.Name) > 0) {
			lit = p.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				for i := 0; i < len(s.Fields); i++ {
					f := &s.Fields[i]
					if len(f.Type.Package) == 0 {
						f.Type.Package = field.Type.Package
					}
				}
				GenerateStructFields(g, r, p, s.Fields, specName, name, currentFields)
				continue
			}
		}

		if _, ok := forbiddenFields[fieldName]; !ok {
			g.StructField(r, p, &field, lit, specName, fieldName, name)
		}
	}
}

func GenerateSliceElement(g Generator, r *Result, p *Parser, elem *Type, specName string, varName string, comments []Comment) {
	if len(elem.Name) > 0 {
		lit := p.FindTypeLit(r.Imports, strings.Or(elem.Package, r.Package), elem.Name)
		if (lit != nil) && (!IsStruct(lit)) {
			GenerateTypeLit(g, r, p, lit, specName, "", lit.String(), varName, comments, false)
			return
		}
	}
	GenerateType(g, r, p, elem, specName, varName, comments, false)
}

func GenerateUnion(g Generator, r *Result, p *Parser) {

}

func GeneratorsAll() []Generator {
	return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
}
