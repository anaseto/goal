package main

func (ctx *Context) ApplyN(v V, args []V) V {
	switch v := v.(type) {
	case Variadic:
		if hasNil(args) {
			return Projection{Fun: v, Args: cloneAV(args)}
		}
		return builtins[v].Func(ctx, args)
	case Projection:
		if len(args) > countNils(v.Args) {
			return errs("too many arguments")
		}
		vargs := cloneAV(v.Args)
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
		for _, arg := range vargs {
			if arg == nil {
				return v
			}
		}
		return ctx.ApplyN(v.Fun, vargs)
	case Array:
		switch len(args) {
		case 1:
			indices := toIndices(args[0])
			if indices == nil {
				return errs("not an integer array")
			}
			return v.Apply(indices)
		default:
			return errf("NYI: deep index %d", len(args))
		}
	case Lambda:
		ctx.stack = append(ctx.stack, args...)
		err := ctx.applyLambda(v, len(args))
		if err != nil {
			return errs(err.Error())
		}
		return ctx.pop()
	default:
		return errf("type %s cannot be applied", v.Type())
	}
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
