package goal

import (
	"strings"
	"time"
)

// VariadicFun represents a variadic function. The array of arguments is in
// stack order: the first argument is its last element.
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
		vRight:    vfRight,
		vAdd:      vfAdd,
		vSubtract: vfSubtract,
		vMultiply: vfMultiply,
		vDivide:   vfDivide,
		vKey:      vfKey,
		vMin:      vfMin,
		vMax:      vfMax,
		vLess:     vfLess,
		vMore:     vfMore,
		vEqual:    vfEqual,
		vMatch:    vfMatch,
		vJoin:     vfJoin,
		vWithout:  vfWithout,
		vTake:     vfTake,
		vDrop:     vfDrop,
		vCast:     vfCast,
		vFind:     vfFind,
		vApply:    vfApply,
		vApplyN:   vfApplyN,
		vList:     vfList,
		vQq:       vfQq,
		vEach:     vfEach,
		vFold:     vfFold,
		vScan:     vfScan,
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
	const size = 80
	const nkeywords = 50
	ctx.variadics = make([]VariadicFun, len(vFuns), size)
	copy(ctx.variadics, vFuns[:])
	ctx.variadicsNames = make([]string, len(vStrings), size)
	copy(ctx.variadicsNames, vStrings[:])
	ctx.vNames = make(map[string]variadic, size)
	for v, s := range ctx.variadicsNames {
		ctx.vNames[s] = variadic(v)
	}
	ctx.variadics = append(ctx.variadics, vfSet)
	ctx.variadicsNames = append(ctx.variadicsNames, "::")
	ctx.vNames["::"] = variadic(len(ctx.variadics) - 1)
	ctx.keywords = make(map[string]IdentType, nkeywords)

	// special form variadics
	ctx.registerVariadic("icount", vfICount)

	// monads
	ctx.RegisterMonad("abs", vfAbs)
	ctx.RegisterMonad("bytes", vfBytes)
	ctx.RegisterMonad("ceil", vfCeil)
	ctx.RegisterMonad("error", vfError)
	ctx.RegisterMonad("eval", vfEval)
	ctx.RegisterMonad("firsts", vfFirsts)
	ctx.RegisterMonad("json", vfjson)
	ctx.RegisterMonad("ocount", vfOCount)
	ctx.RegisterMonad("panic", vfPanic)
	ctx.RegisterMonad("rx", vfRx)
	ctx.RegisterMonad("sign", vfSign)
	ctx.RegisterMonad("utf8.rcount", vfUTF8RCount)
	ctx.RegisterMonad("utf8.valid", vfUTF8Valid)

	// math monads
	ctx.RegisterMonad("acos", vfAcos)
	ctx.RegisterMonad("asin", vfAsin)
	ctx.RegisterMonad("atan", vfAtan)
	ctx.RegisterMonad("cos", vfCos)
	ctx.RegisterMonad("exp", vfExp)
	ctx.RegisterMonad("log", vfLog)
	ctx.RegisterMonad("round", vfRoundToEven)
	ctx.RegisterMonad("sin", vfSin)
	ctx.RegisterMonad("sqrt", vfSqrt)
	ctx.RegisterMonad("tan", vfTan)
	ctx.RegisterDyad("nan", vfNaN)

	// dyads
	ctx.RegisterDyad("and", vfAnd)
	ctx.RegisterDyad("csv", vfCSV)
	ctx.RegisterDyad("in", vfIn)
	ctx.RegisterDyad("or", vfOr)
	ctx.RegisterDyad("mod", vfMod)
	ctx.RegisterDyad("rotate", vfRotate)
	v := ctx.RegisterDyad("rshift", vfRShift)
	ctx.vNames["»"] = v.variadic()
	v = ctx.RegisterDyad("shift", vfShift)
	ctx.vNames["«"] = v.variadic()
	ctx.RegisterDyad("sub", vfSub)
	ctx.RegisterDyad("time", vfTime)

	// runtime functions
	ctx.RegisterMonad("rt.vars", vfRTVars)
	ctx.RegisterMonad("rt.prec", vfRTPrec)
	ctx.RegisterMonad("rt.seed", vfRTSeed)
	ctx.RegisterMonad("rt.time", vfRTTime)
}

// vfRight implements the : variadic verb.
func vfRight(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return args[0]
	case 2:
		return args[0]
	default:
		return panicRank(":")
	}
}

// vfAdd implements the + variadic verb.
func vfAdd(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return flip(args[0])
	case 2:
		return add(args[1], args[0])
	default:
		return panicRank("+")
	}
}

// vfSubtract implements the - variadic verb.
func vfSubtract(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return negate(args[0])
	case 2:
		return subtract(args[1], args[0])
	default:
		return panicRank("-")
	}
}

// vfMultiply implements the * variadic verb.
func vfMultiply(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return first(args[0])
	case 2:
		return multiply(args[1], args[0])
	default:
		return panicRank("*")
	}
}

// vfDivide implements the % variadic verb.
func vfDivide(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return classify(ctx, args[0])
	case 2:
		return divide(args[1], args[0])
	default:
		return panicRank("%")
	}
}

