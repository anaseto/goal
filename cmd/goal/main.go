package main

import (
	"goal"
	"goal/cmd"
	gos "goal/os"
	"os"
)

func main() {
	ctx := goal.NewContext()
	registerVariadics(ctx)
	cmd.Cmd(ctx, cmd.Config{Help: getHelp()})
}

func registerVariadics(ctx *goal.Context) {
	ctx.RegisterDyad("import", gos.VImport)
	ctx.RegisterDyad("print", gos.VPrint)
	ctx.RegisterDyad("say", gos.VSay)
	ctx.RegisterMonad("shell", gos.VShell)
	ctx.RegisterMonad("slurp", gos.VSlurp)
	ctx.AssignGlobal("os.ENV", goal.NewAS(os.Environ()))
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
	return help
}
