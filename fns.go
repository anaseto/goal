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
	return NewV(r)
}

func rangeArray(x AI) V {
	cols := 1
	for _, n := range x {
		if n == 0 {
			return NewV(AV{})
		}
		cols *= n
	}
	r := make([]V, x.Len())
	reps := cols
	for i := range r {
		a := make([]int, cols)
		reps /= x[i]
		clen := reps * x[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < x[i]; j++ {
				for k := 0; k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = NewV(a)
	}
	return NewV(r)
}

// where returns &x.
func where(x V) V {
	if x.IsInt() {
		switch {
		case x.Int() < 0:
			return errf("&x : x negative (%d)", x.Int())
		case x.Int() == 0:
			return NewV(AI{})
		default:
			r := make([]int, x.Int())
			return NewV(r)
		}
	}
	switch x := x.Value.(type) {
	case F:
		if !isI(x) {
			return errf("&x : x non-integer (%g)", x)
		}
		n := int(x)
		switch {
		case n < 0:
			return errf("&x : x negative (%d)", n)
		case n == 0:
			return NewV(AI{})
		default:
			r := make([]int, n)
			return NewV(r)
		}
	case *AB:
		n := 0
		for _, xi := range x {
			n += int(B2I(xi))
		}
		r := make([]int, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, i)
			}
		}
		return NewV(r)
	case *AI:
		n := 0
		for _, xi := range x {
			if xi < 0 {
				return errf("&x : x contains negative integer (%d)", x)
			}
			n += xi
		}
		r := make([]int, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, i)
			}
		}
		return NewV(r)
	case *AF:
		n := 0
		for _, xi := range x {
			if !isI(F(xi)) {
				return errf("&x : x contains non-integer (%g)", xi)
			}
			if xi < 0 {
				return errf("&x : x contains negative (%d)", int(xi))
			}
			n += int(xi)
		}
		r := make([]int, 0, n)
		for i, xi := range x {
			for j := 0; j < int(xi); j++ {
				r = append(r, i)
			}
		}
		return NewV(r)
	case *AV:
		switch aType(x) {
		case tB, tF, tI:
			n := 0
			for _, xi := range x {
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
			for i, xi := range x {
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
			return NewV(r)
		default:
			return errs("&x : x non-integer")
		}
	default:
		return errs("&x : x non-integer")
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
	switch x := x.Value.(type) {
	case F:
		if !isI(x) {
			return errf("f#y : f[y] not an integer (%g)", x)
		}
		n := int(x)
		switch {
		case n < 0:
			return errf("f#y : f[y] negative (%d)", n)
		default:
			return repeat(y, n)
		}
	case *AB:
		if x.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", x.Len(), Length(y))
		}
		return repeatAB(x, y)
	case *AI:
		if x.Len() != Length(y) {
			return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", x.Len(), Length(y))
		}
		return repeatAI(x, y)
	case *AF:
		ix := toAI(x)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return replicate(ix, y)
	case *AV:
		// should be canonical
		assertCanonical(x)
		return errs("f#y : f[y] non-integer")
	default:
		return errs("f#y : f[y] non-integer")
	}
}

func repeat(x V, n int) V {
	if x.IsInt() {
		if isBI(x.Int()) {
			r := make([]bool, n)
			for i := range r {
				r[i] = x.Int() == 1
			}
			return NewV(r)
		}
		r := make([]int, n)
		for i := range r {
			r[i] = x.Int()
		}
		return NewV(r)
	}
	switch xv := x.Value.(type) {
	case F:
		r := make([]float64, n)
		for i := range r {
			r[i] = float64(xv)
		}
		return NewV(r)
	case S:
		r := make([]string, n)
		for i := range r {
			r[i] = string(xv)
		}
		return NewV(r)
	default:
		r := make([]V, n)
		for i := range r {
			r[i] = x
		}
		return NewV(r)
	}
}

func repeatAB(x AB, y V) V {
	n := 0
	for _, xi := range x {
		n += int(B2I(xi))
	}
	switch y := y.Value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AI:
		r := make([]int, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x {
			if xi {
				r = append(r, y.at(i))
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f#y : y not an array")
	}
}

func repeatAI(x AI, y V) V {
	n := 0
	for _, xi := range x {
		if xi < 0 {
			return errf("f#y : f[y] contains negative integer (%d)", xi)
		}
		n += xi
	}
	switch y := y.Value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AI:
		r := make([]int, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x {
			for j := 0; j < xi; j++ {
				r = append(r, y[i])
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f#y : y not an array")
	}
}

// weedOut implements {x}_y
func weedOut(x, y V) V {
	if x.IsInt() {
		if x.Int() != 0 {
			return NewV(AV{})
		}
		return y
	}
	switch x := x.Value.(type) {
	case F:
		if x != 0 {
			return NewV(AV{})
		}
		return y
	case *AB:
		return weedOutAB(x, y)
	case *AI:
		return weedOutAI(x, y)
	case *AF:
		ix := toAI(x)
		if ix.IsErr() {
			return errf("f#y : x %v", ix)
		}
		return weedOut(ix, y)
	case *AV:
		//assertCanonical(x)
		return errs("f#y : f[y] non-integer")
	default:
		return errs("f_y : f[y] non-integer")
	}
}

func weedOutAB(x AB, y V) V {
	n := 0
	for _, xi := range x {
		n += 1 - int(B2I(xi))
	}
	switch y := y.Value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AI:
		r := make([]int, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x {
			if !xi {
				r = append(r, y.at(i))
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f_y : y not an array")
	}
}

func weedOutAI(x AI, y V) V {
	n := 0
	for _, xi := range x {
		n += int(B2I(xi == 0))
	}
	switch y := y.Value.(type) {
	case *AB:
		r := make([]bool, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AF:
		r := make([]float64, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AI:
		r := make([]int, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AS:
		r := make([]string, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(r)
	case *AV:
		r := make([]V, 0, n)
		for i, xi := range x {
			if xi == 0 {
				r = append(r, y[i])
			}
		}
		return NewV(canonical(r))
	default:
		return errs("f_y : y not an array")
	}
}

// eval implements .s.
func eval(ctx *Context, x V) V {
	//assertCanonical(x)
	nctx := ctx.derive()
	switch x := x.Value.(type) {
	case S:
		r, err := nctx.Eval(string(x))
		if err != nil {
			return errf(".s : %v", err)
		}
		ctx.merge(nctx)
		return r
	default:
		return errType(".x", "x", x)
	}
}
