package goal

import "strings"

func fold2(ctx *Context, v, x V) V {
	switch v := v.(type) {
	case Variadic:
		switch v {
		case vAdd:
			return fold2vAdd(x)
		}
	}
	vv, ok := v.(Function)
	if !ok {
		switch v := v.(type) {
		case S:
			return fold2Join(v, x)
		}
		// TODO: join, split, encode, decode
		return errf("not a function left: %s", v.Type())
	}
	if vv.Rank(ctx) != 2 {
		// TODO: converge
		return errf("rank %d verb (expected 2)", vv.Rank(ctx))
	}
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			vv, ok := vv.(zeroFun)
			if ok {
				return vv.zero()
			}
			return I(0)
		}
		res := x.At(0)
		for i := 1; i < x.Len(); i++ {
			ctx.push(x.At(i))
			ctx.push(res)
			res = ctx.applyN(v, 2)
		}
		return canonical(res)
	default:
		return x
	}
}

func fold2vAdd(x V) V {
	switch x := x.(type) {
	case AB:
		n := I(0)
		for _, b := range x {
			if b {
				n++
			}
		}
		return n
	case AI:
		n := 0
		for _, v := range x {
			n += v
		}
		return I(n)
	case AF:
		n := 0.0
		for _, v := range x {
			n += v
		}
		return F(n)
	case AS:
		if len(x) == 0 {
			return S("")
		}
		n := 0
		for _, s := range x {
			n += len(s)
		}
		var b strings.Builder
		b.Grow(n)
		for _, s := range x {
			b.WriteString(s)
		}
		return S(b.String())
	case AV:
		if len(x) == 0 {
			return I(0)
		}
		res := x[0]
		for _, v := range x[1:] {
			res = add(res, v)
		}
		return canonical(res)
	default:
		return x
	}
}

func fold2Join(sep S, x V) V {
	switch x := x.(type) {
	case AS:
		return S(strings.Join([]string(x), string(sep)))
	case AV:
		xx := canonical(x)
		if xx, ok := xx.(AS); ok {
			return S(strings.Join([]string(xx), string(sep)))
		}
		return errs("not a string array")
	default:
		return errs("not a string array")
	}
}

func scan2(ctx *Context, v, x V) V {
	vv, ok := v.(Function)
	if !ok {
		switch v := v.(type) {
		case S:
			return scan2Split(v, x)
		}
		// TODO: split, encode, decode
		return errsw("not a function")
	}
	if vv.Rank(ctx) != 2 {
		// TODO: converge
		return errf("rank %d verb (expected 2)", vv.Rank(ctx))
	}
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			vv, ok := v.(zeroFun)
			if ok {
				return vv.zero()
			}
			return I(0)
		}
		res := AV{x.At(0)}
		for i := 1; i < x.Len(); i++ {
			ctx.push(x.At(i))
			ctx.push(res[len(res)-1])
			next := ctx.applyN(v, 2)
			if err, ok := next.(E); ok {
				return err
			}
			res = append(res, next)
		}
		return canonical(res)
	default:
		return x
	}
}

func scan2Split(sep S, x V) V {
	switch x := x.(type) {
	case S:
		return AS(strings.Split(string(x), string(sep)))
	case AS:
		r := make(AV, len(x))
		for i := range r {
			r[i] = AS(strings.Split(x[i], string(sep)))
		}
		return r
	case AV:
		xx := canonical(x)
		switch xx := xx.(type) {
		case S:
			return scan2Split(sep, xx)
		case AS:
			return scan2Split(sep, xx)
		default:
			return errs("not a string atom or array")
		}
	default:
		return errs("not a string atom or array")
	}
}
