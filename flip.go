package goal

// flip returns +x.
func flip(x V) V {
	x = toArray(x)
	switch xv := x.value.(type) {
	case *AV:
		cols := xv.Len()
		if cols == 0 {
			return NewAV([]V{x})
		}
		lines := -1
		for _, o := range xv.Slice {
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
		t := eType(xv)
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
	for i, xi := range x.Slice {
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
	rc := reuseRCp(x.rc)
	for i, xi := range x.Slice {
		if xi.IsI() {
			b := xi.I() == 1
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = b
			}
		} else {
			*rc++
			xia := xi.getAB()
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xia.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewABWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	return NewAVWithRC(r, rc)
}

func flipAF(x *AV) V {
	r := make([]float64, x.Len())
	for i, xi := range x.Slice {
		if xi.IsI() {
			r[i] = float64(xi.I())
			continue
		}
		if xi.IsF() {
			r[i] = float64(xi.F())
			continue
		}
		switch z := xi.value.(type) {
		case *AB:
			r[i] = float64(b2f(z.At(0)))
		case *AF:
			r[i] = z.At(0)
		case *AI:
			r[i] = float64(z.At(0))
		}
	}
	return NewAF(r)
}

func flipAVAF(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]float64, lines*x.Len())
	xlen := x.Len()
	rc := reuseRCp(x.rc)
	for i, xi := range x.Slice {
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
		switch z := xi.value.(type) {
		case *AB:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = float64(b2f(z.At(j)))
			}
		case *AF:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = z.At(j)
			}
		case *AI:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = float64(z.At(j))
			}
		}
	}
	for j := range r {
		r[j] = NewAFWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	return NewAVWithRC(r, rc)
}

func flipAI(x *AV) V {
	r := make([]int64, x.Len())
	for i, xi := range x.Slice {
		if xi.IsI() {
			r[i] = xi.I()
			continue
		}
		switch z := xi.value.(type) {
		case *AB:
			r[i] = b2i(z.At(0))
		case *AI:
			r[i] = z.At(0)
		}
	}
	return NewAI(r)
}

func flipAVAI(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]int64, lines*x.Len())
	xlen := x.Len()
	rc := reuseRCp(x.rc)
	for i, xi := range x.Slice {
		if xi.IsI() {
			n := xi.I()
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = n
			}
			continue
		}
		switch z := xi.value.(type) {
		case *AB:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = b2i(z.At(j))
			}
		case *AI:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = z.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewAIWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	return NewAVWithRC(r, rc)
}

func flipAS(x *AV) V {
	r := make([]string, x.Len())
	for i, xi := range x.Slice {
		switch z := xi.value.(type) {
		case S:
			r[i] = string(z)
		case *AS:
			r[i] = z.At(0)
		}
	}
	return NewAS(r)
}

func flipAVAS(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]string, lines*x.Len())
	xlen := x.Len()
	rc := reuseRCp(x.rc)
	for i, xi := range x.Slice {
		switch z := xi.value.(type) {
		case S:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = string(z)
			}
		case *AS:
			*rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = z.At(j)
			}
		}
	}
	for j := range r {
		r[j] = NewASWithRC(a[j*xlen:(j+1)*xlen], rc)
	}
	return NewAVWithRC(r, rc)
}

func flipAV(x *AV) V {
	r := make([]V, x.Len())
	for i, xi := range x.Slice {
		switch z := xi.value.(type) {
		case array:
			r[i] = z.at(0)
		default:
			r[i] = xi
		}
	}
	return Canonical(NewAV(r))
}

func flipAVAV(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]V, lines*x.Len())
	xlen := x.Len()
	for i, xi := range x.Slice {
		switch z := xi.value.(type) {
		case array:
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = z.at(j)
			}
		default:
			*x.rc++
			for j := 0; j < lines; j++ {
				a[i+j*xlen] = xi
			}
		}
	}
	for j := range r {
		r[j] = Canonical(NewAVWithRC(a[j*xlen:(j+1)*xlen], x.rc))
	}
	return NewAV(r)
}
