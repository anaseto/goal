package goal

import (
	"math"
	"sort"
	"strings"
)

// Matches returns true if the two values match like in x~y.
func (x V) Matches(y V) bool {
	switch x.kind {
	case valNil:
		return y.kind == valNil
	case valInt:
		return y.kind == valInt && x.n == y.n ||
			y.kind == valFloat && float64(x.n) == y.F()
	case valFloat:
		return y.kind == valInt && x.F() == float64(y.n) ||
			y.kind == valFloat && x.F() == y.F()
	case valVariadic:
		return y.kind == valVariadic && x.n == y.n
	case valLambda:
		// XXX: match lambdas: match the string representations?
		// Currently, self-search operations may use a more tolerant
		// comparison for lambdas by using stringification. Adding
		// context information in Match would be a bit inconvenient.
		// Comparing lambdas is not a common thing, so it does not
		// matter much in practice.
		return y.kind == valLambda && x.n == y.n
	case valPanic:
		return y.kind == valPanic && x.value.Matches(y.value)
	default:
		return y.kind == valBoxed && x.value.Matches(y.value)
	}
}

// Thresholds for brute force. Should all be < 256 because it simplifies
// reasoning about integer size to chose for some results in classify,
// ocurrence count, ...
const bruteForceGeneric = 32
const bruteForceNumeric = 255
const bruteForceBytes = 12
const numericSortedLen = 64
const smallRangeLen = 16
const smallRangeSpan = 16

func smallRange(xv *AI) (min, span int64, ok bool) {
	xlen := int64(xv.Len())
	if xlen > smallRangeLen {
		var max int64
		min, max = minMax(xv)
		span = max - min + 1
		if span < 2*xlen+smallRangeSpan {
			return min, span, true
		}
	}
	return
}

