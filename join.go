package goal

// joinTo returns x,y.
func joinTo(x, y V) V {
	if x.IsInt() {
		return joinToI(x.Int(), y, true)
	}
	switch xv := x.Value.(type) {
	case F:
		return joinToF(xv, y, true)
	case S:
		return joinToS(xv, y, true)
	case AB:
		return joinToAB(y, xv, false)
	case AF:
		return joinToAF(y, xv, false)
	case AI:
		return joinToAI(y, xv, false)
	case AS:
		return joinToAS(y, xv, false)
	case AV:
		return joinToAV(y, xv, false)
	default:
		switch yv := y.Value.(type) {
		case array:
			return NewV(joinAtomToArray(x, yv, true))
		default:
			return NewV(AV{x, y})
		}
	}
}

func joinToI(x int, y V, left bool) V {
	if y.IsInt() {
		if left {
			return NewV(AI{int(x), y.Int()})
		}
		return NewV(AI{y.Int(), int(x)})
	}
	switch yv := y.Value.(type) {
	case F:
		if left {
			return NewV(AF{float64(x), float64(yv)})
		}
		return NewV(AF{float64(yv), float64(x)})
	case S:
		if left {
			return NewV(AV{NewI(x), y})
		}
		return NewV(AV{y, NewI(x)})
	case AB:
		return joinToAB(NewI(x), yv, left)
	case AF:
		return joinToAF(NewI(x), yv, left)
	case AI:
		return joinToAI(NewI(x), yv, left)
	case AS:
		return joinToAS(NewI(x), yv, left)
	case AV:
		return joinToAV(NewI(x), yv, left)
	default:
		return NewV(AV{NewI(x), y})
	}
}

func joinToF(x F, y V, left bool) V {
	if y.IsInt() {
		if left {
			return NewV(AF{float64(x), float64(y.Int())})
		}
		return NewV(AF{float64(y.Int()), float64(x)})
	}
	switch yv := y.Value.(type) {
	case F:
		if left {
			return NewV(AF{float64(x), float64(yv)})
		}
		return NewV(AF{float64(yv), float64(x)})
	case S:
		if left {
			return NewV(AV{NewV(x), y})
		}
		return NewV(AV{y, NewV(x)})
	case AB:
		return joinToAB(NewV(x), yv, left)
	case AF:
		return joinToAF(NewV(x), yv, left)
	case AI:
		return joinToAI(NewV(x), yv, left)
	case AS:
		return joinToAS(NewV(x), yv, left)
	case AV:
		return joinToAV(NewV(x), yv, left)
	default:
		return NewV(AV{NewV(x), y})
	}
}

func joinToS(x S, y V, left bool) V {
	if y.IsInt() {
		if left {
			return NewV(AV{NewV(x), y})
		}
		return NewV(AV{y, NewV(x)})
	}
	switch yv := y.Value.(type) {
	case F:
		if left {
			return NewV(AV{NewV(x), y})
		}
		return NewV(AV{y, NewV(x)})
	case S:
		if left {
			return NewV(AS{string(x), string(yv)})
		}
		return NewV(AS{string(yv), string(x)})
	case AB:
		return joinToAB(NewV(x), yv, left)
	case AF:
		return joinToAF(NewV(x), yv, left)
	case AI:
		return joinToAI(NewV(x), yv, left)
	case AS:
		return joinToAS(NewV(x), yv, left)
	case AV:
		return joinToAV(NewV(x), yv, left)
	default:
		return NewV(AV{NewV(x), y})
	}
}

func joinToAV(x V, y AV, left bool) V {
	switch xv := x.Value.(type) {
	case array:
		if left {
			return NewV(joinArrays(xv, y))
		}
		return NewV(joinArrays(y, xv))
	default:
		r := make(AV, len(y)+1)
		if left {
			r[0] = x
			copy(r[1:], y)
		} else {
			r[len(r)-1] = x
			copy(r[:len(r)-1], y)
		}
		return NewV(r)
	}
}

func joinArrays(x, y array) AV {
	r := make(AV, y.Len()+x.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = x.at(i)
	}
	for i := x.Len(); i < len(r); i++ {
		r[i] = y.at(i - x.Len())
	}
	return r
}

func joinAtomToArray(x V, y array, left bool) AV {
	r := make(AV, y.Len()+1)
	if left {
		r[0] = x
		for i := 1; i < len(r); i++ {
			r[i] = y.at(i - 1)
		}
	} else {
		r[len(r)-1] = x
		for i := 0; i < len(r)-1; i++ {
			r[i] = y.at(i)
		}
	}
	return r
}

