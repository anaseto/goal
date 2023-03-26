package goal

// joinTo returns x,y.
func joinTo(x, y V) V {
	if x.IsI() {
		return joinToI(x.I(), y)
	}
	if x.IsF() {
		return joinToF(x.F(), y)
	}
	switch xv := x.value.(type) {
	case S:
		return joinToS(xv, y)
	case *AB:
		return joinToAB(xv, y, false)
	case *AF:
		return joinToAF(xv, y, false)
	case *AI:
		return joinToAI(xv, y, false)
	case *AS:
		return joinToAS(xv, y, false)
	case *AV:
		return joinToAV(xv, y, false)
	case *Dict:
		switch yv := y.value.(type) {
		case *Dict:
			return dictArith(xv, yv, func(x, y V) V { return y })
		case array:
			return joinAtomToArray(x, yv, true)
		default:
			return NewAV([]V{x, y})
		}
	default:
		switch yv := y.value.(type) {
		case array:
			return joinAtomToArray(x, yv, true)
		default:
			return NewAV([]V{x, y})
		}
	}
}

func joinToI(x int64, y V) V {
	if y.IsI() {
		return NewAI([]int64{int64(x), y.I()})
	}
	if y.IsF() {
		return NewAF([]float64{float64(x), float64(y.F())})
	}
	left := true
	switch yv := y.value.(type) {
	case S:
		return NewAV([]V{NewI(x), y})
	case *AB:
		return joinToAB(yv, NewI(x), left)
	case *AF:
		return joinToAF(yv, NewI(x), left)
	case *AI:
		return joinToAI(yv, NewI(x), left)
	case *AS:
		return joinToAS(yv, NewI(x), left)
	case *AV:
		return joinToAV(yv, NewI(x), left)
	default:
		return NewAV([]V{NewI(x), y})
	}
}

func joinToF(x float64, y V) V {
	if y.IsI() {
		return NewAF([]float64{float64(x), float64(y.I())})
	}
	if y.IsF() {
		return NewAF([]float64{float64(x), float64(y.F())})
	}
	left := true
	switch yv := y.value.(type) {
	case S:
		return NewAV([]V{NewF(x), y})
	case *AB:
		return joinToAB(yv, NewF(x), left)
	case *AF:
		return joinToAF(yv, NewF(x), left)
	case *AI:
		return joinToAI(yv, NewF(x), left)
	case *AS:
		return joinToAS(yv, NewF(x), left)
	case *AV:
		return joinToAV(yv, NewF(x), left)
	default:
		return NewAV([]V{NewF(x), y})
	}
}

func joinToS(x S, y V) V {
	if y.IsI() {
		return NewAV([]V{NewV(x), y})
	}
	if y.IsF() {
		return NewAV([]V{NewV(x), y})
	}
	left := true
	switch yv := y.value.(type) {
	case S:
		return NewAS([]string{string(x), string(yv)})
	case *AB:
		return joinToAB(yv, NewV(x), left)
	case *AF:
		return joinToAF(yv, NewV(x), left)
	case *AI:
		return joinToAI(yv, NewV(x), left)
	case *AS:
		return joinToAS(yv, NewV(x), left)
	case *AV:
		return joinToAV(yv, NewV(x), left)
	default:
		return NewAV([]V{NewV(x), y})
	}
}

