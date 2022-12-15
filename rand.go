package goal

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
	if n < 0 {
		return deal(ctx, -n, y)
		//return Panicf("i?y : i negative integer (%d)", n)
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
		if y.I() <= 0 {
			return Panicf("(-i)?y : y non-positive (%d)", y.I())
		}
		if n > y.I() {
			return Panicf("(-i)?y : i > y (%d vs %d)", n, y.I())
		}
		r := make([]int64, y.I())
		for i := range r {
			r[i] = int64(i)
		}
		// TODO: deal can be improved in cases where n is much smaller
		// than y, by using a more involved algorithm.
		ctx.rand.Shuffle(int(y.I()), func(i, j int) {
			r[i], r[j] = r[j], r[i]
		})
		r = r[:n]
		return NewAI(r)
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
		ctx.rand.Seed(x.I())
		return x
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("seed x : x not an integer (%g)", x.F())
		}
		ctx.rand.Seed(int64(x.F()))
		return NewI(int64(x.F()))
	}
	return panicType("seed x", "x", x)
}
