package main

import (
	"fmt"
	"unicode"
)

type FormatJSON struct{}

func (f *FormatJSON) SerializeSlice(g *Generator, name string, s *Slice) {
	letters := []byte{'i', 'j', 'k', 'l', 'm', 'n'}
	letter := letters[g.Tabs-1]

	g.WriteString("s.PutArrayBegin()\n")
	g.Printf("for %c := 0; %c < len(%s); %c++ {\n", letter, letter, name, letter)
	g.Tabs++
	{
		f.SerializeType(g, fmt.Sprintf("%s[%c]", name, letter), &s.Element)
	}
	g.Tabs--
	g.WriteString("}\n")
	g.WriteString("s.PutArrayEnd()\n")
}

func (f *FormatJSON) SerializeStruct(g *Generator, name string, s *Struct) {
	g.WriteString("s.PutObjectBegin()\n")
	for i := 0; i < len(s.Fields); i++ {
		field := &s.Fields[i]
		if field.Tag == `json:"-"` {
			continue
		}

		/* TODO(anton2920): embed fields without name. */
		fieldName := field.Name
		if len(fieldName) == 0 {
			fieldName = field.Type.Name
		}

		g.Printf("s.PutKey(`%s`)\n", fieldName)
		f.SerializeType(g, name+"."+fieldName, &field.Type)
	}
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
			/* TODO(anton2920): implement parsing of types from other packages. */
			g.AddImports(Import{Path: GOFA + t.Package})

			g.WriteString(t.Package)
			g.Tabs = 0
			g.WriteRune('.')
		}

		g.Printf("Put%sJSON(s, &%s)\n", t.Name, name)
		g.Tabs = tabs
	}
}

func (f *FormatJSON) Serialize(g *Generator, ts *TypeSpec) {
	g.AddImports(Import{Path: GOFA + "encoding/json"})
	name := VariableName(ts.Name, false)

	g.Printf("\nfunc Put%sJSON(s *json.Serializer, %s *%s) {\n", ts.Name, name, ts.Name)
	g.Tabs++

	f.SerializeType(g, name, &ts.Type)

	g.Tabs--
	g.WriteString("}\n")
}
