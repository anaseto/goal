package main

import (
	"fmt"
	"goal"
	"os"
	"os/exec"
)

func vShell(ctx *goal.Context, args []goal.V) goal.V {
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
		return goal.NewError(goal.NewS(fmt.Sprintf("shell[cmd] : %v", err)))
	}
	return goal.NewS(string(bytes))
}
