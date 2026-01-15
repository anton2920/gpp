package main

import (
	"fmt"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct {
	SliceField bool
}

func FillWithFunc(r *Result, ctx GenerationContext, fn string) {
	if !strings.StartsWith(fn, "{") {
		fn = fmt.Sprintf(`%s(vs.Get("%s"))`, fn, ctx.FieldName)
	} else {
		fn = stdstrings.Replace(fn[1:len(fn)-1], "?", ctx.FieldName, 1)
	}
	r.Printf("%s = %s", ctx.VarName, fn)
}

func MergeFillComments(comments []Comment) FillComment {
	var fc FillComment

	for _, comment := range comments {
		if c, ok := comment.(FillComment); ok {
			strings.Replace(&fc.ClampFrom, c.ClampFrom)
			strings.Replace(&fc.ClampTo, c.ClampTo)
			fc.InsertAfter = append(fc.InsertAfter, c.InsertAfter...)
			fc.InsertBefore = append(fc.InsertBefore, c.InsertBefore...)
			fc.Enum = fc.Enum || c.Enum
			strings.Replace(&fc.Func, c.Func)
		}
	}

	return fc
}

func (g GeneratorFillValues) Imports() []string {
	return []string{}
}

func (g GeneratorFillValues) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "net/url")
	r.Printf("func Fill%sFromValues(vs url.Values, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorFillValues) Body(r *Result, ctx GenerationContext, t *Type) {
	/* TODO(anton2920): remove after getting rid of polymorphism. */
	if !IsSlice(t.Literal) {
		fc := MergeFillComments(ctx.Comments)
		Insert(r, ctx, fc.InsertBefore)
		GenerateType(g, r, ctx, t)
		Insert(r, ctx, fc.InsertAfter)
	}
}

func (g GeneratorFillValues) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sFill%sFromValues(vs, %s)", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
}

func (g GeneratorFillValues) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	if g.SliceField {
		g.PrimitiveSlice(r, ctx, lit)
		return
	}

	fc := MergeFillComments(ctx.Comments)
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

			if ((len(ctx.CastName) == 0) || (litName == ctx.CastName)) && (!fc.Enum) && (len(fc.ClampFrom)+len(fc.ClampTo) == 0) {
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

func (g GeneratorFillValues) PrimitiveSlice(r *Result, ctx GenerationContext, lit TypeLit) {
	fc := MergeFillComments(ctx.Comments)
	vn := VariableName(ctx.SpecName)

	/* TODO(anton2920): handle Func with {?}. */
	switch lit.(type) {
	case Bool:

	case Int, Float:
		const tmp = "tmp"

		litName := lit.String()
		r.Line("{")
		{
			i := ctx.LoopVar()
			r.Printf(`%s := vs.GetMany("%s")`, tmp, ctx.FieldName)
			r.Printf("%s = %s[:0]", ctx.Deref(ctx.VarName), ctx.Deref(ctx.VarName))
			r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, tmp, i)
			{
				var app string

				if (!fc.Enum) && (len(fc.ClampFrom)+len(fc.ClampTo) == 0) {
					fn := fc.Func
					if len(fc.Func) == 0 {
						r.AddImport(GOFA + "strings")
						fn = "strings.ToInt"
					}
					if (len(ctx.CastName) == 0) || (litName == ctx.CastName) {
						app = fmt.Sprintf("%s(%s[%s])", fn, tmp, i)
					} else {
						app = fmt.Sprintf("%s(%s(%s(%s[%s])))", ctx.CastName, fn, litName, tmp, i)
					}
				} else if fc.Enum {
					r.AddImport(GOFA + "ints")
					app = fmt.Sprintf("%s(ints.Clamp(int(%s), 1, int(%sCount)))", ctx.CastName, tmp, ctx.CastName)
				} else if len(fc.ClampFrom)+len(fc.ClampTo) > 0 {
					r.AddImport(GOFA + "ints")
					app = fmt.Sprintf("%s(ints.Clamp(int(%s), %s, %s))", litName, tmp, PrependVariableName(fc.ClampFrom, vn), PrependVariableName(fc.ClampTo, vn))
				}

				r.Printf("%s[%s], _ = append(%s[%s], %s)", ctx.Deref(ctx.VarName), i, ctx.Deref(ctx.VarName), i, app)
			}
			r.Line("}")
		}
		r.Line("}")
	case String:
		r.Printf(`%s = vs.GetMany("%s")`, ctx.VarName, ctx.FieldName)
	}
}

func (g GeneratorFillValues) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorFillValues) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	GenerateStructField(g, r, ctx.WithCast(field.Type.String()), field, lit)
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
	g.SliceField = true
	elem := &s.Element
	if len(elem.Name) > 0 {
		lit := ctx.FindTypeLit(r.Imports, strings.Or(elem.Package, r.Package), elem.Name)
		if (lit != nil) && (IsPrimitive(lit)) {
			GenerateTypeLit(g, r, ctx.WithCast(elem.String()), lit)
			return
		}
	}
	GenerateType(g, r, ctx, elem)
	g.SliceField = false
}

func (g GeneratorFillValues) Union(r *Result, ctx GenerationContext, u *Union) {
}
