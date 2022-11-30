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
	switch xv := x.Value.(type) {
	case F:
		return lessF(x, y)
	case I:
		return lessI(x, y)
	case S:
		return lessS(x, y)
	case AB:
		if len(xv) == 0 {
			return Length(y) > 0
		}
		return lessAB(x, y)
	case AF:
		if len(xv) == 0 {
			return Length(y) > 0
		}
		return lessAF(x, y)
	case AI:
		if len(xv) == 0 {
			return Length(y) > 0
		}
		return lessAI(x, y)
	case AS:
		if len(xv) == 0 {
			return Length(y) > 0
		}
		return lessAS(x, y)
	case AV:
		if len(xv) == 0 {
			return Length(y) > 0
		}
		return lessAV(x, y)
	default:
		return false
	}
}

func lessF(x V, y V) bool {
	xv := x.Value.(F)
	switch y := y.Value.(type) {
	case F:
		return xv < y
	case I:
		return xv < F(y)
	case AB:
		if len(y) == 0 {
			return false
		}
		return xv < B2F(y[0]) || xv == B2F(y[0]) && len(y) > 1
	case AF:
		if len(y) == 0 {
			return false
		}
		return xv < F(y[0]) || xv == F(y[0]) && len(y) > 1
	case AI:
		if len(y) == 0 {
			return false
		}
		return xv < F(y[0]) || xv == F(y[0]) && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessF(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessI(x V, y V) bool {
	xv := x.Value.(I)
	switch y := y.Value.(type) {
	case F:
		return F(xv) < y
	case I:
		return xv < y
	case AB:
		if len(y) == 0 {
			return false
		}
		return xv < B2I(y[0]) || xv == B2I(y[0]) && len(y) > 1
	case AF:
		if len(y) == 0 {
			return false
		}
		return float64(xv) < y[0] || float64(xv) == y[0] && len(y) > 1
	case AI:
		if len(y) == 0 {
			return false
		}
		return xv < I(y[0]) || xv == I(y[0]) && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessI(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessS(x V, y V) bool {
	xv := x.Value.(S)
	switch y := y.Value.(type) {
	case S:
		return xv < y
	case AS:
		if len(y) == 0 {
			return false
		}
		return string(xv) < y[0] || string(xv) == y[0] && len(y) > 1
	case AV:
		if len(y) == 0 {
			return false
		}
		return lessS(x, y[0]) || !less(y[0], x) && len(y) > 1
	default:
		return false
	}
}

func lessAB(x V, y V) bool {
	xv := x.Value.(AB)
	switch yv := y.Value.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if xv[i] && !yv[i] {
				return false
			}
		}
		return len(xv) < len(yv)
	case AF:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if B2F(xv[i]) > F(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AI:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if B2I(xv[i]) > I(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AV:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(yv[i], NewV(B2I(xv[i]))) {
				return false
			}
		}
		return len(xv) < len(yv)
	default:
		return false
	}
}

func lessAI(x V, y V) bool {
	xv := x.Value.(AI)
	switch yv := y.Value.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if I(xv[i]) > B2I(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AF:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if F(xv[i]) > F(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AI:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if xv[i] > yv[i] {
				return false
			}
		}
		return len(xv) < len(yv)
	case AV:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(yv[i], NewI(xv[i])) {
				return false
			}
		}
		return len(xv) < len(yv)
	default:
		return false
	}
}

func lessAF(x V, y V) bool {
	xv := x.Value.(AF)
	switch yv := y.Value.(type) {
	case F:
		return !lessF(y, x)
	case I:
		return !lessI(y, x)
	case AB:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if F(xv[i]) > B2F(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AF:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if xv[i] > yv[i] {
				return false
			}
		}
		return len(xv) < len(yv)
	case AI:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if xv[i] > float64(yv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AV:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(yv[i], NewF(xv[i])) {
				return false
			}
		}
		return len(xv) < len(yv)
	default:
		return false
	}
}

func lessAS(x V, y V) bool {
	xv := x.Value.(AS)
	switch yv := y.Value.(type) {
	case S:
		return !lessS(y, x)
	case AS:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if xv[i] > yv[i] {
				return false
			}
		}
		return len(xv) < len(yv)
	case AV:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(yv[i], NewS(xv[i])) {
				return false
			}
		}
		return len(xv) < len(yv)
	default:
		return false
	}
}

func lessAV(x V, y V) bool {
	xv := x.Value.(AV)
	switch yv := y.Value.(type) {
	case F:
		return less(xv[0], y)
	case I:
		return less(xv[0], y)
	case AB:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(NewV(B2I(yv[i])), xv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AF:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(NewF(yv[i]), xv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AI:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(NewI(yv[i]), xv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	case AV:
		for i := 0; i < len(xv) && i < len(yv); i++ {
			if less(yv[i], xv[i]) {
				return false
			}
		}
		return len(xv) < len(yv)
	default:
		return false
	}
}

// sortUp returns ^x.
func sortUp(x V) V {
	// TODO: avoid cases of double clone
	//assertCanonical(x)
	x = cloneShallow(x)
	switch x := x.Value.(type) {
	case AB:
		sort.Stable(sortAB(x))
		return NewV(x)
	case AF:
		sort.Stable(sort.Float64Slice(x))
		return NewV(x)
	case AI:
		sort.Stable(sort.IntSlice(x))
		return NewV(x)
	case AS:
		sort.Stable(sort.StringSlice(x))
		return NewV(x)
	case AV:
		sort.Stable(sortAV(x))
		return NewV(x)
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
	switch x := x.Value.(type) {
	case AB:
		p := &permutationAB{Perm: permRange(len(x)), X: sortAB(x)}
		sort.Stable(p)
		return NewV(p.Perm)
	case AF:
		p := &permutationAF{Perm: permRange(len(x)), X: sort.Float64Slice(x)}
		sort.Stable(p)
		return NewV(p.Perm)
	case AI:
		p := &permutationAI{Perm: permRange(len(x)), X: sort.IntSlice(x)}
		sort.Stable(p)
		return NewV(p.Perm)
	case AS:
		p := &permutationAS{Perm: permRange(len(x)), X: sort.StringSlice(x)}
		sort.Stable(p)
		return NewV(p.Perm)
	case AV:
		p := &permutationAV{Perm: permRange(len(x)), X: sortAV(x)}
		sort.Stable(p)
		return NewV(p.Perm)
	default:
		return errf("<x : x not an array (%s)", x.Type())
	}
}

// descend returns >x.
func descend(x V) V {
	p := ascend(x)
	if isErr(p) {
		return errs(">" + strings.TrimPrefix(p.Value.(errV).Error(), "<"))
	}
	reverseMut(p)
	return p
}

// search implements x$y.
func search(x V, y V) V {
	switch x := x.Value.(type) {
	case AB:
		if !sort.IsSorted(sortAB(x)) {
			return errDomain("x$y", "x is not ascending")
		}
		return searchAI(fromABtoAI(x).Value.(AI), y)
	case AI:
		if !sort.IsSorted(sort.IntSlice(x)) {
			return errDomain("x$y", "x is not ascending")
		}
		return searchAI(x, y)
	case AF:
		if !sort.IsSorted(sort.Float64Slice(x)) {
			return errDomain("x$y", "x is not ascending")
		}
		return searchAF(x, y)
	case AS:
		if !sort.IsSorted(sort.StringSlice(x)) {
			return errDomain("x$y", "x is not ascending")
		}
		return searchAS(x, y)
	case AV:
		if !sort.IsSorted(sortAV(x)) {
			return errDomain("x$y", "x is not ascending")
		}
		return searchAV(x, y)
	default:
		// should not happen
		return errType("x$y", "x", x)
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
	switch y := y.Value.(type) {
	case I:
		return NewI(searchAII(x, y))
	case F:
		return NewI(searchAIF(x, y))
	case AB:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAII(x, B2I(yi))
		}
		return NewV(r)
	case AI:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAII(x, I(yi))
		}
		return NewV(r)
	case AF:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAIF(x, F(yi))
		}
		return NewV(r)
	case array:
		r := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = sort.Search(len(x),
				func(i int) bool { return less(y.at(i), NewI(x[i])) })
		}
		return NewV(r)
	default:
		return NewI(x.Len())
	}
}

func searchAF(x AF, y V) V {
	switch y := y.Value.(type) {
	case I:
		return NewI(searchAFI(x, y))
	case F:
		return NewI(searchAFF(x, y))
	case AB:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAFI(x, B2I(yi))
		}
		return NewV(r)
	case AI:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAFI(x, I(yi))
		}
		return NewV(r)
	case AF:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchAFF(x, F(yi))
		}
		return NewV(r)
	case array:
		r := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = sort.Search(len(x),
				func(i int) bool { return less(y.at(i), NewF(x[i])) })
		}
		return NewV(r)
	default:
		return NewI(x.Len())
	}
}

func searchAS(x AS, y V) V {
	switch y := y.Value.(type) {
	case S:
		return NewI(searchASS(x, y))
	case AS:
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = searchASS(x, S(yi))
		}
		return NewV(r)
	case array:
		r := make(AI, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = sort.Search(len(x),
				func(i int) bool { return less(y.at(i), NewS(x[i])) })
		}
		return NewV(r)
	default:
		return NewI(x.Len())
	}
}

func searchAV(x AV, y V) V {
	switch yv := y.Value.(type) {
	case array:
		r := make(AI, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			r[i] = sort.Search(len(x),
				func(i int) bool { return less(yv.at(i), x[i]) })
		}
		return NewV(r)
	default:
		return NewI(sort.Search(len(x),
			func(i int) bool { return less(y, x[i]) }))

	}
}
