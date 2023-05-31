package goal

import (
	"math"
	"math/rand"
)

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

func roll(ctx *Context, n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		return rollI(ctx, n, yv)
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("i?y : non-integer y (%g)", y.F())
		}
		return roll(ctx, n, NewI(int64(y.F())))
	}
	switch yv := y.bv.(type) {
	case *AB:
		var fl flags
		if yv.IsBoolean() {
			fl |= flagBool
		}
		return NewV(&AB{elts: rollSlice[byte](ctx, n, yv.elts), flags: fl})
	case *AI:
		return NewAI(rollSlice[int64](ctx, n, yv.elts))
	case *AF:
		return NewAF(rollSlice[float64](ctx, n, yv.elts))
	case *AS:
		return NewAS(rollSlice[string](ctx, n, yv.elts))
	case *AV:
		return canonicalVs(rollSlice[V](ctx, n, yv.elts))
	default:
		return panicType("i?y", "y", y)
	}
}

func rollI(ctx *Context, n, y int64) V {
	if y <= 0 {
		return Panicf("i?y : non-positive y (%d)", y)
	}
	if y == 2 {
		r := make([]byte, n)
		var i int64
	loop:
		for {
			u := ctx.rand.Uint64()
			for j := 0; j < 64; j += 8 {
				if i >= n {
					break loop
				}
				r[i] = uint8(u>>j) >> 7
				i++
			}
		}
		return newABb(r)
	}
	if y <= 256 { // <= because up to y (excluded)
		r := make([]byte, n)
		for i := range r {
			r[i] = byte(ctx.rand.Int63n(y))
		}
		return NewAB(r)
	}
	r := make([]int64, n)
	for i := range r {
		r[i] = ctx.rand.Int63n(y)
	}
	return NewAI(r)
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

func deal(ctx *Context, n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		if yv <= 0 {
			return Panicf("(-i)?y : non-positive y (%d)", yv)
		}
		if n > yv {
			return Panicf("(-i)?y : i > y (%d vs %d)", n, yv)
		}
		if yv <= 256 { // <= because up to yv (excluded)
			r := dealI[byte](ctx, n, yv)
			return NewV(&AB{elts: r, flags: flagDistinct})
		}
		r := dealI[int64](ctx, n, yv)
		return NewV(&AI{elts: r, flags: flagDistinct})
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("(-i)?y : non-integer y (%g)", y.F())
		}
		return deal(ctx, n, NewI(int64(y.F())))
	}
	ya, ok := y.bv.(Array)
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
	if n == math.MinInt64 {
		n = int64(ylen)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return NewV(&AB{elts: dealSlice[byte](ctx, n, yv.elts), flags: (flagDistinct | flagBool) & yv.flags})
	case *AI:
		return NewV(&AI{elts: dealSlice[int64](ctx, n, yv.elts), flags: flagDistinct & yv.flags})
	case *AF:
		return NewV(&AF{elts: dealSlice[float64](ctx, n, yv.elts), flags: flagDistinct & yv.flags})
	case *AS:
		return NewV(&AS{elts: dealSlice[string](ctx, n, yv.elts), flags: flagDistinct & yv.flags})
	case *AV:
		return canonicalAV(&AV{elts: dealSlice[V](ctx, n, yv.elts), flags: flagDistinct & yv.flags, rc: yv.rc})
	default:
		panic("deal")
	}
}

func dealI[I integer](ctx *Context, n, y int64) []I {
	if n*4+8 < y {
		// For small n, we use hashing, not the fastest
		// algorithm, but not too bad either.
		r := make([]I, n)
		m := map[I]struct{}{}
		var i int64
		for i < n {
			k := I(ctx.rand.Int63n(y))
			_, ok := m[k]
			if ok {
				continue
			}
			m[k] = struct{}{}
			r[i] = k
			i++
		}
		return r
	}
	r := make([]I, y)
	for i := range r {
		r[i] = I(i)
	}
	ctx.rand.Shuffle(int(y), func(i, j int) {
		r[i], r[j] = r[j], r[i]
	})
	r = r[:n]
	return r
}

func dealSlice[T any](ctx *Context, n int64, y []T) []T {
	ylen := len(y)
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
