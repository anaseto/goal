// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch x := x.Value.(type) {
	case array:
		return x.Len()
	default:
		return 1
	}
}

func reverseMut(x V) {
	switch x := x.Value.(type) {
	case *AB:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case *AF:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case *AI:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case *AS:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case *AV:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case sort.Interface:
		sort.Reverse(x)
	}
}

// reverse returns |x.
func reverse(x V) V {
	switch xv := x.Value.(type) {
	case array:
		r := cloneShallow(x)
		reverseMut(r)
		return r
	default:
		return errType("|x", "x", xv)
	}
}

// Rotate returns f|y.
func rotate(x, y V) V {
	i := 0
	if x.IsInt() {
		i = x.Int()
	} else {
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("f|y : non-integer f[y] (%s)", x.Type())
			}
			i = int(x)
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
	switch y := y.Value.(type) {
	case *AB:
		r := make([]bool, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return NewV(r)
	case *AF:
		r := make([]float64, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return NewV(r)
	case *AI:
		r := make([]int, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return NewV(r)
	case *AS:
		r := make([]string, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return NewV(r)
	case *AV:
		r := make([]V, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return NewV(r)
	default:
		return errType("f|y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch x := x.Value.(type) {
	case array:
		if x.Len() == 0 {
			switch x.(type) {
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
		return x.at(0)
	default:
		return NewV(x)
	}
}

// drop returns i_x and s_x.
func drop(x, y V) V {
	if x.IsInt() {
		return dropi(x.Int(), y)
	}
	switch x := x.Value.(type) {
	case F:
		if !isI(x) {
			return errf("i_y : non-integer i (%s)", x.Type())
		}
		return dropi(int(x), y)
	case S:
		return drops(x, y)
	case *AB:
		return drop(fromABtoAI(x), y)
	case *AI:
		return cutAI(x, y)
	case *AF:
		z := toAI(x)
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

func cutAI(x AI, y V) V {
	if !sort.IsSorted(sort.IntSlice(x)) {
		return errs("x_y : x is not ascending")
	}
	ylen := Length(y)
	for _, i := range x {
		if i < 0 || i > ylen {
			return errf("x_y : x contains out of bound index (%d)", i)
		}
	}
	if x.Len() == 0 {
		return NewV(AV{})
	}
	switch yv := y.Value.(type) {
	case *AB:
		r := make([]V, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = NewV(yv[from:to])
		}
		return NewV(r)
	case *AI:
		r := make([]V, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = NewV(yv[from:to])
		}
		return NewV(r)
	case *AF:
		r := make([]V, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = NewV(yv[from:to])
		}
		return NewV(r)
	case *AS:
		r := make([]V, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = NewV(yv[from:to])
		}
		return NewV(r)
	case *AV:
		r := make([]V, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = NewV(yv[from:to])
		}
		return canonicalV(NewAV(r))
	default:
		return errf("x_y : y not an array (%s)", yv.Type())
	}
}

// take returns i#x.
func take(x, y V) V {
	i := 0
	if x.IsInt() {
		i = x.Int()
	} else {
		switch x := x.Value.(type) {
		case F:
			if !isI(x) {
				return errf("i#y : non-integer i (%s)", x.Type())
			}
			i = int(x)
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
	switch y := y.(type) {
	case *AB:
		r := make([]bool, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return NewV(r)
	case *AF:
		r := make([]float64, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return NewV(r)
	case *AI:
		r := make([]int, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return NewV(r)
	case *AS:
		r := make([]string, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return NewV(r)
	case *AV:
		r := make([]V, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return NewV(r)
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
	switch y := y.Value.(type) {
	case *AB:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]bool, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(B2F(y[i-max]))
			}
			return NewV(r)
		case *AI:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = int(B2I(y[i-max]))
			}
			return NewV(r)
		default:
			return errType("x»y", "y", y)
		}
	case *AF:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(B2F(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		case *AI:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(x[i])
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		default:
			return errType("x»y", "y", y)
		}
	case *AI:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[i] = int(B2I(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(y[i-max])
			}
			return NewV(r)
		case *AI:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		default:
			return errType("x»y", "y", y)
		}
	case *AS:
		switch x := x.Value.(type) {
		case *AS:
			r := make([]string, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return NewV(r)
		default:
			return errType("x»y", "y", y)
		}
	case *AV:
		switch x := x.Value.(type) {
		case array:
			r := make([]V, len(y))
			for i := 0; i < max; i++ {
				r[i] = x.at(i)
			}
			copy(r[max:], y[:len(y)-max])
			return canonicalV(NewAV(r))
		default:
			return errType("x»y", "y", y)
		}
	default:
		return errs("x»y: y not an array")
	}
}

// nudge returns »x. XXX unused for now
func nudge(x V) V {
	switch x := x.Value.(type) {
	case *AB:
		r := make([]bool, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return NewV(r)
	case *AI:
		r := make([]int, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return NewV(r)
	case *AF:
		r := make([]float64, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return NewV(r)
	case *AS:
		r := make([]string, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return NewV(r)
	case *AV:
		r := make([]V, x.Len())
		copy(r[1:], x[0:x.Len()-1])
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
	switch y := y.Value.(type) {
	case *AB:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]bool, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(B2F(y[i]))
			}
			return NewV(r)
		case *AI:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = int(B2I(y[i]))
			}
			return NewV(r)
		default:
			return errType("x«y", "y", y)
		}
	case *AF:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(B2F(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		case *AI:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(x[i])
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		default:
			return errType("x«y", "y", y)
		}
	case *AI:
		switch x := x.Value.(type) {
		case *AB:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = int(B2I(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		case *AF:
			r := make([]float64, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(y[max])
			}
			return NewV(r)
		case *AI:
			r := make([]int, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		default:
			return errType("x«y", "y", y)
		}
	case *AS:
		switch x := x.Value.(type) {
		case *AS:
			r := make([]string, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return NewV(r)
		default:
			return errType("x«y", "y", y)
		}
	case *AV:
		switch x := x.Value.(type) {
		case array:
			r := make([]V, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x.at(i)
			}
			copy(r[:len(y)-max], y[max:])
			return canonicalV(NewAV(r))
		default:
			return errType("x«y", "y", y)
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
	switch x := x.Value.(type) {
	case *AB:
		r := make([]bool, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return NewV(r)
	case *AI:
		r := make([]int, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return NewV(r)
	case *AF:
		r := make([]float64, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return NewV(r)
	case *AS:
		r := make([]string, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return NewV(r)
	case *AV:
		r := make([]V, x.Len())
		copy(r[0:x.Len()-1], x[1:])
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
