package goal

import (
	"strings"
	"unicode/utf8"
)

// vfUTF8RCount implements the "utf8.rcount" variadic verb.
func vfUTF8RCount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return utf8rcount(args[0])
	default:
		return panicRank("utf8.rcount")
	}
}

func utf8rcount(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewI(int64(utf8.RuneCountInString(string(xv))))
	case *AS:
		r := make([]int64, xv.Len())
		for i, s := range xv.elts {
			r[i] = int64(utf8.RuneCountInString(s))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = utf8rcount(xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalFast(NewAV(r))
	default:
		return panicType("utf8.rcount s", "s", x)
	}
}

// vfUTF8Valid implements the "utf8.valid" variadic verb.
func vfUTF8Valid(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return utf8valid(args[0])
	case 2:
		x := args[1]
		s, ok := x.value.(S)
		if !ok {
			return panicType("x utf8.valid s", "x", x)
		}
		return toValidUTF8(string(s), args[0])
	default:
		return panicRank("utf8.valid")
	}
}

func utf8valid(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewI(B2I(utf8.ValidString(string(xv))))
	case *AS:
		r := make([]bool, xv.Len())
		for i, s := range xv.elts {
			r[i] = utf8.ValidString(s)
		}
		return NewAB(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = utf8valid(xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panicType("utf8.valid s", "s", x)
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
		return panicType("x utf8.valid s", "s", x)
	}
}
