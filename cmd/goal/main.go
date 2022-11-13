package main

import (
	"bytes"
	"flag"
	"fmt"
	"goal"
	"strings"
	//"log"
	"os"
)

func main() {
	optE := flag.String("e", "", "command")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-e command] [path]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "See man page goal(1) for details (TODO).")
	}
	flag.Parse()
	args := flag.Args()
	ctx := goal.NewContext()
	registerVariadics(ctx)
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
	v := make(goal.AV, len(args)-1)
	for _, s := range args[1:] {
		v = append(v, goal.S(s))
	}
	ctx.AssignGlobal("args", v)
	bs, err := os.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
	ctx.SetSource(fname, bytes.NewReader(bs))
	_, err = ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
}

func runStdin(ctx *goal.Context) {
	help := ctx.RegisterVariadic("help", goal.VariadicFun{
		Func: func(ctx *goal.Context, args []goal.V) goal.V {
			fmt.Println(strings.TrimSpace(helpTopics))
			return nil
		}})
	ctx.AssignGlobal("help", help)
	ctx.SetSource("-", os.Stdin)
	fmt.Printf("goal repl, type help\"\" for basic info.\n")
	for {
		fmt.Print("  ")
		v, err := ctx.RunExpr()
		if err != nil {
			_, eof := err.(goal.ErrEOF)
			if eof {
				echo(ctx, v)
				return
			}
			fmt.Println(err)
			continue
		}
		echo(ctx, v)
	}
}

func runCommand(ctx *goal.Context, cmd string) {
	_, err := ctx.RunString(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
}

func echo(ctx *goal.Context, v goal.V) {
	if v != nil && !ctx.LastIsAssign() {
		fmt.Printf("%v\n", v)
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
	say := ctx.RegisterVariadic("say", goal.VariadicFun{
		Func: func(ctx *goal.Context, args []goal.V) goal.V {
			for _, v := range args {
				switch v := v.(type) {
				case goal.S:
					fmt.Println(string(v))
					return nil
				default:
					fmt.Printf("%v\n", v)
					return nil
				}
			}
			return nil
		}})
	ctx.AssignGlobal("say", say)
	slurp := ctx.RegisterVariadic("slurp", goal.VariadicFun{
		Func: func(ctx *goal.Context, args []goal.V) goal.V {
			switch len(args) {
			case 1:
				switch v := args[0].(type) {
				case goal.S:
					bytes, err := os.ReadFile(string(v))
					if err != nil {
						return goal.E("slurp:" + err.Error())
					}
					return goal.S(bytes)
				default:
					return goal.E("slurp: non-string filename")
				}
			default:
				return goal.E("slurp: too many arguments")
			}
		}})
	ctx.AssignGlobal("slurp", slurp)
}

const helpTopics = `
VERBS
:x  return	:3 -> return 3 (TODO)
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
%x  classify	%1 2 3 1 2 3 -> 0 1 2 0 1 2
x%y divide	3%2 -> 1.5
!x  enum	!5 -> 0 1 2 3 4
x!y mod		3!5 4 3 -> 2 1 0	
&x  where	&0 0 1 0 0 0 1 -> 2 6
x&y min		2&3 -> 2	4&3 -> 3
|x  reverse	|!5 -> 4 3 2 1 0
x|y max		2|3 -> 3	4|3 -> 4
<x  ascend	<2 4 3 -> 0 2 1
x<y less	2<3 -> 1
>x  descend	>2 4 3 -> 1 2 0
x>y greater	2>3 -> 0
=x  group	=1 0 2 1 2 -> (1;0 3;2 4)
x=y equal	2 3 4=3 -> 0 1 0
~x  not		~0 1 2 -> 1 0 0
x~y match	3~3 -> 1	2 3~3 2 -> 0
,   enlist	,1 -> ,1 (list with one element)
x,y join	1,2 -> 1 2
^   sort	^3 5 0 -> 0 3 5
x^y cut		TODO
#x  length	#2 4 5 -> 3
i#y take	2#4 1 5 -> 4 1	    4#3 1 5 -> 3 1 5 3 (cyclic)
f#y replicate	{0 1 1 0}#4 1 5 3 -> 1 5    {x>0}#2 -3 1 -> 2 1
_N  floor	_2.3 -> 2
_s  to lower	_"ABC" -> "abc"
i_x drop	2_3 4 5 6 -> 5 6
$x  string	$2 3 -> "2 3"
x$y cast	TODO
?x  uniq	?2 2 3 4 3 3 -> 2 3 4
x?y find	TODO
@x  type	@2 -> "i"    @"ab" -> "s"    @2 3 -> "I"
x@y apply	1 2 3@2 -> 3	1 2 3[2] -> 3
.   eval	TODO
x.y applyN	{x+y}.2 3 -> 5    {x+y}[2;3] -> 5

ADVERBS
f'x	each	#'(4 5;6 7 8) -> 2 3	
x F'y   each	2 3#'1 2 -> (1 1;2 2 2)
F/x	fold	+/!10 -> 45
F\x	scan	+\!10 -> 0 1 3 6 10 15 21 28 36 45
x F/y	fold	1 2+/!10 -> 46 47
x F\y   scan	5 6+\1 2 3 -> (6 7;8 9;11 12)
n f/x	do	3{x*2}/4 -> 32
n f\x	do	3{x*2}\4 -> 4 8 16 32
f f/x   while	{x<100}{x*2}/4 -> 128
f f/x   while	{x<100}{x*2}\4 -> 4 8 16 32 64 128
s/x	join	","/"a" "b" "c" -> "a,b,c"
s\x	split	","\"a,b,c" -> "a" "b" "c"

IO
slurp[s]	read file named s
say[x;...]	print value(s) with newline
`
