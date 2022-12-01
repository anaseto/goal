package goal

// flip returns +x.
func flip(x V) V {
	//assertCanonical(x)
	x = toArray(x)
	switch xv := x.Value.(type) {
	case AV:
		cols := xv.Len()
		if cols == 0 {
			return NewV(AV{x})
		}
		lines := -1
		for _, o := range xv {
			nl := Length(o)
			if _, ok := o.Value.(array); !ok {
				continue
			}
			switch {
			case lines < 0:
				lines = nl
			case nl >= 1 && nl != lines:
				return errf("line length mismatch: %d vs %d", nl, lines)
			}
		}
		t := aType(xv)
		switch {
		case lines <= 0:
			return NewV(AV{x})
		case lines == 1:
			switch t {
			case tB, tAB:
				return NewV(AV{NewV(flipAB(xv))})
			case tF, tAF:
				return NewV(AV{NewV(flipAF(xv))})
			case tI, tAI:
				return NewV(AV{NewV(flipAI(xv))})
			case tS, tAS:
				return NewV(AV{NewV(flipAS(xv))})
			default:
				return NewV(AV{flipAV(xv)})
			}
		default:
			switch t {
			case tB, tAB:
				return NewV(flipAVAB(xv, lines))
			case tF, tAF:
				return NewV(flipAVAF(xv, lines))
			case tI, tAI:
				return NewV(flipAVAI(xv, lines))
			case tS, tAS:
				return NewV(flipAVAS(xv, lines))
			default:
				return NewV(flipAVAV(xv, lines))
			}
		}
	default:
		return NewV(AV{x})
	}
}

func flipAB(x AV) AB {
	r := make(AB, x.Len())
	for i, xi := range x {
		if xi.IsInt() {
			r[i] = xi.Int() == 1
		} else {
			r[i] = xi.AB()[0]
		}
	}
	return r
}

func flipAVAB(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AB, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			if xi.IsInt() {
				q[i] = xi.Int() == 1
			} else {
				q[i] = xi.AB()[j]
			}
		}
		r[j] = NewV(q)
	}
	return r
}

func flipAF(x AV) AF {
	r := make(AF, x.Len())
	for i, xi := range x {
		if xi.IsInt() {
			r[i] = float64(xi.Int())
			continue
		}
		switch z := xi.Value.(type) {
		case AB:
			r[i] = float64(B2F(z[0]))
		case F:
			r[i] = float64(z)
		case AF:
			r[i] = z[0]
		case AI:
			r[i] = float64(z[0])
		}
	}
	return r
}

func flipAVAF(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AF, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			if xi.IsInt() {
				q[i] = float64(xi.Int())
				continue
			}
			switch z := xi.Value.(type) {
			case AB:
				q[i] = float64(B2F(z[j]))
			case F:
				q[i] = float64(z)
			case AF:
				q[i] = z[j]
			case AI:
				q[i] = float64(z[j])
			}
		}
		r[j] = NewV(q)
	}
	return r
}

func flipAI(x AV) AI {
	r := make(AI, x.Len())
	for i, xi := range x {
		if xi.IsInt() {
			r[i] = xi.Int()
			continue
		}
		switch z := xi.Value.(type) {
		case AB:
			r[i] = int(B2I(z[0]))
		case AI:
			r[i] = z[0]
		}
	}
	return r
}

func flipAVAI(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AI, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			if xi.IsInt() {
				q[i] = xi.Int()
				continue
			}
			switch z := xi.Value.(type) {
			case AB:
				q[i] = int(B2I(z[j]))
			case AI:
				q[i] = z[j]
			}
		}
		r[j] = NewV(q)
	}
	return r
}

func flipAS(x AV) AS {
	r := make(AS, x.Len())
	for i, xi := range x {
		switch z := xi.Value.(type) {
		case S:
			r[i] = string(z)
		case AS:
			r[i] = z[0]
		}
	}
	return r
}

func flipAVAS(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AS, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.Value.(type) {
			case S:
				q[i] = string(z)
			case AS:
				q[i] = z[j]
			}
		}
		r[j] = NewV(q)
	}
	return r
}

func flipAV(x AV) V {
	r := make(AV, x.Len())
	for i, xi := range x {
		switch z := xi.Value.(type) {
		case array:
			r[i] = z.at(0)
		default:
			r[i] = xi
		}
	}
	return NewV(canonical(r))
}

func flipAVAV(x AV, lines int) AV {
	r := make(AV, lines)
	a := make(AV, lines*x.Len())
	for j := range r {
		q := a[j*x.Len() : (j+1)*x.Len()]
		for i, xi := range x {
			switch z := xi.Value.(type) {
			case array:
				q[i] = z.at(j)
			default:
				q[i] = xi
			}
		}
		r[j] = NewV(q)
	}
	return r
}
