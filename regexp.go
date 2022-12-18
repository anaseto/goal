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
			r.Slice[i] = replaceRx(ctx, xi, y, z)
			if r.Slice[i].IsPanic() {
				return r.Slice[i]
			}
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
