package goal

// array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type array interface {
	RefCountHolder
	Len() int
	at(i int) V           // x[i]
	slice(i, j int) array // x[i:j]
	getFlags() flags
	setFlags(flags)
	set(i int, y V)
	atIndices(y []int64) V // x[y] (goal code)
	shallowClone() array
}

// Len returns the length of the array.
func (x *AB) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AI) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AF) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AS) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AV) Len() int { return len(x.Slice) }

func (x *AB) at(i int) V { return NewI(b2i(x.Slice[i])) }
func (x *AI) at(i int) V { return NewI(x.Slice[i]) }
func (x *AF) at(i int) V { return NewF(x.Slice[i]) }
func (x *AS) at(i int) V { return NewS(x.Slice[i]) }
func (x *AV) at(i int) V { return x.Slice[i] }

// At returns array value at the given index.
func (x *AB) At(i int) bool { return x.Slice[i] }

// At returns array value at the given index.
func (x *AI) At(i int) int64 { return x.Slice[i] }

// At returns array value at the given index.
func (x *AF) At(i int) float64 { return x.Slice[i] }

// At returns array value at the given index.
func (x *AS) At(i int) string { return x.Slice[i] }

// At returns array value at the given index.
func (x *AV) At(i int) V { return x.Slice[i] }

func (x *AB) slice(i, j int) array { return &AB{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AI) slice(i, j int) array { return &AI{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AF) slice(i, j int) array { return &AF{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AS) slice(i, j int) array { return &AS{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AV) slice(i, j int) array { return &AV{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }

func (x *AB) getFlags() flags { return x.flags }
func (x *AI) getFlags() flags { return x.flags }
func (x *AF) getFlags() flags { return x.flags }
func (x *AS) getFlags() flags { return x.flags }
func (x *AV) getFlags() flags { return x.flags }

func (x *AB) setFlags(f flags) { x.flags = f }
func (x *AI) setFlags(f flags) { x.flags = f }
func (x *AF) setFlags(f flags) { x.flags = f }
func (x *AS) setFlags(f flags) { x.flags = f }
func (x *AV) setFlags(f flags) { x.flags = f }

// set changes x at i with y (in place).
func (x *AB) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n != 0
	} else {
		x.Slice[i] = y.F() != 0
	}
}

// set changes x at i with y (in place).
func (x *AI) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n
	} else {
		x.Slice[i] = int64(y.F())
	}
}

// set changes x at i with y (in place).
func (x *AF) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = float64(y.I())
	} else {
		x.Slice[i] = y.F()
	}
}

// set changes x at i with y (in place).
func (x *AS) set(i int, y V) {
	x.Slice[i] = string(y.value.(S))
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	y.InitWithRC(x.rc)
	x.Slice[i] = y
}

func (x *AV) atIndices(y []int64) V {
	r := make([]V, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	nr := &AV{Slice: r}
	nr.InitWithRC(reuseRCp(x.rc))
	return NewV(canonicalAV(nr))
}

func (x *AB) atIndices(y []int64) V {
	r := make([]bool, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewV(&AB{Slice: r, rc: reuseRCp(x.rc)})
}

func (x *AI) atIndices(y []int64) V {
	r := make([]int64, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewV(&AI{Slice: r, rc: reuseRCp(x.rc)})
}

func (x *AF) atIndices(y []int64) V {
	r := make([]float64, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewV(&AF{Slice: r, rc: reuseRCp(x.rc)})
}

func (x *AS) atIndices(y []int64) V {
	r := make([]string, len(y))
	xlen := int64(x.Len())
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("x[y] : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewV(&AS{Slice: r, rc: reuseRCp(x.rc)})
}

func (x *AB) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AB{Slice: make([]bool, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AI) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AI{Slice: make([]int64, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AF) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AF{Slice: make([]float64, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AS) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AS{Slice: make([]string, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AV) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AV{Slice: make([]V, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}
