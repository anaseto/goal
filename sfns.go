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

func reverseMut(x V) {
	switch xv := x.value.(type) {
	case *AB:
		xs := xv.Slice
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
		}
	case *AF:
		xs := xv.Slice
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
		}
	case *AI:
		xs := xv.Slice
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
		}
	case *AS:
		xs := xv.Slice
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
		}
	case *AV:
		xs := xv.Slice
		for i := 0; i < len(xs)/2; i++ {
			xs[i], xs[len(xs)-i-1] = xs[len(xs)-i-1], xs[i]
		}
		//case sort.Interface:
		//sort.Reverse(xv)
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
		return Panicf("|x : x not an array (%s)", x.Type())
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

func rotateI(i int64, y V) V {
	switch yv := y.value.(type) {
	case *AB:
		ylen := int64(yv.Len())
		r := make([]bool, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AF:
		ylen := int64(yv.Len())
		r := make([]float64, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AI:
		ylen := int64(yv.Len())
		r := make([]int64, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AS:
		ylen := int64(yv.Len())
		r := make([]string, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		ylen := int64(yv.Len())
		r := make([]V, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAVWithRC(r, yv.rc)
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
		return Panicf("x rotate y : y not an array (%s)", y.Type())
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

// drop returns i_x and s_x.
func drop(x, y V) V {
	if x.IsI() {
		return dropi(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i_y : non-integer i (%g)", x.F())
		}
		return dropi(int64(x.F()), y)
	}
	switch xv := x.value.(type) {
	case S:
		return drops(xv, y)
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

func dropi(i int64, y V) V {
	switch yv := y.value.(type) {
	case S:
		switch {
		case i >= 0:
			if i > int64(len(yv)) {
				i = int64(len(yv))
			}
			return NewV(yv[i:])
		default:
			i = int64(len(yv)) + i
			if i < 0 {
				i = 0
			}
			return NewV(yv[:i])
		}
	case array:
		switch {
		case i >= 0:
			if i > int64(yv.Len()) {
				i = int64(yv.Len())
			}
			y.value = yv.slice(int(i), yv.Len())
			return Canonical(y)
		default:
			i = int64(yv.Len()) + i
			if i < 0 {
				i = 0
			}
			y.value = yv.slice(0, int(i))
			return Canonical(y)
		}
	case *Dict:
		rk := dropi(i, NewV(yv.keys))
		rk.InitRC()
		rv := dropi(i, NewV(yv.values))
		rv.InitRC()
		return NewV(&Dict{
			keys:   rk.value.(array),
			values: rv.value.(array)})
	default:
		return panics("i_y : y not an array")
	}
}

func cutAIarray(x *AI, y array) V {
	if !x.flags.Has(flagAscending) && !sort.IsSorted(sortAI(x.Slice)) {
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
	if !x.flags.Has(flagAscending) && !sort.IsSorted(sortAI(x.Slice)) {
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
	i := int64(0)
	if x.IsI() {
		i = x.I()
	} else if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i#y : non-integer i (%g)", x.F())
		}
		i = int64(x.F())
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
		rk := takei(i, yv.keys)
		rk.InitRC()
		rv := takei(i, yv.values)
		rv.InitRC()
		return NewV(&Dict{
			keys:   rk.value.(array),
			values: rv.value.(array)})
	case array:
		return takei(i, yv)
	default:
		return takei(i, toArray(y).value.(array))
	}
}

func takei(i int64, y array) V {
	if y.Len() == 0 {
		if i < 0 {
			i = -i
		}
		r := make([]bool, i)
		return NewABWithRC(r, reuseRCp(y.RC()))
	}
	switch {
	case i >= 0:
		if i > int64(y.Len()) {
			return takeCyclic(y, i)
		}
		return Canonical(NewV(y.slice(0, int(i))))
	default:
		if i < int64(-y.Len()) {
			return takeCyclic(y, i)
		}
		return Canonical(NewV(y.slice(y.Len()+int(i), y.Len())))
	}
}

func takeCyclic(y array, n int64) V {
	neg := n < 0
	if neg {
		n = -n
	}
	i := int64(0)
	step := int64(y.Len())
	switch yv := y.(type) {
	case *AB:
		ys := yv.Slice
		r := make([]bool, n)
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
		return NewABWithRC(r, reuseRCp(yv.rc))
	case *AI:
		ys := yv.Slice
		r := make([]int64, n)
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
		return NewAIWithRC(r, reuseRCp(yv.rc))
	case *AF:
		ys := yv.Slice
		r := make([]float64, n)
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
		return NewAFWithRC(r, reuseRCp(yv.rc))
	case *AS:
		ys := yv.Slice
		r := make([]string, n)
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
		return NewASWithRC(r, reuseRCp(yv.rc))
	case *AV:
		ys := yv.Slice
		r := make([]V, n)
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
		return NewAVWithRC(r, yv.rc)
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
		return panics("x rshift y: y not an array")
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
		return panics("rshift x : x not an array")
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
		return panics("x shift y: y not an array")
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
		return panics("shift x : x not an array")
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

// shapeSplit returns i$y.
func shapeSplit(x V, y V) V {
	var i int64
	if x.IsI() {
		i = x.I()
	} else {
		// x.IsF() should be true
		f := x.F()
		if !isI(f) {
			return Panicf("i$y : i non-integer (%g)", f)
		}
		i = int64(f)
	}
	switch yv := y.value.(type) {
	case array:
		ylen := yv.Len()
		if i <= 0 {
			return Panicf("i$y : i not positive (%d)", i)
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
		return panics("i$y : y not an array")
	}
}
