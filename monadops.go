package goal

import (
	"math"
	"strings"
)

// negate returns -x.
func negate(x V) V {
	switch x := x.BV.(type) {
	case F:
		return -x
	case I:
		return -x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(-B2I(x[i]))
		}
		return newBV(r)
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return newBV(r)
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = negate(x[i])
		}
		return newBV(r)
	default:
		return errType("-x", "x", x)
	}
}

func signF(x F) I {
	switch {
	case x > 0:
		return newBV(I(1))
	case x < 0:
		return newBV(I(-1))
	default:
		return newBV(I(0))
	}
}

func signI(x I) I {
	switch {
	case x > 0:
		return newBV(I(1))
	case x < 0:
		return newBV(I(-1))
	default:
		return newBV(I(0))
	}
}

// sign returns sign x.
func sign(x V) V {
	switch x := x.BV.(type) {
	case F:
		return signF(x)
	case I:
		return signI(x)
	case AB:
		return newBV(x)
	case AF:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(signF(F(x[i])))
		}
		return newBV(r)
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(signI(I(x[i])))
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = sign(x[i])
		}
		return newBV(r)
	default:
		return errType("sign x", "x", x)
	}
}

// floor returns _x.
func floor(x V) V {
	switch x := x.BV.(type) {
	case F:
		return F(math.Floor(float64(x)))
	case I:
		return newBV(x)
	case S:
		return S(strings.ToLower(string(x)))
	case AB:
		return newBV(x)
	case AF:
		r := make(AI, len(x))
		for i := range r {
			// NOTE: we assume conversion is possible, leaving
			// handling NaN, INF or big floats to the program.
			r[i] = int(math.Floor(x[i]))
		}
		return newBV(r)
	case AI:
		return newBV(x)
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToLower(x[i])
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = floor(x[i])
		}
		return newBV(r)
	default:
		return errType("_N", "N", x)
	}
}

// ceil returns ⌈x.
func ceil(x V) V {
	switch x := x.BV.(type) {
	case F:
		return F(math.Ceil(float64(x)))
	case I:
		return newBV(x)
	case S:
		return S(strings.ToUpper(string(x)))
	case AB:
		return newBV(x)
	case AF:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(math.Ceil(x[i]))
		}
		return newBV(r)
	case AI:
		return newBV(x)
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToUpper(x[i])
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = ceil(x[i])
		}
		return newBV(r)
	default:
		return errType("ceil x", "x", x)
	}
}

// not returns ~x.
func not(x V) V {
	switch x := x.BV.(type) {
	case F:
		return B2I(x == 0)
	case I:
		return B2I(x == 0)
	case S:
		return B2I(x == "")
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = !x[i]
		}
		return newBV(r)
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = x[i] == 0
		}
		return newBV(r)
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = x[i] == 0
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = not(x[i])
		}
		return newBV(r)
	default:
		return B2I(!isTrue(x))
	}
}

// abs returns abs[x].
func abs(x V) V {
	switch x := x.BV.(type) {
	case F:
		return F(math.Abs(float64(x)))
	case I:
		return absI(x)
	case AB:
		return newBV(x)
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Abs(x[i])
		}
		return newBV(r)
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(absI(I(x[i])))
		}
		return newBV(r)
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = abs(x[i])
		}
		return newBV(r)
	default:
		return errType("abs x", "x", x)
	}
}

func absI(x I) I {
	if x < 0 {
		return -x
	}
	return newBV(x)
}
