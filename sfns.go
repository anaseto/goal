// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// Len returns the length of a value like in #x.
func (x V) Len() int {
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

func reverseMut(x array) {
	switch xv := x.(type) {
	case *AB:
		reverseSlice[byte](xv.elts)
	case *AF:
		reverseSlice[float64](xv.elts)
	case *AI:
		reverseSlice[int64](xv.elts)
	case *AS:
		reverseSlice[string](xv.elts)
	case *AV:
		reverseSlice[V](xv.elts)
	}
}

// reverse returns |x.
func reverse(x V) V {
	switch xv := x.value.(type) {
	case array:
		flags := xv.getFlags() &^ flagAscending
		xv = xv.shallowClone()
		reverseMut(xv)
		xv.setFlags(flags)
		x.value = xv
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
	if x.IsI() {
		return rotateI(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("x rotate y : non-integer x (%g)", x.F())
		}
		return rotateI(int64(x.F()), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := rotateI(int64(xi), y)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *AI:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := rotateI(xi, y)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *AF:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			if !isI(xi) {
				return Panicf("x rotate y : non-integer x (%g)", x.F())
			}
			ri := rotateI(int64(xi), y)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := rotate(xi, y)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	default:
		return panicType("x rotate y", "x", x)
	}
}

func rotateI(i int64, y V) V {
	ylen := int64(y.Len())
	if ylen == 0 {
		return y
	}
	i %= ylen
	if i < 0 {
		i += ylen
	}
	switch yv := y.value.(type) {
	case *AB:
		fl := yv.flags &^ flagAscending
		return NewV(&AB{elts: rotateSlice[byte](i, yv.elts), rc: reuseRCp(yv.rc), flags: fl})
	case *AF:
		fl := yv.flags &^ flagAscending
		return NewV(&AF{elts: rotateSlice[float64](i, yv.elts), rc: reuseRCp(yv.rc), flags: fl})
	case *AI:
		fl := yv.flags &^ flagAscending
		return NewV(&AI{elts: rotateSlice[int64](i, yv.elts), rc: reuseRCp(yv.rc), flags: fl})
	case *AS:
		fl := yv.flags &^ flagAscending
		return NewV(&AS{elts: rotateSlice[string](i, yv.elts), rc: reuseRCp(yv.rc), flags: fl})
	case *AV:
		fl := yv.flags &^ flagAscending
		return NewV(&AV{elts: rotateSlice[V](i, yv.elts), rc: yv.rc, flags: fl})
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

func rotateSlice[T any](i int64, ys []T) []T {
	ylen := int64(len(ys))
	r := make([]T, ylen)
	for j := int64(0); j < ylen; j++ {
		r[j] = ys[int((j+i)%ylen)]
	}
	return r
}

// first returns *x.
func first(x V) V {
	switch xv := x.value.(type) {
	case array:
		if xv.Len() == 0 {
			switch xv.(type) {
			case *AF:
				return NewF(0)
			case *AS:
				return NewS("")
			case *AV:
				return NewAV(nil)
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
	switch xv := x.value.(type) {
	case S:
		return dropS(xv, y)
	case *AB:
		switch yv := y.value.(type) {
		case S:
			return cutABS(xv, yv)
		case array:
			return cutABarray(xv, yv)
		case *Dict:
			k := cutABarray(xv, yv.keys)
			if k.IsPanic() {
				return k
			}
			v := cutABarray(xv, yv.values)
			if v.IsPanic() {
				return v
			}
			return NewDict(k, v)
		default:
			return panicType("I_y", "y", y)
		}
	case *AI:
		switch yv := y.value.(type) {
		case S:
			return cutAIS(xv, yv)
		case array:
			return cutAIarray(xv, yv)
		case *Dict:
			k := cutAIarray(xv, yv.keys)
			if k.IsPanic() {
				return k
			}
			v := cutAIarray(xv, yv.values)
			if v.IsPanic() {
				return v
			}
			return NewDict(k, v)
		default:
			return panicType("I_y", "y", y)
		}
	case *AF:
		x = toAI(xv)
		if x.IsPanic() {
			return x
		}
		return drop(x, y)
	case *AS:
		return dropAS(xv, y)
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

func cutAIarray(x *AI, y array) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_y : non-ascending I")
	}
	x.flags |= flagAscending
	ylen := int64(y.Len())
	for _, i := range x.elts {
		if i < 0 || i > ylen {
			return Panicf("I_y : I contains out of bounds index (%d)", i)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewAVWithRC(nil, reuseRCp(x.rc))
	}
	r := cutIntsArray(x.elts, y)
	return NewAV(r)
}

func cutABarray(x *AB, y array) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_y : non-ascending I")
	}
	x.flags |= flagAscending
	ylen := int(y.Len())
	for _, xi := range x.elts {
		if int(xi) > ylen {
			return Panicf("I_y : I contains out of bounds index (%d)", xi)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewAV(nil)
	}
	r := cutIntsArray(x.elts, y)
	return NewAV(r)
}

func cutIntsArray[I integer](x []I, y array) []V {
	xlen := len(x)
	ylen := int64(y.Len())
	r := make([]V, xlen)
	rc := y.RC()
	*rc += 2
	for i, from := range x {
		to := ylen
		if i+1 < xlen {
			to = int64(x[i+1])
		}
		r[i] = Canonical(NewV(y.slice(int(from), int(to))))
	}
	return r
}

func cutAIS(x *AI, y S) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_s : non-ascending I")
	}
	x.flags |= flagAscending
	ylen := int64(len(y))
	for _, i := range x.elts {
		if i < 0 || i > ylen {
			return Panicf("I_s : I contains out of bounds index (%d)", i)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewASWithRC(nil, reuseRCp(x.rc))
	}
	r := cutIntsS(x.elts, y)
	return NewASWithRC(r, reuseRCp(x.rc))
}

func cutABS(x *AB, y S) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(x) {
		return panics("I_s : non-ascending I")
	}
	x.flags |= flagAscending
	ylen := int64(len(y))
	for _, i := range x.elts {
		if int64(i) > ylen {
			return Panicf("I_s : I contains out of bounds index (%d)", i)
		}
	}
	xlen := x.Len()
	if xlen == 0 {
		return NewASWithRC(nil, reuseRCp(x.rc))
	}
	r := cutIntsS(x.elts, y)
	return NewASWithRC(r, reuseRCp(x.rc))
}

func cutIntsS[I integer](x []I, y S) []string {
	xlen := len(x)
	ylen := int64(len(y))
	r := make([]string, xlen)
	for i, from := range x {
		to := ylen
		if i+1 < xlen {
			to = int64(x[i+1])
		}
		r[i] = string(y[from:to])
	}
	return r
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
			return intersection(xv, y)
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
		return takeNAtom(n, y)
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

func takeNAtom(n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		if isBI(yv) {
			if isbI(yv) {
				r := make([]byte, n)
				if yv != 0 {
					for i := range r {
						r[i] = 1
					}
				}
				return newABb(r)
			}
			r := constArray(n, byte(yv))
			return NewAB(r)
		}
		r := constArray(n, yv)
		return NewAI(r)
	}
	if y.IsF() {
		yv := y.F()
		r := constArray(n, yv)
		return NewAF(r)
	}
	switch yv := y.value.(type) {
	case S:
		r := constArray(n, string(yv))
		return NewAS(r)
	default:
		y.immutable()
		r := constArray(n, y)
		var n int
		rv := &AV{elts: r, rc: &n}
		return NewV(rv)
	}
}

func constArray[T any](n int64, y T) []T {
	r := make([]T, n)
	for i := range r {
		r[i] = y
	}
	return r
}

func takeN(n int64, y array) V {
	if y.Len() == 0 {
		if n < 0 {
			n = -n
		}
		switch y.(type) {
		case *AS:
			r := make([]string, n)
			return NewAS(r)
		case *AF:
			r := make([]float64, n)
			return NewAF(r)
		case *AV:
			r := make([]V, n)
			for i := range r {
				r[i] = NewAV(nil)
			}
			return NewAV(r)
		default:
			r := make([]byte, n)
			return newABb(r)
		}
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

func takeCyclic(n int64, y array) V {
	switch yv := y.(type) {
	case *AB:
		fl := yv.flags & flagBool
		r := takeCyclicSlice[byte](n, yv.elts)
		return NewV(&AB{elts: r, rc: reuseRCp(yv.rc), flags: fl})
	case *AI:
		return NewAIWithRC(takeCyclicSlice[int64](n, yv.elts), reuseRCp(yv.rc))
	case *AF:
		return NewAFWithRC(takeCyclicSlice[float64](n, yv.elts), reuseRCp(yv.rc))
	case *AS:
		return NewASWithRC(takeCyclicSlice[string](n, yv.elts), reuseRCp(yv.rc))
	case *AV:
		*yv.rc += 2
		return NewAVWithRC(takeCyclicSlice[V](n, yv.elts), yv.rc)
	default:
		panic("takeCyclic: y not an array")
	}
}

func takePadN(n int64, x array) V {
	switch {
	case n >= 0:
		if n > int64(x.Len()) {
			return padArrayN(n, x)
		}
		return Canonical(NewV(x.slice(0, int(n))))
	default:
		if n < int64(-x.Len()) {
			return padArrayN(n, x)
		}
		return Canonical(NewV(x.slice(x.Len()+int(n), x.Len())))
	}
}

func padArrayN(n int64, x array) V {
	switch xv := x.(type) {
	case *AB:
		r := padNSlice[byte](n, xv.elts)
		if xv.IsBoolean() {
			return newABb(r)
		}
		return NewAB(r)
	case *AI:
		return NewAI(padNSlice[int64](n, xv.elts))
	case *AF:
		return NewAF(padNSlice[float64](n, xv.elts))
	case *AS:
		return NewAS(padNSlice[string](n, xv.elts))
	case *AV:
		return NewAV(padNSliceVs(n, xv.elts))
	default:
		panic("padArrayN")
	}
}

func padNSlice[T any](n int64, ys []T) []T {
	l := n
	if l < 0 {
		l = -l
	}
	r := make([]T, l)
	if n >= 0 {
		copy(r[:len(ys)], ys)
	} else {
		copy(r[len(r)-len(ys):], ys)
	}
	return r
}

func padNSliceVs(n int64, ys []V) []V {
	l := n
	if l < 0 {
		l = -l
	}
	r := make([]V, l)
	pad := proto(ys)
	pad.immutable()
	if n >= 0 {
		copy(r[:len(ys)], ys)
		for i := len(ys); i < len(r); i++ {
			r[i] = pad
		}
	} else {
		copy(r[len(r)-len(ys):], ys)
		for i := 0; i < len(r)-len(ys); i++ {
			r[i] = pad
		}
	}
	return r
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
		return panicType("x rshift Y", "Y", y)
	}
}

func shiftBeforeAB(x V, yv *AB) V {
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		if isBI(x.I()) {
			var fl flags
			if yv.IsBoolean() && isbI(x.I()) {
				fl = flagBool
			}
			r := yv.reuse()
			copy(r.elts[max:], ys[:len(ys)-max])
			r.elts[0] = byte(x.I())
			r.flags = fl
			return NewV(r)
		}
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = int64(yv.At(i - max))
		}
		r[0] = x.I()
		return NewAIWithRC(r, reuseRCp(yv.rc))
	}
	if x.IsF() {
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = float64(yv.At(i - max))
		}
		r[0] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		fl := xv.flags & yv.flags & flagBool
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		copy(r.elts[:max], xv.elts)
		r.flags = fl
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = float64(ys[i-max])
		}
		copy(r[:max], xv.elts)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = int64(yv.At(i - max))
		}
		copy(r[:max], xv.elts)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		r.elts[0] = x.I()
		return NewV(r)
	} else if x.IsF() {
		if isI(x.F()) {
			r := yv.reuse()
			copy(r.elts[max:], ys[:len(ys)-max])
			r.elts[0] = int64(x.F())
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
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = int64(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i] = float64(ys[i-max])
		}
		copy(r[:max], xv.elts)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = xv.At(i)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		r.elts[0] = float64(x.I())
		return NewV(r)
	} else if x.IsF() {
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		r.elts[0] = x.F()
		return NewV(r)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = float64(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = xv.At(i)
		}
		return NewV(r)
	case *AI:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = float64(xv.At(i))
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	switch xv := x.value.(type) {
	case S:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		r.elts[0] = string(xv)
		return NewV(r)
	case *AS:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = xv.At(i)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	switch xv := x.value.(type) {
	case array:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		for i := 0; i < max; i++ {
			r.elts[i] = xv.at(i)
		}
		return Canonical(NewV(r))
	default:
		r := yv.reuse()
		copy(r.elts[max:], ys[:len(ys)-max])
		r.elts[0] = x
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
	copy(r[:max], xv.elts)
	return NewAV(r)
}

// nudge returns rshift x.
func nudge(x V) V {
	if x.Len() == 0 {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		fl := xv.flags & flagBool
		r := xv.reuse()
		copy(r.elts[1:], xv.elts[:xv.Len()-1])
		r.elts[0] = 0
		r.flags = fl
		return NewV(r)
	case *AI:
		r := xv.reuse()
		copy(r.elts[1:], xv.elts[:xv.Len()-1])
		r.elts[0] = 0
		return NewV(r)
	case *AF:
		r := xv.reuse()
		copy(r.elts[1:], xv.elts[:xv.Len()-1])
		r.elts[0] = 0
		return NewV(r)
	case *AS:
		r := xv.reuse()
		copy(r.elts[1:], xv.elts[:xv.Len()-1])
		r.elts[0] = ""
		return NewV(r)
	case *AV:
		r := xv.reuse()
		r0 := proto(xv.elts)
		copy(r.elts[1:], xv.elts[:xv.Len()-1])
		r.elts[0] = r0
		r0.InitWithRC(r.rc)
		return Canonical(NewV(r))
	case *Dict:
		return newDictValues(xv.keys, nudge(NewV(xv.values)))
	default:
		return panicType("rshift X", "X", x)
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
		return panicType("x shift Y", "Y", y)
	}
}

func shiftAfterAB(x V, yv *AB) V {
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		if isBI(x.I()) {
			var fl flags
			if yv.IsBoolean() && isbI(x.I()) {
				fl = flagBool
			}
			r := yv.reuse()
			copy(r.elts[:len(ys)-max], ys[max:])
			r.elts[len(ys)-1] = byte(x.I())
			r.flags = fl
			return NewV(r)
		}
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = int64(yv.At(i))
		}
		r[len(ys)-1] = x.I()
		return NewAIWithRC(r, reuseRCp(yv.rc))
	} else if x.IsF() {
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = float64(yv.At(i))
		}
		r[len(ys)-1] = x.F()
		return NewAFWithRC(r, reuseRCp(yv.rc))
	}
	switch xv := x.value.(type) {
	case *AB:
		fl := xv.flags & yv.flags & flagBool
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		copy(r.elts[len(ys)-max:], xv.elts)
		r.flags = fl
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = float64(ys[i])
		}
		copy(r[len(ys)-max:], xv.elts)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := make([]int64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = int64(yv.At(i))
		}
		copy(r[len(ys)-max:], xv.elts)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		r.elts[len(ys)-1] = x.I()
		return NewV(r)
	} else if x.IsF() {
		if isI(x.F()) {
			r := yv.reuse()
			copy(r.elts[:len(ys)-max], ys[max:])
			r.elts[len(ys)-1] = int64(x.F())
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
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = int64(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := make([]float64, len(ys))
		for i := max; i < len(ys); i++ {
			r[i-max] = float64(ys[i])
		}
		copy(r[len(ys)-max:], xv.elts)
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = xv.At(i)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	if x.IsI() {
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		r.elts[len(ys)-1] = float64(x.I())
		return NewV(r)
	} else if x.IsF() {
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		r.elts[len(ys)-1] = x.F()
		return NewV(r)
	}
	switch xv := x.value.(type) {
	case *AB:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = float64(xv.At(i))
		}
		return NewV(r)
	case *AF:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = xv.At(i)
		}
		return NewV(r)
	case *AI:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = float64(xv.At(i))
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	switch xv := x.value.(type) {
	case S:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		r.elts[len(ys)-1] = string(xv)
		return NewV(r)
	case *AS:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = xv.At(i)
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
	max := minInt(x.Len(), yv.Len())
	if max == 0 {
		return NewV(yv)
	}
	ys := yv.elts
	switch xv := x.value.(type) {
	case array:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		for i := 0; i < max; i++ {
			r.elts[len(ys)-max+i] = xv.at(i)
		}
		return Canonical(NewV(r))
	default:
		r := yv.reuse()
		copy(r.elts[:len(ys)-max], ys[max:])
		r.elts[len(ys)-1] = x
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
	copy(r[ylen-max:], xv.elts)
	return NewAV(r)
}

// NudgeBack returns shift x.
func nudgeBack(x V) V {
	if x.Len() == 0 {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		fl := xv.flags & flagBool
		r := xv.reuse()
		copy(r.elts[0:xv.Len()-1], xv.elts[1:])
		r.elts[xv.Len()-1] = 0
		r.flags = fl
		return NewV(r)
	case *AI:
		r := xv.reuse()
		copy(r.elts[0:xv.Len()-1], xv.elts[1:])
		r.elts[xv.Len()-1] = 0
		return NewV(r)
	case *AF:
		r := xv.reuse()
		copy(r.elts[0:xv.Len()-1], xv.elts[1:])
		r.elts[xv.Len()-1] = 0
		return NewV(r)
	case *AS:
		r := xv.reuse()
		copy(r.elts[0:xv.Len()-1], xv.elts[1:])
		r.elts[xv.Len()-1] = ""
		return NewV(r)
	case *AV:
		r := xv.reuse()
		rlast := proto(xv.elts)
		copy(r.elts[0:xv.Len()-1], xv.elts[1:])
		r.elts[xv.Len()-1] = rlast
		rlast.InitWithRC(r.rc)
		return Canonical(NewV(r))
	case *Dict:
		return newDictValues(xv.keys, nudgeBack(NewV(xv.values)))
	default:
		return panicType("shift X", "X", x)
	}
}

// windows returns i^y.
func windows(i int64, y V) V {
	switch yv := y.value.(type) {
	case S:
		if i < 0 && -i < int64(len(yv))+1 {
			return windowsString(-i, string(yv))
		}
		if i > 0 && i < int64(len(yv))+1 {
			return windowsString(int64(len(yv))-i+1, string(yv))
		}
		return Panicf("i^y : out of range i !%d (%d)", len(yv)+1, i)
	case array:
		if i < 0 && -i < int64(yv.Len())+1 {
			return windowsArray(-i, yv)
		}
		if i > 0 && i < int64(yv.Len())+1 {
			return windowsArray(int64(yv.Len())-i+1, yv)
		}
		return Panicf("i^y : out of range i !%d (%d)", yv.Len()+1, i)
	default:
		return panicType("i^y", "y", y)
	}
}

func windowsString(i int64, s string) V {
	r := make([]string, 1+len(s)-int(i))
	for j := range r {
		r[j] = s[j : j+int(i)]
	}
	return NewAS(r)
}

func windowsArray(i int64, y array) V {
	r := make([]V, 1+y.Len()-int(i))
	rc := y.RC()
	*rc += 2
	for j := range r {
		r[j] = Canonical(NewV(y.slice(j, j+int(i))))
	}
	var n int
	return NewAVWithRC(r, &n)
}

func shape(ctx *Context, x, y V) V {
	if x.IsI() {
		return cutShapeI(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i$y : non-integer i (%g)", x.F())
		}
		return cutShapeI(int64(x.F()), y)
	}
	switch xv := x.value.(type) {
	case S:
		return cast(ctx, xv, y)
	case array:
		return search(x, y)
	default:
		return panicType("x$y", "x", x)
	}
}

func cutShapeI(i int64, y V) V {
	if y.IsI() {
		return NewI(1 + y.I() - i)
	}
	if y.IsF() {
		return NewF(1 + y.F() - float64(i))
	}
	switch yv := y.value.(type) {
	case S:
		if i < 0 && -i < int64(len(yv))+1 {
			return cutColsString(-i, string(yv))
		}
		if i > 0 && i < int64(len(yv))+1 {
			return cutLinesString(int(i), string(yv))
		}
		return Panicf("i$s : out of range i (%d)", i)
	case array:
		if i < 0 && -i < int64(yv.Len())+1 {
			return cutColsArray(-i, yv)
		}
		if i > 0 && i < int64(yv.Len())+1 {
			return cutLinesArray(int(i), yv)
		}
		return Panicf("i$Y : out of range i (%d)", i)
	case *Dict:
		k := cutShapeI(i, NewV(yv.keys))
		if k.IsPanic() {
			return k
		}
		v := cutShapeI(i, NewV(yv.values))
		if v.IsPanic() {
			return v
		}
		return NewDict(k, v)
	default:
		return panicType("i$y", "y", y)
	}
}

func cutColsString(i int64, s string) V {
	if i >= int64(len(s)) {
		return NewAS([]string{s})
	}
	n := len(s) / int(i)
	if len(s)%int(i) != 0 {
		n++
	}
	r := make([]string, n)
	for j := 0; j < n; j++ {
		from := j * int(i)
		to := minInt(from+int(i), len(s))
		r[j] = s[from:to]
	}
	return NewAS(r)
}

func cutLinesString(n int, s string) V {
	r := make([]string, n)
	if n == 1 {
		return NewAS([]string{s})
	}
	from := 0
	for j := 0; j < n; j++ {
		to := minInt(from+(len(s)-from)/(n-j), len(s))
		r[j] = s[from:to]
		from = to
	}
	return NewAS(r)
}

func cutColsArray(i int64, y array) V {
	ylen := y.Len()
	if i >= int64(ylen) {
		return NewAVWithRC([]V{NewV(y)}, y.RC())
	}
	n := ylen / int(i)
	if ylen%int(i) != 0 {
		n++
	}
	rc := y.RC()
	*rc += 2
	r := make([]V, n)
	for j := 0; j < n; j++ {
		from := j * int(i)
		to := minInt(from+int(i), ylen)
		r[j] = Canonical(NewV(y.slice(from, to)))
	}
	var rcn int
	return NewAVWithRC(r, &rcn)
}

func cutLinesArray(n int, y array) V {
	ylen := y.Len()
	if n == 1 {
		return NewAVWithRC([]V{NewV(y)}, y.RC())
	}
	rc := y.RC()
	*rc += 2
	r := make([]V, n)
	from := 0
	for j := 0; j < n; j++ {
		to := minInt(from+(ylen-from)/(n-j), ylen)
		r[j] = Canonical(NewV(y.slice(from, to)))
		from = to
	}
	var rcn int
	return NewAVWithRC(r, &rcn)
}
