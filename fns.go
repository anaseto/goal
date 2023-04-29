package goal

import (
	"math"
	"strings"
)

// enumFieldsKeys returns !x.
func enumFieldsKeys(x V) V {
	if x.IsI() {
		return rangeI(x.I())
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("!i : non-integer i (%g)", x.F())
		}
		return rangeI(int64(x.F()))
	}
	switch xv := x.value.(type) {
	case S:
		return NewAS(strings.Fields(string(xv)))
	case *AB:
		return enumFieldsKeys(fromABtoAI(xv))
	case *AF:
		x := toAI(xv)
		if x.IsPanic() {
			return ppanic("!I : ", x)
		}
		return enumFieldsKeys(x)
	case *AI:
		return rangeArray(xv)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewAS(strings.Fields(xi))
		}
		return NewAV(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := enumFieldsKeys(xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *Dict:
		return xv.Keys()
	default:
		return panicType("!x", "x", x)
	}
}

func rangeI(n int64) V {
	if n < 0 {
		return panics("!i : i negative")
	}
	if n < 256 {
		r := make([]byte, n)
		for i := range r {
			r[i] = byte(i)
		}
		return NewV(&AB{elts: r, flags: flagAscending | flagUnique})
	}
	r := make([]int64, n)
	for i := range r {
		r[i] = int64(i)
	}
	return NewV(&AI{elts: r, flags: flagAscending | flagUnique})
}

