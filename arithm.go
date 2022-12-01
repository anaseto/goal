package goal

import (
	"math"
	"strings"
)

// negate returns -x.
func negate(x V) V {
	if x.IsInt() {
		return NewI(-x.Int())
	}
	switch x := x.Value.(type) {
	case F:
		return NewV(-x)
	case AB:
		r := make([]int, len(x))
		for i := range r {
			r[i] = int(-B2I(x[i]))
		}
		return NewV(r)
	case AI:
		r := make([]int, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return NewV(r)
	case AF:
		r := make([]float64, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return NewV(r)
	case AV:
		r := make([]V, len(x))
		for i := range r {
			r[i] = negate(x[i])
		}
		return NewV(r)
	default:
		return errType("-x", "x", x)
	}
}

func signF(x F) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	default:
		return 0
	}
}

func signI(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	default:
		return 0
	}
}

// sign returns sign x.
func sign(x V) V {
	if x.IsInt() {
		return NewI(signI(x.Int()))
	}
	switch xv := x.Value.(type) {
	case F:
		return NewI(signF(xv))
	case AB:
		return x
	case AI:
		r := make([]int, xv.Len())
		for i := range r {
			r[i] = int(signI(xv[i]))
		}
		return NewV(r)
	case AF:
		r := make([]int, xv.Len())
		for i := range r {
			r[i] = int(signF(F(xv[i])))
		}
		return NewV(r)
	case AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = sign(xv[i])
		}
		return NewV(r)
	default:
		return errType("sign x", "x", xv)
	}
}

// floor returns _x.
func floor(x V) V {
	if x.IsInt() {
		return x
	}
	switch xv := x.Value.(type) {
	case F:
		return NewF(math.Floor(float64(xv)))
	case S:
		return NewS(strings.ToLower(string(xv)))
	case AB:
		return x
	case AI:
		return x
	case AF:
		r := make([]int, xv.Len())
		for i := range r {
			// NOTE: we assume conversion is possible, leaving
			// handling NaN, INF or big floats to the program.
			r[i] = int(math.Floor(xv[i]))
		}
		return NewV(r)
	case AS:
		r := make([]string, xv.Len())
		for i := range r {
			r[i] = strings.ToLower(xv[i])
		}
		return NewV(r)
	case AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = floor(xv[i])
		}
		return NewV(r)
	default:
		return errType("_N", "N", xv)
	}
}

// ceil returns ⌈x. XXX unused for now
func ceil(x V) V {
	if x.IsInt() {
		return x
	}
	switch xv := x.Value.(type) {
	case F:
		return NewF(math.Ceil(float64(xv)))
	case S:
		return NewS(strings.ToUpper(string(xv)))
	case AB:
		return x
	case AI:
		return x
	case AF:
		r := make([]int, xv.Len())
		for i := range r {
			r[i] = int(math.Ceil(xv[i]))
		}
		return NewV(r)
	case AS:
		r := make([]string, xv.Len())
		for i := range r {
			r[i] = strings.ToUpper(xv[i])
		}
		return NewV(r)
	case AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = ceil(xv[i])
		}
		return NewV(r)
	default:
		return errType("ceil x", "x", xv)
	}
}

// not returns ~x.
func not(x V) V {
	if x.IsInt() {
		return NewI(B2I(x.Int() == 0))
	}
	switch xv := x.Value.(type) {
	case F:
		return NewI(B2I(xv == 0))
	case S:
		return NewI(B2I(xv == ""))
	case AB:
		r := make([]bool, xv.Len())
		for i := range r {
			r[i] = !xv[i]
		}
		return NewV(r)
	case AI:
		r := make([]bool, xv.Len())
		for i := range r {
			r[i] = xv[i] == 0
		}
		return NewV(r)
	case AF:
		r := make([]bool, xv.Len())
		for i := range r {
			r[i] = xv[i] == 0
		}
		return NewV(r)
	case AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = not(xv[i])
		}
		return NewV(r)
	default:
		return NewI(B2I(!isTrue(x)))
	}
}

// abs returns abs[x].
func abs(x V) V {
	if x.IsInt() {
		return NewI(absI(x.Int()))
	}
	switch xv := x.Value.(type) {
	case F:
		return NewF(math.Abs(float64(xv)))
	case AB:
		return x
	case AI:
		r := make([]int, xv.Len())
		for i := range r {
			r[i] = int(absI(xv[i]))
		}
		return NewV(r)
	case AF:
		r := make([]float64, xv.Len())
		for i := range r {
			r[i] = math.Abs(xv[i])
		}
		return NewV(r)
	case AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = abs(xv[i])
		}
		return NewV(r)
	default:
		return errType("abs x", "x", xv)
	}
}

func absI(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
