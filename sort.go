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
	return bs[i].Less(bs[j])
}

func (bs sortVSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

// sortUp returns ^x.
func sortUp(x V) V {
	xa, ok := x.value.(array)
	if !ok {
		d, ok := x.value.(*Dict)
		if ok {
			if d.values.getFlags().Has(flagAscending) {
				return x
			}
			p := ascend(NewV(d.values)).value.(*AI)
			nk := d.keys.atIndices(p)
			nv := d.values.atIndices(p)
			nv.setFlags(nv.getFlags() | flagAscending)
			initRC(nk)
			initRC(nv)
			return NewV(&Dict{keys: nk, values: nv})
		}
		return panicType("^x", "x", x)
	}
	if xa.getFlags().Has(flagAscending) {
		return x
	}
	xa = xa.shallowClone()
	switch xv := xa.(type) {
	case *AB:
		sort.Sort(sortAB(xv.Slice))
		xv.flags |= flagAscending
		return NewV(xv)
	case *AF:
		sort.Float64s(xv.Slice)
		xv.flags |= flagAscending
		return NewV(xv)
	case *AI:
		sort.Sort(sortAI(xv.Slice))
		xv.flags |= flagAscending
		return NewV(xv)
	case *AS:
		sort.Strings(xv.Slice)
		xv.flags |= flagAscending
		return NewV(xv)
	case *AV:
		sort.Stable(sortVSlice(xv.Slice))
		xv.flags |= flagAscending
		return NewV(xv)
	default:
		panic("sortUp")
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
		if !xv.flags.Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *AF:
		p := &permutationAF{Perm: permRange(xv.Len()), X: sort.Float64Slice(xv.Slice)}
		if !xv.flags.Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *AI:
		p := &permutationAI{Perm: permRange(xv.Len()), X: sortAI(xv.Slice)}
		if !xv.flags.Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *AS:
		p := &permutationAS{Perm: permRange(xv.Len()), X: sort.StringSlice(xv.Slice)}
		if !xv.flags.Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *AV:
		p := &permutationAV{Perm: permRange(xv.Len()), X: sortVSlice(xv.Slice)}
		if !xv.flags.Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *Dict:
		p := ascend(NewV(xv.values)).value.(*AI)
		return NewV(xv.keys.atIndices(p))
	default:
		return panicType("<x", "x", x)
	}
}

// descend returns >x.
func descend(x V) V {
	p := ascend(x)
	if p.IsPanic() {
		return panics(">" + strings.TrimPrefix(string(p.value.(panicV)), "<"))
	}
	reverseMut(p)
	return p
}

// search implements x$y.
func search(x V, y V) V {
	switch xv := x.value.(type) {
	case *AB:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(sortAB(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAI(fromABtoAI(xv).value.(*AI), y)
	case *AI:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(sortAI(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAI(xv, y)
	case *AF:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(sort.Float64Slice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAF(xv, y)
	case *AS:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(sort.StringSlice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAS(xv, y)
	case *AV:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(sortVSlice(xv.Slice)) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
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
				func(j int) bool { return yv.at(i).Less(NewI(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).Less(NewF(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).Less(NewS(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).Less(x.At(j)) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(sort.Search(x.Len(),
			func(i int) bool { return y.Less(x.At(i)) })))

	}
}
