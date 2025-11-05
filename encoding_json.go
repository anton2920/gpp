package main

type GeneratorJSON struct{}

func (g GeneratorJSON) Generate(r *Result, p *Parser, ts *TypeSpec) {
	var e Encoding
	e.Parser = p

	e.Serialize(r, ts, "JSON")
	e.Deserialize(r, ts, "JSON")
}
