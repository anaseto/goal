package goal

import "math/rand"

func uniform(ctx *Context, x V) V {
	var n int64
	if x.IsI() {
		n = x.I()
	} else {
		if !isI(x.F()) {
			return Panicf("?i : non-integer i (%g)", x.F())
		}
		n = int64(x.F())
	}
	if ctx.rand == nil {
		ctx.rand = rand.New(rand.NewSource(1))
	}
	if n < 0 {
		r := make([]float64, -n)
		for i := range r {
			r[i] = ctx.rand.NormFloat64()
		}
		return NewAF(r)
	}
	r := make([]float64, n)
	for i := range r {
		r[i] = ctx.rand.Float64()
	}
	return NewAF(r)
}

func rolldeal(ctx *Context, x, y V) V {
	var n int64
	if x.IsI() {
		n = x.I()
	} else {
		if !isI(x.F()) {
			return Panicf("i?y : non-integer y (%g)", x.F())
		}
		n = int64(x.F())
	}
	if ctx.rand == nil {
		ctx.rand = rand.New(rand.NewSource(1))
	}
	if n < 0 {
		return deal(ctx, -n, y)
	}
	return roll(ctx, n, y)
}

func rollSlice[T any](ctx *Context, n int64, y []T) []T {
	r := make([]T, n)
	ylen := len(y)
	if ylen == 0 {
		return nil
	}
	for i := range r {
		r[i] = y[ctx.rand.Intn(ylen)]
	}
	return r
}

func roll(ctx *Context, n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		if yv <= 0 {
			return Panicf("i?y : non-positive y (%d)", yv)
		}
		if yv == 2 {
			r := make([]byte, n)
			var i int64
		loop:
			for {
				u := ctx.rand.Uint64()
				for j := 0; j < 64; j += 8 {
					if i >= n {
						break loop
					}
					b := uint8(u>>j) >> 7
					r[i] = b != 0
					i++
				}
			}
			return NewAB(r)
		} else {
			r := make([]int64, n)
			for i := range r {
				r[i] = ctx.rand.Int63n(yv)
			}
			return NewAI(r)
		}
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("i?y : non-integer y (%g)", y.F())
		}
		return roll(ctx, n, NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
	case *AB:
		return NewAB(rollSlice[bool](ctx, n, yv.elts))
	case *AI:
		return NewAI(rollSlice[int64](ctx, n, yv.elts))
	case *AF:
		return NewAF(rollSlice[float64](ctx, n, yv.elts))
	case *AS:
		return NewAS(rollSlice[string](ctx, n, yv.elts))
	case *AV:
		*yv.rc += 2
		return NewAVWithRC(rollSlice[V](ctx, n, yv.elts), yv.rc)
	default:
		return panicType("i?y", "y", y)
	}
}

func dealSlice[T any](ctx *Context, n int64, y []T) []T {
	ylen := len(y)
	if n == 0 {
		return nil
	}
	if n*4+8 < int64(ylen) {
		r := make([]T, n)
		// For small n, we use hashing, not the fastest
		// algorithm, but not too bad either.
		m := map[int]struct{}{}
		var i int64
		for i < n {
			k := ctx.rand.Intn(ylen)
			_, ok := m[k]
			if ok {
				continue
			}
			m[k] = struct{}{}
			r[i] = y[k]
			i++
		}
		return r
	}
	r := make([]T, ylen)
	copy(r, y)
	ctx.rand.Shuffle(ylen, func(i, j int) {
		r[i], r[j] = r[j], r[i]
	})
	r = r[:n]
	return r
}

func deal(ctx *Context, n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		if yv <= 0 {
			return Panicf("(-i)?y : non-positive y (%d)", yv)
		}
		if n > yv {
			return Panicf("(-i)?y : i > y (%d vs %d)", n, yv)
		}
		if n*4+8 < yv {
			// For small n, we use hashing, not the fastest
			// algorithm, but not too bad either.
			r := make([]int64, n)
			m := map[int64]struct{}{}
			var i int64
			for i < n {
				k := ctx.rand.Int63n(yv)
				_, ok := m[k]
				if ok {
					continue
				}
				m[k] = struct{}{}
				r[i] = k
				i++
			}
			return NewV(&AI{elts: r, flags: flagUnique})
		}
		r := make([]int64, yv)
		for i := range r {
			r[i] = int64(i)
		}
		ctx.rand.Shuffle(int(yv), func(i, j int) {
			r[i], r[j] = r[j], r[i]
		})
		r = r[:n]
		return NewV(&AI{elts: r, flags: flagUnique})
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("(-i)?y : non-integer y (%g)", y.F())
		}
		return deal(ctx, n, NewI(int64(y.F())))
	}
	ya, ok := y.value.(array)
	if !ok {
		return panicType("(-i)?y", "y", y)
	}
	ylen := ya.Len()
	if ylen == 0 {
		return panics("(-i)?Y : empty Y")
	}
	if n > int64(ylen) {
		return Panicf("(-i)?Y : i > #Y (%d vs %d)", n, ylen)
	}
	switch yv := y.value.(type) {
	case *AB:
		return NewV(&AB{elts: dealSlice[bool](ctx, n, yv.elts), flags: flagUnique})
	case *AI:
		return NewV(&AI{elts: dealSlice[int64](ctx, n, yv.elts), flags: flagUnique})
	case *AF:
		return NewV(&AF{elts: dealSlice[float64](ctx, n, yv.elts), flags: flagUnique})
	case *AS:
		return NewV(&AS{elts: dealSlice[string](ctx, n, yv.elts), flags: flagUnique})
	case *AV:
		*yv.rc += 2
		return NewV(&AV{elts: dealSlice[V](ctx, n, yv.elts), flags: flagUnique, rc: yv.rc})
	default:
		panic("deal")
	}
}

func seed(ctx *Context, x V) V {
	if x.IsI() {
		if ctx.rand == nil {
			ctx.rand = rand.New(rand.NewSource(x.I()))
			return x
		}
		ctx.rand.Seed(x.I())
		return x
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf(`goal.seed i : non-integer i (%g)`, x.F())
		}
		if ctx.rand == nil {
			ctx.rand = rand.New(rand.NewSource(int64(x.F())))
			return NewI(int64(x.F()))
		}
		ctx.rand.Seed(int64(x.F()))
		return NewI(int64(x.F()))
	}
	return panicType(`goal.seed i`, "i", x)
}
