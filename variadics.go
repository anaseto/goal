package goal

//import "fmt"

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
		vWithout:  {Func: VWithout},
		vTake:     {Func: VTake},
		vDrop:     {Func: VDrop},
		vCast:     {Func: VCast},
		vFind:     {Func: VFind},
		vApply:    {Func: VApply},
		vApplyN:   {Func: VApplyN},
		vIn:       {Func: VIn},
		vSign:     {Func: VSign},
		vOCount:   {Func: VOCount},
		vICount:   {Func: VICount},
		vBytes:    {Func: VBytes},
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
		vWithout:  "^",
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
		return errRank(":")
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
		return errRank("+")
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
		return errRank("-")
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
		return errRank("*")
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
		return errRank("%")
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
		return errRank("!")
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
		return errRank("&")
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
			if err, ok := res.(errV); ok {
				return err
			}
			return rotate(res, args[0])
		}
		return maximum(args[1], args[0])
	default:
		return errRank("|")
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
		return errRank("<")
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
		return errRank(">")
	}
}

// VEqual implements the = variadic verb.
func VEqual(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return group(args[0])
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			if err, ok := res.(errV); ok {
				return err
			}
			return groupBy(res, args[0])
		}
		return equal(args[1], args[0])
	default:
		return errRank("=")
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
		return errRank("~")
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
		return errRank(",")
	}
}

// VWithout implements the ^ variadic verb.
func VWithout(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sortUp(args[0])
	case 2:
		return without(args[1], args[0])
	default:
		return errRank("^")
	}
}

// VTake implements the # variadic verb.
func VTake(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return I(Length(args[0]))
	case 2:
		v, ok := args[1].(Function)
		if ok {
			ctx.push(args[0])
			res := ctx.applyN(v, 1)
			if err, ok := res.(errV); ok {
				return err
			}
			return replicate(res, args[0])
		}
		return take(args[1], args[0])
	default:
		return errRank("#")
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
			if err, ok := res.(errV); ok {
				return err
			}
			return weedOut(res, args[0])
		}
		return drop(args[1], args[0])
	default:
		return errRank("_")
	}
}

// VCast implements the $ variadic verb.
func VCast(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return S(args[0].Sprint(ctx))
	case 2:
		switch args[1].(type) {
		case array:
			return search(args[1], args[0])
		default:
			return cast(args[1], args[0])
		}
	default:
		return errRank("$")
	}
}

// VFind implements the ? variadic verb.
func VFind(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return uniq(args[0])
	case 2:
		return find(args[1], args[0])
	default:
		return errRank("?")
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
		return errRank("@")
	}
}

// VApplyN implements the . variadic verb.
func VApplyN(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return eval(ctx, args[0])
	case 2:
		v := args[1]
		av := toArray(args[0]).(array)
		for i := av.Len() - 1; i >= 0; i-- {
			ctx.push(av.at(i))
		}
		return ctx.applyN(v, av.Len())
	case 3:
		return try(ctx, args[2], args[1], args[0])
	default:
		return errRank(".")
	}
}

// VIn implements the "in" variadic verb.
func VIn(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return errs("in : got only one argument")
	case 2:
		return memberOf(args[1], args[0])
	default:
		return errRank("in")
	}
}

// VSign implements the "sign" variadic verb.
func VSign(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sign(args[0])
	default:
		return errRank("sign")
	}
}

// VOCount implements the "ocount" variadic verb.
func VOCount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return occurrenceCount(args[0])
	default:
		return errRank("ocount")
	}
}

// VICount implements the "icount" variadic verb.
func VICount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return icount(args[0])
	default:
		return errRank("icount")
	}
}

// VBytes implements the "bytes" variadic verb.
func VBytes(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return bytes(args[0])
	default:
		return errRank("icount")
	}
}

// VList implements (...;y;x) array constructor variadic verb.
func VList(ctx *Context, args []V) V {
	t, ok := isCanonical(AV(args))
	if ok {
		res := cloneArgs(args)
		reverseArgs(res)
		return AV(res)
	}
	switch t {
	case tB, tI, tF, tS:
		res := canonical(AV(args))
		reverseMut(res)
		return res
	default:
		res := cloneArgs(args)
		reverseArgs(res)
		return canonical(AV(res))
	}
}

// VEach implements the ' variadic adverb.
func VEach(ctx *Context, args []V) V {
	switch len(args) {
	case 2:
		return each2(ctx, args)
	case 3:
		return each3(ctx, args)
	default:
		return errRank("'")
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
		return errRank("/")
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
		return errRank("\\")
	}
	return nil
}
