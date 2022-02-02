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
		t := dType(x)
		switch {
		case lines <= 0:
			// (+⟨⟨⟩,…⟩) ≡ ⟨⟩
			// TODO: error if atoms?
			return x[0]
		case lines == 1:
			switch t {
			case tB:
				r := make(AB, cols)
				for i, y := range x {
					switch y := y.(type) {
					case B:
						r[i] = y
					case AB:
						r[i] = y[0]
					}
				}
				return r
			case tF:
				r := make(AF, cols)
				for i, y := range x {
					switch y := y.(type) {
					case B:
						r[i] = B2F(y)
					case AB:
						r[i] = B2F(y[0])
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
			case tI:
				r := make(AI, cols)
				for i, y := range x {
					switch y := y.(type) {
					case B:
						r[i] = B2I(y)
					case AB:
						r[i] = B2I(y[0])
					case I:
						r[i] = y
					case AI:
						r[i] = y[0]
					}
				}
				return r
			case tS:
				r := make(AS, cols)
				for i, y := range x {
					switch y := y.(type) {
					case S:
						r[i] = y
					case AS:
						r[i] = y[0]
					}
				}
				return r
			default:
				r := make(AO, cols)
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
		default:
			switch t {
			case tB:
				r := make(AO, lines)
				for j := range r {
					q := make(AB, cols)
					for i, y := range x {
						switch y := y.(type) {
						case B:
							q[i] = y
						case AB:
							q[i] = y[j]
						}
					}
					r[j] = q
				}
				return r
			case tF:
				r := make(AO, lines)
				for j := range r {
					q := make(AF, cols)
					for i, y := range x {
						switch y := y.(type) {
						case B:
							q[i] = B2F(y)
						case AB:
							q[i] = B2F(y[j])
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
			case tI:
				r := make(AO, lines)
				for j := range r {
					q := make(AI, cols)
					for i, y := range x {
						switch y := y.(type) {
						case B:
							q[i] = B2I(y)
						case AB:
							q[i] = B2I(y[j])
						case I:
							q[i] = y
						case AI:
							q[i] = y[j]
						}
					}
					r[j] = q
				}
				return r
			case tS:
				r := make(AO, lines)
				for j := range r {
					q := make(AS, cols)
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
			default:
				r := make(AO, lines)
				for j := range r {
					q := make(AO, cols)
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
		}
	default:
		return AO{x}
	}
}
