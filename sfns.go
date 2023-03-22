// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch xv := x.value.(type) {
	case countable:
		return xv.Len()
	default:
		return 1
	}
}

func reverseSlice[T any](xs []T) {
	for i := 0; i < len(xs)/2; i++ {
		xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
	}
}

func reverseMut(x V) {
	switch xv := x.value.(type) {
	case *AB:
		reverseSlice[bool](xv.Slice)
	case *AF:
		reverseSlice[float64](xv.Slice)
	case *AI:
		reverseSlice[int64](xv.Slice)
	case *AS:
		reverseSlice[string](xv.Slice)
	case *AV:
		reverseSlice[V](xv.Slice)
	}
}

// reverse returns |x.
func reverse(x V) V {
	switch xv := x.value.(type) {
	case array:
		x.value = xv.shallowClone()
		reverseMut(x)
		return x
	case *Dict:
		k := reverse(NewV(xv.keys))
		k.InitRC()
		v := reverse(NewV(xv.values))
		v.InitRC()
		return NewV(&Dict{keys: k.value.(array), values: v.value.(array)})
	default:
		return panicType("|x", "x", x)
	}
}

// Rotate returns x rotate y.
func rotate(x, y V) V {
	i := int64(0)
	if x.IsI() {
		i = x.I()
	} else if x.IsF() {
		if !isI(x.F()) {
			return Panicf("x rotate y : non-integer f[y] (%g)", x.F())
		}
		i = int64(x.F())
	} else {
		return Panicf("x rotate y : non-integer f[y] (%s)", x.Type())
	}
	ylen := int64(Length(y))
	if ylen == 0 {
		return y
	}
	i %= ylen
	if i < 0 {
		i += ylen
	}
	return rotateI(i, y)
}

func rotateSlice[T any](i int64, ys []T) []T {
	ylen := int64(len(ys))
	r := make([]T, ylen)
	for j := int64(0); j < ylen; j++ {
		r[j] = ys[int((j+i)%ylen)]
	}
	return r
}

func rotateI(i int64, y V) V {
	switch yv := y.value.(type) {
	case *AB:
		return NewABWithRC(rotateSlice[bool](i, yv.Slice), reuseRCp(yv.rc))
	case *AF:
		return NewAFWithRC(rotateSlice[float64](i, yv.Slice), reuseRCp(yv.rc))
	case *AI:
		return NewAIWithRC(rotateSlice[int64](i, yv.Slice), reuseRCp(yv.rc))
	case *AS:
		return NewASWithRC(rotateSlice[string](i, yv.Slice), reuseRCp(yv.rc))
	case *AV:
		return NewAVWithRC(rotateSlice[V](i, yv.Slice), yv.rc)
	case *Dict:
		k := rotateI(i, NewV(yv.keys))
		if k.IsPanic() {
			return k
		}
		k.InitRC()
		v := rotateI(i, NewV(yv.values))
		if v.IsPanic() {
			return v
		}
		v.InitRC()
		return NewV(&Dict{keys: k.value.(array), values: v.value.(array)})
	default:
		return panicType("x rotate y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch xv := x.value.(type) {
	case array:
		if xv.Len() == 0 {
			switch xv.(type) {
			case *AS:
				return NewS("")
			default:
				return NewI(0)
			}
		}
		return xv.at(0)
	case *Dict:
		return first(NewV(xv.values))
	default:
		return x
	}
}

// drop implements i_x, s_x, I_x and x_i.
func drop(x, y V) V {
	if x.IsI() {
		return dropN(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i_y : non-integer i (%g)", x.F())
		}
		return dropN(int64(x.F()), y)
	}
	if y.IsI() {
		return deleteI(x, y.I())
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("x_i : non-integer i (%g)", y.F())
		}
		return deleteI(x, int64(y.F()))
	}
	switch xv := x.value.(type) {
	case S:
		return dropS(xv, y)
	case *AB:
		return drop(fromABtoAI(xv), y)
	case *AI:
		switch yv := y.value.(type) {
		case S:
			return cutAIS(xv, yv)
		case array:
			return cutAIarray(xv, yv)
		default:
			return panicType("I_y", "y", y)
		}
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return drop(z, y)
	case array:
		return panics("x_y : x non-integer array")
	default:
		return panicType("x_y", "x", x)
	}
}

