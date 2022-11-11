// structural functions (Length, Reverse, Take, ...)

package goal

import (
	"sort"
)

// length returns #x.
func length(x V) I {
	switch x := x.(type) {
	case nil:
		return 0
	default:
		return I(x.Len())
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
	case Array:
		r := cloneShallow(x)
		reverseMut(r)
		return r
	default:
		return errType(x)
	}
}

// Rotate returns TODO
func Rotate(w, x V) V {
	i := 0
	switch w := w.(type) {
	case I:
		i = int(w)
	case F:
		if !isI(w) {
			return errsw("not an integer")
		}
		i = int(w)
	default:
		return errsw("not an integer")
	}
	lenx := int(length(x))
	if lenx == 0 {
		return x
	}
	i %= lenx
	if i < 0 {
		i += lenx
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	case AF:
		r := make(AF, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	case AI:
		r := make(AI, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	case AS:
		r := make(AS, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	case AV:
		r := make(AV, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	default:
		return errType(x)
	}
}

// first returns *x.
func first(x V) V {
	switch x := x.(type) {
	case Array:
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
		return x.At(0)
	default:
		return x
	}
}

// Tail returns 1_x. XXX unused?
func Tail(x V) V {
	x = toArray(x)
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return errs("zero length")
		}
		return x.Slice(1, x.Len())
	default:
		return errType(x)
	}
}

// drop returns i_x.
func drop(w, x V) V {
	i := 0
	switch w := w.(type) {
	case I:
		i = int(w)
	case F:
		if !isI(w) {
			return errsw("not an integer")
		}
		i = int(w)
	default:
		return errsw("not an integer")
	}
	x = toArray(x)
	switch x := x.(type) {
	case Array:
		switch {
		case i >= 0:
			if i > x.Len() {
				i = x.Len()
			}
			return x.Slice(i, x.Len())
		default:
			i = x.Len() + i
			if i < 0 {
				i = 0
			}
			return x.Slice(0, i)
		}
	default:
		return x
	}
}

// take returns i#x.
func take(w, x V) V {
	i := 0
	switch w := w.(type) {
	case I:
		i = int(w)
	case F:
		if !isI(w) {
			return errsw("not an integer")
		}
		i = int(w)
	default:
		return errsw("not an integer")
	}
	x = toArray(x)
	switch x := x.(type) {
	case Array:
		switch {
		case i >= 0:
			if i > x.Len() {
				return takeCyclic(x, i)
			}
			return x.Slice(0, i)
		default:
			if i < -x.Len() {
				return takeCyclic(x, i)
			}
			return x.Slice(x.Len()+i, x.Len())
		}
	default:
		return x
	}
}

func takeCyclic(x V, n int) V {
	neg := n < 0
	if neg {
		n = -n
	}
	i := 0
	step := x.Len()
	switch x := x.(type) {
	case AB:
		r := make(AB, n)
		for i+step < n {
			copy(r[i:i+step], x)
			i += step
		}
		if neg {
			copy(r[i:n], x[len(x)-n+i:])
		} else {
			copy(r[i:n], x[:n-i])
		}
		return r
	case AF:
		r := make(AF, n)
		for i+step < n {
			copy(r[i:i+step], x)
			i += step
		}
		if neg {
			copy(r[i:n], x[len(x)-n+i:])
		} else {
			copy(r[i:n], x[:n-i])
		}
		return r
	case AI:
		r := make(AI, n)
		for i+step < n {
			copy(r[i:i+step], x)
			i += step
		}
		if neg {
			copy(r[i:n], x[len(x)-n+i:])
		} else {
			copy(r[i:n], x[:n-i])
		}
		return r
	case AS:
		r := make(AS, n)
		for i+step < n {
			copy(r[i:i+step], x)
			i += step
		}
		if neg {
			copy(r[i:n], x[len(x)-n+i:])
		} else {
			copy(r[i:n], x[:n-i])
		}
		return r
	case AV:
		r := make(AV, n)
		for i+step < n {
			copy(r[i:i+step], x)
			i += step
		}
		if neg {
			copy(r[i:n], x[len(x)-n+i:])
		} else {
			copy(r[i:n], x[:n-i])
		}
		return r
	default:
		return x
	}
}

// ShiftBefore returns w»x.
func ShiftBefore(w, x V) V {
	w = toArray(w)
	max := int(minI(length(w), length(x)))
	if max == 0 {
		return x
	}
	switch x := x.(type) {
	case AB:
		switch w := w.(type) {
		case AB:
			r := make(AB, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			copy(r[max:], x[:len(x)-max])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = float64(B2F(x[i-max]))
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = int(B2I(x[i-max]))
			}
			return r
		default:
			return errType(x)
		}
	case AF:
		switch w := w.(type) {
		case AB:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = float64(B2F(w[i]))
			}
			copy(r[max:], x[:len(x)-max])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			copy(r[max:], x[:len(x)-max])
			return r
		case AI:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = float64(w[i])
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return errType(x)
		}
	case AI:
		switch w := w.(type) {
		case AB:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = int(B2I(w[i]))
			}
			copy(r[max:], x[:len(x)-max])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = float64(x[i-max])
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return errType(x)
		}
	case AS:
		switch w := w.(type) {
		case AS:
			r := make(AS, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return errType(x)
		}
	case AV:
		switch w := w.(type) {
		case Array:
			r := make(AV, len(x))
			for i := 0; i < max; i++ {
				r[i] = w.At(i)
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return errType(x)
		}
	default:
		return errs("not an array")
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
		return r
	default:
		return errs("not an array")
	}
}

// ShiftAfter returns w«x.
func ShiftAfter(w, x V) V {
	w = toArray(w)
	max := int(minI(length(w), length(x)))
	if max == 0 {
		return x
	}
	switch x := x.(type) {
	case AB:
		switch w := w.(type) {
		case AB:
			r := make(AB, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			copy(r[:len(x)-max], x[max:])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = float64(B2F(x[i]))
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = int(B2I(x[i]))
			}
			return r
		default:
			return errType(x)
		}
	case AF:
		switch w := w.(type) {
		case AB:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = float64(B2F(w[i]))
			}
			copy(r[:len(x)-max], x[max:])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			copy(r[:len(x)-max], x[max:])
			return r
		case AI:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = float64(w[i])
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return errType(x)
		}
	case AI:
		switch w := w.(type) {
		case AB:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = int(B2I(w[i]))
			}
			copy(r[:len(x)-max], x[max:])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = float64(x[max])
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return errType(x)
		}
	case AS:
		switch w := w.(type) {
		case AS:
			r := make(AS, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return errType(x)
		}
	case AV:
		switch w := w.(type) {
		case Array:
			r := make(AV, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w.At(i)
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return errType(x)
		}
	default:
		return errs("not an array")
	}
}

// NudgeBack returns «x.
func NudgeBack(x V) V {
	if length(x) == 0 {
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
		return r
	default:
		return errs("not an array")
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
			nl := int(length(o))
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
				return AV{flipAO(x)}
			}
		default:
			switch t {
			case tB, tAB:
				return flipAOAB(x, lines)
			case tF, tAF:
				return flipAOAF(x, lines)
			case tI, tAI:
				return flipAOAI(x, lines)
			case tS, tAS:
				return flipAOAS(x, lines)
			default:
				return flipAOAO(x, lines)
			}
		}
	default:
		return AV{x}
	}
}

