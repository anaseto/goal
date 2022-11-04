package main

import "fmt"

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
			ctx.apply()
		case opApply2:
			ctx.apply2()
		case opApplyN:
			ip++
		case opDrop:
			ctx.sp--
		}
	}
	return nil
}

func (ctx *Context) apply() {
	v := ctx.pop()
	x := ctx.pop()
	switch v := v.(type) {
	case Monad:
		switch v {
		case VReturn:
			// TODO: VReturn
			ctx.push(x)
		case VFlip:
			ctx.push(Flip(x))
		case VNegate:
			ctx.push(Negate(x))
		case VFirst:
			ctx.push(First(x))
		case VClassify:
			ctx.push(Classify(x))
		case VEnum:
			ctx.push(Range(x))
		case VWhere:
			ctx.push(Indices(x))
		case VReverse:
			ctx.push(Reverse(x))
		case VAscend:
			ctx.push(GradeUp(x))
		case VDescend:
			ctx.push(GradeDown(x))
		case VGroup:
			ctx.push(Group(x))
		case VNot:
			ctx.push(Not(x))
		case VEnlist:
			ctx.push(Enlist(x))
		case VSort:
			ctx.push(SortUp(x))
		case VLen:
			ctx.push(Length(x))
		case VFloor:
			ctx.push(Floor(x))
		case VString:
			// TODO: VString
			ctx.push(S(fmt.Sprint(x)))
		case VNub:
			panic("Apply VNub") // TODO
		case VType:
			panic("Apply VType") // TODO
		case VEval:
			panic("Apply VEval") // TODO
		}
	default:
		panic("Apply other") // TODO
	}
}

func (ctx *Context) apply2() {
	v := ctx.pop()
	w, x := ctx.pop2()
	switch v := v.(type) {
	case Dyad:
		switch v {
		case VRight:
			ctx.push(x)
		case VAdd:
			ctx.push(Add(w, x))
		case VSubtract:
			ctx.push(Subtract(w, x))
		case VMultiply:
			ctx.push(Multiply(w, x))
		case VDivide:
			ctx.push(Divide(w, x))
		case VMod:
			ctx.push(Modulus(w, x))
		case VMin:
			ctx.push(Minimum(w, x))
		case VMax:
			ctx.push(Maximum(w, x))
		case VLess:
			ctx.push(Lesser(w, x))
		case VMore:
			ctx.push(Greater(w, x))
		case VEqual:
			ctx.push(Equal(w, x))
		case VMatch:
			ctx.push(Match(w, x))
		case VConcat:
			ctx.push(JoinTo(w, x))
		case VCut:
			panic("Apply2 VCut") // TODO
		case VTake:
			ctx.push(Take(w, x))
		case VDrop:
			ctx.push(Drop(w, x))
		case VCast:
			panic("Apply2 VCast") // TODO
		case VFind:
			panic("Apply2 VFind") // TODO
		case VApply:
			ctx.apply()
		case VApplyN:
			panic("Apply2 VApplyN") // TODO
		}
	default:
		panic("Apply2 other") // TODO
	}
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
