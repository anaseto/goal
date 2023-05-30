package goal

// icountGroup returns =x.
func icountGroup(x V) V {
	switch xv := x.bv.(type) {
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
					return NewAB([]byte{byte(xv.Len())})
				}
				return NewAI([]int64{int64(xv.Len())})
			}
			if xv.Len() < 256 {
				return NewAB([]byte{byte(int64(xv.Len()) - n), byte(n)})
			}
			return NewAI([]int64{int64(xv.Len()) - n, n})
		}
		if xv.Len() < 256 {
			return NewAB(icountBytes[byte](xv.elts))
		}
		return NewAI(icountBytes[int64](xv.elts))
	case *AI:
		if xv.Len() == 0 {
			return newABb(nil)
		}
		if xv.Len() < 256 {
			return NewAB(icountInts[byte](xv.elts))
		}
		return NewAI(icountInts[int64](xv.elts))
	case *AF:
		x = toAI(xv)
		if x.IsPanic() {
			return ppanic("=x : ", x)
		}
		return icountGroup(x)
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = NewV(&AS{elts: lineSplit(xi), flags: flagImmutable})
		}
		return newAVu(r)
	case *Dict:
		return groupBy(NewV(xv.values), NewV(xv.keys))
	case *AV:
		return monadAV(xv, icountGroup)
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
	max := int(maxBytes(x))
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
		return protoAV()
	}
	switch xv := x.bv.(type) {
	case *AB:
		if xv.IsBoolean() {
			return groupByBoolsV(xv.elts, y)
		}
		max := maxBytes(xv.elts)
		if ascending(xv) {
			switch yv := y.bv.(type) {
			case array:
				return groupBySorted(xv.elts, yv, int64(max))
			default:
				return panicType("f=Y", "Y", y)
			}
		}
		switch yv := y.bv.(type) {
		case *AB:
			return groupByBytesBytes(xv.elts, yv.elts, max, yv.IsBoolean())
		case *AI:
			return groupByBytesInt64s(xv.elts, yv.elts, max)
		case *AF:
			return groupByBytesFloat64s(xv.elts, yv.elts, max)
		case *AS:
			return groupByBytesStrings(xv.elts, yv.elts, max)
		case *AV:
			return groupByBytesVs(xv.elts, yv.elts, max)
		default:
			return panicType("f=Y", "Y", y)
		}
	case *AI:
		max := maxIntegers(xv.elts)
		if max < 0 {
			return protoAV()
		}
		if ascending(xv) {
			switch yv := y.bv.(type) {
			case array:
				return groupBySorted(xv.elts, yv, int64(max))
			default:
				return panicType("f=Y", "Y", y)
			}
		}
		switch yv := y.bv.(type) {
		case *AB:
			return groupByInt64sBytes(xv.elts, yv.elts, max, yv.IsBoolean())
		case *AI:
			return groupByInt64sInt64s(xv.elts, yv.elts, max)
		case *AF:
			return groupByInt64sFloat64s(xv.elts, yv.elts, max)
		case *AS:
			return groupByInt64sStrings(xv.elts, yv.elts, max)
		case *AV:
			return groupByInt64sVs(xv.elts, yv.elts, max)
		default:
			return panicType("f=Y", "Y", y)
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
	switch yv := y.bv.(type) {
	case *AB:
		if n == 0 {
			y.MarkImmutable()
			r[0] = y
			return newAVu(r)
		}
		rf, rt := groupByBools[byte](x, yv.elts, n)
		r[0] = NewV(&AB{elts: rf, flags: flagImmutable})
		r[1] = NewV(&AB{elts: rt, flags: flagImmutable})
		return newAVu(r)
	case *AI:
		if n == 0 {
			y.MarkImmutable()
			r[0] = y
			return newAVu(r)
		}
		rf, rt := groupByBools[int64](x, yv.elts, n)
		r[0] = NewV(&AI{elts: rf, flags: flagImmutable})
		r[1] = NewV(&AI{elts: rt, flags: flagImmutable})
		return newAVu(r)
	case *AF:
		if n == 0 {
			y.MarkImmutable()
			r[0] = y
			return newAVu(r)
		}
		rf, rt := groupByBools[float64](x, yv.elts, n)
		r[0] = NewV(&AF{elts: rf, flags: flagImmutable})
		r[1] = NewV(&AF{elts: rt, flags: flagImmutable})
		return newAVu(r)
	case *AS:
		if n == 0 {
			y.MarkImmutable()
			r[0] = y
			return newAVu(r)
		}
		rf, rt := groupByBools[string](x, yv.elts, n)
		r[0] = NewV(&AS{elts: rf, flags: flagImmutable})
		r[1] = NewV(&AS{elts: rt, flags: flagImmutable})
		return newAVu(r)
	case *AV:
		if n == 0 {
			y.MarkImmutable()
			r[0] = y
			return newAVu(r)
		}
		rf, rt := groupByBools[V](x, yv.elts, n)
		r[0] = canonicalVs(rf)
		r[0].MarkImmutable()
		r[1] = canonicalVs(rt)
		r[1].MarkImmutable()
		return newAVu(r)
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
	var fl flags
	if b {
		fl = flagBool
	}
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AB{elts: yg[count : count+n], flags: fl | flagImmutable})
		count += n
	}
	groupByScatterBytes[byte](x, y, yg, offset)
	return newAVu(r)
}

