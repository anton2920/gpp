package main

var (
	EncodingWireSliceLengthType = "int32"
	EncodingWireUnionKindType   = "int32"
)

func GeneratorsEncodingWireAll() []Generator {
	return []Generator{GeneratorEncodingWireSerialize{}, GeneratorEncodingWireDeserialize{}}
}
