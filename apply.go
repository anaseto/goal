package goal

//import "fmt"

// Apply calls a value with a single argument.
func (ctx *Context) Apply(x, y V) V {
	ctx.push(y)
	r := ctx.applyN(x, 1)
	return r
}

// Apply2 calls a value with two arguments.
func (ctx *Context) Apply2(x, y, z V) V {
	ctx.push(z)
	ctx.push(y)
	r := ctx.applyN(x, 2)
	return r
}

// ApplyN calls a value with one or more arguments. The arguments should be
// provided in reverse order, given the stack-based right to left semantics
// used by the language.
func (ctx *Context) ApplyN(x V, args []V) V {
	if len(args) == 0 {
		panic("ApplyArgs: len(args) should be > 0")
	}
	ctx.pushArgs(args)
	r := ctx.applyN(x, len(args))
	return r
}

func rcincr(args []V) {
	for _, v := range args {
		v.rcincr()
	}
}

func rcdecr(args []V) {
	for _, v := range args {
		v.rcdecr()
	}
}

// applyN applies x with the top n arguments in the stack. It consumes the
// arguments, but does not push the result, returing it instead.
func (ctx *Context) applyN(x V, n int) V {
	switch x.kind {
	case valLambda:
		return ctx.applyLambda(x.lambda(), n)
	case valVariadic:
		switch n {
		case 1:
			return ctx.applyVariadic(x.variadic())
		case 2:
			return ctx.apply2Variadic(x.variadic())
		default:
			return ctx.applyNVariadic(x.variadic(), n)
		}
	}
	switch xv := x.value.(type) {
	case derivedVerb:
		ctx.push(xv.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			return NewV(projection{Fun: x, Args: ctx.popN(n + 1)})
		}
		rcincr(args)
		r := ctx.variadics[xv.Fun].Func(ctx, args)
		rcdecr(args)
		ctx.dropN(n + 1)
		return r
	case projectionFirst:
		if n > 1 {
			return panicf("too many arguments: got %d, expected 1", n)
		}
		ctx.push(xv.Arg)
		xv.Arg.rcincr()
		r := ctx.applyN(xv.Fun, 2)
		xv.Arg.rcdecr()
		return r
	case projectionMonad:
		if n > 1 {
			return panicf("too many arguments: got %d, expected 1", n)
		}
		return ctx.applyN(xv.Fun, 1)
	case projection:
		return ctx.applyProjection(xv, n)
	case S:
		switch n {
		case 1:
			y := ctx.pop()
			y.rcincr()
			r := applyS(xv, y)
			y.rcdecr()
			return r
		case 2:
			args := ctx.peekN(n)
			rcincr(args)
			r := applyS2(xv, args[1], args[0])
			rcdecr(args)
			ctx.dropN(n)
			return r
		default:
			return panicf("too many arguments")
		}
	case array:
		switch n {
		case 1:
			return ctx.applyArray(x, ctx.pop())
		default:
			args := ctx.peekN(n)
			rcincr(args)
			r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
			rcdecr(args)
			ctx.dropN(n)
			return r
		}
	default:
		return panicf("type %s cannot be applied", x.Type())
	}
}