func groupByBytesInt64s(x []byte, y []int64, max byte) V {
	r, offset, yg := groupByPrepareBytes[int64](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AI{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatterBytes[int64](x, y, yg, offset)
	return newAVu(r)
}

func groupByBytesFloat64s(x []byte, y []float64, max byte) V {
	r, offset, yg := groupByPrepareBytes[float64](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AF{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatterBytes[float64](x, y, yg, offset)
	return newAVu(r)
}

func groupByBytesStrings(x []byte, y []string, max byte) V {
	r, offset, yg := groupByPrepareBytes[string](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AS{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatterBytes[string](x, y, yg, offset)
	return newAVu(r)
}

func groupByBytesVs(x []byte, y []V, max byte) V {
	r, offset, yg := groupByPrepareBytes[V](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AV{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatterBytes[V](x, y, yg, offset)
	for i, ri := range r {
		r[i] = canonicalImmut(ri)
	}
	return newAVu(r)
}

func groupByPrepareBytes[T any](x []byte, max byte) ([]V, []int, []T) {
	l := int(max) + 1
	r := make([]V, l)
	offset := make([]int, l)
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
	var fl flags
	if b {
		fl = flagBool
	}
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AB{elts: yg[count : count+n], flags: fl | flagImmutable})
		count += n
	}
	groupByScatter[byte](x, y, yg, offset)
	return newAVu(r)
}

func groupByInt64sInt64s(x []int64, y []int64, max int64) V {
	r, offset, yg := groupByPrepare[int64](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AI{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatter[int64](x, y, yg, offset)
	return newAVu(r)
}

func groupByInt64sFloat64s(x []int64, y []float64, max int64) V {
	r, offset, yg := groupByPrepare[float64](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AF{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatter[float64](x, y, yg, offset)
	return newAVu(r)
}

func groupByInt64sStrings(x []int64, y []string, max int64) V {
	r, offset, yg := groupByPrepare[string](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AS{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatter[string](x, y, yg, offset)
	return newAVu(r)
}

func groupByInt64sVs(x []int64, y []V, max int64) V {
	r, offset, yg := groupByPrepare[V](x, max)
	count := 0
	for i, n := range offset {
		offset[i] = count
		r[i] = NewV(&AV{elts: yg[count : count+n], flags: flagImmutable})
		count += n
	}
	groupByScatter[V](x, y, yg, offset)
	for i, ri := range r {
		r[i] = canonicalImmut(ri)
	}
	return newAVu(r)
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

func groupBySorted[I integer](x []I, y array, max int64) V {
	r := make([]V, max+1)
	var from, i0 int
	var n int64
	for i, xi := range x {
		if xi >= 0 {
			i0 = i
			break
		}
	}
	from = i0
	var p V
	y.MarkImmutable()
	for i, xi := range x[i0:] {
		if int64(xi) == n {
			continue
		}
		r[n] = NewV(y.slice(from, i+i0))
		from = i + i0
		n++
		for n < int64(xi) {
			if p.kind == valNil {
				p = NewV(&AV{flags: flagImmutable})
			}
			r[n] = p
			n++
		}
	}
	r[n] = NewV(y.slice(from, len(x)))
	return newAVu(r)
}
