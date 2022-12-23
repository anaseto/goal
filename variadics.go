package goal

import (
	"fmt"
	"math"
)

// VariadicFun represents a variadic function.
type VariadicFun func(*Context, []V) V

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
		vRight:    VRight,
		vAdd:      VAdd,
		vSubtract: VSubtract,
		vMultiply: VMultiply,
		vDivide:   VDivide,
		vMod:      VMod,
		vMin:      VMin,
		vMax:      VMax,
		vLess:     VLess,
		vMore:     VMore,
		vEqual:    VEqual,
		vMatch:    VMatch,
		vJoin:     VJoin,
		vWithout:  VWithout,
		vTake:     VTake,
		vDrop:     VDrop,
		vCast:     VCast,
		vFind:     VFind,
		vApply:    VApply,
		vApplyN:   VApplyN,
		vList:     VList,
		vEach:     VEach,
		vFold:     VFold,
		vScan:     VScan,
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
	ctx.variadics = append(ctx.variadics, VSet)
	ctx.variadicsNames = append(ctx.variadicsNames, "::")
	ctx.vNames["::"] = variadic(len(ctx.variadics) - 1)
	ctx.keywords = map[string]NameType{}
	// monads
	ctx.RegisterMonad("abs", VAbs)
	ctx.RegisterMonad("bytes", VBytes)
	ctx.RegisterMonad("ceil", VCeil)
	ctx.RegisterMonad("error", VError)
	ctx.RegisterMonad("eval", VEval)
	ctx.RegisterMonad("firsts", VFirsts)
	ctx.RegisterMonad("icount", VICount)
	ctx.RegisterMonad("ocount", VOCount)
	ctx.RegisterMonad("panic", VPanic)
	ctx.RegisterMonad("rx", VRx)
	ctx.RegisterMonad("seed", VSeed)
	ctx.RegisterMonad("sign", VSign)
	ctx.RegisterMonad("sub", VSub)

	// math monads
	ctx.RegisterMonad("acos", VAcos)
	ctx.RegisterMonad("asin", VAsin)
	ctx.RegisterMonad("atan", VAtan)
	ctx.RegisterMonad("cos", VCos)
	ctx.RegisterMonad("exp", VExp)
	ctx.RegisterMonad("log", VLog)
	ctx.RegisterMonad("round", VRoundToEven)
	ctx.RegisterMonad("sin", VSin)
	ctx.RegisterMonad("sqrt", VSqrt)
	ctx.RegisterMonad("tan", VTan)
	ctx.RegisterDyad("nan", VNaN)

	// dyads
	ctx.RegisterDyad("and", VAnd)
	ctx.RegisterDyad("csv", VCSV)
	ctx.RegisterDyad("in", VIn)
	ctx.RegisterDyad("or", VOr)
	ctx.RegisterDyad("rotate", VRotate)
	ctx.RegisterDyad("shift", VShift)
	ctx.RegisterDyad("rshift", VRShift)
	ctx.RegisterDyad("time", VTime)
}

type zeroFun interface {
	function
	zero() V
}

func (v variadic) zero() V {
	switch v {
	case vMultiply:
		return NewI(1)
	case vMin:
		return NewF(math.Inf(1))
	case vMax:
		return NewF(math.Inf(-1))
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
		x := args[0]
		if x.IsI() || x.IsF() {
			return uniform(ctx, x)
		}
		return uniq(x)
	case 2:
		x, y := args[1], args[0]
		if x.IsI() || x.IsF() {
			return roll(ctx, x, y)
		}
		return find(x, y)
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
	case 1:
		return panics("' : not enough arguments")
	case 2:
		return each2(ctx, args)
	case 3:
		return each3(ctx, args)
	default:
		return panicRank("'")
	}
}

// VFold implements the / variadic adverb.
func VFold(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("/ : not enough arguments")
	case 2:
		return fold2(ctx, args)
	case 3:
		return fold3(ctx, args)
	default:
		return panicRank("/")
	}
}

// VScan implements the \ variadic adverb.
func VScan(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("\\ : not enough arguments")
	case 2:
		return scan2(ctx, args[1], args[0])
	case 3:
		return scan3(ctx, args)
	default:
		return panicRank("\\")
	}
}

// VAnd implements the "and" variadic verb.
func VAnd(ctx *Context, args []V) V {
	for i := len(args) - 1; i > 0; i-- {
		if isFalse(args[i]) {
			return args[i]
		}
	}
	return args[0]
}

// VCSV implements the "csv" variadic verb.
func VCSV(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return fCSV(',', args[0])
	case 2:
		return fCSV2(args[1], args[0])
	default:
		return panicRank("csv")
	}
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
	case 2:
		return evalPackage(ctx, args[1], args[0], NewS(""))
	case 3:
		return evalPackage(ctx, args[2], args[1], args[0])
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

// VPanic implements the "panic" variadic verb.
func VPanic(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		x := args[0]
		switch xv := x.value.(type) {
		case S:
			return panics(string(xv))
		default:
			return panicType("panic x", "x", x)
		}
	default:
		return panicRank("panic")
	}
}

// VOr implements the "or" variadic verb.
func VOr(ctx *Context, args []V) V {
	for i := len(args) - 1; i > 0; i-- {
		if isTrue(args[i]) {
			return args[i]
		}
	}
	return args[0]
}

// VSeed implements the "seed" variadic verb.
func VSeed(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return seed(ctx, args[0])
	default:
		return panicRank("seed")
	}
}

// VSet implements the "set" variadic verb.
func VSet(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		name, ok := args[0].value.(S)
		if !ok {
			return panicType(":: x", "x", args[0])
		}
		r, ok := ctx.GetGlobal(string(name))
		if !ok {
			return Panicf(":: x : undefined variable (%s)", name)
		}
		return r
	case 2:
		name, ok := args[1].value.(S)
		if !ok {
			return panicType("::[x;y]", "x", args[1])
		}
		ctx.AssignGlobal(string(name), args[0])
		return args[0]
	default:
		return panicRank("::")
	}
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

// VSub implements the sub variadic verb.
func VSub(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sub1(args[0])
	case 2:
		return sub2(args[1], args[0])
	case 3:
		return sub3(args[2], args[1], args[0])
	default:
		return panicRank("sub")
	}
}
