package goal

import (
	"strings"
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
	vKey                      // !
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
	vQq                       // qq/STRING/
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
		vKey:      VKey,
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
		vQq:       VQq,
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
	vKey:      "!",
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
	vQq:       "qq",
	vEach:     "'",
	vFold:     "/",
	vScan:     "\\",
}

func (ctx *Context) initVariadics() {
	const size = 64
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
	ctx.keywords = make(map[string]NameType, 32)

	// special form variadics
	ctx.registerVariadic("icount", VICount)

	// monads
	ctx.RegisterMonad("abs", VAbs)
	ctx.RegisterMonad("bytes", VBytes)
	ctx.RegisterMonad("ceil", VCeil)
	ctx.RegisterMonad("error", VError)
	ctx.RegisterMonad("eval", VEval)
	ctx.RegisterMonad("firsts", VFirsts)
	ctx.RegisterMonad("json", VJSON)
	ctx.RegisterMonad("ocount", VOCount)
	ctx.RegisterMonad("panic", VPanic)
	ctx.RegisterMonad("rx", VRx)
	ctx.RegisterMonad("sign", VSign)
	ctx.RegisterMonad("utf8.rcount", VUTF8RCount)
	ctx.RegisterMonad("utf8.valid", VUTF8Valid)

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
	ctx.RegisterDyad("mod", VMod)
	ctx.RegisterDyad("rotate", VRotate)
	v := ctx.RegisterDyad("rshift", VRShift)
	ctx.vNames["»"] = v.variadic()
	v = ctx.RegisterDyad("shift", VShift)
	ctx.vNames["«"] = v.variadic()
	ctx.RegisterDyad("sub", VSub)
	ctx.RegisterDyad("time", VTime)
	ctx.RegisterDyad("goal", VGoal)
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
		return classify(ctx, args[0])
	case 2:
		return divide(args[1], args[0])
	default:
		return panicRank("%")
	}
}

// VKey implements the ! variadic verb.
func VKey(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enum(args[0])
	case 2:
		x, y := args[1], args[0]
		if x.IsI() || x.IsF() {
			return shapeSplit(x, y)
		}
		return dict(args[1], args[0])
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
			r := x.applyN(ctx, 1)
			ctx.drop()
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
			r := x.applyN(ctx, 1)
			ctx.drop()
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
			r := x.applyN(ctx, 1)
			ctx.drop()
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
		if x.IsI() {
			return padStrings(int(x.I()), y)
		} else if x.IsF() {
			if !isI(x.F()) {
				return Panicf("i$y : non-integer i (%g)", x.F())
			}
			return padStrings(int(x.F()), y)
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
		return uniq(ctx, x)
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
		r := x.applyN(ctx, 1)
		ctx.drop()
		return r
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
		if av.Len() == 0 {
			return x
		}
		for i := av.Len() - 1; i >= 0; i-- {
			ctx.push(av.at(i))
		}
		r := x.applyN(ctx, av.Len())
		ctx.drop()
		return r
	case 3:
		x := args[2]
		if x.IsFunction() {
			return try(ctx, x, args[1], args[0])
		}
		return ctx.deepAmend3(x, args[1], args[0])
	case 4:
		return ctx.deepAmend4(args[3], args[2], args[1], args[0])
	default:
		return panicRank(".")
	}
}

// VList implements (x;y;...) array constructor variadic verb.
func VList(ctx *Context, args []V) V {
	xav := &AV{Slice: args}
	xv, cloned := normalize(xav)
	if cloned {
		r := NewV(xv)
		return r
	}
	xav.Slice = cloneArgs(args)
	return NewV(xav)
}

// VQq implements qq/STRING/ interpolation variadic verb.
func VQq(ctx *Context, args []V) V {
	n := 0
	for _, arg := range args {
		s, ok := arg.value.(S)
		if !ok {
			continue
		}
		n += len(s)
	}
	var sb strings.Builder
	if n > 0 {
		sb.Grow(n)
	}
	for _, arg := range args {
		s, ok := arg.value.(S)
		if !ok {
			sb.WriteString(arg.Sprint(ctx))
		} else {
			sb.WriteString(string(s))
		}
	}
	return NewS(sb.String())
}

// VEach implements the ' variadic adverb.
func VEach(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("' : not enough arguments")
	case 2:
		return each2(ctx, args[1], args[0])
	default:
		return eachN(ctx, args)
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
		return foldN(ctx, args)
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
		return scanN(ctx, args)
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

// VJSON implements the "json" variadic verb.
func VJSON(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return fJSON(args[0])
	default:
		return panicRank("json")
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
		return bytecount(args[0])
	default:
		return panicRank("bytes")
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
		return markFirsts(ctx, args[0])
	default:
		return panicRank("firsts")
	}
}

// VICount implements the "icount" variadic verb (#'=).
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
		return panics("in : not enough arguments")
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
		return occurrenceCount(ctx, args[0])
	default:
		return panicRank("ocount")
	}
}

// VMod implements the mod variadic verb.
func VMod(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("mod : not enough arguments")
	case 2:
		return modulus(args[1], args[0])
	default:
		return panicRank("mod")
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
			return Panicf(":: x : undefined global (%s)", name)
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
		return panics("rotate : not enough arguments")
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

// VGoal implements the goal variadic verb.
func VGoal(ctx *Context, args []V) V {
	x := args[len(args)-1]
	cmd, ok := x.value.(S)
	if !ok {
		return panicType("goal[cmd;...]", "cmd", x)
	}
	switch cmd {
	case "globals":
		if len(args) != 1 {
			return panicRank(`"globals" goal`)
		}
		v := cloneArgs(ctx.globals)
		k := make([]string, len(ctx.gNames))
		copy(k, ctx.gNames)
		return NewDict(NewAS(k), Canonical(NewAV(v)))
	case "prec":
		if len(args) != 2 {
			return panicRank(`"prec" goal`)
		}
		y := args[0]
		if y.IsI() {
			ctx.prec = int(y.I())
		} else if y.IsF() {
			if !isI(y.F()) {
				return Panicf(`goal["prec";n]: non-integer n (%g)`, y.F())
			}
			ctx.prec = int(y.F())
		} else {
			return Panicf(`goal["prec";n]: n bad type (%s)`, y.Type())
		}
		return NewI(1)
	case "seed":
		if len(args) != 2 {
			return panicRank(`"seed" goal`)
		}
		return seed(ctx, args[0])
	case "time":
		return goalTime(ctx, args[:len(args)-1])
	default:
		return Panicf("goal[cmd;...]: invalid cmd (%s)", cmd)
	}
}
