package main

import (
	"fmt"
)

func (ctx *Context) Apply(v V, x V) (res V) {
	switch v := v.(type) {
	case Monad:
		res = ctx.applyMonad(v, x)
	case Dyad:
		res = Projection{Fun: v, Args: AV{x, nil}}
	case Projection:
		args := v.Args
		for i, v := range args {
			if v == nil {
				args[i] = x
				break
			}
		}
		for _, arg := range args {
			if arg == nil {
				return v
			}
		}
		res = ctx.ApplyN(v.Fun, args)
	case Array:
		indices := toIndices(x)
		if indices == nil {
			return errs("not an integer array")
		}
		res = v.Apply(indices)
	default:
		res = errs("atoms cannot be applied")
	}
	return res
}

func (ctx *Context) applyMonad(v Monad, x V) (res V) {
	switch v {
	case VReturn:
		res = x // TODO: VReturn: probably syntax instead of value
	case VFlip:
		res = Flip(x)
	case VNegate:
		res = Negate(x)
	case VFirst:
		res = First(x)
	case VClassify:
		res = Classify(x)
	case VRange:
		res = Range(x)
	case VWhere:
		res = Where(x)
	case VReverse:
		res = Reverse(x)
	case VAscend:
		res = Ascend(x)
	case VDescend:
		res = Descend(x)
	case VGroup:
		res = Group(x)
	case VNot:
		res = Not(x)
	case VEnlist:
		res = Enlist(x)
	case VSort:
		res = SortUp(x)
	case VLen:
		res = Length(x)
	case VFloor:
		res = Floor(x)
	case VString:
		// TODO: VString
		res = S(fmt.Sprint(x))
	case VUniq:
		res = Uniq(x)
	case VType:
		res = S(x.Type())
	case VEval:
		res = errNYI("Apply VEval") // TODO
	}
	return res
}

func (x AV) Apply(y AI) V {
	res := make(AV, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AB) Apply(y AI) V {
	res := make(AB, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AI) Apply(y AI) V {
	res := make(AI, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AF) Apply(y AI) V {
	res := make(AF, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (x AS) Apply(y AI) V {
	res := make(AS, len(y))
	for i := range res {
		idx := y[i]
		if idx < 0 || idx >= len(x) {
			return errf("index out of bounds: %d (length %d)", idx, len(x))
		}
		res[i] = x[y[i]]
	}
	return res
}

func (ctx *Context) Apply2(v, w, x V) (res V) {
	switch v := v.(type) {
	case Monad:
		res = errf("monad %v got too many arguments", v)
	case Dyad:
		res = ctx.applyDyad(v, w, x)
	case Projection:
		args := v.Args
		count := 0
		for i, v := range args {
			if v == nil {
				if count == 0 {
					args[i] = w
				} else {
					args[i] = x
					break
				}
				count++
			}
		}
		for _, arg := range args {
			if arg == nil {
				return v
			}
		}
		res = ctx.ApplyN(v.Fun, args)
	default:
		res = errNYI("Apply2 other") // TODO
	}
	return res
}

func (ctx *Context) applyDyad(v Dyad, w, x V) (res V) {
	switch v {
	case VRight:
		res = x
	case VAdd:
		res = Add(w, x)
	case VSubtract:
		res = Subtract(w, x)
	case VMultiply:
		res = Multiply(w, x)
	case VDivide:
		res = Divide(w, x)
	case VMod:
		res = Modulus(w, x)
	case VMin:
		res = Minimum(w, x)
	case VMax:
		res = Maximum(w, x)
	case VLess:
		res = Lesser(w, x)
	case VMore:
		res = Greater(w, x)
	case VEqual:
		res = Equal(w, x)
	case VMatch:
		res = Match(w, x)
	case VJoin:
		res = JoinTo(w, x)
	case VCut:
		res = errNYI("Apply2 VCut") // TODO
	case VTake:
		res = Take(w, x)
	case VDrop:
		res = Drop(w, x)
	case VCast:
		res = errNYI("Apply2 VCast") // TODO
	case VFind:
		res = errNYI("Apply2 VFind") // TODO
	case VApply:
		res = ctx.Apply(w, x)
	case VApplyN:
		res = errNYI("Apply2 VApplyN") // TODO
	}
	return res
}

func (ctx *Context) ApplyN(v V, argn []V) (res V) {
	switch v := v.(type) {
	case Monad:
		switch len(argn) {
		case 0:
			res = Projection{Fun: v, Args: AV{nil, nil}}
		case 1:
			res = ctx.applyMonad(v, argn[0])
		default:
			res = errf("monad %v got too many arguments", v)
		}
	case Dyad:
		switch len(argn) {
		case 0:
			res = Projection{Fun: v, Args: AV{nil, nil}}
		case 1:
			res = Projection{Fun: v, Args: AV{argn[0], nil}}
		case 2:
			res = ctx.applyDyad(v, argn[0], argn[1])
		default:
			res = errf("dyad %v got too many arguments", v)
		}
	case Projection:
		args := v.Args
		n := 0
		for i, v := range args {
			if v == nil {
				args[i] = argn[n]
				break
			}
		}
		for _, arg := range args {
			if arg == nil {
				return v
			}
		}
		res = ctx.ApplyN(v.Fun, args)
	}
	return res
}
