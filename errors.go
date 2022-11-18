package goal

import (
	"fmt"
	"strings"
)

// Position represents a source location, usually where an error occured.
type Position struct {
	Filename string // file name (as obtained from SetSource)
	Pos      int    // byte offset
	Lambda   *lambdaCode
}

// Error represents an error returned by any Context method.
type Error struct {
	Msg       string     // error text content
	Positions []Position // error location stack

	ctx *Context // context
}

// Error returns the default string representation. In practice, this is only
// used for debugging, because it uses raw byte offsets to report positions.
func (e *Error) Error() string {
	if len(e.Positions) == 0 {
		return e.Msg
	}
	sb := &strings.Builder{}
	sources := e.ctx.sources
	for i, pos := range e.Positions {
		if i == 0 {
			if pos.Filename != "" {
				s, line, col := getPosLine(sources[pos.Filename], pos.Pos)
				fmt.Fprintf(sb, "%s:%d:%d: %s\n",
					pos.Filename, line, col, e.Msg)
				writeLine(sb, s, col)
				continue
			}
			fmt.Fprintf(sb, "%s\n", e.Msg)
		}
		if pos.Filename != "" {
			s, line, col := getPosLine(sources[pos.Filename], pos.Pos)
			fmt.Fprintf(sb, "  (called from) %s:%d:%d:%d\n", pos.Filename, line, col, pos.Pos)
			writeLine(sb, s, col)
		} else if lc := pos.Lambda; lc != nil {
			s, _, col := getPosLine(lc.String, pos.Pos-lc.StartPos)
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

func errNYI(s string) E {
	return E("NYI: " + s)
}

func errType(x V) E {
	return E("bad type: `" + x.Type())
}

func errs(s string) E {
	return E(s)
}

func errf(format string, a ...interface{}) E {
	return E(fmt.Sprintf(format, a...))
}
