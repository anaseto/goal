package io

import (
	"fmt"
	"goal"
	"os"
	"os/exec"
)

func VSay(ctx *goal.Context, args []goal.V) goal.V {
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

func VSlurp(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		switch x := args[0].Value().(type) {
		case goal.S:
			bytes, err := os.ReadFile(string(x))
			if err != nil {
				return goal.NewError(goal.NewS(err.Error()))
			}
			return goal.NewS(string(bytes))
		default:
			return goal.NewPanic("slurp: non-string filename")
		}
	default:
		return goal.NewPanic("slurp: too many arguments")
	}
}

func VShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) == 0 {
		return goal.NewPanic("shell: missing command string")
	}
	if len(args) > 1 {
		return goal.Panicf("shell[cmd] : too many arguments (%d)", len(args))
	}
	var cmds string
	switch arg := args[len(args)-1].Value().(type) {
	case goal.S:
		cmds = string(arg)
	default:
		return goal.Panicf("shell[cmd] : cmd is not a string (%s)", arg.Type())
	}
	cmd := exec.Command("/bin/sh", "-c", cmds)
	cmd.Stderr = os.Stderr
	bytes, err := cmd.Output()
	if err != nil {
		return goal.NewError(goal.NewS(err.Error()))
	}
	return goal.NewS(string(bytes))
}
