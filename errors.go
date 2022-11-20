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

func errNYI(s string) E {
	return E("NYI: " + s)
}

func errType(op, sym string, x V) E {
	return errf("%s : bad type for %s (%s)", op, sym, x.Type())
}

func errs(s string) E {
	return E(s)
}

func errf(format string, a ...interface{}) E {
	return E(fmt.Sprintf(format, a...))
}
