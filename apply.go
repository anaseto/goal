package main

import (
	"fmt"
)

func Apply(v V, x V) (res V) {
	switch v := v.(type) {
	case Monad:
		switch v {
		case VReturn:
			res = x // TODO: VReturn
		case VFlip:
			res = Flip(x)
		case VNegate:
			res = Negate(x)
		case VFirst:
			res = First(x)
		case VClassify:
			res = Classify(x)
		case VEnum:
			res = Range(x)
		case VWhere:
			res = Indices(x)
		case VReverse:
			res = Reverse(x)
		case VAscend:
			res = GradeUp(x)
		case VDescend:
			res = GradeDown(x)
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
		case VNub:
			res = errNYI("Apply VNub") // TODO
		case VType:
			res = errNYI("Apply VType") // TODO
		case VEval:
			res = errNYI("Apply VEval") // TODO
		}
	case Dyad:
		res = Projection{Fun: v, Args: AV{x, nil}}
	case Projection:
		//args := v.Args
		//for
		res = errNYI("Apply Projection") // TODO
	default:
		res = errNYI("Apply other") // TODO
	}
	return res
}

func Apply2(v, w, x V) (res V) {
	switch v := v.(type) {
	case Monad:
		res = errf("monad %v got too many arguments", v)
	case Dyad:
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
		case VConcat:
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
			res = Apply(w, x)
		case VApplyN:
			res = errNYI("Apply2 VApplyN") // TODO
		}
	default:
		res = errNYI("Apply2 other") // TODO
	}
	return res
}
