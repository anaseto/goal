package goal

import (
	"math"
	"strconv"
	"unsafe"
)

func appendFloat(ctx *Context, dst []byte, f float64) []byte {
	switch {
	case math.IsInf(f, 0):
		if f >= 0 {
			return append(dst, "0w"...)
		}
		return append(dst, "-0w"...)
	case math.IsNaN(f):
		return append(dst, "0n"...)
	default:
		return strconv.AppendFloat(dst, f, 'g', ctx.Prec, 64)
	}
}

func appendInt(dst []byte, i int64) []byte {
	if i == math.MinInt64 {
		return append(dst, "0i"...)
	}
	return strconv.AppendInt(dst, i, 10)
}

// Sprint returns a matching program string representation of the value.
func (v V) Sprint(ctx *Context) string {
	// NOTE: optimize allocation away using unsafe. Caveat: Append should
	// never increase the number of references to the dst slice for such an
	// optimization to be correct. TODO: This code should be upgraded to
	// use unsafe.String at a later time.
	b := v.Append(ctx, nil)
	return *(*string)(unsafe.Pointer(&b))
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (v V) Append(ctx *Context, dst []byte) []byte {
	switch v.kind {
	case valInt:
		return appendInt(dst, v.I())
	case valFloat:
		return appendFloat(ctx, dst, v.F())
	case valVariadic:
		// v.n < len(ctx.variadicsNames)
		return append(dst, ctx.variadicsNames[v.n]...)
	case valLambda:
		// v.n < len(ctx.lambdas)
		return append(dst, ctx.lambdas[v.n].Source...)
	case valBoxed, valPanic:
		return v.value.Append(ctx, dst)
	default:
		// Could happen for nil values, but they are not normally
		// created from goal programs.
		return dst
	}
}

func (e *errV) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "error["...)
	dst = e.V.Append(ctx, dst)
	return append(dst, ']')
}

func (e panicV) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "panic["...)
	dst = strconv.AppendQuote(dst, string(e))
	return append(dst, ']')
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (s S) Append(ctx *Context, dst []byte) []byte {
	return strconv.AppendQuote(dst, string(s))
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (x *AB) Append(ctx *Context, dst []byte) []byte {
	if x.Len() == 0 {
		return append(dst, "!0"...)
	}
	if x.Len() == 1 {
		dst = append(dst, ',')
		dst = appendInt(dst, B2I(x.At(0)))
		return dst
	}
	for i, xi := range x.elts {
		dst = appendInt(dst, B2I(xi))
		if i < x.Len()-1 {
			dst = append(dst, ' ')
		}
	}
	return dst
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (x *AI) Append(ctx *Context, dst []byte) []byte {
	if x.Len() == 0 {
		return append(dst, "!0"...)
	}
	if x.Len() == 1 {
		dst = append(dst, ',')
		dst = appendInt(dst, x.At(0))
		return dst
	}
	for i, xi := range x.elts {
		dst = appendInt(dst, xi)
		if i < x.Len()-1 {
			dst = append(dst, ' ')
		}
	}
	return dst
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (x *AF) Append(ctx *Context, dst []byte) []byte {
	if x.Len() == 0 {
		return append(dst, "!0"...)
	}
	if x.Len() == 1 {
		dst = append(dst, ',')
		dst = appendFloat(ctx, dst, x.At(0))
		return dst
	}
	for i, xi := range x.elts {
		dst = appendFloat(ctx, dst, xi)
		if i < x.Len()-1 {
			dst = append(dst, ' ')
		}
	}
	return dst
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (x *AS) Append(ctx *Context, dst []byte) []byte {
	if x.Len() == 0 {
		return append(dst, `0#""`...)
	}
	if x.Len() == 1 {
		dst = append(dst, ',')
		dst = strconv.AppendQuote(dst, x.At(0))
		return dst
	}
	for i, xi := range x.elts {
		dst = strconv.AppendQuote(dst, xi)
		if i < x.Len()-1 {
			dst = append(dst, ' ')
		}
	}
	return dst
}

func needsParens(x []V) bool {
	for _, xi := range x {
		if xi.IsI() || xi.IsF() {
			continue
		}
		if _, ok := xi.value.(S); ok {
			continue
		}
		return true
	}
	return false
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (x *AV) Append(ctx *Context, dst []byte) []byte {
	if x.Len() == 0 {
		return append(dst, "()"...)
	}
	if x.Len() == 1 {
		dst = append(dst, ',')
		dst = x.At(0).Append(ctx, dst)
		return dst
	}
	sep := " "
	parens := needsParens(x.elts)
	if parens {
		dst = append(dst, '(')
		if ctx.compactFmt {
			sep = ";"
		} else {
			sep = "\n "
		}
	}
	osc := ctx.compactFmt
	ctx.compactFmt = true
	defer func() {
		ctx.compactFmt = osc
	}()
	for i, xi := range x.elts {
		if xi.kind != valNil {
			dst = xi.Append(ctx, dst)
		}
		if i < x.Len()-1 {
			dst = append(dst, sep...)
		}
	}
	if parens {
		dst = append(dst, ')')
	}
	return dst
}

// Append appends a unique program representation of the value to dst, and
// returns the extended buffer.
func (d *Dict) Append(ctx *Context, dst []byte) []byte {
	osc := ctx.compactFmt
	ctx.compactFmt = true
	defer func() {
		ctx.compactFmt = osc
	}()
	parens := arrayNeedsParens(d.keys)
	if parens {
		dst = append(dst, '(')
	}
	dst = d.keys.Append(ctx, dst)
	if parens {
		dst = append(dst, ')')
	}
	dst = append(dst, '!')
	dst = d.values.Append(ctx, dst)
	return dst
}

func arrayNeedsParens(x array) bool {
	switch x.Len() {
	case 0:
		switch x.(type) {
		case *AV:
			return false
		default:
			return true
		}
	case 1:
		return true
	default:
		return false
	}
}

func (p *projection) Append(ctx *Context, dst []byte) []byte {
	dst = p.Fun.Append(ctx, dst)
	dst = append(dst, '[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg.kind != valNil {
			dst = arg.Append(ctx, dst)
		}
		if i > 0 {
			dst = append(dst, ';')
		}
	}
	dst = append(dst, ']')
	return dst
}

func (p *projectionFirst) Append(ctx *Context, dst []byte) []byte {
	dst = p.Fun.Append(ctx, dst)
	dst = append(dst, '[')
	dst = p.Arg.Append(ctx, dst)
	dst = append(dst, ";]"...)
	return dst
}

func (p *projectionMonad) Append(ctx *Context, dst []byte) []byte {
	dst = p.Fun.Append(ctx, dst)
	dst = append(dst, "[]"...)
	return dst
}

func (r *derivedVerb) Append(ctx *Context, dst []byte) []byte {
	dst = r.Arg.Append(ctx, dst)
	dst = append(dst, ctx.variadicsNames[r.Fun]...)
	return dst
}
