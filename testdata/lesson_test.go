package main

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/anton2920/gofa/encoding/wire"
)

var testStep = StepTest{
	StepCommon: StepCommon{Name: "Back-end development basics"},
	Questions: []Question{
		Question{
			Name: "What is an API?",
			Answers: []string{
				"One",
				"Two",
				"Three",
				"Four",
			},
			CorrectAnswers: []int32{2},
		},
		Question{
			Name: "To be or not to be?",
			Answers: []string{
				"To be",
				"Not to be",
			},
			CorrectAnswers: []int32{0, 1},
		},
		Question{
			Name: "Third question",
			Answers: []string{
				"What?",
				"Where?",
				"When?",
				"Correct",
			},
			CorrectAnswers: []int32{3},
		},
	},
}

func BenchmarkSerializeStepWire(b *testing.B) {
	var s wire.Serializer
	s.Buffer = make([]byte, 1024)
	step := Step(&testStep)

	for i := 0; i < b.N; i++ {
		s.Reset()
		SerializeStepWire(&s, &step)
	}

	//fmt.Printf("Wire size = %d\n", len(s.Buffer))
}

func BenchmarkSerializeStepGob(b *testing.B) {
	var buf bytes.Buffer

	gob.Register(&StepTest{})
	enc := gob.NewEncoder(&buf)
	step := Step(&testStep)

	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(step)
	}

	//fmt.Printf("Gob size = %d\n", buf.Len())
}