func joinToAB(x *AB, y V, left bool) V {
	if y.IsI() {
		if isBI(y.I()) {
			if left {
				r := make([]bool, x.Len()+1)
				r[0] = y.I() == 1
				copy(r[1:], x.Slice)
				return NewABWithRC(r, reuseRCp(x.rc))
			}
			if reusableRCp(x.RC()) {
				x.Slice = append(x.Slice, y.I() == 1)
				x.flags = flagNone
				return NewV(x)
			}
			r := make([]bool, x.Len()+1)
			r[len(r)-1] = y.I() == 1
			copy(r[:len(r)-1], x.Slice)
			return NewAB(r)
		}
		r := make([]int64, x.Len()+1)
		if left {
			r[0] = y.I()
			for i := 1; i < len(r); i++ {
				r[i] = B2I(x.At(i - 1))
			}
		} else {
			r[len(r)-1] = y.I()
			for i := 0; i < len(r)-1; i++ {
				r[i] = B2I(x.At(i))
			}
		}
		return NewAIWithRC(r, reuseRCp(x.rc))
	}
	if y.IsF() {
		r := make([]float64, x.Len()+1)
		if left {
			r[0] = y.F()
			for i := 1; i < len(r); i++ {
				r[i] = B2F(x.At(i - 1))
			}
		} else {
			r[len(r)-1] = y.F()
			for i := 0; i < len(r)-1; i++ {
				r[i] = B2F(x.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(x.rc))
	}
	switch yv := y.value.(type) {
	case *AB:
		// left == false
		return joinABAB(x, yv)
	case *AI:
		// left == false
		return joinABAI(x, yv)
	case *AF:
		// left == false
		return joinABAF(x, yv)
	case array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinToAI(x *AI, y V, left bool) V {
	if y.IsI() {
		if left {
			r := make([]int64, x.Len()+1)
			r[0] = y.I()
			copy(r[1:], x.Slice)
			return NewAIWithRC(r, reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, y.I())
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]int64, x.Len()+1)
		r[len(r)-1] = y.I()
		copy(r[:len(r)-1], x.Slice)
		return NewAI(r)

	}
	if y.IsF() {
		r := make([]float64, x.Len()+1)
		if left {
			r[0] = y.F()
			for i := 1; i < len(r); i++ {
				r[i] = float64(x.At(i - 1))
			}
		} else {
			r[len(r)-1] = y.F()
			for i := 0; i < len(r)-1; i++ {
				r[i] = float64(x.At(i))
			}
		}
		return NewAFWithRC(r, reuseRCp(x.rc))
	}
	switch yv := y.value.(type) {
	case *AB:
		// left == false
		return joinAIAB(x, yv)
	case *AI:
		// left == false
		return joinAIAI(x, yv)
	case *AF:
		// left == false
		return joinAIAF(x, yv)
	case array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinToAF(x *AF, y V, left bool) V {
	if y.IsI() {
		if left {
			r := make([]float64, x.Len()+1)
			r[0] = float64(y.I())
			copy(r[1:], x.Slice)
			return NewAFWithRC(r, reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, float64(y.I()))
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]float64, x.Len()+1)
		r[len(r)-1] = float64(y.I())
		copy(r[:len(r)-1], x.Slice)
		return NewAF(r)
	}
	if y.IsF() {
		if left {
			r := make([]float64, x.Len()+1)
			r[0] = y.F()
			copy(r[1:], x.Slice)
			return NewAFWithRC(r, reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, y.F())
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]float64, x.Len()+1)
		r[len(r)-1] = y.F()
		copy(r[:len(r)-1], x.Slice)
		return NewAF(r)
	}
	switch yv := y.value.(type) {
	case *AB:
		// left == false
		return joinAFAB(x, yv)
	case *AI:
		// left == false
		return joinAFAI(x, yv)
	case *AF:
		// left == false
		return joinAFAF(x, yv)
	case array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinABAB(x *AB, y *AB) V {
	if reusableRCp(x.RC()) {
		x.Slice = append(x.Slice, y.Slice...)
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]bool, y.Len()+x.Len())
	copy(r[:x.Len()], x.Slice)
	copy(r[x.Len():], y.Slice)
	return NewAB(r)
}

func joinAIAI(x *AI, y *AI) V {
	if reusableRCp(x.RC()) {
		x.Slice = append(x.Slice, y.Slice...)
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]int64, y.Len()+x.Len())
	copy(r[:x.Len()], x.Slice)
	copy(r[x.Len():], y.Slice)
	return NewAI(r)
}

func joinAFAF(x *AF, y *AF) V {
	if reusableRCp(x.RC()) {
		x.Slice = append(x.Slice, y.Slice...)
		x.flags = flagNone
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
		r[i] = B2I(x.At(i))
	}
	copy(r[x.Len():], y.Slice)
	return NewAIWithRC(r, reuseRCp(x.rc))
}

func joinAIAB(x *AI, y *AB) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, B2I(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]int64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = B2I(y.At(i - x.Len()))
	}
	return NewAI(r)
}

func joinABAF(x *AB, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(B2F(x.At(i)))
	}
	copy(r[x.Len():], y.Slice)
	return NewAFWithRC(r, reuseRCp(x.rc))
}

func joinAFAB(x *AF, y *AB) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, float64(B2I(yi)))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(B2F(y.At(i - x.Len())))
	}
	return NewAF(r)
}

func joinAIAF(x *AI, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(x.At(i))
	}
	copy(r[x.Len():], y.Slice)
	return NewAFWithRC(r, reuseRCp(x.rc))
}

func joinAFAI(x *AF, y *AI) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.Slice {
			x.Slice = append(x.Slice, float64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.Slice)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(y.At(i - x.Len()))
	}
	return NewAF(r)
}

func joinToAS(x *AS, y V, left bool) V {
	switch yv := y.value.(type) {
	case S:
		if left {
			r := make([]string, x.Len()+1)
			r[0] = string(yv)
			copy(r[1:], x.Slice)
			return NewASWithRC(r, reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, string(yv))
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]string, x.Len()+1)
		r[len(r)-1] = string(yv)
		copy(r[:len(r)-1], x.Slice)
		return NewAS(r)
	case *AS:
		// left == false
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, yv.Slice...)
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]string, x.Len()+yv.Len())
		copy(r[:x.Len()], x.Slice)
		copy(r[x.Len():], yv.Slice)
		return NewAS(r)
	case array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinToAV(x *AV, y V, left bool) V {
	switch yv := y.value.(type) {
	case *AV:
		// left == false
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, yv.Slice...)
			x.flags = flagNone
			return NewV(x)
		}
		return joinArrays(x, yv)
	case array:
		// left == false
		return joinArrays(x, yv)
	default:
		if x.Len() == 0 {
			return toArray(y)
		}
		if left {
			r := make([]V, x.Len()+1)
			r[0] = y
			copy(r[1:], x.Slice)
			return NewV(&AV{Slice: r})
		}
		if reusableRCp(x.RC()) {
			x.Slice = append(x.Slice, y)
			x.flags = flagNone
			return NewV(x)
		}
		r := make([]V, x.Len()+1)
		r[len(r)-1] = y
		copy(r[:len(r)-1], x.Slice)
		return NewV(&AV{Slice: r})
	}
}

func joinArrays(x, y array) V {
	if y.Len() == 0 {
		return NewV(x)
	}
	if x.Len() == 0 {
		return NewV(y)
	}
	r := make([]V, y.Len()+x.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = x.at(i)
	}
	for i := x.Len(); i < len(r); i++ {
		r[i] = y.at(i - x.Len())
	}
	return NewV(&AV{Slice: r})
}

func joinAtomToArray(x V, y array, left bool) V {
	yv, ok := y.(*AV)
	if ok {
		return joinToAV(yv, x, left)
	}
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
	return NewV(&AV{Slice: r})
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
	switch xv := x.value.(type) {
	case S:
		return NewAS([]string{string(xv)})
	case RefCountHolder:
		return NewAVWithRC([]V{x}, reuseRCp(xv.RC()))
	default:
		return NewAV([]V{x})
	}
}
