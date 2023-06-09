package goal

import (
	"sort"
)

// Less satisfies the specification of sort.Interface.
func (x *AB) Less(i, j int) bool {
	return x.elts[i] < x.elts[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AB) Swap(i, j int) {
	x.elts[i], x.elts[j] = x.elts[j], x.elts[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AI) Less(i, j int) bool {
	return x.elts[i] < x.elts[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AI) Swap(i, j int) {
	x.elts[i], x.elts[j] = x.elts[j], x.elts[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AF) Less(i, j int) bool {
	return x.elts[i] < x.elts[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AF) Swap(i, j int) {
	x.elts[i], x.elts[j] = x.elts[j], x.elts[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AS) Less(i, j int) bool {
	return x.elts[i] < x.elts[j]
}

// Swap satisfies the specification of sort.Interface.
func (x *AS) Swap(i, j int) {
	x.elts[i], x.elts[j] = x.elts[j], x.elts[i]
}

// Less satisfies the specification of sort.Interface.
func (x *AV) Less(i, j int) bool {
	return x.elts[i].LessT(x.elts[j])
}

// Swap satisfies the specification of sort.Interface.
func (x *AV) Swap(i, j int) {
	x.elts[i], x.elts[j] = x.elts[j], x.elts[i]
}

// Less satisfies the specification of sort.Interface.
func (d *D) Less(i, j int) bool {
	return d.values.Less(i, j)
}

// Swap satisfies the specification of sort.Interface.
func (d *D) Swap(i, j int) {
	d.keys.Swap(i, j)
	d.values.Swap(i, j)
}

// sortUp returns ^x.
func sortUp(ctx *Context, x V) V {
	xa, ok := x.bv.(Array)
	if !ok {
		switch xv := x.bv.(type) {
		case *D:
			return NewV(sortUpDictKeys(ctx, xv))
		default:
			return panicType("^X", "X", x)
		}
	}
	flags := xa.getFlags()
	if flags.Has(flagAscending) {
		return x
	}
	switch xv := xa.(type) {
	case *AB:
		xv = scloneAB(xv)
		if flags.Has(flagBool) {
			sortBools(xv.elts)
		} else {
			sortBytes(xv.elts)
		}
		xv.setFlags(flags | flagAscending)
		return NewV(xv)
	case *AI:
		xv = sortAI(ctx, xv)
		xv.setFlags(flags | flagAscending)
		return NewV(xv)
	case *AV:
		xa = xv.sclone()
		sort.Stable(xa)
		xa.setFlags(flags | flagAscending)
		return NewV(xa)
	default:
		xa = xa.sclone()
		sort.Sort(xa)
		xa.setFlags(flags | flagAscending)
		return NewV(xa)
	}
}

func sortBools(xs []byte) {
	var freq [2]int
	for _, xi := range xs {
		freq[xi]++
	}
	for i := range xs[:freq[0]] {
		xs[i] = 0
	}
	txs := xs[freq[0]:]
	for i := range txs {
		txs[i] = 1
	}
}

func sortBytes(xs []byte) {
	var freq [256]int
	for _, xi := range xs {
		freq[xi]++
	}
	i := 0
	for j, n := range freq {
		for k := i; k < i+n; k++ {
			xs[k] = byte(j)
		}
		i += n
	}
}

func sortAI(ctx *Context, xv *AI) *AI {
	if xv.Len() > 32 {
		min, max := minMaxAI(xv)
		span := max - min + 1
		if span == 1 {
			return xv
		}
		if span <= 256 {
			xv = scloneAI(xv)
			sortSmallInt64s(xv.elts, min)
			return xv
		}
		return radixSortAI(ctx, xv, min, max)
	}
	xv = scloneAI(xv)
	sort.Sort(xv)
	return xv
}

func sortSmallInt64s(xs []int64, min int64) {
	var freq [256]int
	for _, xi := range xs {
		freq[xi-min]++
	}
	i := 0
	for j, n := range freq {
		xk := int64(j) + min
		for k := i; k < i+n; k++ {
			xs[k] = xk
		}
		i += n
	}
}

func sortUpDictKeys(ctx *Context, d *D) *D {
	flags := d.keys.getFlags()
	if flags.Has(flagAscending) {
		return d
	}
	d = sortBy(ctx, d.values, d.keys)
	d.keys, d.values = d.values, d.keys
	d.keys.setFlags(flags | flagAscending)
	return d
}

type permutation[I integer] struct {
	Perm []I
	X    Array
}

func (p *permutation[I]) Len() int {
	return p.X.Len()
}

func (p *permutation[I]) Swap(i, j int) {
	p.Perm[i], p.Perm[j] = p.Perm[j], p.Perm[i]
}

func (p *permutation[I]) Less(i, j int) bool {
	return p.X.Less(int(p.Perm[i]), int(p.Perm[j]))
}

func permRange[I integer](n int) []I {
	r := make([]I, n)
	for i := range r {
		r[i] = I(i)
	}
	return r
}

// ascend returns <x.
func ascend(ctx *Context, x V) V {
	switch xv := x.bv.(type) {
	case Array:
		return ascendArray(ctx, xv)
	case *D:
		return NewV(sortUpDict(ctx, xv))
	default:
		return panicType("<X", "X", x)
	}
}

func ascendAB(xv *AB) V {
	if xv.IsBoolean() {
		if xv.Len() < 256 {
			return NewAB(ascendBools[byte](xv.elts))
		}
		return NewAI(ascendBools[int64](xv.elts))
	}
	if xv.Len() < 256 {
		p := permRange[byte](xv.Len())
		radixGradeUint8[byte](xv.elts, p)
		return NewAB(p)
	}
	p := permRange[int64](xv.Len())
	radixGradeUint8[int64](xv.elts, p)
	return NewAI(p)
}

func ascendBools[I integer](xs []byte) []I {
	var offsets [2]I
	for _, xi := range xs {
		offsets[xi]++
	}
	offsets[1] = offsets[0]
	offsets[0] = 0
	r := make([]I, len(xs))
	for i, xi := range xs {
		n := offsets[xi]
		offsets[xi]++
		r[n] = I(i)
	}
	return r
}

func ascendAI(ctx *Context, xv *AI) V {
	xlen := xv.Len()
	if ascending(xv) {
		if xlen < 256 {
			return NewAB(permRange[byte](xlen))
		}
		return NewAI(permRange[int64](xlen))
	}
	if xlen > 32 {
		min, max := minMaxAI(xv)
		span := max - min + 1
		if span == 1 {
			if xlen < 256 {
				return NewAB(permRange[byte](xlen))
			}
			return NewAI(permRange[int64](xlen))
		}
		if span <= 256 {
			return radixGradeSmallRange(ctx, xv, min, max)
		}
		return radixGradeAI(ctx, xv, min, max)
	}
	p := &permutation[byte]{Perm: permRange[byte](xlen), X: xv}
	sort.Stable(p)
	return NewAB(p.Perm)
}

func ascendArray(ctx *Context, x Array) V {
	switch xv := x.(type) {
	case *AB:
		return ascendAB(xv)
	case *AI:
		return ascendAI(ctx, xv)
	case Array:
		if x.Len() < 256 {
			p := &permutation[byte]{Perm: permRange[byte](xv.Len()), X: xv}
			if !ascending(xv) {
				sort.Stable(p)
			}
			return NewAB(p.Perm)
		}
		p := &permutation[int64]{Perm: permRange[int64](xv.Len()), X: xv}
		if !ascending(xv) {
			sort.Stable(p)
		}
		return NewAI(p.Perm)
	default:
		panic("ascendArray")
	}
}

func sortUpDict(ctx *Context, d *D) *D {
	flags := d.values.getFlags()
	if flags.Has(flagAscending) {
		return d
	}
	d = sortBy(ctx, d.keys, d.values)
	d.values.setFlags(flags | flagAscending)
	return d
}

func sortBy(ctx *Context, keys, values Array) *D {
	a := ascendArray(ctx, values)
	switch av := a.bv.(type) {
	case *AB:
		nk := keys.atBytes(av.elts)
		nv := values.atBytes(av.elts)
		return &D{keys: nk, values: nv}
	case *AI:
		nk := keys.atInt64s(av.elts)
		nv := values.atInt64s(av.elts)
		return &D{keys: nk, values: nv}
	default:
		panic("sortBy")
	}
}

// descend returns >x.
func descend(ctx *Context, x V) V {
	switch xv := x.bv.(type) {
	case Array:
		return descendArray(ctx, xv)
	case *D:
		return NewV(sortDownDict(ctx, xv))
	default:
		return panicType(">X", "X", x)
	}
}

func descendArray(ctx *Context, x Array) V {
	x = x.sclone()
	reverseMut(x)
	r := ascendArray(ctx, x).bv.(Array)
	reverseMut(r)
	return subtract(NewI(int64(r.Len())-1), NewV(r))
}

func sortDownDict(ctx *Context, d *D) *D {
	d.values.IncrRC()
	dsc := descendArray(ctx, d.values).bv.(*AI) // subtract nevers returns *AB
	d.values.DecrRC()
	nk := d.keys.atInt64s(dsc.elts)
	nv := d.values.atInt64s(dsc.elts)
	return &D{keys: nk, values: nv}
}

// search implements x$y.
func search(x V, y V) V {
	switch xv := x.bv.(type) {
	case *AB:
		if !ascending(xv) && !sort.IsSorted(xv) {
			return panics("X$y : non-ascending X")
		}
		xv.flags |= flagAscending
		return searchAB(xv, y)
	case *AI:
		if !ascending(xv) && !sort.IsSorted(xv) {
			return panics("X$y : non-ascending X")
		}
		xv.flags |= flagAscending
		return searchAI(xv, y)
	case *AF:
		if !ascending(xv) && !sort.IsSorted(xv) {
			return panics("X$y : non-ascending X")
		}
		xv.flags |= flagAscending
		return searchAF(xv, y)
	case *AS:
		if !ascending(xv) && !sort.IsSorted(xv) {
			return panics("X$y : non-ascending X")
		}
		xv.flags |= flagAscending
		return searchAS(xv, y)
	case *AV:
		if !ascending(xv) && !sort.IsSorted(xv) {
			return panics("X$y : non-ascending X")
		}
		xv.flags |= flagAscending
		return searchAV(xv, y)
	default:
		// should not happen
		return panicType("X$y", "x", x)
	}
}

func searchSlice[T ordered](x []T, y T) int64 {
	i, j := 0, len(x)
	for i < j {
		h := int(uint(i+j) >> 1)
		// i ≤ h < j
		if x[h] <= y {
			i = h + 1
		} else {
			j = h
		}
	}
	return int64(i)
}

func searchABI(x *AB, y int64) int64 {
	if y < 0 {
		return 0
	}
	if y >= 256 {
		return int64(x.Len())
	}
	return searchSlice(x.elts, byte(y))
}

func searchABB(x *AB, y byte) int64 {
	return searchSlice(x.elts, y)
}

func searchABF(x *AB, y float64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return float64(x.At(i)) > y }))
}

func searchAII(x *AI, y int64) int64 {
	return searchSlice(x.elts, y)
}

func searchAIF(x *AI, y float64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return float64(x.At(i)) > y }))
}

