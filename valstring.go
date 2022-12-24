package goal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func sprintFloat(w ValueWriter, f float64) {
	switch {
	case math.IsInf(f, 0):
		if f >= 0 {
			w.WriteString("0w")
		} else {
			w.WriteString("-0w")
		}
	case math.IsNaN(f):
		w.WriteString("0n")
	default:
		fmt.Fprintf(w, "%g", f)
	}
}

// Format returns a matching program string representation of the value.
func (v V) Format(ctx *Context) string {
	var sb strings.Builder
	v.Sprint(ctx, &sb)
	return sb.String()
}

// Sprint writes a matching program string representation of the value.
func (v V) Sprint(ctx *Context, w ValueWriter) {
	switch v.kind {
	case valInt:
		fmt.Fprintf(w, "%d", v.n)
	case valFloat:
		sprintFloat(w, v.F())
	case valVariadic:
		// v.n < len(ctx.variadicsNames)
		w.WriteString(ctx.variadicsNames[v.n])
	case valLambda:
		// v.n < len(ctx.lambdas)
		w.WriteString(ctx.lambdas[v.n].Source)
	case valBoxed:
		v.value.Sprint(ctx, w)
	}
}

func (e *errV) Sprint(ctx *Context, w ValueWriter) {
	w.WriteString("error[")
	e.V.Sprint(ctx, w)
	w.WriteByte(']')
}

func (e panicV) Sprint(ctx *Context, w ValueWriter) {
	w.WriteString("panic[")
	w.WriteString(strconv.Quote(string(e)))
	w.WriteByte(']')
}

// Sprint returns a properly quoted string.
func (s S) Sprint(ctx *Context, w ValueWriter) { fmt.Fprintf(w, "%q", string(s)) }

func (x *AB) Sprint(ctx *Context, w ValueWriter) {
	if x.Len() == 0 {
		w.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		w.WriteRune(',')
		fmt.Fprintf(w, "%d", b2i(x.At(0)))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(w, "%d", b2i(xi))
		if i < x.Len()-1 {
			w.WriteRune(' ')
		}
	}
}

func (x *AI) Sprint(ctx *Context, w ValueWriter) {
	if x.Len() == 0 {
		w.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		w.WriteRune(',')
		fmt.Fprintf(w, "%d", x.At(0))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(w, "%d", xi)
		if i < x.Len()-1 {
			w.WriteRune(' ')
		}
	}
}

func (x *AF) Sprint(ctx *Context, w ValueWriter) {
	if x.Len() == 0 {
		w.WriteString(`!0`)
		return
	}
	if x.Len() == 1 {
		w.WriteRune(',')
		sprintFloat(w, x.At(0))
		return
	}
	for i, xi := range x.Slice {
		sprintFloat(w, xi)
		if i < x.Len()-1 {
			w.WriteRune(' ')
		}
	}
}

func (x *AS) Sprint(ctx *Context, w ValueWriter) {
	if x.Len() == 0 {
		w.WriteString(`0#""`)
		return
	}
	if x.Len() == 1 {
		w.WriteRune(',')
		fmt.Fprintf(w, "%q", x.At(0))
		return
	}
	for i, xi := range x.Slice {
		fmt.Fprintf(w, "%q", xi)
		if i < x.Len()-1 {
			w.WriteRune(' ')
		}
	}
}

func (x *AV) Sprint(ctx *Context, w ValueWriter) {
	if x.Len() == 0 {
		w.WriteString(`()`)
		return
	}
	if x.Len() == 1 {
		w.WriteRune(',')
		x.At(0).Sprint(ctx, w)
		return
	}
	w.WriteRune('(')
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
			xi.Sprint(ctx, w)
		}
		if i < x.Len()-1 {
			w.WriteString(sep)
		}
	}
	w.WriteRune(')')
}

func (p projection) Sprint(ctx *Context, w ValueWriter) {
	p.Fun.Sprint(ctx, w)
	w.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.kind != valNil {
			arg.Sprint(ctx, w)
		}
		if i > 0 {
			w.WriteRune(';')
		}
	}
	w.WriteRune(']')
}

func (p projectionFirst) Sprint(ctx *Context, w ValueWriter) {
	p.Fun.Sprint(ctx, w)
	w.WriteByte('[')
	p.Arg.Sprint(ctx, w)
	w.WriteString(";]")
}

func (p projectionMonad) Sprint(ctx *Context, w ValueWriter) {
	p.Fun.Sprint(ctx, w)
	w.WriteString("[]")
}

func (r derivedVerb) Sprint(ctx *Context, w ValueWriter) {
	r.Arg.Sprint(ctx, w)
	w.WriteString(ctx.variadicsNames[r.Fun])
}
