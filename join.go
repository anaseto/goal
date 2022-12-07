package goal

// joinTo returns x,y.
func joinTo(x, y V) V {
	if x.IsI() {
		return joinToI(x.I(), y, true)
	}
	if x.IsF() {
		return joinToF(x.F(), y, true)
	}
	switch xv := x.Value.(type) {
	case S:
		return joinToS(xv, y, true)
	case *AB:
		return joinToAB(y, xv, false)
	case *AF:
		return joinToAF(y, xv, false)
	case *AI:
		return joinToAI(y, xv, false)
	case *AS:
		return joinToAS(y, xv, false)
	case *AV:
		return joinToAV(y, xv, false)
	default:
		switch yv := y.Value.(type) {
		case array:
			return joinAtomToArray(x, yv, true)
		default:
			return NewAV([]V{x, y})
		}
	}
}

func joinToI(x int64, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAI([]int64{int64(x), y.I()})
		}
		return NewAI([]int64{y.I(), x})
	}
	if y.IsF() {
		if left {
			return NewAF([]float64{float64(x), float64(y.F())})
		}
		return NewAF([]float64{float64(y.F()), float64(x)})
	}
	switch yv := y.Value.(type) {
	case S:
		if left {
			return NewAV([]V{NewI(x), y})
		}
		return NewAV([]V{y, NewI(x)})
	case *AB:
		return joinToAB(NewI(x), yv, left)
	case *AF:
		return joinToAF(NewI(x), yv, left)
	case *AI:
		return joinToAI(NewI(x), yv, left)
	case *AS:
		return joinToAS(NewI(x), yv, left)
	case *AV:
		return joinToAV(NewI(x), yv, left)
	default:
		return NewAV([]V{NewI(x), y})
	}
}

func joinToF(x float64, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAF([]float64{float64(x), float64(y.I())})
		}
		return NewAF([]float64{float64(y.I()), float64(x)})
	}
	if y.IsF() {
		if left {
			return NewAF([]float64{float64(x), float64(y.F())})
		}
		return NewAF([]float64{float64(y.F()), float64(x)})
	}
	switch yv := y.Value.(type) {
	case S:
		if left {
			return NewAV([]V{NewF(x), y})
		}
		return NewAV([]V{y, NewF(x)})
	case *AB:
		return joinToAB(NewF(x), yv, left)
	case *AF:
		return joinToAF(NewF(x), yv, left)
	case *AI:
		return joinToAI(NewF(x), yv, left)
	case *AS:
		return joinToAS(NewF(x), yv, left)
	case *AV:
		return joinToAV(NewF(x), yv, left)
	default:
		return NewAV([]V{NewF(x), y})
	}
}

func joinToS(x S, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAV([]V{NewV(x), y})
		}
		return NewAV([]V{y, NewV(x)})
	}
	if y.IsF() {
		if left {
			return NewAV([]V{NewV(x), y})
		}
		return NewAV([]V{y, NewV(x)})
	}
	switch yv := y.Value.(type) {
	case S:
		if left {
			return NewAS([]string{string(x), string(yv)})
		}
		return NewAS([]string{string(yv), string(x)})
	case *AB:
		return joinToAB(NewV(x), yv, left)
	case *AF:
		return joinToAF(NewV(x), yv, left)
	case *AI:
		return joinToAI(NewV(x), yv, left)
	case *AS:
		return joinToAS(NewV(x), yv, left)
	case *AV:
		return joinToAV(NewV(x), yv, left)
	default:
		return NewAV([]V{NewV(x), y})
	}
}

func joinToAV(x V, y *AV, left bool) V {
	switch xv := x.Value.(type) {
	case array:
		if left {
			return joinArrays(xv, y)
		}
		return joinArrays(y, xv)
	default:
		if left {
			r := make([]V, y.Len()+1)
			r[0] = x
			copy(r[1:], y.Slice)
			return NewAV(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, x)
			return NewV(y)
		}
		r := make([]V, y.Len()+1)
		r[len(r)-1] = x
		copy(r[:len(r)-1], y.Slice)
		return NewAV(r)
	}
}

