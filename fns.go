package goal

import (
	"fmt"
	"math"
	"strings"
)

// enum returns !x.
func enum(x V) V {
	if x.IsI() {
		return enumI(x.I())
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("!i : non-integer i (%g)", x.F())
		}
		return enumI(int64(x.F()))
	}
	switch xv := x.bv.(type) {
	case S:
		return NewAS(strings.Fields(string(xv)))
	case *AB:
		return odometer(xv.elts)
	case *AF:
		x := toAI(xv)
		if x.IsPanic() {
			return ppanic("!I : ", x)
		}
		return enum(x)
	case *AI:
		return odometer(xv.elts)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewV(&AS{elts: strings.Fields(xi), flags: flagImmutable})
		}
		return newAVu(r)
	case *AV:
		return mapAV(xv, enum)
	case *D:
		return xv.Keys()
	default:
		return panicType("!x", "x", x)
	}
}

func enumI(n int64) V {
	if n < 0 {
		r := make([]int64, -n)
		for i := range r {
			r[i] = n + int64(i)
		}
		return NewV(&AI{elts: r, flags: flagAscending | flagDistinct})
	}
	if n < 256 {
		r := permRange[byte](int(n))
		return NewV(&AB{elts: r, flags: flagAscending | flagDistinct})
	}
	r := permRange[int64](int(n))
	return NewV(&AI{elts: r, flags: flagAscending | flagDistinct})
}

func odometer[I integer](x []I) V {
	cols := int64(1)
	bsize := true
	for _, n := range x {
		if n <= 0 {
			return protoAV()
		}
		if int64(n) > math.MaxInt64/cols {
			return panics("!I : too big: overflow")
		}
		if int64(n) >= 256 {
			bsize = false
		}
		cols *= int64(n)
	}
	if bsize {
		a := odometerWithCols[I, byte](x, cols)
		r := make([]V, len(x))
		for i := range r {
			ai := a[i*int(cols) : (i+1)*int(cols)]
			r[i] = NewV(&AB{elts: ai, flags: flagImmutable})
		}
		return newAVu(r)
	}
	a := odometerWithCols[I, int64](x, cols)
	r := make([]V, len(x))
	for i := range r {
		ai := a[i*int(cols) : (i+1)*int(cols)]
		r[i] = NewV(&AI{elts: ai, flags: flagImmutable})
	}
	return newAVu(r)
}

func odometerWithCols[I integer, J integer](x []I, cols int64) []J {
	reps := cols
	a := make([]J, int(cols)*len(x))
	for i, xi := range x {
		ai := a[i*int(cols) : (i+1)*int(cols)]
		reps /= int64(xi)
		clen := reps * int64(x[i])
		for c := int64(0); c < cols/clen; c++ {
			col := c * clen
			for j := int64(0); j < int64(xi); j++ {
				for k := int64(0); k < reps; k++ {
					ai[col+j*reps+k] = J(j)
				}
			}
		}
	}
	return a
}

// where returns &x.
func where(x V) V {
	if x.IsI() {
		switch {
		case x.I() <= 0:
			return newABb(nil)
		default:
			r := make([]byte, x.I())
			return newABb(r)
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("&x : x non-integer (%g)", x.F())
		}
		return where(NewI(int64(x.F())))
	}
	switch xv := x.bv.(type) {
	case *AB:
		return whereAB(xv)
	case *AI:
		return whereAI(xv)
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
		return cmapAV(xv, where)
	case *D:
		if xv.values.numeric() {
			r := where(NewV(xv.values))
			if r.IsPanic() {
				return r
			}
			return NewV(arrayAtIv(xv.keys, r))
		}
		return newDictValues(xv.keys, where(NewV(xv.values)))
	default:
		return panicType("&x", "x", x)
	}
}

func whereAB(xv *AB) V {
	if xv.IsBoolean() {
		if xv.Len() < 256 {
			r := whereBools[byte](xv.elts)
			return NewV(&AB{elts: r, flags: flagAscending})
		}
		r := whereBools[int64](xv.elts)
		return NewV(&AI{elts: r, flags: flagAscending})
	}
	if xv.Len() < 256 {
		r := whereBytes[byte](xv.elts)
		return NewV(&AB{elts: r, flags: flagAscending})
	}
	r := whereBytes[int64](xv.elts)
	return NewV(&AI{elts: r, flags: flagAscending})
}

func whereAI(xv *AI) V {
	if xv.Len() < 256 {
		r := whereIs[byte](xv.elts)
		return NewV(&AB{elts: r, flags: flagAscending})
	}
	r := whereIs[int64](xv.elts)
	return NewV(&AI{elts: r, flags: flagAscending})
}

func whereBools[I integer](x []byte) []I {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	r := make([]I, n+1)
	n = 0
	for i, xi := range x {
		r[n] = I(i)
		n += int64(xi)
	}
	return r[:len(r)-1]
}

