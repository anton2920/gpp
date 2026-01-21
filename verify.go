package main

import (
	"fmt"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type GeneratorVerify struct {
	LoopVariable string
	SOAPrefix    string
	SOA          bool
}

type VerifyComment struct {
	Each *VerifyComment

	Funcs        []string
	InsertAfter  []string
	InsertBefore []string

	Max       string
	MaxLength string
	Min       string
	MinLength string
	Prefix    string
	SOAPrefix string

	Optional bool
	Required bool
	SOA      bool
	Unique   bool
}

func (VerifyComment) Comment() {}

func MergeVerifyComments(comments []Comment) VerifyComment {
	var vc VerifyComment

	for _, comment := range comments {
		if c, ok := comment.(VerifyComment); ok {
			if c.Each != nil {
				vc.Each = c.Each
			}

			strings.Replace(&vc.Max, c.Max)
			strings.Replace(&vc.MaxLength, c.MaxLength)
			strings.Replace(&vc.Min, c.Min)
			strings.Replace(&vc.MinLength, c.MinLength)
			strings.Replace(&vc.Prefix, c.Prefix)
			strings.Replace(&vc.SOAPrefix, c.SOAPrefix)

			vc.Funcs = append(vc.Funcs, c.Funcs...)
			vc.InsertAfter = append(vc.InsertAfter, c.InsertAfter...)
			vc.InsertBefore = append(vc.InsertBefore, c.InsertBefore...)

			vc.Optional = vc.Optional || c.Optional
			vc.Required = vc.Required || c.Required
			vc.SOA = vc.SOA || c.SOA
			vc.Unique = vc.Unique || c.Unique
		}
	}

	return vc
}

func ParseVerifyComment(comment string, vc *VerifyComment) bool {
	var done bool
	for !done {
		s, rest, ok := ProperCut(comment, ",", LBracks, RBracks, LBraces, RBraces)
		if !ok {
			done = true
		}
		s = strings.TrimSpace(s)

		lval, rval, ok := strings.Cut(s, "=")
		if !ok {
			lval = stdstrings.ToLower(strings.TrimSpace(lval))

			switch lval {
			case "optional":
				vc.Optional = true
			case "required":
				vc.Required = true
			case "soa":
				vc.SOA = true
			case "uniq", "unique":
				vc.Unique = true
			}
		} else {
			lval = stdstrings.ToLower(strings.TrimSpace(lval))
			rval = strings.TrimSpace(rval)

			switch lval {
			case "each":
				var each VerifyComment
				rval, _ = StripIfFound(rval, LBracks, RBracks)
				if ParseVerifyComment(rval, &each) {
					vc.Each = &each
				}
			case "insertafter":
				vc.InsertAfter = append(vc.InsertAfter, rval)
			case "insertbefore":
				vc.InsertBefore = append(vc.InsertBefore, rval)
			case "min":
				vc.Min = rval
			case "max":
				vc.Max = rval
			case "minlen", "minlength":
				vc.MinLength = rval
			case "maxlen", "maxlength":
				vc.MaxLength = rval
			case "func":
				vc.Funcs = append(vc.Funcs, rval)
			case "prefix":
				vc.Prefix = rval
			case "soa":
				vc.SOA = true
				vc.SOAPrefix = rval
			}
		}

		comment = rest
	}

	return true
}

func VerifyWithFuncs(r *Result, ctx GenerationContext, funcs []string) {
	for _, fn := range funcs {
		var ok bool
		if fn, ok = StripIfFound(fn, LBraces, RBraces); !ok {
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

func (g GeneratorVerify) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "l10n")
	r.Printf("func Verify%s(l l10n.Language, %s %s) error {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorVerify) Body(r *Result, ctx GenerationContext, t *Type) {
	var vc VerifyComment
	if !IsSlice(t.Literal) {
		vc = MergeVerifyComments(ctx.Comments)
		g.SOA = vc.SOA
		g.SOAPrefix = vc.SOAPrefix
	}

	Insert(r, ctx, vc.InsertBefore)
	GenerateType(g, r, ctx, t)
	Insert(r, ctx, vc.InsertAfter)
	r.Line("return nil")
}

func (g GeneratorVerify) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("if err := %sVerify%s(l, %s); err != nil {", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
	{
		r.Line("return err")
	}
	r.Line("}")
}

func (g GeneratorVerify) NewConstant(r *Result, specName string, fieldName string, prefix string, value string, format string) Constant {
	var c Constant

	if expr, ok := StripIfFound(value, LBraces, RBraces); ok {
		vn := VariableName(specName)
		name := PrependVariableNameAndPrefix(expr, vn, g.SOAPrefix)

		if (g.SOA) && (name != expr) && (name == Singular(name)) {
			name = fmt.Sprintf("%s[%s]", Plural(name), g.LoopVariable)
		}
		c = Constant{Name: name}
	} else {
		if strings.StartsWith(fieldName, "Min") {
			fieldName = fieldName[len("Min"):]
		} else if strings.StartsWith(fieldName, "Max") {
			fieldName = fieldName[len("Max"):]
		}
		fieldName = Singular(fieldName)
		c = r.AddConstant(fmt.Sprintf(format, strings.Or(prefix, specName), fieldName), value)
	}

	return c
}

func (g GeneratorVerify) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	var loopVariable string

	fieldDescription := FieldName2Description(ctx.FieldName)
	if g.SOA {
		prefix := FieldName2Description(g.SOAPrefix)
		desc, _ := StripIfFound(Singular(fieldDescription), prefix+" ", "")
		fieldDescription = fmt.Sprintf("%s %%d: %s", strings.Or(prefix, Singular(FieldName2Description(ctx.SpecName))), desc)
		loopVariable = fmt.Sprintf(", %s+1", g.LoopVariable)
	}

	vc := MergeVerifyComments(ctx.Comments)

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
	vc := MergeVerifyComments(ctx.Comments)
	if vc.Each != nil {
		ctx.Comments = append(ctx.Comments, *vc.Each)
	}

	i := ctx.LoopVar()
	r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, ctx.VarName, i)
	{
		GenerateArrayElement(g, r, ctx.WithVar("%s[%s]", ctx.VarName, i), &a.Element)
		if vc.Unique {
			fieldDescription := FieldName2Description(ctx.FieldName)
			j := ctx.LoopVar()
			r.Printf("for %s := %s + 1; %s < len(%s); %s++ {", j, i, j, ctx.VarName, j)
			{
				r.Printf("if %s[%s] == %s[%s] {", ctx.VarName, i, ctx.VarName, j)
				{
					r.Printf(`return fmt.Errorf("%s %%d and %%d are identical", %s, %s)`, fieldDescription, i, j)
				}
				r.Line("}")
			}
			r.Line("}")
		}
	}
	r.Line("}")
}

func (g GeneratorVerify) Slice(r *Result, ctx GenerationContext, s *Slice) {
	if g.SOA {
		GenerateArrayElement(g, r, ctx, &s.Element)
	} else {
		fieldDescription := FieldName2Description(ctx.FieldName)
		vc := MergeVerifyComments(ctx.Comments)

		minLengthConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.MinLength, "Min%s%ssLen")
		maxLengthConst := g.NewConstant(r, ctx.SpecName, ctx.FieldName, vc.Prefix, vc.MaxLength, "Max%s%ssLen")
		if len(vc.MinLength) > 0 {
			r.AddImport("fmt")
			r.Printf("if len(%s) < %s {", ctx.VarName, minLengthConst.Name)
			{
				r.Printf(`return fmt.Errorf("number of %s must be at least %%d", %s)`, fieldDescription, minLengthConst.Name)
			}
			r.Line("}")
		} else if len(vc.MaxLength) > 0 {
			r.AddImport("fmt")
			r.Printf("if len(%s) > %s {", ctx.VarName, minLengthConst.Name)
			{
				r.Printf(`return fmt.Errorf("number of %s must be less than %%d", %s)`, fieldDescription, maxLengthConst.Name)
			}
			r.Line("}")
		} else if (len(vc.MinLength) > 0) && (len(vc.MaxLength) > 0) {
			r.AddImport("fmt")
			r.Printf("if (len(%s) < %s) || (len(%s) > %s) {", ctx.VarName, minLengthConst.Name, ctx.VarName, maxLengthConst.Name)
			{
				r.Printf(`return fmt.Errorf("%s length must be less between %%d and %%d elements long", %s, %s)`, fieldDescription, minLengthConst.Name, maxLengthConst.Name)
			}
			r.Line("}")
		}

		vc.MinLength = ""
		vc.MaxLength = ""
		ctx.Comments = []Comment{vc}

		a := Array{Element: s.Element}
		g.Array(r, ctx, &a)
	}
}

func (g GeneratorVerify) Union(r *Result, ctx GenerationContext, u *Union) {
	vc := MergeVerifyComments(ctx.Comments)

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
					if vc.Each != nil {
						Insert(r, ctx, vc.Each.InsertBefore)
					}
					g.NamedType(r, ctx, &t)
					if vc.Each != nil {
						Insert(r, ctx, vc.Each.InsertAfter)
					}
				}
			}
			r.Tabs--
		}
	}
	r.Tabs++
	r.Line("}")
}
