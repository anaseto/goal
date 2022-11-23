package goal

// Apply calls a value with a single argument.
func (ctx *Context) Apply(v, x V) V {
	ctx.push(x)
	return ctx.applyN(v, 1)
}

// ApplyN calls a value with one or more arguments. The arguments should be
// provided in reverse order, given the stack-based right to left semantics
// used by the language.
func (ctx *Context) ApplyN(v V, args []V) V {
	if len(args) == 0 {
		panic("ApplyArgs: len(args) should be > 0")
	}
	ctx.pushArgs(args)
	return ctx.applyN(v, len(args))
}

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
		res := ctx.variadics[v.Fun].Func(ctx, args)
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
	case S:
		switch n {
		case 1:
			return applyS(v, ctx.pop())
		case 2:
			args := ctx.peekN(n)
			res := applyS2(v, args[1], args[0])
			ctx.dropN(n)
			return res
		default:
			return errf("too many arguments")
		}
	case array:
		switch n {
		case 1:
			return ctx.applyArray(v, ctx.pop())
		default:
			args := ctx.peekN(n)
			res := ctx.applyArrayArgs(v, args[len(args)-1], args[:len(args)-1])
			ctx.dropN(n)
			return res
		}
	default:
		return errf("type %s cannot be applied", v.Type())
	}
}

func applyS(s S, x V) V {
	switch x := x.(type) {
	case I:
		if x < 0 {
			x += I(len(s))
		}
		if x < 0 || x > I(len(s)) {
			return errf("s[i] : i out of bounds index (%d)", x)
		}
		return s[x:]
	case F:
		if !isI(x) {
			return errf("s[x] : x non-integer (%g)", x)
		}
		return applyS(s, x)
	case AB:
		return applyS(s, fromABtoAI(x))
	case AI:
		res := make(AS, x.Len())
		for i, n := range x {
			if n < 0 {
				n += len(s)
			}
			if n < 0 || n > len(s) {
				return errf("s[i] : i out of bounds index (%d)", n)
			}
			res[i] = string(s[n:])
		}
		return res
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS(s, z)
	case AV:
		res := make(AV, x.Len())
		for i, v := range x {
			res[i] = applyS(s, v)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	default:
		return errf("s[x] : x non-integer (%s)", x.Type())
	}
}

func applyS2(s S, x V, y V) V {
	var l int
	switch y := y.(type) {
	case I:
		if y < 0 {
			return errf("s[x;y] : y negative (%d)", y)
		}
		l = int(y)
	case F:
		if !isI(y) {
			return errf("s[x;y] : y non-integer (%g)", y)
		}
		l = int(y)
	case AI:
	case AB:
		if Length(x) != y.Len() {
		}
		return applyS2(s, x, fromABtoAI(y))
	case AF:
		z := toAI(y)
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS2(s, x, z)
	default:
		return errType("s[x;y]", "y", y)
	}
	switch x := x.(type) {
	case I:
		if x < 0 {
			x += I(len(s))
		}
		if x < 0 || x > I(len(s)) {
			return errf("s[i;y] : i out of bounds index (%d)", x)
		}
		if _, ok := y.(AI); ok {
			return errf("s[x;y] : x is an atom but y is an array")
		}
		if int(x)+l > len(s) {
			l = len(s) - int(x)
		}
		return s[x : int(x)+l]
	case F:
		if !isI(x) {
			return errf("s[x;y] : x non-integer (%g)", x)
		}
		return applyS2(s, x, y)
	case AB:
		return applyS2(s, fromABtoAI(x), y)
	case AI:
		res := make(AS, x.Len())
		if z, ok := y.(AI); ok {
			if z.Len() != x.Len() {
				return errf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					x.Len(), z.Len())
			}
			for i, n := range x {
				if n < 0 {
					n += len(s)
				}
				if n < 0 || n > len(s) {
					return errf("s[i;y] : i out of bounds index (%d)", n)
				}
				l := z[i]
				if n+l > len(s) {
					l = len(s) - n
				}
				res[i] = string(s[n : n+l])
			}
			return res
		}
		for i, n := range x {
			if n < 0 {
				n += len(s)
			}
			if n < 0 || n > len(s) {
				return errf("s[i;y] : i out of bounds index (%d)", n)
			}
			l := l
			if n+l > len(s) {
				l = len(s) - n
			}
			res[i] = string(s[n : n+l])
		}
		return res
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return err
		}
		return applyS2(s, z, y)
	case AV:
		res := make(AV, x.Len())
		for i, v := range x {
			res[i] = applyS2(s, v, y)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	default:
		return errf("s[x;y] : x non-integer (%s)", x.Type())
	}
}

