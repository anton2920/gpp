package main

import (
	"fmt"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type Encoding struct {
	*Parser
}

type KeysSet map[string]struct{}

func (e *Encoding) Serialize(r *Result, ts *TypeSpec, serializerName string) {
	r.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	r.Printf("\nfunc Serialize%s%s(s *%s.Serializer, %s *%s) {", ts.Name, serializerName, stdstrings.ToLower(serializerName), name, ts.Name)
	r.Tabs++

	e.SerializeType(r, name, &ts.Type, ts.Name != LiteralName(ts.Type.Literal), true)

	r.Tabs--
	r.Line("}")
}

func (e *Encoding) Deserialize(r *Result, ts *TypeSpec, deserializerName string) {
	r.AddImport(GOFA + "encoding/json")
	name := VariableName(ts.Name, false)

	r.Printf("\nfunc Deserialize%s%s(d *%s.Deserializer, %s *%s) bool {", ts.Name, deserializerName, stdstrings.ToLower(deserializerName), name, ts.Name)
	r.Tabs++

	e.DeserializeType(r, name, &ts.Type, ts.Name != LiteralName(ts.Type.Literal), true)
	r.WithoutTabs().Rune('\n')
	r.Line("return d.Error == nil")

	r.Tabs--
	r.Line("}")
}

func (e *Encoding) SerializeType(r *Result, name string, t *Type, cast bool, alreadyPointer bool) {
	if t.Literal != nil {
		e.SerializeTypeLit(r, name, t.Literal, cast, alreadyPointer)
	} else {
		tabs := r.Tabs

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.String(t.Package)
			r.Tabs = 0
			r.Rune('.')
		}

		r.Printf("Serialize%sJSON(s, &%s)", t.Name, name)
		r.Tabs = tabs
	}
}

func (e *Encoding) DeserializeType(r *Result, name string, t *Type, cast bool, alreadyPointer bool) {
	if t.Literal != nil {
		e.DeserializeTypeLit(r, name, t.Literal, cast, alreadyPointer)
	} else {
		tabs := r.Tabs

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.String(t.Package)
			r.Tabs = 0
			r.Rune('.')
		}

		r.Printf("Deserialize%sJSON(d, &%s)", t.Name, name)
		r.Tabs = tabs
	}
}

func (e *Encoding) SerializeTypeLit(r *Result, name string, lit TypeLit, cast bool, alreadyPointer bool) {
	switch lit := lit.(type) {
	case *Int, *Float, *String:
		var star string
		if alreadyPointer {
			star = "*"
		}

		s := lit.String()
		if !cast {
			r.Printf("s.%c%s(%s%s)", unicode.ToUpper(rune(s[0])), s[1:], star, name)
		} else {
			r.Printf("s.%c%s(%s(%s%s))", unicode.ToUpper(rune(s[0])), s[1:], s, star, name)
		}
	case *Slice:
		e.SerializeSlice(r, name, lit)
	case *Struct:
		e.SerializeStruct(r, name, lit)
	}
}

func (e *Encoding) DeserializeTypeLit(r *Result, name string, lit TypeLit, cast bool, alreadyPointer bool) {
	switch lit := lit.(type) {
	case *Int, *Float, *String:
		var amp string
		if !alreadyPointer {
			amp = "&"
		}

		s := lit.String()
		if !cast {
			r.Printf("d.%c%s(%s%s)", unicode.ToUpper(rune(s[0])), s[1:], amp, name)
		} else {
			r.AddImport("unsafe")
			r.Printf("d.%c%s((*%s)(unsafe.Pointer(%s%s)))", unicode.ToUpper(rune(s[0])), s[1:], s, amp, name)
		}
	case *Slice:
		e.DeserializeSlice(r, name, lit)
	case *Struct:
		r.Line("var key string")
		e.DeserializeStruct(r, name, lit)
	}
}

func (e *Encoding) SerializeStruct(r *Result, name string, s *Struct) {
	r.Line("s.ObjectBegin()")
	e.SerializeStructFields(r, name, s.Fields, nil)
	r.Line("s.ObjectEnd()")
}

