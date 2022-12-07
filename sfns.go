// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch xv := x.value.(type) {
	case array:
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
	case sort.Interface:
		sort.Reverse(xv)
	}
}

// reverse returns |x.
func reverse(x V) V {
	switch xv := x.value.(type) {
	case array:
		x.value = cloneShallowArray(xv)
		reverseMut(x)
		return x
	default:
		return panicType("|x", "x", x)
	}
}

// Rotate returns f|y.
func rotate(x, y V) V {
	i := int64(0)
	if x.IsI() {
		i = x.I()
	} else if x.IsF() {
		if !isI(x.F()) {
			return panicf("f|y : non-integer f[y] (%g)", x.F())
		}
		i = int64(x.F())
	} else {
		return panicf("f|y : non-integer f[y] (%s)", x.Type())
	}
	ylen := int64(Length(y))
	if ylen == 0 {
		return y
	}
	i %= ylen
	if i < 0 {
		i += ylen
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]bool, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAF(r)
	case *AI:
		r := make([]int64, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAI(r)
	case *AS:
		r := make([]string, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, ylen)
		for j := int64(0); j < ylen; j++ {
			r[j] = yv.At(int((j + i) % ylen))
		}
		return NewAV(r)
	default:
		return panicType("f|y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch xv := x.value.(type) {
	case *errV:
		return xv.V
	case array:
		if xv.Len() == 0 {
			switch xv.(type) {
			case *AB:
				return NewI(0)
			case *AF:
				return NewF(0)
			case *AI:
				return NewI(0)
			case *AS:
				return NewS("")
			default:
				return V{}
			}
		}
		return xv.at(0)
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
			return panicf("i_y : non-integer i (%g)", x.F())
		}
		return dropi(int64(x.F()), y)
	}
	switch xv := x.value.(type) {
	case S:
		return drops(xv, y)
	case *AB:
		return drop(fromABtoAI(xv), y)
	case *AI:
		return cutAI(xv, y)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return drop(z, y)
	case *AV:
		//assertCanonical(x)
		return panics("x_y : x non-integer")
	default:
		return panicf("x_y : bad type i (%s)", x.Type())
	}
}

func dropi(i int64, y V) V {
	switch yv := y.value.(type) {
	case array:
		switch {
		case i >= 0:
			if i > int64(yv.Len()) {
				i = int64(yv.Len())
			}
			y.value = yv.slice(int(i), yv.Len())
			return canonicalV(y)
		default:
			i = int64(yv.Len()) + i
			if i < 0 {
				i = 0
			}
			y.value = yv.slice(0, int(i))
			return canonicalV(y)
		}
	default:
		return panics("i_y : y not an array")
	}
}

