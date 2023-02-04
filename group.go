package goal

// group returns =x.
func group(x V) V {
	if Length(x) == 0 {
		return NewAV(nil)
	}
	switch xv := x.value.(type) {
	case *AB:
		n := int(sumAB(xv))
		r := make([]V, int(b2i(n > 0)+1))
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
		var nrc int = 2
		r[0] = NewAIWithRC(aif, &nrc)
		r[1] = NewAIWithRC(ait, &nrc)
		return NewAVWithRC(r, &nrc)
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
		var n int = 2
		for i := range r {
			r[i] = NewAIWithRC(ai[pj:scounts[i]], &n)
			pj = scounts[i]
		}
		for i, j := range xv.Slice {
			if j < 0 {
				continue
			}
			ai[scounts[j]-counts[j]] = int64(i)
			counts[j]--
		}
		return NewAVWithRC(r, &n)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return group(z)
	default:
		return Panicf("=x : x not an integer array (%s)", x.Type())
	}
}

// icount efficiently returns #'=x.
func icount(x V) V {
	if Length(x) == 0 {
		return NewAI(nil)
	}
	switch xv := x.value.(type) {
	case *AB:
		n := sumAB(xv)
		if n == 0 {
			return NewAI([]int64{int64(xv.Len())})
		}
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
		return Panicf("icount x : x not an integer array (%s)", x.Type())
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if Length(x) != Length(y) {
		return Panicf("f=y : length mismatch for f[y] and y: %d vs %d ",
			Length(x), Length(y))

	}
	x = group(x)
	if x.IsPanic() {
		return panics("f=y : f[y] not an integer array")
	}
	xav := x.value.(*AV) // group should always return *AV or panicV
	switch yv := y.value.(type) {
	case array:
		r := make([]V, xav.Len())
		for i, xi := range xav.Slice {
			r[i] = yv.atIndices(xi.value.(*AI))
		}
		return NewAV(r)
	default:
		return Panicf("f=y : y not array (%s)", y.Type())
	}
}
