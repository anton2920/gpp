package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type FormatJSON struct{}

func FindLiteral(packageName string, typeName string) TypeLit {
	parsedFiles := ParsedPackages[packageName]
	for _, parsedFile := range parsedFiles {
		for _, spec := range parsedFile.Specs {
			if typeName == spec.Name {
				if spec.Type.Literal == nil {
					FindLiteral(strings.Or(spec.Type.Package, packageName), spec.Type.Name)
				}
				return spec.Type.Literal
			}
		}
	}
	return nil
}

func (f *FormatJSON) SerializeSlice(g *Generator, name string, s *Slice) {
	letters := []byte{'i', 'j', 'k', 'l', 'm', 'n'}
	letter := letters[g.Tabs-1]

	g.WriteString("s.PutArrayBegin()\n")
	g.Printf("for %c := 0; %c < len(%s); %c++ {\n", letter, letter, name, letter)
	g.Tabs++
	for {
		if len(s.Element.Name) != 0 {
			lit := FindLiteral(strings.Or(s.Element.Package, g.Package), s.Element.Name)
			if _, ok := lit.(*Struct); !ok {
				f.SerializeTypeLit(g, fmt.Sprintf("%s(%s[%c])", lit, name, letter), lit)
				break
			}
		}
		f.SerializeType(g, fmt.Sprintf("%s[%c]", name, letter), &s.Element)
		break
	}
	g.Tabs--
	g.WriteString("}\n")
	g.WriteString("s.PutArrayEnd()\n")
}

func JSONPrivate(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}

func (f *FormatJSON) SerializeStructFields(g *Generator, name string, fields []StructField) {
	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (JSONPrivate(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (JSONPrivate(field.Type.Literal.String()[0])) {
				continue
			}
		}

		if len(field.Type.Name) > 0 {
			lit := FindLiteral(strings.Or(field.Type.Package, g.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					f.SerializeStructFields(g, name+"."+field.Type.Name, s.Fields)
					continue
				}
			} else {
				g.Printf("s.PutKey(`%s`)\n", field.Type.Name)
				f.SerializeTypeLit(g, fmt.Sprintf("%s(%s.%s)", lit, name, strings.Or(field.Name, field.Type.Name)), lit)
				continue
			}
		}

		g.Printf("s.PutKey(`%s`)\n", field.Name)
		f.SerializeType(g, name+"."+field.Name, &field.Type)
	}
}

func (f *FormatJSON) SerializeStruct(g *Generator, name string, s *Struct) {
	g.WriteString("s.PutObjectBegin()\n")
	f.SerializeStructFields(g, name, s.Fields)
	g.WriteString("s.PutObjectEnd()\n")
}

func (f *FormatJSON) SerializeTypeLit(g *Generator, name string, lit TypeLit) {
	switch lit := lit.(type) {
	case *Int, *Float, *String:
		s := lit.String()
		g.Printf("s.Put%c%s(%s)\n", unicode.ToUpper(rune(s[0])), s[1:], name)
	case *Slice:
		f.SerializeSlice(g, name, lit)
	case *Struct:
		f.SerializeStruct(g, name, lit)
	}
}

func (f *FormatJSON) SerializeType(g *Generator, name string, t *Type) {
	if t.Literal != nil {
		f.SerializeTypeLit(g, name, t.Literal)
	} else {
		tabs := g.Tabs

		if len(t.Package) > 0 {
			g.AddImport(t.Package)
			g.WriteString(t.Package)
			g.Tabs = 0
			g.WriteRune('.')
		}

		g.Printf("Put%sJSON(s, &%s)\n", t.Name, name)
		g.Tabs = tabs
	}
}

func (f *FormatJSON) Serialize(g *Generator, ts *TypeSpec) {
	g.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	g.Printf("\nfunc Put%sJSON(s *json.Serializer, %s *%s) {\n", ts.Name, name, ts.Name)
	g.Tabs++

	f.SerializeType(g, name, &ts.Type)

	g.Tabs--
	g.WriteString("}\n")
}
