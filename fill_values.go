package main

import (
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct{}

func (g GeneratorFillValues) Imports() []string {
	return []string{}
}

func (g GeneratorFillValues) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "net/url")
	r.Printf("func Fill%sFromValues(vs url.Values, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorFillValues) Body(r *Result, ctx GenerationContext, t *Type) {
	GenerateType(g, r, ctx, t)
}

func (g GeneratorFillValues) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sFill%sFromValues(vs, &%s)", t.PackagePrefix(), t.Name, ctx.VarName)
}

func (g GeneratorFillValues) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	var fc FillComment
	for _, comment := range ctx.Comments {
		if c, ok := comment.(FillComment); ok {
			fc.Enum = fc.Enum || c.Enum
			strings.Replace(&fc.Func, c.Func)
		}
	}

	switch lit := lit.(type) {
	case Int, Float:
		litName := lit.String()
		if (len(ctx.CastName) == 0) || (litName == ctx.CastName) {
			r.Printf(`%s, _ = vs.Get%c%s("%s")`, ctx.VarName, unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
		} else {
			const tmp = "tmp"

			r.Line("{")
			{
				r.Printf(`%s, _ := vs.Get%c%s("%s")`, tmp, unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
				if !fc.Enum {
					r.Printf("%s = %s(%s)", ctx.VarName, ctx.CastName, tmp)
				} else {
					r.AddImport(GOFA + "ints")
					r.Printf("%s = %s(ints.Clamp(int(%s), 1, int(%sCount)))", ctx.VarName, ctx.CastName, tmp, ctx.CastName)
				}
			}
			r.Line("}")
		}
	case String:
		if len(fc.Func) == 0 {
			r.Printf(`%s = vs.Get("%s")`, ctx.VarName, ctx.FieldName)
		} else {
			r.Printf(`%s, _ = %s(vs.Get("%s"))`, ctx.VarName, fc.Func, ctx.FieldName)
		}
	}
}

func (g GeneratorFillValues) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorFillValues) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	GenerateStructField(g, r, ctx, field, lit)
}

func (g GeneratorFillValues) StructFieldSkip(field *StructField) bool {
	for _, comment := range field.Comments {
		if fc, ok := comment.(FillComment); ok {
			if fc.NOP {
				return true
			}
		}
	}
	return false
}

func (g GeneratorFillValues) Array(r *Result, ctx GenerationContext, a *Array) {
}

func (g GeneratorFillValues) Slice(r *Result, ctx GenerationContext, s *Slice) {
}

func (g GeneratorFillValues) Union(r *Result, ctx GenerationContext, u *Union) {
}
