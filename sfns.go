// structural functions (Length, Reverse, Take, ...)

package main

import (
	"sort"
)

// Length returns ≠x.
func Length(x O) I {
	switch x := x.(type) {
	case nil:
		return 0
	case Array:
		return x.Len()
	default:
		return 1
	}
}

func reverse(x O) {
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
	case AO:
		for i := 0; i < len(x)/2; i++ {
			x[i], x[len(x)-i-1] = x[len(x)-i-1], x[i]
		}
	case sort.Interface:
		sort.Reverse(x)
	}
}

// Reverse returns ⌽x.
func Reverse(x O) O {
	switch x := x.(type) {
	case Array:
		r := cloneShallow(x)
		reverse(r)
		return r
	default:
		return badtype("⌽")
	}
}

// Rotate returns w⌽x.
func Rotate(w, x O) O {
	i := 0
	switch w := w.(type) {
	case B:
		i = B2I(w)
	case I:
		i = w
	case F:
		i = I(w)
	default:
		// TODO: improve error messages
		return badtype("w⌽")
	}
	lenx := Length(x)
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
	case AO:
		r := make(AO, lenx)
		for j := 0; j < lenx; j++ {
			r[j] = x[(j+i)%lenx]
		}
		return r
	default:
		return badtype("⌽x")
	}
}

// First returns ↑x.
func First(x O) O {
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			switch x.(type) {
			case AB:
				return false
			case AF:
				return F(0)
			case AI:
				return I(0)
			case AS:
				return S("")
			default:
				return O(nil)
			}
		}
		return x.At(0)
	default:
		return x
	}
}

// Tail returns ↓x.
func Tail(x O) O {
	x = toArray(x)
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return badlen("↓")
		}
		return x.Slice(1, x.Len())
	default:
		return badtype("↓")
	}
}

// Drop returns i_x.
func Drop(w, x O) O {
	i := 0
	switch w := w.(type) {
	case B:
		i = B2I(w)
	case I:
		i = w
	case F:
		i = I(w)
	default:
		// TODO: improve error messages
		return badtype("w↓")
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

// Take returns i#x.
func Take(w, x O) O {
	i := 0
	switch w := w.(type) {
	case B:
		i = B2I(w)
	case I:
		i = w
	case F:
		i = I(w)
	default:
		// TODO: improve error messages
		return badtype("w↑")
	}
	x = toArray(x)
	switch x := x.(type) {
	case Array:
		switch {
		case i >= 0:
			if i > x.Len() {
				return growArray(x, i)
			}
			return x.Slice(0, i)
		default:
			if i < -x.Len() {
				return growArray(x, i)
			}
			return x.Slice(x.Len()+i, x.Len())
		}
	default:
		return x
	}
}

// ShiftBefore returns w»x.
func ShiftBefore(w, x O) O {
	w = toArray(w)
	max := minI(Length(w), Length(x))
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
				r[i] = B2F(B(x[i-max]))
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = B2I(B(x[i-max]))
			}
			return r
		default:
			return badtype("» : type mismatch")
		}
	case AF:
		switch w := w.(type) {
		case AB:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = B2F(B(w[i]))
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
				r[i] = F(w[i])
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return badtype("» : type mismatch")
		}
	case AI:
		switch w := w.(type) {
		case AB:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = B2I(B(w[i]))
			}
			copy(r[max:], x[:len(x)-max])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = F(x[i-max])
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
			return badtype("» : type mismatch")
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
			return badtype("» : type mismatch")
		}
	case AO:
		switch w := w.(type) {
		case Array:
			r := make(AO, len(x))
			for i := 0; i < max; i++ {
				r[i] = w.At(i)
			}
			copy(r[max:], x[:len(x)-max])
			return r
		default:
			return badtype("» : type mismatch")
		}
	default:
		return badtype("» : x must be an array")
	}
}

// Nudge returns »x.
func Nudge(x O) O {
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
	case AO:
		r := make(AO, len(x))
		copy(r[1:], x[0:len(x)-1])
		return r
	default:
		return badtype("» : x must be an array")
	}
}

