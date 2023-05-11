package main

import (
	"bufio"
	"os"
	"strings"

	"codeberg.org/anaseto/goal"
	"codeberg.org/anaseto/goal/cmd"
	gos "codeberg.org/anaseto/goal/os"
)

func main() {
	ctx := goal.NewContext()
	registerVariadics(ctx)
	cmd.Cmd(ctx, cmd.Config{Help: getHelp, ProgramName: "goal"})
}

func registerVariadics(ctx *goal.Context) {
	ctx.RegisterMonad("chdir", gos.VFChdir)
	ctx.RegisterMonad("close", gos.VFClose)
	ctx.RegisterMonad("flush", gos.VFFlush)
	ctx.RegisterMonad("shell", gos.VFShell)
	ctx.RegisterDyad("env", gos.VFEnv)
	ctx.RegisterDyad("import", gos.VFImport)
	ctx.RegisterDyad("open", gos.VFOpen)
	ctx.RegisterDyad("print", gos.VFPrint)
	ctx.RegisterDyad("read", gos.VFRead)
	ctx.RegisterDyad("run", gos.VFRun)
	ctx.RegisterDyad("say", gos.VFSay)

	ctx.AssignGlobal("STDOUT", gos.NewStdHandle(os.Stdout))
	ctx.AssignGlobal("STDERR", gos.NewStdHandle(os.Stderr))
	ctx.AssignGlobal("STDIN", gos.NewStdHandle(os.Stdin))
}

func getHelp() map[string]string {
	help := map[string]string{}
	help[""] = helpTopics
	help["s"] = helpSyntax
	help["t"] = helpTypes
	help["v"] = helpVerbs
	help["nv"] = helpNamedVerbs
	help["a"] = helpAdverbs
	help["io"] = helpIO
	help["tm"] = helpTime
	help["time"] = helpTime // for the builtin name
	help["rt"] = helpRuntime
	const vcols = 4
	const scols = 12
	const acols = 5
	const nvcols = 10
	help[":"] = getBuiltin(helpSyntax, "assign", scols) + getBuiltin(helpVerbs, ":", vcols)
	help["::"] = getBuiltin(helpSyntax, "assign", scols) + getBuiltin(helpVerbs, "::", vcols)
	help["»"] = getBuiltin(helpVerbs, "»", vcols)
	help["rshift"] = getBuiltin(helpVerbs, "»", vcols)
	help["«"] = getBuiltin(helpVerbs, "«", vcols)
	help["shift"] = getBuiltin(helpVerbs, "«", vcols)
	for _, v := range []string{"+", "-", "*", "%", "!", "&", "|", "<", ">", "=", "~", ",", "^", "#", "_", "$", "?", "@", "."} {
		help[v] = getBuiltin(helpVerbs, v, vcols)
	}
	for _, v := range []string{"'", "/", "\\"} {
		help[v] = getBuiltin(helpAdverbs, v, acols)
	}
	help["rx"] = getBuiltin(helpSyntax, "regexp", scols) + getBuiltin(helpNamedVerbs, "rx", nvcols)
	for _, v := range []string{"abs", "bytes", "uc", "error", "eval", "firsts", "json", "ocount", "panic", "sign", "csv", "in", "mod", "nan", "rotate", "sub"} {
		help[v] = getBuiltin(helpNamedVerbs, v, nvcols)
	}
	help["¿"] = getBuiltin(helpNamedVerbs, "firsts", nvcols) + getBuiltin(helpNamedVerbs, "in", nvcols)
	for _, v := range []string{"chdir", "close", "env", "flush", "import", "open", "print", "read", "run", "say", "shell", "ARGS", "STDIN", "STDOUT", "STDERR"} {
		help[v] = getBuiltin(helpIO, v, nvcols)
	}
	for _, v := range []string{"rt.vars", "rt.prec", "rt.seed", "rt.time"} {
		help[v] = getBuiltin(helpRuntime, v, nvcols)
	}
	help["qq"] = getBuiltin(helpSyntax, "strings", scols)
	help["rq"] = getBuiltin(helpSyntax, "raw strings", scols)
	return help
}

func getBuiltin(s string, v string, n int) string {
	var sb strings.Builder
	r := strings.NewReader(s)
	sc := bufio.NewScanner(r)
	match := false
	blanks := strings.Repeat(" ", n)
	for sc.Scan() {
		ln := sc.Text()
		if len(ln) < n {
			match = false
			continue
		}
		if strings.Contains(ln[:n], v) || ln[:n] == blanks && match {
			// NOTE: currently no builtin name is a substring of
			// another. Otherwise, this could match more names than
			// wanted.
			match = true
			sb.WriteString(ln)
			sb.WriteByte('\n')
			continue
		}
		match = false
	}
	return sb.String()
}
