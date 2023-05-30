package goal

func encodeBaseDigits(b int64, x int64) int {
	n := 1
	for x >= b {
		x /= b
		n++
	}
	return n
}

func encodeIIv(f int64, x int64) V {
	if x < 0 {
		return Panicf(`i\i : negative integer (%d)`, x)
	}
	if f <= 256 {
		r := encodeII[byte](f, x)
		var fl flags
		if f == 2 {
			fl |= flagBool
		}
		return NewV(&AB{elts: r, flags: fl})
	}
	r := encodeII[int64](f, x)
	return NewAI(r)
}

func encodeII[I integer](f int64, x int64) []I {
	n := encodeBaseDigits(f, x)
	r := make([]I, n)
	for i := n - 1; i >= 0; i-- {
		r[i] = I(x % f)
		x /= f
	}
	return r
}

func encodeIInts(f int64, x []int64) V {
	min, max := minMaxIs(x)
	if min < 0 {
		return Panicf(`i\I : negative integer (%d)`, min)
	}
	n := encodeBaseDigits(f, max)
	a := make([]int64, n*len(x))
	copy(a[(n-1)*len(x):], x)
	encodeIIs(f, a, len(x), n)
	r := make([]V, n)
	for i := range r {
		r[i] = NewV(&AI{elts: a[i*len(x) : (i+1)*len(x)], flags: flagImmutable})
	}
	return newAVu(r)
}

func encodeIBytes(f int64, x []byte) V {
	max := int64(maxBytes(x))
	n := encodeBaseDigits(f, max)
	a := make([]byte, n*len(x))
	copy(a[(n-1)*len(x):], x)
	encodeIIs(f, a, len(x), n)
	r := make([]V, n)
	var fl flags
	if f == 2 {
		fl |= flagBool
	}
	for i := range r {
		r[i] = NewV(&AB{elts: a[i*len(x) : (i+1)*len(x)], flags: fl | flagImmutable})
	}
	return newAVu(r)
}

func encodeIIs[I integer](f int64, a []I, cols, n int) {
	for i := n - 1; i >= 0; i-- {
		for j := 0; j < cols; j++ {
			ox := a[i*cols+j]
			a[i*cols+j] = I(int64(ox) % f)
			if i > 0 {
				a[(i-1)*cols+j] = I(int64(ox) / f)
			}
		}
	}
}

func encodeIsIv[I integer](f []I, x int64) V {
	max := maxIntegers(f)
	if max <= 256 {
		r := encodeIsI[I, byte](f, x)
		return NewAB(r)
	}
	r := encodeIsI[I, int64](f, x)
	return NewAI(r)
}

func encodeIsI[I integer, J integer](f []I, x int64) []J {
	n := len(f)
	r := make([]J, n)
	for i := n - 1; i >= 0; i-- {
		fi := int64(f[i])
		r[i] = J(x % fi)
		x /= fi
	}
	return r
}

func encodeIsBytes[I integer](f []I, x []byte) V {
	n := len(f)
	a := make([]byte, n*len(x))
	copy(a[(n-1)*len(x):], x)
	encodeIsIs(f, a, len(x))
	r := make([]V, n)
	var fl flags
	min, max := minMaxIs(f)
	if min == 2 && max == 2 {
		fl |= flagBool
	}
	for i := range r {
		r[i] = NewV(&AB{elts: a[i*len(x) : (i+1)*len(x)], flags: fl | flagImmutable})
	}
	return newAVu(r)
}

func encodeIsInts[I integer](f []I, x []int64) V {
	min := minIntegers(x)
	if min < 0 {
		return Panicf(`I\I : negative integer (%d)`, min)
	}
	n := len(f)
	a := make([]int64, n*len(x))
	copy(a[(n-1)*len(x):], x)
	encodeIsIs(f, a, len(x))
	r := make([]V, n)
	for i := range r {
		r[i] = NewV(&AI{elts: a[i*len(x) : (i+1)*len(x)], flags: flagImmutable})
	}
	return newAVu(r)
}

func encodeIsIs[I integer, J integer](f []I, a []J, cols int) {
	for i := len(f) - 1; i >= 0; i-- {
		fi := int64(f[i])
		for j := 0; j < cols; j++ {
			ox := a[i*cols+j]
			a[i*cols+j] = J(int64(ox) % fi)
			if i > 0 {
				a[(i-1)*cols+j] = J(int64(ox) / fi)
			}
		}
	}
}

