package goal

import "fmt"

type VariadicFun struct {
	Adverb bool
	Func   func(*Context, []V) V
}

var builtins []VariadicFun
var builtinsNames []string

func init() {
	builtins = []VariadicFun{
		vRight:    {Func: fRight},
		vAdd:      {Func: fAdd},
		vSubtract: {Func: fSubtract},
		vMultiply: {Func: fMultiply},
		vDivide:   {Func: fDivide},
		vMod:      {Func: fMod},
		vMin:      {Func: fMin},
		vMax:      {Func: fMax},
		vLess:     {Func: fLess},
		vMore:     {Func: fMore},
		vEqual:    {Func: fEqual},
		vMatch:    {Func: fMatch},
		vJoin:     {Func: fJoin},
		vCut:      {Func: fCut},
		vTake:     {Func: fTake},
		vDrop:     {Func: fDrop},
		vCast:     {Func: fCast},
		vFind:     {Func: fFind},
		vApply:    {Func: fApply},
		vApplyN:   {Func: fApplyN},
		vList:     {Func: fList},
		vEach:     {Func: fEach, Adverb: true},
		vFold:     {Func: fFold, Adverb: true},
		vScan:     {Func: fScan, Adverb: true},
	}

	builtinsNames = []string{
		vRight:    ":",
		vAdd:      "+",
		vSubtract: "-",
		vMultiply: "*",
		vDivide:   "%",
		vMod:      "!",
		vMin:      "&",
		vMax:      "|",
		vLess:     "<",
		vMore:     ">",
		vEqual:    "=",
		vMatch:    "~",
		vJoin:     ",",
		vCut:      "^",
		vTake:     "#",
		vDrop:     "_",
		vCast:     "$",
		vFind:     "?",
		vApply:    "@",
		vApplyN:   ".",
		vList:     "List",
		vEach:     "'",
		vFold:     "/",
		vScan:     "\\",
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
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			return Replicate(res, args[0])
		}
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
		return ctx.applyN(v, 1)
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
		return ctx.applyN(v, len(args)-1)
	default:
		return errs("too many arguments")
	}
}

func fList(ctx *Context, args []V) V {
	// TODO: avoid redundant cloning if canonical clones already
	res := cloneArgs(args)
	reverseArgs(res)
	return canonical(AV(res))
}

func fEach(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		v, ok := args[1].(Function)
		if !ok {
			// TODO: binary search
			return errsw("not a function")
		}
		x := toArray(args[0])
		switch x := x.(type) {
		case Array:
			res := make(AV, 0, x.Len())
			for i := 0; i < x.Len(); i++ {
				ctx.push(x.At(i))
				next := ctx.applyN(v, 1)
				if err, ok := next.(E); ok {
					return err
				}
				res = append(res, next)
			}
			return canonical(res)
		default:
			return errs("not an array")
		}
	case 3:
		v, ok := args[1].(Function)
		if !ok {
			return errsw("not a function")
		}
		x, ok := args[2].(Array)
		if !ok {
			return errsw("not an array")
		}
		y, ok := args[0].(Array)
		if !ok {
			return errs("not an array")
		}
		xlen := x.Len()
		if xlen != y.Len() {
			return errf("length mismatch: %d vs %d", x.Len(), y.Len())
		}
		res := make(AV, 0, xlen)
		for i := 0; i < xlen; i++ {
			ctx.push(y.At(i))
			ctx.push(x.At(i))
			next := ctx.applyN(v, 2)
			if err, ok := next.(E); ok {
				return err
			}
			res = append(res, next)
		}
		return canonical(res)
	default:
		return errs("too many arguments")
	}
	return nil
}

func fFold(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return fold2(ctx, args[1], args[0])
	case 3:
		v, ok := args[1].(Function)
		if !ok {
			return errs("3-rank form for adverb / expects function")
		}
		if v.Rank(ctx) != 2 {
			// TODO: while
			return errf("rank %d verb (expected 2)", v.Rank(ctx))
		}
		y := args[0]
		switch y := y.(type) {
		case Array:
			res := args[2]
			if y.Len() == 0 {
				return res
			}
			for i := 0; i < y.Len(); i++ {
				ctx.push(y.At(i))
				ctx.push(res)
				res = ctx.applyN(v, 2)
				if err, ok := res.(E); ok {
					return err
				}
			}
			return canonical(res)
		default:
			ctx.push(y)
			ctx.push(args[2])
			return ctx.applyN(v, 2)
		}
	default:
		return errs("too many arguments")
	}
	return nil
}

func fScan(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return scan2(ctx, args[1], args[0])
	case 3:
		v, ok := args[1].(Function)
		if !ok {
			return errs("3-rank form for adverb / expects function")
		}
		if v.Rank(ctx) != 2 {
			// TODO: while
			return errf("rank %d verb (expected 2)", v.Rank(ctx))
		}
		y := args[0]
		switch y := y.(type) {
		case Array:
			res := AV{args[2]}
			if y.Len() == 0 {
				return res
			}
			for i := 0; i < y.Len(); i++ {
				ctx.push(y.At(i))
				ctx.push(res[len(res)-1])
				next := ctx.applyN(v, 2)
				if err, ok := next.(E); ok {
					return err
				}
				res = append(res, next)
			}
			return canonical(res)
		default:
			ctx.push(y)
			ctx.push(args[2])
			return ctx.applyN(v, 2)
		}
	default:
		return errs("too many arguments")
	}
	return nil
}
