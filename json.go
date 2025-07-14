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
			g.WriteString(field.Type.Package)
			g.WriteRune('.')
		}
		g.WriteString(JSONGeneratorPrefix)
		g.WriteString(field.Type.Name)
		takeAddress = true
	} else {
		g.WriteString(JSONGeneratorPrefix)
		if kind == TypeKindSlice {
			kind = field.Type.InnerKind
			takeElement = true
		}
		g.WriteRune(unicode.ToUpper(rune(TypeKind2String[kind][0])))
		g.WriteString(TypeKind2String[kind][1:])
		cast = field.Type.Name != TypeKind2String[kind]
	}
	g.WriteString(`(buffer, `)
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
	if cast {
		g.WriteRune(')')
	}
	if takeElement {
		g.WriteString(`[i]`)
	}
	g.WriteRune(')')
	g.WriteRune('\n')
}

func JSONGenerateStruct(g *Generator, s *Struct) {
	name := VariableName(s.Name, false)

	g.WriteString(`func `)
	g.WriteString(JSONGeneratorPrefix)
	g.WriteString(s.Name)
	g.WriteRune('(')
	g.WriteString(`buffer *[]byte, `)
	g.WriteString(name)
	g.WriteString(` *`)
	g.WriteString(s.Name)
	g.WriteString(") {\n")

	g.WriteString("\t*buffer = append(*buffer, `{`...)\n")
	for i := 0; i < len(s.Fields); i++ {
		field := &s.Fields[i]
		if field.Tag == "`json:\"-\"`" {
			continue
		}
		if i > 0 {
			g.WriteString("\t*buffer = append(*buffer, `,`...)\n")
		}

		g.WriteString("\t*buffer = append(*buffer, `\"")
		g.WriteString(field.Name)
		g.WriteString("\":`...)\n")

		if field.Type.Kind != TypeKindSlice {
			JSONGenerateField(g, name, field)
		} else {
			g.WriteString("\t*buffer = append(*buffer, `[`...)\n")

			g.WriteString("\tfor i := 0; i < len(")
			g.WriteString(field.Name)
			g.WriteString("); i++ {\n")
			{
				g.WriteString("\t\tif i > 0 {\n")
				g.WriteString("\t\t\t*buffer = append(*buffer, `,`...)\n")
				g.WriteString("\t\t}\n")
				g.WriteRune('\t')
				JSONGenerateField(g, name, field)
			}
			g.WriteString("\t\t*buffer = append(*buffer, `]`...)\n")

			g.WriteString("\t}\n")
		}
	}
	g.WriteString("\t*buffer = append(*buffer, `}`...)\n")

	g.WriteString("}\n")
}

func JSONGenerateArray(g *Generator, s *Struct) {
	name := VariableName(s.Name, true)

	g.WriteString(`func `)
	g.WriteString(JSONGeneratorPrefix)
	g.WriteString(s.Name)
	g.WriteString(`s(`)
	g.WriteString(`buffer *[]byte, `)
	g.WriteString(name)
	g.WriteString(` []`)
	g.WriteString(s.Name)
	g.WriteString(") {\n")

	g.WriteString("\t*buffer = append(*buffer, `[`...)\n")
	g.WriteString("\tfor i := 0; i < len(")
	g.WriteString(name)
	g.WriteString("); i++ {\n")
	{
		g.WriteString("\t\t" + JSONGeneratorPrefix)
		g.WriteString(s.Name)
		g.WriteString(`(buffer, &`)
		g.WriteString(name)
		g.WriteString("[i])\n")
	}
	g.WriteString("\t}\n")
	g.WriteString("\t*buffer = append(*buffer, `]`...)\n")

	g.WriteString("}\n")
}

func (fj *FormatJSON) Generate(g *Generator, s *Struct) {
	g.WriteRune('\n')
	JSONGenerateStruct(g, s)
	g.WriteRune('\n')
	JSONGenerateArray(g, s)
}
