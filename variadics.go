package goal

import "fmt"

// VariadicFun represents a variadic function, either a verb or an adverb.
type VariadicFun struct {
	Adverb bool
	Func   func(*Context, []V) V
}

func (ctx *Context) initVariadics() {
	ctx.variadics = []VariadicFun{
		vRight:    {Func: VRight},
		vAdd:      {Func: VAdd},
		vSubtract: {Func: VSubtract},
		vMultiply: {Func: VMultiply},
		vDivide:   {Func: VDivide},
		vMod:      {Func: VMod},
		vMin:      {Func: VMin},
		vMax:      {Func: VMax},
		vLess:     {Func: VLess},
		vMore:     {Func: VMore},
		vEqual:    {Func: VEqual},
		vMatch:    {Func: VMatch},
		vJoin:     {Func: VJoin},
		vCut:      {Func: VCut},
		vTake:     {Func: VTake},
		vDrop:     {Func: VDrop},
		vCast:     {Func: VCast},
		vFind:     {Func: VFind},
		vApply:    {Func: VApply},
		vApplyN:   {Func: VApplyN},
		vList:     {Func: VList},
		vEach:     {Func: VEach, Adverb: true},
		vFold:     {Func: VFold, Adverb: true},
		vScan:     {Func: VScan, Adverb: true},
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

// VRight implements the : variadic verb.
func VRight(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return args[0]
	case 2:
		return args[0]
	default:
		return errs(": got too many arguments")
	}
}

// VAdd implements the + variadic verb.
func VAdd(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return flip(args[0])
	case 2:
		return add(args[1], args[0])
	default:
		return errs("+ got too many arguments")
	}
}

// VSubtract implements the - variadic verb.
func VSubtract(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return negate(args[0])
	case 2:
		return subtract(args[1], args[0])
	default:
		return errs("- got too many arguments")
	}
}

// VMultiply implements the * variadic verb.
func VMultiply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return first(args[0])
	case 2:
		return multiply(args[1], args[0])
	default:
		return errs("* got too many arguments")
	}
}

// VDivide implements the % variadic verb.
func VDivide(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return classify(args[0])
	case 2:
		return divide(args[1], args[0])
	default:
		return errs("%% got too many arguments")
	}
}

// VMod implements the ! variadic verb.
func VMod(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enum(args[0])
	case 2:
		return modulus(args[1], args[0])
	default:
		return errs("! got too many arguments")
	}
}

// VMin implements the & variadic verb.
func VMin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return where(args[0])
	case 2:
		return minimum(args[1], args[0])
	default:
		return errs("& got too many arguments")
	}
}

// VMax implements the | variadic verb.
func VMax(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return reverse(args[0])
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			if err, ok := res.(E); ok {
				return err
			}
			return rotate(res, args[0])
		}
		return maximum(args[1], args[0])
	default:
		return errs("| got too many arguments")
	}
}

// VLess implements the < variadic verb.
func VLess(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return ascend(args[0])
	case 2:
		return lesser(args[1], args[0])
	default:
		return errs("< got too many arguments")
	}
}

// VMore implements the > variadic verb.
func VMore(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return descend(args[0])
	case 2:
		return greater(args[1], args[0])
	default:
		return errs("> got too many arguments")
	}
}

// VEqual implements the = variadic verb.
func VEqual(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return group(args[0])
	case 2:
		return equal(args[1], args[0])
	default:
		return errs("= got too many arguments")
	}
}

// VMatch implements the ~ variadic verb.
func VMatch(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return not(args[0])
	case 2:
		return B2I(Match(args[1], args[0]))
	default:
		return errs("~ got too many arguments")
	}
}

// VJoin implements the , variadic verb.
func VJoin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enlist(args[0])
	case 2:
		return joinTo(args[1], args[0])
	default:
		return errs(", got too many arguments")
	}
}

// VCut implements the ^ variadic verb.
func VCut(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sortUp(args[0])
	case 2:
		return cut(args[1], args[0])
	default:
		return errs("^ got too many arguments")
	}
}

// VTake implements the # variadic verb.
func VTake(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return length(args[0])
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			if err, ok := res.(E); ok {
				return err
			}
			return replicate(res, args[0])
		}
		return take(args[1], args[0])
	default:
		return errs("# got too many arguments")
	}
}

// VDrop implements the _ variadic verb.
func VDrop(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return floor(args[0])
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			if err, ok := res.(E); ok {
				return err
			}
			return weedOut(res, args[0])
		}
		return drop(args[1], args[0])
	default:
		return errs("_ got too many arguments")
	}
}

// VCast implements the $ variadic verb.
func VCast(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return S(fmt.Sprint(args[0]))
	case 2:
		return cast(args[1], args[0])
	default:
		return errs("$ got too many arguments")
	}
}

// VFind implements the ? variadic verb.
func VFind(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return uniq(args[0])
	case 2:
		return errNYI("dyadic ?")
	default:
		return errs("? got too many arguments")
	}
}

// VApply implements the @ variadic verb.
func VApply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return S(args[0].Type())
	case 2:
		v := args[1]
		ctx.push(args[0])
		return ctx.applyN(v, 1)
	default:
		return errs("@ got too many arguments")
	}
}

// VApplyN implements the . variadic verb.
func VApplyN(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return eval(ctx, args[0])
	case 2:
		v := args[1]
		av := toArray(args[0]).(Array)
		for i := av.Len() - 1; i >= 0; i-- {
			ctx.push(av.At(i))
		}
		return ctx.applyN(v, av.Len())
	default:
		return errs(". got too many arguments")
	}
}

// VList implements (...;y;x) array constructor variadic verb.
func VList(ctx *Context, args []V) V {
	// TODO: avoid redundant cloning if canonical clones already
	res := cloneArgs(args)
	reverseArgs(res)
	return canonical(AV(res))
}

// VEach implements the ' variadic adverb.
func VEach(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return each2(ctx, args)
	case 3:
		return each3(ctx, args)
	default:
		return errs("too many arguments")
	}
	return nil
}

// VFold implements the / variadic adverb.
func VFold(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return fold2(ctx, args)
	case 3:
		return fold3(ctx, args)
	default:
		return errs("too many arguments")
	}
	return nil
}

// VScan implements the \ variadic adverb.
func VScan(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return scan2(ctx, args[1], args[0])
	case 3:
		return scan3(ctx, args)
	default:
		return errs("too many arguments")
	}
	return nil
}
