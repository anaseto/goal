package main

//import "fmt"

func (ctx *Context) execute(ops []opcode) error {
	for ip := 0; ip < len(ops); {
		op := ops[ip]
		//fmt.Printf("%s %d\n", op, ctx.sp)
		ip++
		switch op {
		case opNop:
		case opConst:
			ctx.push(ctx.constants[ops[ip]])
			ip++
		case opGlobal:
			ctx.push(ctx.globals[ops[ip]])
			ip++
		case opLocal:
			ctx.push(ctx.stack[ctx.frameIdx+int32(ops[ip])])
			ip++
		case opAssignGlobal:
			ctx.globals[ops[ip]] = ctx.top()
			ip++
		case opAssignLocal:
			ctx.stack[ctx.frameIdx+int32(ops[ip])] = ctx.top()
			ip++
		case opAdverb:
			ctx.push(Adverb(ops[ip]))
			ip++
		case opVariadic:
			ctx.push(Variadic(ops[ip]))
			ip++
		case opLambda:
			ctx.push(Lambda(ops[ip]))
			ip++
		case opApply:
			err := ctx.applyN(1)
			if err != nil {
				return err
			}
		case opApply2:
			err := ctx.applyN(2)
			if err != nil {
				return err
			}
		case opApplyN:
			err := ctx.applyN(int(ops[ip+1]))
			if err != nil {
				return err
			}
			ip++
		case opDrop:
			ctx.stack = ctx.stack[:len(ctx.stack)-1]
		}
	}
	return nil
}

func (ctx *Context) applyN(n int) error {
	v := ctx.pop()
	if id, ok := v.(Lambda); ok {
		return ctx.applyLambda(id, n)
	}
	args := ctx.popN(n)
	//fmt.Printf("args %d: %v\n", n, args)
	res := ctx.ApplyN(v, args)
	err, ok := res.(error)
	if ok {
		//fmt.Printf("applyN %d stack length %d, sp %d\n", n, len(ctx.stack), ctx.sp)
		return err
	}
	ctx.push(res)
	return nil
}

const maxCallDepth = 10000

func (ctx *Context) applyLambda(id Lambda, n int) error {
	if ctx.callDepth > maxCallDepth {
		return errs("exceeded maximum call depth")
	}
	lc := ctx.prog.Lambdas[int(id)]
	if lc.Arity < n {
		return errf("too many arguments: got %d, expected %d", n, lc.Arity)
	} else if lc.Arity > n {
		args := ctx.popN(n)
		ctx.push(Projection{Fun: id, Args: cloneAV(args)})
		return nil
	}
	olen := len(ctx.stack)
	oframeIdx := ctx.frameIdx
	ctx.frameIdx = int32(olen - n)

	ctx.callDepth++
	err := ctx.execute(lc.Body)
	ctx.callDepth--

	if err != nil {
		return err
	}
	var res V
	switch len(ctx.stack) {
	case olen:
		res = nil
	case olen + 1:
		res = ctx.stack[len(ctx.stack)-1]
	default:
		return errf("bad sp %d vs osp %d", len(ctx.stack), olen)
	}
	ctx.dropN(n)
	ctx.frameIdx = oframeIdx
	ctx.push(res)
	return nil
}

func (ctx *Context) push(v V) {
	ctx.stack = append(ctx.stack, v)
}

func (ctx *Context) pop() V {
	arg := ctx.stack[len(ctx.stack)-1]
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
	return arg
}

func (ctx *Context) top() V {
	return ctx.stack[len(ctx.stack)-1]
}

func (ctx *Context) popN(n int) []V {
	args := ctx.stack[len(ctx.stack)-n:]
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
	return args
}

func (ctx *Context) dropN(n int) {
	for i := range ctx.stack[len(ctx.stack)-n:] {
		ctx.stack[i] = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}
