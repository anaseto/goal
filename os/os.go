// Package os provides variadic function definitions for IO/OS builtins.
package os

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"codeberg.org/anaseto/goal"
)

// VFImport implements the import dyad.
//
// import s : evaluate file s+".goal" (or s if it has already an extension)
// with prefix s (without extension) for globals.
//
// x import s : same as import s but use prefix x.  If x is empty, no prefix is
// used.
//
// It returns 0 and does nothing if a file has already been evaluated.
func VFImport(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("import : too many arguments (%d)", len(args))
	}
	var prefix string
	var hasPfx bool
	if len(args) == 2 {
		pfx := args[1]
		p, ok := pfx.BV().(goal.S)
		if !ok {
			return panicType("x import s", "x", pfx)
		}
		prefix = string(p)
		hasPfx = true
	}
	s := args[0]
	switch sv := s.BV().(type) {
	case goal.S:
		return importWithPrefix(ctx, prefix, string(sv), hasPfx)
	case *goal.AS:
		var r goal.V
		for _, si := range sv.Slice() {
			r = importWithPrefix(ctx, prefix, si, hasPfx)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		if len(args) == 2 {
			return panicType("x import s", "s", s)
		}
		return panicType("import s", "s", s)
	}
}

func importWithPrefix(ctx *goal.Context, prefix, name string, hasPfx bool) goal.V {
	if !hasPfx {
		prefix = filepath.Base(name)
		prefix = strings.TrimSuffix(prefix, filepath.Ext(prefix))
	}
	fname := name
	if filepath.Ext(fname) == "" {
		fname += ".goal"
	}
	source, err := readFile(fname)
	if err != nil {
		path, ok := searchIncFile(ctx, fname)
		if ok {
			fname = path
			source, err = readFile(fname)
		}
		if err != nil {
			return goal.Panicf("import : %v", err)
		}
	}
	r, err := ctx.EvalPackage(source, fname, prefix)
	if err != nil {
		_, ok := err.(goal.ErrPackageImported)
		if ok {
			return goal.NewI(0)
		}
		return goal.Panicf("import : %v", err)
	}
	return r
}

// searchIncFile returns the path to filename relative to the current directory
// or the GOALLIB environment variable, and boolean true if such a file exists.
// Otherwise it returns a false boolean.
func searchIncFile(ctx *goal.Context, fname string) (string, bool) {
	goalLIB, ok := os.LookupEnv("GOALLIB")
	if !ok {
		return "", false
	}
	sep := ":"
	if runtime.GOOS == "windows" {
		// Like they do for PERLLIB.
		sep = ";"
	}
	for _, dir := range strings.Split(goalLIB, sep) {
		fpath := filepath.Join(dir, fname)
		fi, err := os.Stat(fpath)
		if err == nil && fi.Mode().IsRegular() {
			return fpath, true
		}
	}
	return "", false
}

func ppanic(pfx string, x goal.V) goal.V {
	return goal.NewPanic(pfx + x.Panic())
}

// VFPrint implements the print dyad.
//
// print x : outputs x to standard output. It returns a true value on success.
//
// x print y : outputs y to x, where x is an io.Writer handle or a filename
// (goal.S).
func VFPrint(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		x := args[0]
		err := printV(ctx, x)
		if err != nil {
			return goal.Errorf("print x : %v", err)
		}
		return goal.NewI(1)
	case 2:
		r := fprintFunc(ctx, args[1], args[0], fprintV)
		if r.IsPanic() {
			return ppanic("h print x : ", r)
		}
		return r
	default:
		return goal.NewPanic("print : too many arguments")
	}
}

// VFSay implements the say dyad. It is the same as print, but appends a newline
// to the result.
func VFSay(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		x := args[0]
		err := sayV(ctx, x)
		if err != nil {
			return goal.Errorf("say x : %v", err)
		}
		return goal.NewI(1)
	case 2:
		r := fprintFunc(ctx, args[1], args[0], fsayV)
		if r.IsPanic() {
			return ppanic("h print x : ", r)
		}
		return r
	default:
		return goal.NewPanic("say : too many arguments")
	}
}

func fprintFunc(ctx *goal.Context, w, x goal.V, f func(*goal.Context, io.Writer, goal.V) error) goal.V {
	switch wv := w.BV().(type) {
	case goal.S:
		file, err := os.Create(string(wv))
		if err != nil {
			return goal.Errorf("%v", err)
		}
		defer func() {
			file.Close()
		}()
		b := bufio.NewWriter(file)
		err = f(ctx, b, x)
		if err != nil {
			return goal.Errorf("%v", err)
		}
		err = b.Flush()
		if err != nil {
			return goal.Errorf("%v", err)
		}
	case *file:
		err := f(ctx, wv.b.Writer, x)
		if err != nil {
			return goal.Errorf("%v", err)
		}
	case *command:
		wout := wv.b.Writer
		if wout == nil {
			return goal.NewPanic("read-only pipe")
		}
		err := f(ctx, wout, x)
		if err != nil {
			return goal.Errorf("%v", err)
		}
	case io.Writer:
		err := f(ctx, wv, x)
		if err != nil {
			return goal.Errorf("%v", err)
		}
	default:
		return goal.NewPanic("h should be a string or writer")
	}
	return goal.NewI(1)
}

