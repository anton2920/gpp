package main

import "unicode"

type GeneratorEncodingWireDeserialize struct{}

func (g GeneratorEncodingWireDeserialize) Decl(r *Result, ctx GenerationContext, t *Type) {
	if IsSlice(t.Literal) {
		t = &Type{Literal: Pointer{BaseType: *t}}
	}
	r.AddImport(GOFA + "encoding/wire")
	r.Printf("func Deserialize%sWire(d *wire.Deserializer, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorEncodingWireDeserialize) Body(r *Result, ctx GenerationContext, t *Type) {
	if IsSlice(t.Literal) {
		ctx.VarName = ctx.Deref(ctx.VarName)
	}
	GenerateType(g, r, ctx, t)
}

func (g GeneratorEncodingWireDeserialize) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sDeserialize%sWire(d, %s)", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
}

func (g GeneratorEncodingWireDeserialize) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	litName := lit.String()
	if len(ctx.CastName) == 0 {
		r.Printf("%s = d.%c%s()", ctx.Deref(ctx.VarName), unicode.ToUpper(rune(litName[0])), litName[1:])
	} else {
		r.Printf("%s = %s(d.%c%s())", ctx.Deref(ctx.VarName), ctx.CastName, unicode.ToUpper(rune(litName[0])), litName[1:])
	}
}

func (g GeneratorEncodingWireDeserialize) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorEncodingWireDeserialize) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	r.AddImport(field.Type.Package)
	GenerateStructField(g, r, ctx.WithCast(field.Type.String()), field, lit)
}

func (g GeneratorEncodingWireDeserialize) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorEncodingWireDeserialize) Array(r *Result, ctx GenerationContext, a *Array) {
	i := ctx.LoopVar()

	if a.Size == 0 {
		r.Printf("%s = make([]%s, d.%c%s())", ctx.VarName, a.Element.String(), unicode.ToUpper(rune(EncodingWireSliceLengthType[0])), EncodingWireSliceLengthType[1:])
	}
	r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, ctx.VarName, i)
	{
		GenerateArrayElement(g, r, ctx.WithVar("%s[%s]", ctx.VarName, i), &a.Element)
	}
	r.Line("}")
}

func (g GeneratorEncodingWireDeserialize) Slice(r *Result, ctx GenerationContext, s *Slice) {
	a := Array{Element: s.Element}
	g.Array(r, ctx, &a)
}

func (g GeneratorEncodingWireDeserialize) Union(r *Result, ctx GenerationContext, u *Union) {
	const value = "value"

	r.Printf("switch d.%c%s() {", unicode.ToUpper(rune(EncodingWireUnionKindType[0])), EncodingWireUnionKindType[1:])
	{
		for i, name := range u.Types {
			if name[0] != '*' {
				ctx.Autoderef = false
			} else {
				ctx.Autoderef = true
				name = name[1:]
			}
			t := Type{Name: name}

			r.Printf("case %d:", i)
			{
				r.Printf("var %s %s", value, t)
				g.NamedType(r, ctx.WithVar(value), &t)
				r.Printf("*%s = %s", ctx.VarName, ctx.AddrOf(value))
			}
			r.Tabs--
		}
	}
	r.Tabs++
	r.Line("}")
}