// vfKey implements the ! variadic verb.
func vfKey(ctx *Context, args []V) V {
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

// vfMin implements the & variadic verb.
func vfMin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return where(args[0])
	case 2:
		return minimum(args[1], args[0])
	default:
		return panicRank("&")
	}
}

// vfMax implements the | variadic verb.
func vfMax(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return reverse(args[0])
	case 2:
		return maximum(args[1], args[0])
	default:
		return panicRank("|")
	}
}

// vfLess implements the < variadic verb.
func vfLess(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return ascend(args[0])
	case 2:
		return lesser(args[1], args[0])
	default:
		return panicRank("<")
	}
}

// vfMore implements the > variadic verb.
func vfMore(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return descend(args[0])
	case 2:
		return greater(args[1], args[0])
	default:
		return panicRank(">")
	}
}

// vfEqual implements the = variadic verb.
func vfEqual(ctx *Context, args []V) V {
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

// vfMatch implements the ~ variadic verb.
func vfMatch(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return not(args[0])
	case 2:
		return NewI(B2I(args[1].Matches(args[0])))
	default:
		return panicRank("~")
	}
}

// vfJoin implements the , variadic verb.
func vfJoin(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return enlist(args[0])
	case 2:
		return joinTo(args[1], args[0])
	default:
		return panicRank(",")
	}
}

// vfWithout implements the ^ variadic verb.
func vfWithout(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sortUp(args[0])
	case 2:
		return without(args[1], args[0])
	default:
		return panicRank("^")
	}
}

// vfTake implements the # variadic verb.
func vfTake(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return NewI(int64(args[0].Len()))
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

// vfDrop implements the _ variadic verb.
func vfDrop(ctx *Context, args []V) V {
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

// vfCast implements the $ variadic verb.
func vfCast(ctx *Context, args []V) V {
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

// vfFind implements the ? variadic verb.
func vfFind(ctx *Context, args []V) V {
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

// vfApply implements the @ variadic verb.
func vfApply(ctx *Context, args []V) V {
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

// vfApplyN implements the . variadic verb.
func vfApplyN(ctx *Context, args []V) V {
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

// vfList implements (x;y;...) array constructor variadic verb.
func vfList(ctx *Context, args []V) V {
	xav := &AV{elts: args}
	xv, cloned := normalize(xav)
	if cloned {
		r := NewV(xv)
		return r
	}
	xav.elts = cloneArgs(args)
	return NewV(xav)
}

// vfQq implements qq/STRING/ interpolation variadic verb.
func vfQq(ctx *Context, args []V) V {
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

// vfEach implements the ' variadic adverb.
func vfEach(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("' : not enough arguments")
	case 2:
		return each2(ctx, args[1], args[0])
	default:
		return eachN(ctx, args)
	}
}

// vfFold implements the / variadic adverb.
func vfFold(ctx *Context, args []V) V {
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

// vfScan implements the \ variadic adverb.
func vfScan(ctx *Context, args []V) V {
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

// vfAnd implements the "and" variadic verb.
func vfAnd(ctx *Context, args []V) V {
	for i := len(args) - 1; i > 0; i-- {
		if args[i].IsFalse() {
			return args[i]
		}
	}
	return args[0]
}

// vfCSV implements the "csv" variadic verb.
func vfCSV(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return fCSV(',', args[0])
	case 2:
		return fCSV2(args[1], args[0])
	default:
		return panicRank("csv")
	}
}

// vfjson implements the "json" variadic verb.
func vfjson(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return fJSON(args[0])
	default:
		return panicRank("json")
	}
}

// vfAbs implements the "abs" variadic verb.
func vfAbs(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return abs(args[0])
	default:
		return panicRank("abs")
	}
}

// vfBytes implements the "bytes" variadic verb.
func vfBytes(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return bytecount(args[0])
	default:
		return panicRank("bytes")
	}
}

// vfCeil implements the "ceil" variadic verb.
func vfCeil(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return ceil(args[0])
	default:
		return panicRank("ceil")
	}
}

// vfError implements the "error" variadic verb.
func vfError(ctx *Context, args []V) V {
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

// vfEval implements the "eval" variadic verb.
func vfEval(ctx *Context, args []V) V {
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

// vfFirsts implements the "firsts" variadic verb.
func vfFirsts(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return markFirsts(ctx, args[0])
	default:
		return panicRank("firsts")
	}
}

// vfICount implements the "icount" variadic verb (#'=).
func vfICount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return icount(args[0])
	default:
		return panicRank("icount")
	}
}

// vfIn implements the "in" variadic verb.
func vfIn(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("in : not enough arguments")
	case 2:
		return memberOf(args[1], args[0])
	default:
		return panicRank("in")
	}
}

// vfOCount implements the "ocount" variadic verb.
func vfOCount(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return occurrenceCount(ctx, args[0])
	default:
		return panicRank("ocount")
	}
}

// vfMod implements the mod variadic verb.
func vfMod(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("mod : not enough arguments")
	case 2:
		return modulus(args[1], args[0])
	default:
		return panicRank("mod")
	}
}

