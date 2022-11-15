package goal

import (
	"fmt"
	"strings"
)

// Position represents a source location, usually where an error occured.
type Position struct {
	Filename string // file name (as obtained from SetSource)
	Pos      int    // byte offset
	Lambda   Lambda
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
	sources := e.ctx.sources
	sb := &strings.Builder{}
	last := e.Positions[len(e.Positions)-1]
	s, lnum, offset := getPosLine(sources[last.Filename], last.Pos)
	if last.Filename != "" {
		fmt.Fprintf(sb, "%s:%d: %s\n", last.Filename, lnum, e.Msg)
	} else {
		fmt.Fprintf(sb, "%s\n", e.Msg)
	}
	writeLine(sb, s, offset)
	for i := len(e.Positions) - 2; i >= 0; i-- {
		pos := e.Positions[i]
		s, lnum, offset := getPosLine(e.ctx.prog.LambdaStrings[int(pos.Lambda)], pos.Pos)
		if pos.Filename != "" {
			fmt.Fprintf(sb, "(called from) %s:%d\n", pos.Filename, lnum)
		} else {
			fmt.Fprint(sb, "(called from)\n")
		}
		writeLine(sb, s, offset)
	}
	return sb.String()
}

func writeLine(sb *strings.Builder, s string, offset int) {
	if s == "" {
		return
	}
	sb.WriteString(s)
	sb.WriteRune('\n')
	if offset > 0 {
		sb.WriteString(strings.Repeat(" ", offset-1))
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