// classify returns %x.
func classify(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, classify(ctx, NewV(xv.values)))
	case array:
		if xv.Len() == 0 {
			return x
		}
		if xv.getFlags().Has(flagUnique) {
			return rangeI(int64(xv.Len()))
		}
	default:
		return panicType("%X", "X", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.IsBoolean() {
			if xv.At(0) == 0 {
				return x
			}
			return not(x)
		}
		if xv.flags.Has(flagAscending) {
			if xv.Len() < 256 {
				r := classifySortedSlice[byte, byte](xv.elts)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			r := classifySortedSlice[byte, int64](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < bruteForceBytes {
			r := classifyBrute(xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < 256 {
			return NewABWithRC(classifyBytes[byte](xv.elts), reuseRCp(xv.rc))
		}
		return NewAIWithRC(classifyBytes[int64](xv.elts), reuseRCp(xv.rc))
	case *AI:
		if xv.flags.Has(flagAscending) {
			if xv.Len() < 256 {
				r := classifySortedSlice[int64, byte](xv.elts)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			r := classifySortedSlice[int64, int64](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		min, span, ok := smallRange(xv)
		if ok {
			// fast path avoiding hash table
			if span < 256 || xv.Len() < 256 {
				r := classifyInts[byte](xv.elts, min, span)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			r := classifyInts[int64](xv.elts, min, span)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() <= bruteForceNumeric {
			r := classifyBrute(xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		r := classifySlice[int64, int64](xv.elts)
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AF:
		if xv.flags.Has(flagAscending) {
			if xv.Len() < 256 {
				r := classifySortedSlice[float64, byte](xv.elts)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			r := classifySortedSlice[float64, int64](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() <= bruteForceNumeric {
			r := classifyBrute(xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		r := classifySlice[float64, int64](xv.elts)
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AS:
		if xv.flags.Has(flagAscending) {
			if xv.Len() < 256 {
				r := classifySortedSlice[string, byte](xv.elts)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			r := classifySortedSlice[string, int64](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() <= bruteForceGeneric {
			r := classifyBrute(xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < 256 {
			r := classifySlice[string, byte](xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		r := classifySlice[string, int64](xv.elts)
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.elts {
				ss[i] = xi.Sprint(ctx)
			}
			if xv.Len() <= bruteForceGeneric {
				r := classifyBrute(ss)
				return NewABWithRC(r, reuseRCp(xv.rc))
			}
			if xv.Len() < 256 {
				return NewAB(classifySlice[string, byte](ss))
			}
			return NewAI(classifySlice[string, int64](ss))
		}
		r := make([]byte, xv.Len())
		n := byte(0)
	loop:
		for i, xi := range xv.elts {
			for j := range xv.elts[:i] {
				if xi.Matches(xv.At(j)) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return NewAB(r)
	default:
		panic("classify")
	}
}

func classifyBytes[T integer](xs []byte) []T {
	var m [256]T
	var n T
	r := make([]T, len(xs))
	for i, xi := range xs {
		c := m[xi]
		if c == 0 {
			r[i] = n
			m[xi] = n + 1
			n++
			continue
		}
		r[i] = c - 1
	}
	return r
}

func classifyInts[T integer](xs []int64, min, span int64) []T {
	r := make([]T, len(xs))
	var n T
	offset := -min
	m := make([]T, span)
	for i, xi := range xs {
		c := m[xi+offset]
		if c == 0 {
			r[i] = n
			m[xi+offset] = n + 1
			n++
			continue
		}
		r[i] = c - 1
	}
	return r
}

func classifySlice[T comparable, I integer](xs []T) []I {
	r := make([]I, len(xs))
	m := map[T]I{}
	n := I(0)
	for i, xi := range xs {
		c, ok := m[xi]
		if !ok {
			r[i] = n
			m[xi] = n
			n++
			continue
		}
		r[i] = c
	}
	return r
}

func classifyBrute[T comparable](xs []T) []byte {
	r := make([]byte, len(xs))
	var n byte
loop:
	for i, xi := range xs {
		for j, xj := range xs[:i] {
			if xi == xj {
				r[i] = r[j]
				continue loop
			}
		}
		r[i] = n
		n++
	}
	return r
}

func classifySortedSlice[T comparable, I integer](xs []T) []I {
	r := make([]I, len(xs))
	var n I
	prev := xs[0]
	r[0] = 0
	i := 1
	for _, xi := range xs[1:] {
		if xi != prev {
			n++
		}
		r[i] = n
		prev = xi
		i++
	}
	return r
}

// uniq returns ?x.
func uniq(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		xv.values.IncrRC()
		mf := markFirsts(ctx, NewV(xv.values))
		xv.values.DecrRC()
		nk := replicate(mf, NewV(xv.keys))
		nk.InitRC()
		nv := replicate(mf, NewV(xv.values))
		nv.InitRC()
		return NewV(&Dict{keys: nk.value.(array), values: nv.value.(array)})
	case array:
		if xv.Len() == 0 || xv.getFlags().Has(flagUnique) {
			return x
		}
	default:
		return panicType("?x", "x", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.IsBoolean() {
			b := xv.At(0)
			for i := 1; i < xv.Len(); i++ {
				if xv.At(i) != b {
					return NewV(&AB{elts: []byte{b, xv.At(i)}, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
				}
			}
			return NewV(&AB{elts: []byte{b}, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
		}
		if xv.flags.Has(flagAscending) {
			r := uniqSortedSlice[byte](xv.elts)
			return NewV(&AB{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
		}
		if xv.Len() < bruteForceBytes {
			r := make([]byte, xv.Len())
			uniqBrute(xv.elts, r)
			return NewV(&AB{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
		}
		return NewV(&AB{elts: uniqBytes(xv.elts), rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AF:
		var r []float64
		if xv.flags.Has(flagAscending) {
			r = uniqSortedSlice[float64](xv.elts)
		} else {
			r = uniqSlice[float64](xv.elts, bruteForceNumeric)
		}
		return NewV(&AF{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AI:
		var r []int64
		min, span, ok := smallRange(xv)
		if ok {
			// fast path avoiding hash table
			r = uniqInts(xv.elts, min, span)
			return NewV(&AI{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
		}
		if xv.flags.Has(flagAscending) {
			r = uniqSortedSlice[int64](xv.elts)
		} else {
			r = uniqSlice[int64](xv.elts, bruteForceNumeric)
		}
		return NewV(&AI{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AS:
		var r []string
		if xv.flags.Has(flagAscending) {
			r = uniqSortedSlice[string](xv.elts)
		} else {
			r = uniqSlice[string](xv.elts, bruteForceGeneric)
		}
		return NewV(&AS{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AV:
		r := []V{}
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.elts {
				ss[i] = xi.Sprint(ctx)
			}
			m := map[string]struct{}{}
			for i, s := range ss {
				_, ok := m[s]
				if !ok {
					r = append(r, xv.At(i))
					m[s] = struct{}{}
					continue
				}
			}
			return NewV(canonicalAV(&AV{elts: r, rc: xv.rc, flags: xv.flags | flagUnique}))
		}
	loop:
		for i, xi := range xv.elts {
			for _, xj := range xv.elts[:i] {
				if xi.Matches(xj) {
					continue loop
				}
			}
			r = append(r, xi)
		}
		return NewV(canonicalAV(&AV{elts: r, rc: xv.rc, flags: xv.flags | flagUnique}))
	default:
		panic("uniq")
	}
}

func uniqInts(xs []int64, min, span int64) []int64 {
	offset := -min
	m := make([]bool, span)
	n := 0
	for _, xi := range xs {
		c := m[xi+offset]
		if !c {
			n++
			m[xi+offset] = true
			continue
		}
	}
	r := make([]int64, n)
	n = 0
	for _, xi := range xs {
		c := m[xi+offset]
		if c {
			r[n] = xi
			m[xi+offset] = false
			n++
		}
	}
	return r
}

func uniqBytes(xs []byte) []byte {
	var m [256]bool
	n := 0
	for _, xi := range xs {
		c := m[xi]
		if !c {
			n++
			m[xi] = true
			continue
		}
	}
	r := make([]byte, n)
	n = 0
	for _, xi := range xs {
		c := m[xi]
		if c {
			r[n] = xi
			m[xi] = false
			n++
		}
	}
	return r
}

func uniqSlice[T comparable](xs []T, bruteForceThreshold int) []T {
	r := []T{}
	if len(xs) <= bruteForceThreshold {
		uniqBrute(xs, r)
		return r
	}
	m := map[T]struct{}{}
	for _, xi := range xs {
		_, ok := m[xi]
		if !ok {
			r = append(r, xi)
			m[xi] = struct{}{}
			continue
		}
	}
	return r
}

func uniqBrute[T comparable](xs, r []T) {
loop:
	for i, xi := range xs {
		for _, xj := range xs[:i] {
			if xi == xj {
				continue loop
			}
		}
		r = append(r, xi)
	}
}

func uniqSortedSlice[T comparable](xs []T) []T {
	prev := xs[0]
	r := []T{prev}
	n := 1
	for _, xi := range xs[1:] {
		if xi == prev {
			continue
		}
		r = append(r, xi)
		prev = xi
		n++
	}
	return r
}

// Mark Firsts returns firsts x.
func markFirsts(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, markFirsts(ctx, NewV(xv.values)))
	case array:
		if xv.Len() == 0 {
			return x
		}
		if xv.getFlags().Has(flagUnique) {
			r := make([]byte, xv.Len())
			for i := range r {
				r[i] = 1
			}
			return newABb(r)
		}
	default:
		return panicType("firsts X", "X", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.IsBoolean() {
			r := make([]byte, xv.Len())
			r[0] = 1
			x0 := xv.At(0)
			for i := 1; i < xv.Len(); i++ {
				if xv.At(i) != x0 {
					r[i] = 1
					break
				}
			}
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		if xv.flags.Has(flagAscending) {
			r := markFirstsSortedSlice[byte](xv.elts)
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < bruteForceBytes {
			r := make([]byte, xv.Len())
			markFirstsBrute(xv.elts, r)
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		return newABbWithRC(markFirstsBytes(xv.elts), reuseRCp(xv.rc))
	case *AF:
		var r []byte
		if xv.flags.Has(flagAscending) {
			r = markFirstsSortedSlice[float64](xv.elts)
		} else {
			r = markFirstsSlice[float64](xv.elts, bruteForceNumeric)
		}
		return newABbWithRC(r, reuseRCp(xv.rc))
	case *AI:
		var r []byte
		if xv.flags.Has(flagAscending) {
			r = markFirstsSortedSlice[int64](xv.elts)
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		min, span, ok := smallRange(xv)
		if ok {
			// fast path avoiding hash table
			r = markFirstsInts(xv.elts, min, span)
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		r = markFirstsSlice[int64](xv.elts, bruteForceNumeric)
		return newABbWithRC(r, reuseRCp(xv.rc))
	case *AS:
		var r []byte
		if xv.flags.Has(flagAscending) {
			r = markFirstsSortedSlice[string](xv.elts)
		} else {
			r = markFirstsSlice[string](xv.elts, bruteForceGeneric)
		}
		return newABbWithRC(r, reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.elts {
				ss[i] = xi.Sprint(ctx)
			}
			return newABbWithRC(markFirstsSlice[string](ss, bruteForceGeneric), reuseRCp(xv.rc))
		}
		r := make([]byte, xv.Len())
	loop:
		for i, xi := range xv.elts {
			for _, xj := range xv.elts[:i] {
				if xi.Matches(xj) {
					continue loop
				}
			}
			r[i] = 1
		}
		return newABbWithRC(r, reuseRCp(xv.rc))
	default:
		panic("firsts")
	}
}

func markFirstsInts(xs []int64, min, span int64) []byte {
	r := make([]byte, len(xs))
	offset := -min
	m := make([]bool, span)
	for i, xi := range xs {
		c := m[xi+offset]
		if !c {
			r[i] = 1
			m[xi+offset] = true
			continue
		}
	}
	return r
}

func markFirstsBytes(xs []byte) []byte {
	var m [256]bool
	r := make([]byte, len(xs))
	for i, xi := range xs {
		c := m[xi]
		if !c {
			r[i] = 1
			m[xi] = true
			continue
		}
	}
	return r
}

func markFirstsSlice[T comparable](xs []T, bruteForceThreshold int) []byte {
	r := make([]byte, len(xs))
	if len(xs) <= bruteForceThreshold {
		markFirstsBrute(xs, r)
		return r
	}
	m := map[T]struct{}{}
	for i, s := range xs {
		_, ok := m[s]
		if !ok {
			r[i] = 1
			m[s] = struct{}{}
			continue
		}
	}
	return r
}

func markFirstsBrute[T comparable](xs []T, r []byte) {
loop:
	for i, xi := range xs {
		for _, xj := range xs[:i] {
			if xi == xj {
				continue loop
			}
		}
		r[i] = 1
	}
}

func markFirstsSortedSlice[T comparable](xs []T) []byte {
	r := make([]byte, len(xs))
	prev := xs[0]
	r[0] = 1
	i := 1
	for _, xi := range xs[1:] {
		if xi != prev {
			r[i] = 1
		}
		prev = xi
		i++
	}
	return r
}

// memberOf returns x in y.
func memberOf(x, y V) V {
	// XXX: maybe we should make a switch first on x, instead of y, to
	// handle simple unboxed cases firsts. Same comment for find.
	if xv, ok := x.value.(*Dict); ok {
		return newDictValues(xv.keys, memberOf(NewV(xv.values), y))
	}
	switch yv := y.value.(type) {
	case S:
		return containedInS(x, string(yv))
	case *AB:
		return memberOfAB(x, yv)
	case *AF:
		return memberOfAF(x, yv)
	case *AI:
		return memberOfAI(x, yv)
	case *AS:
		return memberOfAS(x, yv)
	case *AV:
		return memberOfAV(x, yv)
	case *Dict:
		return memberOf(x, NewV(yv.values))
	default:
		return panicType("x in y", "y", y)
	}
}

func memberOfAB(x V, y *AB) V {
	if x.IsI() {
		xv := x.I()
		for _, yi := range y.elts {
			if xv == int64(yi) {
				return NewI(1)
			}
		}
		return NewI(0)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return NewI(0)
		}
		return memberOfAB(NewI(int64(x.F())), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := memberOfBB(xv.elts, y.elts)
		return newABb(r)
	case *AI:
		r := memberOfIB(xv.elts, y.elts)
		return newABb(r)
	case *AF:
		return memberOf(x, fromABtoAF(y))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfBB(xs []byte, ys []byte) []byte {
	var m [256]bool
	for _, yi := range ys {
		m[yi] = true
	}
	r := make([]byte, len(xs))
	for i, xi := range xs {
		if m[xi] {
			r[i] = 1
		}
	}
	return r
}

func memberOfIB(xs []int64, ys []byte) []byte {
	var m [256]bool
	for _, yi := range ys {
		m[yi] = true
	}
	r := make([]byte, len(xs))
	for i, xi := range xs {
		if xi >= 0 && xi < 256 && m[xi] {
			r[i] = 1
		}
	}
	return r
}

func memberOfSlice[T comparable](xs []T, ys []T, bruteForceThreshold int) []byte {
	r := make([]byte, len(xs))
	if len(xs) <= bruteForceThreshold || len(ys) <= bruteForceThreshold {
		memberOfSliceBrute(xs, ys, r)
		return r
	}
	m := bmapSlice[T](ys)
	for i, xi := range xs {
		_, ok := m[xi]
		r[i] = b2B(ok)
	}
	return r
}

func memberOfSliceBrute[T comparable](xs []T, ys []T, r []byte) {
	for i, xi := range xs {
		for _, yi := range ys {
			if xi == yi {
				r[i] = 1
				break
			}
		}
	}
}

func bmapSlice[T comparable](xs []T) map[T]struct{} {
	m := map[T]struct{}{}
	for _, xi := range xs {
		_, ok := m[xi]
		if !ok {
			m[xi] = struct{}{}
			continue
		}
	}
	return m
}

func memberOfAF(x V, y *AF) V {
	if x.IsI() {
		for _, yi := range y.elts {
			if float64(x.I()) == yi {
				return NewI(1)
			}
		}
		return NewI(0)
	}
	if x.IsF() {
		for _, yi := range y.elts {
			if x.F() == yi {
				return NewI(1)
			}
		}
		return NewI(0)
	}
	switch xv := x.value.(type) {
	case *AB:
		return memberOfAF(fromABtoAF(xv), y)
	case *AI:
		return memberOfAF(toAF(xv), y)
	case *AF:
		return newABbWithRC(memberOfSlice[float64](xv.elts, y.elts, bruteForceNumeric), reuseRCp(xv.rc))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberISortedAI(x int64, y *AI) bool {
	ylen := y.Len()
	i := sort.Search(ylen, func(j int) bool { return y.At(j) >= x })
	return i < ylen && y.At(i) == x
}

func memberOfAI(x V, y *AI) V {
	if x.IsI() {
		if y.flags.Has(flagAscending) && y.Len() > numericSortedLen {
			return NewI(b2I(memberISortedAI(x.I(), y)))
		}
		for _, yi := range y.elts {
			if x.I() == yi {
				return NewI(1)
			}
		}
		return NewI(0)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return NewI(0)
		}
		if y.flags.Has(flagAscending) && y.Len() > numericSortedLen {
			return NewI(b2I(memberISortedAI(int64(x.F()), y)))
		}
		for _, yi := range y.elts {
			if x.F() == float64(yi) {
				return NewI(1)
			}
		}
		return NewI(0)
	}
	switch xv := x.value.(type) {
	case *AB:
		if xv.IsBoolean() {
			m := findAIboolsIdx(y.elts)
			xlen := int64(x.Len())
			mb := [2]byte{b2B(m[0] < xlen), b2B(m[1] < xlen)}
			r := make([]byte, xlen)
			for i, xi := range xv.elts {
				r[i] = mb[xi]
			}
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		r := memberOfBI(xv.elts, y.elts)
		return newABbWithRC(r, reuseRCp(xv.rc))
	case *AI:
		ylen := int64(y.Len())
		if y.flags.Has(flagAscending) && ylen > numericSortedLen {
			r := make([]byte, xv.Len())
			for i, xi := range xv.elts {
				r[i] = b2B(memberISortedAI(xi, y))
			}
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		xlen := int64(xv.Len())
		if ylen > smallRangeLen && xlen > smallRangeLen || xlen > bruteForceNumeric {
			// NOTE: heuristics here might need some adjustments:
			// we used one based on self-search functions, but
			// member of is more complicated, because there are two
			// variables (#x influences allocation, while #y
			// influences number of searches).
			min, max := minMax(y)
			span := max - min + 1
			if span < ylen+xlen+smallRangeSpan {
				// fast path avoiding hash table
				r := memberOfII(xv.elts, y.elts, min, max)
				return newABbWithRC(r, reuseRCp(xv.rc))
			}
		}
		return newABbWithRC(memberOfSlice[int64](xv.elts, y.elts, bruteForceNumeric), reuseRCp(xv.rc))
	case *AF:
		return memberOf(x, toAF(y))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfBI(xs []byte, ys []int64) []byte {
	var m [256]bool
	for _, yi := range ys {
		if yi >= 0 && yi < 256 {
			m[yi] = true
		}
	}
	r := make([]byte, len(xs))
	for i, xi := range xs {
		if m[xi] {
			r[i] = 1
		}
	}
	return r
}

func memberOfII(xs, ys []int64, min, max int64) []byte {
	r := make([]byte, len(xs))
	offset := -min
	m := make([]byte, max-min+1)
	for _, yi := range ys {
		m[yi+offset] = 1
	}
	for i, xi := range xs {
		if xi < min || xi > max {
			continue
		}
		r[i] = m[xi+offset]
	}
	return r
}

func memberSOfAS(x string, y *AS) bool {
	ylen := y.Len()
	i := sort.Search(ylen, func(j int) bool { return y.At(j) >= x })
	return i < ylen && y.At(i) == x
}

func memberOfAS(x V, y *AS) V {
	switch xv := x.value.(type) {
	case S:
		if y.flags.Has(flagAscending) && y.Len() > bruteForceGeneric/4 {
			return NewI(b2I(memberSOfAS(string(xv), y)))
		}
		for _, yi := range y.elts {
			if string(xv) == yi {
				return NewI(1)
			}
		}
		return NewI(0)
	case *AS:
		if y.flags.Has(flagAscending) && y.Len() > bruteForceGeneric/4 {
			r := make([]byte, xv.Len())
			for i, xi := range xv.elts {
				r[i] = b2B(memberSOfAS(xi, y))
			}
			return newABbWithRC(r, reuseRCp(xv.rc))
		}
		return newABbWithRC(memberOfSlice[string](xv.elts, y.elts, bruteForceGeneric), reuseRCp(xv.rc))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfAV(x V, y *AV) V {
	switch xv := x.value.(type) {
	case array:
		return memberOfArray(xv, y)
	default:
		for _, yi := range y.elts {
			if x.Matches(yi) {
				return NewI(1)
			}
		}
		return NewI(0)
	}
}

func memberOfArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make([]byte, x.Len())
	for i := 0; i < x.Len(); i++ {
		for j := 0; j < y.Len(); j++ {
			if x.at(i).Matches(y.at(j)) {
				r[i] = 1
				break
			}
		}
	}
	return newABb(r)
}

// OccurrenceCount returns ocount x.
func occurrenceCount(ctx *Context, x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, occurrenceCount(ctx, NewV(xv.values)))
	case array:
		if xv.Len() == 0 {
			return x
		}
		if xv.getFlags().Has(flagUnique) {
			r := make([]byte, xv.Len())
			return newABb(r)
		}
	default:
		return panicType("ocount X", "X", x)
	}
	// TODO: occurrence count could often return []bytes instead of []int64
	switch xv := x.value.(type) {
	case *AB:
		if xv.IsBoolean() {
			r := make([]int64, xv.Len())
			var counts [2]int64
			for i, xi := range xv.elts {
				r[i] = counts[xi]
				counts[xi]++
			}
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.flags.Has(flagAscending) {
			r := occurrenceCountSortedSlice[byte](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < bruteForceBytes {
			r := occurrenceCountSlice[byte](xv.elts, bruteForceNumeric)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		r := occurrenceCountBytes(xv.elts)
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AI:
		var r []int64
		if xv.flags.Has(flagAscending) {
			r = occurrenceCountSortedSlice[int64](xv.elts)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		min, span, ok := smallRange(xv)
		if ok {
			// fast path avoiding hash table
			r = occurrenceCountInts(xv.elts, min, span)
			return NewAIWithRC(r, reuseRCp(xv.rc))
		}
		r = occurrenceCountSlice[int64](xv.elts, bruteForceNumeric)
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AF:
		var r []int64
		if xv.flags.Has(flagAscending) {
			r = occurrenceCountSortedSlice[float64](xv.elts)
		} else {
			r = occurrenceCountSlice[float64](xv.elts, bruteForceNumeric)
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AS:
		var r []int64
		if xv.flags.Has(flagAscending) {
			r = occurrenceCountSortedSlice[string](xv.elts)
		} else {
			r = occurrenceCountSlice[string](xv.elts, bruteForceGeneric)
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > (2*bruteForceGeneric)/3 {
			ss := make([]string, xv.Len())
			for i, xi := range xv.elts {
				ss[i] = xi.Sprint(ctx)
			}
			return NewAIWithRC(occurrenceCountSlice[string](ss, bruteForceGeneric), reuseRCp(xv.rc))
		}
		r := make([]byte, xv.Len())
	loop:
		for i, xi := range xv.elts {
			for j := i - 1; j >= 0; j-- {
				if xi.Matches(xv.At(j)) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return NewAB(r)
	default:
		panic("ocount")
	}
}

func occurrenceCountBytes(xs []byte) []int64 {
	var m [256]int64
	r := make([]int64, len(xs))
	for i, xi := range xs {
		c := m[xi]
		if c == 0 {
			m[xi] = 1
			r[i] = 0
			continue
		}
		m[xi]++
		r[i] = c
	}
	return r
}

func occurrenceCountInts(xs []int64, min, span int64) []int64 {
	r := make([]int64, len(xs))
	offset := -min
	m := make([]int64, span)
	for i, xi := range xs {
		c := m[xi+offset]
		if c == 0 {
			m[xi+offset] = 1
			r[i] = 0
			continue
		}
		m[xi+offset]++
		r[i] = c
	}
	return r
}

func occurrenceCountSlice[T comparable](xs []T, bruteForceThreshold int) []int64 {
	r := make([]int64, len(xs))
	if len(xs) <= (2*bruteForceThreshold)/3 {
		for i, xi := range xs {
			var n int64
			for _, xj := range xs[:i] {
				if xi == xj {
					n++
				}
			}
			r[i] = n
		}
		return r
	}
	m := map[T]int64{}
	for i, xi := range xs {
		c, ok := m[xi]
		if !ok {
			m[xi] = 0
			continue
		}
		m[xi] = c + 1
		r[i] = c + 1
	}
	return r
}

func occurrenceCountSortedSlice[T comparable](xs []T) []int64 {
	r := make([]int64, len(xs))
	prev := xs[0]
	r[0] = 0
	i := 1
	for _, xi := range xs[1:] {
		if xi != prev {
			r[i] = 0
		} else {
			r[i] = r[i-1] + 1
		}
		prev = xi
		i++
	}
	return r
}

// without returns x^y.
func without(x, y V) V {
	if x.IsI() {
		return windows(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i^y : i non-integer (%g)", x.F())
		}
		return windows(int64(x.F()), y)
	}
	switch xv := x.value.(type) {
	case S:
		if xv == "" {
			return trimSpaces(y)
		}
		return trim(xv, y)
	case array:
		_, ok := y.value.(array)
		if !ok {
			d, ok := y.value.(*Dict)
			if ok {
				return withoutDict(x, d)
			}
			return panicType("X^Y", "Y", y)
		}
		r := memberOf(y, x)
		switch bres := r.value.(type) {
		case *AB:
			for i, b := range bres.elts {
				bres.elts[i] = 1 - b
			}
		}
		return replicate(r, y)
	default:
		return panicType("x^y", "x", x)
	}
}

func withoutDict(x V, y *Dict) V {
	r := memberOf(NewV(y.keys), x)
	switch bres := r.value.(type) {
	case *AB:
		for i, b := range bres.elts {
			bres.elts[i] = 1 - b
		}
	}
	return NewDict(replicate(r, NewV(y.keys)), replicate(r, NewV(y.values)))
}

// intersection implements keep x#y.
func intersection(x array, y V) V {
	switch yv := y.value.(type) {
	case array:
		return replicate(memberOf(y, NewV(x)), y)
	case *Dict:
		return takeKeys(x, yv)
	default:
		return panicType("X#Y", "Y", y)
	}
}

// find returns x?y.
func find(x, y V) V {
	if yv, ok := y.value.(*Dict); ok {
		return newDictValues(yv.keys, find(x, NewV(yv.values)))
	}
	switch xv := x.value.(type) {
	case S:
		return findS(xv, y)
	case *AB:
		return findAB(xv, y)
	case *AF:
		return findAF(xv, y)
	case *AI:
		return findAI(xv, y)
	case *AS:
		return findAS(xv, y)
	case *AV:
		return findAV(xv, y)
	case *Dict:
		return findDict(xv, y)
	default:
		return panicType("x?y", "x", x)
	}
}

func findDict(d *Dict, y V) V {
	idx := find(NewV(d.values), y)
	l := d.keys.Len()
	if idx.IsI() {
		i := idx.I()
		if i == int64(l) {
			switch d.keys.(type) {
			case *AB:
				return NewF(math.NaN())
			case *AI:
				return NewF(math.NaN())
			case *AF:
				return NewF(math.NaN())
			default:
				return NewS("")
			}
		}
		return d.keys.at(int(i))
	}
	idxv := idx.value.(*AI)
	switch keys := d.keys.(type) {
	case *AS:
		r := make([]string, idxv.Len())
		for j := range r {
			i := idxv.At(j)
			if i == int64(l) {
				r[j] = ""
			} else {
				r[j] = keys.At(int(i))
			}
		}
		return NewAS(r)
	default:
		var zero V
		switch d.keys.(type) {
		case *AB:
			zero = NewF(math.NaN())
		case *AI:
			zero = NewF(math.NaN())
		case *AF:
			zero = NewF(math.NaN())
		default:
			zero = NewS("")
		}
		r := make([]V, idxv.Len())
		for j := range r {
			i := idxv.At(j)
			if i == int64(l) {
				r[j] = zero
			} else {
				r[j] = d.keys.at(int(i))
			}
		}
		return Canonical(NewAV(r))
	}
}

func findArray(x array, y V) V {
	switch xv := x.(type) {
	case *AB:
		return findAB(xv, y)
	case *AF:
		return findAF(xv, y)
	case *AI:
		return findAI(xv, y)
	case *AS:
		return findAS(xv, y)
	case *AV:
		return findAV(xv, y)
	default:
		panic("findArray")
	}
}

func findS(s S, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewI(int64(strings.Index(string(s), string(yv))))
	case *rx:
		loc := yv.Regexp.FindStringIndex(string(s))
		if loc == nil {
			return NewAI([]int64{int64(len(s)), int64(len(s))})
		}
		return NewAI([]int64{int64(loc[0]), int64(loc[1] - loc[0])})
	case *AS:
		r := make([]int64, yv.Len())
		for i, ss := range yv.elts {
			r[i] = int64(strings.Index(string(s), string(ss)))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := findS(s, yi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	default:
		return panicType("s?y", "y", y)
	}
}

func imapSlice[T comparable](xs []T) map[T]int64 {
	m := map[T]int64{}
	for i, xi := range xs {
		_, ok := m[xi]
		if !ok {
			m[xi] = int64(i)
			continue
		}
	}
	return m
}

func findAB(x *AB, y V) V {
	if y.IsI() {
		for i, xi := range x.elts {
			if int64(xi) == y.I() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
	if y.IsF() {
		if !isI(y.F()) {
			return NewI(int64(x.Len()))
		}
		return findAB(x, NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
	case *AB:
		if y.Len() < bruteForceBytes {
			r := make([]byte, y.Len())
			findSlicesBrute(x.elts, yv.elts, r)
			return NewAB(r)
		}
		r := findBB(x.elts, yv.elts)
		return NewAI(r)
	case *AI:
		r := findBI(x.elts, yv.elts)
		return NewAI(r)
	case *AF:
		return find(fromABtoAF(x), y)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findBB(xs []byte, ys []byte) []int64 {
	var m [256]int64
	for i, xi := range xs {
		m[xi] = int64(i) + 1
	}
	ylen := int64(len(ys))
	r := make([]int64, ylen)
	for i, yi := range ys {
		if m[yi] != 0 {
			r[i] = m[yi] - 1
		} else {
			r[i] = ylen
		}
	}
	return r
}

func findBI(xs []byte, ys []int64) []int64 {
	var m [256]int64
	for i, xi := range xs {
		m[xi] = int64(i) + 1
	}
	ylen := int64(len(ys))
	r := make([]int64, ylen)
	for i, yi := range ys {
		if yi >= 0 && yi < 256 && m[yi] != 0 {
			r[i] = m[yi] - 1
		} else {
			r[i] = ylen
		}
	}
	return r
}

func findSlices[T comparable](xs, ys []T, bruteForceThreshold int) []int64 {
	r := make([]int64, len(ys))
	xlen := int64(len(xs))
	if len(ys) <= bruteForceThreshold || len(xs) <= bruteForceThreshold {
		findSlicesBrute(xs, ys, r)
		return r
	}
	m := imapSlice[T](xs)
	for i, yi := range ys {
		j, ok := m[yi]
		if ok {
			r[i] = j
		} else {
			r[i] = xlen
		}
	}
	return r
}

func findSlicesBrute[T comparable, I integer](xs, ys []T, r []I) {
	xlen := I(len(xs))
	for i, yi := range ys {
		for j, xi := range xs {
			if yi == xi {
				r[i] = I(j)
				break
			}
		}
		r[i] = xlen
	}
}

func findAF(x *AF, y V) V {
	if y.IsI() {
		for i, xi := range x.elts {
			if xi == float64(y.I()) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	if y.IsF() {
		for i, xi := range x.elts {
			if float64(xi) == y.F() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
	switch yv := y.value.(type) {
	case *AB:
		return findAF(x, fromABtoAF(yv))
	case *AI:
		return findAF(x, toAF(yv))
	case *AF:
		return NewAIWithRC(findSlices[float64](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findAIboolsIdx(xs []int64) (m [2]int64) {
	xlen := int64(len(xs))
	m[0], m[1] = xlen, xlen
loop:
	for i, xi := range xs {
		switch {
		case xi == 1:
			if m[1] == xlen {
				m[1] = int64(i)
				if m[0] < xlen {
					break loop
				}
			}
		case xi == 0:
			if m[0] == xlen {
				m[0] = int64(i)
				if m[1] < xlen {
					break loop
				}
			}
		}
	}
	return
}

func findAII(x *AI, y int64) int64 {
	xlen := x.Len()
	i := int64(sort.Search(xlen, func(i int) bool { return x.At(i) >= y }))
	if i < int64(xlen) && x.At(int(i)) == y {
		return i
	}
	return int64(xlen)
}

func findAI(x *AI, y V) V {
	if y.IsI() {
		yv := y.I()
		if x.flags.Has(flagAscending) && x.Len() > numericSortedLen {
			return NewI(findAII(x, yv))
		}
		for i, xi := range x.elts {
			if xi == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return NewI(int64(x.Len()))
		}
		return findAI(x, NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
	case *AB:
		if yv.IsBoolean() {
			m := findAIboolsIdx(x.elts)
			r := make([]int64, yv.Len())
			for i, yi := range yv.elts {
				r[i] = m[yi]
			}
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		r := findIB(x.elts, yv.elts)
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AI:
		xlen := int64(x.Len())
		if x.flags.Has(flagAscending) && xlen > numericSortedLen {
			r := make([]int64, yv.Len())
			for i, yi := range yv.elts {
				r[i] = findAII(x, yi)
			}
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		ylen := int64(yv.Len())
		if xlen > smallRangeLen && ylen > smallRangeLen || ylen > bruteForceNumeric {
			// NOTE: heuristics here might need some adjustments:
			// we used one based on self-search functions, but find
			// is more complicated, because there are two variables
			// (#x influences allocation, while #y influences
			// number of searches).
			min, max := minMax(x)
			span := max - min + 1
			if span < xlen+ylen+smallRangeSpan {
				// fast path avoiding hash table
				r := findII(x.elts, yv.elts, min, max)
				return NewAIWithRC(r, reuseRCp(yv.rc))
			}
		}
		return NewAIWithRC(findSlices[int64](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
	case *AF:
		return findAF(toAF(x).value.(*AF), y)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findIB(xs []int64, ys []byte) []int64 {
	var m [256]int64
	xlen := int64(len(xs))
	r := make([]int64, len(ys))
	for i, xi := range xs {
		if xi >= 0 && xi < 256 && m[xi] == 0 {
			m[xi] = int64(i) + 1
		}
	}
	for i, yi := range ys {
		if m[yi] > 0 {
			r[i] = m[yi] - 1
		} else {
			r[i] = xlen
		}
	}
	return r
}

func findII(xs, ys []int64, min, max int64) []int64 {
	xlen := int64(len(xs))
	r := make([]int64, len(ys))
	offset := -min
	m := make([]int64, max-min+1)
	for i, xi := range xs {
		c := m[xi+offset]
		if c == 0 {
			m[xi+offset] = int64(i) + 1
			continue
		}
	}
	for i, yi := range ys {
		if yi < min || yi > max {
			r[i] = xlen
			continue
		}
		j := m[yi+offset]
		if j > 0 {
			r[i] = j - 1
		} else {
			r[i] = xlen
		}
	}
	return r
}

func findASS(x *AS, y S) int64 {
	xlen := x.Len()
	i := int64(sort.Search(xlen, func(i int) bool { return S(x.At(i)) >= y }))
	if i < int64(xlen) && x.At(int(i)) == string(y) {
		return i
	}
	return int64(xlen)
}

func findAS(x *AS, y V) V {
	switch yv := y.value.(type) {
	case S:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceGeneric/4 {
			return NewI(findASS(x, yv))
		}
		for i, xi := range x.elts {
			if S(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AS:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceGeneric/4 {
			r := make([]int64, yv.Len())
			for i, yi := range yv.elts {
				r[i] = findASS(x, S(yi))
			}
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		return NewAIWithRC(findSlices[string](x.elts, yv.elts, bruteForceGeneric), reuseRCp(yv.rc))
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findArrays(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make([]int64, y.Len())
	for i := range r {
		r[i] = int64(x.Len())
	}
	for i := 0; i < y.Len(); i++ {
		for j := 0; j < x.Len(); j++ {
			if y.at(i).Matches(x.at(j)) {
				r[i] = int64(j)
				break
			}
		}
	}
	return NewAI(r)
}

func findAV(x *AV, y V) V {
	switch yv := y.value.(type) {
	case array:
		return findArrays(x, yv)
	default:
		for i, xi := range x.elts {
			if y.Matches(xi) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
}
