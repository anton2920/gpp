package main

import (
	"unicode"
)

type GeneratorEncodingWireSerialize struct{}

func (g GeneratorEncodingWireSerialize) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "encoding/wire")
	r.Printf("func Serialize%sWire(s *wire.Serializer, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorEncodingWireSerialize) Body(r *Result, ctx GenerationContext, t *Type) {
	GenerateType(g, r, ctx, t)
}

func (g GeneratorEncodingWireSerialize) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sSerialize%sWire(s, %s)", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
}

func (g GeneratorEncodingWireSerialize) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	litName := lit.String()
	if len(ctx.CastName) == 0 {
		r.Printf("s.%c%s(%s)", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.Deref(ctx.VarName))
	} else {
		r.Printf("s.%c%s(%s(%s))", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.CastName, ctx.Deref(ctx.VarName))
	}
}

func (g GeneratorEncodingWireSerialize) Struct(r *Result, ctx GenerationContext, s *Struct) {
	GenerateStructFields(g, r, ctx, s.Fields, nil)
}

func (g GeneratorEncodingWireSerialize) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	GenerateStructField(g, r, ctx.WithCast(LiteralName(lit)), field, lit)
}

func (g GeneratorEncodingWireSerialize) StructFieldSkip(field *StructField) bool {
	return false
}

func (g GeneratorEncodingWireSerialize) Array(r *Result, ctx GenerationContext, a *Array) {
	i := ctx.LoopVar()

	if a.Size == 0 {
		r.Printf("s.%c%s(%s(len(%s)))", unicode.ToUpper(rune(EncodingWireSliceLengthType[0])), EncodingWireSliceLengthType[1:], EncodingWireSliceLengthType, ctx.VarName)
	}
	r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, ctx.VarName, i)
	{
		GenerateArrayElement(g, r, ctx.WithVar("%s[%s]", ctx.VarName, i), &a.Element)
	}
	r.Line("}")
}

func (g GeneratorEncodingWireSerialize) Slice(r *Result, ctx GenerationContext, s *Slice) {
	a := Array{Element: s.Element}
	g.Array(r, ctx, &a)
}

func (g GeneratorEncodingWireSerialize) Union(r *Result, ctx GenerationContext, u *Union) {
	r.Printf("switch %s := %s.(type) {", ctx.VarName, ctx.Deref(ctx.VarName))
	{
		for i, name := range u.Types {
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
				r.Printf("s.%c%s(%d)", unicode.ToUpper(rune(EncodingWireUnionKindType[0])), EncodingWireUnionKindType[1:], i)
				g.NamedType(r, ctx, &t)
			}
			r.Tabs--
		}
	}
	r.Tabs++
	r.Line("}")
}