func joinToAS(x V, y AS, left bool) V {
	switch xv := x.Value.(type) {
	case S:
		r := make(AS, len(y)+1)
		if left {
			r[0] = string(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = string(xv)
			copy(r[:len(r)-1], y)
		}
		return NewV(r)
	case AS:
		r := make(AS, len(y)+xv.Len())
		if left {
			copy(r[:xv.Len()], xv)
			copy(r[xv.Len():], y)
		} else {
			copy(r[:len(y)], y)
			copy(r[len(y):], xv)
		}
		return NewV(r)
	case array:
		if left {
			return NewV(joinArrays(xv, y))
		}
		return NewV(joinArrays(y, xv))
	default:
		return NewV(joinAtomToArray(x, y, left))
	}
}

func joinToAB(x V, y AB, left bool) V {
	if x.IsInt() {
		if isBI(x.Int()) {
			r := make(AB, len(y)+1)
			if left {
				r[0] = x.Int() == 1
				copy(r[1:], y)
			} else {
				r[len(r)-1] = x.Int() == 1
				copy(r[:len(r)-1], y)
			}
			return NewV(r)
		}
		r := make(AI, len(y)+1)
		if left {
			r[0] = int(x.Int())
			for i := 1; i < len(r); i++ {
				r[i] = int(B2I(y[i-1]))
			}
		} else {
			r[len(r)-1] = int(x.Int())
			for i := 0; i < len(r); i++ {
				r[i] = int(B2I(y[i]))
			}
		}
		return NewV(r)

	}
	switch xv := x.Value.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			for i := 1; i < len(r); i++ {
				r[i] = float64(B2F(y[i-1]))
			}
		} else {
			r[len(r)-1] = float64(xv)
			for i := 0; i < len(r); i++ {
				r[i] = float64(B2F(y[i]))
			}
		}
		return NewV(r)
	case AB:
		if left {
			return NewV(joinABAB(xv, y))
		}
		return NewV(joinABAB(y, xv))
	case AI:
		if left {
			return NewV(joinAIAB(xv, y))
		}
		return NewV(joinABAI(y, xv))
	case AF:
		if left {
			return NewV(joinAFAB(xv, y))
		}
		return NewV(joinABAF(y, xv))
	case array:
		if left {
			return NewV(joinArrays(xv, y))
		}
		return NewV(joinArrays(y, xv))
	default:
		return NewV(joinAtomToArray(x, y, left))
	}
}

func joinToAI(x V, y AI, left bool) V {
	if x.IsInt() {
		r := make(AI, len(y)+1)
		if left {
			r[0] = x.Int()
			copy(r[1:], y)
		} else {
			r[len(r)-1] = x.Int()
			copy(r[:len(r)-1], y)
		}
		return NewV(r)

	}
	switch xv := x.Value.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			for i := 1; i < len(r); i++ {
				r[i] = float64(y[i-1])
			}
		} else {
			r[len(r)-1] = float64(xv)
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(y[i])
			}
		}
		return NewV(r)
	case AB:
		if left {
			return NewV(joinABAI(xv, y))
		}
		return NewV(joinAIAB(y, xv))
	case AI:
		if left {
			return NewV(joinAIAI(xv, y))
		}
		return NewV(joinAIAI(y, xv))
	case AF:
		if left {
			return NewV(joinAFAI(xv, y))
		}
		return NewV(joinAIAF(y, xv))
	case array:
		if left {
			return NewV(joinArrays(xv, y))
		}
		return NewV(joinArrays(y, xv))
	default:
		return NewV(joinAtomToArray(x, y, left))
	}
}

func joinToAF(x V, y AF, left bool) V {
	if x.IsInt() {
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(x.Int())
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(x.Int())
			copy(r[:len(r)-1], y)
		}
		return NewV(r)
	}
	switch xv := x.Value.(type) {
	case F:
		r := make(AF, len(y)+1)
		if left {
			r[0] = float64(xv)
			copy(r[1:], y)
		} else {
			r[len(r)-1] = float64(xv)
			copy(r[:len(r)-1], y)
		}
		return NewV(r)
	case AB:
		if left {
			return NewV(joinABAF(xv, y))
		}
		return NewV(joinAFAB(y, xv))
	case AI:
		if left {
			return NewV(joinAIAF(xv, y))
		}
		return NewV(joinAFAI(y, xv))
	case AF:
		if left {
			return NewV(joinAFAF(xv, y))
		}
		return NewV(joinAFAF(y, xv))
	case array:
		if left {
			return NewV(joinArrays(xv, y))
		}
		return NewV(joinArrays(y, xv))
	default:
		return NewV(joinAtomToArray(x, y, left))
	}
}

func joinABAB(x AB, y AB) AB {
	r := make(AB, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAIAI(x AI, y AI) AI {
	r := make(AI, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAFAF(x AF, y AF) AF {
	r := make(AF, len(y)+len(x))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinABAI(x AB, y AI) AI {
	r := make(AI, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = int(B2I(x[i]))
	}
	copy(r[len(x):], y)
	return r
}

func joinAIAB(x AI, y AB) AI {
	r := make(AI, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = int(B2I(y[i-len(x)]))
	}
	return r
}

func joinABAF(x AB, y AF) AF {
	r := make(AF, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = float64(B2F(x[i]))
	}
	copy(r[len(x):], y)
	return r
}

func joinAFAB(x AF, y AB) AF {
	r := make(AF, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = float64(B2F(y[i-len(x)]))
	}
	return r
}

func joinAIAF(x AI, y AF) AF {
	r := make(AF, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = float64(x[i])
	}
	copy(r[len(x):], y)
	return r
}

func joinAFAI(x AF, y AI) AF {
	r := make(AF, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = float64(y[i-len(x)])
	}
	return r
}

// enlist returns ,x.
func enlist(x V) V {
	if x.IsInt() {
		if isBI(x.Int()) {
			return NewV(AB{x.Int() == 1})
		}
		return NewV(AI{int(x.Int())})
	}
	switch xv := x.Value.(type) {
	case F:
		return NewV(AF{float64(xv)})
	case S:
		return NewV(AS{string(xv)})
	default:
		return NewV(AV{x})
	}
}
