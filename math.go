package goal

import "math"

// vfNaN implements the nan variadic verb.
func vfNaN(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		r := isNaN(args[0])
		return r
	case 2:
		return fillNaN(args[1], args[0])
	default:
		return panicRank("nan")
	}
}

func isNaN(x V) V {
	if x.IsI() {
		return NewI(b2I(false))
	}
	if x.IsF() {
		return NewI(b2I(math.IsNaN(x.F())))
	}
	switch xv := x.bv.(type) {
	case *AB:
		r := xv.reuse()
		for i := range r.elts {
			r.elts[i] = 0
		}
		r.flags = flagBool
		return NewV(r)
	case *AI:
		r := make([]byte, xv.Len())
		return newABb(r)
	case *AF:
		r := make([]byte, xv.Len())
		for i, xi := range xv.elts {
			r[i] = b2B(math.IsNaN(xi))
		}
		return newABb(r)
	case *AV:
		return mapAV(xv, isNaN)
	case *D:
		return newDictValues(xv.keys, isNaN(NewV(xv.values)))
	default:
		return panicType("NaN x", "x", x)
	}
}

func fillNaN(x V, y V) V {
	var fill float64
	if x.IsI() {
		fill = float64(x.I())
	} else if x.IsF() {
		fill = x.F()
	} else {
		return panicType("x NaN y", "x", x)
	}
	r := fillNaNf(fill, y)
	return r
}

func fillNaNf(fill float64, y V) V {
	if y.IsI() {
		return y
	}
	if y.IsF() {
		if math.IsNaN(y.F()) {
			return NewF(fill)
		}
		return y
	}
	switch yv := y.bv.(type) {
	case *AB:
		return y
	case *AI:
		return y
	case *AF:
		var r []float64
		if yv.reusable() {
			r = yv.elts
		} else {
			r = make([]float64, yv.Len())
			copy(r, yv.elts)
		}
		for i, ri := range r {
			if math.IsNaN(ri) {
				r[i] = fill
			}
		}
		return NewAF(r)
	case *AV:
		return mapAV(yv, func(yi V) V { return fillNaNf(fill, yi) })
	case *D:
		return newDictValues(yv.keys, fillNaNf(fill, NewV(yv.values)))
	default:
		return panicType("x NaN y", "y", y)
	}
}

// vfAtan implements the atan variadic verb.
func vfAtan(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		r := mathm(args[0], math.Atan)
		if r.IsPanic() {
			return ppanic("tan x : ", r)
		}
		return r
	case 2:
		return arctan2(args[1], args[0])
	default:
		return panicRank("+")
	}
}
