package goal

// flip returns +x.
func flip(x V) V {
	//assertCanonical(x)
	x = toArray(x)
	switch xv := x.value.(type) {
	case *AV:
		cols := xv.Len()
		if cols == 0 {
			return NewAV([]V{x})
		}
		lines := -1
		for _, o := range xv.Slice {
			nl := int(Length(o))
			if _, ok := o.value.(array); !ok {
				continue
			}
			switch {
			case lines < 0:
				lines = nl
			case nl >= 1 && nl != lines:
				return panicf("line length mismatch: %d vs %d", nl, lines)
			}
		}
		t := aType(xv)
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
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x.Slice {
			if xi.IsI() {
				q[i] = xi.I() == 1
			} else {
				q[i] = xi.getAB().At(j)
			}
		}
		r[j] = NewAB(q)
	}
	return NewAV(r)
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
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x.Slice {
			if xi.IsI() {
				q[i] = float64(xi.I())
				continue
			}
			if xi.IsF() {
				q[i] = float64(xi.F())
				continue
			}
			switch z := xi.value.(type) {
			case *AB:
				q[i] = float64(b2f(z.At(j)))
			case *AF:
				q[i] = z.At(j)
			case *AI:
				q[i] = float64(z.At(j))
			}
		}
		r[j] = NewAF(q)
	}
	return NewAV(r)
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
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x.Slice {
			if xi.IsI() {
				q[i] = xi.I()
				continue
			}
			switch z := xi.value.(type) {
			case *AB:
				q[i] = b2i(z.At(j))
			case *AI:
				q[i] = z.At(j)
			}
		}
		r[j] = NewAI(q)
	}
	return NewAV(r)
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
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x.Slice {
			switch z := xi.value.(type) {
			case S:
				q[i] = string(z)
			case *AS:
				q[i] = z.At(j)
			}
		}
		r[j] = NewAS(q)
	}
	return NewAV(r)
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
	return canonicalV(NewAV(r))
}

func flipAVAV(x *AV, lines int) V {
	r := make([]V, lines)
	a := make([]V, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x.Slice {
			switch z := xi.value.(type) {
			case array:
				q[i] = z.at(j)
			default:
				q[i] = xi
			}
		}
		r[j] = NewAV(q)
	}
	return NewAV(r)
}