func joinArrays(x, y array) V {
	// TODO: joinArrays can use reusable.
	r := make([]V, y.Len()+x.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = x.at(i)
	}
	for i := x.Len(); i < len(r); i++ {
		r[i] = y.at(i - x.Len())
	}
	return NewAV(r)
}

func joinAtomToArray(x V, y array, left bool) V {
	r := make([]V, y.Len()+1)
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
	return NewAV(r)
}

func joinToAS(x V, y *AS, left bool) V {
	switch xv := x.Value.(type) {
	case S:
		if left {
			r := make([]string, y.Len()+1)
			r[0] = string(xv)
			copy(r[1:], y.Slice)
			return NewAS(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, string(xv))
			return NewV(y)
		}
		r := make([]string, y.Len()+1)
		r[len(r)-1] = string(xv)
		copy(r[:len(r)-1], y.Slice)
		return NewAS(r)
	case *AS:
		if left {
			r := make([]string, y.Len()+xv.Len())
			copy(r[:xv.Len()], xv.Slice)
			copy(r[xv.Len():], y.Slice)
			return NewAS(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, xv.Slice...)
			return NewV(y)
		}
		r := make([]string, y.Len()+xv.Len())
		copy(r[:y.Len()], y.Slice)
		copy(r[y.Len():], xv.Slice)
		return NewAS(r)
	case array:
		if left {
			return joinArrays(xv, y)
		}
		return joinArrays(y, xv)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAB(x V, y *AB, left bool) V {
	if x.IsI() {
		if isBI(x.I()) {
			if left {
				r := make([]bool, y.Len()+1)
				r[0] = x.I() == 1
				copy(r[1:], y.Slice)
				return NewAB(r)
			}
			if y.reusable() {
				y.Slice = append(y.Slice, x.I() == 1)
				return NewV(y)
			}
			r := make([]bool, y.Len()+1)
			r[len(r)-1] = x.I() == 1
			copy(r[:len(r)-1], y.Slice)
			return NewAB(r)
		}
		r := make([]int64, y.Len()+1)
		if left {
			r[0] = x.I()
			for i := 1; i < len(r); i++ {
				r[i] = b2i(y.At(i - 1))
			}
		} else {
			r[len(r)-1] = x.I()
			for i := 0; i < len(r); i++ {
				r[i] = b2i(y.At(i))
			}
		}
		return NewAI(r)

	}
	if x.IsF() {
		r := make([]float64, y.Len()+1)
		if left {
			r[0] = x.F()
			for i := 1; i < len(r); i++ {
				r[i] = b2f(y.At(i - 1))
			}
		} else {
			r[len(r)-1] = x.F()
			for i := 0; i < len(r); i++ {
				r[i] = b2f(y.At(i))
			}
		}
		return NewAF(r)
	}
	switch xv := x.Value.(type) {
	case *AB:
		if left {
			return joinABAB(xv, y)
		}
		return joinABAB(y, xv)
	case *AI:
		if left {
			return joinAIAB(xv, y)
		}
		return joinABAI(y, xv)
	case *AF:
		if left {
			return joinAFAB(xv, y)
		}
		return joinABAF(y, xv)
	case array:
		if left {
			return joinArrays(xv, y)
		}
		return joinArrays(y, xv)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAI(x V, y *AI, left bool) V {
	if x.IsI() {
		if left {
			r := make([]int64, y.Len()+1)
			r[0] = x.I()
			copy(r[1:], y.Slice)
			return NewAI(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, x.I())
			return NewV(y)
		}
		r := make([]int64, y.Len()+1)
		r[len(r)-1] = x.I()
		copy(r[:len(r)-1], y.Slice)
		return NewAI(r)

	}
	if x.IsF() {
		r := make([]float64, y.Len()+1)
		if left {
			r[0] = x.F()
			for i := 1; i < len(r); i++ {
				r[i] = float64(y.At(i - 1))
			}
		} else {
			r[len(r)-1] = x.F()
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(y.At(i))
			}
		}
		return NewAF(r)
	}
	switch xv := x.Value.(type) {
	case *AB:
		if left {
			return joinABAI(xv, y)
		}
		return joinAIAB(y, xv)
	case *AI:
		if left {
			return joinAIAI(xv, y)
		}
		return joinAIAI(y, xv)
	case *AF:
		if left {
			return joinAFAI(xv, y)
		}
		return joinAIAF(y, xv)
	case array:
		if left {
			return joinArrays(xv, y)
		}
		return joinArrays(y, xv)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinToAF(x V, y *AF, left bool) V {
	if x.IsI() {
		if left {
			r := make([]float64, y.Len()+1)
			r[0] = float64(x.I())
			copy(r[1:], y.Slice)
			return NewAF(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, float64(x.I()))
			return NewV(y)
		}
		r := make([]float64, y.Len()+1)
		r[len(r)-1] = float64(x.I())
		copy(r[:len(r)-1], y.Slice)
		return NewAF(r)
	}
	if x.IsF() {
		if left {
			r := make([]float64, y.Len()+1)
			r[0] = x.F()
			copy(r[1:], y.Slice)
			return NewAF(r)
		}
		if y.reusable() {
			y.Slice = append(y.Slice, x.F())
			return NewV(y)
		}
		r := make([]float64, y.Len()+1)
		r[len(r)-1] = x.F()
		copy(r[:len(r)-1], y.Slice)
		return NewAF(r)
	}
	switch xv := x.Value.(type) {
	case *AB:
		if left {
			return joinABAF(xv, y)
		}
		return joinAFAB(y, xv)
	case *AI:
		if left {
			return joinAIAF(xv, y)
		}
		return joinAFAI(y, xv)
	case *AF:
		if left {
			return joinAFAF(xv, y)
		}
		return joinAFAF(y, xv)
	case array:
		if left {
			return joinArrays(xv, y)
		}
		return joinArrays(y, xv)
	default:
		return joinAtomToArray(x, y, left)
	}
}

func joinABAB(x *AB, y *AB) V {
	if x.reusable() {
		x.Slice = append(x.Slice, y.Slice...)
		return NewV(x)
	}
	r := make([]bool, y.Len()+x.Len())
	copy(r[:x.Len()], x.Slice)
	copy(r[x.Len():], y.Slice)
	return NewAB(r)
}

func joinAIAI(x *AI, y *AI) V {
	if x.reusable() {
		x.Slice = append(x.Slice, y.Slice...)
		return NewV(x)
	}
	r := make([]int64, y.Len()+x.Len())
	copy(r[:x.Len()], x.Slice)
	copy(r[x.Len():], y.Slice)
	return NewAI(r)
}

func joinAFAF(x *AF, y *AF) V {
	if x.reusable() {
		x.Slice = append(x.Slice, y.Slice...)
		return NewV(x)
	}
	r := make([]float64, y.Len()+x.Len())
	copy(r[:x.Len()], x.Slice)
	copy(r[x.Len():], y.Slice)
	return NewAF(r)
}

func joinABAI(x *AB, y *AI) V {
	r := make([]int64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = b2i(x.At(i))
	}
	copy(r[x.Len():], y.Slice)
	return NewAI(r)
}

func joinAIAB(x *AI, y *AB) V {
	if x.reusable() {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, b2i(yi))
		}
		return NewV(x)
	}
	r := make([]int64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = b2i(y.At(i - x.Len()))
	}
	return NewAI(r)
}

func joinABAF(x *AB, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(b2f(x.At(i)))
	}
	copy(r[x.Len():], y.Slice)
	return NewAF(r)
}

func joinAFAB(x *AF, y *AB) V {
	if x.reusable() {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, float64(b2i(yi)))
		}
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(b2f(y.At(i - x.Len())))
	}
	return NewAF(r)
}

func joinAIAF(x *AI, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(x.At(i))
	}
	copy(r[x.Len():], y.Slice)
	return NewAF(r)
}

func joinAFAI(x *AF, y *AI) V {
	if x.reusable() {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, float64(yi))
		}
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(y.At(i - x.Len()))
	}
	return NewAF(r)
}

// enlist returns ,x.
func enlist(x V) V {
	if x.IsI() {
		if isBI(x.I()) {
			return NewAB([]bool{x.I() == 1})
		}
		return NewAI([]int64{x.I()})
	}
	if x.IsF() {
		return NewAF([]float64{x.F()})
	}
	switch xv := x.Value.(type) {
	case S:
		return NewAS([]string{string(xv)})
	default:
		return NewAV([]V{x})
	}
}
