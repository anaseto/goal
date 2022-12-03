package goal

// enum returns !x.
func enum(x V) V {
	x = toIndices(x)
	if x.IsErr() {
		return errf("!x : %v", x)
	}
	if x.IsInt() {
		return rangeI(x.Int())
	}
	switch xv := x.Value.(type) {
	case *AI:
		return rangeArray(xv)
	default:
		return errs("!x : x nested array")
	}
}

func rangeI(n int) V {
	if n < 0 {
		return errs("!x : x negative")
	}
	r := make([]int, n)
	for i := range r {
		r[i] = i
	}
	return NewAI(r)
}

func rangeArray(x *AI) V {
	cols := 1
	for _, n := range x.Slice {
		if n == 0 {
			return NewAV([]V{})
		}
		cols *= n
	}
	r := make([]V, x.Len())
	reps := cols
	for i := range r {
		a := make([]int, cols)
		reps /= x.At(i)
		clen := reps * x.At(i)
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < x.At(i); j++ {
				for k := 0; k < reps; k++ {
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
	if x.IsInt() {
		switch {
		case x.Int() < 0:
			return errf("&x : x negative (%d)", x.Int())
		case x.Int() == 0:
			return NewAI([]int{})
		default:
			r := make([]int, x.Int())
			return NewAI(r)
		}
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("&x : x non-integer (%g)", xv)
		}
		n := int(xv)
		switch {
		case n < 0:
			return errf("&x : x negative (%d)", n)
		case n == 0:
			return NewAI([]int{})
		default:
			r := make([]int, n)
			return NewAI(r)
		}
	case *AB:
		n := 0
		for _, xi := range xv.Slice {
			n += int(B2I(xi))
		}
		r := make([]int, 0, n)
		for i, xi := range xv.Slice {
			if xi {
				r = append(r, i)
			}
		}
		return NewAI(r)
	case *AI:
		n := 0
		for _, xi := range xv.Slice {
			if xi < 0 {
				return errf("&x : x contains negative integer (%d)", xv)
			}
			n += xi
		}
		r := make([]int, 0, n)
		for i, xi := range xv.Slice {
			for j := 0; j < xi; j++ {
				r = append(r, i)
			}
		}
		return NewAI(r)
	case *AF:
		n := 0
		for _, xi := range xv.Slice {
			if !isI(F(xi)) {
				return errf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return errf("&x : x contains negative (%d)", int(xi))
			}
			n += int(xi)
		}
		r := make([]int, 0, n)
		for i, xi := range xv.Slice {
			for j := 0; j < int(xi); j++ {
				r = append(r, i)
			}
		}
		return NewAI(r)
	case *AV:
		switch aType(xv) {
		case tB, tF, tI:
			n := 0
			for _, xi := range xv.Slice {
				if xi.IsInt() {
					if xi.Int() < 0 {
						return errf("&x : negative integer (%d)", xi.Int())
					}
					n += int(xi.Int())
				} else {
					xif := xi.F()
					if !isI(xif) {
						return errf("&x : not an integer (%g)", xif)
					}
					if xif < 0 {
						return errf("&x : negative integer (%d)", int(xif))
					}
					n += int(xif)
				}
			}
			r := make([]int, 0, n)
			for i, xi := range xv.Slice {
				var max int
				if xi.IsInt() {
					max = xi.Int()
				} else {
					max = int(xi.F())
				}
				for j := 0; j < int(max); j++ {
					r = append(r, i)
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
	if x.IsInt() {
		switch {
		case x.Int() < 0:
			return errf("f#y : f[y] negative integer (%d)", x.Int())
		default:
			return repeat(y, x.Int())
		}
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("f#y : f[y] not an integer (%g)", xv)
		}
		n := int(xv)
		switch {
		case n < 0:
			return errf("f#y : f[y] negative (%d)", n)
		default:
			return repeat(y, n)
		}
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

func repeat(x V, n int) V {
	if x.IsInt() {
		if isBI(x.Int()) {
			r := make([]bool, n)
			for i := range r {
				r[i] = x.Int() == 1
			}
			return NewAB(r)
		}
		r := make([]int, n)
		for i := range r {
			r[i] = x.Int()
		}
		return NewAI(r)
	}
	switch xv := x.Value.(type) {
	case F:
		r := make([]float64, n)
		for i := range r {
			r[i] = float64(xv)
		}
		return NewAF(r)
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
	n := 0
	for _, xi := range x.Slice {
		n += int(B2I(xi))
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
		r := make([]int, 0, n)
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
	n := 0
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
			for j := 0; j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x.Slice {
			for j := 0; j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAF(r)
	case *AI:
		r := make([]int, 0, n)
		for i, xi := range x.Slice {
			for j := 0; j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAI(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x.Slice {
			for j := 0; j < xi; j++ {
				r = append(r, yv.At(i))
			}
		}
		return NewAS(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x.Slice {
			for j := 0; j < xi; j++ {
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
	if x.IsInt() {
		if x.Int() != 0 {
			return NewAV([]V{})
		}
		return y
	}
	switch xv := x.Value.(type) {
	case F:
		if xv != 0 {
			return NewAV([]V{})
		}
		return y
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
	n := 0
	for _, xi := range x.Slice {
		n += 1 - int(B2I(xi))
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
		r := make([]int, 0, n)
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
	n := 0
	for _, xi := range x.Slice {
		n += int(B2I(xi == 0))
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
		r := make([]int, 0, n)
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
