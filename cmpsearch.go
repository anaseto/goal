package goal

import (
	"bytes"
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
		if xv.getFlags().Has(flagDistinct) {
			return enumI(int64(xv.Len()))
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
			r := classifySortedSlice[byte, byte](xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		if xv.Len() < bruteForceBytes {
			r := classifyBrute(xv.elts)
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		return NewABWithRC(classifyBytes(xv.elts), reuseRCp(xv.rc))
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

func classifyBytes(xs []byte) []byte {
	var m [256]int
	var n int
	r := make([]byte, len(xs))
	for i, xi := range xs {
		c := m[xi]
		if c == 0 {
			r[i] = byte(n)
			m[xi] = n + 1
			n++
			continue
		}
		r[i] = byte(c - 1)
	}
	return r
}

func classifyInts[T integer](xs []int64, min, span int64) []T {
	// len(xs) <= MaxIntT so that n+1 fits in T
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
	// pre-condition: len(xs) < 256
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
	var n int
	prev := xs[0]
	r[0] = 0
	i := 1
	for _, xi := range xs[1:] {
		if xi != prev {
			n++
		}
		r[i] = I(n)
		prev = xi
		i++
	}
	return r
}

// distinct returns ?x.
func distinct(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return NewV(distinctDict(xv))
	case array:
		return NewV(distinctArray(xv))
	default:
		return panicType("?x", "x", x)
	}
}

func distinctDict(x *Dict) *Dict {
	if x.values.getFlags().Has(flagDistinct) {
		return x
	}
	x.values.IncrRC()
	mf := markFirsts(NewV(x.values))
	x.values.DecrRC()
	nk := replicate(mf, NewV(x.keys))
	nk.InitRC()
	nv := replicate(mf, NewV(x.values))
	nv.InitRC()
	return &Dict{keys: nk.value.(array), values: nv.value.(array)}
}

func distinctArray(x array) array {
	if x.Len() == 0 || x.getFlags().Has(flagDistinct) {
		return x
	}
	switch xv := x.(type) {
	case *AB:
		if xv.IsBoolean() {
			b := xv.At(0)
			for i := 1; i < xv.Len(); i++ {
				if xv.At(i) != b {
					return &AB{elts: []byte{b, xv.At(i)}, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
				}
			}
			return &AB{elts: []byte{b}, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
		}
		if xv.flags.Has(flagAscending) {
			r := distinctSortedSlice[byte](xv.elts)
			return &AB{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
		}
		if xv.Len() < bruteForceBytes {
			r := distinctBrute(xv.elts)
			return &AB{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
		}
		return &AB{elts: distinctBytes(xv.elts), rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
	case *AF:
		var r []float64
		if xv.flags.Has(flagAscending) {
			r = distinctSortedSlice[float64](xv.elts)
		} else {
			r = distinctSlice[float64](xv.elts, bruteForceNumeric)
		}
		return &AF{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
	case *AI:
		var r []int64
		min, span, ok := smallRange(xv)
		if ok {
			// fast path avoiding hash table
			r = distinctInts(xv.elts, min, span)
			return &AI{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
		}
		if xv.flags.Has(flagAscending) {
			r = distinctSortedSlice[int64](xv.elts)
		} else {
			r = distinctSlice[int64](xv.elts, bruteForceNumeric)
		}
		return &AI{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
	case *AS:
		var r []string
		if xv.flags.Has(flagAscending) {
			r = distinctSortedSlice[string](xv.elts)
		} else {
			r = distinctSlice[string](xv.elts, bruteForceGeneric)
		}
		return &AS{elts: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagDistinct}
	case *AV:
		xv.IncrRC()
		mf := markFirsts(NewV(xv))
		xv.DecrRC()
		return replicate(mf, NewV(xv)).value.(array)
	default:
		panic("distinctArray")
	}
}

func distinctInts(xs []int64, min, span int64) []int64 {
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
			m[xi+offset] = false
			r[n] = xi
			n++
		}
	}
	return r
}

func distinctBytes(xs []byte) []byte {
	var m [256]bool
	n := 0
	for _, xi := range xs {
		if !m[xi] {
			n++
			m[xi] = true
			continue
		}
	}
	r := make([]byte, n)
	n = 0
	for _, xi := range xs {
		if m[xi] {
			m[xi] = false
			r[n] = xi
			n++
		}
	}
	return r
}

func distinctSlice[T comparable](xs []T, bruteForceThreshold int) []T {
	if len(xs) <= bruteForceThreshold {
		return distinctBrute(xs)
	}
	r := []T{}
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

func distinctBrute[T comparable](xs []T) []T {
	r := []T{}
loop:
	for i, xi := range xs {
		for _, xj := range xs[:i] {
			if xi == xj {
				continue loop
			}
		}
		r = append(r, xi)
	}
	return r
}

func distinctSortedSlice[T comparable](xs []T) []T {
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
func markFirsts(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, markFirsts(NewV(xv.values)))
	case array:
		if xv.Len() == 0 {
			return x
		}
		if xv.getFlags().Has(flagDistinct) {
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
			ctx := NewContext()
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
		if isBI(xv) {
			return NewI(b2I(bytes.IndexByte(y.elts, byte(xv)) >= 0))
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
	i := findSortedSlice(y.elts, x)
	return i < y.Len() && y.At(i) == x
}

func findSortedSlice[T ordered](x []T, y T) int {
	i, j := 0, len(x)
	for i < j {
		h := int(uint(i+j) >> 1)
		// i ≤ h < j
		if x[h] < y {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
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
			m := findAIboolsIdx[int64](y.elts)
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
		xlen := int64(xv.Len())
		asc := y.flags.Has(flagAscending)
		if ylen > smallRangeLen && xlen > smallRangeLen || xlen > bruteForceNumeric && ylen > 0 {
			// NOTE: heuristics here might need some adjustments:
			// we used one based on self-search functions, but
			// member of is more complicated, because there are two
			// variables (#x influences allocation, while #y
			// influences number of searches).
			var min, max int64
			if asc {
				min, max = y.elts[0], y.elts[ylen-1]
			} else {
				min, max = minMax(y)
			}
			span := max - min + 1
			if span < ylen+xlen+smallRangeSpan {
				// fast path avoiding hash table
				r := memberOfII(xv.elts, y.elts, min, max)
				return newABbWithRC(r, reuseRCp(xv.rc))
			}
		}
		if asc && ylen > numericSortedLen {
			r := make([]byte, xv.Len())
			for i, xi := range xv.elts {
				r[i] = b2B(memberISortedAI(xi, y))
			}
			return newABbWithRC(r, reuseRCp(xv.rc))
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
	i := findSortedSlice(y.elts, x)
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

const bruteForceGenericDyad = 24

func memberOfArray(x, y array) V {
	if x.Len() > bruteForceGenericDyad && y.Len() > bruteForceGenericDyad {
		ctx := NewContext()
		return memberOf(each2String(ctx, x), each2String(ctx, y))
	}
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
		if xv.getFlags().Has(flagDistinct) {
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
		return withoutVArray(x, y)
	default:
		return panicType("x^y", "x", x)
	}
}

func withoutVArray(x, y V) V {
	r := memberOf(y, x)
	switch bres := r.value.(type) {
	case *AB:
		for i, b := range bres.elts {
			bres.elts[i] = 1 - b
		}
	}
	return replicate(r, y)
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
		return withKeys(x, yv)
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
			return arrayProtoV(d.keys)
		}
		return d.keys.at(int(i))
	}
	return vArrayAtV(d.keys, idx)
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
			return NewAI([]int64{int64(len(s)), 0})
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

func findAB(x *AB, y V) V {
	if y.IsI() {
		yv := y.I()
		if isBI(yv) {
			n := bytes.IndexByte(x.elts, byte(yv))
			if n == -1 {
				n = x.Len()
			}
			return NewI(int64(n))
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
		if x.Len() < bruteForceBytes {
			r := make([]byte, y.Len())
			findSlicesBrute(x.elts, yv.elts, r)
			return NewAB(r)
		}
		if x.Len() < 256 {
			r := findBsBs[byte](x.elts, yv.elts)
			return NewAB(r)
		}
		r := findBsBs[int64](x.elts, yv.elts)
		return NewAI(r)
	case *AI:
		if x.Len() < 256 {
			r := findBsIs[byte](x.elts, yv.elts)
			return NewAB(r)
		}
		r := findBsIs[int64](x.elts, yv.elts)
		return NewAI(r)
	case *AF:
		return find(fromABtoAF(x), y)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findSlicesBrute[T comparable, I integer](xs, ys []T, r []I) {
	xlen := I(len(xs))
loop:
	for i, yi := range ys {
		for j, xi := range xs {
			if yi == xi {
				r[i] = I(j)
				continue loop
			}
		}
		r[i] = xlen
	}
}

func findBsBs[I integer](xs []byte, ys []byte) []I {
	// len(xs) <= MaxIntT so that i+1 fits in T
	var m [256]I
	for i := len(xs) - 1; i >= 0; i-- {
		m[xs[i]] = I(i) + 1
	}
	xlen := I(len(xs))
	r := make([]I, len(ys))
	for i, yi := range ys {
		if m[yi] != 0 {
			r[i] = m[yi] - 1
		} else {
			r[i] = xlen
		}
	}
	return r
}

func findBsIs[I integer](xs []byte, ys []int64) []I {
	// len(xs) <= MaxIntT so that i+1 fits in T
	var m [256]I
	for i := len(xs) - 1; i >= 0; i-- {
		m[xs[i]] = I(i) + 1
	}
	xlen := I(len(xs))
	r := make([]I, len(ys))
	for i, yi := range ys {
		if yi >= 0 && yi < 256 && m[yi] != 0 {
			r[i] = m[yi] - 1
		} else {
			r[i] = xlen
		}
	}
	return r
}

func findAI(x *AI, y V) V {
	if y.IsI() {
		yv := y.I()
		if x.flags.Has(flagAscending) && x.Len() > numericSortedLen {
			return NewI(findIs(x.elts, yv))
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
			if x.Len() < 256 {
				r := findIsBbs[byte](x.elts, yv.elts)
				return NewABWithRC(r, reuseRCp(yv.rc))
			}
			r := findIsBbs[int64](x.elts, yv.elts)
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		if x.Len() < 256 {
			r := findIsBs[byte](x.elts, yv.elts)
			return NewABWithRC(r, reuseRCp(yv.rc))
		}
		r := findIsBs[int64](x.elts, yv.elts)
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AI:
		xlen := int64(x.Len())
		ylen := int64(yv.Len())
		asc := x.flags.Has(flagAscending)
		if xlen > smallRangeLen && ylen > smallRangeLen || ylen > bruteForceNumeric && xlen > 0 {
			// NOTE: heuristics here might need some adjustments:
			// we used one based on self-search functions, but find
			// is more complicated, because there are two variables
			// (#x influences allocation, while #y influences
			// number of searches).
			var min, max int64
			if asc {
				min, max = x.elts[0], x.elts[xlen-1]
			} else {
				min, max = minMax(x)
			}
			span := max - min + 1
			if span < xlen+ylen+smallRangeSpan {
				// fast path avoiding hash table
				if xlen < 256 {
					r := findIsIs[byte](x.elts, yv.elts, min, max)
					return NewABWithRC(r, reuseRCp(yv.rc))
				}
				r := findIsIs[int64](x.elts, yv.elts, min, max)
				return NewAIWithRC(r, reuseRCp(yv.rc))
			}
		}
		if asc && xlen > numericSortedLen {
			if xlen < 256 {
				r := findSortedIsIs[byte](x.elts, yv.elts)
				return NewABWithRC(r, reuseRCp(yv.rc))
			}
			r := findSortedIsIs[int64](x.elts, yv.elts)
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		if xlen < 256 {
			return NewABWithRC(findSlices[int64, byte](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
		}
		return NewAIWithRC(findSlices[int64, int64](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
	case *AF:
		return findAF(toAF(x).value.(*AF), y)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findSortedIsIs[I integer](xs, ys []int64) []I {
	r := make([]I, len(ys))
	for i, yi := range ys {
		r[i] = I(findIs(xs, yi))
	}
	return r
}

func findAIboolsIdx[I integer](xs []int64) (m [2]I) {
	xlen := I(len(xs))
	m[0], m[1] = xlen, xlen
loop:
	for i, xi := range xs {
		switch {
		case xi == 1:
			if m[1] == xlen {
				m[1] = I(i)
				if m[0] < xlen {
					break loop
				}
			}
		case xi == 0:
			if m[0] == xlen {
				m[0] = I(i)
				if m[1] < xlen {
					break loop
				}
			}
		}
	}
	return
}

func findSlices[T comparable, I integer](xs, ys []T, bruteForceThreshold int) []I {
	r := make([]I, len(ys))
	xlen := I(len(xs))
	if len(ys) <= bruteForceThreshold || len(xs) <= bruteForceThreshold {
		findSlicesBrute(xs, ys, r)
		return r
	}
	m := imapSlice[T, I](xs)
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

func imapSlice[T comparable, I integer](xs []T) map[T]I {
	m := map[T]I{}
	for i, xi := range xs {
		_, ok := m[xi]
		if !ok {
			m[xi] = I(i)
			continue
		}
	}
	return m
}

func findIs(x []int64, y int64) int64 {
	xlen := len(x)
	i := findSortedSlice(x, y)
	if i < xlen && x[i] == y {
		return int64(i)
	}
	return int64(xlen)
}

func findIsBbs[I integer](xs []int64, ys []byte) []I {
	m := findAIboolsIdx[I](xs)
	r := make([]I, len(ys))
	for i, yi := range ys {
		r[i] = m[yi]
	}
	return r
}

func findIsBs[I integer](xs []int64, ys []byte) []I {
	// len(xs) <= MaxIntT so that i+1 fits in T
	var m [256]I
	xlen := I(len(xs))
	r := make([]I, len(ys))
	for i, xi := range xs {
		if xi >= 0 && xi < 256 && m[xi] == 0 {
			m[xi] = I(i) + 1
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

func findIsIs[I integer](xs, ys []int64, min, max int64) []I {
	xlen := I(len(xs))
	r := make([]I, len(ys))
	offset := -min
	m := make([]I, max-min+1)
	for i, xi := range xs {
		c := m[xi+offset]
		if c == 0 {
			m[xi+offset] = I(i) + 1
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
		if x.Len() < 256 {
			return NewABWithRC(findSlices[float64, byte](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
		}
		return NewAIWithRC(findSlices[float64, int64](x.elts, yv.elts, bruteForceNumeric), reuseRCp(yv.rc))
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findSs(x []string, y string) int64 {
	xlen := len(x)
	i := findSortedSlice(x, y)
	if i < xlen && x[i] == y {
		return int64(i)
	}
	return int64(xlen)
}

func findAS(x *AS, y V) V {
	switch yv := y.value.(type) {
	case S:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceGeneric/4 {
			return NewI(findSs(x.elts, string(yv)))
		}
		for i, xi := range x.elts {
			if S(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AS:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceGeneric/4 {
			if x.Len() < 256 {
				r := findSortedSsSs[byte](x.elts, yv.elts)
				return NewABWithRC(r, reuseRCp(yv.rc))
			}
			r := findSortedSsSs[int64](x.elts, yv.elts)
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		if x.Len() < 256 {
			return NewABWithRC(findSlices[string, byte](x.elts, yv.elts, bruteForceGeneric), reuseRCp(yv.rc))
		}
		return NewAIWithRC(findSlices[string, int64](x.elts, yv.elts, bruteForceGeneric), reuseRCp(yv.rc))
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findSortedSsSs[I integer](xs, ys []string) []I {
	r := make([]I, len(ys))
	for i, yi := range ys {
		r[i] = I(findSs(xs, yi))
	}
	return r
}

func findArrays(x, y array) V {
	if x.Len() > bruteForceGenericDyad && y.Len() > bruteForceGenericDyad {
		ctx := NewContext()
		return find(each2String(ctx, x), each2String(ctx, y))
	}
	if x.Len() < 256 {
		return NewAB(findArraysBrute[byte](x, y))
	}
	return NewAI(findArraysBrute[int64](x, y))
}

func findArraysBrute[I integer](x, y array) []I {
	r := make([]I, y.Len())
	for i := range r {
		r[i] = I(x.Len())
	}
	for i := 0; i < y.Len(); i++ {
		for j := 0; j < x.Len(); j++ {
			if y.at(i).Matches(x.at(j)) {
				r[i] = I(j)
				break
			}
		}
	}
	return r
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
