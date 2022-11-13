package goal

import (
	"sort"
	"strings"
)

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

func less(x, y V) bool {
	switch x := x.(type) {
	case F:
		return lessF(x, y)
	case I:
		return lessI(x, y)
	case S:
		return lessS(x, y)
	case AB:
		if len(x) == 0 {
			return length(y) > 0
		}
		return lessAB(x, y)
	case AF:
		if len(x) == 0 {
			return length(y) > 0
		}
		return lessAF(x, y)
	case AI:
		if len(x) == 0 {
			return length(y) > 0
		}
		return lessAI(x, y)
	case AS:
		if len(x) == 0 {
			return length(y) > 0
		}
		return lessAS(x, y)
	case AV:
		if len(x) == 0 {
			return length(y) > 0
		}
		return lessAO(x, y)
	default:
		return false
	}
}

func lessF(x F, y V) bool {
	switch y := y.(type) {
	case F:
		return x < y
	case I:
		return x < F(y)
	case AB:
		if len(y) == 0 {
			return false
		}
		return x < B2F(y[0]) || x == B2F(y[0]) && len(y) > 1
	case AF:
		if len(y) == 0 {
			return false
		}
		return x < F(y[0]) || x == F(y[0]) && len(y) > 1
	case AI:
		if len(y) == 0 {
			return false
		}
		return x < F(y[0]) || x == F(y[0]) && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessF(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessI(x I, y V) bool {
	switch y := y.(type) {
	case F:
		return F(x) < y
	case I:
		return x < y
	case AB:
		if len(y) == 0 {
			return false
		}
		return x < B2I(y[0]) || x == B2I(y[0]) && len(y) > 1
	case AF:
		if len(y) == 0 {
			return false
		}
		return float64(x) < y[0] || float64(x) == y[0] && len(y) > 1
	case AI:
		if len(y) == 0 {
			return false
		}
		return x < I(y[0]) || x == I(y[0]) && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessI(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessS(x S, y V) bool {
	switch y := y.(type) {
	case S:
		return x < y
	case AS:
		if len(y) == 0 {
			return false
		}
		return string(x) < y[0] || string(x) == y[0] && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessS(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessAB(x AB, y V) bool {
	switch y := y.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] && !y[i] {
				return false
			}
		}
		return len(x) < len(y)
	case AF:
		for i := 0; i < len(x) && i < len(y); i++ {
			if B2F(x[i]) > F(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AI:
		for i := 0; i < len(x) && i < len(y); i++ {
			if B2I(x[i]) > I(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AV:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(y[i], B2I(x[i])) {
				return false
			}
		}
		return len(x) < len(y)
	default:
		return false
	}
}

func lessAI(x AI, y V) bool {
	switch y := y.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(x) && i < len(y); i++ {
			if I(x[i]) > B2I(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AF:
		for i := 0; i < len(x) && i < len(y); i++ {
			if F(x[i]) > F(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AI:
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] > y[i] {
				return false
			}
		}
		return len(x) < len(y)
	case AV:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(y[i], I(x[i])) {
				return false
			}
		}
		return len(x) < len(y)
	default:
		return false
	}
}

func lessAF(x AF, y V) bool {
	switch y := y.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(x) && i < len(y); i++ {
			if F(x[i]) > B2F(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AF:
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] > y[i] {
				return false
			}
		}
		return len(x) < len(y)
	case AI:
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] > float64(y[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AV:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(y[i], F(x[i])) {
				return false
			}
		}
		return len(x) < len(y)
	default:
		return false
	}
}

func lessAS(x AS, y V) bool {
	switch y := y.(type) {
	case S:
		return !lessS(y, x)
	case AS:
		for i := 0; i < len(x) && i < len(y); i++ {
			if x[i] > y[i] {
				return false
			}
		}
		return len(x) < len(y)
	case AV:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(y[i], S(x[i])) {
				return false
			}
		}
		return len(x) < len(y)
	default:
		return false
	}
}

func lessAO(x AV, y V) bool {
	switch y := y.(type) {
	case F:
		return less(x[0], y)
	case I:
		return less(x[0], y)
	case AB:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(B2I(y[i]), x[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AF:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(F(y[i]), x[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AI:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(I(y[i]), x[i]) {
				return false
			}
		}
		return len(x) < len(y)
	case AV:
		for i := 0; i < len(x) && i < len(y); i++ {
			if less(y[i], x[i]) {
				return false
			}
		}
		return len(x) < len(y)
	default:
		return false
	}
}

// sortUp returns ^x.
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
		return errf("^x : x not an array (%s)", x.Type())
	}
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
		return errf("<x : x not an array (%s)", x.Type())
	}
}

// descend returns >x.
func descend(x V) V {
	p := ascend(x)
	switch p := p.(type) {
	case E:
		return errs(">" + strings.TrimPrefix(p.Error(), "<"))
	}
	reverseMut(p)
	return p
}
