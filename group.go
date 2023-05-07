package goal

// icountLinesGroup returns =x.
func icountLinesGroup(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewAS(lineSplit(string(xv)))
	case *AB:
		if xv.Len() == 0 {
			return newABb(nil)
		}
		if xv.IsBoolean() {
			n := sumIntegers(xv.elts)
			if n == 0 {
				if xv.Len() < 256 {
					return NewABWithRC([]byte{byte(xv.Len())}, reuseRCp(xv.rc))
				}
				return NewAIWithRC([]int64{int64(xv.Len())}, reuseRCp(xv.rc))
			}
			if xv.Len() < 256 {
				return NewABWithRC([]byte{byte(int64(xv.Len()) - n), byte(n)}, reuseRCp(xv.rc))
			}
			return NewAIWithRC([]int64{int64(xv.Len()) - n, n}, reuseRCp(xv.rc))
		}
		if xv.Len() < 256 {
			return NewABWithRC(icountBytes[byte](xv.elts), reuseRCp(xv.rc))
		}
		return NewAIWithRC(icountBytes[int64](xv.elts), reuseRCp(xv.rc))
	case *AI:
		if xv.Len() == 0 {
			return newABb(nil)
		}
		if xv.Len() < 256 {
			return NewABWithRC(icountInts[byte](xv.elts), reuseRCp(xv.rc))
		}
		return NewAIWithRC(icountInts[int64](xv.elts), reuseRCp(xv.rc))
	case *AF:
		x = toAI(xv)
		if x.IsPanic() {
			return ppanic("=x : ", x)
		}
		return icountLinesGroup(x)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewAS(lineSplit(xi))
		}
		return NewAV(r)
	case *Dict:
		return groupBy(NewV(xv.values), NewV(xv.keys))
	case *AV:
		return monadAV(xv, icountLinesGroup)
	default:
		return panicType("=x", "x", x)
	}
}

func icountInts[I integer](x []int64) []I {
	max := maxIntegers(x)
	if max < 0 {
		max = -1
	}
	icounts := make([]I, max+1)
	for _, xi := range x {
		if xi >= 0 {
			icounts[xi]++
		}
	}
	return icounts
}

