package main

import (
	"bufio"
	"flag"
	"fmt"
	"goal"
	"log"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"strings"
)

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	optE := flag.String("e", "", "execute command")
	optD := flag.Bool("d", false, "debug info")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-e command] [path]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "See man page goal(1) for details (TODO).")
	}
	flag.Parse()
	if *cpuprofile != "" {
		// profiling
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goal: %v", err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	args := flag.Args()
	ctx := goal.NewContext()
	registerVariadics(ctx)
	if *optD {
		defer runDebug(ctx)
	}
	if *optE != "" {
		runCommand(ctx, *optE)
	}
	if *optE == "" && len(args) == 0 || len(args) == 1 && args[0] == "-" {
		runStdin(ctx)
		return
	}
	if len(args) == 0 {
		return
	}
	fname := args[0]
	ctx.AssignGlobal("ARGS", goal.NewAS(args[1:]))
	bs, err := os.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
	err = ctx.Compile(fname, string(bs))
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		if *optD {
			printProgram(ctx)
		}
		os.Exit(1)
	}
	if *optD {
		printProgram(ctx)
		os.Exit(0)
	}
	r, err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
	if r.IsError() {
		fmt.Fprint(os.Stderr, r.Error())
		os.Exit(1)
	}
}

func runStdin(ctx *goal.Context) {
	help := ctx.RegisterMonad("help", goal.VariadicFun{
		Func: func(ctx *goal.Context, args []goal.V) goal.V {
			if len(args) >= 1 {
				arg := args[0]
				switch {
				case goal.Match(arg, goal.NewS("+")):
					fmt.Println(strings.TrimSpace(helpVERBS))
				case goal.Match(arg, goal.NewS("nv")):
					fmt.Println(strings.TrimSpace(helpNAMEDVERBS))
				case goal.Match(arg, goal.NewS("'")):
					fmt.Println(strings.TrimSpace(helpADVERBS))
				case goal.Match(arg, goal.NewS("io")):
					fmt.Println(strings.TrimSpace(helpIO))
				case goal.Match(arg, goal.NewS("syn")):
					fmt.Println(strings.TrimSpace(helpSyntax))
				case goal.Match(arg, goal.NewS("types")):
					fmt.Println(strings.TrimSpace(helpTypes))
				default:
					fmt.Println(strings.TrimSpace(helpTopics))
				}
			}
			return goal.V{}
		}})
	ctx.AssignGlobal("help", help)
	ctx.AssignGlobal("h", help)
	lr := lineReader{r: bufio.NewReader(os.Stdin)}
	fmt.Printf("goal repl, type help\"\" for basic info.\n")
	for {
		fmt.Print("  ")
		s, err := lr.readLine()
		if err != nil && s == "" {
			return
		}
		r, err := ctx.Eval(s)
		if err != nil {
			fmt.Println("'ERROR " + strings.TrimSuffix(err.Error(), "\n"))
			continue
		}
		assigned := ctx.AssignedLast()
		if !assigned {
			echo(ctx, r)
		}
	}
}

type lineReader struct {
	r *bufio.Reader
}

func (lr lineReader) readLine() (string, error) {
	sb := strings.Builder{}
	for {
		r, _, err := lr.r.ReadRune()
		if err != nil {
			return sb.String(), err
		}
		if r == '\n' {
			return sb.String(), nil
		}
		if r != '\r' {
			sb.WriteRune(r)
		}
	}
}

func printProgram(ctx *goal.Context) {
	fmt.Fprintf(os.Stderr, "goal: debug info below:\n%v", ctx.Show())
}

func runDebug(ctx *goal.Context) {
	if r := recover(); r != nil {
		printProgram(ctx)
		log.Printf("Caught panic: %v\nStack Trace:\n", r)
		debug.PrintStack()
	}
}

func runCommand(ctx *goal.Context, cmd string) {
	r, err := ctx.Eval(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
	if r.IsError() {
		fmt.Fprint(os.Stderr, r.Error())
		os.Exit(1)
	}
}

func echo(ctx *goal.Context, x goal.V) {
	if x != (goal.V{}) {
		fmt.Printf("%s\n", x.Sprint(ctx))
	}
}

func usageError(usage bool, msgs ...interface{}) {
	s := "goal: "
	s += fmt.Sprint(msgs...)
	fmt.Fprintln(os.Stderr, s)
	if usage {
		flag.Usage()
	}
	os.Exit(1)
}

func registerVariadics(ctx *goal.Context) {
	say := ctx.RegisterMonad("say", goal.VariadicFun{Func: vSay})
	ctx.AssignGlobal("say", say)
	slurp := ctx.RegisterMonad("slurp", goal.VariadicFun{Func: vSlurp})
	ctx.AssignGlobal("slurp", slurp)
	shell := ctx.RegisterMonad("shell", goal.VariadicFun{Func: vShell})
	ctx.AssignGlobal("shell", shell)
}

func vSay(ctx *goal.Context, args []goal.V) goal.V {
	for _, arg := range args {
		switch argv := arg.Value().(type) {
		case goal.S:
			fmt.Println(string(argv))
		default:
			fmt.Println(arg.Sprint(ctx))
		}
	}
	return goal.NewI(1)
}

func vSlurp(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		switch x := args[0].Value().(type) {
		case goal.S:
			bytes, err := os.ReadFile(string(x))
			if err != nil {
				return goal.NewError(goal.NewS(fmt.Sprintf("slurp: %v", err)))
			}
			return goal.NewS(string(bytes))
		default:
			return goal.NewPanic("slurp: non-string filename")
		}
	default:
		return goal.NewPanic("slurp: too many arguments")
	}
}

