package main

func GeneratorsEncodingAll() []Generator {
	return append(GeneratorsEncodingJSONAll(), GeneratorsEncodingWireAll()...)
}
