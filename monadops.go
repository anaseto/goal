package main

import (
	"math"
	"strings"
)

// Negate returns -x.
func Negate(x V) V {
	switch x := x.(type) {
	case B:
		return -B2I(x)
	case F:
		return -x
	case I:
		return -x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -B2I(B(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Negate(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("=")
	}
}

func signF(x F) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

func signI(x I) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

// Sign returns ×x.
func Sign(x V) V {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return signF(x)
	case I:
		return signI(x)
	case AB:
		return x
	case AF:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signF(F(x[i]))
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signI(x[i])
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Sign(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("×")
	}
}

// Floor returns _x.
func Floor(x V) V {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Floor(float64(x))
	case I:
		return x
	case S:
		return strings.ToLower(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Floor(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToLower(x[i])
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Floor(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌊")
	}
}

// Ceil returns ⌈x.
func Ceil(x V) V {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Ceil(float64(x))
	case I:
		return x
	case S:
		return strings.ToUpper(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Ceil(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToUpper(x[i])
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Ceil(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌈")
	}
}

// Not returns ~x.
func Not(x V) V {
	switch x := x.(type) {
	case B:
		return !x
	case F:
		return 1 - x
	case I:
		return 1 - x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = !x[i]
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Not(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}

func absI(x I) I {
	if x < 0 {
		return -x
	}
	return x
}

// Abs returns |x.
func Abs(x V) V {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return F(math.Abs(float64(x)))
	case I:
		return absI(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Abs(x[i])
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = absI(x[i])
		}
		return r
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = Abs(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}
