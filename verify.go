package main

import (
	"fmt"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type GeneratorVerify struct{}

func (g GeneratorVerify) NOP(comments []Comment) bool {
	for _, comment := range comments {
		if _, ok := comment.(VerifyComment); ok {
			return true
			break
		}
	}
	return false
}

func (g GeneratorVerify) Imports() []string {
	return []string{GOFA + "l10n"}
}

func (g GeneratorVerify) Func(specName string, varName string) string {
	return fmt.Sprintf("Verify%s(l l10n.Language, %s *%s) error", specName, varName, specName)
}

func (g GeneratorVerify) Body(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment) {
	GenerateType(g, r, p, t, specName, "", LiteralName(t.Literal), varName, comments, true)
	r.Line("return nil")
}

func (g GeneratorVerify) NamedType(r *Result, p *Parser, t *Type, specName string, varName string, comments []Comment, pointer bool) {
	r.Printf("if err := %sVerify%s(l, &%s); err != nil {", t.PackagePrefix(), t.Name, varName)
	r.Tabs++
	{
		r.Line("return err")
	}
	r.Tabs--
	r.Line("}")
}

func (g GeneratorVerify) Primitive(r *Result, p *Parser, lit TypeLit, specName string, fieldName string, castName string, varName string, comments []Comment, pointer bool) {
	vn := VariableName(fieldName, false)

	var vc VerifyComment
	for _, comment := range comments {
		if c, ok := comment.(VerifyComment); ok {
			strings.Replace(&vc.Min, c.Min)
			strings.Replace(&vc.Max, c.Max)
			strings.Replace(&vc.MinLength, c.MinLength)
			strings.Replace(&vc.MaxLength, c.MaxLength)
			vc.Funcs = append(vc.Funcs, c.Funcs...)
		}
	}

	switch lit.(type) {
	case *Int, *Float:
		minConst := r.AddConstant("Min"+specName+fieldName, vc.Min)
		maxConst := r.AddConstant("Max"+specName+fieldName, vc.Max)

		if (len(vc.Min) > 0) && (len(vc.Max) > 0) {
			r.Printf("if (%s < %s) || (%s > %s) {", varName, minConst.Name, varName, maxConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d and greater than %%d"), %s, %s)`, vn, minConst.Name, maxConst.Name)
			}
			r.Tabs--
			r.Line("}")
		} else if len(vc.Min) > 0 {
			r.Printf("if %s < %s {", varName, minConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d"), %s)`, vn, minConst.Name)
			}
			r.Tabs--
			r.Line("}")
		} else if len(vc.Max) > 0 {
			r.Printf("if %s > %s {", varName, maxConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.Printf(`return fmt.Errorf(l.L("%s must be greater than %%d"), %s)`, vn, maxConst.Name)
			}
			r.Tabs--
			r.Line("}")
		}
	case *String:
		if (len(vc.MinLength) > 0) && (len(vc.MaxLength) > 0) {
			minLengthConst := r.AddConstant("Min"+specName+fieldName+"Len", vc.MinLength)
			maxLengthConst := r.AddConstant("Max"+specName+fieldName+"Len", vc.MaxLength)

			r.Printf("if !strings.LengthInRange(%s, %s, %s) {", varName, minLengthConst.Name, maxLengthConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.AddImport(GOFA + "strings")
				r.Printf(`return fmt.Errorf(l.L("%s length must be between %%d and %%d characters long"), %s, %s)`, vn, minLengthConst.Name, maxLengthConst.Name)
			}
			r.Tabs--
			r.Line("}")
		}
		for _, fn := range vc.Funcs {
			if !strings.StartsWith(fn, "{") {
				fn = fmt.Sprintf("%s(l, %s)", fn, varName)
			} else {
				fn = stdstrings.Replace(fn[1:len(fn)-1], "?", varName, 1)
			}
			r.Printf("if err := %s; err != nil {", fn)
			r.Tabs++
			{
				r.Line("return err")
			}
			r.Tabs--
			r.Line("}")
		}
	}
}

func (g GeneratorVerify) Struct(r *Result, p *Parser, s *Struct, specName string, varName string, comments []Comment) {
	GenerateStructFields(g, r, p, s.Fields, specName, varName, nil)
}

func (g GeneratorVerify) StructField(r *Result, p *Parser, field *StructField, lit TypeLit, specName string, fieldName string, varName string) {
	for _, comment := range field.Comments {
		if _, ok := comment.(VerifyComment); ok {
			GenerateStructField(g, r, p, field, lit, specName, fieldName, field.Type.Name, varName, field.Comments)
			break
		}
	}
}

func (g GeneratorVerify) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorVerify) Array(r *Result, p *Parser, a *Array, specName string, varName string, comments []Comment) {
}

func (g GeneratorVerify) Slice(r *Result, p *Parser, s *Slice, specName string, varName string, comments []Comment) {
}

func (g GeneratorVerify) Union(r *Result, p *Parser, u *Union, specName string, varName string, comments []Comment) {
}