// ShiftAfter returns w«x.
func ShiftAfter(w, x O) O {
	w = toArray(w)
	max := minI(Length(w), Length(x))
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
				r[i-max] = B2F(B(x[i]))
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = B2I(B(x[i]))
			}
			return r
		default:
			return badtype("« : type mismatch")
		}
	case AF:
		switch w := w.(type) {
		case AB:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = B2F(B(w[i]))
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
				r[len(x)-1-i] = F(w[i])
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return badtype("« : type mismatch")
		}
	case AI:
		switch w := w.(type) {
		case AB:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = B2I(B(w[i]))
			}
			copy(r[:len(x)-max], x[max:])
			return r
		case AF:
			r := make(AF, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = F(x[max])
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
			return badtype("« : type mismatch")
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
			return badtype("« : type mismatch")
		}
	case AO:
		switch w := w.(type) {
		case Array:
			r := make(AO, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w.At(i)
			}
			copy(r[:len(x)-max], x[max:])
			return r
		default:
			return badtype("« : type mismatch")
		}
	default:
		return badtype("« : x must be an array")
	}
}

// NudgeBack returns «x.
func NudgeBack(x O) O {
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
	case AO:
		r := make(AO, len(x))
		copy(r[0:len(x)-1], x[1:])
		return r
	default:
		return badtype("« : x must be an array")
	}
}

