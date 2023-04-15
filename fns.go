package goal

// enum returns !x.
func enum(x V) V {
	d, ok := x.value.(*Dict)
	if ok {
		return d.Keys()
	}
	x = toIndices(x)
	if x.IsPanic() {
		return Panicf("!x : %v", x)
	}
	if x.IsI() {
		return rangeI(x.I())
	}
	switch xv := x.value.(type) {
	case *AI:
		return rangeArray(xv)
	default:
		return panicType("!x", "x", x)
	}
}

func rangeI(n int64) V {
	if n < 0 {
		return panics("!i : i negative")
	}
	r := make([]int64, n)
	for i := range r {
		r[i] = int64(i)
	}
	return NewV(&AI{elts: r, flags: flagAscending | flagUnique})
}

func rangeII(from, to int64) V {
	if from > to {
		return NewAI(nil)
	}
	r := make([]int64, to-from)
	for i := range r {
		r[i] = from + int64(i)
	}
	return NewV(&AI{elts: r, flags: flagAscending | flagUnique})
}

func rangeArray(x *AI) V {
	cols := int64(1)
	for _, n := range x.elts {
		if n == 0 {
			return NewAV(nil)
		}
		cols *= n
	}
	r := make([]V, x.Len())
	reps := cols
	ua := make([]int64, int(cols)*len(r))
	var n int = 2
	for i := range r {
		a := ua[i*int(cols) : (i+1)*int(cols)]
		reps /= x.At(i)
		clen := reps * x.At(i)
		for c := int64(0); c < cols/clen; c++ {
			col := c * clen
			for j := int64(0); j < x.At(i); j++ {
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
			n += B2I(xi)
		}
		r := make([]int64, n+1)
		j := int64(0)
		for i, xi := range xv.elts {
			r[j] = int64(i)
			j += B2I(xi)
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
		n := int64(0)
		for _, xi := range xv.elts {
			if !isI(xi) {
				return Panicf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return Panicf("&x : x contains negative (%d)", int64(xi))
			}
			n += int64(xi)
		}
		r := make([]int64, 0, n)
		for i, xi := range xv.elts {
			for j := int64(0); j < int64(xi); j++ {
				r = append(r, int64(i))
			}
		}
		return NewV(&AI{elts: r, rc: reuseRCp(xv.rc), flags: flagAscending})
	case *Dict:
		r := where(NewV(xv.values))
		if r.IsPanic() {
			return r
		}
		return NewV(xv.keys.atIndices(r.value.(*AI)))
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
		if isBI(y.I()) {
			r := make([]bool, n)
			for i := range r {
				r[i] = y.I() == 1
			}
			return NewAB(r)
		}
		r := make([]int64, n)
		for i := range r {
			r[i] = y.I()
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
		r := make([]bool, n*int64(yv.Len()))
		for i, yi := range yv.elts {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = yi
			}
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, n*int64(yv.Len()))
		for i, yi := range yv.elts {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = yi
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AF:
		r := make([]float64, n*int64(yv.Len()))
		for i, yi := range yv.elts {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = yi
			}
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AS:
		r := make([]string, n*int64(yv.Len()))
		for i, yi := range yv.elts {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = yi
			}
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		r := make([]V, n*int64(yv.Len()))
		for i, yi := range yv.elts {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = yi
			}
		}
		return NewAVWithRC(r, reuseRCp(yv.rc))
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

func replicateAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.elts {
		n += B2I(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
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
		r := make([]bool, 0, n)
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
		n += 1 - B2I(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
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
		n += B2I(xi == 0)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
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
