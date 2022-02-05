package main

import "math"

// Range returns ↕x.
func Range(x O) O {
	switch x := x.(type) {
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
			if math.Floor(v) != v {
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
		r[i] = make(AI, cols)
		reps /= y[i]
		clen := reps * y[i]
		for c := 0; c < cols/clen; c++ {
			col := c * clen
			for j := 0; j < y[i]; j++ {
				for k := 0; k < reps; k++ {
					r[i][col+j+k] = y[i]
				}
			}
		}
	}
	return r
}
