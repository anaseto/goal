package main

import "math"

func B2I(b B) (i I) {
	if b {
		i = 1
	}
	return
}

func B2F(b B) (f F) {
	if b {
		f = 1
	}
	return
}

func isNum(x O) bool {
	switch x.(type) {
	case B, I, F:
		return true
	default:
		return false
	}
}

func isArray(x O) bool {
	switch x.(type) {
	case AO, AB, AI, AF, AS:
		return true
	default:
		return false
	}
}

func sign(x F) (sign int) {
	if x > 0 {
		sign = 1
	} else if x < 0 {
		sign = -1
	}
	return sign
}

func divide(w, x F) F {
	if x == 0 {
		return F(math.Inf(sign(w)))
	}
	return w / x
}

func modulus(w, x I) I {
	if x == 0 {
		// XXX: really?
		return x
	}
	return x % w
}

func minI(w, x I) I {
	if w < x {
		return w
	}
	return x
}

func maxI(w, x I) I {
	if w < x {
		return x
	}
	return w
}

func minS(w, x S) S {
	if w < x {
		return w
	}
	return x
}

func maxS(w, x S) S {
	if w < x {
		return x
	}
	return w
}

func clone(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = clone(x[i])
		}
		return r
	case E:
		return x
	default:
		return x
	}
}

func cloneShallow(x O) O {
	switch x := x.(type) {
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		copy(r, x)
		return r
	default:
		return x
	}
}

func toArray(x O) O {
	switch x := x.(type) {
	case B:
		return AB{x}
	case F:
		return AF{x}
	case I:
		return AI{x}
	case S:
		return AS{x}
	default:
		return x
	}
}

func growArray(x O, n I) O {
	l := Length(x)
	if l >= n && n >= -l {
		return x
	}
	i := 0
	if n < 0 {
		i = l + n
		n = -n
		if i < 0 {
			i = 0
		}
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, n)
		copy(r[i:], x)
		return r
	case AF:
		r := make(AF, n)
		copy(r[i:], x)
		return r
	case AI:
		r := make(AI, n)
		copy(r[i:], x)
		return r
	case AS:
		r := make(AS, n)
		copy(r[i:], x)
		return r
	case AO:
		r := make(AO, n)
		copy(r[i:], x)
		return r
	default:
		return x
	}
}

func isFalse(x O) bool {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x == 0
	case I:
		return x == 0
	case S:
		return x == ""
	case AB:
		return len(x) == 0
	case AF:
		return len(x) == 0
	case AI:
		return len(x) == 0
	case AS:
		return len(x) == 0
	case AO:
		return len(x) == 0
	case E:
		return true
	default:
		// TODO: Interface for other objects?
		return false
	}
}
