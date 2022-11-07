package main

func (ctx *Context) ApplyN(v V, n int) V {
	switch v := v.(type) {
	case Lambda:
		return ctx.applyLambda(v, n)
	case Variadic:
		args := ctx.peekN(n)
		if hasNil(args) {
			return Projection{Fun: v, Args: ctx.popN(n)}
		}
		res := builtins[v].Func(ctx, args)
		ctx.dropN(n)
		return res
	case Projection:
		return ctx.applyProjection(v, n)
	case Array:
		args := ctx.peekN(n)
		switch n {
		case 1:
			indices := toIndices(args[0])
			if indices == nil {
				return errs("not an integer array")
			}
			res := v.Apply(indices)
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
		res := ctx.ApplyN(v.Fun, len(v.Args))
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
	if lc.Arity < n {
		return errf("too many arguments: got %d, expected %d", n, lc.Arity)
	} else if lc.Arity > n {
		return Projection{Fun: id, Args: ctx.popN(n)}
	}
	olen := len(ctx.stack)
	oframeIdx := ctx.frameIdx
	ctx.frameIdx = int32(olen - n)

	ctx.callDepth++
	err := ctx.execute(lc.Body)
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

func (x AV) Apply(y AI) V {
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

func (x AB) Apply(y AI) V {
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

func (x AI) Apply(y AI) V {
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

func (x AF) Apply(y AI) V {
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

func (x AS) Apply(y AI) V {
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
