package goal

import "regexp"

type rx struct {
	Regexp *regexp.Regexp
}

func (r *rx) Matches(x Value) bool {
	xv, ok := x.(*rx)
	return ok && r.Regexp.String() == xv.Regexp.String()
}

func (r *rx) Sprint(ctx *Context) string {
	return "rx[" + S(r.Regexp.String()).Sprint(ctx) + "]"
}

func (r *rx) Type() string {
	return "r"
}

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

func replaceRx(ctx *Context, x V, y *rx, z V) V {
	switch xv := x.value.(type) {
	case S:
		return replaceSRx(ctx, xv, y, z)
	case *AS:
		return replaceASRx(ctx, xv, y, z)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := replaceRx(ctx, xi, y, z)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("sub[x;y;z]", "x", x)
	}
}

func replaceSRx(ctx *Context, s S, y *rx, z V) V {
	switch zv := z.value.(type) {
	case S:
		return NewS(y.Regexp.ReplaceAllString(string(s), string(zv)))
	default:
		if z.IsFunction() {
			f := func(s string) string {
				r := ctx.Apply(z, NewS(s))
				switch rv := r.value.(type) {
				case S:
					return string(rv)
				default:
					return r.Sprint(ctx)
				}
			}
			return NewS(y.Regexp.ReplaceAllStringFunc(string(s), f))
		}
		return panicType("sub[s;r;repl]", "repl", z)
	}
}

func replaceASRx(ctx *Context, x *AS, y *rx, z V) V {
	switch zv := z.value.(type) {
	case S:
		x := x.reuse()
		for i, s := range x.Slice {
			x.Slice[i] = y.Regexp.ReplaceAllString(string(s), string(zv))
		}
		return NewV(x)
	default:
		if z.IsFunction() {
			f := func(s string) string {
				r := ctx.Apply(z, NewS(s))
				switch rv := r.value.(type) {
				case S:
					return string(rv)
				default:
					return r.Sprint(ctx)
				}
			}
			x := x.reuse()
			for i, s := range x.Slice {
				x.Slice[i] = y.Regexp.ReplaceAllStringFunc(string(s), f)
			}
			return NewV(x)
		}
		return panicType("sub[s;r;repl]", "repl", z)
	}
}

func applyRx(x *rx, y V) V {
	if x.Regexp.NumSubexp() > 1 {
		return applyRxFindSubmatch(x, y)
	}
	if x.Regexp.NumSubexp() == 1 {
		return applyRxFind(x, y)
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
		return NewAB(r)
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

func applyRxFind(x *rx, y V) V {
	switch yv := y.value.(type) {
	case S:
		r := []string{}
		matches := x.Regexp.FindAllStringSubmatch(string(yv), -1)
		for _, sm := range matches {
			if len(sm) > 1 {
				r = append(r, sm[1])
			}
		}
		return NewAS(r)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyRxFind(x, NewS(yi))
		}
		return NewAV(r)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.Slice {
			ri := applyRxFind(x, yi)
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
		matches := x.Regexp.FindAllStringSubmatch(string(yv), -1)
		r := make([]V, len(matches))
		for i, sm := range matches {
			r[i] = NewAS(sm[1:])
		}
		return NewAV(r)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyRxFindSubmatch(x, NewS(yi))
		}
		return NewAV(r)
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

func scan2SplitRx(f *rx, x V) V {
	switch xv := x.value.(type) {
	case S:
		r := f.Regexp.Split(string(xv), -1)
		return NewAS(r)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = NewAS(f.Regexp.Split(string(xi), -1))
		}
		return NewAV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := scan2SplitRx(f, xi)
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
