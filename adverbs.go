package goal

import "strings"

func fold2(ctx *Context, args []V) V {
	f := args[1]
	switch f.Kind {
	case Variadic:
		switch f.variadic() {
		case vAdd:
			return fold2vAdd(args[0])
		}
	}
	if !f.IsFunction() {
		if f.IsInt() {
			return fold2Decode(f, args[0])
		}
		switch fv := f.Value.(type) {
		case S:
			return fold2Join(fv, args[0])
		case F, *AB, *AI, *AF:
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
	case *AB:
		n := 0
		for _, b := range x.Slice {
			if b {
				n++
			}
		}
		return NewI(n)
	case *AI:
		n := 0
		for _, xi := range x.Slice {
			n += xi
		}
		return NewI(n)
	case *AF:
		n := 0.0
		for _, xi := range x.Slice {
			n += xi
		}
		return NewF(n)
	case *AS:
		if x.Len() == 0 {
			return NewS("")
		}
		n := 0
		for _, s := range x.Slice {
			n += len(s)
		}
		var b strings.Builder
		b.Grow(n)
		for _, s := range x.Slice {
			b.WriteString(s)
		}
		return NewS(b.String())
	case *AV:
		if x.Len() == 0 {
			return NewI(0)
		}
		r := x.At(0)
		for _, xi := range x.Slice[1:] {
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
	case *AS:
		return NewS(strings.Join([]string(xv.Slice), string(sep)))
	case *AV:
		//assertCanonical(xv)
		return errf("s/x : x not a string array (%s)", xv.Type())
	default:
		return errf("s/x : x not a string array (%s)", xv.Type())
	}
}

func fold2Decode(f V, x V) V {
	if f.IsInt() {
		if x.IsInt() {
			return x
		}
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("i/x : x non-integer (%g)", x)
			}
			return NewI(int(x))
		case *AI:
			r := 0
			n := 1
			for i := x.Len() - 1; i >= 0; i-- {
				r += x.At(i) * n
				n *= int(f.Int())
			}
			return NewI(r)
		case *AB:
			r := 0
			n := 1
			for i := x.Len() - 1; i >= 0; i-- {
				r += int(B2I(x.At(i))) * n
				n *= int(f.Int())
			}
			return NewI(r)
		case *AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return fold2Decode(f, aix)
		case *AV:
			r := make([]V, x.Len())
			for i, xi := range x.Slice {
				r[i] = fold2Decode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return canonicalV(NewAV(r))
		default:
			return errType("i/x", "x", x)
		}

	}
	switch fv := f.Value.(type) {
	case F:
		if !isI(fv) {
			return errf("i/x : i non-integer (%g)", fv)
		}
		return fold2Decode(NewI(int(fv)), x)
	case *AB:
		return fold2Decode(fromABtoAI(fv), x)
	case *AI:
		if x.IsInt() {
			r := 0
			n := 1
			for i := fv.Len() - 1; i >= 0; i-- {
				r += int(x.Int()) * n
				n *= fv.At(i)
			}
			return NewI(r)
		}
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return fold2Decode(f, NewI(int(x)))
		case *AI:
			if fv.Len() != x.Len() {
				return errf("I/x : length mismatch: %d (#I) %d (#x)", fv.Len(), x.Len())
			}
			r := 0
			n := 1
			for i := x.Len() - 1; i >= 0; i-- {
				r += x.At(i) * n
				n *= fv.At(i)
			}
			return NewI(r)
		case *AB:
			return fold2Decode(f, fromABtoAI(x))
		case *AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return fold2Decode(f, aix)
		case *AV:
			r := make([]V, x.Len())
			for i, xi := range x.Slice {
				r[i] = fold2Decode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return canonicalV(NewAV(r))
		default:
			return errType("I/x", "x", x)
		}
	case *AF:
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
	if x.IsInt() {
		return fold3doTimes(ctx, x.Int(), f, y)
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("n f/y : non-integer n (%g)", xv)
		}
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
		if f.IsInt() {
			return scan2Encode(f, x)
		}
		switch fv := f.Value.(type) {
		case S:
			return scan2Split(fv, x)
		case F, *AB, *AI, *AF:
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
		r := []V{xv.at(0)}
		for i := 1; i < xv.Len(); i++ {
			ctx.push(xv.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return canonicalV(NewAV(r))
	default:
		return x
	}
}

func scan2Split(sep S, x V) V {
	switch x := x.Value.(type) {
	case S:
		return NewAS(strings.Split(string(x), string(sep)))
	case *AS:
		r := make([]V, x.Len())
		for i := range r {
			r[i] = NewAS(strings.Split(x.At(i), string(sep)))
		}
		return NewAV(r)
	case *AV:
		//assertCanonical(x)
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
	if f.IsInt() {
		if f.Int() == 0 {
			return errs("i\\x : base i is zero")
		}
		if x.IsInt() {
			n := encodeBaseDigits(int(f.Int()), int(x.Int()))
			r := make([]int, n)
			for i := n - 1; i >= 0; i-- {
				r[i] = int(x.Int() % f.Int())
				x.N /= f.Int()
			}
			return NewAI(r)
		}
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("i\\x : x non-integer (%g)", x)
			}
			return scan2Encode(f, NewI(int(x)))
		case *AI:
			min, max := minMax(x)
			max = maxI(absI(min), absI(max))
			n := encodeBaseDigits(f.Int(), max)
			ai := make([]int, n*x.Len())
			copy(ai[(n-1)*x.Len():], x.Slice)
			for i := n - 1; i >= 0; i-- {
				for j := 0; j < x.Len(); j++ {
					ox := ai[i*x.Len()+j]
					ai[i*x.Len()+j] = ox % int(f.Int())
					if i > 0 {
						ai[(i-1)*x.Len()+j] = ox / int(f.Int())
					}
				}
			}
			r := make([]V, n)
			for i := range r {
				r[i] = NewAI(ai[i*x.Len() : (i+1)*x.Len()])
			}
			return NewAV(r)
		case *AB:
			return scan2Encode(f, fromABtoAI(x))
		case *AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return scan2Encode(f, aix)
		case *AV:
			r := make([]V, x.Len())
			for i, xi := range x.Slice {
				r[i] = scan2Encode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return canonicalV(NewAV(r))
		default:
			return errType("i\\x", "x", x)
		}

	}
	switch fv := f.Value.(type) {
	case F:
		if !isI(fv) {
			return errf("i\\x : i non-integer (%g)", fv)
		}
		return scan2Encode(NewI(int(fv)), x)
	case *AB:
		return scan2Encode(fromABtoAI(fv), x)
	case *AI:
		if x.IsInt() {
			// TODO: check for zero division
			n := fv.Len()
			r := make([]int, n)
			for i := n - 1; i >= 0 && x.Int() > 0; i-- {
				r[i] = int(x.Int()) % fv.At(i)
				x.N /= fv.At(i)
			}
			return NewAI(r)

		}
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("I/x : x non-integer (%g)", x)
			}
			return scan2Encode(f, NewI(int(x)))
		case *AI:
			n := fv.Len()
			ai := make([]int, n*x.Len())
			copy(ai[(n-1)*x.Len():], x.Slice)
			for i := n - 1; i >= 0; i-- {
				for j := 0; j < x.Len(); j++ {
					ox := ai[i*x.Len()+j]
					ai[i*x.Len()+j] = ox % fv.At(i)
					if i > 0 {
						ai[(i-1)*x.Len()+j] = ox / fv.At(i)
					}
				}
			}
			r := make([]V, n)
			for i := range r {
				r[i] = NewAI(ai[i*x.Len() : (i+1)*x.Len()])
			}
			return NewAV(r)
		case *AB:
			return scan2Encode(f, fromABtoAI(x))
		case *AF:
			aix := toAI(x)
			if aix.IsErr() {
				return aix
			}
			return scan2Encode(f, aix)
		case *AV:
			r := make([]V, x.Len())
			for i, xi := range x.Slice {
				r[i] = scan2Encode(f, xi)
				if r[i].IsErr() {
					return r[i]
				}
			}
			return canonicalV(NewAV(r))
		default:
			return errType("I\\x", "x", x)
		}
	case *AF:
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
			return NewAV([]V{})
		}
		ctx.push(yv.at(0))
		ctx.push(x)
		first := ctx.applyN(f, 2)
		if first.IsErr() {
			return first
		}
		r := []V{first}
		for i := 1; i < yv.Len(); i++ {
			ctx.push(yv.at(i))
			ctx.push(r[len(r)-1])
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return canonicalV(NewAV(r))
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
		r := []V{y}
		for {
			ctx.push(y)
			cond := ctx.applyN(x, 1)
			if cond.IsErr() {
				return cond
			}
			if !isTrue(cond) {
				return canonicalV(NewAV(r))
			}
			ctx.push(y)
			y = ctx.applyN(f, 1)
			if y.IsErr() {
				return y
			}
			r = append(r, y)
		}
	}
	if x.IsInt() {
		return scan3doTimes(ctx, x.Int(), f, y)
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("n f\\y : non-integer n (%g)", xv)
		}
		return scan3doTimes(ctx, int(xv), f, y)
	default:
		return errType("x f\\y", "x", xv)
	}
}

func scan3doTimes(ctx *Context, n int, f, y V) V {
	r := []V{y}
	for i := 0; i < n; i++ {
		ctx.push(y)
		y = ctx.applyN(f, 1)
		if y.IsErr() {
			return y
		}
		r = append(r, y)
	}
	return canonicalV(NewAV(r))
}

func each2(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		return errf("f'x : f not a function (%s)", f.Type())
	}
	x := toArray(args[0])
	switch x := x.Value.(type) {
	case array:
		r := make([]V, 0, x.Len())
		for i := 0; i < x.Len(); i++ {
			ctx.push(x.at(i))
			next := ctx.applyN(f, 1)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return canonicalV(NewAV(r))
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
		r := make([]V, 0, ylen)
		for i := 0; i < ylen; i++ {
			ctx.push(y.at(i))
			ctx.push(args[2])
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return canonicalV(NewAV(r))
	}
	if !okay {
		xlen := x.Len()
		r := make([]V, 0, xlen)
		for i := 0; i < xlen; i++ {
			ctx.push(args[0])
			ctx.push(x.at(i))
			next := ctx.applyN(f, 2)
			if next.IsErr() {
				return next
			}
			r = append(r, next)
		}
		return canonicalV(NewAV(r))
	}
	xlen := x.Len()
	if xlen != y.Len() {
		return errf("x f'y : length mismatch: %d (#x) vs %d (#y)", x.Len(), y.Len())
	}
	r := make([]V, 0, xlen)
	for i := 0; i < xlen; i++ {
		ctx.push(y.at(i))
		ctx.push(x.at(i))
		next := ctx.applyN(f, 2)
		if next.IsErr() {
			return next
		}
		r = append(r, next)
	}
	return canonicalV(NewAV(r))
}
