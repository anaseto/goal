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
	ctx.RegisterDyad("import", gos.VImport)
	ctx.RegisterDyad("print", gos.VPrint)
	ctx.RegisterDyad("say", gos.VSay)
	ctx.RegisterMonad("shell", gos.VShell)
	ctx.RegisterMonad("run", gos.VRun)
	ctx.RegisterDyad("open", gos.VOpen)
	ctx.RegisterMonad("close", gos.VClose)
	ctx.RegisterDyad("read", gos.VRead)

	ctx.AssignGlobal("os.ENV", gos.Environ())
	ctx.AssignGlobal("os.STDOUT", gos.Stdout)
	ctx.AssignGlobal("os.STDERR", gos.Stderr)
	ctx.AssignGlobal("os.STDIN", gos.Stdin)
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
