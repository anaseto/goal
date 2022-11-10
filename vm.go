package main

//import "fmt"

func (ctx *Context) execute(ops []opcode) error {
	//fmt.PrIntf("ops: %v\n", ops)
	for ip := 0; ip < len(ops); {
		op := ops[ip]
		//fmt.Printf("op: %s\n", op)
		ip++
		switch op {
		case opNop:
		case opConst:
			ctx.push(ctx.constants[ops[ip]])
			ip++
		case opNil:
			ctx.push(nil)
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
		case opVariadic:
			ctx.push(Variadic(ops[ip]))
			ip++
		case opLambda:
			ctx.push(Lambda(ops[ip]))
			ip++
		case opApply:
			err := ctx.popApplyN(1)
			if err != nil {
				return err
			}
		case opApply2:
			err := ctx.popApplyN(2)
			if err != nil {
				return err
			}
		case opApplyN:
			err := ctx.popApplyN(int(ops[ip]))
			if err != nil {
				return err
			}
			ip++
		case opDrop:
			ctx.drop()
		}
		//fmt.Printf("stack: %v\n", ctx.stack)
	}
	return nil
}

func (ctx *Context) popApplyN(n int) error {
	v := ctx.pop()
	res := ctx.applyN(v, n)
	err, ok := res.(error)
	if ok {
		return err
	}
	ctx.push(res)
	return nil
}

const maxCallDepth = 10000

func (ctx *Context) push(v V) {
	ctx.stack = append(ctx.stack, v)
}

func (ctx *Context) pushArgs(args []V) {
	ctx.stack = append(ctx.stack, args...)
}

func (ctx *Context) pop() V {
	arg := ctx.stack[len(ctx.stack)-1]
	ctx.stack[len(ctx.stack)-1] = nil
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
	return arg
}

func (ctx *Context) top() V {
	return ctx.stack[len(ctx.stack)-1]
}

func (ctx *Context) popN(n int) []V {
	args := cloneArgs(ctx.stack[len(ctx.stack)-n:])
	for i := range ctx.stack[len(ctx.stack)-n:] {
		ctx.stack[i] = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
	return args
}

func (ctx *Context) peek() []V {
	return ctx.stack[len(ctx.stack)-1:]
}

func (ctx *Context) peekN(n int) []V {
	return ctx.stack[len(ctx.stack)-n:]
}

func (ctx *Context) drop() {
	ctx.stack[len(ctx.stack)-1] = nil
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
}

func (ctx *Context) dropN(n int) {
	for i := 1; i <= n; i++ {
		ctx.stack[len(ctx.stack)-i] = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}
