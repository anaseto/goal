package os

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"codeberg.org/anaseto/goal"
)

type file struct {
	f    *os.File
	b    *bufio.ReadWriter
	mode string
}

func (f *file) Matches(y goal.Value) bool {
	switch yv := y.(type) {
	case *file:
		return f.f.Fd() == yv.f.Fd()
	case *command:
		return false
	default:
		return false
	}
}

func (f *file) Fprint(ctx *goal.Context, w goal.ValueWriter) (n int, err error) {
	m, err := fmt.Fprint(w, "open[\"", f.mode, "\";")
	n += m
	if err != nil {
		return
	}
	m, err = goal.S(f.f.Name()).Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(']')
	if err != nil {
		return
	}
	n++
	return
}

func (f *file) Type() string {
	return "h"
}

func (f *file) Less(y goal.Value) bool {
	switch yv := y.(type) {
	case *file:
		return f.f.Fd() < yv.f.Fd()
	default:
		return f.Type() < y.Type()
	}
}

func (f *file) Read(p []byte) (n int, err error) {
	return f.b.Read(p)
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.b.Write(p)
}

func (f *file) Close() error {
	f.b.Writer.Flush()
	return f.f.Close()
}

type command struct {
	c      *exec.Cmd
	b      *bufio.ReadWriter
	mode   string // -| or |-
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func cmdToAS(cmd *command) goal.V {
	return goal.NewAS(cmd.c.Args)
}

func (cmd *command) Matches(y goal.Value) bool {
	switch yv := y.(type) {
	case *command:
		return cmd.mode == yv.mode && goal.Match(cmdToAS(cmd), cmdToAS(yv))
	default:
		return false
	}
}

func (cmd *command) Fprint(ctx *goal.Context, w goal.ValueWriter) (n int, err error) {
	m, err := fmt.Fprint(w, "open[\"", cmd.mode, "\";")
	n += m
	if err != nil {
		return
	}
	m, err = cmdToAS(cmd).Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(']')
	if err != nil {
		return
	}
	n++
	return
}

func (cmd *command) Type() string {
	return "h"
}

func (cmd *command) Less(y goal.Value) bool {
	switch yv := y.(type) {
	case *command:
		return cmd.mode < yv.mode || cmd.mode == yv.mode && cmdToAS(cmd).Less(cmdToAS(yv))
	case *file:
		return true
	default:
		return cmd.Type() < y.Type()
	}
}

func (cmd *command) Read(p []byte) (n int, err error) {
	if cmd.b.Reader == nil {
		return 0, errors.New("write-only")
	}
	return cmd.b.Read(p)
}

func (cmd *command) Write(p []byte) (n int, err error) {
	if cmd.b.Writer == nil {
		return 0, errors.New("read-only")
	}
	return cmd.b.Write(p)
}

func (cmd *command) Close() error {
	if cmd.b.Writer != nil {
		cmd.b.Writer.Flush()
	}
	if cmd.stdin != nil {
		cmd.stdin.Close()
	}
	if cmd.stdout != nil {
		cmd.stdout.Close()
	}
	return cmd.c.Wait()
}

// VOpen implements the open dyad.
//
// open "path" : opens file "path" for reading
//
// mode open "path" : opens file "path" using the given fopen(3) mode
//
// mode can be: "r", "r+", "w", "w+", "a", "a+", "|-", "-|"
//
// It returns a filehandle value of type "h" on success, and an error
// otherwise.
func VOpen(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("open : too many arguments (%d)", len(args))
	}
	var mode goal.S = "r"
	if len(args) == 2 {
		var ok bool
		mode, ok = args[1].Value().(goal.S)
		if !ok {
			return goal.Panicf("mode open path : mode not a string (%s)", args[1].Type())
		}
	}
	m := string(mode)
	var flag int
	switch m {
	case "r":
		flag = os.O_RDONLY
	case "r+":
		flag = os.O_RDWR
	case "w":
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case "w+":
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case "a":
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case "a+":
		flag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	case "-|", "|-":
		return openPipe(m, args[0])
	default:
		return goal.Panicf("mode open path : invalid mode (%s)", m)
	}
	path, ok := args[0].Value().(goal.S)
	if !ok {
		pfx := ""
		if len(args) == 2 {
			pfx = "mode "
		}
		return goal.Panicf(pfx+"open path : path not a string (%s)", args[0].Type())
	}
	f, err := os.OpenFile(string(path), flag, 0666)
	if err != nil {
		return goal.Errorf("open : %v", err)
	}
	b := bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f))
	return goal.NewV(&file{f: f, b: b, mode: m})
}

