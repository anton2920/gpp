package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type KeySet map[string]struct{}

type Generator interface {
	NOP([]Comment) bool

	Imports() []string
	Func(string, string) string
	Return() string

	NamedType(*Result, *Parser, *Type, string, string, []Comment, bool)
	Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool)

	//Array(*Result, *Parser, *Array, string, string, []Comment, bool)
	//ArrayElementBegin()
	//ArrayElementEnd()

	Slice(*Result, *Parser, *Slice, string, string, []Comment)

	Struct(*Result, *Parser, *Struct, string, string)
	StructField(*Result, *Parser, *StructField, TypeLit, string, string, string)
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
	if g.NOP(comments) {
		return
	}

	if t.Literal != nil {
		GenerateTypeLit(g, r, p, t.Literal, specName, "", "", varName, comments, varPointer)
	} else {
		if len(t.Package) > 0 {
			r.AddImport(t.Package)
		}
		g.NamedType(r, p, t, specName, varName, comments, varPointer)
	}
}

func GenerateTypeLit(g Generator, r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, varPointer bool) {
	if g.NOP(comments) {
		return
	}

	switch lit := lit.(type) {
	case *Int, *Float, *String:
		g.Primitive(r, p, lit, specName, fieldName, castName, varName, comments, varPointer)
	//case *Array:
	//	g.Array(r, p, lit, specName, varName, comments)
	case *Slice:
		g.Slice(r, p, lit, specName, varName, comments)
	case *Struct:
		g.Struct(r, p, lit, specName, varName)
	}
}

func GenerateSliceElement(g Generator, r *Result, p *Parser, elem *Type, specName string, varName string, comments []Comment) {
	if len(elem.Name) > 0 {
		lit := p.FindTypeLit(r.Imports, strings.Or(elem.Package, r.Package), elem.Name)
		if (lit != nil) && (!IsStruct(lit)) {
			//g.SliceElementBegin()
			GenerateTypeLit(g, r, p, lit, specName, "", lit.String(), varName, comments, false)
			//g.SliceElementEnd()
			return
		}
	}
	//g.SliceElementBegin()
	GenerateType(g, r, p, elem, specName, varName, comments, false)
	//g.SliceElementEnd()
}

func SkipField(g Generator, field *StructField) bool {
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

func GenerateStructFields(g Generator, r *Result, p *Parser, fields []StructField, specName string, varName string, forbiddenFields KeySet) {
	currentFields := make(KeySet)
	for _, field := range fields {
		if SkipField(g, &field) {
			continue
		}
		fieldName := strings.Or(field.Name, field.Type.Name)
		currentFields[fieldName] = struct{}{}
	}

	for _, field := range fields {
		if SkipField(g, &field) {
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

func GeneratorsAll() []Generator {
	//return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
	return nil
}