// Flip returns +x.
func Flip(x O) O {
	x = toArray(x)
	x = canonical(x) // XXX really?
	switch x := x.(type) {
	case AO:
		cols := len(x)
		if cols == 0 {
			// (+⟨⟩) ≡ ⋈⟨⟩
			return AO{x}
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
				return badlen("+")
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
				return AO{flipAB(x)}
			case tF, tAF:
				return AO{flipAF(x)}
			case tI, tAI:
				return AO{flipAI(x)}
			case tS, tAS:
				return AO{flipAS(x)}
			default:
				return AO{flipAO(x)}
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
		return AO{x}
	}
}

func flipAB(x AO) AB {
	r := make(AB, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case B:
			r[i] = bool(y)
		case AB:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAB(x AO, lines int) AO {
	r := make(AO, lines)
	a := make(AB, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case B:
				q[i] = bool(y)
			case AB:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAF(x AO) AF {
	r := make(AF, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case B:
			r[i] = B2F(y)
		case AB:
			r[i] = B2F(B(y[0]))
		case F:
			r[i] = y
		case AF:
			r[i] = y[0]
		case I:
			r[i] = F(y)
		case AI:
			r[i] = F(y[0])
		}
	}
	return r
}

func flipAOAF(x AO, lines int) AO {
	r := make(AO, lines)
	a := make(AF, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case B:
				q[i] = B2F(y)
			case AB:
				q[i] = B2F(B(y[j]))
			case F:
				q[i] = y
			case AF:
				q[i] = y[j]
			case I:
				q[i] = F(y)
			case AI:
				q[i] = F(y[j])
			}
		}
		r[j] = q
	}
	return r
}

func flipAI(x AO) AI {
	r := make(AI, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case B:
			r[i] = B2I(y)
		case AB:
			r[i] = B2I(B(y[0]))
		case I:
			r[i] = y
		case AI:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAI(x AO, lines int) AO {
	r := make(AO, lines)
	a := make(AI, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case B:
				q[i] = B2I(y)
			case AB:
				q[i] = B2I(B(y[j]))
			case I:
				q[i] = y
			case AI:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAS(x AO) AS {
	r := make(AS, len(x))
	for i, y := range x {
		switch y := y.(type) {
		case S:
			r[i] = y
		case AS:
			r[i] = y[0]
		}
	}
	return r
}

func flipAOAS(x AO, lines int) AO {
	r := make(AO, lines)
	a := make(AS, lines*len(x))
	for j := range r {
		q := a[j*len(x) : (j+1)*len(x)]
		for i, y := range x {
			switch y := y.(type) {
			case S:
				q[i] = y
			case AS:
				q[i] = y[j]
			}
		}
		r[j] = q
	}
	return r
}

func flipAO(x AO) AO {
	r := make(AO, len(x))
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

func flipAOAO(x AO, lines int) AO {
	r := make(AO, lines)
	a := make(AO, lines*len(x))
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

// JoinTo returns w,x.
func JoinTo(w, x O) O {
	switch w := w.(type) {
	case B:
		return joinToB(w, x, true)
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
	case AO:
		return joinToAO(x, w, false)
	default:
		switch x := x.(type) {
		case Array:
			return joinAtomToArray(w, x, true)
		default:
			return AO{w, x}
		}
	}
}

func joinToB(w B, x O, left bool) O {
	switch x := x.(type) {
	case B:
		if left {
			return AB{bool(w), bool(x)}
		}
		return AB{bool(x), bool(w)}
	case F:
		if left {
			return AF{B2F(w), x}
		}
		return AF{x, B2F(w)}
	case I:
		if left {
			return AI{B2I(w), x}
		}
		return AI{x, B2I(w)}
	case S:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AO:
		return joinToAO(w, x, left)
	default:
		return AO{w, x}
	}
}

func joinToI(w I, x O, left bool) O {
	switch x := x.(type) {
	case B:
		if left {
			return AI{w, B2I(x)}
		}
		return AI{B2I(x), w}
	case F:
		if left {
			return AF{F(w), x}
		}
		return AF{x, F(w)}
	case I:
		if left {
			return AI{w, x}
		}
		return AI{x, w}
	case S:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AO:
		return joinToAO(w, x, left)
	default:
		return AO{w, x}
	}
}

func joinToF(w F, x O, left bool) O {
	switch x := x.(type) {
	case B:
		if left {
			return AF{w, B2F(x)}
		}
		return AF{B2F(x), w}
	case F:
		if left {
			return AF{w, x}
		}
		return AF{x, w}
	case I:
		if left {
			return AF{w, F(x)}
		}
		return AF{F(x), w}
	case S:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AO:
		return joinToAO(w, x, left)
	default:
		return AO{w, x}
	}
}

func joinToS(w S, x O, left bool) O {
	switch x := x.(type) {
	case B:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case F:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case I:
		if left {
			return AO{w, x}
		}
		return AO{x, w}
	case S:
		if left {
			return AS{w, x}
		}
		return AS{x, w}
	case AB:
		return joinToAB(w, x, left)
	case AF:
		return joinToAF(w, x, left)
	case AI:
		return joinToAI(w, x, left)
	case AS:
		return joinToAS(w, x, left)
	case AO:
		return joinToAO(w, x, left)
	default:
		return AO{w, x}
	}
}

func joinToAO(w O, x AO, left bool) O {
	switch w := w.(type) {
	case Array:
		if left {
			return joinArrays(w, x)
		}
		return joinArrays(x, w)
	default:
		r := make(AO, len(x)+1)
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

func joinArrays(w, x Array) AO {
	r := make(AO, x.Len()+w.Len())
	for i := 0; i < w.Len(); i++ {
		r[i] = w.At(i)
	}
	for i := w.Len(); i < len(r); i++ {
		r[i] = x.At(i - w.Len())
	}
	return r
}

func joinAtomToArray(w O, x Array, left bool) AO {
	r := make(AO, x.Len()+1)
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

func joinToAS(w O, x AS, left bool) O {
	switch w := w.(type) {
	case S:
		r := make(AS, len(x)+1)
		if left {
			r[0] = w
			copy(r[1:], x)
		} else {
			r[len(r)-1] = w
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

func joinToAB(w O, x AB, left bool) O {
	switch w := w.(type) {
	case B:
		r := make(AB, len(x)+1)
		if left {
			r[0] = bool(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = bool(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = w
			for i := 1; i < len(r); i++ {
				r[i] = B2F(B(x[i-1]))
			}
		} else {
			r[len(r)-1] = w
			for i := 0; i < len(r); i++ {
				r[i] = B2F(B(x[i]))
			}
		}
		return r
	case I:
		r := make(AI, len(x)+1)
		if left {
			r[0] = w
			for i := 1; i < len(r); i++ {
				r[i] = B2I(B(x[i-1]))
			}
		} else {
			r[len(r)-1] = w
			for i := 0; i < len(r); i++ {
				r[i] = B2I(B(x[i]))
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

func joinToAI(w O, x AI, left bool) O {
	switch w := w.(type) {
	case B:
		r := make(AI, len(x)+1)
		if left {
			r[0] = B2I(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = B2I(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = w
			for i := 1; i < len(r); i++ {
				r[i] = F(x[i-1])
			}
		} else {
			r[len(r)-1] = w
			for i := 0; i < len(r)-1; i++ {
				r[i] = F(x[i])
			}
		}
		return r
	case I:
		r := make(AI, len(x)+1)
		if left {
			r[0] = w
			copy(r[1:], x)
		} else {
			r[len(r)-1] = w
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

func joinToAF(w O, x AF, left bool) O {
	switch w := w.(type) {
	case B:
		r := make(AF, len(x)+1)
		if left {
			r[0] = B2F(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = B2F(w)
			copy(r[:len(r)-1], x)
		}
		return r
	case F:
		r := make(AF, len(x)+1)
		if left {
			r[0] = w
			copy(r[1:], x)
		} else {
			r[len(r)-1] = w
			copy(r[:len(r)-1], x)
		}
		return r
	case I:
		r := make(AF, len(x)+1)
		if left {
			r[0] = F(w)
			copy(r[1:], x)
		} else {
			r[len(r)-1] = F(w)
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
		r[i] = B2I(B(w[i]))
	}
	copy(r[len(w):], x)
	return r
}

func joinAIAB(w AI, x AB) AI {
	r := make(AI, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = B2I(B(x[i-len(w)]))
	}
	return r
}

func joinABAF(w AB, x AF) AF {
	r := make(AF, len(w)+len(x))
	for i := 0; i < len(w); i++ {
		r[i] = B2F(B(w[i]))
	}
	copy(r[len(w):], x)
	return r
}

func joinAFAB(w AF, x AB) AF {
	r := make(AF, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = B2F(B(x[i-len(w)]))
	}
	return r
}

func joinAIAF(w AI, x AF) AF {
	r := make(AF, len(w)+len(x))
	for i := 0; i < len(w); i++ {
		r[i] = F(w[i])
	}
	copy(r[len(w):], x)
	return r
}

func joinAFAI(w AF, x AI) AF {
	r := make(AF, len(w)+len(x))
	copy(r[:len(w)], w)
	for i := len(w); i < len(r); i++ {
		r[i] = F(x[i-len(w)])
	}
	return r
}

// Enlist returns ,x.
func Enlist(x O) O {
	switch x := x.(type) {
	case B:
		return AB{bool(x)}
	case F:
		return AF{float64(x)}
	case I:
		return AI{int(x)}
	case S:
		return AS{string(x)}
	default:
		return AO{x}
	}
}

//// Pair returns w⋈x.
//func Pair(w, x O) O {
//switch w := w.(type) {
//case B:
//switch x := x.(type) {
//case B:
//return AB{bool(w), bool(x)}
//case I:
//return AI{B2I(w), x}
//case F:
//return AF{B2F(w), x}
//default:
//return AO{w, x}
//}
//case F:
//switch x := x.(type) {
//case B:
//return AF{w, B2F(x)}
//case I:
//return AF{w, F(x)}
//case F:
//return AF{w, x}
//default:
//return AO{w, x}
//}
//case I:
//switch x := x.(type) {
//case B:
//return AI{w, B2I(x)}
//case I:
//return AI{w, x}
//case F:
//return AF{F(w), x}
//default:
//return AO{w, x}
//}
//case S:
//switch x := x.(type) {
//case S:
//return AS{w, x}
//default:
//return AO{w, x}
//}
//default:
//return AO{w, x}
//}
//}

// Windows returns w↕x.
func Windows(w, x O) O {
	i := 0
	switch w := w.(type) {
	case B:
		i = B2I(w)
	case I:
		i = w
	case F:
		i = I(w)
	default:
		// TODO: improve error messages
		return badtype("↕ : w must be un integer")
	}
	switch x := x.(type) {
	case Array:
		if i <= 0 || i >= x.Len()+1 {
			return badtype("↕ : w must be between 0 and 1+≠x")
		}
		r := make(AO, 1+x.Len()-i)
		for j := range r {
			r[j] = x.Slice(j, j+i)
		}
		return r
	default:
		return badtype("↕ : x must be an array")
	}
}

// Group returns ⊔x.
func Group(x O) O {
	if Length(x) == 0 {
		return AO{}
	}
	// TODO: optimize allocations
	switch x := x.(type) {
	case AB:
		_, max := minMaxB(x)
		r := make(AO, B2I(max)+1)
		for i := range r {
			r[i] = AI{}
		}
		for i, v := range x {
			j := B2I(B(v))
			rj := r[j].(AI)
			r[j] = append(rj, i)
		}
		return r
	case AI:
		min, max := minMax(x)
		if min < 0 {
			return badtype("⊔ : x must not contain negative values")
		}
		r := make(AO, max+1)
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
		return badtype("⊔ : x must be a non negative integer array")
	}
}
