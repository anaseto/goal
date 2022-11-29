package goal

import (
	"math"
	"strings"
)

// negate returns -x.
func negate(x V) V {
	switch x := x.BV.(type) {
	case I:
		return newBV(-x)
	case F:
		return newBV(-x)
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = int(-B2I(x[i]))
		}
		return newBV(r)
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return newBV(r)
	case AF:
		r := make(AF, len(x))
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

// sign returns sign x.
func sign(x V) V {
	switch xv := x.BV.(type) {
	case I:
		return newBV(signI(xv))
	case F:
		return newBV(signF(xv))
	case AB:
		return x
	case AI:
		r := make(AI, xv.Len())
		for i := range r {
			r[i] = int(signI(I(xv[i])))
		}
		return newBV(r)
	case AF:
		r := make(AI, xv.Len())
		for i := range r {
			r[i] = int(signF(F(xv[i])))
		}
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		for i := range r {
			r[i] = sign(xv[i])
		}
		return newBV(r)
	default:
		return errType("sign x", "x", xv)
	}
}

// floor returns _x.
func floor(x V) V {
	switch xv := x.BV.(type) {
	case I:
		return x
	case F:
		return newBV(F(math.Floor(float64(xv))))
	case S:
		return newBV(S(strings.ToLower(string(xv))))
	case AB:
		return x
	case AI:
		return x
	case AF:
		r := make(AI, xv.Len())
		for i := range r {
			// NOTE: we assume conversion is possible, leaving
			// handling NaN, INF or big floats to the program.
			r[i] = int(math.Floor(xv[i]))
		}
		return newBV(r)
	case AS:
		r := make(AS, xv.Len())
		for i := range r {
			r[i] = strings.ToLower(xv[i])
		}
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		for i := range r {
			r[i] = floor(xv[i])
		}
		return newBV(r)
	default:
		return errType("_N", "N", xv)
	}
}

// ceil returns ⌈x. XXX unused for now
func ceil(x V) V {
	switch xv := x.BV.(type) {
	case I:
		return x
	case F:
		return newBV(F(math.Ceil(float64(xv))))
	case S:
		return newBV(S(strings.ToUpper(string(xv))))
	case AB:
		return x
	case AI:
		return x
	case AF:
		r := make(AI, xv.Len())
		for i := range r {
			r[i] = int(math.Ceil(xv[i]))
		}
		return newBV(r)
	case AS:
		r := make(AS, xv.Len())
		for i := range r {
			r[i] = strings.ToUpper(xv[i])
		}
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		for i := range r {
			r[i] = ceil(xv[i])
		}
		return newBV(r)
	default:
		return errType("ceil x", "x", xv)
	}
}

// not returns ~x.
func not(x V) V {
	switch xv := x.BV.(type) {
	case I:
		return newBV(B2I(xv == 0))
	case F:
		return newBV(B2I(xv == 0))
	case S:
		return newBV(B2I(xv == ""))
	case AB:
		r := make(AB, xv.Len())
		for i := range r {
			r[i] = !xv[i]
		}
		return newBV(r)
	case AI:
		r := make(AB, xv.Len())
		for i := range r {
			r[i] = xv[i] == 0
		}
		return newBV(r)
	case AF:
		r := make(AB, xv.Len())
		for i := range r {
			r[i] = xv[i] == 0
		}
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		for i := range r {
			r[i] = not(xv[i])
		}
		return newBV(r)
	default:
		return newBV(B2I(!isTrue(x)))
	}
}

// abs returns abs[x].
func abs(x V) V {
	switch xv := x.BV.(type) {
	case I:
		return newBV(absI(xv))
	case F:
		return newBV(F(math.Abs(float64(xv))))
	case AB:
		return x
	case AI:
		r := make(AI, xv.Len())
		for i := range r {
			r[i] = int(absI(I(xv[i])))
		}
		return newBV(r)
	case AF:
		r := make(AF, xv.Len())
		for i := range r {
			r[i] = math.Abs(xv[i])
		}
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		for i := range r {
			r[i] = abs(xv[i])
		}
		return newBV(r)
	default:
		return errType("abs x", "x", xv)
	}
}

func absI(x I) I {
	if x < 0 {
		return -x
	}
	return x
}