const helpTopics = `
Type help TOPIC or h TOPIC where TOPIC is one of:

"+"	verbs (like +*-%,)
"nv"	named verbs (like in, sign)
"'"	adverbs ('/\)
"io"	io functions (slurp, say)
"syn"   syntax
"types" types

Notations:
	s (string) f (function) F (2-args function)
	n (number) i (integer) x,y (any other)
`
const helpVERBS = `
VERBS
:x  return	:3 -> return 3 prematurely
x:y right	2:3 -> 3	name:3 -> assign 3 to name
+x  flip	+(1 2;3 4) -> (1 3;2 4)
x+y add		2+3 -> 5
s+s concat	"a"+"b" -> "ab"
-x  negate	- 2 3 -> -2 -3
x-y subtract	5-3 -> 2
s-s trim suffix "file.txt"-".txt" -> "file"
*x  first	*3 2 4 -> 3
x*y multiply	2*3 -> 6
s*x repeat	"a"*3 2 1 0 -> "aaa" "aa" "a" ""
%x  classify	%1 2 3 1 2 3 -> 0 1 2 0 1 2	%"a" "b" "a" -> 0 1 0
x%y divide	3%2 -> 1.5
!i  enum	!5 -> 0 1 2 3 4
!x  odometer	!2 3 -> (0 0 0 1 1 1;0 1 2 0 1 2)
x!y mod		3!5 4 3 -> 2 1 0	
&x  where	&0 0 1 0 0 0 1 -> 2 6
x&y min		2&3 -> 2	4&3 -> 3
|x  reverse	|!5 -> 4 3 2 1 0
x|y max		2|3 -> 3	4|3 -> 4
f|y rotate	{2}|1 2 3 4 -> 3 4 1 2
<x  ascend	<2 4 3 -> 0 2 1
x<y less	2<3 -> 1
>x  descend	>2 4 3 -> 1 2 0
x>y greater	2>3 -> 0
=x  group	=1 0 2 1 2 -> (,1;0 3;2 4)	=-1 2 -1 2 -> (!0;!0;1 3)
f=x group by	{1=2!x}=!10 -> (0 2 4 6 8;1 3 5 7 9)
x=y equal	2 3 4=3 -> 0 1 0
~x  not		~0 1 2 -> 1 0 0
x~y match	3~3 -> 1	2 3~3 2 -> 0
,x  enlist	,1 -> ,1 (list with one element)
x,y join	1,2 -> 1 2
^x  sort	^3 5 0 -> 0 3 5
i^y windows	2^!4 -> (1 2;2 3;3 4)
s^y trim	" []"^"  [text]  " -> "text"
x^y without	2 3^1 2 3 4 -> 1 4
#x  length	#2 4 5 -> 3
i#y take	2#4 1 5 -> 4 1	    4#3 1 5 -> 3 1 5 3 (cyclic)
f#y replicate	{0 1 1 0}#4 1 5 3 -> 1 5    {x>0}#2 -3 1 -> 2 1
_N  floor	_2.3 -> 2     _1.5 3.7 -> 1 3
_S  to lower	_"ABC" -> "abc"     _"AB" "CD" -> "ab" "cd"
i_x drop	2_3 4 5 6 -> 5 6
s_x trim prefix "pref-"_"pref-name" -> "name"
x_y cut		2 5_!10 -> (2 3 4;5 6 7 8 9)
f_x weed out	{0 1 1 0}_4 1 5 3 -> 4 3    {x>0}_2 -3 1 -> ,-3
$x  string	$2 3 -> "2 3"
i$x split	2$!6 -> (0 1;2 3;4 5)	2$"a" "b" "c" -> ("a" "b";,"c")
s$y cast	"i"$2.3 -> 2    "i"$"ab" -> 97 98   "s"$97 98 -> "ab"
s$y parse num	"n"$"1.5" -> 1.5
x$y binsearch	2 3 5 7$8 2 7 5 5.5 3 0 -> 4 1 4 3 3 2 0
?x  uniq	?2 2 3 4 3 3 -> 2 3 4
s?y index	"a = a + 1"?"=" "+" -> 2 6
x?y find	3 2 1?2 -> 1	3 2 1?0	-> 3
@x  type	@2 -> "i"    @"ab" -> "s"    @2 3 -> "I"
f@y apply	(|)@1 2 -> 2 1 (like |[1 2] -> 2 1 or |1 2)
s@y substr	"012345"[2] -> "2345"	"012345"[2;3] -> "234"
a@y at		1 2 3@2 -> 3	1 2 3[2] -> 3
.s  eval	."2+3" -> 5
x.y applyN	{x+y}.2 3 -> 5    {x+y}[2;3] -> 5    (1 2;3 4)[0;1] -> 2

@[x;y;f]    amend	@[1 2 3;0 1;10+] -> 11 12 3
@[x;y;F;z]  amend	@[8 4 5;(1 2;0);+;(10 5;-2)] -> 6 14 10
.[f;x;f]    try		.[+;2 3;{"msg"}] -> 5	.[+;2 "a";{"msg"}] -> "msg"
`

