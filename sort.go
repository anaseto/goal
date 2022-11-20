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

type sortAV []V

func (bs sortAV) Len() int {
	return len(bs)
}

func (bs sortAV) Less(i, j int) bool {
	return less(bs[i], bs[j])
}

func (bs sortAV) Swap(i, j int) {
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
		return lessAV(x, y)
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

func lessAV(x AV, y V) bool {
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
		sort.Stable(sortAV(x))
		return x
	default:
		return errf("^x : x not an array (%s)", x.Type())
	}
}

type permutationAV struct {
	Perm AI
	X    sortAV
}

func (p *permutationAV) Len() int {
	return p.X.Len()
}

func (p *permutationAV) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAV) Less(i, j int) bool {
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
		p := &permutationAV{Perm: permRange(len(x)), X: sortAV(x)}
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

// search implements x$y.
func search(x V, y V) V {
	switch x := x.(type) {
	case AB:
		if !sort.IsSorted(sortAB(x)) {
			return errs("x$y : x is not ascending")
		}
		return searchAI(fromABtoAI(x).(AI), y)
	case AI:
		if !sort.IsSorted(sort.IntSlice(x)) {
			return errs("x$y : x is not ascending")
		}
		return searchAI(x, y)
	case AF:
		if !sort.IsSorted(sort.Float64Slice(x)) {
			return errs("x$y : x is not ascending")
		}
		return searchAF(x, y)
	case AS:
		if !sort.IsSorted(sort.StringSlice(x)) {
			return errs("x$y : x is not ascending")
		}
		return searchAS(x, y)
	case AV:
		if !sort.IsSorted(sortAV(x)) {
			return errs("x$y : x is not ascending")
		}
		return searchAV(x, y)
	default:
		// should not happen
		return errf("x$y : bad type for x (%s)", x.Type())
	}
}

func searchAII(x AI, y I) int {
	return sort.Search(len(x), func(i int) bool { return I(x[i]) > y })
}

func searchAIF(x AI, y F) int {
	return sort.Search(len(x), func(i int) bool { return F(x[i]) > y })
}

func searchAFI(x AF, y I) int {
	return sort.Search(len(x), func(i int) bool { return x[i] > float64(y) })
}

func searchAFF(x AF, y F) int {
	return sort.Search(len(x), func(i int) bool { return F(x[i]) > y })
}

func searchASS(x AS, y S) int {
	return sort.Search(len(x), func(i int) bool { return S(x[i]) > y })
}

func searchAI(x AI, y V) V {
	switch y := y.(type) {
	case I:
		return I(searchAII(x, y))
	case F:
		return I(searchAIF(x, y))
	case AB:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAII(x, B2I(v))
		}
		return res
	case AI:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAII(x, I(v))
		}
		return res
	case AF:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAIF(x, F(v))
		}
		return res
	case Array:
		res := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			res[i] = sort.Search(len(x),
				func(i int) bool { return less(y.At(i), I(x[i])) })
		}
		return res
	default:
		return I(x.Len())
	}
}

func searchAF(x AF, y V) V {
	switch y := y.(type) {
	case I:
		return I(searchAFI(x, y))
	case F:
		return I(searchAFF(x, y))
	case AB:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAFI(x, B2I(v))
		}
		return res
	case AI:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAFI(x, I(v))
		}
		return res
	case AF:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchAFF(x, F(v))
		}
		return res
	case Array:
		res := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			res[i] = sort.Search(len(x),
				func(i int) bool { return less(y.At(i), F(x[i])) })
		}
		return res
	default:
		return I(x.Len())
	}
}

func searchAS(x AS, y V) V {
	switch y := y.(type) {
	case S:
		return I(searchASS(x, y))
	case AS:
		res := make(AI, y.Len())
		for i, v := range y {
			res[i] = searchASS(x, S(v))
		}
		return res
	case Array:
		res := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			res[i] = sort.Search(len(x),
				func(i int) bool { return less(y.At(i), S(x[i])) })
		}
		return res
	default:
		return I(x.Len())
	}
}

func searchAV(x AV, y V) V {
	switch y := y.(type) {
	case Array:
		res := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			res[i] = sort.Search(len(x),
				func(i int) bool { return less(y.At(i), x[i]) })
		}
		return res
	default:
		return I(sort.Search(len(x),
			func(i int) bool { return less(y, x[i]) }))
	}
}
