package main

type Generator interface {
	Generate(*Result, *TypeSpec)
}

func GeneratorsAll() []Generator {
	return append([]Generator{GeneratorFill{}, GeneratorVerify{}}, GeneratorsEncodingAll()...)
}
