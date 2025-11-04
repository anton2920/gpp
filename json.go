package main

type GeneratorJSON struct {
	*Parser
}

func (g GeneratorJSON) Generate(r *Result, ts *TypeSpec) {
	var e Encoding
	e.Parser = g.Parser

	e.Serialize(r, ts, "JSON")
	e.Deserialize(r, ts, "JSON")
}
