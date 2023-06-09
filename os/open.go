package os

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"codeberg.org/anaseto/goal"
)

type file struct {
	f    *os.File
	b    *bufio.ReadWriter
	mode string
}

func newFile(f *os.File) goal.V {
	b := bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f))
	return goal.NewV(&file{f: f, b: b, mode: "s"})
}

// NewStdHandle returns a buffered handle for the given file.
func NewStdHandle(f *os.File) goal.V {
	return newFile(f)
}

func (f *file) Matches(y goal.BV) bool {
	switch yv := y.(type) {
	case *file:
		return f.f.Fd() == yv.f.Fd()
	case *command:
		return false
	default:
		return false
	}
}

func (f *file) Append(ctx *goal.Context, dst []byte) []byte {
	dst = append(dst, "open["...)
	dst = strconv.AppendQuote(dst, f.mode)
	dst = append(dst, ';')
	dst = strconv.AppendQuote(dst, f.f.Name())
	dst = append(dst, ']')
	return dst
}

func (f *file) Type() string {
	return "h"
}

func (f *file) LessT(y goal.BV) bool {
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
	c     *exec.Cmd
	b     *bufio.ReadWriter
	mode  string // -| or |-
	stdin io.WriteCloser
}

func cmdToAS(cmd *command) goal.V {
	return goal.NewAS(cmd.c.Args)
}

func (cmd *command) Matches(y goal.BV) bool {
	switch yv := y.(type) {
	case *command:
		return cmd.mode == yv.mode && cmdToAS(cmd).Matches(cmdToAS(yv))
	default:
		return false
	}
}

func (cmd *command) Append(ctx *goal.Context, dst []byte) []byte {
	dst = append(dst, "open["...)
	dst = strconv.AppendQuote(dst, cmd.mode)
	dst = append(dst, ';')
	dst = cmdToAS(cmd).Append(ctx, dst)
	dst = append(dst, ']')
	return dst
}

func (cmd *command) Type() string {
	return "h"
}

func (cmd *command) LessT(y goal.BV) bool {
	switch yv := y.(type) {
	case *command:
		return cmd.mode < yv.mode || cmd.mode == yv.mode && cmdToAS(cmd).LessT(cmdToAS(yv))
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
	return cmd.c.Wait()
}

// VFOpen implements the open dyad.
//
// open "path" : opens file "path" for reading.
//
// x open "path" : opens file "path" using the given fopen(3) mode x.
//
// x can be: "r", "r+", "w", "w+", "a", "a+", "|-", "-|". In the last two
// modes, the path is instead interpreted as a command, and can be a list of
// strings.
//
// It returns a filehandle value of type "h" on success, and an error
// otherwise.
func VFOpen(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("open : too many arguments (%d)", len(args))
	}
	var mode goal.S = "r"
	if len(args) == 2 {
		var ok bool
		mode, ok = args[1].BV().(goal.S)
		if !ok {
			return panicType("x open s", "x", args[1])
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
	path, ok := args[0].BV().(goal.S)
	if !ok {
		if len(args) == 2 {
			return panicType("x open s", "s", args[0])
		}
		return panicType("open s", "s", args[0])
	}
	f, err := os.OpenFile(string(path), flag, 0666)
	if err != nil {
		return goal.Errorf("%v", err)
	}
	b := bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f))
	return goal.NewV(&file{f: f, b: b, mode: m})
}

func openPipe(m string, c goal.V) goal.V {
	var cmd *exec.Cmd
	switch cv := c.BV().(type) {
	case goal.S:
		cmd = exec.Command(string(cv))
	case *goal.AS:
		if cv.Len() == 0 {
			return goal.NewPanic("mode open cmd : empty cmd")
		}
		cmd = exec.Command(cv.Slice()[0], cv.Slice()[1:]...)
	default:
		return panicType("x open s", "s", c)
	}
	r := &command{c: cmd, mode: m}
	switch m {
	case "|-":
		wc, err := cmd.StdinPipe()
		if err != nil {
			return goal.Errorf("%v", err)
		}
		cmd.Stdout = os.Stdout
		r.stdin = wc
		r.b = bufio.NewReadWriter(nil, bufio.NewWriter(wc))
	case "-|":
		rc, err := cmd.StdoutPipe()
		if err != nil {
			return goal.Errorf("%v", err)
		}
		cmd.Stdin = os.Stdin
		r.b = bufio.NewReadWriter(bufio.NewReader(rc), nil)
	}
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return goal.Errorf("%v", err)
	}
	return goal.NewV(r)
}

