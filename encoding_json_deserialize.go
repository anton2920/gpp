package main

import "unicode"

type GeneratorEncodingJSONDeserialize struct{}

func (g GeneratorEncodingJSONDeserialize) Decl(r *Result, ctx GenerationContext, t *Type) {
	var star string
	if IsSlice(t.Literal) {
		star = "*"
	}

	r.AddImport(GOFA + "encoding/json")
	r.Printf("func Deserialize%sJSON(d *json.Deserializer, %s %s%s) bool {", ctx.SpecName, ctx.VarName, star, t.String())
}

func (g GeneratorEncodingJSONDeserialize) Body(r *Result, ctx GenerationContext, t *Type) {
	if IsSlice(t.Literal) {
		ctx.VarName = ctx.Deref(ctx.VarName)
	}
	GenerateType(g, r, ctx, t)
	r.Line("return d.Error == nil")
}

func (g GeneratorEncodingJSONDeserialize) NamedType(r *Result, ctx GenerationContext, t *Type) {
	r.Printf("%sDeserialize%sJSON(d, %s)", t.PackagePrefix(), t.Name, ctx.AddrOf(ctx.VarName))
}

func (g GeneratorEncodingJSONDeserialize) Primitive(r *Result, ctx GenerationContext, lit TypeLit) {
	litName := lit.String()
	if len(ctx.CastName) == 0 {
		r.Printf("d.%c%s(%s)", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.AddrOf(ctx.VarName))
	} else {
		r.AddImport("unsafe")
		r.Printf("d.%c%s((*%s)(unsafe.Pointer(%s)))", unicode.ToUpper(rune(litName[0])), litName[1:], ctx.CastName, ctx.AddrOf(ctx.VarName))
	}
}

func (g GeneratorEncodingJSONDeserialize) Struct(r *Result, ctx GenerationContext, s *Struct) {
	r.Line("var key string")
	r.Line("d.ObjectBegin()")
	r.Line("for d.Key(&key) {")
	{
		r.Line("switch key {")
		GenerateStructFields(g, r, ctx, s.Fields, nil)
		r.Tabs++
		r.Line("}")
	}
	r.Line("}")
	r.Line("d.ObjectEnd()")
}

func (g GeneratorEncodingJSONDeserialize) StructField(r *Result, ctx GenerationContext, field *StructField, lit ForeignTypeLit) {
	r.Printf("case \"%s\":", ctx.FieldName)
	{
		GenerateStructField(g, r, ctx.WithCast(LiteralName(lit.TypeLit)), field, lit.TypeLit)
	}
	r.Tabs--
}

func (g GeneratorEncodingJSONDeserialize) StructFieldSkip(field *StructField) bool {
	return JSONStructFieldSkip(field)
}

func (g GeneratorEncodingJSONDeserialize) Array(r *Result, ctx GenerationContext, a *Array) {
	i := ctx.LoopVar()

	r.Line("d.ArrayBegin()")
	r.Printf("for %s := 0; (%s < len(%s)) && (d.Next()); %s++ {", i, i, ctx.VarName, i)
	{
		GenerateArrayElement(g, r, ctx.WithVar("%s[%s]", ctx.VarName, i), &a.Element)
	}
	r.Line("}")
	r.Line("d.ArrayEnd()")
}

func (g GeneratorEncodingJSONDeserialize) Slice(r *Result, ctx GenerationContext, s *Slice) {
	const element = "element"

	r.Line("d.ArrayBegin()")
	r.Line("for d.Next() {")
	{
		r.AddImport(s.Element.Package)
		r.Printf("var %s %s", element, s.Element.String())
		GenerateArrayElement(g, r, ctx.WithVar(element), &s.Element)
		r.Printf("%s = append(%s, %s)", ctx.VarName, ctx.VarName, element)
	}
	r.Line("}")
	r.Line("d.ArrayEnd()")
}

func (g GeneratorEncodingJSONDeserialize) Union(r *Result, ctx GenerationContext, u *Union) {
}
