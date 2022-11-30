package goal

import "strings"

func fold2(ctx *Context, args []V) V {
	f := args[1]
	switch f.Kind {
	case IntVariadic:
		switch f.Variadic() {
		case vAdd:
			return fold2vAdd(args[0])
		}
	}
	if !f.IsFunction() {
		switch fv := f.Value.(type) {
		case S:
			return fold2Join(fv, args[0])
		case I, F, AB, AI, AF:
			return fold2Decode(f, args[0])
		default:
			return errType("F/x", "F", fv)
		}
	}
	if f.Rank(ctx) != 2 {
		// TODO: converge
		return errf("F/x : F rank is %d (expected 2)", f.Rank(ctx))
	}
	x := args[0]
	switch x := x.Value.(type) {
	case array:
		if x.Len() == 0 {
			f, ok := f.Value.(zeroFun)
			if ok {
				return f.zero()
			}
			return NewI(0)
		}
		r := x.at(0)
		for i := 1; i < x.Len(); i++ {
			ctx.push(x.at(i))
			ctx.push(r)
			r = ctx.applyN(f, 2)
		}
		return r
	default:
		return NewV(x)
	}
}

func fold2vAdd(x V) V {
	switch x := x.Value.(type) {
	case AB:
		n := 0
		for _, b := range x {
			if b {
				n++
			}
		}
		return NewI(n)
	case AI:
		n := 0
		for _, xi := range x {
			n += xi
		}
		return NewI(n)
	case AF:
		n := 0.0
		for _, xi := range x {
			n += xi
		}
		return NewF(n)
	case AS:
		if len(x) == 0 {
			return NewS("")
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
		return NewS(b.String())
	case AV:
		if len(x) == 0 {
			return NewI(0)
		}
		r := x[0]
		for _, xi := range x[1:] {
			r = add(r, xi)
		}
		return r
	default:
		return NewV(x)
	}
}

func fold2Join(sep S, x V) V {
	switch xv := x.Value.(type) {
	case S:
		return x
	case AS:
		return NewS(strings.Join([]string(xv), string(sep)))
	case AV:
		assertCanonical(xv)
		return errf("s/x : x not a string array (%s)", xv.Type())
	default:
		return errf("s/x : x not a string array (%s)", xv.Type())
	}
}

func fold2Decode(f V, x V) V {
	switch fv := f.Value.(type) {
	case I:
		switch x := x.Value.(type) {
		case I:
			return NewV(x)
		case F:
			if !isI(x) {
				return errf("i/x : x non-integer (%g)", x)
			}
			return NewI(int(x))
		case AI:
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += x[i] * n
				n *= int(fv)
			}
			return NewI(r)
		case AB:
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += int(B2I(x[i])) * n
				n *= int(fv)
			}
			return NewI(r)
		case AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return fold2Decode(f, aix)
		case AV:
			r := make(AV, x.Len())
			for i, xi := range x {
				r[i] = fold2Decode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return NewV(canonical(r))
		default:
			return errType("i/x", "x", x)
		}
	case F:
		if !isI(fv) {
			return errf("i/x : i non-integer (%g)", fv)
		}
		return fold2Decode(NewI(int(fv)), x)
	case AB:
		return fold2Decode(fromABtoAI(fv), x)
	case AI:
		switch x := x.Value.(type) {
		case I:
			r := 0
			n := 1
			for i := len(fv) - 1; i >= 0; i-- {
				r += int(x) * n
				n *= fv[i]
			}
			return NewI(r)
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return fold2Decode(f, NewI(int(x)))
		case AI:
			if len(fv) != len(x) {
				return errf("I/x : length mismatch: %d (#I) %d (#x)", len(fv), len(x))
			}
			r := 0
			n := 1
			for i := len(x) - 1; i >= 0; i-- {
				r += x[i] * n
				n *= fv[i]
			}
			return NewI(r)
		case AB:
			return fold2Decode(f, fromABtoAI(x))
		case AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return fold2Decode(f, aix)
		case AV:
			r := make(AV, x.Len())
			for i, xi := range x {
				r[i] = fold2Decode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return NewV(canonical(r))
		default:
			return errType("I/x", "x", x)
		}
	case AF:
		aif := toAI(fv)
		if aif.IsErr() {
			return aif
		}
		return fold2Decode(aif, x)
	default:
		// should not happen
		return errType("I/x", "I", fv)
	}
}

func fold3(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		return errf("x F/y : F not a function (%s)", f.Type())
	}
	if f.Rank(ctx) != 2 {
		return fold3While(ctx, args)
	}
	y := args[0]
	switch yv := y.Value.(type) {
	case array:
		r := args[2]
		if yv.Len() == 0 {
			return r
		}
		for i := 0; i < yv.Len(); i++ {
			ctx.push(yv.at(i))
			ctx.push(r)
			r = ctx.applyN(f, 2)
			if r.IsErr() {
				return r
			}
		}
		return canonicalV(r)
	default:
		ctx.push(y)
		ctx.push(args[2])
		return ctx.applyN(f, 2)
	}
}

