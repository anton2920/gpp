package main

type Generator interface {
	Generate(*Result, *Parser, *TypeSpec)
}

func GeneratorsAll() []Generator {
	return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
}
