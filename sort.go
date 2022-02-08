package main

import "sort"

type ABUp []B

func (bs ABUp) Len() int {
	return len(bs)
}

func (bs ABUp) Less(i, j int) bool {
	return bs[j] && !bs[i]
}

func (bs ABUp) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

type AOUp []O

func (bs AOUp) Len() int {
	return len(bs)
}

func (bs AOUp) Less(i, j int) bool {
	return less(bs[i], bs[j])
}

func (bs AOUp) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func less(w, x O) bool {
	switch w := w.(type) {
	case B:
		return lessB(w, x)
	case F:
		return lessF(w, x)
	case I:
		return lessI(w, x)
	case S:
		return lessS(w, x)
	case AB:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAB(w, x)
	case AF:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAF(w, x)
	case AI:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAI(w, x)
	case AS:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAS(w, x)
	case AO:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAO(w, x)
	default:
		return false
	}
}

func lessB(w B, x O) bool {
	switch x := x.(type) {
	case B:
		return !w && x
	case F:
		return B2F(w) < x
	case I:
		return B2I(w) < x
	case AB:
		if len(x) == 0 {
			return false
		}
		return !w && x[0] || w == x[0] && len(x) > 1
	case AF:
		if len(x) == 0 {
			return false
		}
		return B2F(w) < x[0] || B2F(w) == x[0] && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return B2I(w) < x[0] || B2I(w) == x[0] && len(x) > 1
	case AO:
		if len(x) == 0 {
			return false
		}
		return lessB(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessF(w F, x O) bool {
	switch x := x.(type) {
	case B:
		return w < B2F(x)
	case F:
		return w < x
	case I:
		return w < F(x)
	case AB:
		if len(x) == 0 {
			return false
		}
		return w < B2F(x[0]) || w == B2F(x[0]) && len(x) > 1
	case AF:
		if len(x) == 0 {
			return false
		}
		return w < x[0] || w == x[0] && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return w < F(x[0]) || w == F(x[0]) && len(x) > 1
	case AO:
		if len(x) == 0 {
			return false
		}
		return lessF(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessI(w I, x O) bool {
	switch x := x.(type) {
	case B:
		return w < B2I(x)
	case F:
		return F(w) < x
	case I:
		return w < x
	case AB:
		if len(x) == 0 {
			return false
		}
		return w < B2I(x[0]) || w == B2I(x[0]) && len(x) > 1
	case AF:
		if len(x) == 0 {
			return false
		}
		return F(w) < x[0] || F(w) == x[0] && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return w < x[0] || w == x[0] && len(x) > 1
	case AO:
		if len(x) == 0 {
			return false
		}
		return lessI(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessS(w S, x O) bool {
	switch x := x.(type) {
	case S:
		return w < x
	case AS:
		if len(x) == 0 {
			return false
		}
		return w < x[0] || w == x[0] && len(x) > 1
	case AO:
		if len(x) == 0 {
			return false
		}
		return lessS(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessAB(w AB, x O) bool {
	switch x := x.(type) {
	case B:
		return !lessB(x, w)
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] && !x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if B2F(w[i]) > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if B2I(w[i]) > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AO:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	default:
		return false
	}
}

func lessAI(w AI, x O) bool {
	switch x := x.(type) {
	case B:
		return !lessB(x, w)
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > B2I(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if F(w[i]) > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AO:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	default:
		return false
	}
}

func lessAF(w AF, x O) bool {
	switch x := x.(type) {
	case B:
		return !lessB(x, w)
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > B2F(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > F(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AO:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	default:
		return false
	}
}

func lessAS(w AS, x O) bool {
	switch x := x.(type) {
	case S:
		return !lessS(x, w)
	case AS:
		for i := 0; i < len(w) && i < len(x); i++ {
			if w[i] > x[i] {
				return false
			}
		}
		return len(w) < len(x)
	case AO:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	default:
		return false
	}
}

func lessAO(w AO, x O) bool {
	switch x := x.(type) {
	case B:
		return less(w[0], x)
	case F:
		return less(w[0], x)
	case I:
		return less(w[0], x)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AO:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	default:
		return false
	}
}

// SortUp returns <x.
func SortUp(x O) O {
	// XXX: error if length is zero?
	x = canonical(x)
	x = cloneShallow(x)
	switch x := x.(type) {
	case AB:
		sort.Stable(ABUp(x))
		return x
	case AF:
		sort.Stable(sort.Float64Slice(x))
		return x
	case AI:
		sort.Stable(sort.IntSlice(x))
		return x
	case AS:
		sort.Stable(sort.StringSlice(x))
		return x
	case AO:
		sort.Stable(AOUp(x))
		return x
	case E:
		return x
	default:
		return badtype("<")
	}
}

// SortDown returns >x.
func SortDown(x O) O {
	x = SortUp(x)
	switch x.(type) {
	case E:
		// TODO: match only < error type
		return badtype(">")
	}
	reverse(x)
	return x
}
