package goal

import (
	"fmt"
	"strings"
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
	//{`(1;2;(3;4);4+1;)`, "(1;2;(3;4);5;)"},
	{`f:1+`, "+[1;]"},
	{`f:1+;f 5`, "6"},
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
	{`-3.5 2`, "-3.5 2"},
	{`- 3 2`, "-3 -2"},
	{`- 3.5 2`, "-3.5 -2"},
	{`- 0 1`, "0 -1"},
	{`-(0 1;2 -3;5.5)`, "(0 -1;-2 3;-5.5)"},
	{`#-3 2`, "2"},
	{`#3 -2`, "2"},
	{`3-2`, "1"},
	{`#(!5)`, "5"},
	{`+/(!10)`, "45"},
	{`+/[!10]`, "45"},
	{`+/+/(!3;!3)`, "6"},
	{`1+/!10`, "46"},
	{`1 2+/!10`, "46 47"},
	{`~1`, "0"},
	{`~0`, "1"},
	{`~0 1 2`, "1 0 0"},
	{`~(0 1;1 0)`, "(1 0;0 1)"},
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
	{`_(2.3;-2.1 1.7;3)`, `(2;-3 1; 3)`},
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
	{`"s"$"i"$"abceéèd"`, `"abceéèd"`},
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
	{`3 2^(!5)`, `0 1 4`},
	{`(,2)^(!5)`, `0 1 3 4`},
	{`2 3 in 0 2 4`, `1 0`},
	{`0 in 2 0`, `1`},
	{`0 in 2 3`, `0`},
	{`1 2 in !0`, `0 0`},
	{`3 2 1?2`, `1`},
	{`(!0)?2`, `0`},
	{`3 2 1?0`, `3`},
	{`3 2?2 3`, `1 0`},
	{`3 2?!0`, `!0`},
	{`"a" "b" "c"?"c"`, `2`},
	{`("a";"b";1 2)?,1 2`, `,2`},
	{`("a";+;1 2)?(+)`, `1`},
	{`!2 3`, `(0 0 0 1 1 1;0 1 2 0 1 2)`},
	{`0w+3`, `0w`},
	{`-0w + 3`, `-0w`},
	{`2 3 5 7$3`, `2`},
	{`2 3 5 7$2`, `1`},
	{`2 3 5 7$3.2`, `2`},
	{`2 3 5 7$0`, `0`},
	{`2 3 5 7$6 3`, `3 2`},
	{`2 3 5 7$7`, `4`},
	{`2 3 5 7$7.5`, `4`},
	{`2 3 5 7$8 2 7 5 5.5 3 0`, `4 1 4 3 3 2 0`},
	{`.[+;2 3;{"msg"}]`, `5`},
	{`.[+;2 "a";{"msg"}]`, `"msg"`},
	{`"prefix-"_"prefix-name"`, `"name"`},
	{`"prefix-"_"prefix-name" "prefix-name2"`, `"name" "name2"`},
	{`" []"^"  [text]  "`, `"text"`},
	{`(,!0)^(1;!0)`, `,1`},
	{`=1 0 2 1 2`, `(,1;0 3;2 4)`},
	{`{1=2!x}=!10`, `(0 2 4 6 8;1 3 5 7 9)`},
	{`=1 -1 2 1 2`, `(!0;0 3;2 4)`},
	{`{- 1=2!x}=!10`, `,0 2 4 6 8`},
	{`=0 0 0 0`, `,0 1 2 3`},
	{`{(2*1=2!x)-1}=!10`, `(!0; 1 3 5 7 9)`},
	{`=1 0 2 1 2 4 -1 4 2`, `(,1;0 3;2 4 8;!0;5 7)`},
	{`sign 0 1`, `0 1`},
	{`sign (0 1;-2)`, `(0 1;-1)`},
	{`sign -3 -1 0 1 5`, `-1 -1 0 1 1`},
	{`sign -3.5 -1 0 1 5.2`, `-1 -1 0 1 1`},
	{`ocount 3 2 5 3 2 2 7`, `0 0 0 1 1 2 0`},
	{`ocount "a" "b" "c" "b" "b" "a"`, `0 0 0 1 2 1`},
	{`icount 0 0 1 -1 0 1 2 3 2`, `3 2 2 1`},
	{`"a = a + 1"?"="`, `2`},
	{`"a = a + 1"?"=" "+"`, `2 6`},
	{`a:!3;a[(0 1;1 2)]`, `(0 1;1 2)`},
	{`a:!3;a[(0 1;1 2)][1]`, `1 2`},
	{`a:(#;^);a[(0 1;1 0);3 2 1]`, `((3;1 2 3);(1 2 3;3))`},
	{`"012345"[2]`, `"2345"`},
	{`"012345"[(2 3;5)]`, `("2345" "345";"5")`},
	{`"012345"[2;2]`, `"23"`},
	{`"012345"[(2 3;5);2]`, `("23" "34";"5")`},
	{`"012345"[2 3;2 1]`, `"23" "3"`},
	{`bytes "012"`, `3`},
	{`bytes "é" "è"`, `2 2`},
	{`@[1 2 3;1;1+]`, `1 3 3`},
	{`@[1 2 3;0 1;10+]`, `11 12 3`},
	{`@[1 2 3;(0 1;0 1;0);{x+1}]`, `4 4 3`},
	{`@[1 "a" "b";(1 2;2 1);{x+"c"}]`, `1 "acc" "bcc"`},
	{`@[8 4 5;1;+;10]`, `8 14 5`},
	{`@[8 4 5;1 2;+;10 5]`, `8 14 10`},
	{`@[8 4 5;1 2;+;10]`, `8 14 15`},
	{`@[8 4 5;(1 2;0);+;(10 5;-2)]`, `6 14 10`},
	{`a:(+/);1`, `+/;1`},
	{`@[+/]`, `"r"`},
	{`@[1+]`, `"p"`},
	{`+/'(1 2;3 4)`, `3 7`},
	{`f:{x+y};f[1;]2`, `3`},
	{`b:"a" "b";a:1+!2;{x,'y}[a]'b`, `(((1;"a");(2;"a"));((1;"b");(2;"b")))`},
	{`+,1 2`, `(,1;,2)`},
	{`+1 2 "a"`, `,1 2 "a"`},
	{`+,1 2 "a"`, `(,1;,2;,"a")`},
	{`++,1 2 "a"`, `,1 2 "a"`},
	{`+(1 2;3 4)`, `(1 3;2 4)`},
	{`+(1 "a";3 4)`, `(1 3;"a" 4)`},
	{`and[1;2]`, `2`},
	{`and[1;0;3]`, `0`},
	{`or[0;2]`, `2`},
	{`or[0;0;1]`, `1`},
	{`or[0;0;0]`, `0`},
	{"1\n\n2", `2`},
	{"@(:)", `"v"`},
	{"1,:2", `2`},
	{"1,:[2]", `2`},
	{"1,:[2] 3", `2`},
	{"24 60 60/1 2 3", `3723`},
	{"60/1 2 3", `3723`},
	{"2/1 1 0", `6`},
	{"2.0/1 1 0", `6`},
	{"60/(1 2 3;2 3)", `3723 123`},
	{"2/1 0 0 1.0", `9`},
	{`24 60 60\3723`, `1 2 3`},
	{`60\3723`, `1 2 3`},
	{`2\9`, `1 0 0 1`},
	{`2\6`, `1 1 0`},
	{`2\6 9`, `(0 1;1 0;1 0;0 1)`},
	{`2\(6 9;6)`, `((0 1;1 0;1 0;0 1);1 1 0)`},
	{`#()`, `0`},
	// immutability tests
	{`a:3;1+a;a`, `3`},
	{`a:3 5;1-a;a`, `3 5`},
	{`a:3 5;1 2-a;a`, `3 5`},
	{`a:3 5;1 2-a;a`, `3 5`},
	{`a:(1 2;3 4);2 3=a;a`, `(1 2;3 4)`},
	{`{a:3;1+a;a}0`, `3`},
	{`{a:3 5;1-a;a}0`, `3 5`},
	{`{a:3 5;1 2-a;a}0`, `3 5`},
	{`{a:3 5;1 2-a;a}0`, `3 5`},
	{`{a:(1 2;3 4);2 3=a;a}0`, `(1 2;3 4)`},
	{`a:"d" "a";"c" "d"=a;a`, `"d" "a"`},
	{`a:0 1 0;~a;a`, `0 1 0`},
	{`a:0 1 0;-a;a`, `0 1 0`},
	{`a:0 1 2;-a;a`, `0 1 2`},
	{`a:0 1 2.5;-a;a`, `0 1 2.5`},
	{`a:0 1 -2;sign a;a`, `0 1 -2`},
	{`a:0 1 2.5;_a;a`, `0 1 2.5`},
}

