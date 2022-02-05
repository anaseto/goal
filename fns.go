package main

// Range returns ↕x.
func Range(x O) O {
	switch x := x.(type) {
	case B:
		return rangeI(B2I(x))
	case F:
		if !isI(x) {
			return badtype("↕ : non-integer range")
		}
		// TODO: check whether range is too big?
		return rangeI(I(x))
	case I:
		return rangeI(x)
	case Array:
		return rangeArray(x)
	default:
		return badtype("↕")
	}
}

func rangeI(n I) O {
	if n < 0 {
		return badtype("↕ : negative integer")
	}
	r := make(AI, n)
	for i := 0; i < n; i++ {
		r[i] = i
	}
	return r
}

func rangeArray(x Array) O {
	y := make(AI, x.Len())
	for i := range y {
		v := x.At(i)
		switch v := v.(type) {
		case B:
			y[i] = B2I(v)
		case I:
			y[i] = v
		case F:
			if !isI(v) {
				return badtype("↕ : non-integer range")
			}
			y[i] = I(v)
		default:
			return badtype("↕ : non-numeric argument")
		}
	}
	cols := 1
	for _, n := range y {
		if n == 0 {
			return AO{}
		}
		cols *= n
	}
	r := make(AO, x.Len())
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

// Indices returns /x.
func Indices(x O) O {
	switch x := x.(type) {
	case B:
		if x {
			return AI{0}
		}
		return AI{}
	case I:
		switch {
		case x < 0:
			return badtype("/ : negative integer")
		case x == 0:
			return AI{}
		default:
			r := make(AI, x)
			return r
		}
	case F:
		if !isI(x) {
			return badtype("/ : not an integer")
		}
		n := I(x)
		switch {
		case n < 0:
			return badtype("/ : negative integer")
		case n == 0:
			return AI{}
		default:
			r := make(AI, n)
			return r
		}
	case AB:
		n := 0
		for _, v := range x {
			n += B2I(v)
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
				return badtype("/ : negative integer")
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
			if !isI(v) {
				return badtype("/ : not an integer")
			}
			if v < 0 {
				return badtype("/ : negative integer")
			}
			n += I(v)
		}
		r := make(AI, 0, n)
		for i, v := range x {
			for j := 0; j < I(v); j++ {
				r = append(r, i)
			}
		}
		return r
	case AO:
		switch aType(x) {
		case tB, tF, tI:
			n := 0
			for _, v := range x {
				switch v := v.(type) {
				case B:
					n += B2I(v)
				case F:
					if !isI(v) {
						return badtype("/ : not an integer")
					}
					if v < 0 {
						return badtype("/ : negative integer")
					}
					n += I(v)
				case I:
					if v < 0 {
						return badtype("/ : negative integer")
					}
					n += v
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
				for j := 0; j < max; j++ {
					r = append(r, i)
				}
			}
			return r
		default:
			return badtype("/ : expected integer(s)")
		}
	default:
		return badtype("/ : expected integer(s)")
	}
}
