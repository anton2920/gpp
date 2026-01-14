package main

import (
	"fmt"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct{}

func FillInsert(r *Result, ctx GenerationContext, insert string) {
	if strings.StartsEndsWith(insert, "{", "}") {
		lines := stdstrings.Split(PrependVariableName(insert[1:len(insert)-1], VariableName(ctx.SpecName)), "\n")
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

func FillWithFunc(r *Result, ctx GenerationContext, fn string) {
	if !strings.StartsWith(fn, "{") {
		fn = fmt.Sprintf(`%s(vs.Get("%s"))`, fn, ctx.FieldName)
	} else {
		fn = stdstrings.Replace(fn[1:len(fn)-1], "?", ctx.FieldName, 1)
	}
	r.Printf("%s = %s", ctx.VarName, fn)
}

func (g GeneratorFillValues) Imports() []string {
	return []string{}
}

func (g GeneratorFillValues) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "net/url")
	r.Printf("func Fill%sFromValues(vs url.Values, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorFillValues) Body(r *Result, ctx GenerationContext, t *Type) {
	var fc FillComment

	if !IsSlice(t.Literal) {
		for _, comment := range ctx.Comments {
			if c, ok := comment.(FillComment); ok {
				/* NOTE(anton2920): it multiple insert sources are needed, switch to []string. */
				strings.Replace(&fc.InsertAfter, c.InsertAfter)
				strings.Replace(&fc.InsertBefore, c.InsertBefore)

				fc.Enum = fc.Enum || c.Enum
				strings.Replace(&fc.Func, c.Func)
			}
		}
	}

	FillInsert(r, ctx, fc.InsertBefore)
	GenerateType(g, r, ctx, t)
	FillInsert(r, ctx, fc.InsertAfter)
}

func (g GeneratorFillValues) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sFill%sFromValues(vs, %s)", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
}

func (g GeneratorFillValues) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	var fc FillComment
	for _, comment := range ctx.Comments {
		if c, ok := comment.(FillComment); ok {
			strings.Replace(&fc.ClampFrom, c.ClampFrom)
			strings.Replace(&fc.ClampTo, c.ClampTo)

			/* NOTE(anton2920): it multiple insert sources are needed, switch to []string. */
			strings.Replace(&fc.InsertAfter, c.InsertAfter)
			strings.Replace(&fc.InsertBefore, c.InsertBefore)

			fc.Enum = fc.Enum || c.Enum
			strings.Replace(&fc.Func, c.Func)
		}
	}

	if len(fc.Func) > 0 {
		FillWithFunc(r, ctx, fc.Func)
	} else {
		switch lit := lit.(type) {
		case Bool:
			r.Line("{")
			{
				const tmp = "tmp"

				r.Printf(`%s := vs.Get("%s")`, tmp, ctx.FieldName)
				r.Printf(`%s = (%s == "on")`, ctx.VarName, tmp)
			}
			r.Line("}")
		case Int, Float:
			litName := lit.String()

			if ((len(ctx.CastName) == 0) || (litName == ctx.CastName)) && (len(fc.ClampFrom)+len(fc.ClampTo) == 0) {
				r.Printf(`%s, _ = vs.Get%c%s("%s")`, ctx.Deref(ctx.VarName), unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
			} else {
				const tmp = "tmp"

				r.Line("{")
				{
					r.Printf(`%s, _ := vs.Get%c%s("%s")`, tmp, unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
					if len(fc.ClampFrom)+len(fc.ClampTo) > 0 {
						vn := VariableName(ctx.SpecName)
						r.AddImport(GOFA + "ints")
						r.Printf("%s = %s(ints.Clamp(int(%s), %s, %s))", ctx.Deref(ctx.VarName), litName, tmp, PrependVariableName(fc.ClampFrom, vn), PrependVariableName(fc.ClampTo, vn))
					} else {
						if !fc.Enum {
							r.Printf("%s = %s(%s)", ctx.VarName, ctx.CastName, tmp)
						} else {
							r.AddImport(GOFA + "ints")
							r.Printf("%s = %s(ints.Clamp(int(%s), 1, int(%sCount)))", ctx.Deref(ctx.VarName), ctx.CastName, tmp, ctx.CastName)
						}
					}
				}
				r.Line("}")
			}
		case String:
			r.Printf(`%s = vs.Get("%s")`, ctx.VarName, ctx.FieldName)
		}
	}
}

func (g GeneratorFillValues) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorFillValues) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	/* TODO(anton2920): kolhozno! */
	if (lit != nil) && (!IsStruct(lit)) {
		ctx = ctx.WithCast(field.Type.Name)
	}
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
