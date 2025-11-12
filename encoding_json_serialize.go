package main

import "unicode"

type GeneratorEncodingJSONSerialize struct{}

func (g GeneratorEncodingJSONSerialize) Decl(r *Result, ctx GenerationContext, t *Type) {
	r.AddImport(GOFA + "encoding/json")
	r.Printf("func Serialize%sJSON(s *json.Serializer, %s %s) {", ctx.SpecName, ctx.VarName, t.String())
}

func (g GeneratorEncodingJSONSerialize) Body(r *Result, ctx GenerationContext, t *Type) {
	GenerateType(g, r, ctx, t)
}

func (g GeneratorEncodingJSONSerialize) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sSerialize%sJSON(s, &%s)", t.PackagePrefix(), t.Name, ctx.VarName)
}

func (g GeneratorEncodingJSONSerialize) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	litName := lit.String()
	if len(ctx.CastName) == 0 {
		r.Printf("s.%c%s(%s)", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.VarName)
	} else {
		r.Printf("s.%c%s(%s(%s))", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.CastName, ctx.VarName)
	}
}

func (g GeneratorEncodingJSONSerialize) Struct(r *Result, ctx GenerationContext, s *Struct) {
	r.Line("s.ObjectBegin()")
	r.Line("{")
	{
		GenerateStructFields(g, r, ctx, s.Fields, nil)
	}
	r.Line("}")
	r.Line("s.ObjectEnd()")
}

func (g GeneratorEncodingJSONSerialize) StructField(r *Result, ctx GenerationContext, field *StructField, lit TypeLit) {
	r.Printf("s.Key(`%s`)", ctx.FieldName)
	GenerateStructField(g, r, ctx, field, lit)
}

func (g GeneratorEncodingJSONSerialize) StructFieldSkip(field *StructField) bool {
	return JSONStructFieldSkip(field)
}

func (g GeneratorEncodingJSONSerialize) Array(r *Result, ctx GenerationContext, a *Array) {
	i := ctx.LoopVar()

	r.Line("s.ArrayBegin()")
	r.Printf("for %s := 0; %s < len(%s); %s++ {", i, i, ctx.VarName, i)
	{
		GenerateArrayElement(g, r, ctx.WithVar("%s[%s]", ctx.VarName, i), &a.Element)
	}
	r.Line("}")
	r.Line("s.ArrayEnd()")
}

func (g GeneratorEncodingJSONSerialize) Slice(r *Result, ctx GenerationContext, s *Slice) {
	a := Array{Element: s.Element}
	g.Array(r, ctx, &a)
}

func (g GeneratorEncodingJSONSerialize) Union(r *Result, ctx GenerationContext, u *Union) {
}
