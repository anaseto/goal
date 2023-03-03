package goal

import (
	"math"
	"sort"
	"strings"
)

// Match returns true if the two values match like in x~y.
func Match(x, y V) bool {
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

const bruteForceGeneric = 32
const bruteForceNumeric = 128

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
		return panicType("%x", "x", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		if !xv.At(0) {
			return NewV(xv)
		}
		return not(x)
	case *AF:
		return NewAIWithRC(classifySlice[float64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AI:
		if xv.Len() > bruteForceNumeric {
			min, max := minMax(xv)
			n := int64(0)
			if max-min+1 < int64(xv.Len())+8 {
				r := make([]int64, xv.Len())
				// fast path avoiding hash table
				offset := -min
				m := make([]int64, max-min+1)
				for i, xi := range xv.Slice {
					c := m[xi+offset]
					if c == 0 {
						r[i] = n
						m[xi+offset] = n + 1
						n++
						continue
					}
					r[i] = c - 1
				}
				return NewAI(r)
			}
		}
		return NewAIWithRC(classifySlice[int64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AS:
		return NewAIWithRC(classifySlice[string](xv.Slice, bruteForceGeneric), reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.Slice {
				ss[i] = xi.Sprint(ctx)
			}
			return NewAI(classifySlice[string](ss, bruteForceGeneric))
		}
		r := make([]int64, xv.Len())
		n := int64(0)
	loop:
		for i, xi := range xv.Slice {
			for j := range xv.Slice[:i] {
				if Match(xi, xv.At(j)) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return NewAI(r)
	default:
		panic("classify")
	}
}

func classifySlice[T comparable](xs []T, bruteForceThreshold int) []int64 {
	r := make([]int64, len(xs))
	if len(xs) <= bruteForceThreshold {
		n := int64(0)
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
	m := map[T]int64{}
	n := int64(0)
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
		if xv.Len() == 0 {
			return NewV(xv)
		}
		b := xv.At(0)
		for i := 1; i < xv.Len(); i++ {
			if xv.At(i) != b {
				return NewV(&AB{Slice: []bool{b, xv.At(i)}, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
			}
		}
		return NewV(&AB{Slice: []bool{b}, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AF:
		r := uniqSlice[float64](xv.Slice, bruteForceNumeric)
		return NewV(&AF{Slice: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AI:
		r := uniqSlice[int64](xv.Slice, bruteForceNumeric)
		return NewV(&AI{Slice: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AS:
		r := uniqSlice[string](xv.Slice, bruteForceGeneric)
		return NewV(&AS{Slice: r, rc: reuseRCp(xv.rc), flags: xv.flags | flagUnique})
	case *AV:
		r := []V{}
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.Slice {
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
			return NewV(canonicalAV(&AV{Slice: r, rc: xv.rc, flags: xv.flags | flagUnique}))
		}
	loop:
		for i, xi := range xv.Slice {
			for _, xj := range xv.Slice[:i] {
				if Match(xi, xj) {
					continue loop
				}
			}
			r = append(r, xi)
		}
		return NewV(canonicalAV(&AV{Slice: r, rc: xv.rc, flags: xv.flags | flagUnique}))
	default:
		panic("uniq")
	}
}

func uniqSlice[T comparable](xs []T, bruteForceThreshold int) []T {
	r := []T{}
	if len(xs) <= bruteForceThreshold {
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
			r := make([]bool, xv.Len())
			for i := range r {
				r[i] = true
			}
			return NewAB(r)
		}
	default:
		return panicType("firsts x", "x", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		r[0] = true
		x0 := xv.At(0)
		for i := 1; i < xv.Len(); i++ {
			if xv.At(i) != x0 {
				r[i] = true
				break
			}
		}
		return NewAB(r)
	case *AF:
		return NewABWithRC(markFirstsSlice[float64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AI:
		return NewABWithRC(markFirstsSlice[int64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AS:
		return NewABWithRC(markFirstsSlice[string](xv.Slice, bruteForceGeneric), reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.Slice {
				ss[i] = xi.Sprint(ctx)
			}
			return NewABWithRC(markFirstsSlice[string](ss, bruteForceGeneric), reuseRCp(xv.rc))
		}
		r := make([]bool, xv.Len())
	loop:
		for i, xi := range xv.Slice {
			for _, xj := range xv.Slice[:i] {
				if Match(xi, xj) {
					continue loop
				}
			}
			r[i] = true
		}
		return NewAB(r)
	default:
		panic("firsts")
	}
}

func markFirstsSlice[T comparable](xs []T, bruteForceThreshold int) []bool {
	r := make([]bool, len(xs))
	if len(xs) <= bruteForceThreshold {
	loop:
		for i, xi := range xs {
			for _, xj := range xs[:i] {
				if xi == xj {
					continue loop
				}
			}
			r[i] = true
		}
		return r
	}
	m := map[T]struct{}{}
	for i, s := range xs {
		_, ok := m[s]
		if !ok {
			r[i] = true
			m[s] = struct{}{}
			continue
		}
	}
	return r
}

// memberOf returns x in y.
func memberOf(x, y V) V {
	if Length(y) == 0 {
		switch xv := x.value.(type) {
		case array:
			r := make([]bool, xv.Len())
			return NewAB(r)
		default:
			return NewI(b2i(false))
		}
	}
	if Length(x) == 0 {
		return NewAB(nil)
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
	default:
		return Panicf("x in y : y not an array (%s)", y.Type())
	}
}

func boolSliceMembers(xs []bool) (t bool, f bool) {
	for _, xi := range xs {
		if t && f {
			break
		}
		t, f = t || xi, f || !xi
	}
	return
}

func memberOfAB(x V, y *AB) V {
	t, f := boolSliceMembers(y.Slice)
	if t && f {
		switch xv := x.value.(type) {
		case *AB:
			r := make([]bool, xv.Len())
			for i := range r {
				r[i] = true
			}
			return NewAB(r)
		case *AS:
			r := make([]bool, xv.Len())
			return NewAB(r)
		case array:
			r := make([]bool, xv.Len())
			for i := 0; i < xv.Len(); i++ {
				xi := xv.at(i)
				r[i] = xi.IsI() && (xi.n == 0 || xi.n == 1) ||
					xi.IsF() && (xi.F() == 0 || xi.F() == 1)
			}
			return NewAB(r)
		default:
			b := x.IsI() && (x.n == 0 || x.n == 1) ||
				x.IsF() && (x.F() == 0 || x.F() == 1)
			return NewI(b2i(b))
		}
	}
	if t {
		switch xv := x.value.(type) {
		case *AB:
			r := make([]bool, xv.Len())
			copy(r, xv.Slice)
			return NewAB(r)
		case *AS:
			r := make([]bool, xv.Len())
			return NewAB(r)
		case array:
			r := make([]bool, xv.Len())
			for i := 0; i < xv.Len(); i++ {
				xi := xv.at(i)
				r[i] = xi.IsI() && xi.n == 1 ||
					xi.IsF() && xi.F() == 1
			}
			return NewAB(r)
		default:
			b := x.IsI() && x.n == 1 ||
				x.IsF() && x.F() == 1
			return NewI(b2i(b))
		}
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = !xi
		}
		return NewAB(r)
	case *AS:
		r := make([]bool, xv.Len())
		return NewAB(r)
	case array:
		r := make([]bool, xv.Len())
		for i := 0; i < xv.Len(); i++ {
			xi := xv.at(i)
			r[i] = xi.IsI() && xi.n == 0 ||
				xi.IsF() && xi.F() == 0
		}
		return NewAB(r)
	default:
		b := x.IsI() && x.n == 0 ||
			x.IsF() && x.F() == 0
		return NewI(b2i(b))
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

func memberOfSlice[T comparable](xs []T, ys []T, bruteForceThreshold int) []bool {
	r := make([]bool, len(xs))
	if len(xs) <= bruteForceThreshold || len(ys) <= bruteForceThreshold {
		for i, xi := range xs {
			for _, yi := range ys {
				if xi == yi {
					r[i] = true
					break
				}
			}
		}
		return r
	}
	m := bmapSlice[T](ys)
	for i, xi := range xs {
		_, r[i] = m[xi]
	}
	return r
}

func memberOfAF(x V, y *AF) V {
	if x.IsI() {
		for _, yi := range y.Slice {
			if float64(x.I()) == yi {
				return NewI(b2i(true))
			}
		}
		return NewI(b2i(false))
	}
	if x.IsF() {
		for _, yi := range y.Slice {
			if x.F() == yi {
				return NewI(b2i(true))
			}
		}
		return NewI(b2i(false))
	}
	switch xv := x.value.(type) {
	case *AB:
		return memberOfAF(fromABtoAF(xv), y)
	case *AI:
		return memberOfAF(toAF(xv), y)
	case *AF:
		return NewABWithRC(memberOfSlice[float64](xv.Slice, y.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberIOfAI(x int64, y *AI) bool {
	ylen := y.Len()
	i := sort.Search(ylen, func(j int) bool { return y.At(j) >= x })
	return i < ylen && y.At(i) == x
}

func memberOfAI(x V, y *AI) V {
	if x.IsI() {
		if y.flags.Has(flagAscending) && y.Len() > bruteForceNumeric/2 {
			return NewI(b2i(memberIOfAI(x.I(), y)))
		}
		for _, yi := range y.Slice {
			if x.I() == yi {
				return NewI(b2i(true))
			}
		}
		return NewI(b2i(false))
	}
	if x.IsF() {
		if !isI(x.F()) {
			return NewI(b2i(false))
		}
		if y.flags.Has(flagAscending) && y.Len() > bruteForceNumeric/2 {
			return NewI(b2i(memberIOfAI(int64(x.F()), y)))
		}
		for _, yi := range y.Slice {
			if x.F() == float64(yi) {
				return NewI(b2i(true))
			}
		}
		return NewI(b2i(false))
	}
	switch xv := x.value.(type) {
	case *AB:
		return memberOfAI(fromABtoAI(xv), y)
	case *AI:
		if y.flags.Has(flagAscending) && y.Len() > bruteForceNumeric/2 {
			r := make([]bool, xv.Len())
			for i, xi := range xv.Slice {
				r[i] = memberIOfAI(xi, y)
			}
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		return NewABWithRC(memberOfSlice[int64](xv.Slice, y.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AF:
		return memberOf(x, toAF(y))
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
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
			return NewI(b2i(memberSOfAS(string(xv), y)))
		}
		for _, yi := range y.Slice {
			if string(xv) == yi {
				return NewI(1)
			}
		}
		return NewI(0)
	case *AS:
		if y.flags.Has(flagAscending) && y.Len() > bruteForceGeneric/4 {
			r := make([]bool, xv.Len())
			for i, xi := range xv.Slice {
				r[i] = memberSOfAS(xi, y)
			}
			return NewABWithRC(r, reuseRCp(xv.rc))
		}
		return NewABWithRC(memberOfSlice[string](xv.Slice, y.Slice, bruteForceGeneric), reuseRCp(xv.rc))
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
		for _, yi := range y.Slice {
			if Match(x, yi) {
				return NewI(b2i(true))
			}
		}
		return NewI(b2i(false))
	}
}

func memberOfArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make([]bool, x.Len())
	for i := 0; i < x.Len(); i++ {
		for j := 0; j < y.Len(); j++ {
			if Match(x.at(i), y.at(j)) {
				r[i] = true
				break
			}
		}
	}
	return NewAB(r)
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
			r := make([]bool, xv.Len())
			return NewAB(r)
		}
	default:
		return panicType("ocount x", "x", x)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]int64, xv.Len())
		var f, t int64
		for i, xi := range xv.Slice {
			if xi {
				r[i] = t
				t++
				continue
			}
			r[i] = f
			f++
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AF:
		return NewAIWithRC(occurrenceCountSlice[float64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AI:
		return NewAIWithRC(occurrenceCountSlice[int64](xv.Slice, bruteForceNumeric), reuseRCp(xv.rc))
	case *AS:
		return NewAIWithRC(occurrenceCountSlice[string](xv.Slice, bruteForceGeneric), reuseRCp(xv.rc))
	case *AV:
		if xv.Len() > bruteForceGeneric {
			ss := make([]string, xv.Len())
			for i, xi := range xv.Slice {
				ss[i] = xi.Sprint(ctx)
			}
			return NewAIWithRC(occurrenceCountSlice[string](ss, bruteForceGeneric), reuseRCp(xv.rc))
		}
		r := make([]int64, xv.Len())
	loop:
		for i, xi := range xv.Slice {
			for j := i - 1; j >= 0; j-- {
				if Match(xi, xv.At(j)) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return NewAI(r)
	default:
		panic("ocount")
	}
}

func occurrenceCountSlice[T comparable](xs []T, bruteForceThreshold int) []int64 {
	r := make([]int64, len(xs))
	if len(xs) <= bruteForceThreshold {
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
		return trim(xv, y)
	case array:
		_, ok := y.value.(array)
		if !ok {
			d, ok := y.value.(*Dict)
			if ok {
				return withoutDict(x, d)
			}
			return panicType("x^y", "y", y)
		}
		r := memberOf(y, x)
		switch bres := r.value.(type) {
		case *AB:
			for i, b := range bres.Slice {
				bres.Slice[i] = !b
			}
		}
		return replicate(r, y)
	default:
		return panicType("x^y", "x", x)
	}
}

func withoutDict(x V, y *Dict) V {
	r := memberOf(NewV(y.values), x)
	switch bres := r.value.(type) {
	case *AB:
		for i, b := range bres.Slice {
			bres.Slice[i] = !b
		}
	}
	return NewDict(replicate(r, NewV(y.keys)), replicate(r, NewV(y.values)))
}

// intersection implements keep x#y.
func intersection(x, y V) V {
	switch yv := y.value.(type) {
	case array:
		return replicate(memberOf(y, x), y)
	case *Dict:
		r := memberOf(NewV(yv.values), x)
		return NewDict(replicate(r, NewV(yv.keys)), replicate(r, NewV(yv.values)))
	default:
		return panicType("x#y", "y", y)
	}
}

// find returns x?y.
func find(x, y V) V {
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
		for i, ss := range yv.Slice {
			r[i] = int64(strings.Index(string(s), string(ss)))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = findS(s, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return panicType("s?y", "y", y)
	}
}

func imapAB(x *AB) (m [2]int64) {
	m[0] = int64(x.Len())
	m[1] = int64(x.Len())
	if x.Len() == 0 {
		return m
	}
	m[b2i(x.At(0))] = 0
	for i, xi := range x.Slice[1:] {
		if xi != x.At(0) {
			m[b2i(xi)] = int64(i) + 1
			break
		}
	}
	return m
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
		for i, xi := range x.Slice {
			if b2i(xi) == y.I() {
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
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = m[b2i(yi)]
		}
		return NewAI(r)
	case *AI:
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi != 0 && yi != 1 {
				r[i] = int64(x.Len())
			} else {
				r[i] = m[yi]
			}
		}
		return NewAI(r)
	case *AF:
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi != 0 && yi != 1 {
				r[i] = int64(x.Len())
			} else {
				r[i] = m[int(yi)]
			}
		}
		return NewAI(r)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findSlices[T comparable](xs, ys []T, bruteForceThreshold int) []int64 {
	r := make([]int64, len(ys))
	xlen := int64(len(xs))
	if len(ys) <= bruteForceThreshold || len(xs) <= bruteForceThreshold {
		for i, yi := range ys {
			r[i] = xlen
			for j, xi := range xs {
				if yi == xi {
					r[i] = int64(j)
					break
				}
			}
		}
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

func findAFbools(xs []float64) (t int64, f int64) {
	xlen := int64(len(xs))
	t, f = xlen, xlen
	for i, xi := range xs {
		if xi == 1 && t == xlen {
			t = int64(i)
			if f < xlen {
				break
			}
		} else if xi == 0 && f == xlen {
			f = int64(i)
			if t < xlen {
				break
			}
		}
	}
	return
}

func findAF(x *AF, y V) V {
	if y.IsI() {
		for i, xi := range x.Slice {
			if xi == float64(y.I()) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	if y.IsF() {
		for i, xi := range x.Slice {
			if float64(xi) == y.F() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
	switch yv := y.value.(type) {
	case *AB:
		t, f := findAFbools(x.Slice)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi {
				r[i] = t
			} else {
				r[i] = f
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AI:
		return findAF(x, toAF(yv))
	case *AF:
		return NewAIWithRC(findSlices[float64](x.Slice, yv.Slice, bruteForceNumeric), reuseRCp(yv.rc))
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findAIbools(xs []int64) (t int64, f int64) {
	xlen := int64(len(xs))
	t, f = xlen, xlen
	for i, xi := range xs {
		if xi == 1 && t == xlen {
			t = int64(i)
			if f < xlen {
				break
			}
		} else if xi == 0 && f == xlen {
			f = int64(i)
			if t < xlen {
				break
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
		if x.flags.Has(flagAscending) && x.Len() > bruteForceNumeric/2 {
			return NewI(findAII(x, y.I()))
		}
		for i, xi := range x.Slice {
			if xi == y.I() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	if y.IsF() {
		if !isI(y.F()) {
			return NewI(int64(x.Len()))
		}
		yI := int64(y.F())
		if x.flags.Has(flagAscending) && x.Len() > bruteForceNumeric/2 {
			return NewI(findAII(x, yI))
		}
		for i, xi := range x.Slice {
			if xi == yI {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
	switch yv := y.value.(type) {
	case *AB:
		t, f := findAIbools(x.Slice)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi {
				r[i] = t
			} else {
				r[i] = f
			}
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AI:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceNumeric/2 {
			r := make([]int64, yv.Len())
			for i, yi := range yv.Slice {
				r[i] = findAII(x, yi)
			}
			return NewAI(r)
		}
		if yv.Len() > bruteForceNumeric && x.Len() > bruteForceNumeric {
			xlen := int64(x.Len())
			r := make([]int64, yv.Len())
			min, max := minMax(x)
			if max-min+1 < 2*(xlen+8) {
				// fast path avoiding hash table
				offset := -min
				m := make([]int64, max-min+1)
				for i, xi := range x.Slice {
					c := m[xi+offset]
					if c == 0 {
						m[xi+offset] = int64(i) + 1
						continue
					}
				}
				for i, yi := range yv.Slice {
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
				return NewAIWithRC(r, reuseRCp(yv.rc))
			}
		}
		return NewAIWithRC(findSlices[int64](x.Slice, yv.Slice, bruteForceNumeric), reuseRCp(yv.rc))
	case *AF:
		return findAF(toAF(x).value.(*AF), y)
	case array:
		return findArrays(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
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
		for i, xi := range x.Slice {
			if S(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AS:
		if x.flags.Has(flagAscending) && x.Len() > bruteForceGeneric/4 {
			r := make([]int64, yv.Len())
			for i, yi := range yv.Slice {
				r[i] = findASS(x, S(yi))
			}
			return NewAIWithRC(r, reuseRCp(yv.rc))
		}
		return NewAIWithRC(findSlices[string](x.Slice, yv.Slice, bruteForceGeneric), reuseRCp(yv.rc))
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
			if Match(y.at(i), x.at(j)) {
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
		for i, xi := range x.Slice {
			if Match(y, xi) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
}