func icountBytes[I integer](x []byte) []I {
	max := maxBytes(x)
	icounts := make([]I, max+1)
	for _, xi := range x {
		icounts[xi]++
	}
	return icounts
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
		if xv.IsBoolean() {
			return groupByBoolsV(xv.elts, y)
		}
		max := maxBytes(xv.elts)
		switch yv := y.value.(type) {
		case *AB:
			return groupByBytesBytes(xv.elts, yv.elts, max, yv.IsBoolean())
		case *AI:
			return groupByBytesInt64s(xv.elts, yv.elts, max)
		case *AF:
			return groupByBytesFloat64s(xv.elts, yv.elts, max)
		case *AS:
			return groupByBytesStrings(xv.elts, yv.elts, max)
		case *AV:
			return groupByBytesVs(xv.elts, yv.elts, max, yv.rc)
		default:
			return panicType("f=Y", "Y", x)
		}
	case *AI:
		max := maxIntegers(xv.elts)
		if max < 0 {
			return NewAV(nil)
		}
		switch yv := y.value.(type) {
		case *AB:
			return groupByInt64sBytes(xv.elts, yv.elts, max, yv.IsBoolean())
		case *AI:
			return groupByInt64sInt64s(xv.elts, yv.elts, max)
		case *AF:
			return groupByInt64sFloat64s(xv.elts, yv.elts, max)
		case *AS:
			return groupByInt64sStrings(xv.elts, yv.elts, max)
		case *AV:
			return groupByInt64sVs(xv.elts, yv.elts, max, yv.rc)
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

func groupByBoolsV(x []byte, y V) V {
	n := int(sumIntegers(x))
	r := make([]V, int(b2I(n > 0)+1))
	switch yv := y.value.(type) {
	case *AB:
		if n == 0 {
			r[0] = y
			return NewAVWithRC(r, yv.rc)
		}
		rf, rt := groupByBools[byte](x, yv.elts, n)
		var nrc int = 2
		r[0] = NewABWithRC(rf, &nrc)
		r[1] = NewABWithRC(rt, &nrc)
		return NewAVWithRC(r, &nrc)
	case *AI:
		if n == 0 {
			r[0] = y
			return NewAVWithRC(r, yv.rc)
		}
		rf, rt := groupByBools[int64](x, yv.elts, n)
		var nrc int = 2
		r[0] = NewAIWithRC(rf, &nrc)
		r[1] = NewAIWithRC(rt, &nrc)
		return NewAVWithRC(r, &nrc)
	case *AF:
		if n == 0 {
			r[0] = y
			return NewAVWithRC(r, yv.rc)
		}
		rf, rt := groupByBools[float64](x, yv.elts, n)
		var nrc int = 2
		r[0] = NewAFWithRC(rf, &nrc)
		r[1] = NewAFWithRC(rt, &nrc)
		return NewAVWithRC(r, &nrc)
	case *AS:
		if n == 0 {
			r[0] = y
			return NewAVWithRC(r, yv.rc)
		}
		rf, rt := groupByBools[string](x, yv.elts, n)
		var nrc int = 2
		r[0] = NewASWithRC(rf, &nrc)
		r[1] = NewASWithRC(rt, &nrc)
		return NewAVWithRC(r, &nrc)
	case *AV:
		if n == 0 {
			r[0] = y
			return NewAVWithRC(r, yv.rc)
		}
		rf, rt := groupByBools[V](x, yv.elts, n)
		*yv.rc += 2
		r[0] = Canonical(NewAVWithRC(rf, yv.rc))
		r[1] = Canonical(NewAVWithRC(rt, yv.rc))
		return NewAV(r)
	default:
		return panicType("f=Y", "Y", y)
	}
}

func groupByBools[T any](x []byte, y []T, n int) (rf, rt []T) {
	r := make([]T, len(x))
	rf = r[:len(r)-n]
	rt = r[len(r)-n:]
	var offset [2]int
	offset[1] = len(r) - n
	for i, xi := range x {
		j := offset[xi]
		offset[xi]++
		r[j] = y[i]
	}
	return rf, rt
}

func groupByBytesBytes(x []byte, y []byte, max byte, b bool) V {
	r, offset, yg := groupByPrepareBytes[byte](x, max)
	var rc int = 2
	var fl flags
	if b {
		fl = flagBool
	}
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AB{elts: yg[count : count+n], rc: &rc, flags: fl})
		count += n
	}
	groupByScatterBytes[byte](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByBytesInt64s(x []byte, y []int64, max byte) V {
	r, offset, yg := groupByPrepareBytes[int64](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAIWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatterBytes[int64](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByBytesFloat64s(x []byte, y []float64, max byte) V {
	r, offset, yg := groupByPrepareBytes[float64](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAFWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatterBytes[float64](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByBytesStrings(x []byte, y []string, max byte) V {
	r, offset, yg := groupByPrepareBytes[string](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewASWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatterBytes[string](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByBytesVs(x []byte, y []V, max byte, rc *int) V {
	r, offset, yg := groupByPrepareBytes[V](x, max)
	*rc += 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAVWithRC(yg[count:count+n], rc)
		count += n
	}
	groupByScatterBytes[V](x, y, yg, offset)
	for i, ri := range r {
		r[i] = Canonical(ri)
	}
	return NewAV(r)
}

func groupByPrepareBytes[T any](x []byte, max byte) ([]V, []int, []T) {
	r := make([]V, max+1)
	offset := make([]int, max+1)
	count := 0
	for _, xi := range x {
		count++
		offset[xi]++
	}
	yg := make([]T, count)
	return r, offset, yg
}

func groupByScatterBytes[T any](x []byte, y []T, yg []T, offset []int) {
	for i, xi := range x {
		j := offset[xi]
		offset[xi]++
		yg[j] = y[i]
	}
}

func groupByInt64sBytes(x []int64, y []byte, max int64, b bool) V {
	r, offset, yg := groupByPrepare[byte](x, max)
	var rc int = 2
	var fl flags
	if b {
		fl = flagBool
	}
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AB{elts: yg[count : count+n], rc: &rc, flags: fl})
		count += n
	}
	groupByScatter[byte](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByInt64sInt64s(x []int64, y []int64, max int64) V {
	r, offset, yg := groupByPrepare[int64](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAIWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatter[int64](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByInt64sFloat64s(x []int64, y []float64, max int64) V {
	r, offset, yg := groupByPrepare[float64](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAFWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatter[float64](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByInt64sStrings(x []int64, y []string, max int64) V {
	r, offset, yg := groupByPrepare[string](x, max)
	var rc int = 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewASWithRC(yg[count:count+n], &rc)
		count += n
	}
	groupByScatter[string](x, y, yg, offset)
	return NewAVWithRC(r, &rc)
}

func groupByInt64sVs(x []int64, y []V, max int64, rc *int) V {
	r, offset, yg := groupByPrepare[V](x, max)
	*rc += 2
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewAVWithRC(yg[count:count+n], rc)
		count += n
	}
	groupByScatter[V](x, y, yg, offset)
	for i, ri := range r {
		r[i] = Canonical(ri)
	}
	return NewAV(r)
}

func groupByPrepare[T any](x []int64, max int64) ([]V, []int, []T) {
	r := make([]V, max+1)
	offset := make([]int, max+1)
	count := 0
	for _, xi := range x {
		if xi < 0 {
			continue
		}
		count++
		offset[xi]++
	}
	yg := make([]T, count)
	return r, offset, yg
}

func groupByScatter[T any](x []int64, y []T, yg []T, offset []int) {
	for i, xi := range x {
		if xi < 0 {
			continue
		}
		j := offset[xi]
		offset[xi]++
		yg[j] = y[i]
	}
}
