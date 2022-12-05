package goal

import (
	"fmt"
	"math"
)

// variadic represents a built-in function.
type variadic int32

const (
	vRight    variadic = iota // :
	vAdd                      // +
	vSubtract                 // -
	vMultiply                 // *
	vDivide                   // %
	vMod                      // !
	vMin                      // &
	vMax                      // |
	vLess                     // <
	vMore                     // >
	vEqual                    // =
	vMatch                    // ~
	vJoin                     // ,
	vWithout                  // ^
	vTake                     // #
	vDrop                     // _
	vCast                     // $
	vFind                     // ?
	vApply                    // @
	vApplyN                   // .
	vList                     // (...;...;...)
	vEach                     // ' (adverb)
	vFold                     // / (adverb)
	vScan                     // \ (adverb)
	vIn                       // in
	vSign                     // sign
	vOCount                   // ocount (occurrence count)
	vICount                   // icount (index count)
	vBytes                    // bytes (byte count)
	vAnd                      // and
	vOr                       // or
)

func (v variadic) zero() V {
	switch v {
	case vAdd, vSubtract:
		return NewI(0)
	case vMultiply:
		return NewI(1)
	case vMin:
		return NewI(math.MinInt)
	case vMax:
		return NewI(math.MaxInt)
	}
	return V{}
}

var vStrings = [...]string{
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
	vIn:       "in",
	vList:     "list",
	vEach:     "'",
	vFold:     "/",
	vScan:     "\\",
}

func (v variadic) String() string {
	if v <= vScan {
		return vStrings[v]
	}
	return fmt.Sprintf("{Variadic %d}", v)
}

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
		vList:     {Func: VList},
		vEach:     {Func: VEach, Adverb: true},
		vFold:     {Func: VFold, Adverb: true},
		vScan:     {Func: VScan, Adverb: true},
		vIn:       {Func: VIn},
		vSign:     {Func: VSign},
		vOCount:   {Func: VOCount},
		vICount:   {Func: VICount},
		vBytes:    {Func: VBytes},
		vOr:       {Func: VOr},
		vAnd:      {Func: VAnd},
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
		vList:     "list",
		vEach:     "'",
		vFold:     "/",
		vScan:     "\\",
		vIn:       "in",
		vSign:     "sign",
		vOCount:   "ocount",
		vICount:   "icount",
		vBytes:    "bytes",
		vOr:       "or",
		vAnd:      "and",
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
		x, y := args[1], args[0]
		if x.IsFunction() {
			ctx.push(y)
			y.rcincr()
			r := ctx.applyN(x, 1)
			y.rcdecr()
			if r.IsErr() {
				return r
			}
			return rotate(r, y)
		}
		return maximum(x, y)
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
		x, y := args[1], args[0]
		if x.IsFunction() {
			ctx.push(y)
			y.rcincr()
			r := ctx.applyN(x, 1)
			y.rcdecr()
			if r.IsErr() {
				return r
			}
			return groupBy(r, y)
		}
		return equal(x, y)
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
		return NewI(B2I(Match(args[1], args[0])))
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
		return NewI(int64(Length(args[0])))
	case 2:
		x, y := args[1], args[0]
		if x.IsFunction() {
			ctx.push(y)
			y.rcincr()
			r := ctx.applyN(x, 1)
			y.rcdecr()
			if r.IsErr() {
				return r
			}
			return replicate(r, y)
		}
		return take(x, y)
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
		x, y := args[1], args[0]
		if x.IsFunction() {
			ctx.push(y)
			y.rcincr()
			r := ctx.applyN(x, 1)
			y.rcdecr()
			if r.IsErr() {
				return r
			}
			return weedOut(r, y)
		}
		return drop(x, y)
	default:
		return errRank("_")
	}
}

// VCast implements the $ variadic verb.
func VCast(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return NewS(args[0].Sprint(ctx))
	case 2:
		x, y := args[1], args[0]
		switch x.Value.(type) {
		case array:
			return search(x, y)
		default:
			return cast(x, y)
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
		return NewS(args[0].Type())
	case 2:
		x := args[1]
		ctx.push(args[0])
		return ctx.applyN(x, 1)
	case 3:
		return ctx.amend3(args[2], args[1], args[0])
	case 4:
		return ctx.amend4(args[3], args[2], args[1], args[0])
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
		x := args[1]
		av := toArray(args[0]).Value.(array)
		for i := av.Len() - 1; i >= 0; i-- {
			ctx.push(av.at(i))
		}
		return ctx.applyN(x, av.Len())
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

// VAnd implements the "and" variadic verb.
func VAnd(ctx *Context, args []V) V {
	for _, arg := range args {
		if isFalse(arg) {
			return arg
		}
	}
	return args[0]
}

// VOr implements the "or" variadic verb.
func VOr(ctx *Context, args []V) V {
	for _, arg := range args {
		if isTrue(arg) {
			return arg
		}
	}
	return args[0]
}

// VList implements (x;y;...) array constructor variadic verb.
func VList(ctx *Context, args []V) V {
	xv, cloned := normalize(&AV{Slice: args})
	if cloned {
		r := NewV(xv)
		reverseMut(r)
		return r
	}
	r := cloneArgs(args)
	reverseArgs(r)
	return NewAV(r)
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
	return V{}
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
	return V{}
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
	return V{}
}