func fold3While(ctx *Context, args []V) V {
	f := args[1]
	x := args[2]
	y := args[0]
	if x.IsFunction() {
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if cond.IsErr() {
				return cond
			}
			if !isTrue(cond) {
				return y
			}
			ctx.push(y)
			y = ctx.applyN(f, 1)
			if y.IsErr() {
				return y
			}
		}
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("n f/y : non-integer n (%g)", xv)
		}
		return fold3doTimes(ctx, int(xv), f, y)
	case I:
		return fold3doTimes(ctx, int(xv), f, y)
	default:
		return errType("x f/y", "x", xv)
	}
}

func fold3doTimes(ctx *Context, n int, f, y V) V {
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(f, 1)
		if y.IsErr() {
			return y
		}
	}
	return y
}

func scan2(ctx *Context, f, x V) V {
	if !f.IsFunction() {
		switch fv := f.Value.(type) {
		case S:
			return scan2Split(fv, x)
		case I, F, AB, AI, AF:
			return scan2Encode(f, x)
		default:
			return errType("f\\x", "f", fv)
		}
	}
	if f.Rank(ctx) != 2 {
		// TODO: converge
		return errf("f\\x : f rank is %d (expected 2)", f.Rank(ctx))
	}
	switch xv := x.Value.(type) {
	case array:
		if xv.Len() == 0 {
			ff, ok := f.Value.(zeroFun)
			if ok {
				return ff.zero()
			}
			return NewI(0)
		}
		r := AV{xv.at(0)}
		for i := 1; i < xv.Len(); i++ {
			ctx.push(xv.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return NewV(canonical(r))
	default:
		return x
	}
}

func scan2Split(sep S, x V) V {
	switch x := x.Value.(type) {
	case S:
		return NewV(AS(strings.Split(string(x), string(sep))))
	case AS:
		r := make(AV, len(x))
		for i := range r {
			r[i] = NewV(AS(strings.Split(x[i], string(sep))))
		}
		return NewV(r)
	case AV:
		assertCanonical(x)
		return errf("s/x : x not a string atom or array (%s)", x.Type())
	default:
		return errf("s/x : x not a string atom or array (%s)", x.Type())
	}
}

func encodeBaseDigits(b int, x int) int {
	if b < 0 {
		b = -b
	}
	if x < 0 {
		x = -x
	}
	n := 1
	for x >= b {
		x /= b
		n++
	}
	return n
}

func scan2Encode(f V, x V) V {
	switch fv := f.Value.(type) {
	case I:
		if fv == 0 {
			return errs("i\\x : base i is zero")
		}
		switch x := x.Value.(type) {
		case I:
			n := encodeBaseDigits(int(fv), int(x))
			r := make(AI, n)
			for i := n - 1; i >= 0; i-- {
				r[i] = int(x % fv)
				x /= fv
			}
			return NewV(r)
		case F:
			if !isI(x) {
				return errf("i\\x : x non-integer (%g)", x)
			}
			return scan2Encode(f, NewI(int(x)))
		case AI:
			min, max := minMax(x)
			max = int(maxI(absI(I(min)), absI(I(max))))
			n := encodeBaseDigits(int(fv), int(max))
			ai := make(AI, n*len(x))
			copy(ai[(n-1)*len(x):], x)
			for i := n - 1; i >= 0; i-- {
				for j := 0; j < len(x); j++ {
					ox := ai[i*len(x)+j]
					ai[i*len(x)+j] = ox % int(fv)
					if i > 0 {
						ai[(i-1)*len(x)+j] = ox / int(fv)
					}
				}
			}
			r := make(AV, n)
			for i := range r {
				r[i] = NewV(ai[i*len(x) : (i+1)*len(x)])
			}
			return NewV(r)
		case AB:
			return scan2Encode(f, fromABtoAI(x))
		case AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return scan2Encode(f, aix)
		case AV:
			r := make(AV, x.Len())
			for i, xi := range x {
				r[i] = scan2Encode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return NewV(canonical(r))
		default:
			return errType("i\\x", "x", x)
		}
	case F:
		if !isI(fv) {
			return errf("i\\x : i non-integer (%g)", fv)
		}
		return scan2Encode(NewI(int(fv)), x)
	case AB:
		return scan2Encode(fromABtoAI(fv), x)
	case AI:
		switch x := x.Value.(type) {
		case I:
			n := fv.Len()
			r := make(AI, n)
			for i := n - 1; i >= 0 && x > 0; i-- {
				r[i] = int(x) % fv[i]
				x /= I(fv[i])
			}
			return NewV(r)
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return scan2Encode(f, NewI(int(x)))
		case AI:
			n := fv.Len()
			ai := make(AI, n*len(x))
			copy(ai[(n-1)*len(x):], x)
			for i := n - 1; i >= 0; i-- {
				for j := 0; j < len(x); j++ {
					ox := ai[i*len(x)+j]
					ai[i*len(x)+j] = ox % fv[i]
					if i > 0 {
						ai[(i-1)*len(x)+j] = ox / fv[i]
					}
				}
			}
			r := make(AV, n)
			for i := range r {
				r[i] = NewV(ai[i*len(x) : (i+1)*len(x)])
			}
			return NewV(r)
		case AB:
			return scan2Encode(f, fromABtoAI(x))
		case AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return scan2Encode(f, aix)
		case AV:
			r := make(AV, x.Len())
			for i, xi := range x {
				r[i] = scan2Encode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return NewV(canonical(r))
		default:
			return errType("I\\x", "x", x)
		}
	case AF:
		aif := toAI(fv)
		if aif.IsErr() {
			return aif
		}
		return scan2Encode(aif, x)
	default:
		// should not happen
		return errType("I\\x", "I", fv)
	}
}

func scan3(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		return errf("x f'y : f not a function (%s)", f.Type())
	}
	if f.Rank(ctx) != 2 {
		return scan3While(ctx, args)
	}
	y := args[0]
	x := args[2]
	switch yv := y.Value.(type) {
	case array:
		if yv.Len() == 0 {
			return NewV(AV{})
		}
		ctx.push(yv.at(0))
		ctx.push(x)
		first := ctx.applyN(f, 2)
		if first.IsErr() {
			return first
		}
		r := AV{first}
		for i := 1; i < yv.Len(); i++ {
			ctx.push(yv.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return NewV(canonical(r))
	default:
		ctx.push(y)
		ctx.push(x)
		return ctx.applyN(f, 2)
	}
}

func scan3While(ctx *Context, args []V) V {
	f := args[1]
	x := args[2]
	y := args[0]
	if x.IsFunction() {
		r := AV{y}
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if cond.IsErr() {
				return cond
			}
			if !isTrue(cond) {
				return NewV(canonical(r))
			}
			ctx.push(y)
			y = ctx.applyN(f, 1)
			if y.IsErr() {
				return y
			}
			r = append(r, y)
		}
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("n f\\y : non-integer n (%g)", xv)
		}
		return scan3doTimes(ctx, int(xv), f, y)
	case I:
		return scan3doTimes(ctx, int(xv), f, y)
	default:
		return errType("x f\\y", "x", xv)
	}
}

func scan3doTimes(ctx *Context, n int, f, y V) V {
	r := AV{y}
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(f, 1)
		if y.IsErr() {
			return y
		}
		r = append(r, y)
	}
	return NewV(canonical(r))
}

func each2(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		return errf("f'x : f not a function (%s)", f.Type())
	}
	x := toArray(args[0])
	switch x := x.Value.(type) {
	case array:
		r := make(AV, 0, x.Len())
		for i := 0; i < x.Len(); i++ {
			ctx.push(x.at(i))
			next := ctx.applyN(f, 1)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return NewV(canonical(r))
	default:
		// should not happen
		return errf("f'x : x not an array (%s)", x.Type())
	}
}

func each3(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		return errf("x f'y : f not a function (%s)", f.Type())
	}
	x, okax := args[2].Value.(array)
	y, okay := args[0].Value.(array)
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
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return NewV(canonical(r))
	}
	if !okay {
		xlen := x.Len()
		r := make(AV, 0, xlen)
		for i := 0; i < xlen; i++ {
			ctx.push(args[0])
			ctx.push(x.at(i))
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return NewV(canonical(r))
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
		if next.IsErr() {
			return next
		}
		r = append(r, next)
	}
	return NewV(canonical(r))
}
