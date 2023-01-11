package goal

import "fmt"

func (ctx *Context) execute(ops []opcode) (int, error) {
	for ip := 0; ip < len(ops); {
		op := ops[ip]
		//fmt.Printf("op: %s\n", op)
		ip++
		switch op {
		case opConst:
			ctx.push(ctx.constants[ops[ip]])
			ip++
		case opInt:
			ctx.pushNoRC(NewI(int64(ops[ip])))
			ip++
		case opNil:
			ctx.pushNoRC(V{})
		case opGlobal:
			x := ctx.globals[ops[ip]]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined global: %s",
					ctx.gNames[ops[ip]])
			}
			ctx.push(x)
			ip++
		case opGlobalLast:
			x := ctx.globals[ops[ip]]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined global: %s",
					ctx.gNames[ops[ip]])
			}
			ctx.pushNoRC(x)
			ip++
		case opLocal:
			x := ctx.stack[ctx.frameIdx-int32(ops[ip])]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined local: %s",
					ctx.lambdas[ctx.lambda].Names[int32(ops[ip])])
			}
			ctx.push(x)
			ip++
		case opLocalLast:
			x := ctx.stack[ctx.frameIdx-int32(ops[ip])]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined local: %s",
					ctx.lambdas[ctx.lambda].Names[int32(ops[ip])])
			}
			ctx.pushNoRC(x)
			ip++
		case opAssignGlobal:
			x := ctx.top()
			x.IncrRC()
			ctx.globals[ops[ip]] = x
			ip++
		case opAssignLocal:
			x := ctx.top()
			x.IncrRC()
			ctx.stack[ctx.frameIdx-int32(ops[ip])] = x
			ip++
		case opVariadic:
			ctx.pushNoRC(newVariadic(variadic(ops[ip])))
			ip++
		case opLambda:
			ctx.pushNoRC(newLambda(lambda(ops[ip])))
			ip++
		case opApply:
			x := ctx.pop()
			r := x.applyN(ctx, 1)
			if r.IsPanic() {
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
		case opApplyV:
			v := variadic(ops[ip])
			r := v.apply(ctx)
			if r.IsPanic() {
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opDerive:
			v := variadic(ops[ip])
			ctx.stack[len(ctx.stack)-1] = NewV(&derivedVerb{Fun: v, Arg: ctx.top()})
			ip++
		case opApply2:
			x := ctx.pop()
			r := x.applyN(ctx, 2)
			if r.IsPanic() {
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
		case opApply2V:
			v := variadic(ops[ip])
			r := v.apply2(ctx)
			if r.IsPanic() {
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyN:
			x := ctx.pop()
			r := x.applyN(ctx, int(ops[ip]))
			if r.IsPanic() {
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyNV:
			v := variadic(ops[ip])
			ip++
			r := v.applyN(ctx, int(ops[ip]))
			if r.IsPanic() {
				return ip - 2, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opDrop:
			ctx.drop()
		case opJump:
			ip += int(ops[ip])
		case opJumpFalse:
			if isFalse(ctx.top()) {
				ip += int(ops[ip])
			} else {
				ip++
			}
		case opJumpTrue:
			if isTrue(ctx.top()) {
				ip += int(ops[ip])
			} else {
				ip++
			}
		case opReturn:
			return len(ops), nil
		case opTry:
			if ctx.top().IsError() {
				return len(ops), nil
			}
		}
		//fmt.Printf("stack: %v\n", ctx.stack)
	}
	return len(ops), nil
}

const maxCallDepth = 100000

func (ctx *Context) push(x V) {
	x.IncrRC()
	ctx.stack = append(ctx.stack, x)
}

func (ctx *Context) pushNoRC(x V) {
	ctx.stack = append(ctx.stack, x)
}

func (ctx *Context) pushArgs(args []V) {
	rcincrArgs(args)
	ctx.stack = append(ctx.stack, args...)
}

func (ctx *Context) pop() V {
	arg := ctx.stack[len(ctx.stack)-1]
	if arg.kind == valBoxed {
		arg.rcdecrRefCounter()
		ctx.stack[len(ctx.stack)-1].value = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
	return arg
}

func (ctx *Context) top() V {
	return ctx.stack[len(ctx.stack)-1]
}

func (ctx *Context) replaceTop(x V) {
	v := &ctx.stack[len(ctx.stack)-1]
	if v.kind == valBoxed {
		v.rcdecrRefCounter()
	}
	*v = x
	x.IncrRC()
}

func (ctx *Context) peek() []V {
	return ctx.stack[len(ctx.stack)-1:]
}

func (ctx *Context) peekN(n int) []V {
	return ctx.stack[len(ctx.stack)-n:]
}

func (ctx *Context) drop() {
	if v := &ctx.stack[len(ctx.stack)-1]; v.kind == valBoxed {
		v.rcdecrRefCounter()
		v.value = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
}

func (ctx *Context) dropN(n int) {
	topN := ctx.stack[len(ctx.stack)-n:]
	for i := range topN {
		v := &topN[i]
		if v.kind == valBoxed {
			v.rcdecrRefCounter()
			v.value = nil
		}
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}

func (ctx *Context) dropNnoRC(n int) {
	topN := ctx.stack[len(ctx.stack)-n:]
	for i := range topN {
		v := &topN[i]
		if v.kind == valBoxed {
			v.value = nil
		}
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}

func rcincrArgs(args []V) {
	for _, v := range args {
		v.IncrRC()
	}
}
