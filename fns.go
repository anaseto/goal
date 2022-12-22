package goal

// enum returns !x.
func enum(x V) V {
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
		return panics("!x : x nested array")
	}
}

func rangeI(n int64) V {
	if n < 0 {
		return panics("!x : x negative")
	}
	r := make([]int64, n)
	for i := range r {
		r[i] = int64(i)
	}
	return newAscUniqAI(r)
}

func rangeArray(x *AI) V {
	cols := int64(1)
	for _, n := range x.Slice {
		if n == 0 {
			return NewAV(nil)
		}
		cols *= n
	}
	r := make([]V, x.Len())
	reps := cols
	for i := range r {
		a := make([]int64, cols)
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
		r[i] = NewAI(a)
	}
	return NewAV(r)
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
		for _, xi := range xv.Slice {
			n += b2i(xi)
		}
		r := make([]int64, 0, n)
		for i, xi := range xv.Slice {
			if xi {
				r = append(r, int64(i))
			}
		}
		return NewAI(r)
	case *AI:
		n := int64(0)
		for _, xi := range xv.Slice {
			if xi < 0 {
				return Panicf("&x : x contains negative integer (%d)", xv)
			}
			n += xi
		}
		r := make([]int64, 0, n)
		for i, xi := range xv.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, int64(i))
			}
		}
		return NewAI(r)
	case *AF:
		n := int64(0)
		for _, xi := range xv.Slice {
			if !isI(xi) {
				return Panicf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return Panicf("&x : x contains negative (%d)", int64(xi))
			}
			n += int64(xi)
		}
		r := make([]int64, 0, n)
		for i, xi := range xv.Slice {
			for j := int64(0); j < int64(xi); j++ {
				r = append(r, int64(i))
			}
		}
		return NewAI(r)
	case array:
		// assertCanonical(xv)
		return panics("&x : x non-integer array")
	default:
		return Panicf("&x : x non-integer (type %s)", x.Type())
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	if x.IsI() {
		switch {
		case x.I() < 0:
			return Panicf("f#y : f[y] negative integer (%d)", x.I())
		default:
			return repeat(y, x.I())
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("f#y : f[y] not an integer (%g)", x.F())
		}
		return replicate(NewI(int64(x.F())), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.Len() != Length(y) {
			return Panicf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), Length(y))
		}
		return repeatAB(xv, y)
	case *AI:
		if xv.Len() != Length(y) {
			return Panicf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), Length(y))
		}
		return repeatAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsPanic() {
			return Panicf("f#y : x %v", ix)
		}
		return replicate(ix, y)
	case *AV:
		//assertCanonical(xv)
		return Panicf("f#y : f[y] non-integer (%s)", x.Type())
	default:
		return Panicf("f#y : f[y] non-integer (%s)", x.Type())
	}
}

func repeat(x V, n int64) V {
	if x.IsI() {
		if isBI(x.I()) {
			r := make([]bool, n)
			for i := range r {
				r[i] = x.I() == 1
			}
			return NewAB(r)
		}
		r := make([]int64, n)
		for i := range r {
			r[i] = x.I()
		}
		return NewAI(r)
	}
	if x.IsF() {
		r := make([]float64, n)
		for i := range r {
			r[i] = float64(x.F())
		}
		return NewAF(r)
	}
	switch xv := x.value.(type) {
	case S:
		r := make([]string, n)
		for i := range r {
			r[i] = string(xv)
		}
		return NewAS(r)
	case *AB:
		r := make([]bool, n*int64(xv.Len()))
		for i, xi := range xv.Slice {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = xi
			}
		}
		return NewAB(r)
	case *AI:
		r := make([]int64, n*int64(xv.Len()))
		for i, xi := range xv.Slice {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = xi
			}
		}
		return NewAI(r)
	case *AF:
		r := make([]float64, n*int64(xv.Len()))
		for i, xi := range xv.Slice {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = xi
			}
		}
		return NewAF(r)
	case *AS:
		r := make([]string, n*int64(xv.Len()))
		for i, xi := range xv.Slice {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = xi
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, n*int64(xv.Len()))
		for i, xi := range xv.Slice {
			in := int64(i) * n
			for j := int64(0); j < n; j++ {
				r[in+j] = xi
			}
		}
		return NewAV(r)
	default:
		r := make([]V, n)
		for i := range r {
			r[i] = x
		}
		return NewAV(r)
	}
}

func repeatAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		n += b2i(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x.Slice {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.Slice {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAF(r)
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.Slice {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAI(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.Slice {
			if xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.Slice {
			if xi {
				r = append(r, yv.at(i))
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("f#y : y not an array (%s)", y.Type())
	}
}

func repeatAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		if xi < 0 {
			return Panicf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += xi
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAF(r)
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAI(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.Slice {
			for j := int64(0); j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("f#y : y not an array (%s)", y.Type())
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	if x.IsI() {
		if x.I() != 0 {
			return NewAV(nil)
		}
		return y
	}
	if x.IsF() {
		if x.F() != 0 {
			return NewAV(nil)
		}
		return y
	}
	switch xv := x.value.(type) {
	case *AB:
		return weedOutAB(xv, y)
	case *AI:
		return weedOutAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsPanic() {
			return Panicf("f#y : x %v", ix)
		}
		return weedOut(ix, y)
	case *AV:
		//assertCanonical(xv)
		return Panicf("f#y : f[y] non-integer (%s)", x.Type())
	default:
		return Panicf("f_y : f[y] non-integer (%s)", x.Type())
	}
}

func weedOutAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		n += 1 - b2i(xi)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x.Slice {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.Slice {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAF(r)
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.Slice {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAI(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.Slice {
			if !xi {
				r = append(r, yv.At(i))
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.Slice {
			if !xi {
				r = append(r, yv.at(i))
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("f_y : y not an array (%s)", y.Type())
	}
}

func weedOutAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		n += b2i(xi == 0)
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x.Slice {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.Slice {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAF(r)
	case *AI:
		r := make([]int64, 0, n)
		for i, xi := range x.Slice {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAI(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.Slice {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.Slice {
			if xi == 0 {
				r = append(r, yv.At(i))
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("f_y : y not an array (%s)", y.Type())
	}
}

func get(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case S:
		return reval(ctx, xv)
	case *errV:
		return xv.V
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
	return r
}

// eval implements eval x.
func eval(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case S:
		nctx := ctx.derive()
		r, err := nctx.Eval(string(xv))
		if err != nil {
			return Panicf(".s : %v", err)
		}
		ctx.merge(nctx)
		return r
	default:
		return Panicf("eval x : x not a string (%s)", x.Type())
	}
}

// evalPackage implements eval[x;y;z].
func evalPackage(ctx *Context, x V, y V, z V) V {
	s, ok := x.value.(S)
	if !ok {
		return Panicf("eval[x;...] : x not a string (%s)", x.Type())
	}
	name, ok := y.value.(S)
	if !ok {
		return Panicf("eval[x;y;...] : y not a string (%s)", y.Type())
	}
	prefix, ok := z.value.(S)
	if !ok {
		return Panicf("eval[x;y;z] : z not a string (%s)", z.Type())
	}
	for i, r := range prefix {
		if i == 0 && !isAlpha(r) || !isAlphaNum(r) {
			return Panicf("eval[x;y;z] : z invalid identifier prefix (%s)", prefix)
		}
	}
	r, err := ctx.EvalPackage(string(s), string(name), string(prefix))
	if err != nil {
		_, ok := err.(ErrPackageImported)
		if ok {
			return NewI(0)
		}
		return Panicf(".s : %v", err)
	}
	return r
}
