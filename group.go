package goal

// group returns =x.
func group(x V) V {
	if Length(x) == 0 {
		return NewAV([]V{})
	}
	switch xv := x.Value.(type) {
	case *AB:
		n := int(sumAB(xv))
		r := make([]V, int(B2I(n > 0)+1))
		ai := make([]int64, xv.Len())
		if n == 0 {
			for i := range ai {
				ai[i] = int64(i)
			}
			r[0] = NewAI(ai)
			return NewAV(r)
		}
		aif := ai[:len(ai)-n]
		ait := ai[len(ai)-n:]
		iTrue, iFalse := 0, 0
		for i, xi := range xv.Slice {
			if xi {
				ait[iTrue] = int64(i)
				iTrue++
			} else {
				aif[iFalse] = int64(i)
				iFalse++
			}
		}
		r[0] = NewAI(aif)
		r[1] = NewAI(ait)
		return NewAV(r)
	case *AI:
		max := maxAI(xv)
		if max < 0 {
			max = -1
		}
		r := make([]V, max+1)
		counta := make([]int64, 2*(max+1))
		counts := counta[:max+1]
		countn := 0
		for _, j := range xv.Slice {
			if j < 0 {
				countn++
				continue
			}
			counts[j]++
		}
		scounts := counta[max+1:]
		sn := int64(0)
		for i, n := range counts {
			sn += n
			scounts[i] = sn
		}
		pj := int64(0)
		ai := make([]int64, xv.Len()-countn)
		for i := range r {
			r[i] = NewAI(ai[pj:scounts[i]])
			pj = scounts[i]
		}
		for i, j := range xv.Slice {
			if j < 0 {
				continue
			}
			ai[scounts[j]-counts[j]] = int64(i)
			counts[j]--
		}
		return NewAV(r)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return group(z)
	default:
		return panicf("=x : x not an integer array (%s)", x.Type())
	}
}

// icount efficiently returns #'=x.
func icount(x V) V {
	if Length(x) == 0 {
		return NewAI([]int64{})
	}
	switch xv := x.Value.(type) {
	case *AB:
		n := sumAB(xv)
		return NewAI([]int64{int64(xv.Len()) - n, n})
	case *AI:
		max := maxAI(xv)
		if max < 0 {
			max = -1
		}
		counts := make([]int64, max+1)
		for _, j := range xv.Slice {
			if j >= 0 {
				counts[j]++
			}
		}
		return NewAI(counts)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return icount(z)
	default:
		return panicf("icount x : x not an integer array (%s)", x.Type())
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if Length(x) != Length(y) {
		return panicf("f=y : length mismatch for f[y] and y: %d vs %d ",
			Length(x), Length(y))
	}
	x = group(x)
	if x.IsPanic() {
		return panics("f=y : f[y] not an integer array")
	}
	avx := x.Value.(*AV) // group should always return AV or errV
	switch yv := y.Value.(type) {
	case array:
		r := make([]V, avx.Len())
		for i, xi := range avx.Slice {
			r[i] = yv.atIndices(xi.Value.(*AI).Slice)
		}
		return NewAV(r)
	default:
		return panicf("f=y : y not array (%s)", y.Type())
	}
}
