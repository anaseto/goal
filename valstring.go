package goal

import (
	"fmt"
	"math"
	"strings"
)

func sprintFloat(w ValueWriter, f float64) (n int, err error) {
	switch {
	case math.IsInf(f, 0):
		if f >= 0 {
			return w.WriteString("0w")
		}
		return w.WriteString("-0w")
	case math.IsNaN(f):
		return w.WriteString("0n")
	default:
		return fmt.Fprintf(w, "%g", f)
	}
}

// Sprint returns a matching program string representation of the value.
func (v V) Sprint(ctx *Context) string {
	var sb strings.Builder
	v.Fprint(ctx, &sb)
	return sb.String()
}

// Fprint writes a matching program string representation of the value.
func (v V) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	switch v.kind {
	case valInt:
		return fmt.Fprintf(w, "%d", v.n)
	case valFloat:
		return sprintFloat(w, v.F())
	case valVariadic:
		// v.n < len(ctx.variadicsNames)
		return w.WriteString(ctx.variadicsNames[v.n])
	case valLambda:
		// v.n < len(ctx.lambdas)
		return w.WriteString(ctx.lambdas[v.n].Source)
	case valBoxed, valPanic:
		return v.value.Fprint(ctx, w)
	default:
		return 0, nil
	}
}

func (e *errV) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	m, err = w.WriteString("error[")
	n += m
	if err != nil {
		return
	}
	m, err = e.V.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(']')
	if err == nil {
		n++
	}
	return
}

func (e panicV) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	return fmt.Fprintf(w, "panic[%q]", string(e))
}

// Fprint writes a properly quoted string.
func (s S) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	return fmt.Fprintf(w, "%q", string(s))
}

func (x *AB) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	if x.Len() == 0 {
		return w.WriteString(`!0`)
	}
	var m int
	if x.Len() == 1 {
		err = w.WriteByte(',')
		if err != nil {
			return
		}
		n++
		m, err = fmt.Fprintf(w, "%d", b2i(x.At(0)))
		n += m
		return
	}
	for i, xi := range x.Slice {
		m, err = fmt.Fprintf(w, "%d", b2i(xi))
		n += m
		if err != nil {
			return
		}
		if i < x.Len()-1 {
			err = w.WriteByte(' ')
			if err != nil {
				return
			}
			n++
		}
	}
	return
}

func (x *AI) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	if x.Len() == 0 {
		return w.WriteString(`!0`)
	}
	var m int
	if x.Len() == 1 {
		err = w.WriteByte(',')
		if err != nil {
			return
		}
		n++
		m, err = fmt.Fprintf(w, "%d", x.At(0))
		n += m
		return
	}
	for i, xi := range x.Slice {
		m, err = fmt.Fprintf(w, "%d", xi)
		n += m
		if err != nil {
			return
		}
		if i < x.Len()-1 {
			err = w.WriteByte(' ')
			if err != nil {
				return
			}
			n++
		}
	}
	return
}

func (x *AF) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	if x.Len() == 0 {
		return w.WriteString(`!0`)
	}
	var m int
	if x.Len() == 1 {
		err = w.WriteByte(',')
		if err != nil {
			return
		}
		n++
		m, err = sprintFloat(w, x.At(0))
		n += m
		return
	}
	for i, xi := range x.Slice {
		m, err = sprintFloat(w, xi)
		n += m
		if err != nil {
			return
		}
		if i < x.Len()-1 {
			err = w.WriteByte(' ')
			if err != nil {
				return
			}
			n++
		}
	}
	return
}

func (x *AS) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	if x.Len() == 0 {
		return w.WriteString(`0#""`)
	}
	var m int
	if x.Len() == 1 {
		err = w.WriteByte(',')
		if err != nil {
			return
		}
		n++
		m, err = fmt.Fprintf(w, "%q", x.At(0))
		n += m
		return
	}
	for i, xi := range x.Slice {
		m, err = fmt.Fprintf(w, "%q", xi)
		n += m
		if err != nil {
			return
		}
		if i < x.Len()-1 {
			err = w.WriteByte(' ')
			if err != nil {
				return
			}
			n++
		}
	}
	return
}

func (x *AV) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	if x.Len() == 0 {
		return w.WriteString(`()`)
	}
	var m int
	if x.Len() == 1 {
		err = w.WriteByte(',')
		if err != nil {
			return
		}
		n++
		m, err = x.At(0).Fprint(ctx, w)
		n += m
		return
	}
	err = w.WriteByte('(')
	if err != nil {
		return
	}
	n++
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
			m, err = xi.Fprint(ctx, w)
			n += m
			if err != nil {
				return
			}
		}
		if i < x.Len()-1 {
			m, err = w.WriteString(sep)
			n += m
			if err != nil {
				return
			}
		}
	}
	err = w.WriteByte(')')
	if err != nil {
		return
	}
	n++
	return
}

func (p projection) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	m, err = p.Fun.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte('[')
	if err != nil {
		return
	}
	n++
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.kind != valNil {
			m, err = arg.Fprint(ctx, w)
			n += m
			if err != nil {
				return
			}
		}
		if i > 0 {
			err = w.WriteByte(';')
			if err != nil {
				return
			}
			n++
		}
	}
	err = w.WriteByte(']')
	if err != nil {
		return
	}
	n++
	return
}

func (p projectionFirst) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	m, err = p.Fun.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte('[')
	if err != nil {
		return
	}
	n++
	m, err = p.Arg.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	m, err = w.WriteString(";]")
	n += m
	return
}

func (p projectionMonad) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	n, err = p.Fun.Fprint(ctx, w)
	if err != nil {
		return
	}
	m, err = w.WriteString("[]")
	n += m
	return
}

func (r derivedVerb) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	n, err = r.Arg.Fprint(ctx, w)
	if err != nil {
		return
	}
	m, err = w.WriteString(ctx.variadicsNames[r.Fun])
	n += m
	return
}
