package goal

import (
	"sort"
	"strings"
)

type sortAI []int64

func (bs sortAI) Len() int {
	return len(bs)
}

func (bs sortAI) Less(i, j int) bool {
	return bs[i] < bs[j]
}

func (bs sortAI) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

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

type sortVSlice []V

func (bs sortVSlice) Len() int {
	return len(bs)
}

func (bs sortVSlice) Less(i, j int) bool {
	return less(bs[i], bs[j])
}

func (bs sortVSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func less(x, y V) bool {
	switch x.kind {
	case valInt:
		return lessI(x, y)
	case valFloat:
		return lessF(x, y)
	case valVariadic:
		return y.kind == valVariadic && x.n < y.n
	case valLambda:
		return y.kind == valLambda && x.n < y.n
	}
	switch xv := x.value.(type) {
	case S:
		return lessS(x, y)
	case *AB:
		if xv.Len() == 0 {
			return Length(y) > 0
		}
		return lessAB(x, y)
	case *AF:
		if xv.Len() == 0 {
			return Length(y) > 0
		}
		return lessAF(x, y)
	case *AI:
		if xv.Len() == 0 {
			return Length(y) > 0
		}
		return lessAI(x, y)
	case *AS:
		if xv.Len() == 0 {
			return Length(y) > 0
		}
		return lessAS(x, y)
	case *AV:
		if xv.Len() == 0 {
			return Length(y) > 0
		}
		return lessAV(x, y)
	case derivedVerb:
		yv, ok := y.value.(derivedVerb)
		return ok && xv.Fun < yv.Fun ||
			xv.Fun == yv.Fun && less(xv.Arg, yv.Arg)
	case projection:
		yv, ok := y.value.(projection)
		return ok && less(xv.Fun, yv.Fun) ||
			Match(xv.Fun, yv.Fun) && less(NewAV(xv.Args), NewAV(yv.Args))
	case projectionFirst:
		yv, ok := y.value.(projectionFirst)
		return ok && less(xv.Fun, yv.Fun) ||
			Match(xv.Fun, yv.Fun) && less(xv.Arg, yv.Arg)
	case projectionMonad:
		yv, ok := y.value.(projectionMonad)
		return ok && less(xv.Fun, yv.Fun)
	case *errV:
		yv, ok := y.value.(*errV)
		return ok && less(xv.V, yv.V)
	default:
		return false
	}
}

func lessF(x V, y V) bool {
	xv := x.F()
	if y.IsI() {
		return xv < float64(y.I())
	}
	if y.IsF() {
		return xv < y.F()
	}
	switch yv := y.value.(type) {
	case *AB:
		if yv.Len() == 0 {
			return false
		}
		return xv < b2f(yv.At(0)) || xv == b2f(yv.At(0)) && yv.Len() > 1
	case *AF:
		if yv.Len() == 0 {
			return false
		}
		return xv < yv.At(0) || xv == yv.At(0) && yv.Len() > 1
	case *AI:
		if yv.Len() == 0 {
			return false
		}
		return xv < float64(yv.At(0)) || xv == float64(yv.At(0)) && yv.Len() > 1
	case *AV:
		if yv.Len() == 0 {
			return false
		}
		return lessF(x, yv.At(0)) || !less(yv.At(0), x) && yv.Len() > 1
	default:
		return false
	}
}

func lessI(x V, y V) bool {
	xv := x.I()
	if y.IsI() {
		return xv < y.I()
	}
	if y.IsF() {
		return float64(xv) < y.F()
	}
	switch yv := y.value.(type) {
	case *AB:
		if yv.Len() == 0 {
			return false
		}
		return xv < b2i(yv.At(0)) || xv == b2i(yv.At(0)) && yv.Len() > 1
	case *AF:
		if yv.Len() == 0 {
			return false
		}
		return float64(xv) < yv.At(0) || float64(xv) == yv.At(0) && yv.Len() > 1
	case *AI:
		if yv.Len() == 0 {
			return false
		}
		return xv < yv.At(0) || xv == yv.At(0) && yv.Len() > 1
	case *AV:
		if yv.Len() == 0 {
			return false
		}
		return lessI(x, yv.At(0)) || !less(yv.At(0), x) && yv.Len() > 1
	default:
		return false
	}
}

func lessS(x V, y V) bool {
	xv := x.value.(S)
	switch yv := y.value.(type) {
	case S:
		return xv < yv
	case *AS:
		if yv.Len() == 0 {
			return false
		}
		return string(xv) < yv.At(0) || string(xv) == yv.At(0) && yv.Len() > 1
	case *AV:
		if yv.Len() == 0 {
			return false
		}
		return lessS(x, yv.At(0)) || !less(yv.At(0), x) && yv.Len() > 1
	default:
		return false
	}
}

func lessAB(x V, y V) bool {
	xv := x.value.(*AB)
	if y.IsI() {
		return !lessI(y, x)
	}
	if y.IsF() {
		return !lessF(y, x)
	}
	switch yv := y.value.(type) {
	case *AB:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if !xv.At(i) && yv.At(i) {
				return true
			}
			if xv.At(i) && !yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AF:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if b2f(xv.At(i)) < yv.At(i) {
				return true
			}
			if b2f(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AI:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if b2i(xv.At(i)) < yv.At(i) {
				return true
			}
			if b2i(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AV:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(NewI(b2i(xv.At(i))), yv.At(i)) {
				return true
			}
			if less(yv.At(i), NewI(b2i(xv.At(i)))) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	default:
		return false
	}
}

func lessAI(x V, y V) bool {
	xv := x.value.(*AI)
	if y.IsI() {
		return !lessI(y, x)
	}
	if y.IsF() {
		return !lessF(y, x)
	}
	switch yv := y.value.(type) {
	case *AB:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < b2i(yv.At(i)) {
				return true
			}
			if xv.At(i) > b2i(yv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AF:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if float64(xv.At(i)) < yv.At(i) {
				return true
			}
			if float64(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AI:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AV:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(NewI(xv.At(i)), yv.At(i)) {
				return true
			}
			if less(yv.At(i), NewI(xv.At(i))) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	default:
		return false
	}
}

func lessAF(x V, y V) bool {
	xv := x.value.(*AF)
	if y.IsI() {
		return !lessI(y, x)
	}
	if y.IsF() {
		return !lessF(y, x)
	}
	switch yv := y.value.(type) {
	case *AB:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < b2f(yv.At(i)) {
				return true
			}
			if xv.At(i) > b2f(yv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AF:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AI:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < float64(yv.At(i)) {
				return true
			}
			if xv.At(i) > float64(yv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AV:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(NewF(xv.At(i)), yv.At(i)) {
				return true
			}
			if less(yv.At(i), NewF(xv.At(i))) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	default:
		return false
	}
}

func lessAS(x V, y V) bool {
	xv := x.value.(*AS)
	switch yv := y.value.(type) {
	case S:
		return !lessS(y, x)
	case *AS:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AV:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(NewS(xv.At(i)), yv.At(i)) {
				return true
			}
			if less(yv.At(i), NewS(xv.At(i))) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	default:
		return false
	}
}

func lessAV(x V, y V) bool {
	xv := x.value.(*AV)
	if y.IsI() {
		return less(xv.At(0), y)
	}
	if y.IsF() {
		return less(xv.At(0), y)
	}
	switch yv := y.value.(type) {
	case *AB:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(xv.At(i), NewI(b2i(yv.At(i)))) {
				return true
			}
			if less(NewI(b2i(yv.At(i))), xv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AF:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(xv.At(i), NewF(yv.At(i))) {
				return true
			}
			if less(NewF(yv.At(i)), xv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AI:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(xv.At(i), NewI(yv.At(i))) {
				return true
			}
			if less(NewI(yv.At(i)), xv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	case *AV:
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if less(xv.At(i), yv.At(i)) {
				return true
			}
			if less(yv.At(i), xv.At(i)) {
				return false
			}
		}
		return xv.Len() < yv.Len()
	default:
		return false
	}
}

// sortUp returns ^x.
func sortUp(x V) V {
	x = cloneShallow(x)
	switch xv := x.value.(type) {
	case *AB:
		sort.Stable(sortAB(xv.Slice))
		return NewV(xv)
	case *AF:
		sort.Stable(sort.Float64Slice(xv.Slice))
		return NewV(xv)
	case *AI:
		sort.Stable(sortAI(xv.Slice))
		return NewV(xv)
	case *AS:
		sort.Stable(sort.StringSlice(xv.Slice))
		return NewV(xv)
	case *AV:
		sort.Stable(sortVSlice(xv.Slice))
		return NewV(xv)
	default:
		return panicf("^x : x not an array (%s)", x.Type())
	}
}

type permutationAV struct {
	Perm []int64
	X    sortVSlice
}

func (p *permutationAV) Len() int {
	return p.X.Len()
}

func (p *permutationAV) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAV) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

type permutationAB struct {
	Perm []int64
	X    sortAB
}

func (p *permutationAB) Len() int {
	return p.X.Len()
}

func (p *permutationAB) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAB) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

type permutationAI struct {
	Perm []int64
	X    sortAI
}

func (p *permutationAI) Len() int {
	return p.X.Len()
}

func (p *permutationAI) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAI) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

type permutationAF struct {
	Perm []int64
	X    sort.Float64Slice
}

func (p *permutationAF) Len() int {
	return p.X.Len()
}

func (p *permutationAF) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAF) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

type permutationAS struct {
	Perm []int64
	X    sort.StringSlice
}

func (p *permutationAS) Len() int {
	return p.X.Len()
}

func (p *permutationAS) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutationAS) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

func permRange(n int) []int64 {
	r := make([]int64, n)
	for i := range r {
		r[i] = int64(i)
	}
	return r
}

// ascend returns <x.
func ascend(x V) V {
	switch xv := x.value.(type) {
	case *AB:
		p := &permutationAB{Perm: permRange(xv.Len()), X: sortAB(xv.Slice)}
		sort.Stable(p)
		return NewAI(p.Perm)
	case *AF:
		p := &permutationAF{Perm: permRange(xv.Len()), X: sort.Float64Slice(xv.Slice)}
		sort.Stable(p)
		return NewAI(p.Perm)
	case *AI:
		p := &permutationAI{Perm: permRange(xv.Len()), X: sortAI(xv.Slice)}
		sort.Stable(p)
		return NewAI(p.Perm)
	case *AS:
		p := &permutationAS{Perm: permRange(xv.Len()), X: sort.StringSlice(xv.Slice)}
		sort.Stable(p)
		return NewAI(p.Perm)
	case *AV:
		p := &permutationAV{Perm: permRange(xv.Len()), X: sortVSlice(xv.Slice)}
		sort.Stable(p)
		return NewAI(p.Perm)
	default:
		return panicf("<x : x not an array (%s)", x.Type())
	}
}

// descend returns >x.
func descend(x V) V {
	p := ascend(x)
	if p.IsPanic() {
		return panics(">" + strings.TrimPrefix(string(p.value.(S)), "<"))
	}
	reverseMut(p)
	return p
}

// search implements x$y.
func search(x V, y V) V {
	switch xv := x.value.(type) {
	case *AB:
		if !sort.IsSorted(sortAB(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		return searchAI(fromABtoAI(xv).value.(*AI), y)
	case *AI:
		if !sort.IsSorted(sortAI(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		return searchAI(xv, y)
	case *AF:
		if !sort.IsSorted(sort.Float64Slice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		return searchAF(xv, y)
	case *AS:
		if !sort.IsSorted(sort.StringSlice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		return searchAS(xv, y)
	case *AV:
		if !sort.IsSorted(sortVSlice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		return searchAV(xv, y)
	default:
		// should not happen
		return panicType("x$y", "x", x)
	}
}

func searchAII(x *AI, y int64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return x.At(i) > y }))
}

func searchAIF(x *AI, y float64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return float64(x.At(i)) > y }))
}

func searchAFI(x *AF, y int64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return x.At(i) > float64(y) }))
}

func searchAFF(x *AF, y float64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return x.At(i) > y }))
}

