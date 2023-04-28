package goal

import (
	"math"
	"strings"
	"unicode"
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
	case S:
		return NewS(strings.TrimRightFunc(string(xv), unicode.IsSpace))
	case *AB:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
			r[i] = -int64(xi)
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = -xi
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = -xi
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := negate(xi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = strings.TrimRightFunc(xi, unicode.IsSpace)
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, negate(NewV(xv.values)))
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
		if xv.IsBoolean() {
			return x
		}
		r := xv.reuse()
		for i, xi := range xv.elts {
			if xi > 1 {
				r.elts[i] = 1
			}
		}
		return NewV(r)
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = signI(xi)
		}
		return NewV(r)
	case *AF:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
			r[i] = int64(signF(xi))
		}
		return NewAIWithRC(r, reuseRCp(xv.rc))
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := sign(xi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, sign(NewV(xv.values)))
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
		for i, xi := range xv.elts {
			// NOTE: we assume conversion is possible, leaving
			// handling NaN, INF or big floats to the program.
			r.elts[i] = math.Floor(xi)
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = strings.ToLower(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := floor(xi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, floor(NewV(xv.values)))
	default:
		return panicType("_N", "N", x)
	}
}

// ceil returns ceil x.
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
		for i, xi := range xv.elts {
			r.elts[i] = math.Ceil(xi)
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = strings.ToUpper(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := ceil(xi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, ceil(NewV(xv.values)))
	default:
		return panicType("ceil x", "x", x)
	}
}

// not returns ~x.
func not(x V) V {
	if x.IsI() {
		return NewI(b2I(x.I() == 0))
	}
	if x.IsF() {
		return NewI(b2I(x.F() == 0))
	}
	switch xv := x.value.(type) {
	case S:
		return NewI(b2I(xv == ""))
	case *AB:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = b2B(xi == 0)
		}
		return NewV(r)
	case *AI:
		r := make([]byte, xv.Len())
		for i, xi := range xv.elts {
			r[i] = b2B(xi == 0)
		}
		return NewABWithRC(r, reuseRCp(xv.rc))
	case *AF:
		r := make([]byte, xv.Len())
		for i, xi := range xv.elts {
			r[i] = b2B(xi == 0)
		}
		return NewABWithRC(r, reuseRCp(xv.rc))
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = not(xi)
			// never panics
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, not(NewV(xv.values)))
	default:
		return NewI(b2I(!x.IsTrue()))
	}
}

// abs returns abs x.
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
		for i, xi := range xv.elts {
			r.elts[i] = absI(xi)
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = math.Abs(xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := abs(xi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, abs(NewV(xv.values)))
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
