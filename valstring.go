package goal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func sprintFloat(sb *strings.Builder, f float64) {
	switch {
	case math.IsInf(f, 0):
		if f >= 0 {
			sb.WriteString("0w")
		} else {
			sb.WriteString("-0w")
		}
	case math.IsNaN(f):
		sb.WriteString("0n")
	default:
		fmt.Fprintf(sb, "%g", f)
	}
}

// Format returns a matching program string representation of the value.
func (v V) Format(ctx *Context) string {
	var sb strings.Builder
	v.Sprint(ctx, &sb)
	return sb.String()
}

// Sprint writes a matching program string representation of the value.
func (v V) Sprint(ctx *Context, sb *strings.Builder) {
	switch v.kind {
	case valInt:
		fmt.Fprintf(sb, "%d", v.n)
	case valFloat:
		sprintFloat(sb, v.F())
	case valVariadic:
		// v.n < len(ctx.variadicsNames)
		sb.WriteString(ctx.variadicsNames[v.n])
	case valLambda:
		// v.n < len(ctx.lambdas)
		sb.WriteString(ctx.lambdas[v.n].Source)
	case valBoxed:
		v.value.Sprint(ctx, sb)
	}
}

func (e *errV) Sprint(ctx *Context, sb *strings.Builder) {
	sb.WriteString("error[")
	e.V.Sprint(ctx, sb)
	sb.WriteByte(']')
}

func (e panicV) Sprint(ctx *Context, sb *strings.Builder) {
	sb.WriteString("panic[")
	sb.WriteString(strconv.Quote(string(e)))
	sb.WriteByte(']')
}

// Sprint returns a properly quoted string.
func (s S) Sprint(ctx *Context, sb *strings.Builder) { fmt.Fprintf(sb, "%q", string(s)) }

func (x *AB) Sprint(ctx *Context, sb *strings.Builder) {
	if x.Len() == 0 {
		sb.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", b2i(x.At(0)))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%d", b2i(xi))
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
}

func (x *AI) Sprint(ctx *Context, sb *strings.Builder) {
	if x.Len() == 0 {
		sb.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", x.At(0))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%d", xi)
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
}

func (x *AF) Sprint(ctx *Context, sb *strings.Builder) {
	if x.Len() == 0 {
		sb.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		sb.WriteRune(',')
		sprintFloat(sb, x.At(0))
		return
	}
	for i, xi := range x.Slice {
		sprintFloat(sb, xi)
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
}

func (x *AS) Sprint(ctx *Context, sb *strings.Builder) {
	if x.Len() == 0 {
		sb.WriteString(`0#""`)
		return
	}
	if x.Len() == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%q", x.At(0))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(sb, "%q", xi)
		if i < x.Len()-1 {
			sb.WriteRune(' ')
		}
	}
}

func (x *AV) Sprint(ctx *Context, sb *strings.Builder) {
	if x.Len() == 0 {
		sb.WriteString(`()`)
		return
	}
	if x.Len() == 1 {
		sb.WriteRune(',')
		x.At(0).Sprint(ctx, sb)
		return
	}
	sb.WriteRune('(')
	var sep string
	if ctx.sprintCompact {
		sep = ";"
	} else {
		sep = "\n "
	}
	osc := ctx.sprintCompact
	ctx.sprintCompact = true
	defer func() {
		ctx.sprintCompact = osc
	}()
	for i, xi := range x.Slice {
		if xi.kind != valNil {
			xi.Sprint(ctx, sb)
		}
		if i < x.Len()-1 {
			sb.WriteString(sep)
		}
	}
	sb.WriteRune(')')
}

func (p projection) Sprint(ctx *Context, sb *strings.Builder) {
	p.Fun.Sprint(ctx, sb)
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.kind != valNil {
			arg.Sprint(ctx, sb)
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
}

func (p projectionFirst) Sprint(ctx *Context, sb *strings.Builder) {
	p.Fun.Sprint(ctx, sb)
	sb.WriteByte('[')
	p.Arg.Sprint(ctx, sb)
	sb.WriteString(";]")
}

func (p projectionMonad) Sprint(ctx *Context, sb *strings.Builder) {
	p.Fun.Sprint(ctx, sb)
	sb.WriteString("[]")
}

func (r derivedVerb) Sprint(ctx *Context, sb *strings.Builder) {
	r.Arg.Sprint(ctx, sb)
	sb.WriteString(ctx.variadicsNames[r.Fun])
}
