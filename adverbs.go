package main

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
		// TODO: join, split, encode, decode
		return errsw("not a function")
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
			res = Add(res, v)
		}
		return canonical(res)
	default:
		return x
	}
}
