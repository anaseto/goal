package goal

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
	switch x := x.(type) {
	case Lambda:
		return ctx.applyLambda(x, n)
	case Variadic:
		if n == 1 {
			return ctx.applyVariadic(x)
		}
		return ctx.applyNVariadic(x, n)
	case DerivedVerb:
		ctx.push(x.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			return Projection{Fun: x, Args: ctx.popN(n + 1)}
		}
		res := ctx.variadics[x.Fun].Func(ctx, args)
		ctx.dropN(n + 1)
		return res
	case ProjectionOne:
		if n > 1 {
			return errf("too many arguments: got %d, expected 1", n)
		}
		arg := ctx.top()
		if arg == nil {
			ctx.drop()
			return x
		}
		ctx.push(x.Arg)
		return ctx.applyN(x.Fun, 2)
	case Projection:
		return ctx.applyProjection(x, n)
	case Composition:
		res := ctx.applyN(x.Right, n)
		_, ok := res.(error)
		if ok {
			return res
		}
		ctx.push(res)
		return ctx.applyN(x.Left, 1)
	case S:
		switch n {
		case 1:
			return applyS(x, ctx.pop())
		case 2:
			args := ctx.peekN(n)
			res := applyS2(x, args[1], args[0])
			ctx.dropN(n)
			return res
		default:
			return errf("too many arguments")
		}
	case array:
		switch n {
		case 1:
			return ctx.applyArray(x, ctx.pop())
		default:
			args := ctx.peekN(n)
			res := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
			ctx.dropN(n)
			return res
		}
	default:
		return errf("type %s cannot be applied", x.Type())
	}
}

