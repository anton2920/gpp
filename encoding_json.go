package main

func GeneratorsEncodingJSONAll() []Generator {
	return []Generator{GeneratorEncodingJSONSerialize{}, GeneratorEncodingJSONDeserialize{}}
}

func JSONStructFieldSkip(field *StructField) bool {
	return field.Tag == `json:"-"`
}
