package main

import (
	"codeberg.org/anaseto/goal"
	"codeberg.org/anaseto/goal/cmd"
	gos "codeberg.org/anaseto/goal/os"
	"os"
	"strings"
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
	ctx.RegisterMonad("slurp", gos.VSlurp)
	ctx.RegisterDyad("open", gos.VOpen)
	ctx.RegisterMonad("close", gos.VClose)
	ctx.RegisterDyad("read", gos.VRead)
	ctx.AssignGlobal("os.ENV", getEnviron())
}

func getEnviron() goal.V {
	env := os.Environ()
	ss := make([]string, len(env)*2)
	for i, s := range env {
		b, a, _ := strings.Cut(s, "=")
		ss[i] = b
		ss[i+len(env)] = a
	}
	keys := &goal.AS{Slice: ss[:len(env)]}
	values := &goal.AS{Slice: ss[len(env):]}
	var n int = 2
	keys.InitWithRC(&n)
	values.InitWithRC(&n)
	return goal.NewDict(goal.NewV(keys), goal.NewV(values))
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
