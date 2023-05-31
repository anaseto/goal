package goal

import (
	"regexp"
	"strconv"
)

type rx struct {
	Regexp *regexp.Regexp
}

func (r *rx) Matches(x Value) bool {
	xv, ok := x.(*rx)
	return ok && r.Regexp.String() == xv.Regexp.String()
}

func (r *rx) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "rx["...)
	dst = strconv.AppendQuote(dst, r.Regexp.String())
	dst = append(dst, ']')
	return dst
}

func (r *rx) Type() string {
	return "r"
}

type rxReplacer struct {
	r    *rx
	repl V
}

func (r *rxReplacer) Matches(x Value) bool {
	xv, ok := x.(*rxReplacer)
	return ok && r.r.Matches(xv.r) && r.repl.Matches(xv.repl)
}

func (r *rxReplacer) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "sub["...)
	dst = r.r.Append(ctx, dst)
	dst = append(dst, ';')
	dst = r.repl.Append(ctx, dst)
	dst = append(dst, ']')
	return dst
}

func (r *rxReplacer) Type() string {
	return "f"
}

func (r *rxReplacer) stype() string {
	return "rx"
}

func (r *rxReplacer) rank(ctx *Context) int {
	return 1
}

func (r *rxReplacer) replace(ctx *Context, s string) string {
	switch zv := r.repl.bv.(type) {
	case S:
		return r.r.Regexp.ReplaceAllString(string(s), string(zv))
	default:
		// zv is a function
		f := func(s string) string {
			r := ctx.Apply(r.repl, NewS(s))
			switch rv := r.bv.(type) {
			case S:
				return string(rv)
			default:
				return r.Sprint(ctx)
			}
		}
		r.repl.IncrRC()
		rs := r.r.Regexp.ReplaceAllStringFunc(string(s), f)
		r.repl.DecrRC()
		return rs
	}
}

// vfRx implements the rx variadic.
func vfRx(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return compileRegex(args[0])
	default:
		return panicRank("rx")
	}
}

func compileRegex(x V) V {
	switch xv := x.bv.(type) {
	case S:
		r, err := regexp.Compile(string(xv))
		if err != nil {
			return Errorf("rx x : %v", err)
		}
		return NewV(&rx{Regexp: r})
	default:
		return panicType("rx x", "x", x)
	}
}

func applyRx(x *rx, y V) V {
	if x.Regexp.NumSubexp() > 0 {
		return applyRxFindSubmatch(x, y)
	}
	return applyRxMatch(x, y)
}

func applyRxMatch(x *rx, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewI(b2I(x.Regexp.MatchString(string(yv))))
	case *AS:
		r := make([]byte, yv.Len())
		for i, s := range yv.elts {
			r[i] = b2B(x.Regexp.MatchString(s))
		}
		return NewAB(r)
	case *AV:
		return mapAV(yv, func(yi V) V { return applyRxMatch(x, yi) })
	default:
		return panicType("r[y]", "y", y)
	}
}

func applyRxFindSubmatch(x *rx, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewAS(x.Regexp.FindStringSubmatch(string(yv)))
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := applyRxFindSubmatch(x, NewS(yi))
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *AV:
		return mapAV(yv, func(yi V) V { return applyRxFindSubmatch(x, yi) })
	default:
		return panicType("r[y]", "y", y)
	}
}

func applyRx2(x *rx, y, z V) V {
	var n int64
	if z.IsI() {
		n = z.I()
	} else if z.IsF() {
		if !isI(z.F()) {
			return Panicf("r[x;y] : non-integer y (%g)", z.F())
		}
		n = int64(z.F())
	} else {
		return panicType("r[x;y]", "y", z)
	}
	if x.Regexp.NumSubexp() > 0 {
		return applyRxFindAllSubmatch(x, y, n)
	}
	return applyRxFindAll(x, y, n)
}

func applyRxFindAllSubmatch(x *rx, y V, n int64) V {
	switch yv := y.bv.(type) {
	case S:
		matches := x.Regexp.FindAllStringSubmatch(string(yv), int(n))
		r := make([]V, len(matches))
		for i, sm := range matches {
			r[i] = NewV(&AS{elts: sm, flags: flagImmutable})
		}
		return newAVu(r)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := applyRxFindAllSubmatch(x, NewS(yi), n)
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *AV:
		return mapAV(yv, func(yi V) V { return applyRxFindAllSubmatch(x, yi, n) })
	default:
		return panicType("r[y]", "y", y)
	}
}

func applyRxFindAll(x *rx, y V, n int64) V {
	switch yv := y.bv.(type) {
	case S:
		matches := x.Regexp.FindAllString(string(yv), int(n))
		return NewAS(matches)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := applyRxFindAll(x, NewS(yi), n)
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *AV:
		return mapAV(yv, func(yi V) V { return applyRxFindAll(x, yi, n) })
	default:
		return panicType("r[y]", "y", y)
	}
}

func splitRx(f *rx, x V) V {
	switch xv := x.bv.(type) {
	case S:
		r := f.Regexp.Split(string(xv), -1)
		return NewAS(r)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := NewAS(f.Regexp.Split(string(xi), -1))
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *AV:
		return mapAV(xv, func(xi V) V { return splitRx(f, xi) })
	default:
		return panicType("r\\x", "x", x)
	}
}