const helpNAMEDVERBS = `
NAMED VERBS
x in y    member of	2 3 in 0 2 4 -> 1 0
bytes x	  byte-count	bytes "abc" -> 3
error x	  error		r:{?[~x=0;1%x;error "zero"]}0;?["e"~@r;*r;r] -> "zero"
icount x  index-count	icount 0 0 1 -1 0 1 2 3 2 -> 3 2 2 1 (same as #'=x)
ocount x  occur-count	ocount 3 2 5 3 2 2 7 -> 0 0 0 1 1 2 0
sign x    sign		sign -3 -1 0 1.5 5 -> -1 -1 0 1 1
`

const helpADVERBS = `
ADVERBS
f'x	each	#'(4 5;6 7 8) -> 2 3	
x F'y   each	2 3#'1 2 -> (1 1;2 2 2)
F/x	fold	+/!10 -> 45
F\x	scan	+\!10 -> 0 1 3 6 10 15 21 28 36 45
x F/y	fold	1 2+/!10 -> 46 47
x F\y   scan	5 6+\1 2 3 -> (6 7;8 9;11 12)
n f/x	do	3{x*2}/4 -> 32
n f\x	dos	3{x*2}\4 -> 4 8 16 32
f f/x   while	{x<100}{x*2}/4 -> 128
f f\x   whiles	{x<100}{x*2}\4 -> 4 8 16 32 64 128
s/x	join	","/"a" "b" "c" -> "a,b,c"
s\x	split	","\"a,b,c" -> "a" "b" "c"
I/x	encode	24 60 60/1 2 3 -> 3723	2/1 1 0 -> 6
I\x	decode	24 60 60\3723 -> 1 2 3	2\6 -> 1 1 0
`
const helpIO = `
IO
slurp[s]	read file named s	lines:"\n"\slurp["/path/to/file"]
say[x;...]	print value(s) with newline	say 2+3
shell[cmd]	run a command through the shell
`

const helpSyntax = `
literals	1	1.5	"text"
arrays		1 2 -3 4	1 "a" -2 "b"	(1 2;"a";(3;"b"))
variables	a:2 (assign)	a+3 (use)	a::2 (assign global)
expressions	2*3+4 -> 14	1+|1 2 3 -> 4 3 2	+/1 2 3 -> 6
index		1 2 3[1] -> 2 (same as x@1) (1 2;3 4)[0;1] -> 2 (same as x.(0;1))
lambdas		{x+y+z}[2;3;0] -> 5	{[a;b;c]a+b+c}[1;2;3] -> 6
projections	{x+y}[2;] 3 -> 5	(2+) 3 -> 5
cond		?[1;2;3] -> 2	?[0;2;3] -> 3	?[0;2;"";3;4] -> 4
and/or		and[1;2] -> 2   and[1;0;3] -> 0   or[0;2] -> 2   or[0;0;0] -> 0
sequence	[a:2;b:a+3;a+10] -> 12 (bracket block [] at start of expression)
return		[1;:2;3] -> 2 (a : at start of expression)
`

const helpTypes = `
atom	array	name		examples
-----------------------------------------------------------------------
n	N	number		0	1.5	!5	1.2 3 1.8
s	S	string		"abc"	"a" "b" "c"
f		function	+	{x*2}	(1-)	%[;2]
e		error		error "msg"
	A	generic array	("a" 1;"b" 2;"c" 3)	(+;-;*;"any")
-----------------------------------------------------------------------
`
