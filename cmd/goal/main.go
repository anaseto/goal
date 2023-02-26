package main

import (
	"codeberg.org/anaseto/goal"
	"codeberg.org/anaseto/goal/cmd"
	gos "codeberg.org/anaseto/goal/os"
)

func main() {
	ctx := goal.NewContext()
	registerVariadics(ctx)
	cmd.Cmd(ctx, cmd.Config{Help: getHelp(), ProgramName: "goal"})
}

func registerVariadics(ctx *goal.Context) {
	ctx.RegisterMonad("close", gos.VClose)
	ctx.RegisterMonad("flush", gos.VFlush)
	ctx.RegisterMonad("run", gos.VRun)
	ctx.RegisterMonad("shell", gos.VShell)
	ctx.RegisterDyad("env", gos.VEnv)
	ctx.RegisterDyad("import", gos.VImport)
	ctx.RegisterDyad("open", gos.VOpen)
	ctx.RegisterDyad("print", gos.VPrint)
	ctx.RegisterDyad("read", gos.VRead)
	ctx.RegisterDyad("say", gos.VSay)

	ctx.AssignGlobal("STDOUT", gos.Stdout)
	ctx.AssignGlobal("STDERR", gos.Stderr)
	ctx.AssignGlobal("STDIN", gos.Stdin)
}

func getHelp() map[string]string {
	help := map[string]string{}
	help[""] = helpTopics
	help["+"] = helpVERBS
	help["nv"] = helpNAMEDVERBS
	help["'"] = helpADVERBS
	help["io"] = helpIO
	help["time"] = helpTime
	help["syn"] = helpSyntax
	help["types"] = helpTypes
	help["goal"] = helpGoal
	return help
}
