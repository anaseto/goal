package goal

import (
	"fmt"
	"strconv"
	"strings"
)

// Sprint returns a prettified string representation of the value.
func (v V) Sprint(ctx *Context) string {
	switch v.Kind {
	case Int:
		return fmt.Sprintf("%d", v.N)
	case Float:
		return fmt.Sprintf("%g", v.F())
	case Variadic:
		return variadic(v.N).String()
	case Lambda:
		if v.N < 0 || v.N >= int64(len(ctx.lambdas)) {
			return fmt.Sprintf("{Lambda %d}", v.N)
		}
		return ctx.lambdas[v.N].Source
	case Boxed:
		return v.Value.Sprint(ctx)
	default:
		return ""
	}
}

// String returns a prettified string representation of the value.
func (v V) String() string {
	switch v.Kind {
	case Int:
		return fmt.Sprintf("%d", v.N)
	case Float:
		return fmt.Sprintf("%g", v.F())
	case Variadic:
		return variadic(v.N).String()
	case Lambda:
		return fmt.Sprintf("{Lambda %d}", v.N)
	case Boxed:
		return v.Value.String()
	default:
		return ""
	}
}

func (e panicV) Sprint(ctx *Context) string { return e.String() }
func (e panicV) String() string             { return "'ERROR " + string(e) }

// Sprint returns a properly quoted string.
func (s S) Sprint(ctx *Context) string { return strconv.Quote(string(s)) }
func (s S) String() string             { return strconv.Quote(string(s)) }

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
		fmt.Fprintf(sb, "%d", B2I(x.At(0)))
		return sb.String()
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%d", B2I(xi))
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
		fmt.Fprintf(sb, "%g", x.At(0))
		return sb.String()
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%g", xi)
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
	avx, ok := x.Value.(*AV)
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
		if xi.Kind != Nil {
			fmt.Fprintf(sb, "%s", sprintV(ctx, xi))
		}
		if i < x.Len()-1 {
			sb.WriteString(sep)
		}
	}
	sb.WriteRune(')')
	return sb.String()
}

func (x *AV) String() string {
	if x.Len() == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%s", x.At(0).String())
		return sb.String()
	}
	sb.WriteRune('(')
	sep := ";"
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		sep = " "
	}
	for i, xi := range x.Slice {
		if xi.Kind != Nil {
			fmt.Fprintf(sb, "%s", xi.String())
		}
		if i < x.Len()-1 {
			sb.WriteString(sep)
		}
	}
	sb.WriteRune(')')
	return sb.String()
}

func (p Projection) Sprint(ctx *Context) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s", p.Fun.Sprint(ctx))
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.Kind != Nil {
			fmt.Fprintf(sb, "%s", arg.Sprint(ctx))
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p Projection) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s", p.Fun.String())
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.Kind != Nil {
			fmt.Fprintf(sb, "%s", arg.String())
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p ProjectionFirst) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[%s;]", p.Fun.Sprint(ctx), p.Arg.Sprint(ctx))
}

func (p ProjectionFirst) String() string {
	return fmt.Sprintf("%s[%s;]", p.Fun.String(), p.Arg.String())
}

func (p ProjectionMonad) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[]", p.Fun.Sprint(ctx))
}

func (p ProjectionMonad) String() string {
	return fmt.Sprintf("%s[]", p.Fun.String())
}

func (r DerivedVerb) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s%s", r.Arg.Sprint(ctx), r.Fun.String())
}

func (r DerivedVerb) String() string {
	return fmt.Sprintf("%s%s", r.Arg.String(), r.Fun.String())
}
