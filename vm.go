package goal

import "fmt"

func (ctx *Context) execute(ops []opcode) (int, error) {
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
			v := ctx.globals[ops[ip]]
			if v == nil {
				return ip, fmt.Errorf("undefined global: %s",
					ctx.gNames[ops[ip]])
			}
			ctx.push(v)
			ip++
		case opLocal:
			v := ctx.stack[ctx.frameIdx-int32(ops[ip])]
			if v == nil {
				return ip, fmt.Errorf("undefined local: %s",
					ctx.prog.Lambdas[ctx.lambda].Names[int32(ops[ip])])
			}
			ctx.push(v)
			ip++
		case opAssignGlobal:
			ctx.globals[ops[ip]] = ctx.top()
			ip++
		case opAssignLocal:
			ctx.stack[ctx.frameIdx-int32(ops[ip])] = ctx.top()
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
				return ip, err
			}
		case opApplyVariadic:
			v := Variadic(ops[ip])
			res := ctx.applyVariadic(v)
			if err, ok := res.(error); ok && err != nil {
				return ip, err
			}
			ctx.push(res)
			ip++
		case opApply2:
			err := ctx.popApplyN(2)
			if err != nil {
				return ip, err
			}
		case opApply2Variadic:
			v := Variadic(ops[ip])
			res := ctx.applyNVariadic(v, 2)
			if err, ok := res.(error); ok && err != nil {
				return ip, err
			}
			ctx.push(res)
			ip++
		case opApplyN:
			err := ctx.popApplyN(int(ops[ip]))
			if err != nil {
				return ip, err
			}
			ip++
		case opApplyNVariadic:
			v := Variadic(ops[ip])
			ip++
			res := ctx.applyNVariadic(v, int(ops[ip]))
			if err, ok := res.(error); ok && err != nil {
				return ip, err
			}
			ctx.push(res)
			ip++
		case opDrop:
			ctx.drop()
		}
		//fmt.Printf("stack: %v\n", ctx.stack)
	}
	return len(ops), nil
}

func (ctx *Context) popApplyN(n int) error {
	//olen := len(ctx.stack)
	v := ctx.pop()
	res := ctx.applyN(v, n)
	if err, ok := res.(error); ok {
		return err
	}
	ctx.push(res)
	//if len(ctx.stack) != olen-n {
	//return fmt.Errorf("call (%v with %d args): bad stack length: %d vs %d (stack: %v)", v, n, len(ctx.stack), olen-n, ctx.stack)
	//}
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
	topN := ctx.stack[len(ctx.stack)-n:]
	for i := range topN {
		topN[i] = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}
