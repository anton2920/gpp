package main

import (
	"fmt"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type GeneratorVerify struct {
	LoopVariable string
	SOA          bool
}

func (g GeneratorVerify) NewConstant(r *Result, specName string, fieldName string, prefix string, value string, format string) Constant {
	var c Constant

	if strings.StartsWith(fieldName, "Min") {
		fieldName = fieldName[len("Min"):]
	} else if strings.StartsWith(fieldName, "Max") {
		fieldName = fieldName[len("Max"):]
	}
	fieldName = Singular(fieldName)

	if expr, ok := StripIfFound(value, LCompound, RCompound); !ok {
		c = r.AddConstant(fmt.Sprintf(format, strings.Or(prefix, specName), fieldName), value)
	} else {
		vn := VariableName(specName)
		name := PrependVariableName(expr, vn)

		if (g.SOA) && (name != expr) && (name == Singular(name)) {
			name = fmt.Sprintf("%s[%s]", Plural(name), g.LoopVariable)
		}
		c = Constant{Name: name}
	}

	return c
}

func VerifyWithFuncs(r *Result, ctx GenerationContext, funcs []string) {
	var ok bool

	for _, fn := range funcs {
		if fn, ok = StripIfFound(fn, LCompound, RCompound); !ok {
			fn = fmt.Sprintf("%s(l, %s)", fn, ctx.Deref(ctx.VarName))
		} else {
			fn = stdstrings.Replace(fn, "?", ctx.Deref(ctx.VarName), 1)
		}
		r.Printf("if err := %s; err != nil {", fn)
		{
			r.Line("return err")
		}
		r.Line("}")
	}
}

func MergeVerifyComments(comments []Comment) VerifyComment {
	var vc VerifyComment

	for _, comment := range comments {
		if c, ok := comment.(VerifyComment); ok {
			vc.InsertAfter = append(vc.InsertAfter, c.InsertAfter...)
			vc.InsertBefore = append(vc.InsertBefore, c.InsertBefore...)
			vc.Funcs = append(vc.Funcs, c.Funcs...)
			strings.Replace(&vc.Min, c.Min)
			strings.Replace(&vc.Max, c.Max)
			strings.Replace(&vc.MinLength, c.MinLength)
			strings.Replace(&vc.MaxLength, c.MaxLength)
			vc.Optional = vc.Optional || c.Optional
			strings.Replace(&vc.Prefix, c.Prefix)
			vc.Required = vc.Required || c.Required
			vc.SOA = vc.SOA || c.SOA
			strings.Replace(&vc.SOAPrefix, c.SOAPrefix)
		}
	}

	return vc
}

func (g GeneratorVerify) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "l10n")
	r.Printf("func Verify%s(l l10n.Language, %s %s) error {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorVerify) Body(r *Result, ctx GenerationContext, t *Type) {
	if !IsSlice(t.Literal) {
		vc := MergeVerifyComments(ctx.Comments)
		g.SOA = vc.SOA

		Insert(r, ctx, vc.InsertBefore)
		GenerateType(g, r, ctx, t)
		Insert(r, ctx, vc.InsertAfter)
	}
	r.Line("return nil")
}

func (g GeneratorVerify) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("if err := %sVerify%s(l, %s); err != nil {", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
	{
		r.Line("return err")
	}
	r.Line("}")
}

