package main

import (
	"fmt"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type KeySet map[string]struct{}

type GenerationContext struct {
	*Parser

	SpecName  string
	FieldName string
	CastName  string
	VarName   string
	Autoderef bool

	Level    int
	Comments []Comment
}

type Generator interface {
	Decl(r *Result, ctx GenerationContext, t *Type)
	Body(r *Result, ctx GenerationContext, t *Type)

	NamedType(r *Result, ctx GenerationContext, t *Type)
	Primitive(r *Result, ctx GenerationContext, lit TypeLit)

	Struct(r *Result, ctx GenerationContext, s *Struct)
	StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit)
	StructFieldSkip(field *StructField) bool

	Array(r *Result, ctx GenerationContext, a *Array)
	Slice(r *Result, ctx GenerationContext, s *Slice)

	Union(r *Result, ctx GenerationContext, u *Union)
}

/* NOTE(anton2920): this supports only ASCII. */
func VariableName(typeName string) string {
	if typeName == stdstrings.ToUpper(typeName) {
		return stdstrings.ToLower(typeName)
	}

	var lastUpper int
	for i := 0; i < len(typeName); i++ {
		if unicode.IsUpper(rune(typeName[i])) {
			lastUpper = i
		}
	}

	return fmt.Sprintf("%c%s", unicode.ToLower(rune(typeName[lastUpper])), typeName[lastUpper+1:])
}

/* NOTE(anton2920): this supports only ASCII. */
func FieldName2Description(fieldName string) string {
	if fieldName == stdstrings.ToUpper(fieldName) {
		return stdstrings.ToLower(fieldName)
	}

	words := make([]string, 0, 16)
	var lastUpper int

	for i := 1; i < len(fieldName); i++ {
		if unicode.IsUpper(rune(fieldName[i])) {
			words = append(words, stdstrings.ToLower(fieldName[lastUpper:i]))
			lastUpper = i
		}
	}
	words = append(words, stdstrings.ToLower(fieldName[lastUpper:len(fieldName)]))

	return stdstrings.Join(words, " ")
}

func PrependVariableName(s string, vn string) string {
	for dot := 0; dot < len(s); dot++ {
		period := strings.FindChar(s[dot:], '.')
		if period == -1 {
			break
		}
		dot += period

		if (dot == 0) || (s[dot-1] == ' ') || (s[dot-1] == '(') || (s[dot-1] == '[') || (s[dot-1] == '{') || (s[dot-1] == '\t') {
			s = s[:dot] + vn + s[dot:]
			dot += len(vn)
		}
	}
	return s
}

func Insert(r *Result, ctx GenerationContext, inserts []string) {
	for _, insert := range inserts {
		if insert, ok := StripIfFound(insert, LCompound, RCompound); ok {
			lines := stdstrings.Split(PrependVariableName(insert, VariableName(ctx.SpecName)), "\n")
			tabs := r.Tabs
			for i := 0; i < len(lines); i++ {
				if strings.EndsWith(lines[i], "}") {
					r.Tabs++
				}
				r.Line(lines[i])
				r.Tabs = tabs
			}
		}
	}
}

func Generate(g Generator, r *Result, p *Parser, ts *TypeSpec) {
	var ctx GenerationContext

	ctx.Parser = p
	ctx.Comments = ts.Comments

	{
		ctx.SpecName = ts.Name
		ctx.VarName = VariableName(ts.Name)
		ctx.Autoderef = IsStruct(ts.Type.Literal)

		r.Rune('\n')
		g.Decl(r, ctx, &Type{Literal: Pointer{BaseType: Type{Name: ts.Name}}})
		{
			g.Body(r, ctx, &ts.Type)
		}
		r.Line("}")
	}

	{
		ctx.SpecName = Plural(ts.Name)
		ctx.VarName = Plural(VariableName(ts.Name))
		ctx.Autoderef = false

		t := Type{Literal: Slice{Element: Type{Name: ts.Name}}}
		r.Rune('\n')
		g.Decl(r, ctx, &t)
		{
			g.Body(r, ctx, &t)
		}
		r.Line("}")
	}
}

func GenerateType(g Generator, r *Result, ctx GenerationContext, t *Type) {
	if t.Literal != nil {
		GenerateTypeLit(g, r, ctx, t.Literal)
	} else {
		r.AddImport(t.Package)
		g.NamedType(r, ctx.WithAutoderef(true), t)
	}
}