func dropN(n int64, y V) V {
	switch yv := y.value.(type) {
	case S:
		switch {
		case n >= 0:
			if n > int64(len(yv)) {
				n = int64(len(yv))
			}
			return NewV(yv[n:])
		default:
			n = int64(len(yv)) + n
			if n < 0 {
				n = 0
			}
			return NewV(yv[:n])
		}
	case array:
		switch {
		case n >= 0:
			if n > int64(yv.Len()) {
				n = int64(yv.Len())
			}
			y.value = yv.slice(int(n), yv.Len())
			return Canonical(y)
		default:
			n = int64(yv.Len()) + n
			if n < 0 {
				n = 0
			}
			y.value = yv.slice(0, int(n))
			return Canonical(y)
		}
	case *Dict:
		rk := dropN(n, NewV(yv.keys))
		rk.InitRC()
		rv := dropN(n, NewV(yv.values))
		rv.InitRC()
		return NewV(&Dict{
			keys:   rk.value.(array),
			values: rv.value.(array)})
	default:
		return panicType("i_y", "y", y)
	}
}

func deleteiSlice[T any](xs []T, i int64) []T {
	r := make([]T, i, len(xs)-1)
	copy(r, xs[:i])
	r = append(r, xs[i+1:]...)
	return r
}

func deleteI(x V, i int64) V {
	switch xv := x.value.(type) {
	case S:
		if i < 0 {
			i += int64(len(xv))
		}
		if i >= int64(len(xv)) || i < 0 {
			return x
		}
		return NewV(xv[:i] + xv[i+1:])
	case *AB:
		if i < 0 {
			i += int64(xv.Len())
		}
		if i >= int64(xv.Len()) || i < 0 {
			return x
		}
		if reusableRCp(xv.rc) {
			xv.Slice = append(xv.Slice[:i], xv.Slice[i+1:]...)
			return x
		}
		r := deleteiSlice[bool](xv.Slice, i)
		return NewV(&AB{Slice: r, flags: xv.flags})
	case *AI:
		if i < 0 {
			i += int64(xv.Len())
		}
		if i >= int64(xv.Len()) || i < 0 {
			return x
		}
		if reusableRCp(xv.rc) {
			xv.Slice = append(xv.Slice[:i], xv.Slice[i+1:]...)
			return x
		}
		r := deleteiSlice[int64](xv.Slice, i)
		return NewV(&AI{Slice: r, flags: xv.flags})
	case *AF:
		if i < 0 {
			i += int64(xv.Len())
		}
		if i >= int64(xv.Len()) || i < 0 {
			return x
		}
		if reusableRCp(xv.rc) {
			xv.Slice = append(xv.Slice[:i], xv.Slice[i+1:]...)
			return x
		}
		r := deleteiSlice[float64](xv.Slice, i)
		return NewV(&AF{Slice: r, flags: xv.flags})
	case *AS:
		if i < 0 {
			i += int64(xv.Len())
		}
		if i >= int64(xv.Len()) || i < 0 {
			return x
		}
		if reusableRCp(xv.rc) {
			xv.Slice = append(xv.Slice[:i], xv.Slice[i+1:]...)
			return x
		}
		r := deleteiSlice[string](xv.Slice, i)
		return NewV(&AS{Slice: r, flags: xv.flags})
	case *AV:
		if i < 0 {
			i += int64(xv.Len())
		}
		if i >= int64(xv.Len()) || i < 0 {
			return x
		}
		if reusableRCp(xv.rc) {
			xv.Slice = append(xv.Slice[:i], xv.Slice[i+1:]...)
			return canonicalFast(x)
		}
		r := deleteiSlice[V](xv.Slice, i)
		return NewV(canonicalAV(&AV{Slice: r, flags: xv.flags, rc: xv.rc}))
	case *Dict:
		rk := deleteI(NewV(xv.keys), i)
		rk.InitRC()
		rv := deleteI(NewV(xv.values), i)
		rv.InitRC()
		return NewV(&Dict{
			keys:   rk.value.(array),
			values: rv.value.(array)})
	default:
		return panicType("x_i", "x", x)
	}
}

