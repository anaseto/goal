package main

import "fmt"

type VariadicFun struct {
	Name   string
	Adverb bool
	Func   func(*Context, []V) V
}

var builtins []VariadicFun

func init() {
	builtins = []VariadicFun{
		vRight:    {Name: ":", Func: fRight},
		vAdd:      {Name: "+", Func: fAdd},
		vSubtract: {Name: "-", Func: fSubtract},
		vMultiply: {Name: "*", Func: fMultiply},
		vDivide:   {Name: "%", Func: fDivide},
		vMod:      {Name: "!", Func: fMod},
		vMin:      {Name: "&", Func: fMin},
		vMax:      {Name: "|", Func: fMax},
		vLess:     {Name: "<", Func: fLess},
		vMore:     {Name: ">", Func: fMore},
		vEqual:    {Name: "=", Func: fEqual},
		vMatch:    {Name: "~", Func: fMatch},
		vJoin:     {Name: ",", Func: fJoin},
		vCut:      {Name: "^", Func: fCut},
		vTake:     {Name: "#", Func: fTake},
		vDrop:     {Name: "_", Func: fDrop},
		vCast:     {Name: "$", Func: fCast},
		vFind:     {Name: "?", Func: fFind},
		vApply:    {Name: "@", Func: fApply},
		vApplyN:   {Name: ".", Func: fApplyN},
		vList:     {Name: "List", Func: fList},
		vEach:     {Name: "'", Func: fEach, Adverb: true},
		vFold:     {Name: "/", Func: fFold, Adverb: true},
		vScan:     {Name: "\\", Func: fScan, Adverb: true},
	}
}

func fRight(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return args[0]
	case 2:
		return args[0]
	default:
		return errs("too many arguments")
	}
}

func fAdd(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Flip(args[0])
	case 2:
		return Add(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fSubtract(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Negate(args[0])
	case 2:
		return Subtract(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMultiply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return First(args[0])
	case 2:
		return Multiply(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fDivide(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Classify(args[0])
	case 2:
		return Divide(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMod(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Range(args[0])
	case 2:
		return Modulus(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Where(args[0])
	case 2:
		return Minimum(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMax(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Reverse(args[0])
	case 2:
		return Maximum(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fLess(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Ascend(args[0])
	case 2:
		return Lesser(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMore(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Descend(args[0])
	case 2:
		return Greater(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fEqual(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Group(args[0])
	case 2:
		return Equal(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fMatch(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Not(args[0])
	case 2:
		return Match(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fJoin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Enlist(args[0])
	case 2:
		return JoinTo(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fCut(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return SortUp(args[0])
	case 2:
		return errNYI("dyadic ^")
		//return Cut(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fTake(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Length(args[0])
	case 2:
		return Take(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fDrop(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Floor(args[0])
	case 2:
		return Drop(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

func fCast(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return S(fmt.Sprint(args[0]))
	case 2:
		return errNYI("dyadic $")
	default:
		return errs("too many arguments")
	}
}

func fFind(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return Uniq(args[0])
	case 2:
		return errNYI("dyadic ?")
	default:
		return errs("too many arguments")
	}
}

func fApply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return S(args[0].Type())
	case 2:
		v := args[1]
		ctx.push(args[0])
		return ctx.ApplyN(v, 1)
	default:
		return errs("too many arguments")
	}
}

func fApplyN(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return errNYI("monadic .")
	case 2:
		v := args[len(args)-1]
		ctx.pushArgs(args[:len(args)-1])
		return ctx.ApplyN(v, len(args)-1)
	default:
		return errs("too many arguments")
	}
}

func fList(ctx *Context, args []V) V {
	res := cloneArgs(args)
	reverseArgs(res)
	return AV(res)
}

func fEach(ctx *Context, args []V) V {
	return nil
}

func fFold(ctx *Context, args []V) V {
	return nil
}

func fScan(ctx *Context, args []V) V {
	return nil
}