// applyArray applies an array to a value.
func (ctx *Context) applyArray(a array, x V) V {
	if x == nil {
		return a
	}
	switch z := x.(type) {
	case F:
		if !isI(z) {
			return errf("a[x] : non-integer index: %g", z)
		}
		i := int(z)
		if i < 0 {
			i = a.Len() + i
		}
		if i < 0 || i >= a.Len() {
			return errf("a[x] : out of bounds index: %d", i)
		}
		return a.at(i)
	case I:
		i := int(z)
		if i < 0 {
			i = a.Len() + i
		}
		if i < 0 || i >= a.Len() {
			return errf("a[x] : out of bounds index: %d", i)
		}
		return a.at(i)
	case AV:
		res := make(AV, z.Len())
		for i, v := range z {
			res[i] = ctx.applyArray(a, v)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	case array:
		indices := toIndices(x, a.Len())
		if err, ok := indices.(errV); ok {
			return err
		}
		res := a.atIndices(indices.(AI))
		return res
	default:
		return errf("a[x] : x non-array non-integer")
	}
}

func (ctx *Context) applyArrayArgs(v array, arg V, args []V) V {
	// TODO: annotate error with depth?
	if len(args) == 0 {
		return ctx.applyArray(v, arg)
	}
	if arg == nil {
		res := make(AV, v.Len())
		for i := 0; i < len(res); i++ {
			res[i] = ctx.ApplyN(v.at(i), args)
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
			res[i] = ctx.applyArrayArgs(v, arg.at(i), args)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return canonical(res)
	default:
		res := ctx.applyArray(v, arg)
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
	for i := range res {
		idx := y[i]
		if idx < 0 {
			idx += xlen
		}
		if idx < 0 || idx >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", y[i], len(x))
		}
		res[i] = x[idx]
	}
	return canonical(res)
}

func (x AB) atIndices(y AI) V {
	res := make(AB, len(y))
	xlen := x.Len()
	for i := range res {
		idx := y[i]
		if idx < 0 {
			idx += xlen
		}
		if idx < 0 || idx >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", y[i], len(x))
		}
		res[i] = x[idx]
	}
	return res
}

func (x AI) atIndices(y AI) V {
	res := make(AI, len(y))
	xlen := x.Len()
	for i := range res {
		idx := y[i]
		if idx < 0 {
			idx += xlen
		}
		if idx < 0 || idx >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", y[i], len(x))
		}
		res[i] = x[idx]
	}
	return res
}

func (x AF) atIndices(y AI) V {
	res := make(AF, len(y))
	xlen := x.Len()
	for i := range res {
		idx := y[i]
		if idx < 0 {
			idx += xlen
		}
		if idx < 0 || idx >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", y[i], len(x))
		}
		res[i] = x[idx]
	}
	return res
}

func (x AS) atIndices(y AI) V {
	res := make(AS, len(y))
	xlen := x.Len()
	for i := range res {
		idx := y[i]
		if idx < 0 {
			idx += xlen
		}
		if idx < 0 || idx >= len(x) {
			return errf("x[y] : index out of bounds: %d (length %d)", y[i], len(x))
		}
		res[i] = x[idx]
	}
	return res
}