func cutAIarray(x *AI, y array) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_y : I is not ascending")
	}
	x.flags |= flagAscending
	ylen := int64(y.Len())
	for _, i := range x.Slice {
		if i < 0 || i > ylen {
			return Panicf("I_y : I contains out of bound index (%d)", i)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewAVWithRC(nil, reuseRCp(x.rc))
	}
	r := make([]V, xlen)
	rc := y.RC()
	*rc += 2
	for i, from := range x.Slice {
		to := ylen
		if i+1 < xlen {
			to = x.At(i + 1)
		}
		r[i] = Canonical(NewV(y.slice(int(from), int(to))))
	}
	return NewAVWithRC(r, reuseRCp(x.rc))
}

func cutAIS(x *AI, y S) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_s : I is not ascending")
	}
	x.flags |= flagAscending
	ylen := int64(len(y))
	for _, i := range x.Slice {
		if i < 0 || i > ylen {
			return Panicf("I_s : I contains out of bound index (%d)", i)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewASWithRC(nil, reuseRCp(x.rc))
	}
	r := make([]string, xlen)
	for i, from := range x.Slice {
		to := ylen
		if i+1 < xlen {
			to = x.At(i + 1)
		}
		r[i] = string(y[from:to])
	}
	return NewASWithRC(r, reuseRCp(x.rc))
}

// take returns i#y.
func take(x, y V) V {
	n := int64(0)
	if x.IsI() {
		n = x.I()
	} else if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i#y : non-integer i (%g)", x.F())
		}
		n = int64(x.F())
	} else {
		switch xv := x.value.(type) {
		case S:
			return scount(xv, y)
		case array:
			return intersection(x, y)
		default:
			return panicType("x#y", "x", x)
		}
	}
	switch yv := y.value.(type) {
	case *Dict:
		rk := takeN(n, yv.keys)
		rk.InitRC()
		rv := takeN(n, yv.values)
		rv.InitRC()
		return NewV(&Dict{
			keys:   rk.value.(array),
			values: rv.value.(array)})
	case array:
		return takeN(n, yv)
	default:
		return takeN(n, toArray(y).value.(array))
	}
}

func takeN(n int64, y array) V {
	if y.Len() == 0 {
		if n < 0 {
			n = -n
		}
		r := make([]bool, n)
		return NewABWithRC(r, reuseRCp(y.RC()))
	}
	switch {
	case n >= 0:
		if n > int64(y.Len()) {
			return takeCyclic(n, y)
		}
		return Canonical(NewV(y.slice(0, int(n))))
	default:
		if n < int64(-y.Len()) {
			return takeCyclic(n, y)
		}
		return Canonical(NewV(y.slice(y.Len()+int(n), y.Len())))
	}
}

func takeCyclicSlice[T any](n int64, ys []T) []T {
	neg := n < 0
	if neg {
		n = -n
	}
	i := int64(0)
	step := int64(len(ys))
	r := make([]T, n)
	if neg {
		res := n % step
		if res > 0 {
			copy(r[0:res], ys[step-res:])
			i += res
		}
	}
	for i+step <= n {
		copy(r[i:i+step], ys)
		i += step
	}
	if !neg && i < n {
		copy(r[i:n], ys[:n-i])
	}
	return r
}

