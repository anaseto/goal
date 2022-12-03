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
	switch xv := x.Value.(type) {
	case F:
		return NewV(-xv)
	case *AB:
		r := make([]int, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = -B2I(xi)
		}
		return NewAI(r)
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = -xi
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = -xi
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = negate(xi)
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
	case *AB:
		return x
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = signI(xi)
		}
		return NewV(r)
	case *AF:
		r := make([]int, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = signF(F(xi))
		}
		return NewAI(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = sign(xi)
		}
		return NewV(r)
	default:
		return errType("sign x", "x", x)
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
	case *AB:
		return x
	case *AI:
		return x
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			// NOTE: we assume conversion is possible, leaving
			// handling NaN, INF or big floats to the program.
			r.Slice[i] = math.Floor(xi)
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = strings.ToLower(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = floor(xi)
		}
		return NewV(r)
	default:
		return errType("_N", "N", x)
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
	case *AB:
		return x
	case *AI:
		return x
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = math.Ceil(xi)
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = strings.ToUpper(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = ceil(xi)
		}
		return NewV(r)
	default:
		return errType("ceil x", "x", x)
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
	case *AB:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = !xi
		}
		return NewV(r)
	case *AI:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = xi == 0
		}
		return NewAB(r)
	case *AF:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = xi == 0
		}
		return NewAB(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = not(xi)
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
	case *AB:
		return x
	case *AI:
		r := make([]int, xv.Len())
		for i := range r {
			r[i] = int(absI(xv.At(i)))
		}
		return NewAI(r)
	case *AF:
		r := make([]float64, xv.Len())
		for i := range r {
			r[i] = math.Abs(xv.At(i))
		}
		return NewAF(r)
	case *AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = abs(xv.At(i))
		}
		return NewAV(r)
	default:
		return errType("abs x", "x", x)
	}
}

func absI(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
