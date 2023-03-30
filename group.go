package goal

import "strings"

// group returns =x.
func group(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewAS(strings.Fields(string(xv)))
	case *AB:
		if xv.Len() == 0 {
			return NewAV(nil)
		}
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
		for i, xi := range xv.elts {
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
		if xv.Len() == 0 {
			return NewAV(nil)
		}
		max := maxAI(xv)
		if max < 0 {
			max = -1
		}
		r := make([]V, max+1)
		counta := make([]int64, 2*(max+1))
		counts := counta[:max+1]
		countn := 0
		for _, j := range xv.elts {
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
		for i, j := range xv.elts {
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
			return ppanic("=x : ", z)
		}
		return group(z)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewAS(strings.Fields(xi))
		}
		return NewAV(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := group(xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *Dict:
		gi := group(NewV(xv.values))
		if gi.IsPanic() {
			return gi
		}
		gv := gi.value.(*AV)
		r := gv.reuse()
		for i, gi := range gv.elts {
			r.elts[i] = NewV(xv.keys.atIndices(gi.value.(*AI)))
		}
		if r.rc == nil {
			var n = 2
			r.rc = &n
		} else {
			*r.rc += 2
		}
		r.InitWithRC(r.rc)
		return NewV(r)
	default:
		if x.Len() == 0 {
			return NewAV(nil)
		}
		return panicType("=x", "x", x)
	}
}

// icount efficiently returns #'=x.
func icount(x V) V {
	switch xv := x.value.(type) {
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
		counts := make([]int64, max+1)
		for _, j := range xv.elts {
			if j >= 0 {
				counts[j]++
			}
		}
		return NewAI(counts)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return ppanic("=x : ", z)
		}
		return icount(z)
	case *Dict:
		return icount(NewV(xv.values))
	default:
		if x.Len() == 0 {
			return NewAI(nil)
		}
		return panicType("=x", "x", x)
	}
}

// groupBy by returns {x}=y.
func groupBy(x, y V) V {
	if x.Len() != y.Len() {
		return Panicf("f=Y : length mismatch for f[Y] and Y: %d vs %d ",
			x.Len(), y.Len())

	}
	gx := group(x)
	if gx.IsPanic() {
		return panicType("f=Y", "f[Y]", x)
	}
	xav := gx.value.(*AV) // group should always return *AV or panicV
	switch yv := y.value.(type) {
	case array:
		r := make([]V, xav.Len())
		for i, xi := range xav.elts {
			r[i] = NewV(yv.atIndices(xi.value.(*AI)))
		}
		return NewAV(r)
	default:
		return panicType("f=Y", "Y", y)
	}
}