func takeCyclic(n int64, y array) V {
	switch yv := y.(type) {
	case *AB:
		return NewABWithRC(takeCyclicSlice[bool](n, yv.Slice), reuseRCp(yv.rc))
	case *AI:
		return NewAIWithRC(takeCyclicSlice[int64](n, yv.Slice), reuseRCp(yv.rc))
	case *AF:
		return NewAFWithRC(takeCyclicSlice[float64](n, yv.Slice), reuseRCp(yv.rc))
	case *AS:
		return NewASWithRC(takeCyclicSlice[string](n, yv.Slice), reuseRCp(yv.rc))
	case *AV:
		*yv.rc += 2
		return NewAVWithRC(takeCyclicSlice[V](n, yv.Slice), yv.rc)
	default:
		panic("takeCyclic: y not an array")
	}
}

// ShiftBefore returns x rshift y.
func shiftBefore(x, y V) V {
	switch yv := y.value.(type) {
	case *AB:
		return shiftBeforeAB(x, yv)
	case *AI:
		return shiftBeforeAI(x, yv)
	case *AF:
		return shiftBeforeAF(x, yv)
	case *AS:
		return shiftBeforeAS(x, yv)
	case *AV:
		return shiftBeforeAV(x, yv)
	case *Dict:
		return newDictValues(yv.keys, shiftBefore(x, NewV(yv.values)))
	default:
		return panicType("x rshift y", "y", y)
	}
}

func shiftBeforeAB(x V, yv *AB) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		if isBI(x.I()) {
			r := yv.reuse()
			copy(r.Slice[max:], ys[:len(ys)-max])
			r.Slice[0] = x.I() == 1
			return NewV(r)
		}
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = b2i(yv.At(i - max))
		}
		r[0] = x.I()
		return NewAIWithRC(r, reuseRCp(yv.rc))
	}
	if x.IsF() {
		if isBF(x.F()) {
			r := yv.reuse()
			copy(r.Slice[max:], ys[:len(ys)-max])
			r.Slice[0] = x.F() == 1
			return NewV(r)
		}
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = b2f(yv.At(i - max))
		}
		r[0] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		copy(r.Slice[:max], xv.Slice)
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = b2f(ys[i-max])
		}
		copy(r[:max], xv.Slice)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = b2i(yv.At(i - max))
		}
		copy(r[:max], xv.Slice)
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AV:
		return shiftAVBeforeArray(xv, yv)
	case array:
		return shiftArrayBeforeArray(xv, yv)
	default:
		return shiftVBeforeArray(x, yv)
	}
}

func shiftBeforeAI(x V, yv *AI) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		r.Slice[0] = x.I()
		return NewV(r)
	} else if x.IsF() {
		if isI(x.F()) {
			r := yv.reuse()
			copy(r.Slice[max:], ys[:len(ys)-max])
			r.Slice[0] = int64(x.F())
			return NewV(r)
		}
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = float64(yv.At(i - max))
		}
		r[0] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = b2i(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = float64(ys[i-max])
		}
		copy(r[:max], xv.Slice)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = xv.At(i)
		}
		return NewV(r)
	case *AV:
		return shiftAVBeforeArray(xv, yv)
	case array:
		return shiftArrayBeforeArray(xv, yv)
	default:
		return shiftVBeforeArray(x, yv)
	}
}

func shiftBeforeAF(x V, yv *AF) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		r.Slice[0] = float64(x.I())
		return NewV(r)
	} else if x.IsF() {
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		r.Slice[0] = x.F()
		return NewV(r)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = float64(b2f(xv.At(i)))
		}
		return NewV(r)
	case *AF:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = xv.At(i)
		}
		return NewV(r)
	case *AI:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = float64(xv.At(i))
		}
		return NewV(r)
	case *AV:
		return shiftAVBeforeArray(xv, yv)
	case array:
		return shiftArrayBeforeArray(xv, yv)
	default:
		return shiftVBeforeArray(x, yv)
	}
}

func shiftBeforeAS(x V, yv *AS) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	switch xv := x.value.(type) {
	case S:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		r.Slice[0] = string(xv)
		return NewV(r)
	case *AS:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = xv.At(i)
		}
		return NewV(r)
	case *AV:
		return shiftAVBeforeArray(xv, yv)
	case array:
		return shiftArrayBeforeArray(xv, yv)
	default:
		return shiftVBeforeArray(x, yv)
	}
}

