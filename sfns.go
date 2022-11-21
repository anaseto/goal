// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
	"strings"
)

// Length returns the length of a value like in #x.
func Length(x V) int {
	switch x := x.(type) {
	case array:
		return x.Len()
	default:
		return 1
	}
}

func reverseMut(x V) {
	switch x := x.(type) {
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
	switch x := x.(type) {
	case array:
		r := cloneShallow(x)
		reverseMut(r)
		return r
	default:
		return errType("|x", "x", x)
	}
}

// Rotate returns f|y.
func rotate(x, y V) V {
	i := 0
	switch x := x.(type) {
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
	switch y := y.(type) {
	case AB:
		r := make(AB, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return r
	case AF:
		r := make(AF, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return r
	case AI:
		r := make(AI, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return r
	case AS:
		r := make(AS, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return r
	case AV:
		r := make(AV, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = y[(j+i)%lenx]
		}
		return r
	default:
		return errType("f|y", "y", y)
	}
}

// first returns *x.
func first(x V) V {
	switch x := x.(type) {
	case array:
		if x.Len() == 0 {
			switch x.(type) {
			case AB:
				return I(0)
			case AF:
				return F(0)
			case AI:
				return I(0)
			case AS:
				return S("")
			default:
				return V(nil)
			}
		}
		return x.at(0)
	default:
		return x
	}
}

// drop returns i_x and s_x.
func drop(x, y V) V {
	switch x := x.(type) {
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
		if err, ok := z.(errV); ok {
			return err
		}
		return drop(z, y)
	case AV:
		z := canonical(x)
		if _, ok := z.(AV); ok {
			return errs("x_y : x non-integer")
		}
		return drop(z, y)
	default:
		return errf("x_y : bad type i (%s)", x.Type())
	}
}

func dropi(i int, y V) V {
	switch y := y.(type) {
	case array:
		switch {
		case i >= 0:
			if i > y.Len() {
				i = y.Len()
			}
			return canonical(y.slice(i, y.Len()))
		default:
			i = y.Len() + i
			if i < 0 {
				i = 0
			}
			return canonical(y.slice(0, i))
		}
	default:
		return errs("i_y : y not an array")
	}
}

func cutAI(x AI, y V) V {
	if !sort.IsSorted(sort.IntSlice(x)) {
		return errs("x^y : x is not ascending")
	}
	ylen := Length(y)
	for _, i := range x {
		if i < 0 || i > ylen {
			return errf("x^y : x contains out of bound index (%d)", i)
		}
	}
	if len(x) == 0 {
		return AV{}
	}
	switch y := y.(type) {
	case AB:
		res := make(AV, len(x))
		for i, from := range x {
			to := len(y)
			if i+1 < len(x) {
				to = x[i+1]
			}
			res[i] = y[from:to]
		}
		return canonical(res)
	case AI:
		res := make(AV, len(x))
		for i, from := range x {
			to := len(y)
			if i+1 < len(x) {
				to = x[i+1]
			}
			res[i] = y[from:to]
		}
		return canonical(res)
	case AF:
		res := make(AV, len(x))
		for i, from := range x {
			to := len(y)
			if i+1 < len(x) {
				to = x[i+1]
			}
			res[i] = y[from:to]
		}
		return canonical(res)
	case AS:
		res := make(AV, len(x))
		for i, from := range x {
			to := len(y)
			if i+1 < len(x) {
				to = x[i+1]
			}
			res[i] = y[from:to]
		}
		return res
	case AV:
		res := make(AV, len(x))
		for i, from := range x {
			to := len(y)
			if i+1 < len(x) {
				to = x[i+1]
			}
			res[i] = y[from:to]
		}
		return canonical(res)
	default:
		return errs("x^y : y not an array")
	}
}

func drops(s S, y V) V {
	switch y := y.(type) {
	case S:
		return S(strings.TrimPrefix(string(y), string(s)))
	case AS:
		res := make(AS, y.Len())
		for i, z := range y {
			res[i] = strings.TrimPrefix(string(z), string(s))
		}
		return res
	case AV:
		res := make(AV, y.Len())
		for i, z := range y {
			res[i] = drops(s, z)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return res
	default:
		return errType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch y := y.(type) {
	case S:
		return S(strings.Trim(string(y), string(s)))
	case AS:
		res := make(AS, y.Len())
		for i, z := range y {
			res[i] = strings.Trim(string(z), string(s))
		}
		return res
	case AV:
		res := make(AV, y.Len())
		for i, z := range y {
			res[i] = trim(s, z)
			if err, ok := res[i].(errV); ok {
				return err
			}
		}
		return res
	default:
		return errType("s^y", "y", y)
	}
}

// take returns i#x.
func take(x, y V) V {
	i := 0
	switch x := x.(type) {
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
	y = toArray(y)
	switch y := y.(type) {
	case array:
		switch {
		case i >= 0:
			if i > y.Len() {
				return takeCyclic(y, i)
			}
			return y.slice(0, i)
		default:
			if i < -y.Len() {
				return takeCyclic(y, i)
			}
			return y.slice(y.Len()+i, y.Len())
		}
	default:
		return y
	}
}

func takeCyclic(y V, n int) V {
	neg := n < 0
	if neg {
		n = -n
	}
	i := 0
	step := Length(y)
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
		return r
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
		return r
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
		return r
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
		return r
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
		return r
	default:
		return y
	}
}

// ShiftBefore returns x»y. XXX: unused for now.
func shiftBefore(x, y V) V {
	x = toArray(x)
	max := int(minI(I(Length(x)), I(Length(y))))
	if max == 0 {
		return y
	}
	switch y := y.(type) {
	case AB:
		switch x := x.(type) {
		case AB:
			r := make(AB, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(B2F(y[i-max]))
			}
			return r
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = int(B2I(y[i-max]))
			}
			return r
		default:
			return errType("x»y", "y", y)
		}
	case AF:
		switch x := x.(type) {
		case AB:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(B2F(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return r
		case AI:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = float64(x[i])
			}
			copy(r[max:], y[:len(y)-max])
			return r
		default:
			return errType("x»y", "y", y)
		}
	case AI:
		switch x := x.(type) {
		case AB:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = int(B2I(x[i]))
			}
			copy(r[max:], y[:len(y)-max])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i] = float64(y[i-max])
			}
			return r
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return r
		default:
			return errType("x»y", "y", y)
		}
	case AS:
		switch x := x.(type) {
		case AS:
			r := make(AS, len(y))
			for i := 0; i < max; i++ {
				r[i] = x[i]
			}
			copy(r[max:], y[:len(y)-max])
			return r
		default:
			return errType("x»y", "y", y)
		}
	case AV:
		switch x := x.(type) {
		case array:
			r := make(AV, len(y))
			for i := 0; i < max; i++ {
				r[i] = x.at(i)
			}
			copy(r[max:], y[:len(y)-max])
			return canonical(r)
		default:
			return errType("x»y", "y", y)
		}
	default:
		return errs("x»y: y not an array")
	}
}

// nudge returns »x.
func nudge(x V) V {
	switch x := x.(type) {
	case AB:
		r := make(AB, len(x))
		copy(r[1:], x[0:len(x)-1])
		return r
	case AI:
		r := make(AI, len(x))
		copy(r[1:], x[0:len(x)-1])
		return r
	case AF:
		r := make(AF, len(x))
		copy(r[1:], x[0:len(x)-1])
		return r
	case AS:
		r := make(AS, len(x))
		copy(r[1:], x[0:len(x)-1])
		return r
	case AV:
		r := make(AV, len(x))
		copy(r[1:], x[0:len(x)-1])
		return canonical(r)
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
	switch y := y.(type) {
	case AB:
		switch x := x.(type) {
		case AB:
			r := make(AB, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(B2F(y[i]))
			}
			return r
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = int(B2I(y[i]))
			}
			return r
		default:
			return errType("x«y", "y", y)
		}
	case AF:
		switch x := x.(type) {
		case AB:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(B2F(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return r
		case AI:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = float64(x[i])
			}
			copy(r[:len(y)-max], y[max:])
			return r
		default:
			return errType("x«y", "y", y)
		}
	case AI:
		switch x := x.(type) {
		case AB:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = int(B2I(x[i]))
			}
			copy(r[:len(y)-max], y[max:])
			return r
		case AF:
			r := make(AF, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			for i := max; i < len(y); i++ {
				r[i-max] = float64(y[max])
			}
			return r
		case AI:
			r := make(AI, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return r
		default:
			return errType("x«y", "y", y)
		}
	case AS:
		switch x := x.(type) {
		case AS:
			r := make(AS, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x[i]
			}
			copy(r[:len(y)-max], y[max:])
			return r
		default:
			return errType("x«y", "y", y)
		}
	case AV:
		switch x := x.(type) {
		case array:
			r := make(AV, len(y))
			for i := 0; i < max; i++ {
				r[len(y)-1-i] = x.at(i)
			}
			copy(r[:len(y)-max], y[max:])
			return canonical(r)
		default:
			return errType("x«y", "y", y)
		}
	default:
		return errs("x«y: y not an array")
	}
}

// NudgeBack returns «x.
func nudgeBack(x V) V {
	if Length(x) == 0 {
		return x
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, len(x))
		copy(r[0:len(x)-1], x[1:])
		return r
	case AI:
		r := make(AI, len(x))
		copy(r[0:len(x)-1], x[1:])
		return r
	case AF:
		r := make(AF, len(x))
		copy(r[0:len(x)-1], x[1:])
		return r
	case AS:
		r := make(AS, len(x))
		copy(r[0:len(x)-1], x[1:])
		return r
	case AV:
		r := make(AV, len(x))
		copy(r[0:len(x)-1], x[1:])
		return canonical(r)
	default:
		return errs("«x : x not an array")
	}
}

// flip returns +x.
func flip(x V) V {
	x = toArray(x)
	x = canonical(x) // XXX really?
	switch x := x.(type) {
	case AV:
		cols := len(x)
		if cols == 0 {
			// (+⟨⟩) ≡ ⋈⟨⟩
			return AV{x}
		}
		lines := -1
		for _, o := range x {
			nl := Length(o)
			if !isArray(o) {
				continue
			}
			switch {
			case lines < 0:
				lines = nl
			case nl >= 1 && nl != lines:
				return errf("line length mismatch: %d vs %d", nl, lines)
			}
		}
		t := aType(x)
		switch {
		case lines <= 0:
			// (+⟨⟨⟩,…⟩) ≡ ⟨⟩
			// TODO: error if atoms?
			return x[0]
		case lines == 1:
			switch t {
			case tB, tAB:
				return AV{flipAB(x)}
			case tF, tAF:
				return AV{flipAF(x)}
			case tI, tAI:
				return AV{flipAI(x)}
			case tS, tAS:
				return AV{flipAS(x)}
			default:
				return AV{flipAV(x)}
			}
		default:
			switch t {
			case tB, tAB:
				return flipAVAB(x, lines)
			case tF, tAF:
				return flipAVAF(x, lines)
			case tI, tAI:
				return flipAVAI(x, lines)
			case tS, tAS:
				return flipAVAS(x, lines)
			default:
				return flipAVAV(x, lines)
			}
		}
	default:
		return AV{x}
	}
}

func flipAB(x AV) AB {
	r := make(AB, len(x))
	for i, z := range x {
		switch z := z.(type) {
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
	a := make(AB, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, z := range x {
			switch z := z.(type) {
			case I:
				q[i] = z == 1
			case AB:
				q[i] = z[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAF(x AV) AF {
	r := make(AF, len(x))
	for i, z := range x {
		switch z := z.(type) {
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
	a := make(AF, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, z := range x {
			switch z := z.(type) {
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
		r[j] = q
	}
	return r
}

func flipAI(x AV) AI {
	r := make(AI, len(x))
	for i, z := range x {
		switch z := z.(type) {
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
	a := make(AI, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, z := range x {
			switch z := z.(type) {
			case AB:
				q[i] = int(B2I(z[j]))
			case I:
				q[i] = int(z)
			case AI:
				q[i] = z[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAS(x AV) AS {
	r := make(AS, len(x))
	for i, z := range x {
		switch z := z.(type) {
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
	a := make(AS, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, z := range x {
			switch z := z.(type) {
			case S:
				q[i] = string(z)
			case AS:
				q[i] = z[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAV(x AV) AV {
	r := make(AV, len(x))
	for i, z := range x {
		switch z := z.(type) {
		case array:
			r[i] = z.at(0)
		default:
			r[i] = z
		}
	}
	return r
}

func flipAVAV(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AV, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, z := range x {
			switch z := z.(type) {
			case array:
				q[i] = z.at(j)
			default:
				q[i] = z
			}
		}
		r[j] = q
	}
	return r
}

// joinTo returns x,y.
func joinTo(x, y V) V {
	switch x := x.(type) {
	case F:
		return joinToF(x, y, true)
	case I:
		return joinToI(x, y, true)
	case S:
		return joinToS(x, y, true)
	case AB:
		return joinToAB(y, x, false)
	case AF:
		return joinToAF(y, x, false)
	case AI:
		return joinToAI(y, x, false)
	case AS:
		return joinToAS(y, x, false)
	case AV:
		return joinToAV(y, x, false)
	default:
		switch y := y.(type) {
		case array:
			return joinAtomToArray(x, y, true)
		default:
			return AV{x, y}
		}
	}
}

func joinToI(x I, y V, left bool) V {
	switch y := y.(type) {
	case F:
		if left {
			return AF{float64(x), float64(y)}
		}
		return AF{float64(y), float64(x)}
	case I:
		if left {
			return AI{int(x), int(y)}
		}
		return AI{int(y), int(x)}
	case S:
		if left {
			return AV{x, y}
		}
		return AV{y, x}
	case AB:
		return joinToAB(x, y, left)
	case AF:
		return joinToAF(x, y, left)
	case AI:
		return joinToAI(x, y, left)
	case AS:
		return joinToAS(x, y, left)
	case AV:
		return joinToAV(x, y, left)
	default:
		return AV{x, y}
	}
}

func joinToF(x F, y V, left bool) V {
	switch y := y.(type) {
	case F:
		if left {
			return AF{float64(x), float64(y)}
		}
		return AF{float64(y), float64(x)}
	case I:
		if left {
			return AF{float64(x), float64(y)}
		}
		return AF{float64(y), float64(x)}
	case S:
		if left {
			return AV{x, y}
		}
		return AV{y, x}
	case AB:
		return joinToAB(x, y, left)
	case AF:
		return joinToAF(x, y, left)
	case AI:
		return joinToAI(x, y, left)
	case AS:
		return joinToAS(x, y, left)
	case AV:
		return joinToAV(x, y, left)
	default:
		return AV{x, y}
	}
}

func joinToS(x S, y V, left bool) V {
	switch y := y.(type) {
	case F:
		if left {
			return AV{x, y}
		}
		return AV{y, x}
	case I:
		if left {
			return AV{x, y}
		}
		return AV{y, x}
	case S:
		if left {
			return AS{string(x), string(y)}
		}
		return AS{string(y), string(x)}
	case AB:
		return joinToAB(x, y, left)
	case AF:
		return joinToAF(x, y, left)
	case AI:
		return joinToAI(x, y, left)
	case AS:
		return joinToAS(x, y, left)
	case AV:
		return joinToAV(x, y, left)
	default:
		return AV{x, y}
	}
}

func joinToAV(x V, y AV, left bool) V {
	switch x := x.(type) {
	case array:
		if left {
			return joinArrays(x, y)
		}
		return joinArrays(y, x)
	default:
		r := make(AV, len(y)+1)
		if left {
			r[0] = x
			copy(r[1:], y)
		} else {
			r[len(r)-1] = x
			copy(r[:len(r)-1], y)
		}
		return r
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
	switch x := x.(type) {
	case S:
		r := make(AS, len(y)+1)
		if left {
			r[0] = string(x)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = string(x)
			copy(r[:len(r)-1], y)
		}
		return r
	case AS:
		r := make(AS, len(y)+len(x))
		if left {
			copy(r[:len(x)], x)
			copy(r[len(x):], y)
		} else {
			copy(r[:len(y)], y)
			copy(r[len(y):], x)
		}
		return r
	case array:
		if left {
			return joinArrays(x, y)
		}
		return joinArrays(y, x)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAB(x V, y AB, left bool) V {
	switch x := x.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(x)
			for i := 1; i < len(r); i++ {
				r[i] = float64(B2F(y[i-1]))
			}
		} else {
			r[len(r)-1] = float64(x)
			for i := 0; i < len(r); i++ {
				r[i] = float64(B2F(y[i]))
			}
		}
		return r
	case I:
		if isBI(x) {
			r := make(AB, len(y)+1)
			if left {
				r[0] = x == 1
				copy(r[1:], y)
			} else {
				r[len(r)-1] = x == 1
				copy(r[:len(r)-1], y)
			}
			return r
		}
		r := make(AI, len(y)+1)
		if left {
			r[0] = int(x)
			for i := 1; i < len(r); i++ {
				r[i] = int(B2I(y[i-1]))
			}
		} else {
			r[len(r)-1] = int(x)
			for i := 0; i < len(r); i++ {
				r[i] = int(B2I(y[i]))
			}
		}
		return r
	case AB:
		if left {
			return joinABAB(x, y)
		}
		return joinABAB(y, x)
	case AI:
		if left {
			return joinAIAB(x, y)
		}
		return joinABAI(y, x)
	case AF:
		if left {
			return joinAFAB(x, y)
		}
		return joinABAF(y, x)
	case array:
		if left {
			return joinArrays(x, y)
		}
		return joinArrays(y, x)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAI(x V, y AI, left bool) V {
	switch x := x.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(x)
			for i := 1; i < len(r); i++ {
				r[i] = float64(y[i-1])
			}
		} else {
			r[len(r)-1] = float64(x)
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(y[i])
			}
		}
		return r
	case I:
		r := make(AI, len(y)+1)
		if left {
			r[0] = int(x)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = int(x)
			copy(r[:len(r)-1], y)
		}
		return r
	case AB:
		if left {
			return joinABAI(x, y)
		}
		return joinAIAB(y, x)
	case AI:
		if left {
			return joinAIAI(x, y)
		}
		return joinAIAI(y, x)
	case AF:
		if left {
			return joinAFAI(x, y)
		}
		return joinAIAF(y, x)
	case array:
		if left {
			return joinArrays(x, y)
		}
		return joinArrays(y, x)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAF(x V, y AF, left bool) V {
	switch x := x.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(x)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(x)
			copy(r[:len(r)-1], y)
		}
		return r
	case I:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(x)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(x)
			copy(r[:len(r)-1], y)
		}
		return r
	case AB:
		if left {
			return joinABAF(x, y)
		}
		return joinAFAB(y, x)
	case AI:
		if left {
			return joinAIAF(x, y)
		}
		return joinAFAI(y, x)
	case AF:
		if left {
			return joinAFAF(x, y)
		}
		return joinAFAF(y, x)
	case array:
		if left {
			return joinArrays(x, y)
		}
		return joinArrays(y, x)
	default:
		return joinAtomToArray(x, y, left)
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
	switch x := x.(type) {
	case F:
		return AF{float64(x)}
	case I:
		if isBI(x) {
			return AB{x == 1}
		}
		return AI{int(x)}
	case S:
		return AS{string(x)}
	default:
		return AV{x}
	}
}

// windows returns i^y.
func windows(i int, y V) V {
	switch y := y.(type) {
	case array:
		if i <= 0 || i >= y.Len()+1 {
			return errf("i^y : i out of range !%d (%d)", y.Len()+1, i)
		}
		r := make(AV, 1+y.Len()-i)
		for j := range r {
			r[j] = y.slice(j, j+i)
		}
		return r
	default:
		return errs("i^y : y not an array")
	}
}

func sumAB(x AB) int {
	n := 0
	for _, v := range x {
		if v {
			n++
		}
	}
	return n
}

// group returns =x.
func group(x V) V {
	if Length(x) == 0 {
		return AV{}
	}
	switch x := x.(type) {
	case AB:
		n := sumAB(x)
		r := make(AV, int(B2I(n > 0)+1))
		ai := make(AI, x.Len())
		if n == 0 {
			for i := range ai {
				ai[i] = i
			}
			r[0] = ai
			return r
		}
		aif := ai[:len(ai)-n]
		ait := ai[len(ai)-n:]
		iTrue, iFalse := 0, 0
		for i, v := range x {
			if v {
				ait[iTrue] = i
				iTrue++
			} else {
				aif[iFalse] = i
				iFalse++
			}
		}
		r[0] = aif
		r[1] = ait
		return r
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
			r[i] = ai[pj:scounts[i]]
			pj = scounts[i]
		}
		for i, j := range x {
			if j < 0 {
				continue
			}
			ai[scounts[j]-counts[j]] = i
			counts[j]--
		}
		return r
	case AF:
		z := toAI(x)
		if err, ok := z.(errV); ok {
			return err
		}
		return group(z)
	case AV:
		z := canonical(x)
		if _, ok := z.(AV); ok {
			return errs("=x : x non-integer array")
		}
		return group(z)
	default:
		return errs("=x : x not an integer array")
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if Length(x) != Length(y) {
		return errf("f=y : length mismatch for f[y] and y: %d vs %d ",
			Length(x), Length(y))
	}
	x = group(x)
	if _, ok := x.(errV); ok {
		return errs("f=y : f[y] not an integer array")
	}
	ax := x.(AV) // group should always return AV or errV
	switch y := y.(type) {
	case array:
		r := make(AV, ax.Len())
		for i, v := range ax {
			r[i] = y.atIndices(v.(AI))
		}
		return r
	default:
		return errs("f=y : y not array")
	}
}
