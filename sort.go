package goal

import "sort"

type sortAB []bool

func (bs sortAB) Len() int {
	return len(bs)
}

func (bs sortAB) Less(i, j int) bool {
	return bs[j] && !bs[i]
}

func (bs sortAB) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

type sortAO []V

func (bs sortAO) Len() int {
	return len(bs)
}

func (bs sortAO) Less(i, j int) bool {
	return less(bs[i], bs[j])
}

func (bs sortAO) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func less(w, x V) bool {
	switch w := w.(type) {
	case F:
		return lessF(w, x)
	case I:
		return lessI(w, x)
	case S:
		return lessS(w, x)
	case AB:
		if len(w) == 0 {
			return length(x) > 0
		}
		return lessAB(w, x)
	case AF:
		if len(w) == 0 {
			return length(x) > 0
		}
		return lessAF(w, x)
	case AI:
		if len(w) == 0 {
			return length(x) > 0
		}
		return lessAI(w, x)
	case AS:
		if len(w) == 0 {
			return length(x) > 0
		}
		return lessAS(w, x)
	case AV:
		if len(w) == 0 {
			return length(x) > 0
		}
		return lessAO(w, x)
	default:
		return false
	}
}

func lessF(w F, x V) bool {
	switch x := x.(type) {
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
			if B2F(w[i]) > F(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if B2I(w[i]) > I(x[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AV:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(x[i], B2I(w[i])) {
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
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if I(w[i]) > B2I(x[i]) {
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
			if less(x[i], I(w[i])) {
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
	case F:
		return !lessF(x, w)
	case I:
		return !lessI(x, w)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if F(w[i]) > B2F(x[i]) {
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
			if less(x[i], F(w[i])) {
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
			if less(x[i], S(w[i])) {
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
	case F:
		return less(w[0], x)
	case I:
		return less(w[0], x)
	case AB:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(B2I(x[i]), w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AF:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(F(x[i]), w[i]) {
				return false
			}
		}
		return len(w) < len(x)
	case AI:
		for i := 0; i < len(w) && i < len(x); i++ {
			if less(I(x[i]), w[i]) {
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

// sortUp returns <x.
func sortUp(x V) V {
	// XXX: error if atom?
	x = canonical(x)
	x = cloneShallow(x)
	switch x := x.(type) {
	case AB:
		sort.Stable(sortAB(x))
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
		sort.Stable(sortAO(x))
		return x
	default:
		return errs("not an array")
	}
}

// SortDown returns >x.
func SortDown(x V) V {
	x = sortUp(x)
	switch x.(type) {
	case E:
		return x
	}
	reverseMut(x)
	return x
}

type permutationAO struct {
	Perm AI
	X    sortAO
}

func (p *permutationAO) Len() int {
	return p.X.Len()
}

func (p *permutationAO) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAO) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type permutationAB struct {
	Perm AI
	X    sortAB
}

func (p *permutationAB) Len() int {
	return p.X.Len()
}

func (p *permutationAB) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAB) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type permutationAI struct {
	Perm AI
	X    sort.IntSlice
}

func (p *permutationAI) Len() int {
	return p.X.Len()
}

func (p *permutationAI) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAI) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type permutationAF struct {
	Perm AI
	X    sort.Float64Slice
}

func (p *permutationAF) Len() int {
	return p.X.Len()
}

func (p *permutationAF) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAF) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

type permutationAS struct {
	Perm AI
	X    sort.StringSlice
}

func (p *permutationAS) Len() int {
	return p.X.Len()
}

func (p *permutationAS) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAS) Less(i, j int) bool {
	return p.X.Less(p.Perm[i], p.Perm[j])
}

func permRange(n int) AI {
	r := make(AI, n)
	for i := range r {
		r[i] = i
	}
	return r
}

// ascend returns <x.
func ascend(x V) V {
	switch x := x.(type) {
	case AB:
		p := &permutationAB{Perm: permRange(len(x)), X: sortAB(x)}
		sort.Stable(p)
		return p.Perm
	case AF:
		p := &permutationAF{Perm: permRange(len(x)), X: sort.Float64Slice(x)}
		sort.Stable(p)
		return p.Perm
	case AI:
		p := &permutationAI{Perm: permRange(len(x)), X: sort.IntSlice(x)}
		sort.Stable(p)
		return p.Perm
	case AS:
		p := &permutationAS{Perm: permRange(len(x)), X: sort.StringSlice(x)}
		sort.Stable(p)
		return p.Perm
	case AV:
		p := &permutationAO{Perm: permRange(len(x)), X: sortAO(x)}
		sort.Stable(p)
		return p.Perm
	default:
		return errs("not an array")
	}
}

// descend returns >x.
func descend(x V) V {
	p := ascend(x)
	switch p.(type) {
	case E:
		return p
	}
	reverseMut(p)
	return p
}
