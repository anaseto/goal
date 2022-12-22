package io

import (
	"bufio"
	"fmt"
	"goal"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

// VImport implements the import dyad.
func VImport(ctx *goal.Context, args []goal.V) goal.V {
	var fname string
	var prefix string
	if len(args) > 2 {
		return goal.Panicf("import: too many arguments (%d)", len(args))
	}
	// TODO: VImport: support importing several files at once
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
		// TODO: check that fname is valid (otherwise identifiers could
		// not be written).
		prefix = string(s)
	}
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return goal.Panicf("import : %v", err)
	}
	r, err := ctx.EvalPackage(string(bytes), path.Base(fname), string(prefix))
	if err != nil {
		_, ok := err.(goal.ErrPackageImported)
		if ok {
			return goal.NewI(0)
		}
		return goal.Panicf("import : %v", err)
	}
	return r
}

// VPrint implements the print dyad.
//
// print x : outputs x to standard output. It returns a true value on success.
//
// w print x : outputs x to w, where w is an io.Writer or a filename (goal.S).
func VPrint(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		x := args[0]
		err := printV(ctx, x, false)
		if err != nil {
			return goal.Errorf("print x : %v", err)
		}
	case 2:
		w := args[1]
		var wout io.Writer
		switch wv := w.Value().(type) {
		case goal.S:
			var err error
			wout, err = os.Create(string(wv))
			if err != nil {
				return goal.Errorf("w print x : %v", err)
			}
		case io.Writer:
			wout = wv
		default:
			return goal.NewPanic("w print x : w should be a string or writer")
		}
		x := args[0]
		err := fprintV(ctx, wout, x, false)
		if err != nil {
			return goal.Errorf("w print x : %v", err)
		}
	default:
		return goal.NewPanic("print : too many arguments")
	}
	return goal.NewI(1)
}

// VSay implements the say dyad. It is the same as print, but appends a newline
// to the result.
func VSay(ctx *goal.Context, args []goal.V) goal.V {
	switch len(args) {
	case 1:
		x := args[0]
		err := printV(ctx, x, true)
		if err != nil {
			return goal.Errorf("say x : %v", err)
		}
	case 2:
		w := args[1]
		var wout io.Writer
		switch wv := w.Value().(type) {
		case goal.S:
			var err error
			wout, err = os.Create(string(wv))
			if err != nil {
				return goal.Errorf("w say x : %v", err)
			}
		case io.Writer:
			wout = wv
		default:
			return goal.NewPanic("w say x : w should be a string or writer")
		}
		x := args[0]
		err := fprintV(ctx, wout, x, true)
		if err != nil {
			return goal.Errorf("w say x : %v", err)
		}
	default:
		return goal.NewPanic("say : too many arguments")
	}
	return goal.NewI(1)
}

func printV(ctx *goal.Context, x goal.V, newline bool) error {
	switch xv := x.Value().(type) {
	case goal.S:
		if newline {
			_, err := fmt.Println(string(xv))
			return err
		}
		_, err := fmt.Print(string(xv))
		return err
	case *goal.AS:
		buf := bufio.NewWriter(os.Stdout)
		for i, s := range xv.Slice {
			buf.WriteString(s)
		}
		if newline {
			buf.WriteRune('\n')
		}
		return buf.Flush()
	default:
		if newline {
			_, err := fmt.Println(x.Sprint(ctx))
			return err
		}
		_, err := fmt.Print(x.Sprint(ctx))
		return err
	}
}

func fprintV(ctx *goal.Context, w io.Writer, x goal.V, newline bool) error {
	buf := bufio.NewWriter(w)
	switch xv := x.Value().(type) {
	case goal.S:
		buf.WriteString(string(xv))
		if newline {
			buf.WriteRune('\n')
		}
	case *goal.AS:
		for i, s := range xv.Slice {
			buf.WriteString(s)
		}
		if newline {
			buf.WriteRune('\n')
		}
	default:
		if newline {
			fmt.Fprintln(buf, x.Sprint(ctx))
		} else {
			fmt.Fprint(buf, x.Sprint(ctx))
		}
	}
	return buf.Flush()
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
			return goal.NewS(string(bytes))
		default:
			return goal.NewPanic("slurp: non-string filename")
		}
	default:
		return goal.NewPanic("slurp: too many arguments")
	}
}

// VShell implements the shell monad.
//
// shell cmd : sends cmd to the shell as-is. It returns the standard output of
// the command, or an error.
func VShell(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) == 0 {
		return goal.NewPanic("shell: missing command string")
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
