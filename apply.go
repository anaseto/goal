package goal

//import "fmt"

// Apply calls a value with a single argument.
func (ctx *Context) Apply(x, y V) V {
	ctx.push(y)
	r := x.applyN(ctx, 1)
	ctx.drop()
	return r
}

// Apply2 calls a value with two arguments.
func (ctx *Context) Apply2(x, y, z V) V {
	ctx.push(z)
	ctx.push(y)
	r := x.applyN(ctx, 2)
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
	r := x.applyN(ctx, len(args))
	ctx.drop()
	return r
}

// applicable represents boxed values than can be applied.
type applicable interface {
	Value

	// applyN applies the value with the top n arguments in the stack and
	// returns the result. It consumes only n-1 arguments, but does not
	// replace top with the result.
	applyN(*Context, int) V
}

// applyN applies x with the top n arguments in the stack and returns the
// result. It consumes only n-1 arguments, but does not replace top with the
// result.
func (x V) applyN(ctx *Context, n int) V {
	//slen := len(ctx.stack)
	//defer func() {
	//if len(ctx.stack)+n-1 != slen {
	//panic(fmt.Sprintf("bad stack len: %d vs %d (x: %s, n: %d)", len(ctx.stack)+n-1, slen, x.Sprint(ctx), n))
	//}
	//}()
	switch x.kind {
	case valLambda:
		return x.lambda().applyN(ctx, n)
	case valVariadic:
		switch n {
		case 1:
			return x.variadic().apply(ctx)
		case 2:
			return x.variadic().apply2(ctx)
		default:
			return x.variadic().applyN(ctx, n)
		}
	}
	switch xv := x.value.(type) {
	case applicable:
		return xv.applyN(ctx, n)
	default:
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return Panicf("type %s cannot be applied", x.Type())
	}
}

func (v variadic) apply(ctx *Context) V {
	args := ctx.peek()
	x := args[0]
	if x.kind == valNil {
		return NewV(&projectionMonad{Fun: newVariadic(v)})
	}
	r := ctx.variadics[v](ctx, args)
	r.InitRC()
	return r
}

func (v variadic) apply2(ctx *Context) V {
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
	r.InitRC()
	ctx.drop()
	return r
}

func (v variadic) applyN(ctx *Context, n int) V {
	args := ctx.peekN(n)
	if hasNil(args) {
		args := cloneArgs(args)
		ctx.dropN(n - 1)
		return NewV(&projection{Fun: newVariadic(v), Args: args})
	}
	r := ctx.variadics[v](ctx, args)
	r.InitRC()
	ctx.dropN(n - 1)
	return r
}

func (id lambda) applyN(ctx *Context, n int) V {
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
		if v := &args[i]; v.kind == valBoxed {
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
		if v := &args[i]; v.kind == valBoxed {
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

func (dv *derivedVerb) applyN(ctx *Context, n int) V {
	ctx.push(dv.Arg)
	args := ctx.peekN(n + 1)
	if hasNil(args) {
		args := cloneArgs(args[:len(args)-1])
		ctx.dropN(n)
		return NewV(&projection{Fun: NewV(dv), Args: args})
	}
	r := ctx.variadics[dv.Fun](ctx, args)
	r.InitRC()
	ctx.dropN(n)
	return r
}

func (p *projectionFirst) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("too many arguments: got %d, expected 1", n)
	}
	ctx.push(p.Arg)
	return p.Fun.applyN(ctx, 2)
}

func (p *projectionMonad) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("too many arguments: got %d, expected 1", n)
	}
	return p.Fun.applyN(ctx, 1)
}

func (p *projection) applyN(ctx *Context, n int) V {
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
		r := p.Fun.applyN(ctx, len(p.Args))
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

func (s S) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyS(s, ctx.top())
		r.InitRC()
		return r
	case 2:
		args := ctx.peekN(n)
		r := applyS2(s, args[1], args[0])
		r.InitRC()
		ctx.dropN(n - 1)
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("string got too many arguments")
	}
}

func (re *rx) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyRx(re, ctx.top())
		r.InitRC()
		return r
	case 2:
		args := ctx.peekN(2)
		r := applyRx2(re, args[1], args[0])
		r.InitRC()
		ctx.drop()
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("regexp got too many arguments")
	}
}

func (x *AB) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		return applyArray(x, ctx.top())
	default:
		ctx.dropN(n - 1)
		return Panicf("x[y] : out of depth")
	}
}

func (x *AI) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		return applyArray(x, ctx.top())
	default:
		ctx.dropN(n - 1)
		return Panicf("x[y] : out of depth")
	}
}

func (x *AF) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		return applyArray(x, ctx.top())
	default:
		ctx.dropN(n - 1)
		return Panicf("x[y] : out of depth")
	}
}

func (x *AS) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		return applyArray(x, ctx.top())
	default:
		args := ctx.peekN(n)
		r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
		r.InitRC()
		ctx.dropN(n - 1)
		return r
	}
}

func (x *AV) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		return applyArray(x, ctx.top())
	default:
		args := ctx.peekN(n)
		r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
		r.InitRC()
		ctx.dropN(n - 1)
		return r
	}
}

// applyArray applies an array to a value.
func applyArray(x array, y V) V {
	if y.kind == valNil {
		return NewV(x)
	}
	if y.IsI() {
		i := y.I()
		if i < 0 {
			i = int64(x.Len()) + i
		}
		if i < 0 || i >= int64(x.Len()) {
			return Panicf("x[y] : out of bounds index: %d", i)
		}
		return x.at(int(i))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("x[y] : non-integer index (%g)", y.F())
		}
		i := int64(y.F())
		if i < 0 {
			i = int64(x.Len()) + i
		}
		if i < 0 || i >= int64(x.Len()) {
			return Panicf("x[y] : out of bounds index: %d", i)
		}
		return x.at(int(i))
	}
	if isStar(y) {
		return NewV(x)
	}
	switch yv := y.value.(type) {
	case *AI:
		return x.atIndices(yv.Slice)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = applyArray(x, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewV(&AV{Slice: r, rc: x.RC()}))
	case array:
		iy := toIndices(y)
		if iy.IsPanic() {
			return Panicf("x[y] : %v", iy.value)
		}
		r := x.atIndices(iy.value.(*AI).Slice)
		return r
	default:
		return Panicf("x[y] : y non-integer (%s)", y.Type())
	}
}

func (ctx *Context) applyArrayArgs(x array, arg V, args []V) V {
	if len(args) == 0 {
		return applyArray(x, arg)
	}
	if arg.kind == valNil {
		r := make([]V, x.Len())
		for i := 0; i < len(r); i++ {
			r[i] = ctx.ApplyN(x.at(i), args)
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
		r := applyArray(x, arg)
		if r.IsPanic() {
			return r
		}
		return ctx.ApplyN(r, args)
	}
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
	return Canonical(NewV(&AV{Slice: r, rc: x.rc}))
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
	return NewV(&AB{Slice: r, rc: x.rc})
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
	return NewV(&AI{Slice: r, rc: x.rc})
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
	return NewV(&AF{Slice: r, rc: x.rc})
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
	return NewV(&AS{Slice: r, rc: x.rc})
}

func (r *nReplacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	nr := ctx.replace(r, ctx.top())
	nr.InitRC()
	return nr
}

func (r *replacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	nr := ctx.replace(r, ctx.top())
	nr.InitRC()
	return nr
}

func (r *rxReplacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	nr := ctx.replace(r, ctx.top())
	nr.InitRC()
	return nr
}
