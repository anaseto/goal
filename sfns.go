// structural functions (Length, Reverse, Take, ...)

package main

import "sort"

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

// First returns ↑x.
func First(x O) O {
	switch x := x.(type) {
	case Array:
		if x.Len() == 0 {
			return badlen("↑")
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

// Drop returns w↓x.
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

// Take returns w↑x.
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
				r[i] = B2F(x[i-max])
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i] = B2I(x[i-max])
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
				r[i] = B2F(w[i])
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
				r[i] = B2I(w[i])
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
				r[i-max] = B2F(x[i])
			}
			return r
		case AI:
			r := make(AI, len(x))
			for i := 0; i < max; i++ {
				r[len(x)-1-i] = w[i]
			}
			for i := max; i < len(x); i++ {
				r[i-max] = B2I(x[i])
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
				r[len(x)-1-i] = B2F(w[i])
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
				r[len(x)-1-i] = B2I(w[i])
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