func searchAFI(x *AF, y int64) int64 {
	return int64(sort.Search(x.Len(), func(i int) bool { return x.At(i) > float64(y) }))
}

func searchAFF(x *AF, y float64) int64 {
	return searchSlice(x.elts, y)
}

func searchASS(x *AS, y S) int64 {
	return searchSlice(x.elts, string(y))
}

func searchAB(x *AB, y V) V {
	if y.IsI() {
		return NewI(searchABI(x, y.I()))
	}
	if y.IsF() {
		return NewI(searchABF(x, y.F()))
	}
	switch yv := y.bv.(type) {
	case Array:
		if x.Len() < 256 {
			return NewAB(searchABArray[byte](x, yv))
		}
		return NewAI(searchABArray[int64](x, yv))
	default:
		return NewI(int64(x.Len()))
	}
}

func searchABArray[I integer](x *AB, y Array) []I {
	switch yv := y.(type) {
	case *AB:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchABB(x, yi))
		}
		return r
	case *AI:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchABI(x, yi))
		}
		return r
	case *AF:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchABF(x, yi))
		}
		return r
	default:
		r := make([]I, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = I(sort.Search(x.Len(),
				func(j int) bool { return y.VAt(i).LessT(NewI(int64(x.At(j)))) }))
		}
		return r
	}
}