func whereBytes[I integer](x []byte) []I {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	r := make([]I, n)
	n = 0
	for i, xi := range x {
		for j := byte(0); j < xi; j++ {
			r[n] = I(i)
			n++
		}
	}
	return r
}

func whereIs[I integer](x []int64) []I {
	var n int64
	for _, xi := range x {
		if xi < 0 {
			continue
		}
		n += xi
	}
	r := make([]I, n)
	n = 0
	for i, xi := range x {
		for j := int64(0); j < xi; j++ {
			r[n] = I(i)
			n++
		}
	}
	return r
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
	switch xv := x.bv.(type) {
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
		if isBI(y.uv) {
			r := make([]byte, n)
			for i := range r {
				r[i] = byte(y.uv)
			}
			var fl flags
			if isbI(y.uv) {
				fl |= flagBool
			}
			return NewV(&AB{elts: r, flags: fl})
		}
		r := make([]int64, n)
		for i := range r {
			r[i] = y.uv
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
	switch yv := y.bv.(type) {
	case S:
		r := make([]string, n)
		for i := range r {
			r[i] = string(yv)
		}
		return NewAS(r)
	case *AB:
		r := replicateISlice(n, yv.elts)
		return NewAB(r)
	case *AI:
		r := replicateISlice(n, yv.elts)
		return NewAI(r)
	case *AF:
		r := replicateISlice(n, yv.elts)
		return NewAF(r)
	case *AS:
		r := replicateISlice(n, yv.elts)
		return NewAS(r)
	case *AV:
		r := replicateISlice(n, yv.elts)
		return newAVu(r)
	case *D:
		keys := replicateI(n, NewV(yv.keys))
		values := replicateI(n, NewV(yv.values))
		return NewD(keys, values)
	default:
		r := make([]V, n)
		y.MarkImmutable()
		for i := range r {
			r[i] = y
		}
		return newAVu(r)
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
	switch yv := y.bv.(type) {
	case *AB:
		if x.IsBoolean() {
			r := replicateBools(x.elts, yv.elts)
			return NewV(&AB{elts: r, flags: yv.flags &^ flagImmutable})
		}
		r := replicateBytes(x.elts, yv.elts)
		var fl flags
		if yv.IsBoolean() {
			fl = flagBool
		}
		return NewV(&AB{elts: r, flags: fl})
	case *AI:
		if x.IsBoolean() {
			r := replicateBools(x.elts, yv.elts)
			return NewV(&AI{elts: r, flags: yv.flags &^ flagImmutable})
		}
		r := replicateBytes(x.elts, yv.elts)
		return NewAI(r)
	case *AF:
		if x.IsBoolean() {
			r := replicateBools(x.elts, yv.elts)
			return NewV(&AF{elts: r, flags: yv.flags &^ flagImmutable})
		}
		r := replicateBytes(x.elts, yv.elts)
		return NewAF(r)
	case *AS:
		if x.IsBoolean() {
			r := replicateBools(x.elts, yv.elts)
			return NewV(&AS{elts: r, flags: yv.flags &^ flagImmutable})
		}
		r := replicateBytes(x.elts, yv.elts)
		return NewAS(r)
	case *AV:
		if x.IsBoolean() {
			r := replicateBools(x.elts, yv.elts)
			return canonicalAV(&AV{elts: r, flags: yv.flags &^ flagImmutable})
		}
		r := replicateBytes(x.elts, yv.elts)
		return canonicalVs(r)
	case *D:
		keys := replicateAB(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := replicateAB(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewD(keys, values)
	default:
		return panicType("f#y", "y", y)
	}
}

func replicateBools[T any](x []byte, y []T) []T {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	if n == int64(len(y)) {
		return y
	}
	r := make([]T, n)
	n = 0
	for i, xi := range x {
		if xi > 0 {
			r[n] = y[i]
			n++
		}
	}
	return r
}

func replicateBytes[T any](x []byte, y []T) []T {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	r := make([]T, n)
	n = 0
	for i, xi := range x {
		for j := byte(0); j < xi; j++ {
			r[n] = y[i]
			n++
		}
	}
	return r
}

func replicateAI(x *AI, y V) V {
	switch yv := y.bv.(type) {
	case *AB:
		r, err := replicateIs(x.elts, yv.elts)
		if err != nil {
			return panicErr(err)
		}
		var fl flags
		if yv.IsBoolean() {
			fl = flagBool
		}
		return NewV(&AB{elts: r, flags: fl})
	case *AI:
		r, err := replicateIs(x.elts, yv.elts)
		if err != nil {
			return panicErr(err)
		}
		return NewAI(r)
	case *AF:
		r, err := replicateIs(x.elts, yv.elts)
		if err != nil {
			return panicErr(err)
		}
		return NewAF(r)
	case *AS:
		r, err := replicateIs(x.elts, yv.elts)
		if err != nil {
			return panicErr(err)
		}
		return NewAS(r)
	case *AV:
		r, err := replicateIs(x.elts, yv.elts)
		if err != nil {
			return panicErr(err)
		}
		return canonicalVs(r)
	case *D:
		keys := replicateAI(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := replicateAI(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewD(keys, values)
	default:
		return panicType("f#y", "y", y)
	}
}

func replicateIs[T any](x []int64, y []T) ([]T, error) {
	var n int64
	for _, xi := range x {
		if xi < 0 {
			return nil, fmt.Errorf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += int64(xi)
	}
	r := make([]T, n)
	n = 0
	for i, xi := range x {
		for j := int64(0); j < xi; j++ {
			r[n] = y[i]
			n++
		}
	}
	return r, nil
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	if x.IsI() {
		if x.I() != 0 {
			return protoArrayForV(y)
		}
		return toArray(y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf(`f_y : non-integer f[y] (%g)`, x.F())
		}
		return weedOut(NewI(int64(x.F())), y)
	}
	switch xv := x.bv.(type) {
	case *AB:
		if xv.Len() != y.Len() {
			return Panicf("f_y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return weedOutAIs((*A[byte])(xv), y)
	case *AI:
		if xv.Len() != y.Len() {
			return Panicf("f_y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), y.Len())
		}
		return weedOutAIs((*A[int64])(xv), y)
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

func weedOutAIs[I integer](x *A[I], y V) V {
	switch yv := y.bv.(type) {
	case *AB:
		r := weedOutIs(x.elts, yv.elts)
		return NewV(&AB{elts: r, flags: yv.flags &^ flagImmutable})
	case *AI:
		r := weedOutIs(x.elts, yv.elts)
		return NewV(&AI{elts: r, flags: yv.flags &^ flagImmutable})
	case *AF:
		r := weedOutIs(x.elts, yv.elts)
		return NewV(&AF{elts: r, flags: yv.flags &^ flagImmutable})
	case *AS:
		r := weedOutIs(x.elts, yv.elts)
		return NewV(&AS{elts: r, flags: yv.flags &^ flagImmutable})
	case *AV:
		r := weedOutIs(x.elts, yv.elts)
		return canonicalAV(&AV{elts: r, flags: yv.flags &^ flagImmutable})
	case *D:
		keys := weedOutAIs(x, NewV(yv.keys))
		if keys.IsPanic() {
			return keys
		}
		values := weedOutAIs(x, NewV(yv.values))
		if values.IsPanic() {
			return values
		}
		return NewD(keys, values)
	default:
		return panicType("f_y", "y", y)
	}
}

func weedOutIs[I integer, T any](x []I, y []T) []T {
	var n int64
	for _, xi := range x {
		n += b2I(xi == 0)
	}
	r := make([]T, n)
	n = 0
	for i, xi := range x {
		if xi == 0 {
			r[n] = y[i]
			n++
		}
	}
	return r
}

// get implements .x.
func get(ctx *Context, x V) V {
	switch xv := x.bv.(type) {
	case S:
		return reval(ctx, xv)
	case *errV:
		return xv.V
	case *D:
		return xv.Values()
	case Array:
		return NewV(&D{keys: xv, values: xv})
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
	switch xv := x.bv.(type) {
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
	case *D:
		ks := recompileLambdas(ctx, nctx, xv.Keys())
		if ks.IsPanic() {
			return ks
		}
		xv.keys = ks.bv.(Array)
		vs := recompileLambdas(ctx, nctx, xv.Values())
		if vs.IsPanic() {
			return vs
		}
		xv.values = vs.bv.(Array)
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
	switch xv := x.bv.(type) {
	case S:
		return evalString(ctx, string(xv))
	case *AS:
		return cdoN(xv.Len(), func(i int) V { return evalString(ctx, xv.At(i)) })
	case *AV:
		return cmapAV(xv, func(xi V) V { return eval(ctx, xi) })
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
	s, ok := x.bv.(S)
	if !ok {
		return panicType("eval[s;loc;pfx]", "s", x)
	}
	loc, ok := y.bv.(S)
	if !ok {
		return panicType("eval[s;loc;pfx]", "loc", y)
	}
	pfx, ok := z.bv.(S)
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
	av := toArray(x).bv.(Array)
	if av.Len() == 0 {
		return panics(".[f1;x;f2] : empty x")
	}
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.VAt(i))
	}
	r := f1.applyN(ctx, av.Len())
	if r.IsPanic() {
		r = NewS(string(r.bv.(panicV)))
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
		r = NewS(string(r.bv.(panicV)))
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
