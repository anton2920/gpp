package main

import "unicode"

type Generator interface {
	Generate(*Result, *Parser, *TypeSpec)
}

func GeneratorsAll() []Generator {
	return append(append(GeneratorsFillAll(), GeneratorVerify{}), GeneratorsEncodingAll()...)
}

func Private(c byte) bool {
	return (c == '_') || (unicode.IsLower(rune(c)))
}
