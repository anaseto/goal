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
			return dictMerge(xv, yv)
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
	const left = true
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
	const left = true
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
	const left = true
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
			var fl flags
			b := x.IsBoolean() && isbI(y.I())
			if b {
				fl = flagBool
			}
			if left {
				return NewV(&AB{elts: joinToAnyLeft(x.elts, byte(y.I())), rc: reuseRCp(x.rc), flags: fl})
			}
			if reusableRCp(x.RC()) {
				x.elts = append(x.elts, byte(y.I()))
				x.flags = fl
				return NewV(x)
			}
			return NewV(&AB{elts: joinToAny(x.elts, byte(y.I())), flags: fl})
		}
		return NewAIWithRC(joinToIntegersN(x.elts, y.I(), left), reuseRCp(x.rc))
	}
	if y.IsF() {
		return NewAFWithRC(joinToIntegersN(x.elts, y.F(), left), reuseRCp(x.rc))
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

func joinToIntegersN[I integer, N number](x []I, y N, left bool) []N {
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

func joinToAny[T any](x []T, y T) []T {
	r := make([]T, len(x)+1)
	r[len(r)-1] = y
	copy(r[:len(r)-1], x)
	return r
}

func joinToAnyLeft[T any](x []T, y T) []T {
	r := make([]T, len(x)+1)
	r[0] = y
	copy(r[1:], x)
	return r
}

func joinToAI(x *AI, y V, left bool) V {
	if y.IsI() {
		if left {
			return NewAIWithRC(joinToAnyLeft(x.elts, y.I()), reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, y.I())
			x.flags = flagNone
			return NewV(x)
		}
		return NewAI(joinToAny(x.elts, y.I()))

	}
	if y.IsF() {
		return NewAFWithRC(joinToIntegersN(x.elts, y.F(), left), reuseRCp(x.rc))
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
			return NewAFWithRC(joinToAnyLeft(x.elts, float64(y.I())), reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, float64(y.I()))
			x.flags = flagNone
			return NewV(x)
		}
		return NewAF(joinToAny(x.elts, float64(y.I())))
	}
	if y.IsF() {
		if left {
			return NewAFWithRC(joinToAnyLeft(x.elts, y.F()), reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, y.F())
			x.flags = flagNone
			return NewV(x)
		}
		return NewAF(joinToAny(x.elts, y.F()))
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
	b := x.IsBoolean() && y.IsBoolean()
	var fl flags
	if b {
		fl = flagBool
	}
	if reusableRCp(x.RC()) {
		x.elts = append(x.elts, y.elts...)
		x.flags = fl
		return NewV(x)
	}
	return NewV(&AB{elts: joinAnyAny(x.elts, y.elts), flags: fl})
}

func joinAnyAny[T any](x, y []T) []T {
	r := make([]T, len(x)+len(y))
	copy(r[:len(x)], x)
	copy(r[len(x):], y)
	return r
}

func joinAIAI(x *AI, y *AI) V {
	if reusableRCp(x.RC()) {
		x.elts = append(x.elts, y.elts...)
		x.flags = flagNone
		return NewV(x)
	}
	return NewAI(joinAnyAny(x.elts, y.elts))
}

func joinAFAF(x *AF, y *AF) V {
	if reusableRCp(x.RC()) {
		x.elts = append(x.elts, y.elts...)
		x.flags = flagNone
		return NewV(x)
	}
	return NewAF(joinAnyAny(x.elts, y.elts))
}

func joinABAI(x *AB, y *AI) V {
	r := make([]int64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = int64(x.At(i))
	}
	copy(r[x.Len():], y.elts)
	return NewAIWithRC(r, reuseRCp(x.rc))
}

func joinAIAB(x *AI, y *AB) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.elts {
			x.elts = append(x.elts, int64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]int64, x.Len()+y.Len())
	copy(r[:x.Len()], x.elts)
	for i := x.Len(); i < len(r); i++ {
		r[i] = int64(y.At(i - x.Len()))
	}
	return NewAI(r)
}

func joinABAF(x *AB, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(x.At(i))
	}
	copy(r[x.Len():], y.elts)
	return NewAFWithRC(r, reuseRCp(x.rc))
}

func joinAFAB(x *AF, y *AB) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.elts {
			x.elts = append(x.elts, float64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.elts)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(y.At(i - x.Len()))
	}
	return NewAF(r)
}

func joinAIAF(x *AI, y *AF) V {
	r := make([]float64, x.Len()+y.Len())
	for i := 0; i < x.Len(); i++ {
		r[i] = float64(x.At(i))
	}
	copy(r[x.Len():], y.elts)
	return NewAFWithRC(r, reuseRCp(x.rc))
}

func joinAFAI(x *AF, y *AI) V {
	if reusableRCp(x.RC()) {
		for _, yi := range y.elts {
			x.elts = append(x.elts, float64(yi))
		}
		x.flags = flagNone
		return NewV(x)
	}
	r := make([]float64, x.Len()+y.Len())
	copy(r[:x.Len()], x.elts)
	for i := x.Len(); i < len(r); i++ {
		r[i] = float64(y.At(i - x.Len()))
	}
	return NewAF(r)
}

func joinToAS(x *AS, y V, left bool) V {
	switch yv := y.value.(type) {
	case S:
		if left {
			return NewASWithRC(joinToAnyLeft(x.elts, string(yv)), reuseRCp(x.rc))
		}
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, string(yv))
			x.flags = flagNone
			return NewV(x)
		}
		return NewAS(joinToAny(x.elts, string(yv)))
	case *AS:
		// left == false
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, yv.elts...)
			x.flags = flagNone
			return NewV(x)
		}
		return NewAS(joinAnyAny(x.elts, yv.elts))
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
			x.elts = append(x.elts, yv.elts...)
			x.flags = flagNone
			x.rc = nil
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
			return NewV(&AV{elts: joinToAnyLeft(x.elts, y)})
		}
		if reusableRCp(x.RC()) {
			x.elts = append(x.elts, y)
			y.InitWithRC(x.RC())
			x.flags = flagNone
			return NewV(x)
		}
		return NewV(&AV{elts: joinToAny(x.elts, y)})
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
	return NewV(&AV{elts: r})
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
	switch xv := x.value.(type) {
	case S:
		return NewAS([]string{string(xv)})
	case RefCountHolder:
		return NewAVWithRC([]V{x}, reuseRCp(xv.RC()))
	default:
		return NewAV([]V{x})
	}
}
