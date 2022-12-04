package goal

//import "fmt"

// Apply calls a value with a single argument.
func (ctx *Context) Apply(x, y V) V {
	ctx.push(y)
	r := ctx.applyN(x, 1)
	return r
}

// Apply2 calls a value with a two arguments.
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
	switch x.Kind {
	case Lambda:
		return ctx.applyLambda(x.lambda(), n)
	case Variadic:
		if n == 1 {
			return ctx.applyVariadic(x.variadic())
		}
		return ctx.applyNVariadic(x.variadic(), n)
	}
	switch xv := x.Value.(type) {
	case DerivedVerb:
		ctx.push(xv.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			return NewV(Projection{Fun: x, Args: ctx.popN(n + 1)})
		}
		rcincr(args)
		r := ctx.variadics[xv.Fun].Func(ctx, args)
		rcdecr(args)
		ctx.dropN(n + 1)
		return r
	case ProjectionFirst:
		if n > 1 {
			return errf("too many arguments: got %d, expected 1", n)
		}
		ctx.push(xv.Arg)
		xv.Arg.rcincr()
		r := ctx.applyN(xv.Fun, 2)
		xv.Arg.rcdecr()
		return r
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
			return errf("too many arguments")
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
		return errf("type %s cannot be applied", x.Type())
	}
}

// applyArray applies an array to a value.
func (ctx *Context) applyArray(x V, y V) V {
	xv := x.Value.(array)
	if y == (V{}) {
		return x
	}
	if y.IsInt() {
		i := y.Int()
		if i < 0 {
			i = xv.Len() + i
		}
		if i < 0 || i >= xv.Len() {
			return errf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(i)

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
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = ctx.applyArray(x, yi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	case array:
		iy := toIndices(y)
		if iy.IsErr() {
			return errf("x[y] : %v", iy.Value)
		}
		r := xv.atIndices(iy.Value.(*AI).Slice)
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
		r := make([]V, xv.Len())
		for i := 0; i < len(r); i++ {
			r[i] = ctx.ApplyN(xv.at(i), args)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	}
	switch argv := arg.Value.(type) {
	case array:
		r := make([]V, argv.Len())
		for i := 0; i < argv.Len(); i++ {
			r[i] = ctx.applyArrayArgs(x, argv.at(i), args)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		r := ctx.applyArray(x, arg)
		if r.IsErr() {
			return r
		}
		return ctx.ApplyN(r, args)
	}
}

func (ctx *Context) applyVariadic(v variadic) V {
	args := ctx.peek()
	x := args[0]
	if x == (V{}) {
		ctx.drop()
		return NewV(ProjectionMonad{Fun: NewVariadic(v)})
	}
	if ctx.variadics[v].Adverb {
		ctx.drop()
		return NewV(DerivedVerb{Fun: v, Arg: x})
	}
	x.rcincr()
	r := ctx.variadics[v].Func(ctx, args)
	x.rcdecr()
	ctx.drop()
	return r
}

func (ctx *Context) applyNVariadic(v variadic, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		if n == 2 {
			if args[1] != (V{}) {
				arg := args[1]
				ctx.dropN(n)
				return NewV(ProjectionFirst{Fun: NewVariadic(v), Arg: arg})
			}
		}
		return NewV(Projection{Fun: NewVariadic(v), Args: ctx.popN(n)})
	}
	rcincr(args)
	r := ctx.variadics[v].Func(ctx, args)
	rcdecr(args)
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

func (ctx *Context) applyLambda(id lambda, n int) V {
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
				return NewV(ProjectionMonad{Fun: NewLambda(id)})
			}
			return NewV(ProjectionFirst{Fun: NewLambda(id), Arg: ctx.pop()})
		}
		if n == 2 && args[1] == (V{}) && args[0] != (V{}) {
			return NewV(ProjectionFirst{Fun: NewLambda(id), Arg: ctx.pop()})
		}
		return NewV(Projection{Fun: NewLambda(id), Args: ctx.popN(n)})
	}
	for i, arg := range args {
		if lc.lastUses[i].bn >= 0 {
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

func (x *AV) atIndices(y []int) V {
	r := make([]V, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= x.Len() {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, x.Len())
		}
		r[i] = x.At(yi)
	}
	return canonicalV(NewAV(r))
}

func (x *AB) atIndices(y []int) V {
	r := make([]bool, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= x.Len() {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, x.Len())
		}
		r[i] = x.At(yi)
	}
	return NewAB(r)
}

func (x *AI) atIndices(y []int) V {
	r := make([]int, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= x.Len() {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, x.Len())
		}
		r[i] = x.At(yi)
	}
	return NewAI(r)
}

func (x *AF) atIndices(y []int) V {
	r := make([]float64, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= x.Len() {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, x.Len())
		}
		r[i] = x.At(yi)
	}
	return NewAF(r)
}

func (x *AS) atIndices(y []int) V {
	r := make([]string, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= x.Len() {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, x.Len())
		}
		r[i] = x.At(yi)
	}
	return NewAS(r)
}

// set changes x at i with y (in place).
func (x AV) set(i int, y V) {
	x.Slice[i] = y
}

// set changes x at i with y (in place).
func (x AB) set(i int, y V) {
	x.Slice[i] = y.N == 1
}

// set changes x at i with y (in place).
func (x AI) set(i int, y V) {
	x.Slice[i] = y.N
}

// set changes x at i with y (in place).
func (x AF) set(i int, y V) {
	x.Slice[i] = float64(y.Value.(F))
}

// set changes x at i with y (in place).
func (x AS) set(i int, y V) {
	x.Slice[i] = string(y.Value.(S))
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