func encode(f V, x V) V {
	if f.IsI() {
		if f.I() <= 1 {
			return panics("i\\x : base i is not > 1")
		}
		if x.IsI() {
			return encodeIIv(f.I(), x.I())
		}
		if x.IsF() {
			if !isI(x.F()) {
				return Panicf("i\\x : x non-integer (%g)", x.F())
			}
			return encode(f, NewI(int64(x.F())))
		}
		switch xv := x.bv.(type) {
		case *AI:
			return encodeIInts(f.I(), xv.elts)
		case *AB:
			return encodeIBytes(f.I(), xv.elts)
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("i\\x : i ", aix)
			}
			return encode(f, aix)
		case *AV:
			return Canonical(monadAV(xv, func(xi V) V { return encode(f, xi) }))
		default:
			return panicType("i\\x", "x", x)
		}
	}
	if f.IsF() {
		if !isI(f.F()) {
			return Panicf("i\\x : i non-integer (%g)", f.F())
		}
		return encode(NewI(int64(f.F())), x)
	}
	switch fv := f.bv.(type) {
	case *AB:
		return encodeIs(fv.elts, x)
	case *AI:
		return encodeIs(fv.elts, x)
	case *AF:
		aif := toAI(fv)
		if aif.IsPanic() {
			return aif
		}
		return encode(aif, x)
	default:
		// should not happen
		return panicType("I\\x", "I", f)
	}
}

func encodeIs[I integer](f []I, x V) V {
	for _, b := range f {
		if b <= 1 {
			return panics("I\\x : I contains base < 2")
		}
	}
	if x.IsI() {
		return encodeIsIv(f, x.I())
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("I/x : x non-integer (%g)", x.F())
		}
		return encodeIs(f, NewI(int64(x.F())))
	}
	switch xv := x.bv.(type) {
	case *AI:
		return encodeIsInts(f, xv.elts)
	case *AB:
		return encodeIsBytes(f, xv.elts)
	case *AF:
		aix := toAI(xv)
		if aix.IsPanic() {
			return ppanic("I\\x : I ", aix)
		}
		return encodeIs(f, aix)
	case *AV:
		return Canonical(monadAV(xv, func(xi V) V { return encodeIs(f, xi) }))
	default:
		return panicType("I\\x", "x", x)
	}
}

func decode(f V, x V) V {
	if f.IsI() {
		if f.I() <= 0 {
			return panics("i/x : base i is not positive")
		}
		if x.IsI() {
			return x
		}
		if x.IsF() {
			if !isI(x.F()) {
				return Panicf("i/x : x non-integer (%g)", x.F())
			}
			return NewI(int64(x.F()))
		}
		switch xv := x.bv.(type) {
		case *AI:
			return NewI(decodeIIs(f.I(), xv.elts))
		case *AB:
			return NewI(decodeIIs(f.I(), xv.elts))
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("i/x : i ", aix)
			}
			return decode(f, aix)
		case *AV:
			return Canonical(monadAV(xv, func(xi V) V { return decode(f, xi) }))
		default:
			return panicType("i/x", "x", x)
		}

	}
	if f.IsF() {
		if !isI(f.F()) {
			return Panicf("i/x : i non-integer (%g)", f.F())
		}
		return decode(NewI(int64(f.F())), x)
	}
	switch fv := f.bv.(type) {
	case *AB:
		return decodeIs(fv.elts, x)
	case *AI:
		return decodeIs(fv.elts, x)
	case *AF:
		aif := toAI(fv)
		if aif.IsPanic() {
			return aif
		}
		return decode(aif, x)
	default:
		// should not happen
		return panicType("I/x", "I", f)
	}
}

func decodeIs[I integer](f []I, x V) V {
	for _, b := range f {
		if b <= 0 {
			return panics("I/x : I contains non positive")
		}
	}
	if x.IsI() {
		return NewI(decodeIsI(f, x.I()))
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("I/x : x non-integer (%g)", x.F())
		}
		return decodeIs(f, NewI(int64(x.F())))
	}
	switch xv := x.bv.(type) {
	case *AI:
		if len(f) != xv.Len() {
			return panicLength("I/x", len(f), xv.Len())
		}
		return NewI(decodeIsIs(f, xv.elts))
	case *AB:
		if len(f) != xv.Len() {
			return panicLength("I/x", len(f), xv.Len())
		}
		return NewI(decodeIsIs(f, xv.elts))
	case *AF:
		aix := toAI(xv)
		if aix.IsPanic() {
			return ppanic("I/x : I ", aix)
		}
		return decodeIs(f, aix)
	case *AV:
		return Canonical(monadAV(xv, func(xi V) V { return decodeIs(f, xi) }))
	default:
		return panicType("I/x", "x", x)
	}
}

func decodeIsIs[I integer, J integer](f []I, x []J) int64 {
	var r, n int64 = 0, 1
	for i := len(x) - 1; i >= 0; i-- {
		r += int64(x[i]) * n
		n *= int64(f[i])
	}
	return r
}

func decodeIIs[I integer, J integer](f I, x []J) int64 {
	var r, n int64 = 0, 1
	for i := len(x) - 1; i >= 0; i-- {
		r += int64(x[i]) * n
		n *= int64(f)
	}
	return r
}

func decodeIsI[I integer, J integer](f []I, x J) int64 {
	var r, n int64 = 0, 1
	for i := len(f) - 1; i >= 0; i-- {
		r += int64(x) * n
		n *= int64(f[i])
	}
	return r
}
