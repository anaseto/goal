package main

import (
	"math"
	"sort"
	"strings"
)

func Length(x O) I {
	switch x := x.(type) {
	case AB:
		return len(x)
	case AF:
		return len(x)
	case AI:
		return len(x)
	case AS:
		return len(x)
	case AO:
		return len(x)
	default:
		return 1
	}
}

func Negate(x O) O {
	switch x := x.(type) {
	case B:
		return -B2I(x)
	case F:
		return -x
	case I:
		return -x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Negate(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("=")
	}
}

func signF(x F) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

func signI(x I) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

func Sign(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return signF(x)
	case I:
		return signI(x)
	case AB:
		return x
	case AF:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signF(x[i])
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signI(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Sign(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("×")
	}
}

func Reciprocal(x O) O {
	switch x := x.(type) {
	case B:
		return divide(1, B2F(x))
	case F:
		return divide(1, x)
	case I:
		return divide(1, F(x))
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, F(x[i]))
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Reciprocal(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("÷")
	}
}

func Floor(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Floor(x)
	case I:
		return x
	case S:
		return strings.ToLower(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Floor(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToLower(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Floor(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌊")
	}
}

func Ceil(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Ceil(x)
	case I:
		return x
	case S:
		return strings.ToUpper(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Ceil(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToUpper(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Ceil(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌈")
	}
}

func Not(x O) O {
	switch x := x.(type) {
	case B:
		return !x
	case F:
		return 1 - x
	case I:
		return 1 - x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = !x[i]
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Not(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}

func absI(x I) I {
	if x < 0 {
		return -x
	}
	return x
}

func Abs(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Abs(x)
	case I:
		return absI(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Abs(x[i])
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = absI(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Abs(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}

func clone(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = clone(x[i])
		}
		return r
	case E:
		return x
	default:
		return x
	}
}

func cloneShallow(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		copy(r, x)
		return r
	case E:
		return x
	default:
		return x
	}
}

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

func SortUp(x O) O {
	// XXX: error if length is zero?
	switch x := x.(type) {
	case AB:
		x = cloneShallow(x)
		sort.Stable(ABUp(x))
		return x
	case AF:
		x = cloneShallow(x)
		sort.Stable(sort.Float64Slice(x))
		return x
	case AI:
		x = cloneShallow(x)
		sort.Stable(sort.IntSlice(x))
		return x
	case AS:
		x = cloneShallow(x)
		sort.Stable(sort.StringSlice(x))
		return x
	case AO:
		x = cloneShallow(x)
		sort.Stable(AOUp(x))
		return x
	case E:
		return x
	default:
		return badtype("<")
	}
}

func SortDown(x O) O {
	x = SortUp(x)
	switch x := x.(type) {
	case E:
		return badtype(">")
	}
	reverse(x)
	return x
}

func reverse(x O) {
	switch x := x.(type) {
	case AB:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AF:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AI:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AS:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AO:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	}
}

// Reverse applies to a list and returns a new list in reverse order.
func Reverse(x O) O {
	if !isArray(x) {
		return badtype("⌽")
	}
	x = cloneShallow(x)
	reverse(x)
	return x
}
