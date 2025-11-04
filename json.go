package main

type GeneratorJSON struct {
}

func (g GeneratorJSON) Generate(r *Result, ts *TypeSpec) {
	println("FROM JSON")
}
