package main

import (
	"fmt"
	stdstrings "strings"

	"github.com/anton2920/gofa/strings"
)

type GeneratorVerify struct{}

func (g GeneratorVerify) Generate(r *Result, p *Parser, ts *TypeSpec) {
	r.AddImport(GOFA + "l10n")
	name := VariableName(ts.Name, false)

	r.Printf("\nfunc Verify%s(l l10n.Language, %s *%s) error {", ts.Name, name, ts.Name)
	r.Tabs++

	g.GenerateType(r, p, ts.Name, name, &ts.Type, VerifyComment{})

	r.Line("return nil")
	r.Tabs--
	r.Line("}")
}

func (g GeneratorVerify) GenerateType(r *Result, p *Parser, key string, name string, t *Type, vc VerifyComment) {
	if t.Literal != nil {
		g.GenerateTypeLit(r, p, key, name, t.Literal, vc)
	} else {
		tabs := r.Tabs

		r.String("if err := ")

		if len(t.Package) > 0 {
			r.AddImport(t.Package)
			r.Tabs = 0
			r.String(t.Package)
			r.Rune('.')
		}

		r.Printf("Verify%s(l, &%s); err != nil {", t.Name, name)
		r.Tabs = tabs + 1
		{
			r.Line("return err")
		}
		r.Tabs--
		r.Line("}")
	}
}

func (g GeneratorVerify) GenerateTypeLit(r *Result, p *Parser, key string, name string, lit TypeLit, vc VerifyComment) {
	vn := VariableName(key, false)

	switch lit := lit.(type) {
	case *Int, *Float:
		minConst := r.AddConstant("Min"+key, vc.Min)
		maxConst := r.AddConstant("Max"+key, vc.Max)

		if (len(vc.Min) > 0) && (len(vc.Max) > 0) {
			r.Printf("if (%s < %s) || (%s > %s) {", name, minConst.Name, name, maxConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d and greater than %%d"), %s, %s)`, vn, minConst.Name, maxConst.Name)
			}
			r.Tabs--
			r.Line("}")
		} else if len(vc.Min) > 0 {
			r.Printf("if %s < %s {", name, minConst.Name)
			r.Tabs++
			{
				r.AddImport("fmt")
				r.Printf(`return fmt.Errorf(l.L("%s must not be less than %%d"), %s)`, vn, minConst.Name)
			}
			r.Tabs--
			r.Line("}")
		} else if len(vc.Max) > 0 {
			r.Printf("if %s > %s {", name, maxConst.Name)
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
			minLengthConst := r.AddConstant("Min"+key+"Len", vc.MinLength)
			maxLengthConst := r.AddConstant("Max"+key+"Len", vc.MaxLength)

			r.Printf("if !strings.LengthInRange(%s, %s, %s) {", name, minLengthConst.Name, maxLengthConst.Name)
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
				fn = fmt.Sprintf("%s(l, %s)", fn, name)
			} else {
				fn = stdstrings.Replace(fn[1:len(fn)-1], "?", name, 1)
			}
			r.Printf("if err := %s; err != nil {", fn)
			r.Tabs++
			{
				r.Line("return err")
			}
			r.Tabs--
			r.Line("}")
		}
	case *Struct:
		g.GenerateStruct(r, p, key, name, lit)
	case *Slice:
	}
}

func (g GeneratorVerify) GenerateStruct(r *Result, p *Parser, key string, name string, s *Struct) {
	g.GenerateStructFields(r, p, key, name, s.Fields)
}

func (g GeneratorVerify) GenerateStructFields(r *Result, p *Parser, key string, name string, fields []StructField) {
	for _, field := range fields {
		for _, comment := range field.Comments {
			if vc, ok := comment.(VerifyComment); ok {
				fieldName := strings.Or(field.Name, field.Type.Name)
				g.GenerateType(r, p, key+fieldName, fmt.Sprintf("%s.%s", name, fieldName), &field.Type, vc)
				// fmt.Printf("For field %q: %#v\n", key, vc)
			}
		}
	}
}