// vfPanic implements the "panic" variadic verb.
func vfPanic(ctx *Context, args []V) V {
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

// vfOr implements the "or" variadic verb.
func vfOr(ctx *Context, args []V) V {
	for i := len(args) - 1; i > 0; i-- {
		if args[i].IsTrue() {
			return args[i]
		}
	}
	return args[0]
}

// vfSet implements the "set" variadic verb.
func vfSet(ctx *Context, args []V) V {
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

// vfSign implements the "sign" variadic verb.
func vfSign(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return sign(args[0])
	default:
		return panicRank("sign")
	}
}

// vfRotate implements the rotate variadic verb.
func vfRotate(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return panics("rotate : not enough arguments")
	case 2:
		return rotate(args[1], args[0])
	default:
		return panicRank("rotate")
	}
}

// vfShift implements the shift variadic verb.
func vfShift(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return nudgeBack(args[0])
	case 2:
		return shiftAfter(args[1], args[0])
	default:
		return panicRank("shift")
	}
}

// vfRShift implements the rshift variadic verb.
func vfRShift(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return nudge(args[0])
	case 2:
		return shiftBefore(args[1], args[0])
	default:
		return panicRank("rshift")
	}
}

// vfSub implements the sub variadic verb.
func vfSub(ctx *Context, args []V) V {
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

// vfRTPrec implements the rt.prec variadic verb.
func vfRTPrec(ctx *Context, args []V) V {
	if len(args) > 1 {
		return panicRank(`rt.prec`)
	}
	x := args[0]
	if x.IsI() {
		ctx.prec = int(x.I())
	} else if x.IsF() {
		if !isI(x.F()) {
			return Panicf(`rt.prec i : non-integer i (%g)`, x.F())
		}
		ctx.prec = int(x.F())
	} else {
		return Panicf(`rt.prec i : i bad type (%s)`, x.Type())
	}
	return NewI(1)
}

// vfRTSeed implements the rt.seed variadic verb.
func vfRTSeed(ctx *Context, args []V) V {
	if len(args) > 1 {
		return panicRank(`rt.seed`)
	}
	return seed(ctx, args[0])
}

// vfRTVars implements the rt.vars variadic verb.
func vfRTVars(ctx *Context, args []V) V {
	if len(args) > 1 {
		return panicRank(`rt.vars`)
	}
	x := args[0]
	cmd, ok := x.value.(S)
	if !ok {
		return panicType("rt.vars s", "s", x)
	}
	switch cmd {
	case "":
		v := cloneArgs(ctx.globals)
		k := make([]string, len(ctx.gNames))
		copy(k, ctx.gNames)
		return NewDict(NewAS(k), Canonical(NewAV(v)))
	case "f":
		v := []V{}
		k := []string{}
		for i, x := range ctx.globals {
			if x.IsFunction() {
				v = append(v, x)
				k = append(k, ctx.gNames[i])
			}
		}
		return NewDict(NewAS(k), Canonical(NewAV(v)))
	case "v":
		v := []V{}
		k := []string{}
		for i, x := range ctx.globals {
			if !x.IsFunction() {
				v = append(v, x)
				k = append(k, ctx.gNames[i])
			}
		}
		return NewDict(NewAS(k), Canonical(NewAV(v)))
	default:
		return Panicf("rt.vars s : invalid value (%s)", cmd)
	}
}

// vfRTTime implements the rt.time variadic verb.
func vfRTTime(ctx *Context, args []V) V {
	x := args[len(args)-1]
	var n int64 = 1
	switch xv := x.value.(type) {
	case S:
		if len(args) > 2 {
			return panicRank(`rt.time[s;n]`)
		}
		if len(args) == 2 {
			n = getN(args[0]).I()
		}
		t := time.Now()
		for i := int64(0); i < n; i++ {
			r := evalString(ctx, string(xv))
			if r.IsPanic() {
				return r
			}
		}
		d := time.Since(t)
		return NewI(int64(d) / n)
	default:
		if !x.IsFunction() {
			return panicType(`rt.time[x;n]`, "x", x)
		}
		if len(args) == 1 {
			return panics(`rt.time[f;x;n] : not enough arguments`)
		}
		y := args[len(args)-2]
		if len(args) > 3 {
			return panicRank(`rt.time[f;x;n]`)
		}
		if len(args) == 3 {
			n = getN(args[0]).I()
		}
		x.IncrRC()
		av := toArray(y).value.(array)
		av.IncrRC()
		t := time.Now()
		for i := int64(0); i < n; i++ {
			if av.Len() == 0 {
				continue
			}
			for i := av.Len() - 1; i >= 0; i-- {
				ctx.push(av.at(i))
			}
			r := x.applyN(ctx, av.Len())
			if r.IsPanic() {
				x.DecrRC()
				av.DecrRC()
				ctx.drop()
				return r
			}
			ctx.drop()
		}
		x.DecrRC()
		av.DecrRC()
		d := time.Since(t)
		return NewI(int64(d) / n)
	}
}
