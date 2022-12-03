// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch xv := x.Value.(type) {
	case array:
		return xv.Len()
	default:
		return 1
	}
}

func reverseMut(x V) {
	switch xv := x.Value.(type) {
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
	switch xv := x.Value.(type) {
	case array:
		if !xv.reusable() {
			x = cloneShallow(x)
		}
		reverseMut(x)
		return x
	default:
		return errType("|x", "x", x)
	}
}

// Rotate returns f|y.
func rotate(x, y V) V {
	i := 0
	if x.IsInt() {
		i = x.Int()
	} else {
		switch xv := x.Value.(type) {
		case F:
			if !isI(xv) {
				return errf("f|y : non-integer f[y] (%s)", xv.Type())
			}
			i = int(xv)
		default:
			return errf("f|y : non-integer f[y] (%s)", x.Type())
		}
	}
	lenx := Length(y)
	if lenx == 0 {
		return y
	}
	i %= lenx
	if i < 0 {
		i += lenx
	}
	switch yv := y.Value.(type) {
	case *AB:
		r := make([]bool, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = yv.At((j + i) % lenx)
		}
		return NewAB(r)
	case *AF:
		r := make([]float64, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = yv.At((j + i) % lenx)
		}
		return NewAF(r)
	case *AI:
		r := make([]int, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = yv.At((j + i) % lenx)
		}
		return NewAI(r)
	case *AS:
		r := make([]string, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = yv.At((j + i) % lenx)
		}
		return NewAS(r)
	case *AV:
		r := make([]V, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = yv.At((j + i) % lenx)
		}
		return NewAV(r)
	default:
		return errType("f|y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch xv := x.Value.(type) {
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
	if x.IsInt() {
		return dropi(x.Int(), y)
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("i_y : non-integer i (%s)", xv.Type())
		}
		return dropi(int(xv), y)
	case S:
		return drops(xv, y)
	case *AB:
		return drop(fromABtoAI(xv), y)
	case *AI:
		return cutAI(xv, y)
	case *AF:
		z := toAI(xv)
		if z.IsErr() {
			return z
		}
		return drop(z, y)
	case *AV:
		//assertCanonical(x)
		return errs("x_y : x non-integer")
	default:
		return errf("x_y : bad type i (%s)", x.Type())
	}
}

func dropi(i int, y V) V {
	switch yv := y.Value.(type) {
	case array:
		switch {
		case i >= 0:
			if i > yv.Len() {
				i = yv.Len()
			}
			y.Value = yv.slice(i, yv.Len())
			return canonicalV(y)
		default:
			i = yv.Len() + i
			if i < 0 {
				i = 0
			}
			y.Value = yv.slice(0, i)
			return canonicalV(y)
		}
	default:
		return errs("i_y : y not an array")
	}
}

func cutAI(x *AI, y V) V {
	if !sort.IsSorted(sort.IntSlice(x.Slice)) {
		return errs("x_y : x is not ascending")
	}
	ylen := Length(y)
	for _, i := range x.Slice {
		if i < 0 || i > ylen {
			return errf("x_y : x contains out of bound index (%d)", i)
		}
	}
	if x.Len() == 0 {
		return NewAV([]V{})
	}
	switch yv := y.Value.(type) {
	case *AB:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := yv.Len()
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAB(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AI:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := yv.Len()
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAI(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AF:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := yv.Len()
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAF(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AS:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := yv.Len()
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAS(yv.Slice[from:to])
		}
		return NewAV(r)
	case *AV:
		r := make([]V, x.Len())
		for i, from := range x.Slice {
			to := yv.Len()
			if i+1 < x.Len() {
				to = x.At(i + 1)
			}
			r[i] = NewAV(yv.Slice[from:to])
		}
		return canonicalV(NewAV(r))
	default:
		return errf("x_y : y not an array (%s)", y.Type())
	}
}

// take returns i#x.
func take(x, y V) V {
	i := 0
	if x.IsInt() {
		i = x.Int()
	} else {
		switch xv := x.Value.(type) {
		case F:
			if !isI(xv) {
				return errf("i#y : non-integer i (%s)", xv.Type())
			}
			i = int(xv)
		default:
			return errf("i#y : non-integer i (%s)", x.Type())
		}
	}
	yv := toArray(y).Value.(array)
	switch {
	case i >= 0:
		if i > yv.Len() {
			return takeCyclic(yv, i)
		}
		y.Value = yv.slice(0, i)
		return canonicalV(y)
	default:
		if i < -yv.Len() {
			return takeCyclic(yv, i)
		}
		y.Value = yv.slice(yv.Len()+i, yv.Len())
		return canonicalV(y)
	}
}

func takeCyclic(y array, n int) V {
	neg := n < 0
	if neg {
		n = -n
	}
	i := 0
	step := y.Len()
	switch yv := y.(type) {
	case *AB:
		ys := yv.Slice
		r := make([]bool, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[len(ys)-n+i:])
		} else {
			copy(r[i:n], ys[:n-i])
		}
		return NewAB(r)
	case *AI:
		ys := yv.Slice
		r := make([]int, n)
		for i+step < n {
			copy(r[i:i+step], ys)
			i += step
		}
		if neg {
			copy(r[i:n], ys[len(ys)-n+i:])
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
			copy(r[i:n], ys[len(ys)-n+i:])
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
			copy(r[i:n], ys[len(ys)-n+i:])
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
			copy(r[i:n], ys[len(ys)-n+i:])
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
	max := minI(Length(x), Length(y))
	if max == 0 {
		return y
	}
	switch yv := y.Value.(type) {
	case *AB:
		ys := yv.Slice
		switch xv := x.Value.(type) {
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
				r[i] = float64(B2F(ys[i-max]))
			}
			return NewAF(r)
		case *AI:
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i] = int(B2I(ys[i-max]))
			}
			return NewAI(r)
		default:
			return errType("x»y", "x", x)
		}
	case *AF:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AB:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[i] = float64(B2F(xv.At(i)))
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
			return errType("x»y", "x", x)
		}
	case *AI:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AB:
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[i] = int(B2I(xv.At(i)))
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
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAI(r)
		default:
			return errType("x»y", "x", x)
		}
	case *AS:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AS:
			r := make([]string, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.At(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return NewAS(r)
		default:
			return errType("x»y", "x", x)
		}
	case *AV:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case array:
			r := make([]V, len(ys))
			for i := 0; i < max; i++ {
				r[i] = xv.at(i)
			}
			copy(r[max:], ys[:len(ys)-max])
			return canonicalV(NewAV(r))
		default:
			return errType("x»y", "x", x)
		}
	default:
		return errs("x»y: y not an array")
	}
}

