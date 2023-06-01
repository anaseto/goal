package goal

// flip returns +x.
func flip(x V) V {
	if xv, ok := x.bv.(*D); ok {
		return NewV(&D{keys: xv.values, values: xv.keys})
	}
	x = toArray(x)
	switch xv := x.bv.(type) {
	case *AV:
		if xv.Len() == 0 {
			x.MarkImmutable()
			return newAVu([]V{x})
		}
		l := -1
		for _, xi := range xv.elts {
			xia, ok := xi.bv.(Array)
			if !ok {
				x = extendAV(xv)
				return flip(x)
			}
			nl := xia.Len()
			switch {
			case l == -1:
				l = nl
			case nl != l:
				x = extendAV(xv)
				return flip(x)
			}
		}
		t := eType(xv)
		switch {
		case l == 0:
			x.MarkImmutable()
			return newAVu([]V{x})
		case l == 1:
			switch t {
			case tAb:
				return newAVu([]V{flipAB(xv, true)})
			case tAB:
				return newAVu([]V{flipAB(xv, false)})
			case tAF:
				return newAVu([]V{flipAF(xv)})
			case tAI:
				return newAVu([]V{flipAI(xv)})
			case tAS:
				return newAVu([]V{flipAS(xv)})
			default:
				return newAVu([]V{flipAV(xv)})
			}
		default:
			switch t {
			case tAb:
				return flipAVAB(xv, l, true)
			case tAB:
				return flipAVAB(xv, l, false)
			case tAF:
				return flipAVAF(xv, l)
			case tAI:
				return flipAVAI(xv, l)
			case tAS:
				return flipAVAS(xv, l)
			default:
				return flipAVAV(xv, l)
			}
		}
	default:
		x.MarkImmutable()
		return newAVu([]V{x})
	}
}

func extendAV(x *AV) V {
	n := 0
	for _, xi := range x.elts {
		xl := xi.Len()
		if n < xl {
			n = xl
		}
	}
	r := x.reuse()
	for i, xi := range x.elts {
		xia, ok := xi.bv.(Array)
		if ok {
			if xia.Len() == n {
				r.elts[i] = xi
				continue
			}
			ri := takeN(int64(n), xia)
			ri.MarkImmutable()
			r.elts[i] = ri
			continue
		}
		ri := takeNAtom(int64(n), xi)
		ri.MarkImmutable()
		r.elts[i] = ri
	}
	return NewV(r)
}

// getAB retrieves the *getAB value. It assumes Value type is *getAB.
func (x V) getAB() *AB {
	return x.bv.(*AB)
}

func flipAB(x *AV, b bool) V {
	r := make([]byte, x.Len())
	for i, xi := range x.elts {
		r[i] = xi.getAB().At(0)
	}
	var fl flags
	if b {
		fl |= flagBool
	}
	return NewV(&AB{elts: r, flags: fl | flagImmutable})
}

func flipAVAB(x *AV, lines int, b bool) V {
	r := make([]V, lines)
	a := make([]byte, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		xia := xi.getAB()
		for j := 0; j < lines; j++ {
			a[i+j*xlen] = xia.At(j)
		}
	}
	var fl flags
	if b {
		fl |= flagBool
	}
	for j := range r {
		r[j] = NewV(&AB{elts: a[j*xlen : (j+1)*xlen], flags: fl | flagImmutable})
	}
	return newAVu(r)
}

func flipAF(x *AV) V {
	r := make([]float64, x.Len())
	for i, xi := range x.elts {
		switch xiv := xi.bv.(type) {
		case *AB:
			r[i] = float64(xiv.At(0))
		case *AF:
			r[i] = xiv.At(0)
		case *AI:
			r[i] = float64(xiv.At(0))
		}
	}
	return NewV(&AF{elts: r, flags: flagImmutable})
}

func flipAVAF(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]float64, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		switch xiv := xi.bv.(type) {
		case *AB:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = float64(xiv.At(j))
			}
		case *AF:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xiv.At(j)
			}
		case *AI:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = float64(xiv.At(j))
			}
		}
	}
	for j := range r {
		r[j] = NewV(&AF{elts: a[j*xlen : (j+1)*xlen], flags: flagImmutable})
	}
	return newAVu(r)
}

func flipAI(x *AV) V {
	r := make([]int64, x.Len())
	for i, xi := range x.elts {
		switch xiv := xi.bv.(type) {
		case *AB:
			r[i] = int64(xiv.At(0))
		case *AI:
			r[i] = xiv.At(0)
		}
	}
	return NewV(&AI{elts: r, flags: flagImmutable})
}

func flipAVAI(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]int64, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		switch xiv := xi.bv.(type) {
		case *AB:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = int64(xiv.At(j))
			}
		case *AI:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xiv.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewV(&AI{elts: a[j*xlen : (j+1)*xlen], flags: flagImmutable})
	}
	return newAVu(r)
}

func flipAS(x *AV) V {
	r := make([]string, x.Len())
	for i, xi := range x.elts {
		xiv := xi.bv.(*AS)
		r[i] = xiv.At(0)
	}
	return NewV(&AS{elts: r, flags: flagImmutable})
}

func flipAVAS(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]string, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		xiv := xi.bv.(*AS)
		for j := 0; j < lines; j++ {
			a[i+j*xlen] = xiv.At(j)
		}
	}
	for j := range r {
		r[j] = NewV(&AS{elts: a[j*xlen : (j+1)*xlen], flags: flagImmutable})
	}
	return newAVu(r)
}

func flipAV(x *AV) V {
	r := make([]V, x.Len())
	for i, xi := range x.elts {
		xiv := xi.bv.(Array)
		r[i] = xiv.VAt(0)
	}
	return canonicalAVImmut(&AV{elts: r, flags: flagImmutable})
}

func flipAVAV(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]V, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		xiv := xi.bv.(Array)
		for j := 0; j < lines; j++ {
			a[i+j*xlen] = xiv.VAt(j)
		}
	}
	for j := range r {
		r[j] = canonicalAVImmut(&AV{elts: a[j*xlen : (j+1)*xlen], flags: flagImmutable})
	}
	return newAVu(r)
}
