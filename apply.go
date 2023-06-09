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
// provided in stack order, as in the right to left semantics used by the
// language: the first argument is the last element.
func (ctx *Context) ApplyN(x V, args []V) V {
	if len(args) == 0 {
		panic("ApplyN: len(args) should be > 0")
	}
	ctx.pushArgs(args)
	r := x.applyN(ctx, len(args))
	ctx.drop()
	return r
}

// callable represents boxed values than can be applied.
type callable interface {
	BV

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
	case valInt:
		return applyI(n, x.I(), ctx.top())
	case valFloat:
		if !isI(x.F()) {
			return Panicf("i@y : non-integer i (%g)", x.F())
		}
		return applyI(n, int64(x.F()), ctx.top())
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
	switch xv := x.bv.(type) {
	case callable:
		return xv.applyN(ctx, n)
	default:
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return Panicf("type \"%s\" is not callable", x.Type())
	}
}

func (v variadic) apply(ctx *Context) V {
	args := ctx.peek()
	x := args[0]
	if x.kind == valNil {
		return NewV(&projectionMonad{Fun: newVariadic(v)})
	}
	r := ctx.variadics[v](ctx, args)
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
			v.bv = nil
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
			v.bv = nil
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
	if unusedFirst && err == nil {
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
	ctx.dropN(n)
	return r
}

func (p *projectionFirst) applyN(ctx *Context, n int) V {
	ctx.push(p.Arg)
	return p.Fun.applyN(ctx, n+1)
}

func (p *projectionMonad) applyN(ctx *Context, n int) V {
	return p.Fun.applyN(ctx, n)
}

func (p *projection) applyN(ctx *Context, n int) V {
	args := ctx.peekN(n)
	nilN := countNils(p.Args)
	switch {
	case n >= nilN:
		for _, arg := range args[nilN:] {
			ctx.pushNoRC(arg)
		}
		nilc := 0
		for _, arg := range p.Args {
			switch {
			case arg.kind != valNil:
				ctx.push(arg)
			default:
				ctx.pushNoRC(args[nilc])
				nilc++
			}
		}
		// argument stack len: n+(#p.Args)+n-nilN
		r := p.Fun.applyN(ctx, len(p.Args)+n-nilN)
		ctx.drop()
		// argument stack len: n
		ctx.dropNnoRC(n - 1)
		// restore refcount of remaining last argument on the stack
		args[0].IncrRC()
		return r
	default:
		vargs := cloneArgs(p.Args)
		nilc := 1
		for i := len(vargs) - 1; i >= 0; i-- {
			if vargs[i].kind == valNil {
				if nilc > n {
					break
				}
				vargs[i] = args[n-nilc]
				nilc++
			}
		}
		if n > 1 {
			ctx.dropN(n - 1)
		}
		return NewV(&projection{Fun: p.Fun, Args: vargs})
	}
}

func (s S) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyS(s, ctx.top())
		return r
	case 2:
		args := ctx.peekN(n)
		r := applyS2(s, args[1], args[0])
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
		x := ctx.top()
		if x.kind == valNil {
			return NewV(&projectionMonad{Fun: NewV(re)})
		}
		r := applyRx(re, x)
		return r
	case 2:
		args := ctx.peekN(2)
		if args[0].kind == valNil {
			if args[1].kind != valNil {
				arg := args[1]
				ctx.drop()
				return NewV(&projectionFirst{Fun: NewV(re), Arg: arg})
			}
			args := cloneArgs(args)
			ctx.drop()
			return NewV(&projection{Fun: NewV(re), Args: args})
		}
		if args[1].kind == valNil {
			args := cloneArgs(args)
			ctx.drop()
			return NewV(&projection{Fun: NewV(re), Args: args})
		}
		r := applyRx2(re, args[1], args[0])
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
		r := applyArray(x, ctx.top())
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("X.y : out of depth indexing")
	}
}

func (x *AI) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyArray(x, ctx.top())
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("X.y : out of depth indexing")
	}
}

func (x *AF) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyArray(x, ctx.top())
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("X.y : out of depth indexing")
	}
}

func (x *AS) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyArray(x, ctx.top())
		return r
	default:
		ctx.dropN(n - 1)
		return Panicf("X.y : out of depth indexing")
	}
}

func (x *AV) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		r := applyArray(x, ctx.top())
		return r
	default:
		args := ctx.peekN(n)
		r := ctx.applyArrayArgs(x, args[len(args)-1], args[:len(args)-1])
		ctx.dropN(n - 1)
		return r
	}
}

func (d *D) applyN(ctx *Context, n int) V {
	switch n {
	case 1:
		y := ctx.top()
		r := applyDict(d, y)
		return r
	default:
		args := ctx.peekN(n)
		r := ctx.applyDictArgs(d, args[len(args)-1], args[:len(args)-1])
		ctx.dropN(n - 1)
		return r
	}
}

