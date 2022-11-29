package goal

import (
	"fmt"
	"strings"
)

// position represents a source location, usually where an error occured.
type position struct {
	Filename string // file name (as obtained from SetSource)
	Pos      int    // byte offset
	lambda   *lambdaCode
}

// Error represents an error returned by any Context method.
type Error struct {
	Msg string // error message (without location)

	compile   bool
	positions []position        // error location stack
	sources   map[string]string // filename: source
}

// Error returns the default string representation. It makes uses of position
// information obtained from its running context.
func (e *Error) Error() string {
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

func errNYI(s string) V {
	return errs("NYI: " + s)
}

func errs(s string) V {
	return newBV(errV(s))
}

func errf(format string, a ...interface{}) V {
	return errs(fmt.Sprintf(format, a...))
}

// Errorf returns a formatted error value.
func Errorf(format string, a ...interface{}) V {
	return errs(fmt.Sprintf(format, a...))
}

// NewError returns an error value.
func NewError(s string) V {
	return errs(s)
}

func errType(op, sym string, x V) V {
	return errf("%s : bad type for %s (%s)", op, sym, x.Type())
}

func errDomain(op, s string) V {
	return errs(op + " : " + s)
}

func errRank(op string) V {
	return errs(op + " got too many arguments")
}