// VFClose implements the close monad.
//
// close h : closes a filehandle.
//
// It returns a true value on success, and an error otherwise.
func VFClose(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 1 {
		return goal.Panicf("close : too many arguments (%d)", len(args))
	}
	switch h := args[0].BV().(type) {
	case io.Closer:
		err := h.Close()
		if err != nil {
			if e, ok := err.(*exec.ExitError); ok {
				return exitError(e)
			}
			return goal.Errorf("%v", err)
		}
		return goal.NewI(1)
	default:
		return panicType("close h", "h", args[0])
	}
}

func exitError(err *exec.ExitError) goal.V {
	keys := goal.NewAS([]string{"code", "msg"})
	values := goal.NewAV([]goal.V{goal.NewI(int64(err.ProcessState.ExitCode())), goal.NewS(err.Error())})
	return goal.NewError(goal.NewD(keys, values))
}

func isI(x float64) bool {
	return x == float64(int64(x))
}

// VFRead implements the read dyad.
//
// read h : reads from filehandle h until EOF or an error occurs.
//
// read s : reads file named s
//
// s read h : reads from filehandle h until delimiter s or EOF, or an error
// occurs.
//
// i read h : reads i bytes from filehandle h until EOF (returning possibly
// less than i bytes) or an error occurs.
//
// It returns the read content as a string on success, and an error otherwise.
func VFRead(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 2 {
		return goal.Panicf("read : too many arguments (%d)", len(args))
	}
	var n int64 = -1
	y := args[0]
	if len(args) == 2 {
		x := args[1]
		if x.IsI() {
			n = x.I()
		} else if x.IsF() {
			if !isI(x.F()) {
				return goal.Panicf("i read h : non-integer i (%g)", x.F())
			}
			n = int64(x.F())
		} else {
			s, ok := x.BV().(goal.S)
			if ok {
				return readString(y, string(s))
			}
			return panicType("x read h", "x", x)
		}
	}
	switch yv := y.BV().(type) {
	case goal.S:
		if len(args) != 1 {
			break
		}
		s, err := readFile(string(yv))
		if err != nil {
			return goal.NewError(goal.NewS(err.Error()))
		}
		return goal.NewS(s)
	case io.Reader:
		if n < 0 {
			sb := strings.Builder{}
			_, err := io.Copy(&sb, yv)
			if err != nil {
				return goal.Errorf("%v", err)
			}
			return goal.NewS(sb.String())
		}
		sb := strings.Builder{}
		_, err := io.CopyN(&sb, yv, n)
		s := sb.String()
		if err != nil && (err != io.EOF || s == "") {
			return goal.Errorf("%v", err)
		}
		return goal.NewS(s)
	}
	if len(args) == 2 {
		return panicType("read h", "h", y)
	}
	return panicType("x read h", "h", y)
}

func readFile(fname string) (string, error) {
	bytes, err := os.ReadFile(fname)
	if err != nil {
		return "", err
	}
	s := *(*string)(unsafe.Pointer(&bytes))
	return s, nil
}

func readString(h goal.V, delim string) goal.V {
	if len(delim) != 1 {
		return goal.Panicf("s read h : s not a 1-byte string (got %d bytes)", len(delim))
	}
	switch hv := h.BV().(type) {
	case *file:
		s, err := hv.b.Reader.ReadString(delim[0])
		if err != nil && (err != io.EOF || s == "") {
			return goal.Errorf("%v", err)
		}
		return goal.NewS(s)
	case *command:
		if hv.b.Reader == nil {
			return goal.NewPanic("write-only handle")
		}
		s, err := hv.b.Reader.ReadString(delim[0])
		if err != nil && (err != io.EOF || s == "") {
			return goal.Errorf("%v", err)
		}
		return goal.NewS(s)
	case io.Reader:
		b := bufio.NewReader(hv)
		s, err := b.ReadString(delim[0])
		if err != nil && (err != io.EOF || s == "") {
			return goal.Errorf("%v", err)
		}
		return goal.NewS(s)
	default:
		return panicType("s read h", "h", h)
	}
}

// VFFlush implements the flush monad.
//
// flush h : flushes any buffered data to h.
//
// It returns a true value on success.
func VFFlush(ctx *goal.Context, args []goal.V) goal.V {
	if len(args) > 1 {
		return goal.Panicf("flush : too many arguments (%d)", len(args))
	}
	x := args[0]
	switch xv := x.BV().(type) {
	case *file:
		err := xv.b.Writer.Flush()
		if err != nil {
			return goal.Errorf("%v", err)
		}
		return goal.NewI(1)
	case *command:
		if xv.b.Writer == nil {
			return goal.NewError(goal.NewS("read-only pipe"))
		}
		err := xv.b.Writer.Flush()
		if err != nil {
			return goal.Errorf("%v", err)
		}
		return goal.NewI(1)
	default:
		return panicType("flush h", "h", x)
	}
}