// nudge returns »x. XXX unused for now
func nudge(x V) V {
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r[1:], xv.Slice[0:xv.Len()-1])
		return NewAB(r)
	case *AI:
		r := make([]int, xv.Len())
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
		return errs("»x : not an array")
	}
}

// ShiftAfter returns x«y. XXX: unused for now.
func shiftAfter(x, y V) V {
	x = toArray(x)
	max := minI(Length(x), Length(y))
	if max == 0 {
		return y
	}
	switch yv := y.Value.(type) {
	case *AB:
		ys := yv.Slice
		switch xv := x.Value.(type) {
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
				r[i-max] = float64(B2F(ys[i]))
			}
			return NewAF(r)
		case *AI:
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			for i := max; i < len(ys); i++ {
				r[i-max] = int(B2I(ys[i]))
			}
			return NewAI(r)
		default:
			return errType("x«y", "x", x)
		}
	case *AF:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AB:
			r := make([]float64, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = float64(B2F(xv.At(i)))
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
			return errType("x«y", "x", x)
		}
	case *AI:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AB:
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = int(B2I(xv.At(i)))
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
			r := make([]int, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAI(r)
		default:
			return errType("x«y", "x", x)
		}
	case *AS:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case *AS:
			r := make([]string, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.At(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return NewAS(r)
		default:
			return errType("x«y", "x", x)
		}
	case *AV:
		ys := yv.Slice
		switch xv := x.Value.(type) {
		case array:
			r := make([]V, len(ys))
			for i := 0; i < max; i++ {
				r[len(ys)-1-i] = xv.at(i)
			}
			copy(r[:len(ys)-max], ys[max:])
			return canonicalV(NewAV(r))
		default:
			return errType("x«y", "x", x)
		}
	default:
		return errs("x«y: y not an array")
	}
}

// NudgeBack returns «x. XXX unused for now
func nudgeBack(x V) V {
	if Length(x) == 0 {
		return x
	}
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r[0:xv.Len()-1], xv.Slice[1:])
		return NewAB(r)
	case *AI:
		r := make([]int, xv.Len())
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
		return errs("«x : x not an array")
	}
}

// windows returns i^y.
func windows(i int, y V) V {
	switch yv := y.Value.(type) {
	case array:
		if i <= 0 || i >= yv.Len()+1 {
			return errf("i^y : i out of range !%d (%d)", yv.Len()+1, i)
		}
		r := make([]V, 1+yv.Len()-i)
		for j := range r {
			yc := y
			yc.Value = yv.slice(j, j+i)
			r[j] = canonicalV(yc)
		}
		return canonicalV(NewAV(r))
	default:
		return errs("i^y : y not an array")
	}
}
