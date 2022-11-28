package goal

import "strings"

func fold2(ctx *Context, args []V) V {
	fv := args[1]
	switch fv := fv.(type) {
	case Variadic:
		switch fv {
		case vAdd:
			return fold2vAdd(args[0])
		}
	}
	f, ok := fv.(Function)
	if !ok {
		switch fv := fv.(type) {
		case S:
			return fold2Join(fv, args[0])
		case I, F, AB, AI, AF:
			return fold2Decode(fv, args[0])
		}
		return errType("F/x", "F", fv)
	}
	if f.Rank(ctx) != 2 {
		// TODO: converge
		return errf("F/x : F rank is %d (expected 2)", f.Rank(ctx))
	}
	switch x := args[0].(type) {
	case array:
		if x.Len() == 0 {
			f, ok := f.(zeroFun)
			if ok {
				return f.zero()
			}
			return I(0)
		}
		r := x.at(0)
		for i := 1; i < x.Len(); i++ {
			ctx.push(x.at(i))
			ctx.push(r)
			r = ctx.applyN(fv, 2)
		}
		return canonical(r)
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
		for _, xi := range x {
			n += xi
		}
		return I(n)
	case AF:
		n := 0.0
		for _, xi := range x {
			n += xi
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
		r := x[0]
		for _, xi := range x[1:] {
			r = add(r, xi)
		}
		return canonical(r)
	default:
		return x
	}
}

func fold2Join(sep S, x V) V {
	switch x := x.(type) {
	case S:
		return x
	case AS:
		return S(strings.Join([]string(x), string(sep)))
	case AV:
		assertCanonical(x)
		return errf("s/x : x not a string array (%s)", x.Type())
	default:
		return errf("s/x : x not a string array (%s)", x.Type())
	}
}

func fold2Decode(f V, x V) V {
	switch f := f.(type) {
	case I:
		switch x := x.(type) {
		case I:
			return x
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return I(x)
		case AI:
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += x[i] * n
				n *= int(f)
			}
			return I(r)
		case AB:
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += int(B2I(x[i])) * n
				n *= int(f)
			}
			return I(r)
		case AF:
			aix := toAI(x)
			if err, ok := aix.(errV); ok {
				return err
			}
			return fold2Decode(f, aix)
		default:
			return errType("I/x", "x", x)
		}
	case F:
		if !isI(f) {
			return errf("I/x : I non-integer (%g)", f)
		}
		return fold2Decode(I(f), x)
	case AB:
		return fold2Decode(fromABtoAI(f), x)
	case AI:
		switch x := x.(type) {
		case I:
			r := 0
			n := 1
			for i := len(f) - 1; i >= 0; i-- {
				r += int(x) * n
				n *= f[i]
			}
			return I(r)
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return fold2Decode(f, I(x))
		case AI:
			if len(f) != len(x) {
				return errf("I/x : length mismatch: %d (#I) %d (#x)", len(f), len(x))
			}
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += x[i] * n
				n *= f[i]
			}
			return I(r)
		case AB:
			return fold2Decode(f, fromABtoAI(x))
		case AF:
			aix := toAI(x)
			if err, ok := aix.(errV); ok {
				return err
			}
			return fold2Decode(f, aix)
		default:
			return errType("I/x", "x", x)
		}
	case AF:
		aif := toAI(f)
		if err, ok := aif.(errV); ok {
			return err
		}
		return fold2Decode(aif, x)
	default:
		// should not happen
		return errType("I/x", "I", f)
	}
}

func fold3(ctx *Context, args []V) V {
	f, ok := args[1].(Function)
	if !ok {
		return errf("x F/y : F not a function (%s)", args[1].Type())
	}
	if f.Rank(ctx) != 2 {
		return fold3While(ctx, args)
	}
	y := args[0]
	switch y := y.(type) {
	case array:
		r := args[2]
		if y.Len() == 0 {
			return r
		}
		for i := 0; i < y.Len(); i++ {
			ctx.push(y.at(i))
			ctx.push(r)
			r = ctx.applyN(f, 2)
			if err, ok := r.(errV); ok {
				return err
			}
		}
		return canonical(r)
	default:
		ctx.push(y)
		ctx.push(args[2])
		return ctx.applyN(f, 2)
	}
}

func fold3While(ctx *Context, args []V) V {
	f := args[1]
	switch x := args[2].(type) {
	case F:
		if !isI(x) {
			return errf("n f/y : non-integer n (%g)", x)
		}
		return fold3doTimes(ctx, int(x), f, args[0])
	case I:
		return fold3doTimes(ctx, int(x), f, args[0])
	case Function:
		y := args[0]
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if err, ok := cond.(errV); ok {
				return err
			}
			if !isTrue(cond) {
				return y
			}
			ctx.push(y)
			y = ctx.applyN(f, 1)
			if err, ok := y.(errV); ok {
				return err
			}
		}
	default:
		return errType("x f/y", "x", x)
	}
}

func fold3doTimes(ctx *Context, n int, f, y V) V {
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(f, 1)
		if err, ok := y.(errV); ok {
			return err
		}
	}
	return y
}

