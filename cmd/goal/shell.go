package main

import (
	"goal"
	"os"
	"os/exec"
)

func vShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) == 0 {
		return goal.NewError("shell: missing command string")
	}
	if len(args) > 1 {
		return goal.Errorf("shell[cmd;opts] : too many arguments (%d)", len(args))
	}
	var cmds string
	switch arg := args[len(args)-1].(type) {
	case goal.S:
		cmds = string(arg)
	default:
		return goal.Errorf("shell[cmd;opts] : cmd is not a string (%s)", arg.Type())
	}
	cmd := exec.Command("/bin/sh", "-c", cmds)
	cmd.Stderr = os.Stderr
	bytes, err := cmd.Output()
	if err != nil {
		return goal.Errorf("shell[cmd;opts] : %v", err)
	}
	return goal.S(bytes)
}
