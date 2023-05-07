package goal

import (
	"strings"
	"unicode/utf8"
)

// vfUTF8 implements the utf8 variadic verb.
func vfUTF8(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return utf8valid(args[0])
	case 2:
		x := args[1]
		s, ok := x.value.(S)
		if !ok {
			return panicType("x utf8 s", "x", x)
		}
		return toValidUTF8(string(s), args[0])
	default:
		return panicRank("utf8")
	}
}

func utf8valid(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewI(b2I(utf8.ValidString(string(xv))))
	case *AS:
		r := make([]byte, xv.Len())
		for i, s := range xv.elts {
			r[i] = b2B(utf8.ValidString(s))
		}
		return newABb(r)
	case *AV:
		return Canonical(monadAV(xv, utf8valid))
	default:
		return panicType("utf8 s", "s", x)
	}
}

func toValidUTF8(repl string, x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewS(strings.ToValidUTF8(string(xv), repl))
	case *AS:
		r := make([]string, xv.Len())
		for i, s := range xv.elts {
			r[i] = strings.ToValidUTF8(s, repl)
		}
		return NewAS(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = toValidUTF8(repl, xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panicType("x utf8 s", "s", x)
	}
}
