package main

import "sort"

type ABUp []bool

func (bs ABUp) Len() int {
	return len(bs)
}

func (bs ABUp) Less(i, j int) bool {
	return bs[j] && !bs[i]
}

func (bs ABUp) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

type AOUp []V

func (bs AOUp) Len() int {
	return len(bs)
}

func (bs AOUp) Less(i, j int) bool {
	return less(bs[i], bs[j])
}

func (bs AOUp) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func less(w, x V) bool {
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
	case AV:
		if len(w) == 0 {
			return Length(x) > 0
		}
		return lessAO(w, x)
	default:
		return false
	}
}

func lessB(w B, x V) bool {
	switch x := x.(type) {
	case B:
		return bool(!w && x)
	case F:
		return B2F(w) < x
	case I:
		return B2I(w) < x
	case AB:
		if len(x) == 0 {
			return false
		}
		return bool(!w && B(x[0]) || w == B(x[0]) && len(x) > 1)
	case AF:
		if len(x) == 0 {
			return false
		}
		return B2F(w) < F(x[0]) || B2F(w) == F(x[0]) && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return B2I(w) < I(x[0]) || B2I(w) == I(x[0]) && len(x) > 1
	case AV:
		if len(x) == 0 {
			return false
		}
		return lessB(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessF(w F, x V) bool {
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
		return w < B2F(B(x[0])) || w == B2F(B(x[0])) && len(x) > 1
	case AF:
		if len(x) == 0 {
			return false
		}
		return w < F(x[0]) || w == F(x[0]) && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return w < F(x[0]) || w == F(x[0]) && len(x) > 1
	case AV:
		if len(x) == 0 {
			return false
		}
		return lessF(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessI(w I, x V) bool {
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
		return w < B2I(B(x[0])) || w == B2I(B(x[0])) && len(x) > 1
	case AF:
		if len(x) == 0 {
			return false
		}
		return float64(w) < x[0] || float64(w) == x[0] && len(x) > 1
	case AI:
		if len(x) == 0 {
			return false
		}
		return w < I(x[0]) || w == I(x[0]) && len(x) > 1
	case AV:
		if len(x) == 0 {
			return false
		}
		return lessI(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessS(w S, x V) bool {
	switch x := x.(type) {
	case S:
		return w < x
	case AS:
		if len(x) == 0 {
			return false
		}
		return string(w) < x[0] || string(w) == x[0] && len(x) > 1
	case AV:
		if len(x) == 0 {
			return false
		}
		return lessS(w, x[0]) || !less(x[0], w) && len(x) > 1
	default:
		return false
	}
}

func lessAB(w AB, x V) bool {
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
			if B2F(B(w[i])) > F(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if B2I(B(w[i])) > I(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AV:
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

func lessAI(w AI, x V) bool {
	switch x := x.(type) {
	case B:
		return !lessB(x, w)
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if I(w[i]) > B2I(B(x[i])) {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if F(w[i]) > F(x[i]) {
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
	case AV:
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

func lessAF(w AF, x V) bool {
	switch x := x.(type) {
	case B:
		return !lessB(x, w)
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if F(w[i]) > B2F(B(x[i])) {
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
			if w[i] > float64(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AV:
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

func lessAS(w AS, x V) bool {
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
	case AV:
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

func lessAO(w AV, x V) bool {
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
	case AV:
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
func SortUp(x V) V {
	// XXX: error if atom?
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
	case AV:
		sort.Stable(AOUp(x))
		return x
	default:
		return badtype("<")
	}
}

// SortDown returns >x.
func SortDown(x V) V {
	x = SortUp(x)
	switch x.(type) {
	case E:
		return x
	}
	reverse(x)
	return x
}

type PermutationAO struct {
	Perm []int
	X    AOUp
}

func (p *PermutationAO) Len() int {
	return p.X.Len()
}

func (p *PermutationAO) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *PermutationAO) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type PermutationAB struct {
	Perm AI
	X    ABUp
}

func (p *PermutationAB) Len() int {
	return p.X.Len()
}

func (p *PermutationAB) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *PermutationAB) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type PermutationAI struct {
	Perm AI
	X    sort.IntSlice
}

func (p *PermutationAI) Len() int {
	return p.X.Len()
}

func (p *PermutationAI) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *PermutationAI) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type PermutationAF struct {
	Perm AI
	X    sort.Float64Slice
}

func (p *PermutationAF) Len() int {
	return p.X.Len()
}

func (p *PermutationAF) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *PermutationAF) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type PermutationAS struct {
	Perm AI
	X    sort.StringSlice
}

func (p *PermutationAS) Len() int {
	return p.X.Len()
}

func (p *PermutationAS) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *PermutationAS) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

func permRange(n int) AI {
	r := make(AI, n)
	for i := range r {
		r[i] = i
	}
	return r
}

// GradeUp returns <x.
func GradeUp(x V) V {
	switch x := x.(type) {
	case AB:
		p := &PermutationAB{Perm: permRange(len(x)), X: ABUp(x)}
		sort.Stable(p)
		return p.Perm
	case AF:
		p := &PermutationAF{Perm: permRange(len(x)), X: sort.Float64Slice(x)}
		sort.Stable(p)
		return p.Perm
	case AI:
		p := &PermutationAI{Perm: permRange(len(x)), X: sort.IntSlice(x)}
		sort.Stable(p)
		return p.Perm
	case AS:
		p := &PermutationAS{Perm: permRange(len(x)), X: sort.StringSlice(x)}
		sort.Stable(p)
		return p.Perm
	case AV:
		p := &PermutationAO{Perm: permRange(len(x)), X: AOUp(x)}
		sort.Stable(p)
		return p.Perm
	default:
		return badtype("⍋ : x must be an array")
	}
}

// GradeDown returns >x.
func GradeDown(x V) V {
	p := GradeUp(x)
	switch p.(type) {
	case E:
		return p
	}
	reverse(p)
	return p
}
