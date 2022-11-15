package goal

import (
	"fmt"
	"strings"
)

// Error represents an error returned by any Context method.
type Error struct {
	Msg       string     // error text content
	Positions []Position // error location stack

	ctx *Context
}

// Error returns the default string representation. In practice, this is only
// used for debugging, because it uses raw byte offsets to report positions.
func (e *Error) Error() string {
	if len(e.Positions) == 0 {
		return e.Msg
	}
	sb := &strings.Builder{}
	last := e.Positions[len(e.Positions)-1]
	fmt.Fprintf(sb, "%s:%d:%s\n", last.Filename, last.Pos, e.Msg)
	for i := len(e.Positions) - 2; i >= 0; i-- {
		fmt.Fprintf(sb, "(call stack) %s:%d\n", last.Filename, last.Pos)
	}
	return sb.String()
}

// Position represents a source location, usually where an error occured.
type Position struct {
	Filename string // file name (as obtained from SetSource)
	Pos      int    // byte offset
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