func cutAI(x *AI, y V) V {
	if !sort.IsSorted(sortAI(x.Slice)) {
		return panics("x_y : x is not ascending")
	}
	ylen := int64(Length(y))
	for _, i := range x.Slice {
		if i < 0 || i > ylen {
			return panicf("x_y : x contains out of bound index (%d)", i)
		}
	}
	if x.Len() == 0 {
		return NewAV([]V{})
	}
	switch yv := y.value.(type) {
	case *AB:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := int64(yv.Len())
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAB(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AI:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := int64(yv.Len())
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAI(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AF:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := int64(yv.Len())
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAF(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AS:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := int64(yv.Len())
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAS(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AV:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := int64(yv.Len())
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAV(yv.Slice[from:to])
		}
		return canonicalV(NewAV(r))
	default:
		return panicf("x_y : y not an array (%s)", y.Type())
	}
}

// take returns i#x.
func take(x, y V) V {
	i := int64(0)
	if x.IsI() {
		i = x.I()
	} else if x.IsF() {
		if !isI(x.F()) {
			return panicf("i#y : non-integer i (%g)", x.F())
		}
		i = int64(x.F())
	} else {
		return panicf("i#y : non-integer i (%s)", x.Type())
	}
	yv := toArray(y).value.(array)
	switch {
	case i >= 0:
		if i > int64(yv.Len()) {
			return takeCyclic(yv, i)
		}
		y.value = yv.slice(0, int(i))
		return canonicalV(y)
	default:
		if i < int64(-yv.Len()) {
			return takeCyclic(yv, i)
		}
		y.value = yv.slice(yv.Len()+int(i), yv.Len())
		return canonicalV(y)
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
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[int64(len(ys))-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAB(r)
	case *AI:
		ys := yv.Slice
		r := make([]int64, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[int64(len(ys))-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAI(r)
	case *AF:
		ys := yv.Slice
		r := make([]float64, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[int64(len(ys))-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAF(r)
	case *AS:
		ys := yv.Slice
		r := make([]string, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[int64(len(ys))-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAS(r)
	case *AV:
		ys := yv.Slice
		r := make([]V, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[int64(len(ys))-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAV(r)
	default:
		return NewV(y)
	}
}

// ShiftBefore returns x»y. XXX: unused for now.
func shiftBefore(x, y V) V {
	x = toArray(x)
	max := minInt(Length(x), Length(y))
	if max == 0 {
		return y
	}
	switch yv := y.value.(type) {
	case *AB:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]bool, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAB(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i] = float64(b2f(ys[i-max]))
			}
			return NewAF(r)
		case *AI:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i] = b2i(ys[i-max])
			}
			return NewAI(r)
		default:
			return panicType("x»y", "x", x)
		}
	case *AF:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = float64(b2f(xv.At(i)))
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAF(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAF(r)
		case *AI:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = float64(xv.At(i))
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAF(r)
		default:
			return panicType("x»y", "x", x)
		}
	case *AI:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = b2i(xv.At(i))
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAI(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i] = float64(ys[i-max])
			}
			return NewAF(r)
		case *AI:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAI(r)
		default:
			return panicType("x»y", "x", x)
		}
	case *AS:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AS:
			r := make([]string, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAS(r)
		default:
			return panicType("x»y", "x", x)
		}
	case *AV:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case array:
			r := make([]V, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.at(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return canonicalV(NewAV(r))
		default:
			return panicType("x»y", "x", x)
		}
	default:
		return panics("x»y: y not an array")
	}
}

// nudge returns »x. XXX unused for now
func nudge(x V) V {
	switch xv := x.value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return NewAB(r)
	case *AI:
		r := make([]int64, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return NewAI(r)
	case *AF:
		r := make([]float64, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return NewAF(r)
	case *AS:
		r := make([]string, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return NewAS(r)
	case *AV:
		r := make([]V, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return canonicalV(NewAV(r))
	default:
		return panics("»x : not an array")
	}
}

// ShiftAfter returns x«y. XXX: unused for now.
func shiftAfter(x, y V) V {
	x = toArray(x)
	max := minInt(Length(x), Length(y))
	if max == 0 {
		return y
	}
	switch yv := y.value.(type) {
	case *AB:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]bool, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAB(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i-max] = float64(b2f(ys[i]))
			}
			return NewAF(r)
		case *AI:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i-max] = b2i(ys[i])
			}
			return NewAI(r)
		default:
			return panicType("x«y", "x", x)
		}
	case *AF:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = float64(b2f(xv.At(i)))
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAF(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAF(r)
		case *AI:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = float64(xv.At(i))
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAF(r)
		default:
			return panicType("x«y", "x", x)
		}
	case *AI:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AB:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = b2i(xv.At(i))
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAI(r)
		case *AF:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i-max] = float64(ys[max])
			}
			return NewAF(r)
		case *AI:
			r := make([]int64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAI(r)
		default:
			return panicType("x«y", "x", x)
		}
	case *AS:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case *AS:
			r := make([]string, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAS(r)
		default:
			return panicType("x«y", "x", x)
		}
	case *AV:
		ys := yv.Slice
		switch xv := x.value.(type) {
		case array:
			r := make([]V, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.at(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return canonicalV(NewAV(r))
		default:
			return panicType("x«y", "x", x)
		}
	default:
		return panics("x«y: y not an array")
	}
}

// NudgeBack returns «x. XXX unused for now
func nudgeBack(x V) V {
	if Length(x) == 0 {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return NewAB(r)
	case *AI:
		r := make([]int64, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return NewAI(r)
	case *AF:
		r := make([]float64, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return NewAF(r)
	case *AS:
		r := make([]string, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return NewAS(r)
	case *AV:
		r := make([]V, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return canonicalV(NewAV(r))
	default:
		return panics("«x : x not an array")
	}
}

// windows returns i^y.
func windows(i int64, y V) V {
	switch yv := y.value.(type) {
	case array:
		if i <= 0 || i >= int64(yv.Len()+1) {
			return panicf("i^y : i out of range !%d (%d)", yv.Len()+1, i)
		}
		r := make([]V, 1+yv.Len()-int(i))
		for j := range r {
			yc := y
			yc.value = yv.slice(j, j+int(i))
			r[j] = canonicalV(yc)
		}
		return NewAV(r)
	default:
		return panics("i^y : y not an array")
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
			return panicf("i$y : i non-integer (%g)", f)
		}
		i = int64(f)
	}
	switch yv := y.value.(type) {
	case array:
		ylen := yv.Len()
		if i <= 0 {
			return panicf("i$y : i not positive (%d)", i)
		}
		if i >= int64(ylen) {
			return NewAV([]V{y})
		}
		n := ylen / int(i)
		if ylen%int(i) != 0 {
			n++
		}
		r := make([]V, n)
		for j := 0; j < n; j++ {
			yc := y
			from := j * int(i)
			to := minInt(from+int(i), ylen)
			yc.value = yv.slice(from, to)
			r[j] = canonicalV(yc)
		}
		return NewAV(r)
	default:
		return panics("i$y : y not an array")
	}
}
