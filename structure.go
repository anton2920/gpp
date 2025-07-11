package main

import (
	"bytes"
	"fmt"
	"go/token"
)

type StructField struct {
	Name     string
	Type     Type
	Tag      string
	Embedded bool
}

type Struct struct {
	Name     string
	Fields   []StructField
	TypeArgs []TypeArg
}

func (s *Struct) String() string {
	var buf bytes.Buffer

	if len(s.Name) > 0 {
		buf.WriteString(s.Name)
	} else {
		buf.WriteString(`struct`)
	}
	if len(s.TypeArgs) > 0 {
		buf.WriteString(`[`)
		for i := 0; i < len(s.TypeArgs); i++ {
			arg := &s.TypeArgs[i]

			if i > 0 {
				buf.WriteString(`, `)
			}
			buf.WriteString(arg.Name)
			buf.WriteRune(' ')
			buf.WriteString(arg.Type.String())
		}
		buf.WriteRune(']')
	}
	buf.WriteString(" {\n")
	for i := 0; i < len(s.Fields); i++ {
		field := &s.Fields[i]

		buf.WriteRune('\t')
		if (len(field.Name) > 0) && (!field.Embedded) {
			buf.WriteString(field.Name)
			buf.WriteRune(' ')
		}
		buf.WriteString(field.Type.String())
		if len(field.Tag) > 0 {
			buf.WriteRune(' ')
			buf.WriteString(field.Tag)
		}
		buf.WriteRune('\n')
	}
	buf.WriteRune('}')

	return buf.String()
}

func ParseStructFields(l *Lexer, fs *[]StructField) bool {
	pos := l.Position

	/* Option 1: IdentList Type. */
	var idents []string
	if ParseIdentList(l, &idents) {
		var t Type
		if ParseType(l, &t) {
			var tag string
			ParseString(l, &tag)
			l.Error = nil

			for i := 0; i < len(idents); i++ {
				*fs = append(*fs, StructField{Name: idents[i], Type: t, Tag: tag})
			}

			if ParseToken(l, token.SEMICOLON) {
				return true
			}
		}
	}

	l.Position = pos
	l.Error = nil

	/* Option 2: Type. */
	var t Type
	if ParseType(l, &t) {
		var tag string
		ParseString(l, &tag)
		l.Error = nil

		*fs = append(*fs, StructField{Name: t.Name, Type: t, Tag: tag, Embedded: true})

		if ParseToken(l, token.SEMICOLON) {
			return true
		}
	}

	return false
}

/* TODO(anton2920): split into 'ParseTypedecl' and 'ParseStruct'. */
func ParseStruct(l *Lexer, s *Struct) bool {
	if ParseToken(l, token.TYPE) {
		if ParseIdent(l, &s.Name) {
			/* Optional TypeArgs. */
			if ParseToken(l, token.LBRACK) {
				var idents []string
				if ParseIdentList(l, &idents) {
					var t Type
					if ParseType(l, &t) {
						if ParseToken(l, token.RBRACK) {
							for i := 0; i < len(idents); i++ {
								s.TypeArgs = append(s.TypeArgs, TypeArg{Name: idents[i], Type: t})
							}
						}
					}
				}
			}
			l.Error = nil

			if ParseToken(l, token.STRUCT) {
				if ParseToken(l, token.LBRACE) {
					for l.Peek().GoToken != token.RBRACE {
						if !ParseStructFields(l, &s.Fields) {
							return false
						}
					}
					fmt.Println(s)
					return true
				}
			}
		}
	}
	return false
}
