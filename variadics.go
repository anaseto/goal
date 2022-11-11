package goal

import "fmt"

// VariadicFun represents a variadic function, either a verb or an adverb.
type VariadicFun struct {
	Adverb bool
	Func   func(*Context, []V) V
}

func (ctx *Context) initVariadics() {
	ctx.variadics = []VariadicFun{
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

	ctx.variadicsNames = []string{
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

// fRight implements the : variadic verb.
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

// fAdd implements the + variadic verb.
func fAdd(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return flip(args[0])
	case 2:
		return add(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fSubtract implements the - variadic verb.
func fSubtract(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return negate(args[0])
	case 2:
		return subtract(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMultiply implements the * variadic verb.
func fMultiply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return first(args[0])
	case 2:
		return multiply(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fDivide implements the % variadic verb.
func fDivide(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return classify(args[0])
	case 2:
		return divide(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMod implements the ! variadic verb.
func fMod(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enum(args[0])
	case 2:
		return modulus(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMin implements the & variadic verb.
func fMin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return where(args[0])
	case 2:
		return minimum(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMax implements the | variadic verb.
func fMax(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return reverse(args[0])
	case 2:
		return maximum(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fLess implements the < variadic verb.
func fLess(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return ascend(args[0])
	case 2:
		return lesser(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMore implements the > variadic verb.
func fMore(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return descend(args[0])
	case 2:
		return greater(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fEqual implements the = variadic verb.
func fEqual(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return group(args[0])
	case 2:
		return equal(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fMatch implements the ~ variadic verb.
func fMatch(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return not(args[0])
	case 2:
		return match(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fJoin implements the , variadic verb.
func fJoin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enlist(args[0])
	case 2:
		return joinTo(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fCut implements the ^ variadic verb.
func fCut(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sortUp(args[0])
	case 2:
		return errNYI("dyadic ^")
		//return Cut(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fTake implements the # variadic verb.
func fTake(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return length(args[0])
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			return replicate(res, args[0])
		}
		return take(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fDrop implements the _ variadic verb.
func fDrop(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return floor(args[0])
	case 2:
		return drop(args[1], args[0])
	default:
		return errs("too many arguments")
	}
}

// fCast implements the $ variadic verb.
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

// fFind implements the ? variadic verb.
func fFind(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return uniq(args[0])
	case 2:
		return errNYI("dyadic ?")
	default:
		return errs("too many arguments")
	}
}

// fApply implements the @ variadic verb.
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

// fApplyN implements the . variadic verb.
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

// fList implements (x;y;...) array building variadic verb.
func fList(ctx *Context, args []V) V {
	// TODO: avoid redundant cloning if canonical clones already
	res := cloneArgs(args)
	reverseArgs(res)
	return canonical(AV(res))
}

// fEach implements the ' variadic adverb.
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

// fFold implements the / variadic adverb.
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

// fScan implements the \ variadic adverb.
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
