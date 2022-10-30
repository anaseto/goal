package main

// Range returns ↕x.
func Range(x V) V {
	switch x := x.(type) {
	case B:
		return rangeI(B2I(x))
	case F:
		if !isI(x) {
			return errs("non-integer range")
		}
		// TODO: check whether range is too big?
		return rangeI(I(x))
	case I:
		return rangeI(x)
	case Array:
		return rangeArray(x)
	default:
		return errs("bad type")
	}
}

func rangeI(n I) V {
	if n < 0 {
		return errs("negative integer")
	}
	r := make(AI, n)
	for i := range r {
		r[i] = i
	}
	return r
}

func rangeArray(x Array) V {
	y := make(AI, x.Len())
	for i := range y {
		v := x.At(i)
		switch v := v.(type) {
		case B:
			y[i] = int(B2I(v))
		case I:
			y[i] = int(v)
		case F:
			if !isI(v) {
				return errs("non-integer range")
			}
			y[i] = int(v)
		default:
			return errs("non-numeric")
		}
	}
	cols := 1
	for _, n := range y {
		if n == 0 {
			return AV{}
		}
		cols *= n
	}
	r := make(AV, x.Len())
	reps := cols
	for i := range r {
		a := make(AI, cols)
		reps /= y[i]
		clen := reps * y[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < y[i]; j++ {
				for k := 0; k < reps; k++ {
					a[col+j*reps+k] = j
				}
			}
		}
		r[i] = a
	}
	return r
}

// Indices returns &x.
func Indices(x V) V {
	switch x := x.(type) {
	case B:
		if x {
			return AI{0}
		}
		return AI{}
	case I:
		switch {
		case x < 0:
			return errs("negative integer")
		case x == 0:
			return AI{}
		default:
			r := make(AI, x)
			return r
		}
	case F:
		if !isI(x) {
			return errs("not an integer")
		}
		n := I(x)
		switch {
		case n < 0:
			return errs("negative integer")
		case n == 0:
			return AI{}
		default:
			r := make(AI, n)
			return r
		}
	case AB:
		n := 0
		for _, v := range x {
			n += int(B2I(B(v)))
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
				return errs("negative integer")
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
				return errs("not an integer")
			}
			if v < 0 {
				return errs("negative integer")
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
				case B:
					n += int(B2I(v))
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
				case B:
					max = B2I(v)
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

// Replicate returns {w}#x.
func Replicate(w, x V) V {
	if Length(w) != Length(x) {
		return errf("length mismatch: %d vs %d", Length(w), Length(x))
	}
	switch w := w.(type) {
	case B:
		return repeat(x, int(B2I(w)))
	case I:
		switch {
		case w < 0:
			return errs("negative integer")
		default:
			return repeat(x, int(w))
		}
	case F:
		if !isI(w) {
			return errs("not an integer")
		}
		n := int(w)
		switch {
		case n < 0:
			return errs("negative integer")
		default:
			return repeat(x, n)
		}
	case AB:
		return repeatAB(w, x)
	case AI:
		return repeatAI(w, x)
	case AF:
		return repeatAF(w, x)
	case AV:
		return repeatAO(w, x)
	default:
		return errsw("non-integer")
	}
}

func repeat(x V, n int) V {
	switch x := x.(type) {
	case B:
		r := make(AB, n)
		for i := range r {
			r[i] = bool(x)
		}
		return r
	case F:
		r := make(AF, n)
		for i := range r {
			r[i] = float64(x)
		}
		return r
	case I:
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

func repeatAB(w AB, x V) V {
	n := 0
	for _, v := range w {
		n += int(B2I(B(v)))
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range w {
			if v {
				r = append(r, x[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range w {
			if v {
				r = append(r, x[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range w {
			if v {
				r = append(r, x[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range w {
			if v {
				r = append(r, x[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range w {
			if v {
				r = append(r, x.At(i))
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAI(w AI, x V) V {
	n := 0
	for _, v := range w {
		if v < 0 {
			return errs("negative integer")
		}
		n += v
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range w {
			for j := 0; j < v; j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range w {
			for j := 0; j < v; j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range w {
			for j := 0; j < v; j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range w {
			for j := 0; j < v; j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range w {
			for j := 0; j < v; j++ {
				r = append(r, x[i])
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAF(w AF, x V) V {
	n := 0
	for _, v := range w {
		if !isI(F(v)) {
			return errs("not an integer")
		}
		if v < 0 {
			return errs("negative integer")
		}
		n += int(v)
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, 0, n)
		for i, v := range w {
			for j := 0; j < int(v); j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AF:
		r := make(AF, 0, n)
		for i, v := range w {
			for j := 0; j < int(v); j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AI:
		r := make(AI, 0, n)
		for i, v := range w {
			for j := 0; j < int(v); j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AS:
		r := make(AS, 0, n)
		for i, v := range w {
			for j := 0; j < int(v); j++ {
				r = append(r, x[i])
			}
		}
		return r
	case AV:
		r := make(AV, 0, n)
		for i, v := range w {
			for j := 0; j < int(v); j++ {
				r = append(r, x[i])
			}
		}
		return r
	default:
		return errs("not an array")
	}
}

func repeatAO(w AV, x V) V {
	switch aType(w) {
	case tB, tF, tI:
		n := 0
		for _, v := range w {
			switch v := v.(type) {
			case B:
				n += int(B2I(v))
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
		switch x := x.(type) {
		case AB:
			r := make(AB, 0, n)
			for i, v := range w {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, x[i])
				}
			}
			return r
		case AI:
			r := make(AI, 0, n)
			for i, v := range w {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, x[i])
				}
			}
			return r
		case AF:
			r := make(AF, 0, n)
			for i, v := range w {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, x[i])
				}
			}
			return r
		case AS:
			r := make(AS, 0, n)
			for i, v := range w {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, x[i])
				}
			}
			return r
		case AV:
			r := make(AV, 0, n)
			for i, v := range w {
				max := num2I(v)
				for j := 0; j < int(max); j++ {
					r = append(r, x[i])
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
