package goal

import "strings"

func fold2(ctx *Context, args []V) V {
	v := args[1]
	switch v := v.(type) {
	case Variadic:
		switch v {
		case vAdd:
			return fold2vAdd(args[0])
		}
	}
	vv, ok := v.(Function)
	if !ok {
		switch v := v.(type) {
		case S:
			return fold2Join(v, args[0])
		}
		// TODO: decode
		return errf("F/x : F not a function (%s)", v.Type())
	}
	if vv.Rank(ctx) != 2 {
		// TODO: converge
		return errf("F/x : F rank is %d (expected 2)", vv.Rank(ctx))
	}
	switch x := args[0].(type) {
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
		return errf("s/x : x not a string array (%s)", x.Type())
	default:
		return errf("s/x : x not a string array (%s)", x.Type())
	}
}

func fold3(ctx *Context, args []V) V {
	v, ok := args[1].(Function)
	if !ok {
		return errf("x F/y : F not a function (%s)", v.Type())
	}
	if v.Rank(ctx) != 2 {
		return fold3While(ctx, args)
	}
	y := args[0]
	switch y := y.(type) {
	case Array:
		res := args[2]
		if y.Len() == 0 {
			return res
		}
		for i := 0; i < y.Len(); i++ {
			ctx.push(y.At(i))
			ctx.push(res)
			res = ctx.applyN(v, 2)
			if err, ok := res.(E); ok {
				return err
			}
		}
		return canonical(res)
	default:
		ctx.push(y)
		ctx.push(args[2])
		return ctx.applyN(v, 2)
	}
}

func fold3While(ctx *Context, args []V) V {
	v := args[1]
	switch x := args[2].(type) {
	case F:
		if !isI(x) {
			return errf("n f/y : non-integer n (%g)", x)
		}
		return fold3doTimes(ctx, int(x), v, args[0])
	case I:
		return fold3doTimes(ctx, int(x), v, args[0])
	case Function:
		y := args[0]
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if err, ok := cond.(E); ok {
				return err
			}
			if !isTrue(cond) {
				return y
			}
			ctx.push(y)
			y = ctx.applyN(v, 1)
			if err, ok := y.(E); ok {
				return err
			}
		}
	default:
		return errf("x f/y : bad type `%v for x", x.Type())
	}
}

func fold3doTimes(ctx *Context, n int, v, y V) V {
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(v, 1)
		if err, ok := y.(E); ok {
			return err
		}
	}
	return y
}

func scan2(ctx *Context, v, x V) V {
	vv, ok := v.(Function)
	if !ok {
		switch v := v.(type) {
		case S:
			return scan2Split(v, x)
		}
		// TODO: encode
		return errf("f\\x : f not a function (%s)", v.Type())
	}
	if vv.Rank(ctx) != 2 {
		// TODO: converge
		return errf("f\\x : f rank is %d (expected 2)", vv.Rank(ctx))
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
			return errf("s/x : x not a string atom or array (%s)", x.Type())
		}
	default:
		return errf("s/x : x not a string atom or array (%s)", x.Type())
	}
}

func scan3(ctx *Context, args []V) V {
	v, ok := args[1].(Function)
	if !ok {
		return errf("x f'y : f not a function (%s)", v.Type())
	}
	if v.Rank(ctx) != 2 {
		return scan3While(ctx, args)
		//return errf("rank %d verb (expected 2)", v.Rank(ctx))
	}
	y := args[0]
	switch y := y.(type) {
	case Array:
		if y.Len() == 0 {
			return AV{}
		}
		ctx.push(y.At(0))
		ctx.push(args[2])
		first := ctx.applyN(v, 2)
		if err, ok := first.(E); ok {
			return err
		}
		res := AV{first}
		for i := 1; i < y.Len(); i++ {
			ctx.push(y.At(i))
			ctx.push(res[len(res)-1])
			next := ctx.applyN(v, 2)
			if err, ok := next.(E); ok {
				return err
			}
			res = append(res, next)
		}
		return canonical(res)
	default:
		ctx.push(y)
		ctx.push(args[2])
		return ctx.applyN(v, 2)
	}
}

func scan3While(ctx *Context, args []V) V {
	v := args[1]
	switch x := args[2].(type) {
	case F:
		if !isI(x) {
			return errf("n f\\y : non-integer n (%g)", x)
		}
		return scan3doTimes(ctx, int(x), v, args[0])
	case I:
		return scan3doTimes(ctx, int(x), v, args[0])
	case Function:
		y := args[0]
		res := AV{y}
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if err, ok := cond.(E); ok {
				return err
			}
			if !isTrue(cond) {
				return canonical(res)
			}
			ctx.push(y)
			y = ctx.applyN(v, 1)
			if err, ok := y.(E); ok {
				return err
			}
			res = append(res, y)
		}
	default:
		return errf("x f\\y : bad type `%v for x", x.Type())
	}
}

func scan3doTimes(ctx *Context, n int, v, y V) V {
	res := AV{y}
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(v, 1)
		if err, ok := y.(E); ok {
			return err
		}
		res = append(res, y)
	}
	return canonical(res)
}

func each2(ctx *Context, args []V) V {
	v, ok := args[1].(Function)
	if !ok {
		// TODO: binary search
		return errf("f'x : f not a function (%s)", v.Type())
	}
	x := toArray(args[0])
	switch x := x.(type) {
	case Array:
		res := make(AV, 0, x.Len())
		for i := 0; i < x.Len(); i++ {
			ctx.push(x.At(i))
			next := ctx.applyN(v, 1)
			if err, ok := next.(E); ok {
				return err
			}
			res = append(res, next)
		}
		return canonical(res)
	default:
		// should not happen
		return errf("f'x : x not an array (%s)", x.Type())
	}
}

func each3(ctx *Context, args []V) V {
	v, ok := args[1].(Function)
	if !ok {
		return errf("x f'y : f not a function (%s)", v.Type())
	}
	x, ok := args[2].(Array)
	if !ok {
		return errf("x f'y : x not an array (%s)", x.Type())
	}
	y, ok := args[0].(Array)
	if !ok {
		return errf("x f'y : y not an array (%s)", y.Type())
	}
	xlen := x.Len()
	if xlen != y.Len() {
		return errf("x f'y : length mismatch: %d (#x) vs %d (#y)", x.Len(), y.Len())
	}
	res := make(AV, 0, xlen)
	for i := 0; i < xlen; i++ {
		ctx.push(y.At(i))
		ctx.push(x.At(i))
		next := ctx.applyN(v, 2)
		if err, ok := next.(E); ok {
			return err
		}
		res = append(res, next)
	}
	return canonical(res)
}
