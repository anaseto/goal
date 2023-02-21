package os

import (
	"bufio"
	"codeberg.org/anaseto/goal"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

// VImport implements the import dyad.
//
// import "name" : evaluate file "name.goal" with prefix "name"
//
// pfx import "name" : evaluate file "name.goal" with custom prefix pfx
//
// It returns 0 and does nothing if a file has already been evaluated.
func VImport(ctx *goal.Context, args []goal.V) goal.V {
	var fname string
	var prefix string
	if len(args) > 2 {
		return goal.Panicf("import : too many arguments (%d)", len(args))
	}
	// TODO: VImport: support importing several files at once?
	s, ok := args[0].Value().(goal.S)
	if !ok {
		return goal.Panicf("import name : name not a string (%s)", args[0].Type())
	}
	if strings.ContainsRune(string(s), '.') {
		return goal.Panicf("import name : name should not include extension .", args[0].Type())
	}
	fname = string(s) + ".goal"
	if len(args) == 2 {
		p, ok := args[1].Value().(goal.S)
		if !ok {
			return goal.Panicf("prefix import name : prefix not a string (%s)", args[1].Type())
		}
		prefix = string(p)
	} else {
		// TODO: check that prefix is valid (otherwise identifiers could
		// not be written).
		prefix = path.Base(string(s))
	}
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return goal.Panicf("import : %v", err)
	}
	r, err := ctx.EvalPackage(string(bytes), fname, string(prefix))
	if err != nil {
		_, ok := err.(goal.ErrPackageImported)
		if ok {
			return goal.NewI(0)
		}
		return goal.Panicf("import : %v", err)
	}
	return r
}

func ppanic(pfx string, x goal.V) goal.V {
	return goal.NewPanic(pfx + x.Panic())
}

// VPrint implements the print dyad.
//
// print x : outputs x to standard output. It returns a true value on success.
//
// h print y : outputs y to w, where w is an io.Writer or a filename (goal.S).
func VPrint(ctx *goal.Context, args []goal.V) goal.V {
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

// VSay implements the say dyad. It is the same as print, but appends a newline
// to the result.
func VSay(ctx *goal.Context, args []goal.V) goal.V {
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
		for _, s := range xv.Slice {
			buf.WriteString(s)
		}
		return buf.Flush()
	default:
		_, err := fmt.Print(x.Append(ctx, nil))
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
		for _, s := range xv.Slice {
			buf.WriteString(s)
		}
		buf.WriteByte('\n')
		return buf.Flush()
	default:
		_, err := fmt.Println(x.Append(ctx, nil))
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
		for _, s := range xv.Slice {
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
		for _, s := range xv.Slice {
			fmt.Fprint(w, s)
		}
		_, err := fmt.Fprint(w, '\n')
		return err
	default:
		_, err := w.Write(append(x.Append(ctx, nil), '\n'))
		return err
	}
}

// VSlurp implements the slurp monad.
//
// slurp x returns the contents of the file with filename x, or an error.
func VSlurp(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		switch x := args[0].Value().(type) {
		case goal.S:
			bytes, err := os.ReadFile(string(x))
			if err != nil {
				return goal.NewError(goal.NewS(err.Error()))
			}
			// TODO: avoid allocation by using Copy and
			// strings.Builder, or maybe unsafe.
			return goal.NewS(string(bytes))
		default:
			return goal.NewPanic("slurp : non-string filename")
		}
	default:
		return goal.NewPanic("slurp : too many arguments")
	}
}

// VShell implements the shell monad.
//
// shell cmd : sends cmd to the shell as-is. It returns the standard output of
// the command, or an error. Standard error is inherited from the parent.
func VShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) == 0 {
		return goal.NewPanic("shell : missing command string")
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
		return goal.NewError(goal.NewS(err.Error()))
	}
	return goal.NewS(string(bytes))
}

// VRun implements the run monad.
//
// run s : run command s, with arguments if s is an array.
//
// Standard input, output, and error are inherited from the parent.
// It returns a true value on success, and an error otherwise.
func VRun(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) == 0 {
		return goal.NewPanic("run : missing command string")
	}
	if len(args) > 1 {
		return goal.Panicf("run : too many arguments (%d)", len(args))
	}
	var cmds []string
	switch arg := args[len(args)-1].Value().(type) {
	case goal.S:
		cmds = []string{string(arg)}
	case *goal.AS:
		cmds = arg.Slice
	default:
		return goal.Panicf("run s : bad type (%s)", arg.Type())
	}
	if len(cmds) == 0 {
		return goal.NewPanic("run s : empty command")
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		return goal.NewError(goal.NewS(err.Error()))
	}
	return goal.NewI(1)
}