func searchAI(x *AI, y V) V {
	if y.IsI() {
		return NewI(searchAII(x, y.I()))
	}
	if y.IsF() {
		return NewI(searchAIF(x, y.F()))
	}
	switch yv := y.bv.(type) {
	case Array:
		if x.Len() < 256 {
			return NewAB(searchAIArray[byte](x, yv))
		}
		return NewAI(searchAIArray[int64](x, yv))
	default:
		return NewI(int64(x.Len()))
	}
}

func searchAIArray[I integer](x *AI, y Array) []I {
	switch yv := y.(type) {
	case *AB:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAII(x, int64(yi)))
		}
		return r
	case *AI:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAII(x, yi))
		}
		return r
	case *AF:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAIF(x, yi))
		}
		return r
	default:
		r := make([]I, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = I(sort.Search(x.Len(),
				func(j int) bool { return y.VAt(i).LessT(NewI(x.At(j))) }))
		}
		return r
	}
}

func searchAF(x *AF, y V) V {
	if y.IsI() {
		return NewI(searchAFI(x, y.I()))
	}
	if y.IsF() {
		return NewI(searchAFF(x, y.F()))
	}
	switch yv := y.bv.(type) {
	case Array:
		if x.Len() < 256 {
			return NewAB(searchAFArray[byte](x, yv))
		}
		return NewAI(searchAFArray[int64](x, yv))
	default:
		return NewI(int64(x.Len()))
	}
}

