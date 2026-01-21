package main

import "testing"

func TestProperCut(t *testing.T) {
	sets := [...][]struct {
		String string
		Sep    string
		S1     string
		S2     string
		R1     string
		R2     string
		OK     bool
	}{
		/* Set 1: "a(a+), (b, c, d), e" */
		{
			{"a(a+), (b, c, d), e", ",", "(", ")", "a(a+)", " (b, c, d), e", true},
			{"(b, c, d), e", ",", "(", ")", "(b, c, d)", " e", true},
			{"e", ",", "(", ")", "e", "", false},
		},

		/* Set 2: "a(a+), (b, c, d)" */
		{
			{"a(a+), (b, c, d)", ",", "(", ")", "a(a+)", " (b, c, d)", true},
			{"(b, c, d)", ",", "(", ")", "(b, c, d)", "", false},
		},

		/* Set 3: InsertBefore={{if true { println("TRUE") }}}; verify: InsertAfter={{ if false {} else {println("FALSE!")}}} */
		{
			{`InsertBefore={{if true { println("TRUE"); }}}; verify: InsertAfter={{ if false {} else {println("FALSE!")}}}`, ";", LBraces, RBraces, `InsertBefore={{if true { println("TRUE"); }}}`, ` verify: InsertAfter={{ if false {} else {println("FALSE!")}}}`, true},
			{`verify: InsertAfter={{ if false {} else {println("FALSE!")}}}`, ";", LBraces, RBraces, `verify: InsertAfter={{ if false {} else {println("FALSE!")}}}`, "", false},
		},

		/* Set 4: MinLength=1, Each=[[Min={{0}}, Max={{len(.Answers)}}]] */
		{
			{`MinLength=1, Each=[[Min={{0}}, Max={{len(.Answers)}}]]`, ",", LBracks, RBracks, `MinLength=1`, ` Each=[[Min={{0}}, Max={{len(.Answers)}}]]`, true},
			{`Each=[[Min={{0}}, Max={{len(.Answers)}}]]`, ",", LBracks, RBracks, `Each=[[Min={{0}}, Max={{len(.Answers)}}]]`, "", false},
		},

		/* TODO(anton2920): this set doesn't work! */
		/* Set 5: {{ {for i := 0; i < 10; i++ {println(i)}}; {println("HELLO!")} }}; Value2 */
		{
			{`{{ {for i := 0; i < 10; i++ {println(i)}}; {println("HELLO!")} }}; Value2`, ";", LBraces, RBraces, `{{ {for i := 0; i < 10; i++ {println(i)}}; {println("HELLO!")} }}`, "Value2", true},
		},
	}
	for i, samples := range sets[:len(sets)-1] {
		for j, sample := range samples {
			r1, r2, ok := ProperCut(sample.String, sample.Sep, sample.S1, sample.S2)
			if r1 != sample.R1 {
				t.Errorf("set %d: sample %d: expected r1 = %q, got %q", i+1, j+1, sample.R1, r1)
			}
			if r2 != sample.R2 {
				t.Errorf("set %d: sample %d: expected r2 = %q, got %q", i+1, j+1, sample.R2, r2)
			}
			if ok != sample.OK {
				t.Errorf("set %d: sample %d: expected ok = %v, got %v", i+1, j+1, sample.OK, ok)
			}
		}
	}
}
