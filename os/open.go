package os

import (
	"bufio"
	"codeberg.org/anaseto/goal"
	"fmt"
	"io"
	"os"
	"strings"
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
	default:
		return false
	}
}

func (f *file) Fprint(ctx *goal.Context, w goal.ValueWriter) (n int, err error) {
	return fmt.Fprintf(w, "open[", goal.S(f.mode), ",", goal.S(f.f.Name()), "]")
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
	if f.b.Writer.Buffered() > 0 {
		f.b.Writer.Flush()
	}
	return f.f.Close()
}

// VOpen implements the open dyad.
//
// open "path" : opens file "path" for reading
//
// mode open "path" : opens file "path" using the given fopen(3) mode
//
// mode can be: "r", "r+", "w", "w+", "a", "a+"
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
	path, ok := args[0].Value().(goal.S)
	if !ok {
		pfx := ""
		if len(args) == 2 {
			pfx = "mode "
		}
		return goal.Panicf(pfx+"open path : path not a string (%s)", args[0].Type())
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
	default:
		return goal.Panicf("mode open path : invalid mode (%s)", m)
	}
	f, err := os.OpenFile(string(path), flag, 0666)
	if err != nil {
		return goal.Errorf("open : %v", err)
	}
	b := bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f))
	return goal.NewV(&file{f: f, b: b, mode: m})
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
			return goal.NewError(goal.NewS("close : " + err.Error()))
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
				return goal.NewError(goal.NewS("read : " + err.Error()))
			}
			return goal.NewS(sb.String())
		}
		sb := strings.Builder{}
		_, err := io.CopyN(&sb, h, n)
		if err != nil {
			return goal.NewError(goal.NewS("read : " + err.Error()))
		}
		return goal.NewS(sb.String())
	default:
		return goal.Panicf("read h : h not a handle (%s)", args[0].Type())
	}
}
