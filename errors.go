package goal

import (
	"errors"
	"fmt"
	"strings"
)

// PanicError represents a fatal error returned by any Context method.
type PanicError struct {
	Msg string // error message (without location)

	compile   bool
	positions []position        // error location stack
	sources   map[string]string // filename: source
}

// position represents a source location, usually where an error occured.
type position struct {
	Filename string // file name (as obtained from SetSource)
	Pos      int    // byte offset
	lambda   *lambdaCode
}

// Error returns the default string representation. It makes uses of position
// information obtained from its running context.
func (e *PanicError) Error() string {
	if len(e.positions) == 0 {
		return e.Msg
	}
	sb := &strings.Builder{}
	sources := e.sources
	for i, pos := range e.positions {
		if i == 0 {
			if pos.Filename != "" {
				s, line, col := getPosLine(sources[pos.Filename], pos.Pos)
				fmt.Fprintf(sb, "%s:%d:%d: %s\n",
					pos.Filename, line, col+1, e.Msg)
				writeLine(sb, s, col)
				continue
			}
			fmt.Fprintf(sb, "%s\n", e.Msg)
		}
		if pos.Filename != "" {
			s, line, col := getPosLine(sources[pos.Filename], pos.Pos)
			ctxs := "called from"
			if e.compile {
				ctxs = "from"
			}
			fmt.Fprintf(sb, "  (%s) %s:%d:%d:%d\n", ctxs, pos.Filename, line, col+1, pos.Pos)
			writeLine(sb, s, col)
		} else if lc := pos.lambda; lc != nil {
			s, _, col := getPosLine(lc.Source, pos.Pos-lc.StartPos)
			writeLine(sb, s, col)
		} else {
			s, _, col := getPosLine(sources[""], pos.Pos)
			writeLine(sb, s, col)
		}
	}
	return sb.String()
}

func writeLine(sb *strings.Builder, s string, col int) {
	if s == "" {
		return
	}
	sb.WriteString(s)
	sb.WriteRune('\n')
	if col > 0 {
		sb.WriteString(strings.Repeat(" ", col))
	}
	sb.WriteRune('^')
	sb.WriteRune('\n')
}

func getPosLine(s string, pos int) (string, int, int) {
	if pos > len(s) {
		return "", 0, 0
	}
	start := 0
	count := 1
	for i, r := range s {
		if r == '\n' {
			if i < pos {
				start = i + 1
				count++
				continue
			}
			return s[start:i], count, pos - start
		}
	}
	return s[start:], count, pos - start
}

// newExecError returns an execution error from a panicV value. It assumes
// isPanic(x) is true and the Value type is panicV.
func newExecError(x V) error {
	return errors.New(string(x.value.(panicV)))
}

func panics(s string) V {
	return V{kind: valPanic, value: panicV(s)}
}

func panicf(format string, a ...interface{}) V {
	return panics(fmt.Sprintf(format, a...))
}

// Panicf returns a formatted fatal error value.
func Panicf(format string, a ...interface{}) V {
	return panics(fmt.Sprintf(format, a...))
}

// NewPanic returns a fatal error value.
func NewPanic(s string) V {
	return panics(s)
}

func panicType(op, sym string, x V) V {
	return panicf("%s : bad type for %s (%s)", op, sym, x.Type())
}

func panicTypeElt(op, sym string, x V) V {
	return panicf("%s : bad type in %s (%s)", op, sym, x.Type())
}

func panicDomain(op, s string) V {
	return panics(op + " : " + s)
}

func panicRank(op string) V {
	return panics(op + " got too many arguments")
}
