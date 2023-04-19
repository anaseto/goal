package goal

// icountLinesGroup returns =x.
func icountLinesGroup(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewAS(lineSplit(string(xv)))
	case *AB:
		if xv.Len() == 0 {
			return NewAI(nil)
		}
		n := sumAB(xv)
		if n == 0 {
			return NewAI([]int64{int64(xv.Len())})
		}
		return NewAI([]int64{int64(xv.Len()) - n, n})
	case *AI:
		if xv.Len() == 0 {
			return NewAI(nil)
		}
		max := maxAI(xv)
		if max < 0 {
			max = -1
		}
		icounts := make([]int64, max+1)
		for _, j := range xv.elts {
			if j >= 0 {
				icounts[j]++
			}
		}
		return NewAI(icounts)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return ppanic("=x : ", z)
		}
		return icountLinesGroup(z)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewAS(lineSplit(xi))
		}
		return NewAV(r)
	case *Dict:
		return groupBy(NewV(xv.values), NewV(xv.keys))
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := icountLinesGroup(xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	default:
		return panicType("=x", "x", x)
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	xlen := x.Len()
	if xlen != y.Len() {
		return Panicf("f=Y : length mismatch for f[Y] and Y: %d vs %d ",
			x.Len(), y.Len())

	}
	if xlen == 0 {
		return NewAV(nil)
	}
	switch xv := x.value.(type) {
	case *AB:
		n := int(sumAB(xv))
		r := make([]V, int(B2I(n > 0)+1))
		switch yv := y.value.(type) {
		case *AB:
			if n == 0 {
				r[0] = y
				return NewAVWithRC(r, yv.rc)
			}
			rf, rt := groupByBools[bool](xv.elts, yv.elts, n)
			var nrc int = 2
			r[0] = NewABWithRC(rf, &nrc)
			r[1] = NewABWithRC(rt, &nrc)
			return NewAVWithRC(r, &nrc)
		case *AI:
			if n == 0 {
				r[0] = y
				return NewAVWithRC(r, yv.rc)
			}
			rf, rt := groupByBools[int64](xv.elts, yv.elts, n)
			var nrc int = 2
			r[0] = NewAIWithRC(rf, &nrc)
			r[1] = NewAIWithRC(rt, &nrc)
			return NewAVWithRC(r, &nrc)
		case *AF:
			if n == 0 {
				r[0] = y
				return NewAVWithRC(r, yv.rc)
			}
			rf, rt := groupByBools[float64](xv.elts, yv.elts, n)
			var nrc int = 2
			r[0] = NewAFWithRC(rf, &nrc)
			r[1] = NewAFWithRC(rt, &nrc)
			return NewAVWithRC(r, &nrc)
		case *AS:
			if n == 0 {
				r[0] = y
				return NewAVWithRC(r, yv.rc)
			}
			rf, rt := groupByBools[string](xv.elts, yv.elts, n)
			var nrc int = 2
			r[0] = NewASWithRC(rf, &nrc)
			r[1] = NewASWithRC(rt, &nrc)
			return NewAVWithRC(r, &nrc)
		case *AV:
			if n == 0 {
				r[0] = y
				return NewAVWithRC(r, yv.rc)
			}
			rf, rt := groupByBools[V](xv.elts, yv.elts, n)
			*yv.rc += 2
			r[0] = Canonical(NewAVWithRC(rf, yv.rc))
			r[1] = Canonical(NewAVWithRC(rt, yv.rc))
			return NewAV(r)
		default:
			return panicType("f=Y", "Y", y)
		}
	case *AI:
		max := maxAI(xv)
		if max < 0 {
			max = -1
		}
		r := make([]V, max+1)
		// NOTE: we could do a stack-allocating variant for small max.
		// Also, if groups are big, we could maybe just allocate them
		// separately.
		counta := make([]int64, 2*(max+1))
		icounts := counta[:max+1]
		countn := 0
		for _, j := range xv.elts {
			if j < 0 {
				countn++
				continue
			}
			icounts[j]++
		}
		scounts := counta[max+1:]
		sn := int64(0)
		for i, n := range icounts {
			sn += n
			scounts[i] = sn
		}
		switch yv := y.value.(type) {
		case *AB:
			rg := groupByInt64s[bool](xv.elts, icounts, scounts, yv.elts, countn)
			var n int = 2
			pj := int64(0)
			for i := range r {
				r[i] = NewABWithRC(rg[pj:scounts[i]], &n)
				pj = scounts[i]
			}
			return NewAVWithRC(r, &n)
		case *AI:
			rg := groupByInt64s[int64](xv.elts, icounts, scounts, yv.elts, countn)
			var n int = 2
			pj := int64(0)
			for i := range r {
				r[i] = NewAIWithRC(rg[pj:scounts[i]], &n)
				pj = scounts[i]
			}
			return NewAVWithRC(r, &n)
		case *AF:
			rg := groupByInt64s[float64](xv.elts, icounts, scounts, yv.elts, countn)
			var n int = 2
			pj := int64(0)
			for i := range r {
				r[i] = NewAFWithRC(rg[pj:scounts[i]], &n)
				pj = scounts[i]
			}
			return NewAVWithRC(r, &n)
		case *AS:
			rg := groupByInt64s[string](xv.elts, icounts, scounts, yv.elts, countn)
			var n int = 2
			pj := int64(0)
			for i := range r {
				r[i] = NewASWithRC(rg[pj:scounts[i]], &n)
				pj = scounts[i]
			}
			return NewAVWithRC(r, &n)
		case *AV:
			rg := groupByInt64s[V](xv.elts, icounts, scounts, yv.elts, countn)
			pj := int64(0)
			*yv.rc += 2
			for i := range r {
				r[i] = Canonical(NewAVWithRC(rg[pj:scounts[i]], yv.rc))
				pj = scounts[i]
			}
			return NewAV(r)
		default:
			return panicType("f=Y", "Y", x)
		}
	case *AF:
		ix := toAI(xv)
		if ix.IsPanic() {
			return ppanic("f=x : f[Y]", ix)
		}
		return groupBy(ix, y)
	default:
		return panicType("f=Y", "f[Y]", x)
	}
}

func groupByBools[T any](x []bool, y []T, n int) (rf, rt []T) {
	r := make([]T, len(x))
	rf = r[:len(r)-n]
	rt = r[len(r)-n:]
	iTrue, iFalse := 0, 0
	for i, xi := range x {
		if xi {
			rt[iTrue] = y[i]
			iTrue++
		} else {
			rf[iFalse] = y[i]
			iFalse++
		}
	}
	return rf, rt
}

func groupByInt64s[T any](x, icounts, scounts []int64, y []T, n int) []T {
	r := make([]T, len(x)-n)
	for i, j := range x {
		if j < 0 {
			continue
		}
		r[scounts[j]-icounts[j]] = y[i]
		icounts[j]--
	}
	return r
}
