package goal

import (
	"regexp"
)

type rx struct {
	Regexp *regexp.Regexp
}

func (r *rx) Matches(x Value) bool {
	xv, ok := x.(*rx)
	return ok && r.Regexp.String() == xv.Regexp.String()
}

func (r *rx) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	n, err = w.WriteString("rx[")
	if err != nil {
		return
	}
	m, err = S(r.Regexp.String()).Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(']')
	if err != nil {
		return
	}
	n++
	return
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
	return ok && r.r.Matches(xv.r) && Match(r.repl, xv.repl)
}

func (r *rxReplacer) Fprint(ctx *Context, w ValueWriter) (n int, err error) {
	var m int
	n, err = w.WriteString("sub[")
	if err != nil {
		return
	}
	m, err = r.r.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(';')
	if err != nil {
		return
	}
	n++
	m, err = r.repl.Fprint(ctx, w)
	n += m
	if err != nil {
		return
	}
	err = w.WriteByte(']')
	if err != nil {
		return
	}
	n++
	return
}

func (r *rxReplacer) Type() string {
	return "f"
}

func (r *rxReplacer) rank(ctx *Context) int {
	return 1
}

func (r *rxReplacer) replace(ctx *Context, s string) string {
	switch zv := r.repl.value.(type) {
	case S:
		return r.r.Regexp.ReplaceAllString(string(s), string(zv))
	default:
		// zv is a function
		f := func(s string) string {
			r := ctx.Apply(r.repl, NewS(s))
			switch rv := r.value.(type) {
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

// VRx implements the rx variadic.
func VRx(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return compileRegex(args[0])
	default:
		return panicRank("rx")
	}
}

func compileRegex(x V) V {
	switch xv := x.value.(type) {
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
	switch yv := y.value.(type) {
	case S:
		return NewI(b2i(x.Regexp.MatchString(string(yv))))
	case *AS:
		r := make([]bool, yv.Len())
		for i, s := range yv.Slice {
			r[i] = x.Regexp.MatchString(s)
		}
		return NewABRC(r, yv.rc)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.Slice {
			ri := applyRxMatch(x, yi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("r[y]", "y", y)
	}
}

func applyRxFindSubmatch(x *rx, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewAS(x.Regexp.FindStringSubmatch(string(yv)))
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyRxFindSubmatch(x, NewS(yi))
		}
		return NewAVRC(r, yv.rc)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.Slice {
			ri := applyRxFindSubmatch(x, yi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
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
			return Panicf("r[x;y] : y not an integer (%g)", z.F())
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
	switch yv := y.value.(type) {
	case S:
		matches := x.Regexp.FindAllStringSubmatch(string(yv), int(n))
		r := make([]V, len(matches))
		for i, sm := range matches {
			r[i] = NewAS(sm)
		}
		return NewAV(r)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyRxFindAllSubmatch(x, NewS(yi), n)
		}
		return NewAVRC(r, yv.rc)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.Slice {
			ri := applyRxFindAllSubmatch(x, yi, n)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("r[y]", "y", y)
	}
}

func applyRxFindAll(x *rx, y V, n int64) V {
	switch yv := y.value.(type) {
	case S:
		matches := x.Regexp.FindAllString(string(yv), int(n))
		return NewAS(matches)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyRxFindAll(x, NewS(yi), n)
		}
		return NewAVRC(r, yv.rc)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.Slice {
			ri := applyRxFindAll(x, yi, n)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("r[y]", "y", y)
	}
}

func splitRx(f *rx, x V) V {
	switch xv := x.value.(type) {
	case S:
		r := f.Regexp.Split(string(xv), -1)
		return NewAS(r)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = NewAS(f.Regexp.Split(string(xi), -1))
		}
		return NewAVRC(r, xv.rc)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := splitRx(f, xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("r\\x", "x", x)
	}
}
