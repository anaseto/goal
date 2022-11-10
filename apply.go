package goal

// applyN applies v with the top n arguments in the stack. It consumes the
// arguments, but does not push the result, returing it instead.
func (ctx *Context) applyN(v V, n int) V {
	switch v := v.(type) {
	case Lambda:
		return ctx.applyLambda(v, n)
	case Variadic:
		if n == 1 {
			return ctx.applyVariadic(v)
		}
		return ctx.applyNVariadic(v, n)
	case DerivedVerb:
		ctx.push(v.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			return Projection{Fun: v, Args: ctx.popN(n + 1)}
		}
		res := builtins[v.Fun].Func(ctx, args)
		ctx.dropN(n + 1)
		return res
	case ProjectionOne:
		if n > 1 {
			return errf("too many arguments: got %d, expected 1", n)
		}
		arg := ctx.top()
		if arg == nil {
			ctx.drop()
			return v
		}
		ctx.push(v.Arg)
		return ctx.applyN(v.Fun, 2)
	case Projection:
		return ctx.applyProjection(v, n)
	case Composition:
		res := ctx.applyN(v.Right, n)
		_, ok := res.(error)
		if ok {
			return res
		}
		ctx.push(res)
		return ctx.applyN(v.Left, 1)
	case Array:
		args := ctx.peekN(n)
		switch n {
		case 1:
			if args[0] == nil {
				ctx.drop()
				return v
			}
			indices := toIndices(args[0])
			if indices == nil {
				return errs("not an integer array")
			}
			res := v.Select(indices)
			ctx.drop()
			return res
		default:
			ctx.dropN(n)
			return errf("NYI: deep index %d", n)
		}
	default:
		return errf("type %s cannot be applied", v.Type())
	}
}

func (ctx *Context) applyVariadic(v Variadic) V {
	args := ctx.peek()
	vv := args[0]
	if builtins[v].Adverb {
		ctx.drop()
		return DerivedVerb{Fun: v, Arg: vv}
	}
	if vv == nil {
		ctx.drop()
		return Projection{Fun: v, Args: []V{vv}}
	}
	switch vv := vv.(type) {
	case Variadic:
		return Composition{Left: v, Right: vv}
	default:
		if vv == nil {
			return Projection{Fun: v, Args: []V{vv}}
		}
		res := builtins[v].Func(ctx, args)
		ctx.drop()
		return res
	}
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
	if n == 2 && !builtins[v].Adverb {
		switch arg := args[1].(type) {
		case Lambda:
		case Function:
			res := Composition{
				Left:  ProjectionOne{Fun: v, Arg: args[0]},
				Right: arg,
			}
			ctx.dropN(2)
			return res
		}
	}
	res := builtins[v].Func(ctx, args)
	ctx.dropN(n)
	return res
}

func (ctx *Context) applyProjection(v Projection, n int) V {
	args := ctx.peekN(n)
	nNils := countNils(v.Args)
	switch {
	case len(args) > nNils:
		return errs("too many arguments")
	case len(args) == nNils:
		n := 0
		for _, v := range v.Args {
			switch {
			case v != nil:
				ctx.push(v)
			default:
				ctx.push(args[n])
				n++
			}
		}
		res := ctx.applyN(v.Fun, len(v.Args))
		ctx.dropN(len(v.Args))
		ctx.dropN(n)
		return res
	default:
		vargs := cloneArgs(v.Args)
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
		return Projection{Fun: v, Args: vargs}
	}
}

func (ctx *Context) applyLambda(id Lambda, n int) V {
	if ctx.callDepth > maxCallDepth {
		return errs("exceeded maximum call depth")
	}
	lc := ctx.prog.Lambdas[int(id)]
	if lc.Rank < n {
		return errf("too many arguments: got %d, expected %d", n, lc.Rank)
	} else if lc.Rank > n {
		if lc.Rank == 2 && n == 1 {
			return ProjectionOne{Fun: id, Arg: ctx.pop()}
		}
		return Projection{Fun: id, Args: ctx.popN(n)}
	}
	olen := len(ctx.stack)
	oframeIdx := ctx.frameIdx
	ctx.frameIdx = int32(olen - n)

	ctx.callDepth++
	_, err := ctx.execute(lc.Body)
	ctx.callDepth--

	if err != nil {
		return errf("lambda execute: %v", err)
	}
	var res V
	switch len(ctx.stack) {
	case olen:
	case olen + 1:
		res = ctx.stack[len(ctx.stack)-1]
	default:
		return errf("bad sp %d vs osp %d", len(ctx.stack), olen)
	}
	ctx.dropN(n)
	ctx.frameIdx = oframeIdx
	return res
}

func (x AV) Select(y AI) V {
	res := make(AV, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AB) Select(y AI) V {
	res := make(AB, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AI) Select(y AI) V {
	res := make(AI, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AF) Select(y AI) V {
	res := make(AF, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AS) Select(y AI) V {
	res := make(AS, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}
