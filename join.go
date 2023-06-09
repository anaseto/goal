package goal

// join returns x,y.
func join(x, y V) V {
	if x.IsI() {
		return joinI(x.I(), y)
	}
	if x.IsF() {
		return joinF(x.F(), y)
	}
	switch xv := x.bv.(type) {
	case S:
		return joinS(xv, y)
	case *AB:
		return joinAB(xv, y, false)
	case *AF:
		return joinAF(xv, y, false)
	case *AI:
		return joinAI(xv, y, false)
	case *AS:
		return joinAS(xv, y, false)
	case *AV:
		return joinAV(xv, y, false)
	case *D:
		switch yv := y.bv.(type) {
		case *D:
			return dictMerge(xv, yv)
		case Array:
			return joinAtomToArray(x, yv, true)
		default:
			return newAVv([]V{x, y})
		}
	default:
		switch yv := y.bv.(type) {
		case Array:
			return joinAtomToArray(x, yv, true)
		default:
			return newAVv([]V{x, y})
		}
	}
}

func joinI(x int64, y V) V {
	if y.IsI() {
		return NewAI([]int64{int64(x), y.I()})
	}
	if y.IsF() {
		return NewAF([]float64{float64(x), float64(y.F())})
	}
	const left = true
	switch yv := y.bv.(type) {
	case S:
		return newAVv([]V{NewI(x), y})
	case *AB:
		return joinAB(yv, NewI(x), left)
	case *AF:
		return joinAF(yv, NewI(x), left)
	case *AI:
		return joinAI(yv, NewI(x), left)
	case *AS:
		return joinAS(yv, NewI(x), left)
	case *AV:
		return joinAV(yv, NewI(x), left)
	default:
		return newAVv([]V{NewI(x), y})
	}
}

func joinF(x float64, y V) V {
	if y.IsI() {
		return NewAF([]float64{float64(x), float64(y.I())})
	}
	if y.IsF() {
		return NewAF([]float64{float64(x), float64(y.F())})
	}
	const left = true
	switch yv := y.bv.(type) {
	case S:
		return newAVv([]V{NewF(x), y})
	case *AB:
		return joinAB(yv, NewF(x), left)
	case *AF:
		return joinAF(yv, NewF(x), left)
	case *AI:
		return joinAI(yv, NewF(x), left)
	case *AS:
		return joinAS(yv, NewF(x), left)
	case *AV:
		return joinAV(yv, NewF(x), left)
	default:
		return newAVv([]V{NewF(x), y})
	}
}

func joinS(x S, y V) V {
	if y.IsI() {
		return newAVv([]V{NewV(x), y})
	}
	if y.IsF() {
		return newAVv([]V{NewV(x), y})
	}
	const left = true
	switch yv := y.bv.(type) {
	case S:
		return NewAS([]string{string(x), string(yv)})
	case *AB:
		return joinAB(yv, NewV(x), left)
	case *AF:
		return joinAF(yv, NewV(x), left)
	case *AI:
		return joinAI(yv, NewV(x), left)
	case *AS:
		return joinAS(yv, NewV(x), left)
	case *AV:
		return joinAV(yv, NewV(x), left)
	default:
		return newAVv([]V{NewV(x), y})
	}
}