func openPipe(m string, c goal.V) goal.V {
	var cmd *exec.Cmd
	switch cv := c.Value().(type) {
	case goal.S:
		cmd = exec.Command(string(cv))
	case *goal.AS:
		if cv.Len() == 0 {
			return goal.NewPanic("mode open cmd : empty cmd")
		}
		cmd = exec.Command(cv.Slice[0], cv.Slice[1:]...)
	default:
		return goal.Panicf("mode open cmd : non-string cmd (%s)", c.Type())
	}
	r := &command{c: cmd, mode: m}
	switch m {
	case "|-":
		wc, err := cmd.StdinPipe()
		if err != nil {
			return goal.Errorf("\"|-\" open cmd : %v", err)
		}
		cmd.Stdout = os.Stdout
		r.stdin = wc
		r.b = bufio.NewReadWriter(nil, bufio.NewWriter(wc))
	case "-|":
		rc, err := cmd.StdoutPipe()
		if err != nil {
			return goal.Errorf("\"-|\" open cmd : %v", err)
		}
		cmd.Stdin = os.Stdin
		r.stdout = rc
		r.b = bufio.NewReadWriter(bufio.NewReader(rc), nil)
	}
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return goal.Errorf("open : %v", err)
	}
	return goal.NewV(r)
}

// VClose implements the close monad.
//
// close h : closes a filehandle
//
// It returns a true value on success, and an error otherwise.
func VClose(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 1 {
		return goal.Panicf("close : too many arguments (%d)", len(args))
	}
	switch h := args[0].Value().(type) {
	case io.Closer:
		err := h.Close()
		if err != nil {
			return goal.Errorf("close : %v", err)
		}
		return goal.NewI(1)
	default:
		return goal.Panicf("close h : h not a handle (%s)", args[0].Type())
	}
}

func isI(x float64) bool {
	return x == float64(int64(x))
}

// VRead implements the read monad.
//
// read h : reads from filehandle h until EOF or an error occurs.
// n read h : reads n bytes from filehandle h until EOF or an error occurs.
//
// It returns the read content as a string on success, and an error otherwise.
func VRead(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("read : too many arguments (%d)", len(args))
	}
	var n int64 = -1
	if len(args) == 2 {
		x := args[1]
		if x.IsI() {
			n = x.I()
		} else if x.IsF() {
			if !isI(x.F()) {
				return goal.Panicf("n read h : n not a integer (%g)", x.F())
			}
			n = int64(x.F())
		} else {
			return goal.Panicf("n read h : n not a integer (%s)", x.Type())
		}
	}
	switch h := args[0].Value().(type) {
	case io.Reader:
		if n < 0 {
			sb := strings.Builder{}
			_, err := io.Copy(&sb, h)
			if err != nil {
				return goal.Errorf("read : %v", err)
			}
			return goal.NewS(sb.String())
		}
		sb := strings.Builder{}
		_, err := io.CopyN(&sb, h, n)
		if err != nil {
			return goal.Errorf("read : %v", err)
		}
		return goal.NewS(sb.String())
	default:
		return goal.Panicf("read h : h not a handle (%s)", args[0].Type())
	}
}
