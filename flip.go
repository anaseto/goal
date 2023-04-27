package goal

// flip returns +x.
func flip(x V) V {
	if xv, ok := x.value.(*Dict); ok {
		return NewV(&Dict{keys: xv.values, values: xv.keys})
	}
	x = toArray(x)
	switch xv := x.value.(type) {
	case *AV:
		cols := xv.Len()
		if cols == 0 {
			return NewAV([]V{x})
		}
		lines := -1
		for _, o := range xv.elts {
			var nl int
			switch a := o.value.(type) {
			case array:
				nl = a.Len()
			default:
				continue
			}
			switch {
			case lines < 0:
				lines = nl
			case nl != lines:
				return Panicf("line length mismatch: %d vs %d", nl, lines)
			}
		}
		t := rType(xv)
		switch {
		case lines <= 0:
			return NewAV([]V{x})
		case lines == 1:
			switch t {
			case tB, tAB:
				return NewAV([]V{flipAB(xv)})
			case tF, tAF:
				return NewAV([]V{flipAF(xv)})
			case tI, tAI:
				return NewAV([]V{flipAI(xv)})
			case tS, tAS:
				return NewAV([]V{flipAS(xv)})
			default:
				return NewAV([]V{flipAV(xv)})
			}
		default:
			switch t {
			case tB, tAB:
				return flipAVAB(xv, lines)
			case tF, tAF:
				return flipAVAF(xv, lines)
			case tI, tAI:
				return flipAVAI(xv, lines)
			case tS, tAS:
				return flipAVAS(xv, lines)
			default:
				return flipAVAV(xv, lines)
			}
		}
	default:
		return NewAV([]V{x})
	}
}

// getAB retrieves the *getAB value. It assumes Value type is *getAB.
func (x V) getAB() *AB {
	return x.value.(*AB)
}

func flipAB(x *AV) V {
	r := make([]bool, x.Len())
	for i, xi := range x.elts {
		if xi.IsI() {
			r[i] = xi.I() == 1
		} else {
			r[i] = xi.getAB().At(0)
		}
	}
	return NewAB(r)
}

func flipAVAB(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]bool, lines*x.Len())
	xlen := x.Len()
	var n int = 2
	rc := &n
	for i, xi := range x.elts {
		if xi.IsI() {
			b := xi.I() == 1
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = b
			}
		} else {
			xia := xi.getAB()
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xia.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewABWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	var rn int
	return NewAVWithRC(r, &rn)
}

func flipAF(x *AV) V {
	r := make([]float64, x.Len())
	for i, xi := range x.elts {
		if xi.IsI() {
			r[i] = float64(xi.I())
			continue
		}
		if xi.IsF() {
			r[i] = float64(xi.F())
			continue
		}
		switch xiv := xi.value.(type) {
		case *AB:
			r[i] = float64(B2F(xiv.At(0)))
		case *AF:
			r[i] = xiv.At(0)
		case *AI:
			r[i] = float64(xiv.At(0))
		}
	}
	return NewAF(r)
}

func flipAVAF(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]float64, lines*x.Len())
	xlen := x.Len()
	var n int = 2
	rc := &n
	for i, xi := range x.elts {
		if xi.IsI() {
			f := float64(xi.I())
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = f
			}
			continue
		}
		if xi.IsF() {
			f := xi.F()
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = f
			}
			continue
		}
		switch xiv := xi.value.(type) {
		case *AB:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = float64(B2F(xiv.At(j)))
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
		r[j] = NewAFWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	var rn int
	return NewAVWithRC(r, &rn)
}

func flipAI(x *AV) V {
	r := make([]int64, x.Len())
	for i, xi := range x.elts {
		if xi.IsI() {
			r[i] = xi.I()
			continue
		}
		switch xiv := xi.value.(type) {
		case *AB:
			r[i] = B2I(xiv.At(0))
		case *AI:
			r[i] = xiv.At(0)
		}
	}
	return NewAI(r)
}

func flipAVAI(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]int64, lines*x.Len())
	xlen := x.Len()
	var n int = 2
	rc := &n
	for i, xi := range x.elts {
		if xi.IsI() {
			n := xi.I()
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = n
			}
			continue
		}
		switch xiv := xi.value.(type) {
		case *AB:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = B2I(xiv.At(j))
			}
		case *AI:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xiv.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewAIWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	var rn int
	return NewAVWithRC(r, &rn)
}

func flipAS(x *AV) V {
	r := make([]string, x.Len())
	for i, xi := range x.elts {
		switch xiv := xi.value.(type) {
		case S:
			r[i] = string(xiv)
		case *AS:
			r[i] = xiv.At(0)
		}
	}
	return NewAS(r)
}

func flipAVAS(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]string, lines*x.Len())
	xlen := x.Len()
	var n int = 2
	rc := &n
	for i, xi := range x.elts {
		switch xiv := xi.value.(type) {
		case S:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = string(xiv)
			}
		case *AS:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xiv.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewASWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	var rn int
	return NewAVWithRC(r, &rn)
}

func flipAV(x *AV) V {
	r := make([]V, x.Len())
	for i, xi := range x.elts {
		switch xiv := xi.value.(type) {
		case array:
			r[i] = xiv.at(0)
		default:
			r[i] = xi
		}
	}
	return Canonical(NewAVWithRC(r, x.rc))
}

func flipAVAV(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]V, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.elts {
		switch xiv := xi.value.(type) {
		case array:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xiv.at(j)
			}
		default:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xi
			}
		}
	}
	*x.rc += 2
	for j := range r {
		r[j] = Canonical(NewAVWithRC(a[j*xlen:(j+1)*xlen], x.rc))
	}
	return NewAV(r)
}