func shiftBeforeAV(x V, yv *AV) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	switch xv := x.value.(type) {
	case array:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.Slice[i] = xv.at(i)
		}
		return Canonical(NewV(r))
	default:
		r := yv.reuse()
		copy(r.Slice[max:], ys[:len(ys)-max])
		r.Slice[0] = x
		return NewV(r)
	}
}

func shiftArrayBeforeArray(xv, yv array) V {
	ylen := yv.Len()
	max := minInt(xv.Len(), ylen)
	r := make([]V, ylen)
	rc := yv.RC()
	for i := max; i < ylen; i++ {
		r[i] = yv.at(i - max)
	}
	for i := 0; i < max; i++ {
		xi := xv.at(i)
		xi.InitWithRC(rc)
		r[i] = xi
	}
	return NewAVWithRC(r, rc)
}

func shiftVBeforeArray(x V, yv array) V {
	ylen := yv.Len()
	r := make([]V, ylen)
	for i := 1; i < ylen; i++ {
		r[i] = yv.at(i - 1)
	}
	r[0] = x
	return NewAV(r)
}

func shiftAVBeforeArray(xv *AV, yv array) V {
	ylen := yv.Len()
	max := minInt(xv.Len(), ylen)
	r := make([]V, ylen)
	for i := max; i < ylen; i++ {
		r[i] = yv.at(i - max)
	}
	copy(r[:max], xv.Slice)
	return NewAV(r)
}

// nudge returns rshift x.
func nudge(x V) V {
	if Length(x) == 0 {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		r := xv.reuse()
		copy(r.Slice[1:], xv.Slice[:xv.Len()-1])
		r.Slice[0] = false
		return NewV(r)
	case *AI:
		r := xv.reuse()
		copy(r.Slice[1:], xv.Slice[:xv.Len()-1])
		r.Slice[0] = 0
		return NewV(r)
	case *AF:
		r := xv.reuse()
		copy(r.Slice[1:], xv.Slice[:xv.Len()-1])
		r.Slice[0] = 0
		return NewV(r)
	case *AS:
		r := xv.reuse()
		copy(r.Slice[1:], xv.Slice[:xv.Len()-1])
		r.Slice[0] = ""
		return NewV(r)
	case *AV:
		r := xv.reuse()
		copy(r.Slice[1:], xv.Slice[:xv.Len()-1])
		r.Slice[0] = NewI(0)
		return Canonical(NewV(r))
	case *Dict:
		return newDictValues(xv.keys, nudge(NewV(xv.values)))
	default:
		return panicType("rshift x", "x", x)
	}
}

// ShiftAfter returns x shift y.
func shiftAfter(x, y V) V {
	switch yv := y.value.(type) {
	case *AB:
		return shiftAfterAB(x, yv)
	case *AI:
		return shiftAfterAI(x, yv)
	case *AF:
		return shiftAfterAF(x, yv)
	case *AS:
		return shiftAfterAS(x, yv)
	case *AV:
		return shiftAfterAV(x, yv)
	case *Dict:
		return newDictValues(yv.keys, shiftAfter(x, NewV(yv.values)))
	default:
		return panicType("x shift y", "y", y)
	}
}