func joinAB(x *AB, y V, left bool) V {
	if y.IsI() {
		if isBI(y.I()) {
			var fl flags
			b := x.IsBoolean() && isbI(y.I())
			if b {
				fl = flagBool
			}
			if left {
				return NewV(&AB{elts: joinSliceLeft(x.elts, byte(y.I())), flags: fl})
			}
			if x.reusable() {
				x.elts = append(x.elts, byte(y.I()))
				x.flags = fl
				return NewV(x)
			}
			return NewV(&AB{elts: joinSlice(x.elts, byte(y.I())), flags: fl})
		}
		return NewAI(joinIsN(x.elts, y.I(), left))
	}
	if y.IsF() {
		return NewAF(joinIsN(x.elts, y.F(), left))
	}
	switch yv := y.bv.(type) {
	case *AB:
		// left == false
		return joinABAB(x, yv)
	case *AI:
		// left == false
		return joinABAI(x, yv)
	case *AF:
		// left == false
		return joinABAF(x, yv)
	case Array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinIsN[I integer, N number](x []I, y N, left bool) []N {
	r := make([]N, len(x)+1)
	if left {
		r[0] = y
		for i := 1; i < len(r); i++ {
			r[i] = N(x[i-1])
		}
	} else {
		r[len(r)-1] = y
		for i := 0; i < len(r)-1; i++ {
			r[i] = N(x[i])
		}
	}
	return r
}

func joinSlice[T any](x []T, y T) []T {
	r := make([]T, len(x)+1)
	r[len(r)-1] = y
	copy(r[:len(r)-1], x)
	return r
}

func joinSliceLeft[T any](x []T, y T) []T {
	r := make([]T, len(x)+1)
	r[0] = y
	copy(r[1:], x)
	return r
}

func joinAI(x *AI, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAI(joinSliceLeft(x.elts, y.I()))
		}
		if x.reusable() {
			x.elts = append(x.elts, y.I())
			x.flags = flagNone
			return NewV(x)
		}
		return NewAI(joinSlice(x.elts, y.I()))

	}
	if y.IsF() {
		return NewAF(joinIsN(x.elts, y.F(), left))
	}
	switch yv := y.bv.(type) {
	case *AB:
		// left == false
		return joinAIAB(x, yv)
	case *AI:
		// left == false
		return joinAIAI(x, yv)
	case *AF:
		// left == false
		return joinAIAF(x, yv)
	case Array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinAF(x *AF, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAF(joinSliceLeft(x.elts, float64(y.I())))
		}
		if x.reusable() {
			x.elts = append(x.elts, float64(y.I()))
			x.flags = flagNone
			return NewV(x)
		}
		return NewAF(joinSlice(x.elts, float64(y.I())))
	}
	if y.IsF() {
		if left {
			return NewAF(joinSliceLeft(x.elts, y.F()))
		}
		if x.reusable() {
			x.elts = append(x.elts, y.F())
			x.flags = flagNone
			return NewV(x)
		}
		return NewAF(joinSlice(x.elts, y.F()))
	}
	switch yv := y.bv.(type) {
	case *AB:
		// left == false
		return joinAFAB(x, yv)
	case *AI:
		// left == false
		return joinAFAI(x, yv)
	case *AF:
		// left == false
		return joinAFAF(x, yv)
	case Array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinABAB(x *AB, y *AB) V {
	b := x.IsBoolean() && y.IsBoolean()
	var fl flags
	if b {
		fl = flagBool
	}
	if x.reusable() {
		x.elts = append(x.elts, y.elts...)
		x.flags = fl
		return NewV(x)
	}
	return NewV(&AB{elts: joinSlices(x.elts, y.elts), flags: fl})
}

func joinSlices[T any](x, y []T) []T {
	r := make([]T, len(x)+len(y))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAIAI(x *AI, y *AI) V {
	if x.reusable() {
		x.elts = append(x.elts, y.elts...)
		x.flags = flagNone
		return NewV(x)
	}
	return NewAI(joinSlices(x.elts, y.elts))
}

func joinAFAF(x *AF, y *AF) V {
	if x.reusable() {
		x.elts = append(x.elts, y.elts...)
		x.flags = flagNone
		return NewV(x)
	}
	return NewAF(joinSlices(x.elts, y.elts))
}

func joinABAI(x *AB, y *AI) V {
	return NewAI(joinSliceToNums(x.elts, y.elts))
}

func joinSliceToNums[N number, M number](x []N, y []M) []M {
	r := make([]M, len(x)+len(y))
	for i := 0; i < len(x); i++ {
		r[i] = M(x[i])
	}
	copy(r[len(x):], y)
	return r
}

func joinAIAB(x *AI, y *AB) V {
	if x.reusable() {
		for _, yi := range y.elts {
			x.elts = append(x.elts, int64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	return NewAI(joinNumsToSlice(x.elts, y.elts))
}

func joinNumsToSlice[N number, M number](x []N, y []M) []N {
	r := make([]N, len(x)+len(y))
	copy(r[:len(x)], x)
	for i := len(x); i < len(r); i++ {
		r[i] = N(y[i-len(x)])
	}
	return r
}

func joinABAF(x *AB, y *AF) V {
	return NewAF(joinSliceToNums(x.elts, y.elts))
}

func joinAFAB(x *AF, y *AB) V {
	if x.reusable() {
		for _, yi := range y.elts {
			x.elts = append(x.elts, float64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	return NewAF(joinNumsToSlice(x.elts, y.elts))
}

func joinAIAF(x *AI, y *AF) V {
	return NewAF(joinSliceToNums(x.elts, y.elts))
}

func joinAFAI(x *AF, y *AI) V {
	if x.reusable() {
		for _, yi := range y.elts {
			x.elts = append(x.elts, float64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	return NewAF(joinNumsToSlice(x.elts, y.elts))
}

func joinAS(x *AS, y V, left bool) V {
	switch yv := y.bv.(type) {
	case S:
		if left {
			return NewAS(joinSliceLeft(x.elts, string(yv)))
		}
		if x.reusable() {
			x.elts = append(x.elts, string(yv))
			x.flags = flagNone
			return NewV(x)
		}
		return NewAS(joinSlice(x.elts, string(yv)))
	case *AS:
		// left == false
		if x.reusable() {
			x.elts = append(x.elts, yv.elts...)
			x.flags = flagNone
			return NewV(x)
		}
		return NewAS(joinSlices(x.elts, yv.elts))
	case Array:
		// left == false
		return joinArrays(x, yv)
	default:
		return joinAtomToArray(y, x, left)
	}
}

func joinAV(x *AV, y V, left bool) V {
	switch yv := y.bv.(type) {
	case *AV:
		// left == false
		if x.reusable() {
			x.elts = append(x.elts, yv.elts...)
			x.flags = flagNone
			return NewV(x)
		}
		return joinArrays(x, yv)
	case Array:
		// left == false
		return joinArrays(x, yv)
	default:
		if x.Len() == 0 {
			return toArray(y)
		}
		y.MarkImmutable()
		if left {
			return NewV(&AV{elts: joinSliceLeft(x.elts, y)})
		}
		if x.reusable() {
			x.elts = append(x.elts, y)
			x.flags = flagNone
			return NewV(x)
		}
		return NewV(&AV{elts: joinSlice(x.elts, y)})
	}
}

func joinArrays(x, y Array) V {
	if y.Len() == 0 {
		return NewV(x)
	}
	if x.Len() == 0 {
		return NewV(y)
	}
	r := make([]V, y.Len()+x.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = x.VAt(i)
	}
	for i := x.Len(); i < len(r); i++ {
		r[i] = y.VAt(i - x.Len())
	}
	return NewV(&AV{elts: r})
}

func joinAtomToArray(x V, y Array, left bool) V {
	yv, ok := y.(*AV)
	if ok {
		return joinAV(yv, x, left)
	}
	r := make([]V, y.Len()+1)
	if left {
		r[0] = x
		x.MarkImmutable()
		for i := 1; i < len(r); i++ {
			r[i] = y.VAt(i - 1)
		}
	} else {
		r[len(r)-1] = x
		x.MarkImmutable()
		for i := 0; i < len(r)-1; i++ {
			r[i] = y.VAt(i)
		}
	}
	return NewV(&AV{elts: r})
}

// enlist returns ,x.
func enlist(x V) V {
	if x.IsI() {
		if isBI(x.I()) {
			b := isbI(x.I())
			var fl flags
			if b {
				fl = flagBool
			}
			return NewV(&AB{elts: []byte{byte(x.I())}, flags: fl})
		}
		return NewAI([]int64{x.I()})
	}
	if x.IsF() {
		return NewAF([]float64{x.F()})
	}
	switch xv := x.bv.(type) {
	case S:
		return NewAS([]string{string(xv)})
	default:
		x.MarkImmutable()
		return newAVu([]V{x})
	}
}
