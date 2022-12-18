package goal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func sprintFloat(f float64) string {
	switch {
	case math.IsInf(f, 0):
		if f >= 0 {
			return "0w"
		}
		return "-0w"
	case math.IsNaN(f):
		return "0n"
	default:
		return fmt.Sprintf("%g", f)
	}
}

// Sprint returns a prettified string representation of the value.
func (v V) Sprint(ctx *Context) string {
	switch v.kind {
	case valInt:
		return fmt.Sprintf("%d", v.n)
	case valFloat:
		return sprintFloat(v.F())
	case valVariadic:
		if v.n < int64(len(ctx.variadicsNames)) {
			return ctx.variadicsNames[v.n]
		}
		return variadic(v.n).String()
	case valLambda:
		if v.n < 0 || v.n >= int64(len(ctx.lambdas)) {
			return fmt.Sprintf("{Lambda %d}", v.n)
		}
		return ctx.lambdas[v.n].Source
	case valBoxed:
		return v.value.Sprint(ctx)
	default:
		return ""
	}
}

func (e *errV) Sprint(ctx *Context) string { return "error[" + e.V.Sprint(ctx) + "]" }

func (e panicV) Sprint(ctx *Context) string { return "panic[" + string(e) + "]" }

// Sprint returns a properly quoted string.
func (s S) Sprint(ctx *Context) string { return strconv.Quote(string(s)) }

func (x *AB) Sprint(ctx *Context) string {
	return x.String()
}

func (x *AB) String() string {
	if x.Len() == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", b2i(x.At(0)))
		return sb.String()
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%d", b2i(xi))
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x *AI) Sprint(ctx *Context) string {
	return x.String()
}

func (x *AI) String() string {
	if x.Len() == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", x.At(0))
		return sb.String()
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%d", xi)
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x *AF) Sprint(ctx *Context) string {
	return x.String()
}

func (x *AF) String() string {
	if x.Len() == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		sb.WriteString(sprintFloat(x.At(0)))
		return sb.String()
	}
	for i, xi := range x.Slice {
		sb.WriteString(sprintFloat(xi))
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x *AS) Sprint(ctx *Context) string {
	return x.String()
}

func (x *AS) String() string {
	if x.Len() == 0 {
		return `0#""`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%q", x.At(0))
		return sb.String()
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%q", xi)
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

// sprintV returns a string for a V deep in an AV.
func sprintV(ctx *Context, x V) string {
	avx, ok := x.value.(*AV)
	if ok {
		return avx.sprint(ctx, true)
	}
	return x.Sprint(ctx)
}

func (x *AV) Sprint(ctx *Context) string {
	return x.sprint(ctx, false)
}

func (x *AV) sprint(ctx *Context, deep bool) string {
	if x.Len() == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%s", x.At(0).Sprint(ctx))
		return sb.String()
	}
	sb.WriteRune('(')
	var sep string
	if deep {
		sep = ";"
	} else {
		sep = "\n "
	}
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		sep = " "
	}
	for i, xi := range x.Slice {
		if xi.kind != valNil {
			fmt.Fprintf(sb, "%s", sprintV(ctx, xi))
		}
		if i < x.Len()-1 {
			sb.WriteString(sep)
		}
	}
	sb.WriteRune(')')
	return sb.String()
}

func (p projection) Sprint(ctx *Context) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s", p.Fun.Sprint(ctx))
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.kind != valNil {
			fmt.Fprintf(sb, "%s", arg.Sprint(ctx))
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p projectionFirst) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[%s;]", p.Fun.Sprint(ctx), p.Arg.Sprint(ctx))
}

func (p projectionMonad) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[]", p.Fun.Sprint(ctx))
}

func (r derivedVerb) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s%s", r.Arg.Sprint(ctx), r.Fun.String())
}