func printV(ctx *goal.Context, x goal.V) error {
	switch xv := x.BV().(type) {
	case goal.S:
		_, err := fmt.Print(string(xv))
		return err
	case *goal.AS:
		buf := bufio.NewWriter(os.Stdout)
		imax := xv.Len() - 1
		for i, s := range xv.Slice() {
			buf.WriteString(s)
			if i < imax {
				buf.WriteString(ctx.OFS)
			}
		}
		return buf.Flush()
	default:
		_, err := fmt.Printf("%s", x.Append(ctx, nil))
		return err
	}
}

func sayV(ctx *goal.Context, x goal.V) error {
	switch xv := x.BV().(type) {
	case goal.S:
		_, err := fmt.Println(string(xv))
		return err
	case *goal.AS:
		buf := bufio.NewWriter(os.Stdout)
		imax := xv.Len() - 1
		for i, s := range xv.Slice() {
			buf.WriteString(s)
			if i < imax {
				buf.WriteString(ctx.OFS)
			}
		}
		buf.WriteByte('\n')
		return buf.Flush()
	default:
		_, err := fmt.Printf("%s\n", x.Append(ctx, nil))
		return err
	}
}

func fprintV(ctx *goal.Context, w io.Writer, x goal.V) error {
	switch xv := x.BV().(type) {
	case goal.S:
		_, err := fmt.Fprint(w, string(xv))
		return err
	case *goal.AS:
		var err error
		imax := xv.Len() - 1
		for i, s := range xv.Slice() {
			_, err = fmt.Fprint(w, s)
			if i < imax {
				fmt.Fprint(w, ctx.OFS)
			}
		}
		return err
	default:
		_, err := w.Write(x.Append(ctx, nil))
		return err
	}
}

func fsayV(ctx *goal.Context, w io.Writer, x goal.V) error {
	switch xv := x.BV().(type) {
	case goal.S:
		_, err := fmt.Fprintln(w, string(xv))
		return err
	case *goal.AS:
		imax := xv.Len() - 1
		for i, s := range xv.Slice() {
			fmt.Fprint(w, s)
			if i < imax {
				fmt.Fprint(w, ctx.OFS)
			}
		}
		_, err := fmt.Fprint(w, "\n")
		return err
	default:
		_, err := w.Write(append(x.Append(ctx, nil), '\n'))
		return err
	}
}

// VFShell implements the shell dyad. It works like the run dyad, but runs a
// single string command through /bin/sh instead.
func VFShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("shell : too many arguments (%d)", len(args))
	}
	var shellcmd string
	y := args[0]
	switch yv := y.BV().(type) {
	case goal.S:
		shellcmd = string(yv)
	default:
		if len(args) == 2 {
			return panicType("x shell s", "s", y)
		}
		return panicType("shell s", "s", y)
	}
	cmd := exec.Command("/bin/sh", "-c", shellcmd)
	var sb strings.Builder
	cmd.Stdout = &sb
	cmd.Stderr = os.Stderr
	switch len(args) {
	case 1:
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return cmdError(err, cmd, sb.String())
		}
		return goal.NewS(sb.String())
	case 2:
		x := args[1]
		s, ok := x.BV().(goal.S)
		if !ok {
			return panicType("x run s", "x", x)
		}
		cmd.Stdin = strings.NewReader(string(s))
	}
	err := cmd.Run()
	if err != nil {
		return cmdError(err, cmd, sb.String())
	}
	return goal.NewS(sb.String())
}

func cmdError(err error, cmd *exec.Cmd, out string) goal.V {
	keys := goal.NewAS([]string{"code", "msg", "out"})
	values := goal.NewAV([]goal.V{goal.NewI(int64(cmd.ProcessState.ExitCode())), goal.NewS(err.Error()), goal.NewS(out)})
	return goal.NewError(goal.NewD(keys, values))
}

// VFRun implements the run monad.
//
// run s : run command s, with arguments if s is an array.
//
// x run s : run command s, with input string x as standard input.
//
// In the first form, standard input and error are inherited from the
// parent. In the second form, only standard error is inherited.
// Both commands return their own standard output, or an error.
func VFRun(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("run : too many arguments (%d)", len(args))
	}
	var cmds []string
	y := args[0]
	switch yv := y.BV().(type) {
	case goal.S:
		cmds = []string{string(yv)}
	case *goal.AS:
		cmds = yv.Slice()
	default:
		if len(args) == 2 {
			return panicType("x run s", "s", y)
		}
		return panicType("run s", "s", y)
	}
	if len(cmds) == 0 {
		return goal.NewPanic("run : empty command")
	}
	if len(args) == 1 {
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdin = os.Stdin
		var sb strings.Builder
		cmd.Stdout = &sb
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return cmdError(err, cmd, sb.String())
		}
		return goal.NewS(sb.String())
	}
	x := args[1]
	s, ok := x.BV().(goal.S)
	if !ok {
		return panicType("x run s", "x", x)
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = strings.NewReader(string(s))
	var sb strings.Builder
	cmd.Stdout = &sb
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return cmdError(err, cmd, sb.String())
	}
	return goal.NewS(sb.String())
}

// VFChdir implements the chdir monad.
//
// chdir s : change current directory to s, or return an error
//
// It returns a true value on success.
func VFChdir(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 1 {
		return goal.Panicf("chdir : too many arguments (%d)", len(args))
	}
	x := args[0]
	switch dir := x.BV().(type) {
	case goal.S:
		err := os.Chdir(string(dir))
		if err != nil {
			return goal.Errorf("%v", err)
		}
		return goal.NewI(1)
	default:
		return panicType("chdir s", "s", x)
	}
}

func panicType(op, sym string, x goal.V) goal.V {
	return goal.Panicf("%s : bad type \"%s\" in %s", op, x.Type(), sym)
}
