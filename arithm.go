package goal

import (
	"math"
	"strings"
)

// negate returns -x.
func negate(x V) V {
	if x.IsI() {
		return NewI(-x.I())
	}
	if x.IsF() {
		return NewF(-x.F())
	}
	switch xv := x.value.(type) {
	case *AB:
		r := make([]int64, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = -b2i(xi)
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
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
			ri := negate(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("-x", "x", x)
	}
}

func signF(x float64) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	default:
		return 0
	}
}

func signI(x int64) int64 {
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
	if x.IsI() {
		return NewI(signI(x.I()))
	}
	if x.IsF() {
		return NewI(int64(signF(x.F())))
	}
	switch xv := x.value.(type) {
	case *AB:
		return x
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = signI(xi)
		}
		return NewV(r)
	case *AF:
		r := make([]int64, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = int64(signF(xi))
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := sign(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("sign x", "x", x)
	}
}

// floor returns _x.
func floor(x V) V {
	if x.IsI() {
		return x
	}
	if x.IsF() {
		return NewF(math.Floor(float64(x.F())))
	}
	switch xv := x.value.(type) {
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
			ri := floor(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("_N", "N", x)
	}
}

// ceil returns ⌈x.
func ceil(x V) V {
	if x.IsI() {
		return x
	}
	if x.IsF() {
		return NewF(math.Ceil(float64(x.F())))
	}
	switch xv := x.value.(type) {
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
			ri := ceil(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("ceil x", "x", x)
	}
}

// not returns ~x.
func not(x V) V {
	if x.IsI() {
		return NewI(b2i(x.I() == 0))
	}
	if x.IsF() {
		return NewI(b2i(x.F() == 0))
	}
	switch xv := x.value.(type) {
	case S:
		return NewI(b2i(xv == ""))
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
		return NewABWithRC(r, reuseRCp(xv.rc))
	case *AF:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = xi == 0
		}
		return NewABWithRC(r, reuseRCp(xv.rc))
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := not(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return NewI(b2i(!isTrue(x)))
	}
}

// abs returns abs[x].
func abs(x V) V {
	if x.IsI() {
		return NewI(absI(x.I()))
	}
	if x.IsF() {
		return NewF(math.Abs(float64(x.F())))
	}
	switch xv := x.value.(type) {
	case *AB:
		return x
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = absI(xi)
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = math.Abs(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := abs(xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("abs x", "x", x)
	}
}

func absI(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
