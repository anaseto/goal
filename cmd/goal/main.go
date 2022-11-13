package main

import (
	"bytes"
	"flag"
	"fmt"
	"goal"
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
	ctx.SetSource("-", os.Stdin)
	for {
		fmt.Print("  ")
		v, err := ctx.RunExpr()
		if err != nil {
			_, eof := err.(goal.ErrEOF)
			if eof {
				echo(v)
				return
			}
			fmt.Println(err)
		}
		echo(v)
	}
}

func runCommand(ctx *goal.Context, cmd string) {
	_, err := ctx.RunString(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "goal: %v", err)
		os.Exit(1)
	}
}

func echo(v goal.V) {
	if v != nil {
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