func rangeArray(x *AI) V {
	cols := int64(1)
	for _, n := range x.elts {
		if n <= 0 {
			return NewAV(nil)
		}
		if n > math.MaxInt64/cols {
			return panics("!I : too big: overflow")
		}
		cols *= n
	}
	r := make([]V, x.Len())
	reps := cols
	ua := make([]int64, int(cols)*len(r))
	var n int = 2
	for i, xi := range x.elts {
		a := ua[i*int(cols) : (i+1)*int(cols)]
		reps /= xi
		clen := reps * x.At(i)
		for c := int64(0); c < cols/clen; c++ {
			col := c * clen
			for j := int64(0); j < xi; j++ {
				for k := int64(0); k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = NewAIWithRC(a, &n)
	}
	var rn int
	return NewAVWithRC(r, &rn)
}

func rangeII(from, to int64) V {
	if from > to {
		return NewAI(nil)
	}
	if from >= 0 && to < 256 {
		r := make([]byte, to-from)
		for i := range r {
			r[i] = byte(from) + byte(i)
		}
		return NewV(&AB{elts: r, flags: flagAscending | flagUnique})
	}
	r := make([]int64, to-from)
	for i := range r {
		r[i] = from + int64(i)
	}
	return NewV(&AI{elts: r, flags: flagAscending | flagUnique})
}

// where returns &x.
func where(x V) V {
	if x.IsI() {
		switch {
		case x.I() < 0:
			return Panicf("&x : x negative (%d)", x.I())
		case x.I() == 0:
			return NewAI(nil)
		default:
			r := make([]int64, x.I())
			return NewAI(r)
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("&x : x non-integer (%g)", x.F())
		}
		n := int64(x.F())
		switch {
		case n < 0:
			return Panicf("&x : x negative (%d)", n)
		case n == 0:
			return NewAI(nil)
		default:
			r := make([]int64, n)
			return NewAI(r)
		}

	}
	switch xv := x.value.(type) {
	case *AB:
		n := int64(0)
		for _, xi := range xv.elts {
			n += b2I(xi)
		}
		r := make([]int64, n+1)
		j := int64(0)
		for i, xi := range xv.elts {
			r[j] = int64(i)
			j += b2I(xi)
		}
		return NewV(&AI{elts: r[:len(r)-1], rc: reuseRCp(xv.rc), flags: flagAscending})
	case *AI:
		n := int64(0)
		for _, xi := range xv.elts {
			if xi < 0 {
				return Panicf("&x : x contains negative integer (%d)", xv)
			}
			n += xi
		}
		r := make([]int64, 0, n)
		for i, xi := range xv.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, int64(i))
			}
		}
		return NewV(&AI{elts: r, rc: reuseRCp(xv.rc), flags: flagAscending})
	case *AF:
		x := toAI(xv)
		if x.IsPanic() {
			return ppanic("&x : ", x)
		}
		return where(x)
	case S:
		return NewI(int64(len(xv)))
	case *AS:
		r := make([]int64, xv.Len())
		for i, s := range xv.elts {
			r[i] = int64(len(s))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := where(xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return canonicalFast(NewAV(r))
	case *Dict:
		switch xv.values.(type) {
		case *AB:
			r := where(NewV(xv.values))
			if r.IsPanic() {
				return r
			}
			return NewV(xv.keys.atInts(r.value.(*AI).elts))
		case *AI:
			r := where(NewV(xv.values))
			if r.IsPanic() {
				return r
			}
			return NewV(xv.keys.atInts(r.value.(*AI).elts))
		case *AF:
			r := where(NewV(xv.values))
			if r.IsPanic() {
				return r
			}
			return NewV(xv.keys.atInts(r.value.(*AI).elts))
		default:
			return newDictValues(xv.keys, where(NewV(xv.values)))
		}
	default:
		return panicType("&x", "x", x)
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	if x.IsI() {
		switch {
		case x.I() < 0:
			return Panicf("f#y : f[y] negative integer (%d)", x.I())
		default:
			return replicateI(x.I(), y)
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("f#y : non-integer f[y] (%g)", x.F())
		}
		return replicate(NewI(int64(x.F())), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.Len() != y.Len() {
			return Panicf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return replicateAB(xv, y)
	case *AI:
		if xv.Len() != y.Len() {
			return Panicf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return replicateAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsPanic() {
			return Panicf("f#y : x %v", ix)
		}
		return replicate(ix, y)
	default:
		return panicType("f#y", "f[y]", x)
	}
}

func replicateI(n int64, y V) V {
	if y.IsI() {
		if isBI(y.n) {
			r := make([]byte, n)
			for i := range r {
				r[i] = y.n
			}
			var fl flags
			if isbI(y.n) {
				fl |= flagBool
			}
			return NewV(&AB{elts: r, flags: fl})
		}
		r := make([]int64, n)
		for i := range r {
			r[i] = y.n
		}
		return NewAI(r)
	}
	if y.IsF() {
		r := make([]float64, n)
		for i := range r {
			r[i] = float64(y.F())
		}
		return NewAF(r)
	}
	switch yv := y.value.(type) {
	case S:
		r := make([]string, n)
		for i := range r {
			r[i] = string(yv)
		}
		return NewAS(r)
	case *AB:
		r := replicateISlice(n, yv.elts)
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := replicateISlice(n, yv.elts)
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := replicateISlice(n, yv.elts)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := replicateISlice(n, yv.elts)
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := replicateISlice(n, yv.elts)
		*yv.rc += 2
		return NewAVWithRC(r, yv.rc)
	case *Dict:
		keys := replicateI(n, NewV(yv.keys))
		values := replicateI(n, NewV(yv.values))
		return NewDict(keys, values)
	default:
		r := make([]V, n)
		for i := range r {
			r[i] = y
		}
		return NewAV(r)
	}
}

func replicateISlice[T any](n int64, ys []T) []T {
	r := make([]T, n*int64(len(ys)))
	for i, yi := range ys {
		in := int64(i) * n
		for j := int64(0); j < n; j++ {
			r[in+j] = yi
		}
	}
	return r
}

func replicateAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.elts {
		n += b2I(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]byte, 0, n)
		for i, xi := range x.elts {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.elts {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.elts {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.elts {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.elts {
			if xi {
				r = append(r, yv.at(i))
			}
		}
		return NewV(canonicalAV(&AV{elts: r, rc: yv.rc}))
	case *Dict:
		keys := replicateAB(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := replicateAB(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewDict(keys, values)
	default:
		return panicType("f#y", "y", y)
	}
}

func replicateAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.elts {
		if xi < 0 {
			return Panicf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += xi
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]byte, 0, n)
		for i, xi := range x.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.elts {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewV(canonicalAV(&AV{elts: r, rc: yv.rc}))
	case *Dict:
		keys := replicateAI(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := replicateAI(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewDict(keys, values)
	default:
		return panicType("f#y", "y", y)
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	if x.IsI() {
		if x.I() != 0 {
			return NewAV(nil)
		}
		return toArray(y)
	}
	if x.IsF() {
		if x.F() != 0 {
			return NewAV(nil)
		}
		return toArray(y)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.Len() != y.Len() {
			return Panicf("f_y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return weedOutAB(xv, y)
	case *AI:
		if xv.Len() != y.Len() {
			return Panicf("f_y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return weedOutAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsPanic() {
			return Panicf("f#y : x %v", ix)
		}
		return weedOut(ix, y)
	default:
		return panicType("f_y", "f[y]", x)
	}
}

func weedOutAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.elts {
		n += 1 - b2I(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]byte, 0, n)
		for i, xi := range x.elts {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.elts {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.elts {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.elts {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.elts {
			if !xi {
				r = append(r, yv.at(i))
			}
		}
		return NewV(canonicalAV(&AV{elts: r, rc: yv.rc}))
	case *Dict:
		keys := weedOutAB(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := weedOutAB(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewDict(keys, values)
	default:
		return panicType("f_y", "y", y)
	}
}

func weedOutAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.elts {
		n += b2I(xi == 0)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]byte, 0, n)
		for i, xi := range x.elts {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.elts {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.elts {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.elts {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.elts {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewV(canonicalAV(&AV{elts: r, rc: yv.rc}))
	case *Dict:
		keys := weedOutAI(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := weedOutAI(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewDict(keys, values)
	default:
		return panicType("f_y", "y", y)
	}
}

// get implements .x.
func get(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case S:
		return reval(ctx, xv)
	case *errV:
		return xv.V
	case *Dict:
		return xv.Values()
	case array:
		return NewV(&Dict{keys: xv, values: xv})
	default:
		return panicType(".x", "x", x)
	}
}

// reval implements .s.
func reval(ctx *Context, s S) V {
	nctx := NewContext()
	r, err := nctx.Eval(string(s))
	if err != nil {
		return Panicf(".s : %v", err)
	}
	return recompileLambdas(ctx, nctx, r)
}

func recompileLambdas(ctx, nctx *Context, x V) V {
	if x.kind == valLambda {
		return evalString(ctx, x.Sprint(nctx))
	}
	if x.kind != valBoxed {
		return x
	}
	switch xv := x.value.(type) {
	case S:
		return x
	case *AB:
		return x
	case *AI:
		return x
	case *AF:
		return x
	case *AS:
		return x
	case *nReplacer:
		return x
	case *replacer:
		return x
	case *rx:
		return x
	case *Dict:
		ks := recompileLambdas(ctx, nctx, xv.Keys())
		if ks.IsPanic() {
			return ks
		}
		xv.keys = ks.value.(array)
		vs := recompileLambdas(ctx, nctx, xv.Values())
		if vs.IsPanic() {
			return vs
		}
		xv.values = vs.value.(array)
		return x
	case *errV:
		xv.V = recompileLambdas(ctx, nctx, xv.V)
		if xv.V.IsPanic() {
			return xv.V
		}
		return x
	case *derivedVerb:
		xv.Arg = recompileLambdas(ctx, nctx, xv.Arg)
		if xv.Arg.IsPanic() {
			return xv.Arg
		}
		return x
	case *projection:
		xv.Fun = recompileLambdas(ctx, nctx, xv.Fun)
		if xv.Fun.IsPanic() {
			return xv.Fun
		}
		for i, arg := range xv.Args {
			xi := recompileLambdas(ctx, nctx, arg)
			if xi.IsPanic() {
				return xi
			}
			xv.Args[i] = xi
		}
		return x
	case *projectionMonad:
		xv.Fun = recompileLambdas(ctx, nctx, xv.Fun)
		if xv.Fun.IsPanic() {
			return xv.Fun
		}
		return x
	case *projectionFirst:
		xv.Fun = recompileLambdas(ctx, nctx, xv.Fun)
		if xv.Fun.IsPanic() {
			return xv.Fun
		}
		xv.Arg = recompileLambdas(ctx, nctx, xv.Arg)
		if xv.Arg.IsPanic() {
			return xv.Arg
		}
		return x
	case *rxReplacer:
		xv.repl = recompileLambdas(ctx, nctx, xv.repl)
		if xv.repl.IsPanic() {
			return xv.repl
		}
		return x
	case *AV:
		for i, xi := range xv.elts {
			xv.elts[i] = recompileLambdas(ctx, nctx, xi)
		}
		return x
	default:
		return Panicf(".s : unsupported return value type (%s)", x.Type())
	}
}

// eval implements eval x.
func eval(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case S:
		return evalString(ctx, string(xv))
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := evalString(ctx, string(xi))
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return Canonical(NewAV(r))
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := eval(ctx, xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return Canonical(NewAV(r))
	default:
		return panicType("eval x", "x", x)
	}
}

func evalString(ctx *Context, s string) V {
	if ctx.fname == "" {
		osource := ctx.sources[""]
		defer func() {
			ctx.sources[""] = osource
		}()
	}
	nctx := ctx.derive()
	r, err := nctx.Eval(s)
	ctx.merge(nctx)
	if err != nil {
		return Panicf("eval s : %v", err)
	}
	return r
}

// evalPackage implements eval[s;loc;pfx].
func evalPackage(ctx *Context, x V, y V, z V) V {
	s, ok := x.value.(S)
	if !ok {
		return panicType("eval[s;loc;pfx]", "s", x)
	}
	loc, ok := y.value.(S)
	if !ok {
		return panicType("eval[s;loc;pfx]", "loc", y)
	}
	pfx, ok := z.value.(S)
	if !ok {
		return panicType("eval[s;loc;pfx]", "pfx", z)
	}
	for i, r := range pfx {
		if i == 0 && !isAlpha(r) || !isAlphaNum(r) {
			return Panicf("eval[s;loc;pfx] : non-identifier prefix (%s)", pfx)
		}
	}
	if ctx.fname == "" {
		osource := ctx.sources[""]
		defer func() {
			ctx.sources[""] = osource
		}()
	}
	r, err := ctx.EvalPackage(string(s), string(loc), string(pfx))
	if err != nil {
		_, ok := err.(ErrPackageImported)
		if ok {
			return NewI(0)
		}
		return Panicf("eval[s;loc;pfx] : %v", err)
	}
	return r
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).value.(array)
	if av.Len() == 0 {
		return panics(".[f1;x;f2] : empty x")
	}
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	r := f1.applyN(ctx, av.Len())
	if r.IsPanic() {
		r = NewS(string(r.value.(panicV)))
		ctx.replaceTop(r)
		r = f2.applyN(ctx, 1)
		if r.IsPanic() {
			ctx.drop()
			return Panicf(".[f1;x;f2] : f2 call: %v", r)
		}
	}
	ctx.drop()
	return r
}

// tryAt implements @[f1;x;f2].
func tryAt(ctx *Context, f1, x, f2 V) V {
	r := ctx.Apply(f1, x)
	if r.IsPanic() {
		r = NewS(string(r.value.(panicV)))
		ctx.replaceTop(r)
		r = f2.applyN(ctx, 1)
		if r.IsPanic() {
			return Panicf("@[f1;x;f2] : f2 call: %v", r)
		}
	}
	return r
}

func getN(y V) V {
	var n int64 = 1
	if y.IsI() {
		n = y.I()
	} else if y.IsF() {
		if !isI(y.F()) {
			return Panicf(`goal["time";x;n] : non-integer n (%g)`, y.F())
		}
		n = int64(y.F())
	} else {
		return panicType(`goal["time";x;n]`, "n", y)
	}
	return NewI(n)
}