// applyArray applies an array to a value.
func (ctx *Context) applyArray(x V, y V) V {
	xv := x.value.(array)
	if y.kind == valNil {
		return x
	}
	if y.IsI() {
		i := y.I()
		if i < 0 {
			i = int64(xv.Len()) + i
		}
		if i < 0 || i >= int64(xv.Len()) {
			return panicf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(int(i))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return panicf("x[y] : non-integer index (%g)", y.F())
		}
		i := int64(y.F())
		if i < 0 {
			i = int64(xv.Len()) + i
		}
		if i < 0 || i >= int64(xv.Len()) {
			return panicf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(int(i))
	}
	switch yv := y.value.(type) {
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = ctx.applyArray(x, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	case array:
		iy := toIndices(y)
		if iy.IsPanic() {
			return panicf("x[y] : %v", iy.value)
		}
		r := xv.atIndices(iy.value.(*AI).Slice)
		return r
	default:
		return panicf("x[y] : y non-array non-integer (%s)", y.Type())
	}
}

func (ctx *Context) applyArrayArgs(x V, arg V, args []V) V {
	xv := x.value.(array)
	// TODO: annotate error with depth?
	if len(args) == 0 {
		return ctx.applyArray(x, arg)
	}
	if arg.kind == valNil {
		r := make([]V, xv.Len())
		for i := 0; i < len(r); i++ {
			r[i] = ctx.ApplyN(xv.at(i), args)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	}
	switch argv := arg.value.(type) {
	case array:
		r := make([]V, argv.Len())
		for i := 0; i < argv.Len(); i++ {
			r[i] = ctx.applyArrayArgs(x, argv.at(i), args)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		r := ctx.applyArray(x, arg)
		if r.IsPanic() {
			return r
		}
		return ctx.ApplyN(r, args)
	}
}

func (ctx *Context) applyVariadic(v variadic) V {
	args := ctx.peek()
	x := args[0]
	if x.kind == valNil {
		ctx.drop()
		return NewV(projectionMonad{Fun: newVariadic(v)})
	}
	if ctx.variadics[v].Adverb {
		ctx.drop()
		return NewV(derivedVerb{Fun: v, Arg: x})
	}
	x.rcincr()
	r := ctx.variadics[v].Func(ctx, args)
	x.rcdecr()
	ctx.drop()
	return r
}

func (ctx *Context) apply2Variadic(v variadic) V {
	args := ctx.peekN(2)
	if args[0].kind == valNil {
		if args[1].kind != valNil {
			arg := args[1]
			ctx.drop2()
			return NewV(projectionFirst{Fun: newVariadic(v), Arg: arg})
		}
		return NewV(projection{Fun: newVariadic(v), Args: ctx.popN(2)})
	} else if args[1].kind == valNil {
		return NewV(projection{Fun: newVariadic(v), Args: ctx.popN(2)})
	}
	args[0].rcincr()
	args[1].rcincr()
	r := ctx.variadics[v].Func(ctx, args)
	args[0].rcdecr()
	args[1].rcdecr()
	ctx.drop2()
	return r
}

func (ctx *Context) applyNVariadic(v variadic, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		return NewV(projection{Fun: newVariadic(v), Args: ctx.popN(n)})
	}
	rcincr(args)
	r := ctx.variadics[v].Func(ctx, args)
	rcdecr(args)
	ctx.dropN(n)
	return r
}

func (ctx *Context) applyProjection(p projection, n int) V {
	args := ctx.peekN(n)
	nNils := countNils(p.Args)
	switch {
	case len(args) > nNils:
		return panics("too many arguments")
	case len(args) == nNils:
		nilc := 0
		for _, arg := range p.Args {
			switch {
			case arg.kind != valNil:
				ctx.push(arg)
			default:
				ctx.push(args[nilc])
				nilc++
			}
		}
		r := ctx.applyN(p.Fun, len(p.Args))
		ctx.dropN(n)
		return r
	default:
		vargs := cloneArgs(p.Args)
		nilc := 1
		for i := len(vargs) - 1; i >= 0; i-- {
			if vargs[i].kind == valNil {
				if nilc > len(args) {
					break
				}
				vargs[i] = args[len(args)-nilc]
				nilc++
			}
		}
		ctx.dropN(n)
		return NewV(projection{Fun: NewV(p), Args: vargs})
	}
}

func (ctx *Context) applyLambda(id lambda, n int) V {
	if ctx.callDepth > maxCallDepth {
		return panics("lambda: exceeded maximum call depth")
	}
	lc := ctx.lambdas[int(id)]
	if lc.Rank < n {
		return panicf("lambda: too many arguments: got %d, expected %d", n, lc.Rank)
	}
	args := ctx.peekN(n)
	if lc.Rank > n || hasNil(args) {
		if n == 1 {
			if args[0].kind == valNil {
				ctx.drop() // drop nil
				return NewV(projectionMonad{Fun: newLambda(id)})
			}
			return NewV(projectionFirst{Fun: newLambda(id), Arg: ctx.pop()})
		}
		if n == 2 && args[1].kind == valNil && args[0].kind != valNil {
			return NewV(projectionFirst{Fun: newLambda(id), Arg: ctx.pop()})
		}
		return NewV(projection{Fun: newLambda(id), Args: ctx.popN(n)})
	}
	for i, arg := range args {
		if lc.lastUses[i].bn >= 1 {
			arg.rcincr()
		}
	}
	nVars := len(lc.Names) - lc.Rank
	olen := len(ctx.stack)
	for i := 0; i < nVars; i++ {
		ctx.push(V{})
	}
	oframeIdx := ctx.frameIdx
	ctx.frameIdx = int32(len(ctx.stack) - 1)

	olambda := ctx.lambda
	ctx.lambda = int(id)
	ctx.callDepth++
	ip, err := ctx.execute(lc.Body)
	ctx.callDepth--
	ctx.lambda = olambda

	if err != nil {
		ctx.updateErrPos(ip, lc)
		return panics(err.Error())
	}
	var r V
	switch len(ctx.stack) {
	case olen + nVars:
	case olen + nVars + 1:
		r = ctx.stack[len(ctx.stack)-1]
		ctx.drop()
	default:
		ctx.updateErrPos(ip, lc)
		// should not happen
		return panicf("lambda %d: bad len %d vs old %d (depth: %d): %v", id, len(ctx.stack), olen, ctx.callDepth, ctx.stack)
	}
	if nVars > 0 {
		ctx.dropN(nVars)
	}
	ctx.dropN(n)
	ctx.frameIdx = oframeIdx
	return r
}

func (x *AV) atIndices(y []int64) V {
	r := make([]V, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return canonicalV(NewAV(r))
}

func (x *AB) atIndices(y []int64) V {
	r := make([]bool, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAB(r)
}

func (x *AI) atIndices(y []int64) V {
	r := make([]int64, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAI(r)
}

func (x *AF) atIndices(y []int64) V {
	r := make([]float64, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAF(r)
}

func (x *AS) atIndices(y []int64) V {
	r := make([]string, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAS(r)
}

// set changes x at i with y (in place).
func (x AV) set(i int, y V) {
	x.Slice[i] = y
}

// set changes x at i with y (in place).
func (x AB) set(i int, y V) {
	x.Slice[i] = y.n == 1
}

// set changes x at i with y (in place).
func (x AI) set(i int, y V) {
	x.Slice[i] = y.n
}

// set changes x at i with y (in place).
func (x AF) set(i int, y V) {
	x.Slice[i] = y.F()
}

// set changes x at i with y (in place).
func (x AS) set(i int, y V) {
	x.Slice[i] = string(y.value.(S))
}

//// setIndices x at y with z (in place).
//func (x AV) setIndices(y AI, z V) error {
//az := z.BV.(array)
//for i, yi := range y.Slice {
//if yi < 0 {
//yi += len(x)
//}
//if yi < 0 || yi >= len(x) {
//return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x)).BV
//}
//x[yi] = az.at(i)
//}
//return nil
//}

//// setIndices x at y with z (in place).
//func (x AI) setIndices(y AI, z V) error {
//az := z.BV.(AI)
//for i, yi := range y.Slice {
//if yi < 0 {
//yi += len(x)
//}
//if yi < 0 || yi >= len(x) {
//return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x)).BV
//}
//x[yi] = az[i]
//}
//return nil
//}

//// setIndices x at y with z (in place).
//func (x AF) setIndices(y AI, z V) error {
//az := z.BV.(AF)
//for i, yi := range y.Slice {
//if yi < 0 {
//yi += len(x)
//}
//if yi < 0 || yi >= len(x) {
//return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x)).BV
//}
//x[yi] = az[i]
//}
//return nil
//}

//// setIndices x at y with z (in place).
//func (x AB) setIndices(y AI, z V) error {
//az := z.BV.(AB)
//for i, yi := range y.Slice {
//if yi < 0 {
//yi += len(x)
//}
//if yi < 0 || yi >= len(x) {
//return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x)).BV
//}
//x[yi] = az[i]
//}
//return nil
//}

//// setIndices x at y with z (in place).
//func (x AS) setIndices(y AI, z V) error {
//az := z.BV.(AS)
//for i, yi := range y.Slice {
//if yi < 0 {
//yi += len(x)
//}
//if yi < 0 || yi >= len(x) {
//return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x)).BV
//}
//x[yi] = az[i]
//}
//return nil
//}