func searchASS(x *AS, y S) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return S(x.At(i)) > y }))
}

func searchAI(x *AI, y V) V {
	if y.IsI() {
		return NewI(searchAII(x, y.I()))
	}
	if y.IsF() {
		return NewI(searchAIF(x, y.F()))
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAII(x, b2i(yi))
		}
		return NewAI(r)
	case *AI:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAII(x, yi)
		}
		return NewAI(r)
	case *AF:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAIF(x, yi)
		}
		return NewAI(r)
	case array:
		r := make([]int64, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			r[i] = int64(sort.Search(x.Len(),
				func(i int) bool { return less(yv.at(i), NewI(x.At(i))) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(x.Len()))
	}
}

func searchAF(x *AF, y V) V {
	if y.IsI() {
		return NewI(searchAFI(x, y.I()))
	}
	if y.IsF() {
		return NewI(searchAFF(x, y.F()))
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAFI(x, b2i(yi))
		}
		return NewAI(r)
	case *AI:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAFI(x, yi)
		}
		return NewAI(r)
	case *AF:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchAFF(x, yi)
		}
		return NewAI(r)
	case array:
		r := make([]int64, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			r[i] = int64(sort.Search(x.Len(),
				func(i int) bool { return less(yv.at(i), NewF(x.At(i))) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(x.Len()))
	}
}

func searchAS(x *AS, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewI(searchASS(x, yv))
	case *AS:
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = searchASS(x, S(yi))
		}
		return NewAI(r)
	case array:
		r := make([]int64, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			r[i] = int64(sort.Search(x.Len(),
				func(i int) bool { return less(yv.at(i), NewS(x.At(i))) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(x.Len()))
	}
}

func searchAV(x *AV, y V) V {
	switch yv := y.value.(type) {
	case array:
		r := make([]int64, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			r[i] = int64(sort.Search(x.Len(),
				func(i int) bool { return less(yv.at(i), x.At(i)) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(sort.Search(x.Len(),
			func(i int) bool { return less(y, x.At(i)) })))

	}
}
