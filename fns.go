package goal

// enum returns !x.
func enum(x V) V {
	switch x := x.(type) {
	case F:
		if !isI(x) {
			return errs("!x : x non-integer")
		}
		// TODO: check whether range is too big?
		return rangeI(I(x))
	case I:
		return rangeI(x)
	case Array:
		return rangeArray(x)
	default:
		return errType(x)
	}
}

func rangeI(n I) V {
	if n < 0 {
		return errs("!x : x negative")
	}
	r := make(AI, n)
	for i := range r {
		r[i] = i
	}
	return r
}

func rangeArray(x Array) V {
	z, ok := x.(AI)
	if !ok {
		z = make(AI, x.Len())
		for i := range z {
			v := x.At(i)
			switch v := v.(type) {
			case I:
				z[i] = int(v)
			case F:
				if !isI(v) {
					return errs("!x : x contains non-integer")
				}
				z[i] = int(v)
			default:
				return errs("!x : x is not numeric")
			}
		}
	}
	cols := 1
	for _, n := range z {
		if n == 0 {
			return AV{}
		}
		cols *= n
	}
	r := make(AV, x.Len())
	reps := cols
	for i := range r {
		a := make(AI, cols)
		reps /= z[i]
		clen := reps * z[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < z[i]; j++ {
				for k := 0; k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = a
	}
	return r
}

// where returns &x.
func where(x V) V {
	switch x := x.(type) {
	case I:
		switch {
		case x < 0:
			return errs("&x : x negative")
		case x == 0:
			return AI{}
		default:
			r := make(AI, x)
			return r
		}
	case F:
		if !isI(x) {
			return errs("&x : x non-integer")
		}
		n := I(x)
		switch {
		case n < 0:
			return errs("&x : x negative")
		case n == 0:
			return AI{}
		default:
			r := make(AI, n)
			return r
		}
	case AB:
		n := 0
		for _, v := range x {
			n += int(B2I(v))
		}
		r := make(AI, 0, n)
		for i, v := range x {
			if v {
				r = append(r, i)
			}
		}
		return r
	case AI:
		n := 0
		for _, v := range x {
			if v < 0 {
				return errs("&x : x contains negative integer")
			}
			n += v
		}
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, i)
			}
		}
		return r
	case AF:
		n := 0
		for _, v := range x {
			if !isI(F(v)) {
				return errs("&x : x contains non-integer")
			}
			if v < 0 {
				return errs("&x : x contains negative")
			}
			n += int(v)
		}
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, i)
			}
		}
		return r
	case AV:
		switch aType(x) {
		case tB, tF, tI:
			n := 0
			for _, v := range x {
				switch v := v.(type) {
				case F:
					if !isI(v) {
						return errs("not an integer")
					}
					if v < 0 {
						return errs("negative integer")
					}
					n += int(v)
				case I:
					if v < 0 {
						return errs("negative integer")
					}
					n += int(v)
				}
			}
			r := make(AI, 0, n)
			for i, v := range x {
				var max I
				switch v := v.(type) {
				case I:
					max = v
				case F:
					max = I(v)
				}
				for j := 0; j < int(max); j++ {
					r = append(r, i)
				}
			}
			return r
		default:
			return errs("non-integer")
		}
	default:
		return errs("non-integer")
	}
}

// replicate returns {x}#y.
func replicate(x, y V) V {
	if length(x) != length(y) {
		return errf("f#y : length mismatch: %d (f[y]) vs %d (y)", length(x), length(y))
	}
	switch x := x.(type) {
	case I:
		switch {
		case x < 0:
			return errs("f#y : f[y] negative integer")
		default:
			return repeat(y, int(x))
		}
	case F:
		if !isI(x) {
			return errs("f#y : f[y] not an integer")
		}
		n := int(x)
		switch {
		case n < 0:
			return errs("f#y : f[y] negative")
		default:
			return repeat(y, n)
		}
	case AB:
		return repeatAB(x, y)
	case AI:
		return repeatAI(x, y)
	case AF:
		return repeatAF(x, y)
	case AV:
		return repeatAO(x, y)
	default:
		return errs("f#y : f[y] non-integer")
	}
}

func repeat(x V, n int) V {
	switch x := x.(type) {
	case F:
		r := make(AF, n)
		for i := range r {
			r[i] = float64(x)
		}
		return r
	case I:
		if isBI(x) {
			r := make(AB, n)
			for i := range r {
				r[i] = x == 1
			}
			return r
		}
		r := make(AI, n)
		for i := range r {
			r[i] = int(x)
		}
		return r
	case S:
		r := make(AS, n)
		for i := range r {
			r[i] = string(x)
		}
		return r
	default:
		r := make(AV, n)
		for i := range r {
			r[i] = x
		}
		return r
	}
}

func repeatAB(x AB, y V) V {
	n := 0
	for _, v := range x {
		n += int(B2I(v))
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			if v {
				r = append(r, y.At(i))
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAI(x AI, y V) V {
	n := 0
	for _, v := range x {
		if v < 0 {
			return errs("negative integer")
		}
		n += v
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			for j := 0; j < v; j++ {
				r = append(r, y[i])
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAF(x AF, y V) V {
	n := 0
	for _, v := range x {
		if !isI(F(v)) {
			return errs("not an integer")
		}
		if v < 0 {
			return errs("negative integer")
		}
		n += int(v)
	}
	switch y := y.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, y[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range x {
			for j := 0; j < int(v); j++ {
				r = append(r, y[i])
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAO(x AV, y V) V {
	switch aType(x) {
	case tB, tF, tI:
		n := 0
		for _, v := range x {
			switch v := v.(type) {
			case F:
				if !isI(v) {
					return errsw("non-integer")
				}
				if v < 0 {
					return errsw("negative integer")
				}
				n += int(v)
			case I:
				if v < 0 {
					return errsw("negative integer")
				}
				n += int(v)
			}
		}
		switch y := y.(type) {
		case AB:
			r := make(AB, 0, n)
			for i, v := range x {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, y[i])
				}
			}
			return r
		case AI:
			r := make(AI, 0, n)
			for i, v := range x {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, y[i])
				}
			}
			return r
		case AF:
			r := make(AF, 0, n)
			for i, v := range x {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, y[i])
				}
			}
			return r
		case AS:
			r := make(AS, 0, n)
			for i, v := range x {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, y[i])
				}
			}
			return r
		case AV:
			r := make(AV, 0, n)
			for i, v := range x {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, y[i])
				}
			}
			return r
		default:
			return errs("not an array")
		}
	default:
		return errsw("non-integer")
	}
}
