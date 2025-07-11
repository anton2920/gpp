package main

import (
	"bytes"
	"fmt"
	"unicode"
)

type Generator func(*bytes.Buffer, *Struct)

const (
	JSONGeneratorPrefix = "JSONSerialize"
)

func SerialGenerator(w *bytes.Buffer, s *Struct) {
	w.Write([]byte("TODO(anton2920): serial generator\n"))
}

func JSONGenerateField(w *bytes.Buffer, name string, field *StructField) {
	var takeAddress, takeElement, cast bool

	kind := field.Type.Kind

	w.WriteRune('\t')
	if field.Type.Kind == TypeKindUnknown {
		if len(field.Type.Package) > 0 {
			w.WriteString(field.Type.Package)
			w.WriteRune('.')
		}
		w.WriteString(JSONGeneratorPrefix)
		w.WriteString(field.Type.Name)
		takeAddress = true
	} else {
		w.WriteString(JSONGeneratorPrefix)
		if kind == TypeKindSlice {
			kind = field.Type.InnerKind
			takeElement = true
		}
		w.WriteRune(unicode.ToUpper(rune(TypeKind2String[kind][0])))
		w.WriteString(TypeKind2String[kind][1:])
		cast = field.Type.Name != TypeKind2String[kind]
	}
	w.WriteString(`(buffer, `)
	if cast {
		w.WriteString(TypeKind2String[kind])
		w.WriteRune('(')
	}
	if takeAddress {
		w.WriteRune('&')
	}
	w.WriteString(name)
	w.WriteRune('.')
	w.WriteString(field.Name)
	if cast {
		w.WriteRune(')')
	}
	if takeElement {
		w.WriteString(`[i]`)
	}
	w.WriteRune(')')
	w.WriteRune('\n')
}

func JSONGenerateStruct(w *bytes.Buffer, s *Struct) {
	name := VariableName(s.Name, false)

	w.WriteString(`func `)
	w.WriteString(JSONGeneratorPrefix)
	w.WriteString(s.Name)
	w.WriteRune('(')
	w.WriteString(`buffer *[]byte, `)
	w.WriteString(name)
	w.WriteString(` *`)
	w.WriteString(s.Name)
	w.WriteString(") {\n")

	w.WriteString("\t*buffer = append(*buffer, `{`...)\n")
	for i := 0; i < len(s.Fields); i++ {
		field := &s.Fields[i]
		if field.Tag == "`json:\"-\"`" {
			continue
		}
		if i > 0 {
			w.WriteString("\t*buffer = append(*buffer, `,`...)\n")
		}

		w.WriteString("\t*buffer = append(*buffer, `\"")
		w.WriteString(field.Name)
		w.WriteString("\":`...)\n")

		if field.Type.Kind != TypeKindSlice {
			JSONGenerateField(w, name, field)
		} else {
			w.WriteString("\t*buffer = append(*buffer, `[`...)\n")

			w.WriteString("\tfor i := 0; i < len(")
			w.WriteString(field.Name)
			w.WriteString("); i++ {\n")
			{
				w.WriteString("\t\tif i > 0 {\n")
				w.WriteString("\t\t\t*buffer = append(*buffer, `,`...)\n")
				w.WriteString("\t\t}\n")
				w.WriteRune('\t')
				JSONGenerateField(w, name, field)
			}
			w.WriteString("\t\t*buffer = append(*buffer, `]`...)\n")

			w.WriteString("\t}\n")
		}
	}
	w.WriteString("\t*buffer = append(*buffer, `}`...)\n")

	w.WriteString("}\n")
}

func JSONGenerateArray(w *bytes.Buffer, s *Struct) {
	name := VariableName(s.Name, true)

	w.WriteString(`func `)
	w.WriteString(JSONGeneratorPrefix)
	w.WriteString(s.Name)
	w.WriteString(`s(`)
	w.WriteString(`buffer *[]byte, `)
	w.WriteString(name)
	w.WriteString(` []`)
	w.WriteString(s.Name)
	w.WriteString(") {\n")

	w.WriteString("\t*buffer = append(*buffer, `[`...)\n")
	w.WriteString("\tfor i := 0; i < len(")
	w.WriteString(name)
	w.WriteString("); i++ {\n")
	{
		w.WriteString("\t\t" + JSONGeneratorPrefix)
		w.WriteString(s.Name)
		w.WriteString(`(buffer, `)
		w.WriteString(name)
		w.WriteString("[i])\n")
	}
	w.WriteString("\t}\n")
	w.WriteString("\t*buffer = append(*buffer, `]`...)\n")

	w.WriteString("}\n")
}

func JSONGenerator(w *bytes.Buffer, s *Struct) {
	JSONGenerateStruct(w, s)
	w.WriteRune('\n')
	JSONGenerateArray(w, s)
}

/* NOTE(anton2920): this supports only ASCII. */
func VariableName(typeName string, array bool) string {
	var lastUpper int
	for i := 0; i < len(typeName); i++ {
		if unicode.IsUpper(rune(typeName[i])) {
			lastUpper = i
		}
	}

	var suffix string
	if array {
		suffix = "s"
	}

	return fmt.Sprintf("%c%s%s", unicode.ToLower(rune(typeName[lastUpper])), typeName[lastUpper+1:], suffix)
}