// applyArray applies an array to a value.
func (ctx *Context) applyArray(x array, y V) V {
	if y == nil {
		return x
	}
	switch y := y.(type) {
	case F:
		if !isI(y) {
			return errf("a[x] : non-integer index (%g)", y)
		}
		i := int(y)
		if i < 0 {
			i = x.Len() + i
		}
		if i < 0 || i >= x.Len() {
			return errf("a[x] : out of bounds index: %d", i)
		}
		return x.at(i)
	case I:
		i := int(y)
		if i < 0 {
			i = x.Len() + i
		}
		if i < 0 || i >= x.Len() {
			return errf("a[x] : out of bounds index: %d", i)
		}
		return x.at(i)
	case AV:
		res := make(AV, y.Len())
		for i, yi := range y {
			res[i] = ctx.applyArray(x, yi)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	case array:
		indices := toIndices(y)
		if err, ok := indices.(errV); ok {
			return errV("x[y] :") + err
		}
		res := x.atIndices(indices.(AI))
		return res
	default:
		return errf("a[x] : x non-array non-integer")
	}
}

func (ctx *Context) applyArrayArgs(x array, arg V, args []V) V {
	// TODO: annotate error with depth?
	if len(args) == 0 {
		return ctx.applyArray(x, arg)
	}
	if arg == nil {
		res := make(AV, x.Len())
		for i := 0; i < len(res); i++ {
			res[i] = ctx.ApplyN(x.at(i), args)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	}
	switch arg := arg.(type) {
	case array:
		res := make(AV, arg.Len())
		for i := 0; i < arg.Len(); i++ {
			res[i] = ctx.applyArrayArgs(x, arg.at(i), args)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	default:
		res := ctx.applyArray(x, arg)
		if _, ok := res.(errV); ok {
			return res
		}
		return ctx.ApplyN(res, args)
	}
}

func (ctx *Context) applyVariadic(v Variadic) V {
	args := ctx.peek()
	vv := args[0]
	if ctx.variadics[v].Adverb {
		ctx.drop()
		return DerivedVerb{Fun: v, Arg: vv}
	}
	if vv == nil {
		ctx.drop()
		return Projection{Fun: v, Args: []V{nil}}
	}
	switch vv := vv.(type) {
	case Composition:
		ctx.drop()
		return Composition{Left: v, Right: vv}
	case Projection:
		ctx.drop()
		return Composition{Left: v, Right: vv}
	case ProjectionOne:
		ctx.drop()
		return Composition{Left: v, Right: vv}
	}
	res := ctx.variadics[v].Func(ctx, args)
	ctx.drop()
	return res
}

func (ctx *Context) applyNVariadic(v Variadic, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		if n == 2 && args[1] != nil {
			arg := args[1]
			ctx.dropN(n)
			return ProjectionOne{Fun: v, Arg: arg}
		}
		return Projection{Fun: v, Args: ctx.popN(n)}
	}
	if n == 2 && !ctx.variadics[v].Adverb {
		switch arg := args[0].(type) {
		case Composition:
			left := ProjectionOne{Fun: v, Arg: args[1]}
			ctx.dropN(2)
			return Composition{Left: left, Right: arg}
		case Projection:
			left := ProjectionOne{Fun: v, Arg: args[1]}
			ctx.dropN(2)
			return Composition{Left: left, Right: arg}
		case ProjectionOne:
			left := ProjectionOne{Fun: v, Arg: args[1]}
			ctx.dropN(2)
			return Composition{Left: left, Right: arg}
		}
	}
	res := ctx.variadics[v].Func(ctx, args)
	ctx.dropN(n)
	return res
}

func (ctx *Context) applyProjection(p Projection, n int) V {
	args := ctx.peekN(n)
	nNils := countNils(p.Args)
	switch {
	case len(args) > nNils:
		return errs("too many arguments")
	case len(args) == nNils:
		n := 0
		for _, arg := range p.Args {
			switch {
			case arg != nil:
				ctx.push(arg)
			default:
				ctx.push(args[n])
				n++
			}
		}
		res := ctx.applyN(p.Fun, len(p.Args))
		ctx.dropN(len(p.Args))
		return res
	default:
		vargs := cloneArgs(p.Args)
		n := 1
		for i := len(vargs) - 1; i >= 0; i-- {
			if vargs[i] == nil {
				if n > len(args) {
					break
				}
				vargs[i] = args[len(args)-n]
				n++
			}
		}
		ctx.dropN(n)
		return Projection{Fun: p, Args: vargs}
	}
}

func (ctx *Context) applyLambda(id Lambda, n int) V {
	if ctx.callDepth > maxCallDepth {
		return errs("lambda: exceeded maximum call depth")
	}
	lc := ctx.lambdas[int(id)]
	if lc.Rank < n {
		return errf("lambda: too many arguments: got %d, expected %d", n, lc.Rank)
	} else if lc.Rank > n {
		if lc.Rank == 2 && n == 1 {
			return ProjectionOne{Fun: id, Arg: ctx.pop()}
		}
		return Projection{Fun: id, Args: ctx.popN(n)}
	}
	nVars := len(lc.Names) - lc.Rank
	olen := len(ctx.stack)
	for i := 0; i < nVars; i++ {
		ctx.push(nil)
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
		return errV(err.Error())
	}
	var res V
	switch len(ctx.stack) {
	case olen + nVars:
	case olen + nVars + 1:
		res = ctx.stack[len(ctx.stack)-1]
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
	return res
}

func (x AV) atIndices(y AI) V {
	res := make(AV, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		res[i] = x[yi]
	}
	return canonical(res)
}

func (x AB) atIndices(y AI) V {
	res := make(AB, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		res[i] = x[yi]
	}
	return res
}

func (x AI) atIndices(y AI) V {
	res := make(AI, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		res[i] = x[yi]
	}
	return res
}

func (x AF) atIndices(y AI) V {
	res := make(AF, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		res[i] = x[yi]
	}
	return res
}

func (x AS) atIndices(y AI) V {
	res := make(AS, len(y))
	xlen := x.Len()
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		res[i] = x[yi]
	}
	return res
}

// set changes x at i with y (in place).
func (x AV) set(i int, y V) {
	x[i] = y
}

// set changes x at i with y (in place).
func (x AB) set(i int, y V) {
	x[i] = y.(I) == 1
}

// set changes x at i with y (in place).
func (x AI) set(i int, y V) {
	x[i] = int(y.(I))
}

// set changes x at i with y (in place).
func (x AF) set(i int, y V) {
	x[i] = float64(y.(F))
}

// set changes x at i with y (in place).
func (x AS) set(i int, y V) {
	x[i] = string(y.(S))
}

// setIndices x at y with z (in place).
func (x AV) setIndices(y AI, z V) error {
	az := z.(array)
	for i, yi := range y {
		if yi < 0 {
			yi += len(x)
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		x[yi] = az.at(i)
	}
	return nil
}

// setIndices x at y with z (in place).
func (x AI) setIndices(y AI, z V) error {
	az := z.(AI)
	for i, yi := range y {
		if yi < 0 {
			yi += len(x)
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		x[yi] = az[i]
	}
	return nil
}

// setIndices x at y with z (in place).
func (x AF) setIndices(y AI, z V) error {
	az := z.(AF)
	for i, yi := range y {
		if yi < 0 {
			yi += len(x)
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		x[yi] = az[i]
	}
	return nil
}

// setIndices x at y with z (in place).
func (x AB) setIndices(y AI, z V) error {
	az := z.(AB)
	for i, yi := range y {
		if yi < 0 {
			yi += len(x)
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		x[yi] = az[i]
	}
	return nil
}

// setIndices x at y with z (in place).
func (x AS) setIndices(y AI, z V) error {
	az := z.(AS)
	for i, yi := range y {
		if yi < 0 {
			yi += len(x)
		}
		if yi < 0 || yi >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", yi, len(x))
		}
		x[yi] = az[i]
	}
	return nil
}
