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
			ctx.push(ctx.frame[ops[ip]])
			ip++
		case opAssignGlobal:
			ctx.globals[ops[ip]] = ctx.top()
			ip++
		case opAssignLocal:
			ctx.frame[ops[ip]] = ctx.top()
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
			ctx.sp--
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

const maxCallDepth = 100000

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
	osp := ctx.sp
	oframe := ctx.frame
	ctx.frame = ctx.stack[ctx.sp-n:]

	ctx.callDepth++
	err := ctx.execute(lc.Body)
	ctx.callDepth--

	ctx.frame = oframe

	if err != nil {
		return err
	}
	if ctx.sp == osp {
		ctx.push(nil)
	}
	return nil
}

func (ctx *Context) push(v V) {
	if ctx.sp >= len(ctx.stack) {
		ctx.stack = append(ctx.stack, nil)
	}
	ctx.stack[ctx.sp] = v
	ctx.sp++
}

func (ctx *Context) pop() V {
	ctx.sp--
	return ctx.stack[ctx.sp]
}

func (ctx *Context) top() V {
	return ctx.stack[ctx.sp-1]
}

func (ctx *Context) pop2() (V, V) {
	ctx.sp -= 2
	return ctx.stack[ctx.sp], ctx.stack[ctx.sp+1]
}

func (ctx *Context) popN(n int) []V {
	values := make([]V, n)
	copy(values, ctx.stack[ctx.sp-n:])
	ctx.sp -= n
	return values
}