func scan2(ctx *Context, fv, x V) V {
	f, ok := fv.(Function)
	if !ok {
		switch fv := fv.(type) {
		case S:
			return scan2Split(fv, x)
		}
		// TODO: encode
		return errf("f\\x : f not a function (%s)", fv.Type())
	}
	if f.Rank(ctx) != 2 {
		// TODO: converge
		return errf("f\\x : f rank is %d (expected 2)", f.Rank(ctx))
	}
	switch x := x.(type) {
	case array:
		if x.Len() == 0 {
			f, ok := fv.(zeroFun)
			if ok {
				return f.zero()
			}
			return I(0)
		}
		r := AV{x.at(0)}
		for i := 1; i < x.Len(); i++ {
			ctx.push(x.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(fv, 2)
			if err, ok := next.(errV); ok {
				return err
			}
			r = append(r, next)
		}
		return canonical(r)
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
		assertCanonical(x)
		return errf("s/x : x not a string atom or array (%s)", x.Type())
	default:
		return errf("s/x : x not a string atom or array (%s)", x.Type())
	}
}

func scan3(ctx *Context, args []V) V {
	f, ok := args[1].(Function)
	if !ok {
		return errf("x f'y : f not a function (%s)", args[1].Type())
	}
	if f.Rank(ctx) != 2 {
		return scan3While(ctx, args)
	}
	y := args[0]
	switch y := y.(type) {
	case array:
		if y.Len() == 0 {
			return AV{}
		}
		ctx.push(y.at(0))
		ctx.push(args[2])
		first := ctx.applyN(f, 2)
		if err, ok := first.(errV); ok {
			return err
		}
		r := AV{first}
		for i := 1; i < y.Len(); i++ {
			ctx.push(y.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(f, 2)
			if err, ok := next.(errV); ok {
				return err
			}
			r = append(r, next)
		}
		return canonical(r)
	default:
		ctx.push(y)
		ctx.push(args[2])
		return ctx.applyN(f, 2)
	}
}

func scan3While(ctx *Context, args []V) V {
	f := args[1]
	switch x := args[2].(type) {
	case F:
		if !isI(x) {
			return errf("n f\\y : non-integer n (%g)", x)
		}
		return scan3doTimes(ctx, int(x), f, args[0])
	case I:
		return scan3doTimes(ctx, int(x), f, args[0])
	case Function:
		y := args[0]
		r := AV{y}
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if err, ok := cond.(errV); ok {
				return err
			}
			if !isTrue(cond) {
				return canonical(r)
			}
			ctx.push(y)
			y = ctx.applyN(f, 1)
			if err, ok := y.(errV); ok {
				return err
			}
			r = append(r, y)
		}
	default:
		return errType("x f\\y", "x", x)
	}
}

func scan3doTimes(ctx *Context, n int, f, y V) V {
	r := AV{y}
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(f, 1)
		if err, ok := y.(errV); ok {
			return err
		}
		r = append(r, y)
	}
	return canonical(r)
}

func each2(ctx *Context, args []V) V {
	f, ok := args[1].(Function)
	if !ok {
		return errf("f'x : f not a function (%s)", args[1].Type())
	}
	x := toArray(args[0])
	switch x := x.(type) {
	case array:
		r := make(AV, 0, x.Len())
		for i := 0; i < x.Len(); i++ {
			ctx.push(x.at(i))
			next := ctx.applyN(f, 1)
			if err, ok := next.(errV); ok {
				return err
			}
			r = append(r, next)
		}
		return canonical(r)
	default:
		// should not happen
		return errf("f'x : x not an array (%s)", x.Type())
	}
}

func each3(ctx *Context, args []V) V {
	f, ok := args[1].(Function)
	if !ok {
		return errf("x f'y : f not a function (%s)", args[1].Type())
	}
	x, okax := args[2].(array)
	y, okay := args[0].(array)
	if !okax && !okay {
		return ctx.ApplyN(f, args)
	}
	if !okax {
		ylen := y.Len()
		r := make(AV, 0, ylen)
		for i := 0; i < ylen; i++ {
			ctx.push(y.at(i))
			ctx.push(args[2])
			next := ctx.applyN(f, 2)
			if err, ok := next.(errV); ok {
				return err
			}
			r = append(r, next)
		}
		return canonical(r)
	}
	if !okay {
		xlen := x.Len()
		r := make(AV, 0, xlen)
		for i := 0; i < xlen; i++ {
			ctx.push(args[0])
			ctx.push(x.at(i))
			next := ctx.applyN(f, 2)
			if err, ok := next.(errV); ok {
				return err
			}
			r = append(r, next)
		}
		return canonical(r)
	}
	xlen := x.Len()
	if xlen != y.Len() {
		return errf("x f'y : length mismatch: %d (#x) vs %d (#y)", x.Len(), y.Len())
	}
	r := make(AV, 0, xlen)
	for i := 0; i < xlen; i++ {
		ctx.push(y.at(i))
		ctx.push(x.at(i))
		next := ctx.applyN(f, 2)
		if err, ok := next.(errV); ok {
			return err
		}
		r = append(r, next)
	}
	return canonical(r)
}