func shiftAfterAB(x V, yv *AB) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		if isBI(x.I()) {
			r := yv.reuse()
			copy(r.Slice[:len(ys)-max], ys[max:])
			r.Slice[len(ys)-1] = x.I() == 1
			return NewV(r)
		}
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = b2i(yv.At(i))
		}
		r[len(ys)-1] = x.I()
		return NewAIWithRC(r, reuseRCp(yv.rc))
	} else if x.IsF() {
		if isBF(x.F()) {
			r := yv.reuse()
			copy(r.Slice[:len(ys)-max], ys[max:])
			r.Slice[len(ys)-1] = x.F() == 1
			return NewV(r)
		}
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = b2f(yv.At(i))
		}
		r[len(ys)-1] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		copy(r.Slice[len(ys)-max:], xv.Slice)
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = b2f(ys[i])
		}
		copy(r[len(ys)-max:], xv.Slice)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = b2i(yv.At(i))
		}
		copy(r[len(ys)-max:], xv.Slice)
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AV:
		return shiftAVAfterArray(xv, yv)
	case array:
		return shiftArrayAfterArray(xv, yv)
	default:
		return shiftVAfterArray(x, yv)
	}
}

func shiftAfterAI(x V, yv *AI) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		r.Slice[len(ys)-1] = x.I()
		return NewV(r)
	} else if x.IsF() {
		if isI(x.F()) {
			r := yv.reuse()
			copy(r.Slice[:len(ys)-max], ys[max:])
			r.Slice[len(ys)-1] = int64(x.F())
			return NewV(r)
		}
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = float64(yv.At(i))
		}
		r[len(ys)-1] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = b2i(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = float64(ys[i])
		}
		copy(r[len(ys)-max:], xv.Slice)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = xv.At(i)
		}
		return NewV(r)
	case *AV:
		return shiftAVAfterArray(xv, yv)
	case array:
		return shiftArrayAfterArray(xv, yv)
	default:
		return shiftVAfterArray(x, yv)
	}
}

func shiftAfterAF(x V, yv *AF) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	if x.IsI() {
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		r.Slice[len(ys)-1] = float64(x.I())
		return NewV(r)
	} else if x.IsF() {
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		r.Slice[len(ys)-1] = x.F()
		return NewV(r)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = float64(b2f(xv.At(i)))
		}
		return NewV(r)
	case *AF:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = xv.At(i)
		}
		return NewV(r)
	case *AI:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = float64(xv.At(i))
		}
		return NewV(r)
	case *AV:
		return shiftAVAfterArray(xv, yv)
	case array:
		return shiftArrayAfterArray(xv, yv)
	default:
		return shiftVAfterArray(x, yv)
	}
}

func shiftAfterAS(x V, yv *AS) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	switch xv := x.value.(type) {
	case S:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		r.Slice[len(ys)-1] = string(xv)
		return NewV(r)
	case *AS:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = xv.At(i)
		}
		return NewV(r)
	case *AV:
		return shiftAVAfterArray(xv, yv)
	case array:
		return shiftArrayAfterArray(xv, yv)
	default:
		return shiftVAfterArray(x, yv)
	}
}

func shiftAfterAV(x V, yv *AV) V {
	max := minInt(Length(x), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.Slice
	switch xv := x.value.(type) {
	case array:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.Slice[len(ys)-max+i] = xv.at(i)
		}
		return Canonical(NewV(r))
	default:
		r := yv.reuse()
		copy(r.Slice[:len(ys)-max], ys[max:])
		r.Slice[len(ys)-1] = x
		return NewV(r)
	}
}

func shiftArrayAfterArray(xv, yv array) V {
	ylen := yv.Len()
	max := minInt(xv.Len(), ylen)
	r := make([]V, ylen)
	rc := yv.RC()
	for i := max; i < ylen; i++ {
		r[i-max] = yv.at(i)
	}
	for i := 0; i < max; i++ {
		xi := xv.at(i)
		xi.InitWithRC(rc)
		r[ylen-max+i] = xi
	}
	return NewAVWithRC(r, rc)
}

func shiftVAfterArray(x V, yv array) V {
	ylen := yv.Len()
	r := make([]V, ylen)
	for i := 1; i < ylen; i++ {
		r[i-1] = yv.at(i)
	}
	r[ylen-1] = x
	return NewAV(r)
}

func shiftAVAfterArray(xv *AV, yv array) V {
	ylen := yv.Len()
	max := minInt(xv.Len(), ylen)
	r := make([]V, ylen)
	for i := max; i < ylen; i++ {
		r[i-max] = yv.at(i)
	}
	copy(r[ylen-max:], xv.Slice)
	return NewAV(r)
}

