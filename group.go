package goal

// group returns =x.
func group(x V) V {
	if Length(x) == 0 {
		return NewAV([]V{})
	}
	switch x := x.Value.(type) {
	case *AB:
		n := sumAB(x)
		r := make([]V, int(B2I(n > 0)+1))
		ai := make([]int, x.Len())
		if n == 0 {
			for i := range ai {
				ai[i] = i
			}
			r[0] = NewAI(ai)
			return NewAV(r)
		}
		aif := ai[:len(ai)-n]
		ait := ai[len(ai)-n:]
		iTrue, iFalse := 0, 0
		for i, xi := range x.Slice {
			if xi {
				ait[iTrue] = i
				iTrue++
			} else {
				aif[iFalse] = i
				iFalse++
			}
		}
		r[0] = NewAI(aif)
		r[1] = NewAI(ait)
		return NewAV(r)
	case *AI:
		max := maxAI(x)
		if max < 0 {
			max = -1
		}
		r := make([]V, max+1)
		counta := make([]int, 2*(max+1))
		counts := counta[:max+1]
		countn := 0
		for _, j := range x.Slice {
			if j < 0 {
				countn++
				continue
			}
			counts[j]++
		}
		scounts := counta[max+1:]
		sn := 0
		for i, n := range counts {
			sn += n
			scounts[i] = sn
		}
		pj := 0
		ai := make([]int, x.Len()-countn)
		for i := range r {
			r[i] = NewAI(ai[pj:scounts[i]])
			pj = scounts[i]
		}
		for i, j := range x.Slice {
			if j < 0 {
				continue
			}
			ai[scounts[j]-counts[j]] = i
			counts[j]--
		}
		return NewAV(r)
	case *AF:
		z := toAI(x)
		if z.IsErr() {
			return z
		}
		return group(z)
	case *AV:
		//assertCanonical(x)
		return errs("=x : x non-integer array")
	default:
		return errs("=x : x not an integer array")
	}
}

// icount efficiently returns #'=x.
func icount(x V) V {
	if Length(x) == 0 {
		return NewAI([]int{})
	}
	switch x := x.Value.(type) {
	case *AB:
		n := sumAB(x)
		return NewAI([]int{x.Len() - n, n})
	case *AI:
		max := maxAI(x)
		if max < 0 {
			max = -1
		}
		counts := make([]int, max+1)
		for _, j := range x.Slice {
			if j >= 0 {
				counts[j]++
			}
		}
		return NewAI(counts)
	case *AF:
		z := toAI(x)
		if z.IsErr() {
			return z
		}
		return icount(z)
	case *AV:
		//assertCanonical(x)
		return errs("icount x : x non-integer array")
	default:
		return errf("icount x : x not an integer array (%s)", x.Type())
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if Length(x) != Length(y) {
		return errf("f=y : length mismatch for f[y] and y: %d vs %d ",
			Length(x), Length(y))
	}
	x = group(x)
	if x.IsErr() {
		return errs("f=y : f[y] not an integer array")
	}
	avx := x.Value.(*AV) // group should always return AV or errV
	switch y := y.Value.(type) {
	case array:
		r := make([]V, avx.Len())
		for i, xi := range avx.Slice {
			r[i] = y.atIndices(xi.Value.(*AI).Slice)
		}
		return NewAV(r)
	default:
		return errs("f=y : y not array")
	}
}
