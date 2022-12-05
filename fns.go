package goal

// enum returns !x.
func enum(x V) V {
	x = toIndices(x)
	if x.IsErr() {
		return errf("!x : %v", x)
	}
	if x.IsI() {
		return rangeI(x.I())
	}
	switch xv := x.Value.(type) {
	case *AI:
		return rangeArray(xv)
	default:
		return errs("!x : x nested array")
	}
}

func rangeI(n int64) V {
	if n < 0 {
		return errs("!x : x negative")
	}
	r := make([]int64, n)
	for i := range r {
		r[i] = int64(i)
	}
	return NewAI(r)
}

func rangeArray(x *AI) V {
	cols := int64(1)
	for _, n := range x.Slice {
		if n == 0 {
			return NewAV([]V{})
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
			return errf("&x : x negative (%d)", x.I())
		case x.I() == 0:
			return NewAI([]int64{})
		default:
			r := make([]int64, x.I())
			return NewAI(r)
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return errf("&x : x non-integer (%g)", x.F())
		}
		n := int64(x.F())
		switch {
		case n < 0:
			return errf("&x : x negative (%d)", n)
		case n == 0:
			return NewAI([]int64{})
		default:
			r := make([]int64, n)
			return NewAI(r)
		}

	}
	switch xv := x.Value.(type) {
	case *AB:
		n := int64(0)
		for _, xi := range xv.Slice {
			n += B2I(xi)
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
				return errf("&x : x contains negative integer (%d)", xv)
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
				return errf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return errf("&x : x contains negative (%d)", int64(xi))
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
	case *AV:
		switch aType(xv) {
		case tB, tF, tI:
			n := int64(0)
			for _, xi := range xv.Slice {
				if xi.IsI() {
					if xi.I() < 0 {
						return errf("&x : negative integer (%d)", xi.I())
					}
					n += xi.I()
				} else {
					xif := xi.F()
					if !isI(xif) {
						return errf("&x : not an integer (%g)", xif)
					}
					if xif < 0 {
						return errf("&x : negative integer (%d)", int64(xif))
					}
					n += int64(xif)
				}
			}
			r := make([]int64, 0, n)
			for i, xi := range xv.Slice {
				var max int64
				if xi.IsI() {
					max = xi.I()
				} else {
					max = int64(xi.F())
				}
				for j := int64(0); j < max; j++ {
					r = append(r, int64(i))
				}
			}
			return NewAI(r)
		default:
			return errs("&x : x non-integer array")
		}
	default:
		return errf("&x : x non-integer (type %s)", x.Type())
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	if x.IsI() {
		switch {
		case x.I() < 0:
			return errf("f#y : f[y] negative integer (%d)", x.I())
		default:
			return repeat(y, x.I())
		}
	}
	if x.IsF() {
		if !isI(x.F()) {
			return errf("f#y : f[y] not an integer (%g)", x.F())
		}
		return replicate(NewI(int64(x.F())), y)
	}
	switch xv := x.Value.(type) {
	case *AB:
		if xv.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), Length(y))
		}
		return repeatAB(xv, y)
	case *AI:
		if xv.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", xv.Len(), Length(y))
		}
		return repeatAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return replicate(ix, y)
	case *AV:
		// should be canonical
		//assertCanonical(x)
		return errf("f#y : f[y] non-integer (%s)", x.Type())
	default:
		return errf("f#y : f[y] non-integer (%s)", x.Type())
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
	switch xv := x.Value.(type) {
	case S:
		r := make([]string, n)
		for i := range r {
			r[i] = string(xv)
		}
		return NewAS(r)
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
		n += B2I(xi)
	}
	switch yv := y.Value.(type) {
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
		return canonicalV(NewAV(r))
	default:
		return errf("f#y : y not an array (%s)", y.Type())
	}
}

func repeatAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		if xi < 0 {
			return errf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += xi
	}
	switch yv := y.Value.(type) {
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
		return canonicalV(NewAV(r))
	default:
		return errf("f#y : y not an array (%s)", y.Type())
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	if x.IsI() {
		if x.I() != 0 {
			return NewAV([]V{})
		}
		return y
	}
	if x.IsF() {
		if x.F() != 0 {
			return NewAV([]V{})
		}
		return y
	}
	switch xv := x.Value.(type) {
	case *AB:
		return weedOutAB(xv, y)
	case *AI:
		return weedOutAI(xv, y)
	case *AF:
		ix := toAI(xv)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return weedOut(ix, y)
	case *AV:
		//assertCanonical(x)
		return errf("f#y : f[y] non-integer (%s)", x.Type())
	default:
		return errf("f_y : f[y] non-integer (%s)", x.Type())
	}
}

func weedOutAB(x *AB, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		n += 1 - B2I(xi)
	}
	switch yv := y.Value.(type) {
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
		return canonicalV(NewAV(r))
	default:
		return errf("f_y : y not an array (%s)", y.Type())
	}
}

func weedOutAI(x *AI, y V) V {
	n := int64(0)
	for _, xi := range x.Slice {
		n += B2I(xi == 0)
	}
	switch yv := y.Value.(type) {
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
		return canonicalV(NewAV(r))
	default:
		return errf("f_y : y not an array (%s)", y.Type())
	}
}

// eval implements .s.
func eval(ctx *Context, x V) V {
	//assertCanonical(x)
	nctx := ctx.derive()
	switch xv := x.Value.(type) {
	case S:
		r, err := nctx.Eval(string(xv))
		if err != nil {
			return errf(".s : %v", err)
		}
		ctx.merge(nctx)
		return r
	default:
		return errType(".x", "x", x)
	}
}
