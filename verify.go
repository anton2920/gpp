package main

import (
	"fmt"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type GeneratorVerify struct{}

func (g GeneratorVerify) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "l10n")
	r.Printf("func Verify%s(l l10n.Language, %s %s) error {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorVerify) Body(r *Result, ctx GenerationContext, t *Type) {
	GenerateType(g, r, ctx, t)
	r.Line("return nil")
}

func (g GeneratorVerify) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("if err := %sVerify%s(l, %s); err != nil {", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
	{
		r.Line("return err")
	}
	r.Line("}")
}

func VerifyWithFuncs(r *Result, ctx GenerationContext, funcs []string) {
	for _, fn := range funcs {
		if !strings.StartsWith(fn, "{") {
			fn = fmt.Sprintf("%s(l, %s)", fn, ctx.Deref(ctx.VarName))
		} else {
			fn = stdstrings.Replace(fn[1:len(fn)-1], "?", ctx.Deref(ctx.VarName), 1)
		}
		r.Printf("if err := %s; err != nil {", fn)
		{
			r.Line("return err")
		}
		r.Line("}")
	}
}

func (g GeneratorVerify) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	vn := VariableName(ctx.FieldName)

	var vc VerifyComment
	for _, comment := range ctx.Comments {
		if c, ok := comment.(VerifyComment); ok {
			strings.Replace(&vc.Min, c.Min)
			strings.Replace(&vc.Max, c.Max)
			strings.Replace(&vc.MinLength, c.MinLength)
			strings.Replace(&vc.MaxLength, c.MaxLength)
			vc.Optional = vc.Optional || c.Optional
			strings.Replace(&vc.Prefix, c.Prefix)
			vc.Required = vc.Required || c.Required
			vc.Funcs = append(vc.Funcs, c.Funcs...)
		}
	}

	switch lit.(type) {
	case Int, Float:
		fieldName := ctx.FieldName
		if strings.StartsWith(ctx.FieldName, "Min") {
			fieldName = fieldName[len("Min"):]
		} else if strings.StartsWith(ctx.FieldName, "Max") {
			fieldName = fieldName[len("Max"):]
		}
		minConst := r.AddConstant(fmt.Sprintf("Min%s%s", strings.Or(vc.Prefix, ctx.SpecName), fieldName), vc.Min)
		maxConst := r.AddConstant(fmt.Sprintf("Max%s%s", strings.Or(vc.Prefix, ctx.SpecName), fieldName), vc.Max)

		if vc.Required {
			r.Printf("if %s == 0 {", ctx.Deref(ctx.VarName))
			{
				r.AddImport(GOFA + "errors")
				r.Printf(`return errors.New("you must spefify %s")`, vn)
			}
			r.Line("}")
			VerifyWithFuncs(r, ctx, vc.Funcs)
		} else {
			if vc.Optional {
				r.Printf("if %s > 0 {", ctx.Deref(ctx.VarName))
			}
			if (len(vc.Min) > 0) && (len(vc.Max) > 0) {
				r.Printf("if (%s < %s) || (%s > %s) {", ctx.Deref(ctx.VarName), minConst.Name, ctx.Deref(ctx.VarName), maxConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d and greater than %%d"), %s, %s)`, vn, minConst.Name, maxConst.Name)
				}
				r.Line("}")
			} else if len(vc.Min) > 0 {
				r.Printf("if %s < %s {", ctx.Deref(ctx.VarName), minConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d"), %s)`, vn, minConst.Name)
				}
				r.Line("}")
			} else if len(vc.Max) > 0 {
				r.Printf("if %s > %s {", ctx.Deref(ctx.VarName), maxConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must be greater than %%d"), %s)`, vn, maxConst.Name)
				}
				r.Line("}")
			}
			VerifyWithFuncs(r, ctx, vc.Funcs)
			if vc.Optional {
				r.Line("}")
			}
		}
	case String:
		if vc.Required {
			r.Printf("if len(%s) == 0 {", ctx.Deref(ctx.VarName))
			{
				r.AddImport(GOFA + "errors")
				r.Printf(`return errors.New("you must spefify %s")`, vn)
			}
			r.Line("}")
			VerifyWithFuncs(r, ctx, vc.Funcs)
		} else {
			if vc.Optional {
				r.Printf("if len(%s) > 0 {", ctx.Deref(ctx.VarName))
			}
			if (len(vc.MinLength) > 0) && (len(vc.MaxLength) > 0) {
				fieldName := ctx.FieldName
				if strings.StartsWith(ctx.FieldName, "Min") {
					fieldName = fieldName[len("Min"):]
				} else if strings.StartsWith(ctx.FieldName, "Max") {
					fieldName = fieldName[len("Max"):]
				}
				minLengthConst := r.AddConstant(fmt.Sprintf("Min%s%sLen", strings.Or(vc.Prefix, ctx.SpecName), fieldName), vc.MinLength)
				maxLengthConst := r.AddConstant(fmt.Sprintf("Max%s%sLen", strings.Or(vc.Prefix, ctx.SpecName), fieldName), vc.MaxLength)

				r.Printf("if !strings.LengthInRange(%s, %s, %s) {", ctx.Deref(ctx.VarName), minLengthConst.Name, maxLengthConst.Name)
				{
					r.AddImport("fmt")
					r.AddImport(GOFA + "strings")
					r.Printf(`return fmt.Errorf(l.L("%s length must be between %%d and %%d characters long"), %s, %s)`, vn, minLengthConst.Name, maxLengthConst.Name)
				}
				r.Line("}")
			}
			VerifyWithFuncs(r, ctx, vc.Funcs)
			if vc.Optional {
				r.Line("}")
			}
		}
	}
}

func (g GeneratorVerify) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorVerify) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	for _, comment := range field.Comments {
		if _, ok := comment.(VerifyComment); ok {
			GenerateStructField(g, r, ctx, field, lit)
			break
		}
	}
}

func (g GeneratorVerify) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorVerify) Array(r *Result, ctx GenerationContext, a *Array) {
}

func (g GeneratorVerify) Slice(r *Result, ctx GenerationContext, s *Slice) {
}

func (g GeneratorVerify) Union(r *Result, ctx GenerationContext, u *Union) {
}
