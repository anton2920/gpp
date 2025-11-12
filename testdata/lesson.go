package main

//gpp:generate: encoding(wire)
type (
	//gpp:union: *StepTest, *StepProgramming
	Step interface{}

	//gpp:nop
	StepCommon struct {
		Name string
	}

	Question struct {
		Name           string
		Answers        []string
		CorrectAnswers []int32
	}
	Check struct {
		Input  string
		Output string
	}

	StepTest struct {
		StepCommon

		Questions []Question
	}
	StepProgramming struct {
		StepCommon

		Description string
		Checks      [2][]Check
	}
)
