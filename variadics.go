package goal

import (
	"fmt"
	"math"
)

// VariadicFun represents a variadic function, either a verb or an adverb.
type VariadicFun struct {
	Adverb bool
	Func   func(*Context, []V) V
}

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
)

var vFuns []VariadicFun

func init() {
	vFuns = []VariadicFun{
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
	}
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
	vList:     "list",
	vEach:     "'",
	vFold:     "/",
	vScan:     "\\",
}

func (v variadic) String() string {
	if int(v) <= len(vStrings)-1 {
		return vStrings[v]
	}
	return fmt.Sprintf("{Variadic %d}", v)
}

func (ctx *Context) initVariadics() {
	const size = 32
	ctx.variadics = make([]VariadicFun, len(vFuns), size)
	copy(ctx.variadics, vFuns[:])
	ctx.variadicsNames = make([]string, len(vStrings), size)
	copy(ctx.variadicsNames, vStrings[:])
	ctx.vNames = make(map[string]variadic, size)
	for v, s := range ctx.variadicsNames {
		ctx.vNames[s] = variadic(v)
	}
	ctx.vNames["::"] = vRight
	ctx.keywords = map[string]NameType{}
	ctx.RegisterMonad("abs", VariadicFun{Func: VAbs})
	ctx.RegisterMonad("bytes", VariadicFun{Func: VBytes})
	ctx.RegisterMonad("ceil", VariadicFun{Func: VCeil})
	ctx.RegisterMonad("error", VariadicFun{Func: VError})
	ctx.RegisterMonad("eval", VariadicFun{Func: VEval})
	ctx.RegisterMonad("firsts", VariadicFun{Func: VFirsts})
	ctx.RegisterMonad("icount", VariadicFun{Func: VICount})
	ctx.RegisterMonad("ocount", VariadicFun{Func: VOCount})
	ctx.RegisterMonad("sign", VariadicFun{Func: VSign})
	ctx.RegisterDyad("and", VariadicFun{Func: VAnd})
	ctx.RegisterDyad("in", VariadicFun{Func: VIn})
	ctx.RegisterDyad("or", VariadicFun{Func: VOr})
	ctx.RegisterDyad("rotate", VariadicFun{Func: VRotate})
	ctx.RegisterDyad("shift", VariadicFun{Func: VShift})
	ctx.RegisterDyad("rshift", VariadicFun{Func: VRShift})
}

type zeroFun interface {
	function
	zero() V
}

func (v variadic) zero() V {
	switch v {
	case vAdd, vSubtract:
		return NewI(0)
	case vMultiply:
		return NewI(1)
	case vMin:
		return NewI(math.MinInt64)
	case vMax:
		return NewI(math.MaxInt64)
	default:
		return NewI(0)
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
		return panicRank(":")
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
		return panicRank("+")
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
		return panicRank("-")
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
		return panicRank("*")
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
		return panicRank("%")
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
		return panicRank("!")
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
		return panicRank("&")
	}
}

// VMax implements the | variadic verb.
func VMax(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return reverse(args[0])
	case 2:
		return maximum(args[1], args[0])
	default:
		return panicRank("|")
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
		return panicRank("<")
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
		return panicRank(">")
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
			r := ctx.applyN(x, 1)
			if r.IsPanic() {
				return r
			}
			return groupBy(r, y)
		}
		return equal(x, y)
	default:
		return panicRank("=")
	}
}

// VMatch implements the ~ variadic verb.
func VMatch(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return not(args[0])
	case 2:
		return NewI(b2i(Match(args[1], args[0])))
	default:
		return panicRank("~")
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
		return panicRank(",")
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
		return panicRank("^")
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
			r := ctx.applyN(x, 1)
			if r.IsPanic() {
				return r
			}
			return replicate(r, y)
		}
		return take(x, y)
	default:
		return panicRank("#")
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
			r := ctx.applyN(x, 1)
			if r.IsPanic() {
				return r
			}
			return weedOut(r, y)
		}
		return drop(x, y)
	default:
		return panicRank("_")
	}
}

// VCast implements the $ variadic verb.
func VCast(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return NewS(args[0].Sprint(ctx))
	case 2:
		x, y := args[1], args[0]
		if x.IsI() || x.IsF() {
			return shapeSplit(x, y)
		}
		switch xv := x.value.(type) {
		case array:
			return search(x, y)
		case S:
			return cast(xv, y)
		default:
			return panicType("x$y", "x", x)
		}
	default:
		return panicRank("$")
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
		return panicRank("?")
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
		return panicRank("@")
	}
}

// VApplyN implements the . variadic verb.
func VApplyN(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return get(ctx, args[0])
	case 2:
		x := args[1]
		av := toArray(args[0]).value.(array)
		for i := av.Len() - 1; i >= 0; i-- {
			ctx.push(av.at(i))
		}
		return ctx.applyN(x, av.Len())
	case 3:
		return try(ctx, args[2], args[1], args[0])
	default:
		return panicRank(".")
	}
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
		return panicRank("'")
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
		return panicRank("/")
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
		return panicRank("\\")
	}
	return V{}
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

// VAbs implements the "abs" variadic verb.
func VAbs(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return abs(args[0])
	default:
		return panicRank("abs")
	}
}

// VBytes implements the "bytes" variadic verb.
func VBytes(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return bytes(args[0])
	default:
		return panicRank("icount")
	}
}

// VCeil implements the "ceil" variadic verb.
func VCeil(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return ceil(args[0])
	default:
		return panicRank("ceil")
	}
}

// VError implements the "error" variadic verb.
func VError(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		x := args[0]
		if x.IsError() {
			return panics("error x : x is already an error")
		}
		return NewError(x)
	default:
		return panicRank("error")
	}
}

// VEval implements the "eval" variadic verb.
func VEval(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return eval(ctx, args[0])
	default:
		return panicRank("eval")
	}
}

// VFirsts implements the "firsts" variadic verb.
func VFirsts(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return markFirsts(args[0])
	default:
		return panicRank("firsts")
	}
}

// VICount implements the "icount" variadic verb.
func VICount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return icount(args[0])
	default:
		return panicRank("icount")
	}
}

// VIn implements the "in" variadic verb.
func VIn(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("in : got only one argument")
	case 2:
		return memberOf(args[1], args[0])
	default:
		return panicRank("in")
	}
}

// VOCount implements the "ocount" variadic verb.
func VOCount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return occurrenceCount(args[0])
	default:
		return panicRank("ocount")
	}
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

// VSign implements the "sign" variadic verb.
func VSign(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sign(args[0])
	default:
		return panicRank("sign")
	}
}

// VRotate implements the rotate variadic verb.
func VRotate(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("rotate : got only one argument")
	case 2:
		return rotate(args[1], args[0])
	default:
		return panicRank("rotate")
	}
}

// VShift implements the shift variadic verb.
func VShift(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return nudgeBack(args[0])
	case 2:
		return shiftAfter(args[1], args[0])
	default:
		return panicRank("shift")
	}
}

// VRShift implements the rshift variadic verb.
func VRShift(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return nudge(args[0])
	case 2:
		return shiftBefore(args[1], args[0])
	default:
		return panicRank("rshift")
	}
}