// NudgeBack returns shift x.
func nudgeBack(x V) V {
	if Length(x) == 0 {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		r := xv.reuse()
		copy(r.Slice[0:xv.Len()-1], xv.Slice[1:])
		r.Slice[xv.Len()-1] = false
		return NewV(r)
	case *AI:
		r := xv.reuse()
		copy(r.Slice[0:xv.Len()-1], xv.Slice[1:])
		r.Slice[xv.Len()-1] = 0
		return NewV(r)
	case *AF:
		r := xv.reuse()
		copy(r.Slice[0:xv.Len()-1], xv.Slice[1:])
		r.Slice[xv.Len()-1] = 0
		return NewV(r)
	case *AS:
		r := xv.reuse()
		copy(r.Slice[0:xv.Len()-1], xv.Slice[1:])
		r.Slice[xv.Len()-1] = ""
		return NewV(r)
	case *AV:
		r := xv.reuse()
		copy(r.Slice[0:xv.Len()-1], xv.Slice[1:])
		r.Slice[xv.Len()-1] = NewI(0)
		return Canonical(NewV(r))
	case *Dict:
		return newDictValues(xv.keys, nudgeBack(NewV(xv.values)))
	default:
		return panicType("shift x", "x", x)
	}
}

// windows returns i^y.
func windows(i int64, y V) V {
	switch yv := y.value.(type) {
	case S:
		if i <= 0 || i >= int64(len(yv)+1) {
			return Panicf("i^y : i out of range !%d (%d)", len(yv)+1, i)
		}
		r := make([]string, 1+len(yv)-int(i))
		for j := range r {
			r[j] = string(yv[j : j+int(i)])
		}
		return NewAS(r)
	case array:
		if i <= 0 || i >= int64(yv.Len()+1) {
			return Panicf("i^y : i out of range !%d (%d)", yv.Len()+1, i)
		}
		r := make([]V, 1+yv.Len()-int(i))
		rc := yv.RC()
		*rc += 2
		for j := range r {
			yc := y
			yc.value = yv.slice(j, j+int(i))
			r[j] = Canonical(yc)
		}
		var n int
		return NewAVWithRC(r, &n)
	default:
		return panicType("i^y", "y", y)
	}
}

// shapeSplit returns i!y.
func shapeSplit(x V, y V) V {
	var i int64
	if x.IsI() {
		i = x.I()
	} else {
		// x.IsF() should be true
		f := x.F()
		if !isI(f) {
			return Panicf("i!y : i non-integer (%g)", f)
		}
		i = int64(f)
	}
	switch yv := y.value.(type) {
	case S:
		ylen := len(yv)
		if i <= 0 {
			return Panicf("i!s : i not positive (%d)", i)
		}
		if i >= int64(ylen) {
			return NewAS([]string{string(yv)})
		}
		n := ylen / int(i)
		if ylen%int(i) != 0 {
			n++
		}
		r := make([]string, n)
		for j := 0; j < n; j++ {
			from := j * int(i)
			to := minInt(from+int(i), ylen)
			r[j] = string(yv[from:to])
		}
		return NewAS(r)
	case array:
		ylen := yv.Len()
		if i <= 0 {
			return Panicf("i!y : i not positive (%d)", i)
		}
		if i >= int64(ylen) {
			return NewAVWithRC([]V{y}, yv.RC())
		}
		n := ylen / int(i)
		if ylen%int(i) != 0 {
			n++
		}
		rc := yv.RC()
		*rc += 2
		r := make([]V, n)
		for j := 0; j < n; j++ {
			yc := y
			from := j * int(i)
			to := minInt(from+int(i), ylen)
			yc.value = yv.slice(from, to)
			r[j] = Canonical(yc)
		}
		var rcn int
		return NewAVWithRC(r, &rcn)
	default:
		return panicType("i!y", "y", y)
	}
}