func (g GeneratorVerify) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	var loopVariable string

	vc := MergeVerifyComments(ctx.Comments)
	fieldDescription := FieldName2Description(ctx.FieldName)
	if g.SOA {
		fieldDescription = fmt.Sprintf("%s %%d: %s", strings.Or(FieldName2Description(vc.SOAPrefix), Singular(FieldName2Description(ctx.SpecName))), Singular(fieldDescription))
		loopVariable = fmt.Sprintf(", %s+1", g.LoopVariable)
	}

	switch lit.(type) {
	case Int, Float:
		minConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.Min, "Min%s%s")
		maxConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.Max, "Max%s%s")

		if vc.Required {
			r.Printf("if %s == 0 {", ctx.Deref(ctx.VarName))
			{
				r.AddImport(GOFA + "errors")
				r.Printf(`return errors.New("you must spefify %s")`, fieldDescription)
			}
			r.Line("}")
			VerifyWithFuncs(r, ctx, vc.Funcs)
		} else {
			if vc.Optional {
				r.Printf("if %s != 0 {", ctx.Deref(ctx.VarName))
			}
			if (len(vc.Min) > 0) && (len(vc.Max) > 0) {
				r.Printf("if (%s < %s) || (%s > %s) {", ctx.Deref(ctx.VarName), minConst.Name, ctx.Deref(ctx.VarName), maxConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%v and greater than %%v")%s, %s, %s)`, fieldDescription, loopVariable, minConst.Name, maxConst.Name)
				}
				r.Line("}")
			} else if len(vc.Min) > 0 {
				r.Printf("if %s < %s {", ctx.Deref(ctx.VarName), minConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%v")%s, %s)`, fieldDescription, loopVariable, minConst.Name)
				}
				r.Line("}")
			} else if len(vc.Max) > 0 {
				r.Printf("if %s > %s {", ctx.Deref(ctx.VarName), maxConst.Name)
				{
					r.AddImport("fmt")
					r.Printf(`return fmt.Errorf(l.L("%s must be greater than %%v")%s, %s)`, fieldDescription, loopVariable, maxConst.Name)
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
				r.Printf(`return errors.New("you must spefify %s")`, fieldDescription)
			}
			r.Line("}")
			VerifyWithFuncs(r, ctx, vc.Funcs)
		} else {
			if vc.Optional {
				r.Printf("if len(%s) != 0 {", ctx.Deref(ctx.VarName))
			}
			if (len(vc.MinLength) > 0) && (len(vc.MaxLength) > 0) {
				minLengthConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.MinLength, "Min%s%sLen")
				maxLengthConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.MaxLength, "Max%s%sLen")

				r.Printf("if !strings.LengthInRange(%s, %s, %s) {", ctx.Deref(ctx.VarName), minLengthConst.Name, maxLengthConst.Name)
				{
					r.AddImport("fmt")
					r.AddImport(GOFA + "strings")
					r.Printf(`return fmt.Errorf(l.L("%s length must be between %%d and %%d characters long")%s, %s, %s)`, fieldDescription, loopVariable, minLengthConst.Name, maxLengthConst.Name)
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
	soa := (g.SOA) && len(s.Fields) > 0
	if soa {
		field0 := fmt.Sprintf("%s.%s", ctx.VarName, s.Fields[0].Name)

		/* NOTE(anton2920): this ignores banned fields. */
		if len(s.Fields) >= 2 {
			field1 := fmt.Sprintf("%s.%s", ctx.VarName, s.Fields[1].Name)

			tabs := r.Tabs
			r.Printf("if (len(%s) != len(%s))", field0, field1)
			r.Backspace()
			r.Tabs = 0
			for _, field := range s.Fields[2:] {
				r.Printf(" || (len(%s) != len(%s.%s))", field0, ctx.VarName, field.Name)
				r.Backspace()
			}
			r.Line(" {")
			r.Tabs = tabs + 1
			{
				r.AddImport(GOFA + "errors")
				r.AddImport(GOFA + "l10n")
				r.Printf(`return errors.New(l.L("number of all fields should be the same"))`)
			}
			r.Line("}")
		}

		i := ctx.LoopVar()
		g.LoopVariable = i
		r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, field0, i)
	}
	GenerateStructFields(g, r, ctx, s.Fields, nil)
	if soa {
		r.Line("}")
	}
}

func (g GeneratorVerify) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	for _, comment := range field.Comments {
		if _, ok := comment.(VerifyComment); ok {
			if g.SOA {
				ctx = ctx.WithVar("%s[%s]", ctx.VarName, g.LoopVariable)
			}
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
	GenerateArrayElement(g, r, ctx, &s.Element)
}

func (g GeneratorVerify) Union(r *Result, ctx GenerationContext, u *Union) {
}
