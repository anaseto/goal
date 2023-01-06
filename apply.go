package goal

//import "fmt"

// Apply calls a value with a single argument.
func (ctx *Context) Apply(x, y V) V {
	ctx.push(y)
	r := ctx.applyN(x, 1)
	ctx.drop()
	return r
}

// Apply2 calls a value with two arguments.
func (ctx *Context) Apply2(x, y, z V) V {
	ctx.push(z)
	ctx.push(y)
	r := ctx.applyN(x, 2)
	ctx.drop()
	return r
}

// ApplyN calls a value with one or more arguments. The arguments should be
// provided in reverse order, given the stack-based right to left semantics
// used by the language.
func (ctx *Context) ApplyN(x V, args []V) V {
	if len(args) == 0 {
		panic("ApplyN: len(args) should be > 0")
	}
	ctx.pushArgs(args)
	r := ctx.applyN(x, len(args))
	ctx.drop()
	return r
}

// applyN applies x with the top n arguments in the stack. It consumes only n-1
// arguments, but does not replace top with the result.
func (ctx *Context) applyN(x V, n int) V {
	//slen := len(ctx.stack)
	//defer func() {
	//if len(ctx.stack)+n-1 != slen {
	//panic(fmt.Sprintf("bad stack len: %d vs %d (x: %s, n: %d)", len(ctx.stack)+n-1, slen, x.Sprint(ctx), n))
	//}
	//}()
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
	case *derivedVerb:
		ctx.push(xv.Arg)
		args := ctx.peekN(n + 1)
		if hasNil(args) {
			args := cloneArgs(args[:len(args)-1])
			ctx.dropN(n)
			return NewV(&projection{Fun: x, Args: args})
		}
		if n > 1 {
			ctx.swap()
		}
		r := ctx.variadics[xv.Fun](ctx, args)
		ctx.dropN(n)
		return r
	case *projectionFirst:
		if n > 1 {
			ctx.dropN(n - 1)
			return Panicf("too many arguments: got %d, expected 1", n)
		}
		ctx.push(xv.Arg)
		return ctx.applyN(xv.Fun, 2)
	case *projectionMonad:
		if n > 1 {
			ctx.dropN(n - 1)
			return Panicf("too many arguments: got %d, expected 1", n)
		}
		return ctx.applyN(xv.Fun, 1)
	case *projection:
		return ctx.applyProjection(xv, n)
	case S:
		switch n {
		case 1:
			return applyS(xv, ctx.top())
		case 2:
			args := ctx.peekN(n)
			r := applyS2(xv, args[1], args[0])
			ctx.dropN(n - 1)
			return r
		default:
			ctx.dropN(n - 1)
			return Panicf("string got too many arguments")
		}
	case array:
		switch n {
		case 1:
			return ctx.applyArray(x, ctx.top())
		default:
			args := ctx.peekN(n)
			r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
			ctx.dropN(n - 1)
			return r
		}
	case *rx:
		switch n {
		case 1:
			return applyRx(xv, ctx.top())
		case 2:
			args := ctx.peekN(2)
			r := applyRx2(xv, args[1], args[0])
			ctx.drop()
			return r
		default:
			ctx.dropN(n - 1)
			return Panicf("regexp got too many arguments")
		}
	case stringReplacer:
		if n > 1 {
			ctx.dropN(n - 1)
			return Panicf("substitution got too many arguments")
		}
		return ctx.replace(xv, ctx.top())
	default:
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return Panicf("type %s cannot be applied", x.Type())
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
			return Panicf("x[y] : out of bounds index: %d", i)
		}
		return xv.at(int(i))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("x[y] : non-integer index (%g)", y.F())
		}
		i := int64(y.F())
		if i < 0 {
			i = int64(xv.Len()) + i
		}
		if i < 0 || i >= int64(xv.Len()) {
			return Panicf("x[y] : out of bounds index: %d", i)
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
		return Canonical(NewAV(r))
	case array:
		iy := toIndices(y)
		if iy.IsPanic() {
			return Panicf("x[y] : %v", iy.value)
		}
		r := xv.atIndices(iy.value.(*AI).Slice)
		return r
	default:
		return Panicf("x[y] : y non-integer (%s)", y.Type())
	}
}

func (ctx *Context) applyArrayArgs(x V, arg V, args []V) V {
	xv := x.value.(array)
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
		return Canonical(NewAV(r))
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
		return Canonical(NewAV(r))
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
		return NewV(&projectionMonad{Fun: newVariadic(v)})
	}
	return ctx.variadics[v](ctx, args)
}

