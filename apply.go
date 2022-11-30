package goal

//import "fmt"

// Apply calls a value with a single argument.
func (ctx *Context) Apply(x, y V) V {
	ctx.push(y)
	return ctx.applyN(x, 1)
}

// Apply2 calls a value with a two arguments.
func (ctx *Context) Apply2(x, y, z V) V {
	ctx.push(z)
	ctx.push(y)
	return ctx.applyN(x, 2)
}

// ApplyN calls a value with one or more arguments. The arguments should be
// provided in reverse order, given the stack-based right to left semantics
// used by the language.
func (ctx *Context) ApplyN(x V, args []V) V {
	if len(args) == 0 {
		panic("ApplyArgs: len(args) should be > 0")
	}
	ctx.pushArgs(args)
	return ctx.applyN(x, len(args))
}

// applyN applies x with the top n arguments in the stack. It consumes the
// arguments, but does not push the result, returing it instead.
func (ctx *Context) applyN(x V, n int) V {
	switch xv := x.Value.(type) {
	case Lambda:
		return ctx.applyLambda(xv, n)
	case Variadic:
		if n == 1 {
			return ctx.applyVariadic(xv)
		}
		return ctx.applyNVariadic(xv, n)
	case DerivedVerb:
		ctx.push(xv.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			return NewV(Projection{Fun: x, Args: ctx.popN(n + 1)})
		}
		r := ctx.variadics[xv.Fun].Func(ctx, args)
		ctx.dropN(n + 1)
		return r
	case ProjectionFirst:
		if n > 1 {
			return errf("too many arguments: got %d, expected 1", n)
		}
		ctx.push(xv.Arg)
		return ctx.applyN(xv.Fun, 2)
	case ProjectionMonad:
		if n > 1 {
			return errf("too many arguments: got %d, expected 1", n)
		}
		return ctx.applyN(xv.Fun, 1)
	case Projection:
		return ctx.applyProjection(xv, n)
	case S:
		switch n {
		case 1:
			return applyS(xv, ctx.pop())
		case 2:
			args := ctx.peekN(n)
			r := applyS2(xv, args[1], args[0])
			ctx.dropN(n)
			return r
		default:
			return errf("too many arguments")
		}
	case array:
		switch n {
		case 1:
			return ctx.applyArray(x, ctx.pop())
		default:
			args := ctx.peekN(n)
			r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
			ctx.dropN(n)
			return r
		}
	default:
		return errf("type %s cannot be applied", x.Type())
	}
}

