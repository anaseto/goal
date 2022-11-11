package goal

import (
	"fmt"
	"testing"
)

type matchTest struct {
	Left  string
	Right string
}

var matchTests = [...]matchTest{
	{"2+3", "5"},
	{"1 2+3", "4 5"},
	{"a:2 5;b:a+5;|b", "10 7"},
	{`a:1;b:{x+1} a`, "2"},
	{`a:1;b:{x+y+2}[a;4]`, "7"},
	{`a:1 3 5;b:3;a+b`, "4 6 8"},
	{`a:1 3 5;f:{x+3};f[a]`, "4 6 8"},
	{`a:1 3 5;;;|a`, "5 3 1"},
	{`a:1 3 5;a[2 0 1 0]`, "5 1 3 1"},
	{`(1;2;(3;4);4+1)`, "(1;2;(3;4);5)"},
	{`(1;2;(3;4);4+1;)`, "(1;2;(3;4);5;)"},
	{`f:1+`, "+[1;]"},
	{`f:1+;f 5`, "6"},
	{`f:-+;f[5;2]`, "-7"},
	{`#(1;2;3)`, "3"},
	{`#((2;3);(1;2;5))`, "2"},
	{`#'((2;3);(1;2;5))`, "2 3"},
	{`2 3#'1 2`, "(1 1;2 2 2)"},
	{`{0 1 0 1} 0`, "0 1 0 1"},
	{`{0 1 0 1}#1 2 3 4`, "2 4"},
	{`+/!10`, "45"},
	{`#0#1`, "0"},
	{`+/0#1`, "0"},
	{`+\!10`, "0 1 3 6 10 15 21 28 36 45"},
	{`","/"a" "b" "c" "d"`, `"a,b,c,d"`},
	{`-3 2`, "-3 2"},
	{`- 3 2`, "-3 -2"},
	{`#-3 2`, "2"},
	{`#3 -2`, "2"},
	{`3-2`, "1"},
	{`#(!5)`, "5"},
	{`+/(!10)`, "45"},
	{`+/[!10]`, "45"},
	{`+/+/(!3;!3)`, "6"},
	{`1+/!10`, "46"},
	{`~1`, "0"},
	{`~0`, "1"},
	{`~0 1 2`, "1 0 0"},
	{`"name"+".suffix"`, `"name.suffix"`},
	{`"name.suffix"-".suffix"`, `"name"`},
	{`2!!5`, `0 1 0 1 0`},
	{`1%2`, `0.5`},
	{`1 2%2`, `0.5 1`},
	{`1 2*3`, `3 6`},
	{`2*3 5`, `6 10`},
	{`{x+y}/!5`, `+/!5`},
	{`{x-y}\!5`, `-\!5`},
	{`{|x}[!5]`, `|!5`},
	{`{x;y}[1;2]`, `2`},
	{`{0;y;x}[1;2]`, `1`},
	{`|/!5`, `4`},
	{`&/!5`, `0`},
	{`=/!5`, `0`},
	{`2 0 2=2`, `1 0 1`},
	{`=/2 2 1`, `1`},
	{`a:3;f:{a:3;a+x};(a;f 2)`, `(3;5)`},
}

func TestRunString(t *testing.T) {
	for i, mt := range matchTests {
		mt := mt
		name := fmt.Sprintf("String%d", i)
		matchString := fmt.Sprintf("(%v) ~ (%v)", mt.Left, mt.Right)
		t.Run(name, func(t *testing.T) {
			ctxLeft := NewContext()
			vLeft, errLeft := ctxLeft.RunString(mt.Left)
			ctxRight := NewContext()
			vRight, errRight := ctxRight.RunString(mt.Right)
			if !match(vLeft, vRight) {
				t.Log(ctxLeft.ast.String())
				t.Log(ctxLeft.ProgramString())
				t.Log(matchString)
				t.Logf("results: %v vs %v", vLeft, vRight)
				t.Fail()
			}
			if errLeft != nil || errRight != nil {
				if !t.Failed() {
					t.Log(ctxLeft.ast.String())
					t.Log(ctxLeft.ProgramString())
					t.Log(matchString)
				}
				t.Logf("return error: `%v` vs `%v`", errLeft, errRight)
				t.Fail()
			}
		})
	}
}

func BenchmarkFoldMinus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.RunString("-/!1000")
	}
}

func BenchmarkFoldPlus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.RunString("+/!1000")
	}
}

func BenchmarkNewContext(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewContext()
	}
}