func TestEval(t *testing.T) {
	for i, mt := range matchTests {
		mt := mt
		name := fmt.Sprintf("String%d", i)
		matchString := fmt.Sprintf("(%s) ~ (%s)", mt.Left, mt.Right)
		t.Run(name, func(t *testing.T) {
			ctxLeft := NewContext()
			err := ctxLeft.Compile("", mt.Left)
			ps := ctxLeft.programString()
			if err != nil {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("compile error: %v", err)
				t.Fail()
				return
			}
			vLeft, errLeft := ctxLeft.Run()
			ctxRight := NewContext()
			vRight, errRight := ctxRight.Eval(mt.Right)
			if errLeft != nil || errRight != nil {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("return error: `%v` vs `%v`", errLeft, errRight)
				t.Fail()
				return
			}
			if !Match(vLeft, vRight) {
				t.Log(ps)
				t.Log(matchString)
				if vLeft != (V{}) {
					t.Logf("results: %s vs %s\n", vLeft.Sprint(ctxLeft), vRight.Sprint(ctxRight))
				} else {
					t.Logf("results: %v vs %s\n", vLeft, vRight.Sprint(ctxRight))
				}
				//t.Logf("results (go): %#v vs %#v", vLeft, vRight)
				t.Fail()
			}
		})
	}
}

