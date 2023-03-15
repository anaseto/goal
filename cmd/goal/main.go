package main

import (
	"os"

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
	ctx.RegisterMonad("shell", gos.VShell)
	ctx.RegisterDyad("env", gos.VEnv)
	ctx.RegisterDyad("import", gos.VImport)
	ctx.RegisterDyad("open", gos.VOpen)
	ctx.RegisterDyad("print", gos.VPrint)
	ctx.RegisterDyad("read", gos.VRead)
	ctx.RegisterDyad("run", gos.VRun)
	ctx.RegisterDyad("say", gos.VSay)

	ctx.AssignGlobal("STDOUT", gos.NewStdHandle(os.Stdout))
	ctx.AssignGlobal("STDERR", gos.NewStdHandle(os.Stderr))
	ctx.AssignGlobal("STDIN", gos.NewStdHandle(os.Stdin))
}

func getHelp() map[string]string {
	help := map[string]string{}
	help[""] = helpTopics
	help["syn"] = helpSyntax
	help["types"] = helpTypes
	help["+"] = helpVerbs
	help["nv"] = helpNamedVerbs
	help["'"] = helpAdverbs
	help["io"] = helpIO
	help["time"] = helpTime
	help["goal"] = helpGoal
	return help
}
