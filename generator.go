package main

import "unicode"

type Generator interface {
	Generate(*Result, *Parser, *TypeSpec)
}

func GeneratorsAll() []Generator {
	return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
}

func Private(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}

/*
type KeySet map[string]struct{}

type Generator interface {
	Func() string
	NOP([]Comment) bool

	Type(*Result, *Parser, *Type, string, string, []Comment, bool)

	Int(*Result, *Parser, *Int, string, string, []Comment, bool)
	Float(*Result, *Parser, *Float, string, string, []Comment, bool)
	String(*Result, *Parser, *String, string, string, []Comment, bool)

	Array(*Result, *Parser, *Array, string, string, []Comment, bool)
	ArrayElementBegin()
	ArrayElementEnd()

	Slice(*Result, *Parser, *Slice, string, string, []Comment, bool)
	SliceElementBegin()
	SliceElementEnd()

	Struct(*Result, *Parser, *Struct, string, string, []Comment, bool)
	StructFieldBegin()
	StructFieldEnd()
}

func Generate(g Generator, r *Result, p *Parser, ts *TypeSpec) {
	r.AddImports(g.Imports())
	varName := VariableName(ts.Name, false)

	r.Printf("\nfunc %s {", g.Func(ts.Name, varName))
	r.Tabs++

	GenerateType(g, r, p, ts.Name, varName, &ts.Type, ts.Comments, true)

	r.Tabs--
	r.Line("}")
}

func GenerateType(g Generator, r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, varPointer bool) {
	if g.NOP(comments) {
		return
	}

	if t.Literal != nil {
		GenerateLiteral(g, r, p, specName, varName, t.Literal, comments, varPointer)
	} else {
		if len(t.Package) > 0 {
			r.AddImport(t.Package)
		}
		g.Type(r, p, specName, varName, t, comments, varPointer)
	}
}

func GenerateTypeLit(g Generator, r *Result, p *Parser, lit TypeLit, specName string, varName string, comments []Comment, varPointer bool) {
	if g.NOP(comments) {
		return
	}

	switch lit := lit.(type) {
	case *Int:
		g.Int(r, p, lit, specName, varName, comments, varPointer)
	case *Float:
		g.Float(r, p, lit, specName, varName, comments, varPointer)
	case *String:
		g.String(r, p, lit, specName, varName, comments, varPointer)
	case *Array:
		g.Array(r, p, lit, specName, varName, comments, varPointer)
	case *Slice:
		g.Slice(r, p, lit, specName, varName, comments, varPointer)
	case *Struct:
		g.Struct(r, p, lit, specName, varName, comments, varPointer)
	}
}

func SkipField(g Generator, field *Field) bool {
	if g.SkipField(field) {
		return true
	}

	if len(field.Name) == 0 {
		*//* struct { myType } *//*
		if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
			return true
		}
		*//* struct { int } *//*
		if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
			return true
		}
	}

	return false
}

func GenerateStructFields(r *Result, p *Parser, fields []StructField, specName string, varName string, comments []Comment, forbiddenFields KeySet) {
	fields := make(KeySet)
	for _, field := range fields {
		if SkipField(g, field) {
			continue
		}
		fieldName := strings.Or(field.Name, field.Type.Name)
		fields[fieldName] = struct{}{}
	}

	for _, field := range fields {
		if SkipField(g, field) {
			continue
		}

		fieldName := strings.Or(field.Name, field.Type.Name)
		varName = fmt.Sprinf("%s.%s", varName, fieldName)

		if (field.Name == 0) && (len(field.Type.Name) > 0) {
			lit := e.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				for i := 0; i < len(s.Fields); i++ {
					f := &s.Fields[i]
					if len(f.Type.Package) == 0 {
						f.Type.Package = field.Type.Package
					}
				}
				GenerateStructFields(r, p, s.Fields, specName, varName, comments, fields)
				continue
			} else if lit != nil {
				if _, ok := forbiddenFields[fieldName]; !ok {
					g.StructFieldBegin(...)
					GenerateTypeLit(...)
					g.StructFieldEnd(...)
				}
				continue
			}
		}

		if _, ok := forbiddenFields[fieldName]; !ok {
			g.StructFieldBegin(...)
			GenerateType(...)
			g.StructFieldEnd(...)
		}
	}
}

func GenerateArrayElements() {

}
*/
