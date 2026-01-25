package main

import (
	"fmt"
	"os"
	stdstrings "strings"
	"unicode"

	"github.com/anton2920/gofa/bools"
	"github.com/anton2920/gofa/strings"
)

type GeneratorFillValues struct {
	SliceField bool
}

type FillComment struct {
	Each *FillComment

	Func string

	InsertAfter  []string
	InsertBefore []string

	Enum     bool
	NOP      bool
	Optional bool
}

func (FillComment) Comment() {}

func MergeFillComments(comments []Comment) FillComment {
	var fc FillComment

	for _, comment := range comments {
		if c, ok := comment.(FillComment); ok {
			if c.Each != nil {
				fc.Each = c.Each
			}

			strings.Replace(&fc.Func, c.Func)

			fc.InsertAfter = append(fc.InsertAfter, c.InsertAfter...)
			fc.InsertBefore = append(fc.InsertBefore, c.InsertBefore...)

			fc.Enum = fc.Enum || c.Enum
			fc.Optional = fc.Optional || c.Optional

		}
	}

	return fc
}

func ParseFillComment(comment string, fc *FillComment) bool {
	var done bool

	for !done {
		s, rest, ok := ProperCut(comment, ",", LBracks, RBracks, LBraces, RBraces)
		if !ok {
			done = true
		}
		s = strings.TrimSpace(s)

		switch stdstrings.ToLower(s) {
		case "enum":
			fc.Enum = true
		case "nop":
			fc.NOP = true
		case "optional":
			fc.Optional = true
		default:
			lval, rval, ok := strings.Cut(s, "=")
			if ok {
				lval = stdstrings.ToLower(strings.TrimSpace(lval))
				rval = strings.TrimSpace(rval)

				switch lval {
				case "each":
					var each FillComment
					rval, _ = StripIfFound(rval, LBracks, RBracks)
					if ParseFillComment(rval, &each) {
						fc.Each = &each
					}
				case "insertafter":
					fc.InsertAfter = append(fc.InsertAfter, rval)
				case "insertbefore":
					fc.InsertBefore = append(fc.InsertBefore, rval)
				case "func":
					fc.Func = rval
				}
			}
		}

		comment = rest
	}

	if (fc.Optional) && (!fc.Enum) {
		fmt.Fprintf(os.Stderr, "WARNING: ignoring 'Optional' without 'Enum'")
		fc.Optional = false
	}

	return true
}

func FillWithFunc(r *Result, ctx GenerationContext, fn string) {
	var ok bool

	v := fmt.Sprintf(`vs.Get("%s")`, ctx.FieldName)
	if fn, ok = StripIfFound(fn, LBraces, RBraces); !ok {
		fn = fmt.Sprintf(`%s(%s)`, fn, v)
	} else {
		fn = PrependVariableName(stdstrings.Replace(fn, "?", v, 1), VariableName(ctx.SpecName))
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

			if (!fc.Enum) && ((len(ctx.CastName) == 0) || (litName == ctx.CastName)) {
				r.Printf(`%s, _ = vs.Get%c%s("%s")`, ctx.Deref(ctx.VarName), unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
			} else {
				const tmp = "tmp"

				r.Line("{")
				{
					r.Printf(`%s, _ := vs.Get%c%s("%s")`, tmp, unicode.ToUpper(rune(litName[0])), litName[1:], ctx.FieldName)
					if !fc.Enum {
						r.Printf("%s = %s(%s)", ctx.VarName, ctx.CastName, tmp)
					} else {
						r.AddImport(GOFA + "ints")
						r.Printf("%s = %s(ints.Clamp(int(%s), %d, int(%sCount)))", ctx.Deref(ctx.VarName), ctx.CastName, tmp, bools.ToInt(!fc.Optional), ctx.CastName)
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

	switch lit.(type) {
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

				if fc.Enum {
					r.AddImport(GOFA + "ints")
					app = fmt.Sprintf("%s(ints.Clamp(strings.ToInt(%s[%s]), 1, int(%sCount)))", ctx.CastName, tmp, i, ctx.CastName)
				} else {
					/* TODO(anton2920): handle Func with {{?}}. */
					fn := fc.Func
					if len(fc.Func) == 0 {
						r.AddImport(GOFA + "strings")
						if _, ok := lit.(Int); ok {
							fn = "strings.ToInt"
						} else {
							fn = "strings.ToFloat"
						}
					}
					if (len(ctx.CastName) == 0) || (litName == ctx.CastName) {
						app = fmt.Sprintf("%s(%s[%s])", fn, tmp, i)
					} else {
						app = fmt.Sprintf("%s(%s(%s[%s]))", ctx.CastName, fn, tmp, i)
					}
				}

				r.Printf("%s = append(%s, %s)", ctx.Deref(ctx.VarName), ctx.Deref(ctx.VarName), app)
			}
			r.Line("}")
		}
		r.Line("}")
	case String:
		/* TODO(anton2920): handle Func. */
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
		lit := ctx.Parser.FindTypeLit(r.File.Imports, strings.Or(elem.Package, r.File.Package), elem.Name)
		if (lit != nil) && (IsPrimitive(lit)) {
			GenerateTypeLit(g, r, ctx.WithCast(elem.String()), lit)
			return
		}
	}
	GenerateType(g, r, ctx, elem)
	g.SliceField = false
}

func (g GeneratorFillValues) Union(r *Result, ctx GenerationContext, u *Union) {
	fc := MergeFillComments(ctx.Comments)

	r.Printf("switch %s := %s.(type) {", ctx.VarName, ctx.Deref(ctx.VarName))
	{
		for _, name := range u.Types {
			var star string
			if name[0] != '*' {
				ctx.Autoderef = true
			} else {
				ctx.Autoderef = false
				name = name[1:]
				star = "*"
			}
			t := Type{Name: name}

			r.Printf("case %s%s:", star, t)
			{
				if name != "nil" {
					if fc.Each != nil {
						Insert(r, ctx, fc.Each.InsertBefore)
					}
					g.NamedType(r, ctx, &t)
					if fc.Each != nil {
						Insert(r, ctx, fc.Each.InsertAfter)
					}
				}
			}
			r.Tabs--
		}
	}
	r.Tabs++
	r.Line("}")
}