// applyArray applies an array to a value.
func (ctx *Context) applyArray(x V, y V) V {
	xv := x.Value.(array)
	if y == (V{}) {
		return x
	}
	switch yv := y.Value.(type) {
	case F:
		if !isI(yv) {
			return errf("x[y] : non-integer index (%g)", yv)
		}
		i := int(yv)
		if i < 0 {
			i = xv.Len() + i
		}
		if i < 0 || i >= xv.Len() {
			return errf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(i)
	case I:
		i := int(yv)
		if i < 0 {
			i = xv.Len() + i
		}
		if i < 0 || i >= xv.Len() {
			return errf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(i)
	case AV:
		r := make(AV, yv.Len())
		for i, yi := range yv {
			r[i] = ctx.applyArray(x, yi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	case array:
		iy := toIndices(y)
		if isErr(iy) {
			return errf("x[y] : %v", iy.Value)
		}
		r := xv.atIndices(iy.Value.(AI))
		return r
	default:
		return errf("x[y] : y non-array non-integer (%s)", y.Type())
	}
}

func (ctx *Context) applyArrayArgs(x V, arg V, args []V) V {
	xv := x.Value.(array)
	// TODO: annotate error with depth?
	if len(args) == 0 {
		return ctx.applyArray(x, arg)
	}
	if arg == (V{}) {
		r := make(AV, xv.Len())
		for i := 0; i < len(r); i++ {
			r[i] = ctx.ApplyN(xv.at(i), args)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	}
	switch argv := arg.Value.(type) {
	case array:
		r := make(AV, argv.Len())
		for i := 0; i < argv.Len(); i++ {
			r[i] = ctx.applyArrayArgs(x, argv.at(i), args)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(canonical(r))
	default:
		r := ctx.applyArray(x, arg)
		if isErr(r) {
			return r
		}
		return ctx.ApplyN(r, args)
	}
}

func (ctx *Context) applyVariadic(v Variadic) V {
	args := ctx.peek()
	x := args[0]
	if x == (V{}) {
		ctx.drop()
		return NewV(ProjectionMonad{Fun: NewV(v)})
	}
	if ctx.variadics[v].Adverb {
		ctx.drop()
		return NewV(DerivedVerb{Fun: v, Arg: x})
	}
	r := ctx.variadics[v].Func(ctx, args)
	ctx.drop()
	return r
}

func (ctx *Context) applyNVariadic(v Variadic, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		if n == 2 {
			if args[1] != (V{}) {
				arg := args[1]
				ctx.dropN(n)
				return NewV(ProjectionFirst{Fun: NewV(v), Arg: arg})
			}
		}
		return NewV(Projection{Fun: NewV(v), Args: ctx.popN(n)})
	}
	r := ctx.variadics[v].Func(ctx, args)
	ctx.dropN(n)
	return r
}

func (ctx *Context) applyProjection(p Projection, n int) V {
	args := ctx.peekN(n)
	nNils := countNils(p.Args)
	switch {
	case len(args) > nNils:
		return errs("too many arguments")
	case len(args) == nNils:
		nilc := 0
		for _, arg := range p.Args {
			switch {
			case arg != (V{}):
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
			if vargs[i] == (V{}) {
				if nilc > len(args) {
					break
				}
				vargs[i] = args[len(args)-nilc]
				nilc++
			}
		}
		ctx.dropN(n)
		return NewV(Projection{Fun: NewV(p), Args: vargs})
	}
}

func (ctx *Context) applyLambda(id Lambda, n int) V {
	if ctx.callDepth > maxCallDepth {
		return errs("lambda: exceeded maximum call depth")
	}
	lc := ctx.lambdas[int(id)]
	if lc.Rank < n {
		return errf("lambda: too many arguments: got %d, expected %d", n, lc.Rank)
	}
	args := ctx.peekN(n)
	if lc.Rank > n || hasNil(args) {
		if n == 1 {
			if args[0] == (V{}) {
				ctx.drop() // drop nil
				return NewV(ProjectionMonad{Fun: NewV(id)})
			}
			return NewV(ProjectionFirst{Fun: NewV(id), Arg: ctx.pop()})
		}
		if n == 2 && args[1] == (V{}) && args[0] != (V{}) {
			return NewV(ProjectionFirst{Fun: NewV(id), Arg: ctx.pop()})
		}
		return NewV(Projection{Fun: NewV(id), Args: ctx.popN(n)})
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
		return errs(err.Error())
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
		return errf("lambda %d: bad len %d vs old %d (depth: %d): %v", id, len(ctx.stack), olen, ctx.callDepth, ctx.stack)
	}
	if nVars > 0 {
		ctx.dropN(nVars)
	}
	ctx.dropN(n)
	ctx.frameIdx = oframeIdx
	return r
}

func (x AV) atIndices(y AI) V {
	r := make(AV, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		r[i] = x[yi]
	}
	return NewV(canonical(r))
}

func (x AB) atIndices(y AI) V {
	r := make(AB, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		r[i] = x[yi]
	}
	return NewV(r)
}

func (x AI) atIndices(y AI) V {
	r := make(AI, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		r[i] = x[yi]
	}
	return NewV(r)
}

func (x AF) atIndices(y AI) V {
	r := make(AF, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		r[i] = x[yi]
	}
	return NewV(r)
}

func (x AS) atIndices(y AI) V {
	r := make(AS, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		r[i] = x[yi]
	}
	return NewV(r)
}

// set changes x at i with y (in place).
func (x AV) set(i int, y V) {
	x[i] = y
}

// set changes x at i with y (in place).
func (x AB) set(i int, y V) {
	x[i] = y.Value.(I) == 1
}

// set changes x at i with y (in place).
func (x AI) set(i int, y V) {
	x[i] = int(y.Value.(I))
}

// set changes x at i with y (in place).
func (x AF) set(i int, y V) {
	x[i] = float64(y.Value.(F))
}

// set changes x at i with y (in place).
func (x AS) set(i int, y V) {
	x[i] = string(y.Value.(S))
}

//// setIndices x at y with z (in place).
//func (x AV) setIndices(y AI, z V) error {
//az := z.BV.(array)
//for i, yi := range y {
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
//for i, yi := range y {
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
//for i, yi := range y {
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
//for i, yi := range y {
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
//for i, yi := range y {
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
