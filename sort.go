package goal

import (
	"sort"
	"strings"
)

// Less satisfies the specification of sort.Interface.
func (x *AB) Less(i, j int) bool {
	return x.Slice[j] && !x.Slice[i]
}

// Swap satisfies the specification of sort.Interface.
func (x *AB) Swap(i, j int) {
	x.Slice[i], x.Slice[j] = x.Slice[j], x.Slice[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AI) Less(i, j int) bool {
	return x.Slice[i] < x.Slice[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AI) Swap(i, j int) {
	x.Slice[i], x.Slice[j] = x.Slice[j], x.Slice[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AF) Less(i, j int) bool {
	return x.Slice[i] < x.Slice[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AF) Swap(i, j int) {
	x.Slice[i], x.Slice[j] = x.Slice[j], x.Slice[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AS) Less(i, j int) bool {
	return x.Slice[i] < x.Slice[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AS) Swap(i, j int) {
	x.Slice[i], x.Slice[j] = x.Slice[j], x.Slice[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AV) Less(i, j int) bool {
	return x.Slice[i].LessT(x.Slice[j])
}

// Swap satisfies the specification of sort.Interface.
func (x *AV) Swap(i, j int) {
	x.Slice[i], x.Slice[j] = x.Slice[j], x.Slice[i]
}

// Less satisfies the specification of sort.Interface.
func (x *Dict) Less(i, j int) bool {
	return x.values.Less(i, j)
}

// Swap satisfies the specification of sort.Interface.
func (x *Dict) Swap(i, j int) {
	x.keys.Swap(i, j)
	x.values.Swap(i, j)
}

// sortUp returns ^x.
func sortUp(x V) V {
	xa, ok := x.value.(array)
	if !ok {
		d, ok := x.value.(*Dict)
		if ok {
			return NewV(sortUpDict(d))
		}
		return panicType("^x", "x", x)
	}
	flags := xa.getFlags()
	if flags.Has(flagAscending) {
		return x
	}
	xa = xa.shallowClone()
	switch xa.(type) {
	case *AV:
		sort.Stable(xa)
	default:
		sort.Sort(xa)
	}
	xa.setFlags(flags | flagAscending)
	return NewV(xa)
}

func sortUpDict(d *Dict) *Dict {
	flags := d.values.getFlags()
	if flags.Has(flagAscending) {
		return d
	}
	nk := d.keys.shallowClone()
	nv := d.values.shallowClone()
	initRC(nk)
	initRC(nv)
	nd := &Dict{keys: nk, values: nv}
	sort.Stable(nd)
	nv.setFlags(flags | flagAscending)
	return nd
}

type permutation struct {
	Perm []int64
	X    array
}

func (p *permutation) Len() int {
	return p.X.Len()
}

func (p *permutation) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutation) Less(i, j int) bool {
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
	case array:
		p := &permutation{Perm: permRange(xv.Len()), X: xv}
		if !xv.getFlags().Has(flagAscending) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	case *Dict:
		return NewV(sortUpDict(xv).keys)
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
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(xv) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAI(fromABtoAI(xv).value.(*AI), y)
	case *AI:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(xv) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAI(xv, y)
	case *AF:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(xv) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAF(xv, y)
	case *AS:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(xv) {
			return panicDomain("x$y", "x is not ascending")
		}
		xv.flags |= flagAscending
		return searchAS(xv, y)
	case *AV:
		if !xv.flags.Has(flagAscending) && !sort.IsSorted(xv) {
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
				func(j int) bool { return yv.at(i).LessT(NewI(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).LessT(NewF(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).LessT(NewS(x.At(j))) }))
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
				func(j int) bool { return yv.at(i).LessT(x.At(j)) }))
		}
		return NewAI(r)
	default:
		return NewI(int64(sort.Search(x.Len(),
			func(i int) bool { return y.LessT(x.At(i)) })))

	}
}