func (e *Encoding) DeserializeStruct(r *Result, name string, s *Struct) {
	r.Line("d.ObjectBegin()")
	r.Line("for d.Key(&key) {")
	r.Tabs++
	{
		r.Line("switch key {")
		e.DeserializeStructFields(r, name, s.Fields, nil)
		r.Line("}")
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.ObjectEnd()")
}

func (e *Encoding) SerializeStructFields(r *Result, name string, fields []StructField, forbiddenKeys KeysSet) {
	keys := make(KeysSet)

	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
				continue
			}
		}
		keys[strings.Or(field.Name, field.Type.Name)] = struct{}{}
	}

	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
				continue
			}
		}

		if len(field.Type.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					e.SerializeStructFields(r, fmt.Sprintf("%s.%s", name, field.Type.Name), s.Fields, keys)
					continue
				}
			} else if lit != nil {
				key := field.Type.Name
				if _, ok := forbiddenKeys[key]; !ok {
					r.Printf("s.Key(`%s`)", key)
					e.SerializeTypeLit(r, fmt.Sprintf("%s.%s", name, strings.Or(field.Name, key)), lit, true, false)
				}
				continue
			}
		}

		key := strings.Or(field.Name, field.Type.Name)
		if _, ok := forbiddenKeys[key]; !ok {
			r.Printf("s.Key(`%s`)", key)
			e.SerializeType(r, fmt.Sprintf("%s.%s", name, key), &field.Type, false, false)
		}
	}
}

func (e *Encoding) DeserializeStructFields(r *Result, name string, fields []StructField, forbiddenKeys KeysSet) {
	keys := make(KeysSet)

	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
				continue
			}
		}
		keys[strings.Or(field.Name, field.Type.Name)] = struct{}{}
	}

	for _, field := range fields {
		if field.Tag == `json:"-"` {
			continue
		}
		if len(field.Name) == 0 {
			/* struct { myType } */
			if (len(field.Type.Name) > 0) && (Private(field.Type.Name[0])) {
				continue
			}
			/* struct { int } */
			if (field.Type.Literal != nil) && (Private(field.Type.Literal.String()[0])) {
				continue
			}
		}

		if len(field.Type.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(*Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					e.DeserializeStructFields(r, fmt.Sprintf("%s.%s", name, field.Type.Name), s.Fields, keys)
					continue
				}
			} else if lit != nil {
				key := field.Type.Name
				if _, ok := forbiddenKeys[key]; !ok {
					r.Printf("case \"%s\":", key)
					r.Tabs++
					e.DeserializeTypeLit(r, fmt.Sprintf("%s.%s", name, strings.Or(field.Name, key)), lit, true, false)
					r.Tabs--
				}
				continue
			}
		}

		key := strings.Or(field.Name, field.Type.Name)
		if _, ok := forbiddenKeys[key]; !ok {
			r.Printf("case \"%s\":", key)
			r.Tabs++
			e.DeserializeType(r, fmt.Sprintf("%s.%s", name, key), &field.Type, false, false)
			r.Tabs--
		}
	}
}

func (e *Encoding) SerializeSlice(r *Result, name string, s *Slice) {
	const element = "element"

	r.Line("s.ArrayBegin()")
	r.Printf("for _, element := range %s {", name)
	r.Tabs++
	for {
		if len(s.Element.Name) != 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(s.Element.Package, r.Package), s.Element.Name)
			if lit != nil {
				if !IsStruct(lit) {
					e.SerializeTypeLit(r, element, lit, true, false)
					break
				}
			}
		}
		e.SerializeType(r, element, &s.Element, false, false)
		break
	}
	r.Tabs--
	r.Line("}")
	r.Line("s.ArrayEnd()")
}

func (e *Encoding) DeserializeSlice(r *Result, name string, s *Slice) {
	const element = "element"

	var elementType string
	if len(s.Element.Name) == 0 {
		elementType = s.Element.Literal.String()
	} else {
		if len(s.Element.Package) > 0 {
			r.AddImport(s.Element.Package)
			elementType = s.Element.Package + "."
		}
		elementType += s.Element.Name
	}

	r.Line("d.ArrayBegin()")
	r.Line("for d.Next() {")
	r.Tabs++
	for {
		r.Printf("var element %s", elementType)
		if len(s.Element.Name) > 0 {
			lit := e.FindTypeLit(r.Imports, strings.Or(s.Element.Package, r.Package), s.Element.Name)
			if lit != nil {
				if !IsStruct(lit) {
					r.AddImport("unsafe")
					e.DeserializeTypeLit(r, element, lit, true, false)
					r.Printf("%s = append(%s, %s)", name, name, element)
					break
				}
			}
		}
		e.DeserializeType(r, element, &s.Element, false, false)
		r.Printf("%s = append(%s, element)", name, name)
		break
	}
	r.Tabs--
	r.Line("}")
	r.Line("d.ArrayEnd()")
}

func GeneratorsEncodingAll() []Generator {
	return []Generator{GeneratorJSON{}, GeneratorWire{}}
}
