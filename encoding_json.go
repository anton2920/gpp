package main

func GeneratorsEncodingJSONAll() []Generator {
	return []Generator{GeneratorEncodingJSONSerialize{}, GeneratorEncodingJSONDeserialize{}}
}