var matchErrors = [...]matchTest{
	{"(1)2", "type n cannot be applied"}, // exec
	{"1[2]", "type n cannot be applied"}, // compiling
	{"{x}[2;3]", "too many arguments"},
	{"{x+y}[2][2;3]", "too many arguments"},
	{"(!5)[7]", "out of bounds"},
	{"2.3 5[7]", "out of bounds"},
	{`"a" "b"[7]`, "out of bounds"},
	{`0 1 0[7]`, "out of bounds"},
	{"(!5)[`a`]", "non-array"},
	{"{", "EOF"},
	{")", "unexpected ) without opening"},
	{"{)", "unexpected ) without closing"},
	{"]", "unexpected ] without opening"},
	{"{]", "unexpected ] without closing"},
	{"{[]}", "empty argument list"},
	{"{[1]}", "expected identifier or ] in argument list"},
	{"{[a 1]}", "expected ; or ] in argument list"},
	{"1.a", "number: invalid syntax"},
	{`"\%"`, "string: invalid syntax"},
	{"{[a;a]a}", "name a appears twice"},
	{"?[1;2;3;4]", "even number of statements"},
	{"and[1;;3;4]", "empty argument (2-th)"},
	{"or[1;;3;4]", "empty argument (2-th)"},
	{"1 2+1 2 3", "length mismatch"},
	{"{}", "empty lambda"},
	{"[]", "empty sequence"},
	{"(;)", "empty slot in list"},
	{`a:3 "a";1 2=a`, `bad type`},
	{`a:3 "a";"c" 2=a`, `bad type`},
}

func TestErrors(t *testing.T) {
	for i, mt := range matchErrors {
		mt := mt
		name := fmt.Sprintf("String%d", i)
		matchString := fmt.Sprintf("%s", mt.Left)
		t.Run(name, func(t *testing.T) {
			ctx := NewContext()
			err := ctx.Compile("", mt.Left)
			ps := ctx.programString()
			if err == nil {
				var v V
				v, err = ctx.Run()
				if err == nil {
					t.Log(ps)
					t.Log(matchString)
					t.Errorf("no error left: result: %v\nexpected: %v", v, mt.Right)
				}
			}
			e, ok := err.(*Error)
			if !ok {
				// should never happen
				t.Log(ps)
				t.Log(matchString)
				t.Errorf("bad error: `%v`\nexpected:`%v`", err, mt.Right)
			}
			if !strings.Contains(e.Msg, mt.Right) {
				t.Log(ps)
				t.Log(matchString)
				t.Logf("\n   error: %v\nexpected: %v", e.Msg, mt.Right)
				t.Fail()
				return
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

func BenchmarkFib(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("fib:{?[x~0;0;x~1;1;(fib x-1)+(fib x-2)]}; fib 35")
	}
}

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

func BenchmarkWhileN(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := NewContext()
		ctx.Eval("100 {x+1}/!10000")
	}
}
