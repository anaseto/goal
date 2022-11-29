// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
	"strings"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch x := x.BV.(type) {
	case array:
		return x.Len()
	default:
		return 1
	}
}

func reverseMut(x V) {
	switch x := x.BV.(type) {
	case AB:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AF:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AI:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AS:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case AV:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case sort.Interface:
		sort.Reverse(x)
	}
}

// reverse returns |x.
func reverse(x V) V {
	switch xv := x.BV.(type) {
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
	switch x := x.BV.(type) {
	case I:
		i = int(x)
	case F:
		if !isI(x) {
			return errf("f|y : non-integer f[y] (%s)", x.Type())
		}
		i = int(x)
	default:
		return errf("f|y : non-integer f[y] (%s)", x.Type())
	}
	lenx := Length(y)
	if lenx == 0 {
		return y
	}
	i %= lenx
	if i < 0 {
		i += lenx
	}
	switch y := y.BV.(type) {
	case AB:
		r := make(AB, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return newBV(r)
	case AF:
		r := make(AF, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return newBV(r)
	case AI:
		r := make(AI, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return newBV(r)
	case AS:
		r := make(AS, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return newBV(r)
	case AV:
		r := make(AV, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return newBV(r)
	default:
		return errType("f|y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch x := x.BV.(type) {
	case array:
		if x.Len() == 0 {
			switch x.(type) {
			case AB:
				return newBV(I(0))
			case AF:
				return newBV(F(0))
			case AI:
				return newBV(I(0))
			case AS:
				return newBV(S(""))
			default:
				return V{}
			}
		}
		return x.at(0)
	default:
		return newBV(x)
	}
}

// drop returns i_x and s_x.
func drop(x, y V) V {
	switch x := x.BV.(type) {
	case I:
		return dropi(int(x), y)
	case F:
		if !isI(x) {
			return errf("i_y : non-integer i (%s)", x.Type())
		}
		return dropi(int(x), y)
	case S:
		return drops(x, y)
	case AB:
		return drop(fromABtoAI(x), y)
	case AI:
		return cutAI(x, y)
	case AF:
		z := toAI(x)
		if isErr(z) {
			return z
		}
		return drop(z, y)
	case AV:
		//assertCanonical(x)
		return errs("x_y : x non-integer")
	default:
		return errf("x_y : bad type i (%s)", x.Type())
	}
}

func dropi(i int, y V) V {
	switch y := y.BV.(type) {
	case array:
		switch {
		case i >= 0:
			if i > y.Len() {
				i = y.Len()
			}
			return newBV(canonicalArray(y.slice(i, y.Len())))
		default:
			i = y.Len() + i
			if i < 0 {
				i = 0
			}
			return newBV(canonicalArray(y.slice(0, i)))
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
		return newBV(AV{})
	}
	switch yv := y.BV.(type) {
	case AB:
		r := make(AV, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = newBV(yv[from:to])
		}
		return newBV(r)
	case AI:
		r := make(AV, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = newBV(yv[from:to])
		}
		return newBV(r)
	case AF:
		r := make(AV, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = newBV(yv[from:to])
		}
		return newBV(r)
	case AS:
		r := make(AV, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = newBV(yv[from:to])
		}
		return newBV(r)
	case AV:
		r := make(AV, x.Len())
		for i, from := range x {
			to := len(yv)
			if i+1 < x.Len() {
				to = x[i+1]
			}
			r[i] = newBV(yv[from:to])
		}
		return newBV(canonical(r))
	default:
		return errf("x_y : y not an array (%s)", yv.Type())
	}
}

func drops(s S, y V) V {
	switch y := y.BV.(type) {
	case S:
		return newBV(S(strings.TrimPrefix(string(y), string(s))))
	case AS:
		r := make(AS, y.Len())
		for i, yi := range y {
			r[i] = strings.TrimPrefix(string(yi), string(s))
		}
		return newBV(r)
	case AV:
		r := make(AV, y.Len())
		for i, yi := range y {
			r[i] = drops(s, yi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return newBV(r)
	default:
		return errType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch y := y.BV.(type) {
	case S:
		return newBV(S(strings.Trim(string(y), string(s))))
	case AS:
		r := make(AS, y.Len())
		for i, yi := range y {
			r[i] = strings.Trim(string(yi), string(s))
		}
		return newBV(r)
	case AV:
		r := make(AV, y.Len())
		for i, yi := range y {
			r[i] = trim(s, yi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return newBV(r)
	default:
		return errType("s^y", "y", y)
	}
}

// take returns i#x.
func take(x, y V) V {
	i := 0
	switch x := x.BV.(type) {
	case I:
		i = int(x)
	case F:
		if !isI(x) {
			return errf("i#y : non-integer i (%s)", x.Type())
		}
		i = int(x)
	default:
		return errf("i#y : non-integer i (%s)", x.Type())
	}
	yv := toArray(y).BV.(array)
	switch {
	case i >= 0:
		if i > yv.Len() {
			return takeCyclic(yv, i)
		}
		return newBV(yv.slice(0, i))
	default:
		if i < -yv.Len() {
			return takeCyclic(yv, i)
		}
		return newBV(yv.slice(yv.Len()+i, yv.Len()))
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
	case AB:
		r := make(AB, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return newBV(r)
	case AF:
		r := make(AF, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return newBV(r)
	case AI:
		r := make(AI, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return newBV(r)
	case AS:
		r := make(AS, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return newBV(r)
	case AV:
		r := make(AV, n)
		for i+step < n {
			copy(r[i:i+step], y)
			i += step
		}
		if neg {
			copy(r[i:n], y[len(y)-n+i:])
		} else {
			copy(r[i:n], y[:n-i])
		}
		return newBV(r)
	default:
		return newBV(y)
	}
}

// ShiftBefore returns x»y. XXX: unused for now.
func shiftBefore(x, y V) V {
	x = toArray(x)
	max := int(minI(I(Length(x)), I(Length(y))))
	if max == 0 {
		return y
	}
	switch y := y.BV.(type) {
	case AB:
		switch x := x.BV.(type) {
		case AB:
			r := make(AB, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(B2F(y[i-max]))
			}
			return newBV(r)
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = int(B2I(y[i-max]))
			}
			return newBV(r)
		default:
			return errType("x»y", "y", y)
		}
	case AF:
		switch x := x.BV.(type) {
		case AB:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(B2F(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		case AI:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(x[i])
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		default:
			return errType("x»y", "y", y)
		}
	case AI:
		switch x := x.BV.(type) {
		case AB:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = int(B2I(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(y[i-max])
			}
			return newBV(r)
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		default:
			return errType("x»y", "y", y)
		}
	case AS:
		switch x := x.BV.(type) {
		case AS:
			r := make(AS, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(r)
		default:
			return errType("x»y", "y", y)
		}
	case AV:
		switch x := x.BV.(type) {
		case array:
			r := make(AV, len(y))
			for i := 0; i < max; i++ {
				r[i] = x.at(i)
			}
			copy(r[max:], y[:len(y)-max])
			return newBV(canonical(r))
		default:
			return errType("x»y", "y", y)
		}
	default:
		return errs("x»y: y not an array")
	}
}

// nudge returns »x. XXX unused for now
func nudge(x V) V {
	switch x := x.BV.(type) {
	case AB:
		r := make(AB, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return newBV(r)
	case AI:
		r := make(AI, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return newBV(r)
	case AF:
		r := make(AF, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return newBV(r)
	case AS:
		r := make(AS, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return newBV(r)
	case AV:
		r := make(AV, x.Len())
		copy(r[1:], x[0:x.Len()-1])
		return newBV(canonical(r))
	default:
		return errs("»x : not an array")
	}
}

// ShiftAfter returns x«y. XXX: unused for now.
func shiftAfter(x, y V) V {
	x = toArray(x)
	max := int(minI(I(Length(x)), I(Length(y))))
	if max == 0 {
		return y
	}
	switch y := y.BV.(type) {
	case AB:
		switch x := x.BV.(type) {
		case AB:
			r := make(AB, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(B2F(y[i]))
			}
			return newBV(r)
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = int(B2I(y[i]))
			}
			return newBV(r)
		default:
			return errType("x«y", "y", y)
		}
	case AF:
		switch x := x.BV.(type) {
		case AB:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(B2F(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		case AI:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(x[i])
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		default:
			return errType("x«y", "y", y)
		}
	case AI:
		switch x := x.BV.(type) {
		case AB:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = int(B2I(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(y[max])
			}
			return newBV(r)
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		default:
			return errType("x«y", "y", y)
		}
	case AS:
		switch x := x.BV.(type) {
		case AS:
			r := make(AS, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(r)
		default:
			return errType("x«y", "y", y)
		}
	case AV:
		switch x := x.BV.(type) {
		case array:
			r := make(AV, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x.at(i)
			}
			copy(r[:len(y)-max], y[max:])
			return newBV(canonical(r))
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
	switch x := x.BV.(type) {
	case AB:
		r := make(AB, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return newBV(r)
	case AI:
		r := make(AI, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return newBV(r)
	case AF:
		r := make(AF, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return newBV(r)
	case AS:
		r := make(AS, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return newBV(r)
	case AV:
		r := make(AV, x.Len())
		copy(r[0:x.Len()-1], x[1:])
		return newBV(canonical(r))
	default:
		return errs("«x : x not an array")
	}
}

// flip returns +x.
func flip(x V) V {
	//assertCanonical(x)
	x = toArray(x)
	switch xv := x.BV.(type) {
	case AV:
		cols := xv.Len()
		if cols == 0 {
			return newBV(AV{x})
		}
		lines := -1
		for _, o := range xv {
			nl := Length(o)
			if _, ok := o.BV.(array); !ok {
				continue
			}
			switch {
			case lines < 0:
				lines = nl
			case nl >= 1 && nl != lines:
				return errf("line length mismatch: %d vs %d", nl, lines)
			}
		}
		t := aType(xv)
		switch {
		case lines <= 0:
			return newBV(AV{x})
		case lines == 1:
			switch t {
			case tB, tAB:
				return newBV(AV{newBV(flipAB(xv))})
			case tF, tAF:
				return newBV(AV{newBV(flipAF(xv))})
			case tI, tAI:
				return newBV(AV{newBV(flipAI(xv))})
			case tS, tAS:
				return newBV(AV{newBV(flipAS(xv))})
			default:
				return newBV(AV{flipAV(xv)})
			}
		default:
			switch t {
			case tB, tAB:
				return newBV(flipAVAB(xv, lines))
			case tF, tAF:
				return newBV(flipAVAF(xv, lines))
			case tI, tAI:
				return newBV(flipAVAI(xv, lines))
			case tS, tAS:
				return newBV(flipAVAS(xv, lines))
			default:
				return newBV(flipAVAV(xv, lines))
			}
		}
	default:
		return newBV(AV{x})
	}
}

func flipAB(x AV) AB {
	r := make(AB, x.Len())
	for i, xi := range x {
		switch z := xi.BV.(type) {
		case I:
			r[i] = z == 1
		case AB:
			r[i] = z[0]
		}
	}
	return r
}

func flipAVAB(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AB, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.BV.(type) {
			case I:
				q[i] = z == 1
			case AB:
				q[i] = z[j]
			}
		}
		r[j] = newBV(q)
	}
	return r
}

func flipAF(x AV) AF {
	r := make(AF, x.Len())
	for i, xi := range x {
		switch z := xi.BV.(type) {
		case AB:
			r[i] = float64(B2F(z[0]))
		case F:
			r[i] = float64(z)
		case AF:
			r[i] = z[0]
		case I:
			r[i] = float64(z)
		case AI:
			r[i] = float64(z[0])
		}
	}
	return r
}

func flipAVAF(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AF, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.BV.(type) {
			case AB:
				q[i] = float64(B2F(z[j]))
			case F:
				q[i] = float64(z)
			case AF:
				q[i] = z[j]
			case I:
				q[i] = float64(z)
			case AI:
				q[i] = float64(z[j])
			}
		}
		r[j] = newBV(q)
	}
	return r
}

func flipAI(x AV) AI {
	r := make(AI, x.Len())
	for i, xi := range x {
		switch z := xi.BV.(type) {
		case AB:
			r[i] = int(B2I(z[0]))
		case I:
			r[i] = int(z)
		case AI:
			r[i] = z[0]
		}
	}
	return r
}

func flipAVAI(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AI, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.BV.(type) {
			case AB:
				q[i] = int(B2I(z[j]))
			case I:
				q[i] = int(z)
			case AI:
				q[i] = z[j]
			}
		}
		r[j] = newBV(q)
	}
	return r
}

func flipAS(x AV) AS {
	r := make(AS, x.Len())
	for i, xi := range x {
		switch z := xi.BV.(type) {
		case S:
			r[i] = string(z)
		case AS:
			r[i] = z[0]
		}
	}
	return r
}

func flipAVAS(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AS, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.BV.(type) {
			case S:
				q[i] = string(z)
			case AS:
				q[i] = z[j]
			}
		}
		r[j] = newBV(q)
	}
	return r
}

func flipAV(x AV) V {
	r := make(AV, x.Len())
	for i, xi := range x {
		switch z := xi.BV.(type) {
		case array:
			r[i] = z.at(0)
		default:
			r[i] = xi
		}
	}
	return newBV(canonical(r))
}

func flipAVAV(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AV, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.BV.(type) {
			case array:
				q[i] = z.at(j)
			default:
				q[i] = xi
			}
		}
		r[j] = newBV(q)
	}
	return r
}

// joinTo returns x,y.
func joinTo(x, y V) V {
	switch xv := x.BV.(type) {
	case F:
		return joinToF(xv, y, true)
	case I:
		return joinToI(xv, y, true)
	case S:
		return joinToS(xv, y, true)
	case AB:
		return joinToAB(y, xv, false)
	case AF:
		return joinToAF(y, xv, false)
	case AI:
		return joinToAI(y, xv, false)
	case AS:
		return joinToAS(y, xv, false)
	case AV:
		return joinToAV(y, xv, false)
	default:
		switch yv := y.BV.(type) {
		case array:
			return newBV(joinAtomToArray(x, yv, true))
		default:
			return newBV(AV{x, y})
		}
	}
}

func joinToI(x I, y V, left bool) V {
	switch yv := y.BV.(type) {
	case F:
		if left {
			return newBV(AF{float64(x), float64(yv)})
		}
		return newBV(AF{float64(yv), float64(x)})
	case I:
		if left {
			return newBV(AI{int(x), int(yv)})
		}
		return newBV(AI{int(yv), int(x)})
	case S:
		if left {
			return newBV(AV{newBV(x), y})
		}
		return newBV(AV{y, newBV(x)})
	case AB:
		return joinToAB(newBV(x), yv, left)
	case AF:
		return joinToAF(newBV(x), yv, left)
	case AI:
		return joinToAI(newBV(x), yv, left)
	case AS:
		return joinToAS(newBV(x), yv, left)
	case AV:
		return joinToAV(newBV(x), yv, left)
	default:
		return newBV(AV{newBV(x), y})
	}
}

func joinToF(x F, y V, left bool) V {
	switch yv := y.BV.(type) {
	case F:
		if left {
			return newBV(AF{float64(x), float64(yv)})
		}
		return newBV(AF{float64(yv), float64(x)})
	case I:
		if left {
			return newBV(AF{float64(x), float64(yv)})
		}
		return newBV(AF{float64(yv), float64(x)})
	case S:
		if left {
			return newBV(AV{newBV(x), y})
		}
		return newBV(AV{y, newBV(x)})
	case AB:
		return joinToAB(newBV(x), yv, left)
	case AF:
		return joinToAF(newBV(x), yv, left)
	case AI:
		return joinToAI(newBV(x), yv, left)
	case AS:
		return joinToAS(newBV(x), yv, left)
	case AV:
		return joinToAV(newBV(x), yv, left)
	default:
		return newBV(AV{newBV(x), y})
	}
}

func joinToS(x S, y V, left bool) V {
	switch yv := y.BV.(type) {
	case F:
		if left {
			return newBV(AV{newBV(x), y})
		}
		return newBV(AV{y, newBV(x)})
	case I:
		if left {
			return newBV(AV{newBV(x), y})
		}
		return newBV(AV{y, newBV(x)})
	case S:
		if left {
			return newBV(AS{string(x), string(yv)})
		}
		return newBV(AS{string(yv), string(x)})
	case AB:
		return joinToAB(newBV(x), yv, left)
	case AF:
		return joinToAF(newBV(x), yv, left)
	case AI:
		return joinToAI(newBV(x), yv, left)
	case AS:
		return joinToAS(newBV(x), yv, left)
	case AV:
		return joinToAV(newBV(x), yv, left)
	default:
		return newBV(AV{newBV(x), y})
	}
}

func joinToAV(x V, y AV, left bool) V {
	switch xv := x.BV.(type) {
	case array:
		if left {
			return newBV(joinArrays(xv, y))
		}
		return newBV(joinArrays(y, xv))
	default:
		r := make(AV, len(y)+1)
		if left {
			r[0] = x
			copy(r[1:], y)
		} else {
			r[len(r)-1] = x
			copy(r[:len(r)-1], y)
		}
		return newBV(r)
	}
}

func joinArrays(x, y array) AV {
	r := make(AV, y.Len()+x.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = x.at(i)
	}
	for i := x.Len(); i < len(r); i++ {
		r[i] = y.at(i - x.Len())
	}
	return r
}

func joinAtomToArray(x V, y array, left bool) AV {
	r := make(AV, y.Len()+1)
	if left {
		r[0] = x
		for i := 1; i < len(r); i++ {
			r[i] = y.at(i - 1)
		}
	} else {
		r[len(r)-1] = x
		for i := 0; i < len(r)-1; i++ {
			r[i] = y.at(i)
		}
	}
	return r
}

func joinToAS(x V, y AS, left bool) V {
	switch xv := x.BV.(type) {
	case S:
		r := make(AS, len(y)+1)
		if left {
			r[0] = string(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = string(xv)
			copy(r[:len(r)-1], y)
		}
		return newBV(r)
	case AS:
		r := make(AS, len(y)+xv.Len())
		if left {
			copy(r[:xv.Len()], xv)
			copy(r[xv.Len():], y)
		} else {
			copy(r[:len(y)], y)
			copy(r[len(y):], xv)
		}
		return newBV(r)
	case array:
		if left {
			return newBV(joinArrays(xv, y))
		}
		return newBV(joinArrays(y, xv))
	default:
		return newBV(joinAtomToArray(x, y, left))
	}
}

func joinToAB(x V, y AB, left bool) V {
	switch xv := x.BV.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			for i := 1; i < len(r); i++ {
				r[i] = float64(B2F(y[i-1]))
			}
		} else {
			r[len(r)-1] = float64(xv)
			for i := 0; i < len(r); i++ {
				r[i] = float64(B2F(y[i]))
			}
		}
		return newBV(r)
	case I:
		if isBI(xv) {
			r := make(AB, len(y)+1)
			if left {
				r[0] = xv == 1
				copy(r[1:], y)
			} else {
				r[len(r)-1] = xv == 1
				copy(r[:len(r)-1], y)
			}
			return newBV(r)
		}
		r := make(AI, len(y)+1)
		if left {
			r[0] = int(xv)
			for i := 1; i < len(r); i++ {
				r[i] = int(B2I(y[i-1]))
			}
		} else {
			r[len(r)-1] = int(xv)
			for i := 0; i < len(r); i++ {
				r[i] = int(B2I(y[i]))
			}
		}
		return newBV(r)
	case AB:
		if left {
			return newBV(joinABAB(xv, y))
		}
		return newBV(joinABAB(y, xv))
	case AI:
		if left {
			return newBV(joinAIAB(xv, y))
		}
		return newBV(joinABAI(y, xv))
	case AF:
		if left {
			return newBV(joinAFAB(xv, y))
		}
		return newBV(joinABAF(y, xv))
	case array:
		if left {
			return newBV(joinArrays(xv, y))
		}
		return newBV(joinArrays(y, xv))
	default:
		return newBV(joinAtomToArray(x, y, left))
	}
}

func joinToAI(x V, y AI, left bool) V {
	switch xv := x.BV.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			for i := 1; i < len(r); i++ {
				r[i] = float64(y[i-1])
			}
		} else {
			r[len(r)-1] = float64(xv)
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(y[i])
			}
		}
		return newBV(r)
	case I:
		r := make(AI, len(y)+1)
		if left {
			r[0] = int(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = int(xv)
			copy(r[:len(r)-1], y)
		}
		return newBV(r)
	case AB:
		if left {
			return newBV(joinABAI(xv, y))
		}
		return newBV(joinAIAB(y, xv))
	case AI:
		if left {
			return newBV(joinAIAI(xv, y))
		}
		return newBV(joinAIAI(y, xv))
	case AF:
		if left {
			return newBV(joinAFAI(xv, y))
		}
		return newBV(joinAIAF(y, xv))
	case array:
		if left {
			return newBV(joinArrays(xv, y))
		}
		return newBV(joinArrays(y, xv))
	default:
		return newBV(joinAtomToArray(x, y, left))
	}
}

func joinToAF(x V, y AF, left bool) V {
	switch xv := x.BV.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(xv)
			copy(r[:len(r)-1], y)
		}
		return newBV(r)
	case I:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(xv)
			copy(r[:len(r)-1], y)
		}
		return newBV(r)
	case AB:
		if left {
			return newBV(joinABAF(xv, y))
		}
		return newBV(joinAFAB(y, xv))
	case AI:
		if left {
			return newBV(joinAIAF(xv, y))
		}
		return newBV(joinAFAI(y, xv))
	case AF:
		if left {
			return newBV(joinAFAF(xv, y))
		}
		return newBV(joinAFAF(y, xv))
	case array:
		if left {
			return newBV(joinArrays(xv, y))
		}
		return newBV(joinArrays(y, xv))
	default:
		return newBV(joinAtomToArray(x, y, left))
	}
}

func joinABAB(x AB, y AB) AB {
	r := make(AB, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAIAI(x AI, y AI) AI {
	r := make(AI, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAFAF(x AF, y AF) AF {
	r := make(AF, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinABAI(x AB, y AI) AI {
	r := make(AI, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = int(B2I(x[i]))
	}
	copy(r[len(x):], y)
	return r
}

func joinAIAB(x AI, y AB) AI {
	r := make(AI, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = int(B2I(y[i-len(x)]))
	}
	return r
}

func joinABAF(x AB, y AF) AF {
	r := make(AF, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = float64(B2F(x[i]))
	}
	copy(r[len(x):], y)
	return r
}

func joinAFAB(x AF, y AB) AF {
	r := make(AF, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = float64(B2F(y[i-len(x)]))
	}
	return r
}

func joinAIAF(x AI, y AF) AF {
	r := make(AF, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = float64(x[i])
	}
	copy(r[len(x):], y)
	return r
}

func joinAFAI(x AF, y AI) AF {
	r := make(AF, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = float64(y[i-len(x)])
	}
	return r
}

// enlist returns ,x.
func enlist(x V) V {
	switch xv := x.BV.(type) {
	case F:
		return newBV(AF{float64(xv)})
	case I:
		if isBI(xv) {
			return newBV(AB{xv == 1})
		}
		return newBV(AI{int(xv)})
	case S:
		return newBV(AS{string(xv)})
	default:
		return newBV(AV{x})
	}
}

// windows returns i^y.
func windows(i int, y V) V {
	switch y := y.BV.(type) {
	case array:
		if i <= 0 || i >= y.Len()+1 {
			return errf("i^y : i out of range !%d (%d)", y.Len()+1, i)
		}
		r := make(AV, 1+y.Len()-i)
		for j := range r {
			r[j] = newBV(y.slice(j, j+i))
		}
		return newBV(canonical(r))
	default:
		return errs("i^y : y not an array")
	}
}

func sumAB(x AB) int {
	n := 0
	for _, xi := range x {
		if xi {
			n++
		}
	}
	return n
}

// group returns =x.
func group(x V) V {
	if Length(x) == 0 {
		return newBV(AV{})
	}
	switch x := x.BV.(type) {
	case AB:
		n := sumAB(x)
		r := make(AV, int(B2I(n > 0)+1))
		ai := make(AI, x.Len())
		if n == 0 {
			for i := range ai {
				ai[i] = i
			}
			r[0] = newBV(ai)
			return newBV(r)
		}
		aif := ai[:len(ai)-n]
		ait := ai[len(ai)-n:]
		iTrue, iFalse := 0, 0
		for i, xi := range x {
			if xi {
				ait[iTrue] = i
				iTrue++
			} else {
				aif[iFalse] = i
				iFalse++
			}
		}
		r[0] = newBV(aif)
		r[1] = newBV(ait)
		return newBV(r)
	case AI:
		max := maxAI(x)
		if max < 0 {
			max = -1
		}
		r := make(AV, max+1)
		counta := make(AI, 2*(max+1))
		counts := counta[:max+1]
		countn := 0
		for _, j := range x {
			if j < 0 {
				countn++
				continue
			}
			counts[j]++
		}
		scounts := counta[max+1:]
		sn := 0
		for i, n := range counts {
			sn += n
			scounts[i] = sn
		}
		pj := 0
		ai := make(AI, x.Len()-countn)
		for i := range r {
			r[i] = newBV(ai[pj:scounts[i]])
			pj = scounts[i]
		}
		for i, j := range x {
			if j < 0 {
				continue
			}
			ai[scounts[j]-counts[j]] = i
			counts[j]--
		}
		return newBV(r)
	case AF:
		z := toAI(x)
		if isErr(z) {
			return z
		}
		return group(z)
	case AV:
		//assertCanonical(x)
		return errs("=x : x non-integer array")
	default:
		return errs("=x : x not an integer array")
	}
}

// icount efficiently returns #'=x.
func icount(x V) V {
	if Length(x) == 0 {
		return newBV(AI{})
	}
	switch x := x.BV.(type) {
	case AB:
		n := sumAB(x)
		return newBV(AI{x.Len() - n, n})
	case AI:
		max := maxAI(x)
		if max < 0 {
			max = -1
		}
		counts := make(AI, max+1)
		for _, j := range x {
			if j >= 0 {
				counts[j]++
			}
		}
		return newBV(counts)
	case AF:
		z := toAI(x)
		if isErr(z) {
			return z
		}
		return icount(z)
	case AV:
		//assertCanonical(x)
		return errs("icount x : x non-integer array")
	default:
		return errf("icount x : x not an integer array (%s)", x.Type())
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if Length(x) != Length(y) {
		return errf("f=y : length mismatch for f[y] and y: %d vs %d ",
			Length(x), Length(y))
	}
	x = group(x)
	if isErr(x) {
		return errs("f=y : f[y] not an integer array")
	}
	avx := x.BV.(AV) // group should always return AV or errV
	switch y := y.BV.(type) {
	case array:
		r := make(AV, avx.Len())
		for i, xi := range avx {
			r[i] = y.atIndices(xi.BV.(AI))
		}
		return newBV(r)
	default:
		return errs("f=y : y not array")
	}
}
