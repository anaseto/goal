package goal

func encodeBaseDigits(b int64, x int64) int {
	if x < 0 {
		x = -x
	}
	n := 1
	for x >= b {
		x /= b
		n++
	}
	return n
}

func encode(f V, x V) V {
	if f.IsI() {
		if f.I() <= 1 {
			return panics("i\\x : base i is not > 1")
		}
		if x.IsI() {
			n := encodeBaseDigits(f.I(), x.I())
			r := make([]int64, n)
			for i := n - 1; i >= 0; i-- {
				r[i] = x.I() % f.I()
				x.n /= f.I()
			}
			return NewAI(r)
		}
		if x.IsF() {
			if !isI(x.F()) {
				return Panicf("i\\x : x non-integer (%g)", x.F())
			}
			return encode(f, NewI(int64(x.F())))
		}
		switch xv := x.value.(type) {
		case *AI:
			a, n := encodeIIs(f.I(), xv.elts)
			r := make([]V, n)
			for i := range r {
				r[i] = NewAI(a[i*xv.Len() : (i+1)*xv.Len()])
			}
			return NewAV(r)
		case *AB:
			a, n := encodeIIs(f.I(), xv.elts)
			r := make([]V, n)
			for i := range r {
				r[i] = NewAB(a[i*xv.Len() : (i+1)*xv.Len()])
			}
			return NewAV(r)
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("i\\x : i ", aix)
			}
			return encode(f, aix)
		case *AV:
			r := make([]V, xv.Len())
			for i, xi := range xv.elts {
				r[i] = encode(f, xi)
				if r[i].IsPanic() {
					return r[i]
				}
			}
			return canonicalFast(NewAV(r))
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
	switch fv := f.value.(type) {
	case *AB:
		return encode(fromABtoAI(fv), x)
	case *AI:
		for _, b := range fv.elts {
			if b <= 1 {
				return panics("I\\x : I contains base < 2")
			}
		}
		if x.IsI() {
			n := fv.Len()
			r := make([]int64, n)
			for i := n - 1; i >= 0 && x.I() > 0; i-- {
				fi := fv.At(i)
				r[i] = x.I() % fi
				x.n /= fi
			}
			return NewAI(r)
		}
		if x.IsF() {
			if !isI(x.F()) {
				return Panicf("I/x : x non-integer (%g)", x.F())
			}
			return encode(f, NewI(int64(x.F())))
		}
		switch xv := x.value.(type) {
		case *AI:
			n := fv.Len()
			ai := make([]int64, n*xv.Len())
			copy(ai[(n-1)*xv.Len():], xv.elts)
			for i := n - 1; i >= 0; i-- {
				for j := 0; j < xv.Len(); j++ {
					fi := fv.At(i)
					ox := ai[i*xv.Len()+j]
					ai[i*xv.Len()+j] = ox % fi
					if i > 0 {
						ai[(i-1)*xv.Len()+j] = ox / fi
					}
				}
			}
			r := make([]V, n)
			for i := range r {
				r[i] = NewAI(ai[i*xv.Len() : (i+1)*xv.Len()])
			}
			return NewAV(r)
		case *AB:
			return encode(f, fromABtoAI(xv))
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("I\\x : I ", aix)
			}
			return encode(f, aix)
		case *AV:
			r := make([]V, xv.Len())
			for i, xi := range xv.elts {
				r[i] = encode(f, xi)
				if r[i].IsPanic() {
					return r[i]
				}
			}
			return canonicalFast(NewAV(r))
		default:
			return panicType("I\\x", "x", x)
		}
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

func encodeIIs[I integer](f int64, x []I) ([]I, int) {
	min, max := minMaxIntegers(x)
	max = I(maxI(absI(int64(min)), absI(int64(max))))
	n := encodeBaseDigits(f, int64(max))
	a := make([]I, n*len(x))
	copy(a[(n-1)*len(x):], x)
	for i := n - 1; i >= 0; i-- {
		for j := 0; j < len(x); j++ {
			ox := a[i*len(x)+j]
			a[i*len(x)+j] = ox % I(f)
			if i > 0 {
				a[(i-1)*len(x)+j] = ox / I(f)
			}
		}
	}
	return a, n
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
		switch xv := x.value.(type) {
		case *AI:
			var r, n int64 = 0, 1
			for i := xv.Len() - 1; i >= 0; i-- {
				r += xv.At(i) * n
				n *= f.I()
			}
			return NewI(r)
		case *AB:
			var r, n int64 = 0, 1
			for i := xv.Len() - 1; i >= 0; i-- {
				r += int64(xv.At(i)) * n
				n *= f.I()
			}
			return NewI(r)
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("i/x : i ", aix)
			}
			return decode(f, aix)
		case *AV:
			r := make([]V, xv.Len())
			for i, xi := range xv.elts {
				r[i] = decode(f, xi)
				if r[i].IsPanic() {
					return r[i]
				}
			}
			return canonicalFast(NewAV(r))
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
	switch fv := f.value.(type) {
	case *AB:
		return decode(fromABtoAI(fv), x)
	case *AI:
		for _, b := range fv.elts {
			if b <= 0 {
				return panics("I/x : I contains non positive")
			}
		}
		if x.IsI() {
			var r, n int64 = 0, 1
			for i := fv.Len() - 1; i >= 0; i-- {
				r += x.I() * n
				n *= fv.At(i)
			}
			return NewI(r)
		}
		if x.IsF() {
			if !isI(x.F()) {
				return Panicf("I/x : x non-integer (%g)", x.F())
			}
			return decode(f, NewI(int64(x.F())))
		}
		switch xv := x.value.(type) {
		case *AI:
			if fv.Len() != xv.Len() {
				return panicLength("I/x", fv.Len(), xv.Len())
			}
			var r, n int64 = 0, 1
			for i := xv.Len() - 1; i >= 0; i-- {
				r += xv.At(i) * n
				n *= fv.At(i)
			}
			return NewI(r)
		case *AB:
			return decode(f, fromABtoAI(xv))
		case *AF:
			aix := toAI(xv)
			if aix.IsPanic() {
				return ppanic("I/x : I ", aix)
			}
			return decode(f, aix)
		case *AV:
			r := make([]V, xv.Len())
			for i, xi := range xv.elts {
				r[i] = decode(f, xi)
				if r[i].IsPanic() {
					return r[i]
				}
			}
			return canonicalFast(NewAV(r))
		default:
			return panicType("I/x", "x", x)
		}
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
