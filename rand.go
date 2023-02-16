package goal

import "math/rand"

func uniform(ctx *Context, x V) V {
	var n int64
	if x.IsI() {
		n = x.I()
	} else {
		if !isI(x.F()) {
			return Panicf("?i : not an integer (%g)", x.F())
		}
		n = int64(x.F())
	}
	if n < 0 {
		return Panicf("?i : negative integer (%d)", n)
	}
	if ctx.rand == nil {
		ctx.rand = rand.New(rand.NewSource(1))
	}
	r := make([]float64, n)
	for i := range r {
		r[i] = ctx.rand.Float64()
	}
	return NewAF(r)
}

func roll(ctx *Context, x, y V) V {
	var n int64
	if x.IsI() {
		n = x.I()
	} else {
		if !isI(x.F()) {
			return Panicf("i?y : i not an integer (%g)", x.F())
		}
		n = int64(x.F())
	}
	if ctx.rand == nil {
		ctx.rand = rand.New(rand.NewSource(1))
	}
	if n < 0 {
		return deal(ctx, -n, y)
	}
	if y.IsI() {
		if y.I() <= 0 {
			return Panicf("i?y : y non-positive (%d)", y.I())
		}
		r := make([]int64, n)
		for i := range r {
			r[i] = ctx.rand.Int63n(y.I())
		}
		return NewAI(r)
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("i?y : y not an integer (%g)", y.F())
		}
		return roll(ctx, x, NewI(int64(y.F())))
	}
	return panicType("i?y", "y", y)
}

func deal(ctx *Context, n int64, y V) V {
	if y.IsI() {
		yv := y.I()
		if yv <= 0 {
			return Panicf("(-i)?y : y non-positive (%d)", yv)
		}
		if n > yv {
			return Panicf("(-i)?y : i > y (%d vs %d)", n, yv)
		}
		if n == 0 {
			return NewAI(nil)
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
			return NewV(&AI{Slice: r, flags: flagUnique})
		}
		r := make([]int64, yv)
		for i := range r {
			r[i] = int64(i)
		}
		ctx.rand.Shuffle(int(yv), func(i, j int) {
			r[i], r[j] = r[j], r[i]
		})
		r = r[:n]
		return NewV(&AI{Slice: r, flags: flagUnique})
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("(-i)?y : y not an integer (%g)", y.F())
		}
		return deal(ctx, n, NewI(int64(y.F())))
	}
	return panicType("(-i)?y", "y", y)
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
			return Panicf("seed x : x not an integer (%g)", x.F())
		}
		if ctx.rand == nil {
			ctx.rand = rand.New(rand.NewSource(int64(x.F())))
			return NewI(int64(x.F()))
		}
		ctx.rand.Seed(int64(x.F()))
		return NewI(int64(x.F()))
	}
	return panicType("seed x", "x", x)
}