func searchAFArray[I integer](x *AF, y Array) []I {
	switch yv := y.(type) {
	case *AB:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAFI(x, int64(yi)))
		}
		return r
	case *AI:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAFI(x, yi))
		}
		return r
	case *AF:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchAFF(x, yi))
		}
		return r
	default:
		r := make([]I, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = I(sort.Search(x.Len(),
				func(j int) bool { return y.VAt(i).LessT(NewF(x.At(j))) }))
		}
		return r
	}
}

func searchAS(x *AS, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewI(searchASS(x, yv))
	case Array:
		if x.Len() < 256 {
			return NewAB(searchASArray[byte](x, yv))
		}
		return NewAI(searchASArray[int64](x, yv))
	default:
		return NewI(int64(x.Len()))
	}
}

func searchASArray[I integer](x *AS, y Array) []I {
	switch yv := y.(type) {
	case *AS:
		r := make([]I, yv.Len())
		for i, yi := range yv.elts {
			r[i] = I(searchASS(x, S(yi)))
		}
		return r
	default:
		r := make([]I, y.Len())
		for i := 0; i < y.Len(); i++ {
			r[i] = I(sort.Search(x.Len(),
				func(j int) bool { return y.VAt(i).LessT(NewS(x.At(j))) }))
		}
		return r
	}
}

func searchAV(x *AV, y V) V {
	switch yv := y.bv.(type) {
	case Array:
		if x.Len() < 256 {
			return NewAB(searchAVArray[byte](x, yv))
		}
		return NewAI(searchAVArray[int64](x, yv))
	default:
		return NewI(int64(sort.Search(x.Len(),
			func(i int) bool { return y.LessT(x.At(i)) })))

	}
}

func searchAVArray[I integer](x *AV, y Array) []I {
	r := make([]I, y.Len())
	for i := 0; i < y.Len(); i++ {
		r[i] = I(sort.Search(x.Len(),
			func(j int) bool { return y.VAt(i).LessT(x.At(j)) }))
	}
	return r
}
