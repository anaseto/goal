// Package os provides variadic function definitions for IO/OS builtins.
package os

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
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
		p, ok := pfx.Value().(goal.S)
		if !ok {
			return panicType("x import s", "x", pfx)
		}
		prefix = string(p)
		hasPfx = true
	}
	s := args[0]
	switch sv := s.Value().(type) {
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
		prefix = path.Base(name)
		prefix = strings.TrimSuffix(prefix, path.Ext(prefix))
	}
	fname := name
	if path.Ext(fname) == "" {
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
		fpath := path.Join(dir, fname)
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
	switch wv := w.Value().(type) {
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
	switch xv := x.Value().(type) {
	case goal.S:
		_, err := fmt.Print(string(xv))
		return err
	case *goal.AS:
		buf := bufio.NewWriter(os.Stdout)
		for _, s := range xv.Slice() {
			buf.WriteString(s)
		}
		return buf.Flush()
	default:
		_, err := fmt.Printf("%s", x.Append(ctx, nil))
		return err
	}
}

func sayV(ctx *goal.Context, x goal.V) error {
	switch xv := x.Value().(type) {
	case goal.S:
		_, err := fmt.Println(string(xv))
		return err
	case *goal.AS:
		buf := bufio.NewWriter(os.Stdout)
		for _, s := range xv.Slice() {
			buf.WriteString(s)
		}
		buf.WriteByte('\n')
		return buf.Flush()
	default:
		_, err := fmt.Printf("%s\n", x.Append(ctx, nil))
		return err
	}
}

func fprintV(ctx *goal.Context, w io.Writer, x goal.V) error {
	switch xv := x.Value().(type) {
	case goal.S:
		_, err := fmt.Fprint(w, string(xv))
		return err
	case *goal.AS:
		var err error
		for _, s := range xv.Slice() {
			_, err = fmt.Fprint(w, s)
		}
		return err
	default:
		_, err := w.Write(x.Append(ctx, nil))
		return err
	}
}

func fsayV(ctx *goal.Context, w io.Writer, x goal.V) error {
	switch xv := x.Value().(type) {
	case goal.S:
		_, err := fmt.Fprintln(w, string(xv))
		return err
	case *goal.AS:
		for _, s := range xv.Slice() {
			fmt.Fprint(w, s)
		}
		_, err := fmt.Fprint(w, '\n')
		return err
	default:
		_, err := w.Write(append(x.Append(ctx, nil), '\n'))
		return err
	}
}

// VFShell implements the shell monad.
//
// shell s : sends command s to the shell as-is. It returns the standard output
// of the command, or an error. Standard error is inherited from the parent.
func VFShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 1 {
		return goal.Panicf("shell : too many arguments (%d)", len(args))
	}
	var cmds string
	switch arg := args[len(args)-1].Value().(type) {
	case goal.S:
		cmds = string(arg)
	default:
		return panicType("shell s", "s", args[len(args)-1])
	}
	cmd := exec.Command("/bin/sh", "-c", cmds)
	cmd.Stderr = os.Stderr
	var sb strings.Builder
	cmd.Stdout = &sb
	err := cmd.Run()
	if err != nil {
		return goal.Errorf("%v", err)
	}
	return goal.NewS(sb.String())
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
	var cmds []string
	y := args[0]
	switch yv := y.Value().(type) {
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
	switch len(args) {
	case 1:
		cmd := exec.Command(cmds[0], cmds[1:]...)
		cmd.Stdin = os.Stdin
		var sb strings.Builder
		cmd.Stdout = &sb
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return goal.Errorf("%v", err)
		}
		return goal.NewS(sb.String())
	case 2:
		x := args[1]
		s, ok := x.Value().(goal.S)
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
			return goal.Errorf("%v", err)
		}
		return goal.NewS(sb.String())
	default:
		return goal.Panicf("run : too many arguments (%d)", len(args))
	}
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
	switch dir := x.Value().(type) {
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