func applyDict(d *D, y V) V {
	if y.kind == valNil {
		return NewV(d.values)
	}
	dlen := d.keys.Len()
	ky := findArray(d.keys, y)
	if ky.IsI() {
		i := ky.I() // i >= 0
		if i >= int64(dlen) {
			return arrayProtoV(d.values)
		}
		return d.values.VAt(int(i))
	}
	r := atIv(d.values, ky)
	return r
}

func (ctx *Context) applyDictArgs(x *D, arg V, args []V) V {
	if len(args) == 0 {
		return applyDict(x, arg)
	}
	if arg.kind == valNil {
		return cdoN(x.keys.Len(), func(i int) V { return ctx.applyDictArgs(x, x.keys.VAt(i), args) })
	}
	switch argv := arg.bv.(type) {
	case Array:
		return cdoN(argv.Len(), func(i int) V { return ctx.applyDictArgs(x, argv.VAt(i), args) })
	default:
		r := applyDict(x, arg)
		// applyDict never panics
		switch rv := r.bv.(type) {
		case Array:
			return ctx.applyArrayArgs(rv, args[len(args)-1], args[:len(args)-1])
		case *D:
			return ctx.applyDictArgs(rv, args[len(args)-1], args[:len(args)-1])
		default:
			return panics("d.y : out of depth indexing")
		}
	}
}

// applyArray applies an array to a value.
func applyArray(x Array, y V) V {
	if y.IsI() {
		i := y.I()
		if i < 0 {
			i = int64(x.Len()) + i
		}
		if i < 0 || i >= int64(x.Len()) {
			return arrayProtoV(x)
		}
		return x.VAt(int(i))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("X@i : non-integer index (%g)", y.F())
		}
		return applyArray(x, NewI(int64(y.F())))
	}
	if y.kind == valNil || isStar(y) {
		return NewV(x)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return x.vAtAB(yv)
	case *AF:
		y = toAI(yv)
		if y.IsPanic() {
			return ppanic("X@i : ", y)
		}
		return applyArray(x, y)
	case *AI:
		return x.vAtAI(yv)
	case *AV:
		return cmapAV(yv, func(yi V) V { return applyArray(x, yi) })
	case *D:
		return newDictValues(yv.keys, applyArray(x, NewV(yv.values)))
	default:
		return panicType("X@i", "i", y)
	}
}

func (ctx *Context) applyArrayArgs(x Array, arg V, args []V) V {
	if len(args) == 0 {
		return applyArray(x, arg)
	}
	if arg.kind == valNil {
		return cdoN(x.Len(), func(i int) V { return ctx.applyArrayArgs(x, NewI(int64(i)), args) })
	}
	switch argv := arg.bv.(type) {
	case Array:
		return cdoN(argv.Len(), func(i int) V { return ctx.applyArrayArgs(x, argv.VAt(i), args) })
	default:
		r := applyArray(x, arg)
		if r.IsPanic() {
			return r
		}
		switch rv := r.bv.(type) {
		case Array:
			return ctx.applyArrayArgs(rv, args[len(args)-1], args[:len(args)-1])
		case *D:
			return ctx.applyDictArgs(rv, args[len(args)-1], args[:len(args)-1])
		default:
			return panics("X.y : out of depth indexing")
		}
	}
}

func (r *nReplacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	x := ctx.top()
	if x.kind == valNil {
		return NewV(&projectionMonad{Fun: NewV(r)})
	}
	nr := ctx.replace(r, x)
	return nr
}

func (r *replacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	x := ctx.top()
	if x.kind == valNil {
		return NewV(&projectionMonad{Fun: NewV(r)})
	}
	nr := ctx.replace(r, x)
	return nr
}

func (r *rxReplacer) applyN(ctx *Context, n int) V {
	if n > 1 {
		ctx.dropN(n - 1)
		return Panicf("substitution got too many arguments")
	}
	x := ctx.top()
	if x.kind == valNil {
		return NewV(&projectionMonad{Fun: NewV(r)})
	}
	nr := ctx.replace(r, x)
	return nr
}

func applyI(n int, i int64, y V) V {
	if n > 1 {
		return panicRank("i.y")
	}
	switch yv := y.bv.(type) {
	case *D:
		rk := takePadN(i, yv.keys)
		rv := takePadN(i, yv.values)
		return NewV(&D{
			keys:   rk.bv.(Array),
			values: rv.bv.(Array)})
	case Array:
		r := takePadN(i, yv)
		return r
	default:
		r := takePadN(i, toArray(y).bv.(Array))
		return r
	}
}
