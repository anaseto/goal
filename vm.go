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
		case opListAssignGlobal:
			x := ctx.top()
			ids := ctx.gAssignLists[ops[ip]]
			err := ctx.assignGlobals(ids, x)
			if err != nil {
				return ip - 1, err
			}
			ip++
		case opListAssignLocal:
			x := ctx.top()
			ids := ctx.lambdas[ctx.lambda].AssignLists[ops[ip]]
			err := ctx.assignLocals(ids, x)
			if err != nil {
				return ip - 1, err
			}
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
				ctx.clearAssignOnPanic(ops, ip) // currently unnecessary case (but could be in the future)
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyGlobal:
			x := ctx.globals[ops[ip]]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined global: %s",
					ctx.gNames[ops[ip]])
			}
			r := x.applyN(ctx, 1)
			if r.IsPanic() {
				ctx.clearAssignOnPanic(ops, ip) // currently unnecessary case
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
				ctx.clearAssignOnPanic(ops, ip-1) // currently unnecessary case
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
		case opApply2V:
			v := variadic(ops[ip])
			r := v.apply2(ctx)
			if r.IsPanic() {
				ctx.clearAssignOnPanic(ops, ip)
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyN:
			x := ctx.pop()
			r := x.applyN(ctx, int(ops[ip]))
			if r.IsPanic() {
				ctx.clearAssignOnPanic(ops, ip) // currently unnecessary case
				return ip - 1, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyNV:
			v := variadic(ops[ip])
			ip++
			r := v.applyN(ctx, int(ops[ip]))
			if r.IsPanic() {
				ctx.clearAssignOnPanic(ops, ip)
				return ip - 2, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opApplyNGlobal:
			x := ctx.globals[ops[ip]]
			if x.kind == valNil {
				return ip - 1, fmt.Errorf("undefined global: %s",
					ctx.gNames[ops[ip]])
			}
			ip++
			r := x.applyN(ctx, int(ops[ip]))
			if r.IsPanic() {
				ctx.clearAssignOnPanic(ops, ip) // currently unnecessary case
				return ip - 2, newExecError(r)
			}
			ctx.replaceTop(r)
			ip++
		case opDrop:
			ctx.drop()
		case opJump:
			ip += int(ops[ip])
		case opJumpFalse:
			if ctx.top().IsFalse() {
				ip += int(ops[ip])
			} else {
				ip++
			}
		case opJumpTrue:
			if ctx.top().IsTrue() {
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
		ctx.stack[len(ctx.stack)-1].bv = nil
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
		v.bv = nil
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
}

func (ctx *Context) dropN(n int) {
	topN := ctx.stack[len(ctx.stack)-n:]
	for i := range topN {
		v := &topN[i]
		if v.kind == valBoxed {
			v.rcdecrRefCounter()
			v.bv = nil
		}
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}

func (ctx *Context) dropNnoRC(n int) {
	topN := ctx.stack[len(ctx.stack)-n:]
	for i := range topN {
		v := &topN[i]
		if v.kind == valBoxed {
			v.bv = nil
		}
	}
	ctx.stack = ctx.stack[:len(ctx.stack)-n]
}

func (ctx *Context) assignGlobals(ids []int, x V) error {
	switch xv := x.bv.(type) {
	case Array:
		if len(ids) > xv.Len() {
			return fmt.Errorf("length mismatch in list assignment (%d > %d)", len(ids), xv.Len())
		}
		for i, id := range ids {
			xi := xv.VAt(i)
			xi.IncrRC()
			ctx.globals[id] = xi
		}
		return nil
	default:
		return fmt.Errorf("non-array value in list assignment (%s)", x.Type())
	}
}

func (ctx *Context) assignLocals(ids []int32, x V) error {
	switch xv := x.bv.(type) {
	case Array:
		if len(ids) > xv.Len() {
			return fmt.Errorf("length error in list assignment (%d > %d)", len(ids), xv.Len())
		}
		for i, id := range ids {
			xi := xv.VAt(i)
			xi.IncrRC()
			ctx.stack[ctx.frameIdx-id] = xi
		}
		return nil
	default:
		return fmt.Errorf("non-array value in list assignment (%s)", x.Type())
	}
}

func rcincrArgs(args []V) {
	for _, v := range args {
		v.IncrRC()
	}
}

func (ctx *Context) clearAssignOnPanic(ops []opcode, ip int) {
	// NOTE: it's not a perfect solution, because ideally we would want to
	// preserve the old value, instead of clearing.
	// With current optimizations, only simple global assignement patterns
	// can get in-place modification.  Local assignement doesn't need
	// clearing, because on panic, local variables in panic's context are
	// no longer accessible.
	ip++
	if ip >= len(ops) {
		return
	}
	op := ops[ip]
	ip++
	if op == opAssignGlobal {
		ctx.globals[ops[ip]] = V{}
	}
}