func (ctx *Context) apply2Variadic(v variadic) V {
	args := ctx.peekN(2)
	if args[0].kind == valNil {
		if args[1].kind != valNil {
			arg := args[1]
			ctx.drop()
			return NewV(&projectionFirst{Fun: newVariadic(v), Arg: arg})
		}
		args := cloneArgs(args)
		ctx.drop()
		return NewV(&projection{Fun: newVariadic(v), Args: args})
	}
	if args[1].kind == valNil {
		args := cloneArgs(args)
		ctx.drop()
		return NewV(&projection{Fun: newVariadic(v), Args: args})
	}
	r := ctx.variadics[v](ctx, args)
	ctx.drop()
	return r
}

func (ctx *Context) applyNVariadic(v variadic, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		args := cloneArgs(args)
		ctx.dropN(n - 1)
		return NewV(&projection{Fun: newVariadic(v), Args: args})
	}
	r := ctx.variadics[v](ctx, args)
	ctx.dropN(n - 1)
	return r
}

func (ctx *Context) applyProjection(p *projection, n int) V {
	args := ctx.peekN(n)
	nNils := countNils(p.Args)
	switch {
	case len(args) > nNils:
		if n > 1 {
			ctx.dropN(n - 1)
		}
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
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return NewV(&projection{Fun: NewV(p), Args: vargs})
	}
}

func (ctx *Context) applyLambda(id lambda, n int) V {
	if ctx.callDepth > maxCallDepth {
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return panics("lambda: exceeded maximum call depth")
	}
	lc := ctx.lambdas[int(id)]
	if lc.Rank < n {
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return Panicf("lambda: too many arguments: got %d, expected %d", n, lc.Rank)
	}
	args := ctx.peekN(n)
	if lc.Rank > n || hasNil(args) {
		if n == 1 {
			if args[0].kind == valNil {
				return NewV(&projectionMonad{Fun: newLambda(id)})
			}
			return NewV(&projectionFirst{Fun: newLambda(id), Arg: ctx.top()})
		}
		if n == 2 && args[1].kind != valNil && args[0].kind == valNil {
			x := args[1]
			ctx.drop() // drop nil
			return NewV(&projectionFirst{Fun: newLambda(id), Arg: x})
		}
		args := cloneArgs(args)
		ctx.dropN(n - 1)
		return NewV(&projection{Fun: newLambda(id), Args: args})
	}
	unusedFirst := false
	for _, i := range lc.UnusedArgs {
		if v := args[i]; v.kind == valBoxed {
			v.rcdecrRefCounter()
			v.value = nil
			if i == 0 {
				unusedFirst = true
			}
		}
	}
	nVars := lc.nVars
	for i := 0; i < nVars; i++ {
		ctx.stack = append(ctx.stack, V{})
	}
	oframeIdx := ctx.frameIdx
	ctx.frameIdx = int32(len(ctx.stack) - 1)

	olambda := ctx.lambda
	ctx.lambda = int(id)
	ctx.callDepth++
	ip, err := ctx.execute(lc.Body)
	ctx.callDepth--
	ctx.lambda = olambda
	ctx.frameIdx = oframeIdx

	for _, i := range lc.UsedArgs {
		if v := args[i]; v.kind == valBoxed {
			if i == 0 {
				v.IncrRC()
				continue
			}
			v.value = nil
		}
	}
	var r V
	switch {
	case err != nil:
		ctx.updateErrPos(ip, lc)
		r = panics(err.Error())
	default:
		r = ctx.stack[len(ctx.stack)-1]
	}
	ctx.drop()
	if nVars > 0 {
		ctx.dropNnoRC(nVars)
	}
	if n > 1 {
		ctx.stack = ctx.stack[:len(ctx.stack)-n+1]
	}
	if unusedFirst {
		ctx.stack[len(ctx.stack)-1].IncrRC()
	}
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
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return Canonical(NewAV(r))
}

func (x *AB) atIndices(y []int64) V {
	r := make([]bool, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
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
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
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
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
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
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAS(r)
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	x.Slice[i] = y
}

// set changes x at i with y (in place).
func (x *AB) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n != 0
	} else {
		x.Slice[i] = y.F() != 0
	}
}

// set changes x at i with y (in place).
func (x *AI) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n
	} else {
		x.Slice[i] = int64(y.F())
	}
}

// set changes x at i with y (in place).
func (x *AF) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = float64(y.I())
	} else {
		x.Slice[i] = y.F()
	}
}

// set changes x at i with y (in place).
func (x *AS) set(i int, y V) {
	x.Slice[i] = string(y.value.(S))
}
