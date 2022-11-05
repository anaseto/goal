package main

func (ctx *Context) execute(ops []opcode) error {
	for ip := 0; ip < len(ops); {
		op := ops[ip]
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
		case opMonad:
			ctx.push(Monad(ops[ip]))
			ip++
		case opDyad:
			ctx.push(Dyad(ops[ip]))
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
			err := ctx.apply()
			if err != nil {
				return err
			}
		case opApply2:
			err := ctx.apply2()
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

func (ctx *Context) apply() error {
	v := ctx.pop()
	x := ctx.pop()
	res := Apply(v, x)
	err, ok := res.(error)
	if ok {
		return err
	}
	ctx.push(res)
	return nil
}

func (ctx *Context) apply2() error {
	v := ctx.pop()
	w, x := ctx.pop2()
	res := Apply2(v, w, x)
	err, ok := res.(error)
	if ok {
		return err
	}
	ctx.push(res)
	return nil
}

func (ctx *Context) applyN(n int) error {
	//v := ctx.pop()
	//vals := ctx.popN(n)
	// TODO
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
	ctx.sp -= n
	values := make([]V, 0, n)
	for i := 0; i < n; i++ {
		values[i] = ctx.stack[ctx.sp+i]
	}
	return values
}
