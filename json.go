package main

import (
	"fmt"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type FormatJSON struct{}

func FindLiteral(is Imports, name string, typeName string) TypeLit {
	packageName := FindPackageName(is, name)
	parsedFiles := ParsedPackages[packageName]
	for _, parsedFile := range parsedFiles {
		for _, spec := range parsedFile.Specs {
			if typeName == spec.Name {
				if spec.Type.Literal == nil {
					FindLiteral(is, strings.Or(spec.Type.Package, packageName), spec.Type.Name)
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
			lit := FindLiteral(g.Imports, strings.Or(s.Element.Package, g.Package), s.Element.Name)
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
			lit := FindLiteral(g.Imports, strings.Or(field.Type.Package, g.Package), field.Type.Name)
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
			} else if lit != nil {
				g.Printf("s.PutKey(`%s`)\n", field.Type.Name)
				f.SerializeTypeLit(g, fmt.Sprintf("%s(%s.%s)", lit, name, strings.Or(field.Name, field.Type.Name)), lit)
				continue
			}
		}

		g.Printf("s.PutKey(`%s`)\n", strings.Or(field.Name, field.Type.Name))
		f.SerializeType(g, name+"."+strings.Or(field.Name, field.Type.Name), &field.Type)
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

func (f *FormatJSON) DeserializeSlice(g *Generator, name string, s *Slice) {
	var elementType string

	letters := []byte{'i', 'j', 'k', 'l', 'm', 'n'}
	letter := letters[g.Tabs-1]

	if len(s.Element.Name) == 0 {
		elementType = s.Element.Literal.String()
	} else {
		if len(s.Element.Package) > 0 {
			g.AddImport(s.Element.Package)
			elementType = s.Element.Package + "."
		}
		elementType += s.Element.Name
	}

	g.WriteString("var n int\n")
	g.WriteString("d.GetSliceBegin(&n)\n")
	g.Printf("%s = make([]%s, n)\n", name, elementType)
	g.Printf("for %c := 0; %c < len(%s); %c++ {\n", letter, letter, name, letter)
	g.Tabs++
	for {
		if len(s.Element.Name) > 0 {
			lit := FindLiteral(g.Imports, strings.Or(s.Element.Package, g.Package), s.Element.Name)
			if _, ok := lit.(*Struct); !ok {
				f.DeserializeTypeLit(g, fmt.Sprintf("(*%s)(unsafe.Pointer(&%s[%c]))", lit, name, letter), lit, true)
				break
			}
		}
		f.DeserializeType(g, fmt.Sprintf("%s[%c]", name, letter), &s.Element)
		break
	}
	g.Tabs--
	g.WriteString("}\n")
	g.WriteString("d.GetSliceEnd()\n")
}

func (f *FormatJSON) DeserializeStructFields(g *Generator, name string, fields []StructField) {
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
			lit := FindLiteral(g.Imports, strings.Or(field.Type.Package, g.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					f.DeserializeStructFields(g, name+"."+field.Type.Name, s.Fields)
					continue
				}
			} else if lit != nil {
				g.Printf("case \"%s\":\n", field.Type.Name)
				g.Tabs++
				f.DeserializeTypeLit(g, fmt.Sprintf("(*%s)(unsafe.Pointer(&%s.%s))", lit, name, strings.Or(field.Name, field.Type.Name)), lit, true)
				g.Tabs--
				continue
			}
		}

		g.Printf("case \"%s\":\n", strings.Or(field.Name, field.Type.Name))
		g.Tabs++
		f.DeserializeType(g, name+"."+strings.Or(field.Name, field.Type.Name), &field.Type)
		g.Tabs--
	}
}

func (f *FormatJSON) DeserializeStruct(g *Generator, name string, s *Struct) {
	g.WriteString("d.GetObjectBegin()\n")
	g.WriteString("for d.GetKey(&key) {\n")
	g.Tabs++
	{
		g.WriteString("switch key {\n")
		f.DeserializeStructFields(g, name, s.Fields)
		g.WriteString("}\n")
	}
	g.Tabs--
	g.WriteString("}\n")
	g.WriteString("d.GetObjectEnd()\n")
}

func (f *FormatJSON) DeserializeTypeLit(g *Generator, name string, lit TypeLit, alreadyPointer bool) {
	var amp string
	if !alreadyPointer {
		amp = "&"
	}

	switch lit := lit.(type) {
	case *Int, *Float, *String:
		s := lit.String()
		g.Printf("d.Get%c%s(%s%s)\n", unicode.ToUpper(rune(s[0])), s[1:], amp, name)
	case *Slice:
		f.DeserializeSlice(g, name, lit)
	case *Struct:
		f.DeserializeStruct(g, name, lit)
	}
}

func (f *FormatJSON) DeserializeType(g *Generator, name string, t *Type) {
	if t.Literal != nil {
		f.DeserializeTypeLit(g, name, t.Literal, false)
	} else {
		tabs := g.Tabs

		if len(t.Package) > 0 {
			g.AddImport(t.Package)
			g.WriteString(t.Package)
			g.Tabs = 0
			g.WriteRune('.')
		}

		g.Printf("Get%sJSON(d, &%s)\n", t.Name, name)
		g.Tabs = tabs
	}
}

func (f *FormatJSON) Deserialize(g *Generator, ts *TypeSpec) {
	g.AddImport("unsafe")
	g.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	g.Printf("\nfunc Get%sJSON(d *json.Deserializer, %s *%s) bool {\n", ts.Name, name, ts.Name)
	g.Tabs++

	g.WriteString("var key string\n\n")
	f.DeserializeType(g, name, &ts.Type)
	g.Tabs--
	g.WriteRune('\n')
	g.Tabs++
	g.WriteString("return d.Error == nil\n")

	g.Tabs--
	g.WriteString("}\n")
}
