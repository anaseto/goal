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
	{`f:-+[];f[5]`, ",,-5"},
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
	{`","\"a,b,c"`, `"a" "b" "c"`},
	{`","\("a,b,c";"d,e")`, `("a" "b" "c";"d" "e")`},
	{`f:+/;f[!10]`, `45`},
	{`(+/)[!10]`, `45`},
	{`(+/) @ !10`, `45`},
	{`{x>0}{x-1}/2`, `0`},
	{`{x>0}{x-1}/-2`, `-2`},
	{`{x>0}{x-1}\2`, `2 1 0`},
	{`3{x-1}/4`, `1`},
	{`3.0{x-1}/4`, `1`},
	{`3{x-1}\4`, `4 3 2 1`},
	{`3.0{x-1}\4`, `4 3 2 1`},
	{`5 6+\1 2 3`, `(6 7;8 9;11 12)`},
	{`5 6+/1 2 3`, `11 12`},
	{`5 6+\0#0`, `!0`},
	{"dec:{({0};{dec x-1})[x>0]x};dec 3", "0"},
	{`{x+y}.2 3`, `5`},
	{`$2 3`, `"2 3"`},
	{`2_3 4 5 6`, `5 6`},
	{`?2 2 3 4 3 3`, `2 3 4`},
	{`1 2 3@2`, `3`},
	{`{0 1 1 0}#4 1 5 3`, `1 5`},
	{`_"ABC"`, `"abc"`},
	{`_2.3`, `2`},
	{`^3 5 0`, `0 3 5`},
	{`#,1 2`, `1`},
	{`1,2`, `1 2`},
	{`;-3 -2`, `-3 -2`},
	{`+/-3 -2`, `-5`},
	{`1+/-3 -2`, `-4`},
	{`(2)-3`, `-1`},
	{`[(2)-3]`, `-1`},
	{`{[a;b]a+b}[2;3]`, `5`},
	{`?[1;2;3]`, `2`},
	{`?[0;2;3]`, `3`},
	{"fib:{(({(fib x-1)+(fib x-2)};{1})[x=1];{0})[x=0]x}; fib 2", `1`},
	{"fib:{(({(fib x-1)+(fib x-2)};{1})[x=1];{0})[x=0]x}; fib 3", `2`},
	{"fib:{(({(fib x-1)+(fib x-2)};{1})[x=1];{0})[x=0]x}; fib 4", `3`},
	{`fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}; fib 4`, `3`},
	{`fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}; fib 1`, `1`},
	{`fibrec:{?[x~0;y;x~1;z;fibrec[x-1;z;y+z]]};fib: fibrec[4;0;1]`, `3`},
	{`fibrec:{?[x~0;y;x~1;z;fibrec[x-1;z;y+z]]};fib: fibrec[0;0;1]`, `0`},
	{`fib:{*x{x[1],+/x}/0 1}; fib 4`, `3`},
	{`a:!10;a[-1]`, `9`},
	{`a:!10;a[-1 -3 0]`, `9 7 0`},
	{`a:(!5;!5);a[0;2 3]`, `2 3`},
	{`a:(!5;!5);a[0 1;2 3]`, `(2 3;2 3)`},
	{`a:(!5;!5);a[;2 3]`, `(2 3;2 3)`},
	{`a:(!5;!5);a[0;]`, `!5`},
	{`a:(1 "a";2 "b");a[0;]`, `1 "a"`},
	{`a:(1 "a";2 "b");a[;1]`, `"a" "b"`},
	{`[1;:2;3]`, `2`},
	{`i:2;1 2 3 4 i`, `3`},
	{`/23`, ``},
	{` /23`, ``},
	{"/\n23\n\\\n", ``},
	{`a:3;f:{a:a+x};f 2;a`, `3`},
	{`a:3;f:{a::a+x};f 2;a`, `5`},
	{`{2}|1 2 3 4`, `3 4 1 2`},
	{`{x<5}_!10`, `5 6 7 8 9`},
	{`{0 1 1 0}_4 1 5 3`, `4 3`},
	{`{x>0}_2 -3 1`, `,-3`},
	{`"i"$2.3`, `2`},
	{`"i"$2 0`, `2 0`},
	{`"i"$"ab"`, `97 98`},
	{`"i"$"a" "b"`, `(,97;,98)`},
	{`"n"$"1.5"`, `1.5`},
	{`"n"$"1.5" "2"`, `1.5 2`},
	{`"s"$97`, `"a"`},
	{`"s"$97 98`, `"ab"`},
	{`."2+3"`, `5`},
	{`a:."2+3";a`, `5`},
	{`.'"2+3" "2"`, `5 2`},
	{`."a:3";a`, `3`},
	{`."f:{x}";f 3`, `3`},
	{`2^1 2 3`, `(1 2;2 3)`},
	{`2 5_!10`, `(2 3 4;5 6 7 8 9)`},
	{`(!0)_!10`, `!0`},
	{`(,10)_!10`, `,!0`},
	{`0 10_!10`, `(!10;!0)`},
	{`(!5)^3 2`, `0 1 4`},
	{`2 3 in 0 2 4`, `1 0`},
	{`0w+3`, `0w`},
	{`-0w + 3`, `-0w`},
}

func TestEval(t *testing.T) {
	for i, mt := range matchTests {
		mt := mt
		name := fmt.Sprintf("String%d", i)
		matchString := fmt.Sprintf("(%v) ~ (%v)", mt.Left, mt.Right)
		t.Run(name, func(t *testing.T) {
			ctxLeft := NewContext()
			vLeft, errLeft := ctxLeft.Eval(mt.Left)
			ctxRight := NewContext()
			vRight, errRight := ctxRight.Eval(mt.Right)
			if !Match(vLeft, vRight) {
				t.Log(ctxLeft.programString())
				t.Log(matchString)
				t.Logf("results: %s vs %s", vLeft.Sprint(ctxLeft), vRight.Sprint(ctxRight))
				t.Fail()
			}
			if errLeft != nil || errRight != nil {
				if !t.Failed() {
					t.Log(ctxLeft.programString())
					t.Log(ctxRight.programString())
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
		ctx.Eval("-/!1000")
	}
}

func BenchmarkFoldPlus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("+/!1000")
	}
}

func BenchmarkFoldLambdaPlus(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("{x+y}/!1000")
	}
}

func BenchmarkFoldWhile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("{x<1000}{x+1}/1")
	}
}

func BenchmarkFoldDo(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("1000{x+1}/1")
	}
}

//func BenchmarkFib(b *testing.B) {
//for n := 0; n < b.N; n++ {
//ctx := NewContext()
//ctx.Eval("fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}; fib 35")
//}
//}

func BenchmarkFibTailRec(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("fibrec:{?[x~0;y;x~1;z;fibrec[x-1;z;y+z]]};fibrec[35;0;1]")
	}
}

func BenchmarkFibDoWhile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("*35{x[1],+/x}/0 1")
	}
}

func BenchmarkNewContext(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewContext()
	}
}
