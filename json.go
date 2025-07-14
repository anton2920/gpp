package main

import (
	"unicode"
)

type FormatJSON struct {
}

func JSONGenerateField(g *Generator, name string, field *StructField) {
	var takeAddress, takeElement, cast bool

	kind := field.Type.Kind

	g.WriteRune('\t')
	if field.Type.Kind == TypeKindUnknown {
		if len(field.Type.Package) > 0 {
			g.Printf("%s.", field.Type.Package)
			g.AddImports(Import{Path: GOFA + field.Type.Package})
		}
		g.Printf("Put%sJSON(s, ", field.Type.Name)
		takeAddress = true
	} else {
		if kind == TypeKindSlice {
			kind = field.Type.InnerKind
			takeElement = true
		}
		g.Printf(`s.Put%c%s(`, unicode.ToUpper(rune(TypeKind2String[kind][0])), TypeKind2String[kind][1:])
		cast = field.Type.Name != TypeKind2String[kind]
	}

	if cast {
		g.WriteString(TypeKind2String[kind])
		g.WriteRune('(')
	}
	if takeAddress {
		g.WriteRune('&')
	}
	g.WriteString(name)
	g.WriteRune('.')
	g.WriteString(field.Name)
	if takeElement {
		g.WriteString(`[i]`)
	}
	if cast {
		g.WriteRune(')')
	}
	g.WriteRune(')')
	g.WriteRune('\n')
}

func JSONGenerateStruct(g *Generator, s *Struct) {
	name := VariableName(s.Name, false)

	g.Printf("func Put%sJSON(s *json.Serializer, %s *%s) {\n", s.Name, name, s.Name)

	g.WriteString("\ts.PutObjectBegin()\n")
	for i := 0; i < len(s.Fields); i++ {
		field := &s.Fields[i]
		if field.Tag == "`json:\"-\"`" {
			continue
		}

		g.Printf("\ts.PutKey(`%s`)\n", field.Name)

		if field.Type.Kind != TypeKindSlice {
			JSONGenerateField(g, name, field)
		} else {
			g.WriteString("\ts.PutArrayBegin()\n")
			g.Printf("\tfor i := 0; i < len(%s.%s); i++ {\n", name, field.Name)
			{
				g.WriteRune('\t')
				JSONGenerateField(g, name, field)
			}
			g.WriteString("\t}\n")
			g.WriteString("\ts.PutArrayEnd()\n")
		}
	}
	g.WriteString("\ts.PutObjectEnd()\n")

	g.WriteString("}\n")
}

func JSONGenerateArray(g *Generator, s *Struct) {
	name := VariableName(s.Name, true)

	g.Printf("func Put%ssJSON(s *json.Serializer, %s []%s) {\n", s.Name, name, s.Name)

	g.WriteString("\ts.PutArrayBegin()\n")
	g.Printf("\tfor i := 0; i < len(%s); i++ {\n", name)
	{
		g.Printf("\t\tPut%sJSON(s, &%s[i])\n", s.Name, name)
	}
	g.WriteString("\t}\n")
	g.WriteString("\ts.PutArrayEnd()\n")

	g.WriteString("}\n")
}

func (fj *FormatJSON) Generate(g *Generator, s *Struct) {
	g.AddImports(Import{Path: GOFA + "encoding/json"})

	g.WriteRune('\n')
	JSONGenerateStruct(g, s)
	g.WriteRune('\n')
	JSONGenerateArray(g, s)
}