func GenerateTypeLit(g Generator, r *Result, ctx GenerationContext, lit TypeLit) {
	switch lit := lit.(type) {
	case Bool, Int, Float, Pointer, String:
		g.Primitive(r, ctx, lit)
	case Array:
		g.Array(r, ctx, &lit)
	case Slice:
		g.Slice(r, ctx, &lit)
	case Struct:
		g.Struct(r, ctx, &lit)
	case Union:
		g.Union(r, ctx, &lit)
	}
}

func Private(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}

func StructFieldSkip(g Generator, field *StructField) bool {
	if g.StructFieldSkip(field) {
		return true
	}

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

func GenerateStructFields(g Generator, r *Result, ctx GenerationContext, fields []StructField, forbiddenFields KeySet) {
	currentFields := make(KeySet)
	for field := range forbiddenFields {
		currentFields[field] = struct{}{}
	}
	for _, field := range fields {
		if StructFieldSkip(g, &field) {
			continue
		}
		fieldName := strings.Or(field.Name, field.Type.Name)
		currentFields[fieldName] = struct{}{}
	}

	for _, field := range fields {
		if StructFieldSkip(g, &field) {
			continue
		}

		ctx.FieldName = strings.Or(field.Name, field.Type.Name)
		name := fmt.Sprintf("%s.%s", ctx.VarName, ctx.FieldName)

		var lit TypeLit
		if (field.Type.Literal == nil) && (len(field.Type.Name) > 0) {
			lit = ctx.FindTypeLit(r.Imports, strings.Or(field.Type.Package, r.Package), field.Type.Name)
			if s, ok := lit.(Struct); ok {
				if len(field.Name) == 0 {
					for i := 0; i < len(s.Fields); i++ {
						f := &s.Fields[i]
						if len(f.Type.Package) == 0 {
							f.Type.Package = field.Type.Package
						}
					}
					GenerateStructFields(g, r, ctx.WithVar(name).WithComments(field.Comments), s.Fields, currentFields)
					continue
				}
			}
		}

		if _, ok := forbiddenFields[ctx.FieldName]; !ok {
			g.StructField(r, ctx.WithVar(name).WithComments(field.Comments), &field, lit)
			if forbiddenFields != nil {
				forbiddenFields[ctx.FieldName] = struct{}{}
			}
		}
	}
}

func GenerateStructField(g Generator, r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	if lit != nil {
		GenerateTypeLit(g, r, ctx, lit)
	} else {
		GenerateType(g, r, ctx.WithCast(""), &field.Type)
	}
}

func GenerateArrayElement(g Generator, r *Result, ctx GenerationContext, elem *Type) {
	if len(elem.Name) > 0 {
		lit := ctx.FindTypeLit(r.Imports, strings.Or(elem.Package, r.Package), elem.Name)
		if (lit != nil) && (IsPrimitive(lit)) {
			GenerateTypeLit(g, r, ctx.WithCast(lit.String()), lit)
			return
		}
	}
	GenerateType(g, r, ctx, elem)
}

func (ctx *GenerationContext) LoopVar() string {
	i := fmt.Sprintf("i%d", ctx.Level)
	ctx.Level++
	return i
}

func (ctx GenerationContext) WithCast(castName string) GenerationContext {
	nctx := ctx
	nctx.CastName = castName
	return nctx
}

func (ctx GenerationContext) WithVar(format string, args ...interface{}) GenerationContext {
	nctx := ctx
	nctx.VarName = fmt.Sprintf(format, args...)
	return nctx
}

func (ctx GenerationContext) WithComments(comments []Comment) GenerationContext {
	nctx := ctx
	nctx.Comments = append(nctx.Comments, comments...)
	return nctx
}

func (ctx GenerationContext) WithAutoderef(state bool) GenerationContext {
	nctx := ctx
	nctx.Autoderef = state
	return nctx
}

func (ctx GenerationContext) Deref(s string) string {
	if ctx.Autoderef {
		return s

	}
	return fmt.Sprintf("(*%s)", s)
}

func (ctx GenerationContext) AddrOf(s string) string {
	if !ctx.Autoderef {
		return s
	}
	return fmt.Sprintf("&%s", s)
}

func GeneratorsAll() []Generator {
	return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
}