func flipAB(x AV) AB {
	r := make(AB, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case I:
			r[i] = y == 1
		case AB:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAB(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AB, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case I:
				q[i] = y == 1
			case AB:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAF(x AV) AF {
	r := make(AF, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case AB:
			r[i] = float64(B2F(y[0]))
		case F:
			r[i] = float64(y)
		case AF:
			r[i] = y[0]
		case I:
			r[i] = float64(y)
		case AI:
			r[i] = float64(y[0])
		}
	}
	return r
}

func flipAOAF(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AF, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case AB:
				q[i] = float64(B2F(y[j]))
			case F:
				q[i] = float64(y)
			case AF:
				q[i] = y[j]
			case I:
				q[i] = float64(y)
			case AI:
				q[i] = float64(y[j])
			}
		}
		r[j] = q
	}
	return r
}

func flipAI(x AV) AI {
	r := make(AI, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case AB:
			r[i] = int(B2I(y[0]))
		case I:
			r[i] = int(y)
		case AI:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAI(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AI, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case AB:
				q[i] = int(B2I(y[j]))
			case I:
				q[i] = int(y)
			case AI:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAS(x AV) AS {
	r := make(AS, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case S:
			r[i] = string(y)
		case AS:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAS(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AS, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case S:
				q[i] = string(y)
			case AS:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAO(x AV) AV {
	r := make(AV, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case Array:
			r[i] = y.At(0)
		default:
			r[i] = y
		}
	}
	return r
}

func flipAOAO(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AV, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case Array:
				q[i] = y.At(j)
			default:
				q[i] = y
			}
		}
		r[j] = q
	}
	return r
}

// joinTo returns w,x.
func joinTo(w, x V) V {
	switch w := w.(type) {
	case F:
		return joinToF(w, x, true)
	case I:
		return joinToI(w, x, true)
	case S:
		return joinToS(w, x, true)
	case AB:
		return joinToAB(x, w, false)
	case AF:
		return joinToAF(x, w, false)
	case AI:
		return joinToAI(x, w, false)
	case AS:
		return joinToAS(x, w, false)
	case AV:
		return joinToAO(x, w, false)
	default:
		switch x := x.(type) {
		case Array:
			return joinAtomToArray(w, x, true)
		default:
			return AV{w, x}
		}
	}
}

func joinToI(w I, x V, left bool) V {
	switch x := x.(type) {
	case F:
		if left {
			return AF{float64(w), float64(x)}
		}
		return AF{float64(x), float64(w)}
	case I:
		if left {
			return AI{int(w), int(x)}
		}
		return AI{int(x), int(w)}
	case S:
		if left {
			return AV{w, x}
		}
		return AV{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AV:
		return joinToAO(w, x, left)
	default:
		return AV{w, x}
	}
}

func joinToF(w F, x V, left bool) V {
	switch x := x.(type) {
	case F:
		if left {
			return AF{float64(w), float64(x)}
		}
		return AF{float64(x), float64(w)}
	case I:
		if left {
			return AF{float64(w), float64(x)}
		}
		return AF{float64(x), float64(w)}
	case S:
		if left {
			return AV{w, x}
		}
		return AV{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AV:
		return joinToAO(w, x, left)
	default:
		return AV{w, x}
	}
}

func joinToS(w S, x V, left bool) V {
	switch x := x.(type) {
	case F:
		if left {
			return AV{w, x}
		}
		return AV{x, w}
	case I:
		if left {
			return AV{w, x}
		}
		return AV{x, w}
	case S:
		if left {
			return AS{string(w), string(x)}
		}
		return AS{string(x), string(w)}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AV:
		return joinToAO(w, x, left)
	default:
		return AV{w, x}
	}
}

func joinToAO(w V, x AV, left bool) V {
	switch w := w.(type) {
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		r := make(AV, len(x)+1)
		if left {
			r[0] = w
			copy(r[1:], x)
		} else {
			r[len(r)-1] = w
			copy(r[:len(r)-1], x)
		}
		return r
	}
}

func joinArrays(w, x Array) AV {
	r := make(AV, x.Len()+w.Len())
	for i := 0; i < w.Len(); i++ {
		r[i] = w.At(i)
	}
	for i := w.Len(); i < len(r); i++ {
		r[i] = x.At(i - w.Len())
	}
	return r
}

func joinAtomToArray(w V, x Array, left bool) AV {
	r := make(AV, x.Len()+1)
	if left {
		r[0] = w
		for i := 1; i < len(r); i++ {
			r[i] = x.At(i - 1)
		}
	} else {
		r[len(r)-1] = w
		for i := 0; i < len(r)-1; i++ {
			r[i] = x.At(i)
		}
	}
	return r
}

func joinToAS(w V, x AS, left bool) V {
	switch w := w.(type) {
	case S:
		r := make(AS, len(x)+1)
		if left {
			r[0] = string(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = string(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case AS:
		r := make(AS, len(x)+len(w))
		if left {
			copy(r[:len(w)], w)
			copy(r[len(w):], x)
		} else {
			copy(r[:len(x)], x)
			copy(r[len(x):], w)
		}
		return r
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		return joinAtomToArray(w, x, left)
	}
}

func joinToAB(w V, x AB, left bool) V {
	switch w := w.(type) {
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = float64(w)
			for i := 1; i < len(r); i++ {
				r[i] = float64(B2F(x[i-1]))
			}
		} else {
			r[len(r)-1] = float64(w)
			for i := 0; i < len(r); i++ {
				r[i] = float64(B2F(x[i]))
			}
		}
		return r
	case I:
		if isBI(w) {
			r := make(AB, len(x)+1)
			if left {
				r[0] = w == 1
				copy(r[1:], x)
			} else {
				r[len(r)-1] = w == 1
				copy(r[:len(r)-1], x)
			}
			return r
		}
		r := make(AI, len(x)+1)
		if left {
			r[0] = int(w)
			for i := 1; i < len(r); i++ {
				r[i] = int(B2I(x[i-1]))
			}
		} else {
			r[len(r)-1] = int(w)
			for i := 0; i < len(r); i++ {
				r[i] = int(B2I(x[i]))
			}
		}
		return r
	case AB:
		if left {
			return joinABAB(w, x)
		}
		return joinABAB(x, w)
	case AI:
		if left {
			return joinAIAB(w, x)
		}
		return joinABAI(x, w)
	case AF:
		if left {
			return joinAFAB(w, x)
		}
		return joinABAF(x, w)
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		return joinAtomToArray(w, x, left)
	}
}

func joinToAI(w V, x AI, left bool) V {
	switch w := w.(type) {
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = float64(w)
			for i := 1; i < len(r); i++ {
				r[i] = float64(x[i-1])
			}
		} else {
			r[len(r)-1] = float64(w)
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(x[i])
			}
		}
		return r
	case I:
		r := make(AI, len(x)+1)
		if left {
			r[0] = int(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = int(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case AB:
		if left {
			return joinABAI(w, x)
		}
		return joinAIAB(x, w)
	case AI:
		if left {
			return joinAIAI(w, x)
		}
		return joinAIAI(x, w)
	case AF:
		if left {
			return joinAFAI(w, x)
		}
		return joinAIAF(x, w)
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		return joinAtomToArray(w, x, left)
	}
}

func joinToAF(w V, x AF, left bool) V {
	switch w := w.(type) {
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = float64(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = float64(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case I:
		r := make(AF, len(x)+1)
		if left {
			r[0] = float64(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = float64(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case AB:
		if left {
			return joinABAF(w, x)
		}
		return joinAFAB(x, w)
	case AI:
		if left {
			return joinAIAF(w, x)
		}
		return joinAFAI(x, w)
	case AF:
		if left {
			return joinAFAF(w, x)
		}
		return joinAFAF(x, w)
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		return joinAtomToArray(w, x, left)
	}
}

func joinABAB(w AB, x AB) AB {
	r := make(AB, len(x)+len(w))
	copy(r[:len(w)], w)
	copy(r[len(w):], x)
	return r
}

func joinAIAI(w AI, x AI) AI {
	r := make(AI, len(x)+len(w))
	copy(r[:len(w)], w)
	copy(r[len(w):], x)
	return r
}

func joinAFAF(w AF, x AF) AF {
	r := make(AF, len(x)+len(w))
	copy(r[:len(w)], w)
	copy(r[len(w):], x)
	return r
}

func joinABAI(w AB, x AI) AI {
	r := make(AI, len(w)+len(x))
	for i := 0; i < len(w); i++ {
		r[i] = int(B2I(w[i]))
	}
	copy(r[len(w):], x)
	return r
}

func joinAIAB(w AI, x AB) AI {
	r := make(AI, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = int(B2I(x[i-len(w)]))
	}
	return r
}

func joinABAF(w AB, x AF) AF {
	r := make(AF, len(w)+len(x))
	for i := 0; i < len(w); i++ {
		r[i] = float64(B2F(w[i]))
	}
	copy(r[len(w):], x)
	return r
}

func joinAFAB(w AF, x AB) AF {
	r := make(AF, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = float64(B2F(x[i-len(w)]))
	}
	return r
}

func joinAIAF(w AI, x AF) AF {
	r := make(AF, len(w)+len(x))
	for i := 0; i < len(w); i++ {
		r[i] = float64(w[i])
	}
	copy(r[len(w):], x)
	return r
}

func joinAFAI(w AF, x AI) AF {
	r := make(AF, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = float64(x[i-len(w)])
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

// Windows returns w↕x.
func Windows(w, x V) V {
	i := 0
	switch w := w.(type) {
	case I:
		i = int(w)
	case F:
		if !isI(w) {
			return errsw("not an integer")
		}
		i = int(w)
	default:
		return errsw("not an integer")
	}
	switch x := x.(type) {
	case Array:
		if i <= 0 || i >= x.Len()+1 {
			return errsw("out of range [0, length]")
		}
		r := make(AV, 1+x.Len()-i)
		for j := range r {
			r[j] = x.Slice(j, j+i)
		}
		return r
	default:
		return errs("not an array")
	}
}

// group returns ⊔x. XXX Classify by default?
func group(x V) V {
	if length(x) == 0 {
		return AV{}
	}
	// TODO: optimize allocations
	switch x := x.(type) {
	case AB:
		_, max := minMaxB(x)
		r := make(AV, max+1)
		for i := range r {
			r[i] = AI{}
		}
		for i, v := range x {
			j := B2I(v)
			rj := r[j].(AI)
			r[j] = append(rj, i)
		}
		return r
	case AI:
		min, max := minMax(x)
		if min < 0 {
			return errs("negative integer")
		}
		r := make(AV, max+1)
		for i := range r {
			r[i] = AI{}
		}
		for i, j := range x {
			rj := r[j].(AI)
			r[j] = append(rj, i)
		}
		return r
		// TODO: AF and AO
	default:
		return errs("non-integer array")
	}
}
