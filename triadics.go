package goal

//import "fmt"

// amend3 implements @[x;y;f].
func (ctx *Context) amend3(x, y, f V) V {
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return y
		}
		return Canonical(ctx.amend3array(cloneShallowArray(xv), y, f))
	default:
		return panicType("@[x;y;f]", "x", x)
	}
}

func (ctx *Context) amend3arrayI(x array, y int64, f V) V {
	if y < 0 || y >= int64(x.Len()) {
		return Panicf("@[x;y;f] : x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply(f, xy)
	if repl.IsPanic() {
		return Panicf("f call in @[x;y;f] : %v", repl)
	}
	if isEltType(x, repl) {
		x.set(int(y), repl)
		return NewV(x)
	}
	a := make([]V, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a[y] = repl
	return NewAV(a)
}

func (ctx *Context) amend3array(x array, y, f V) V {
	if y.IsI() {
		return ctx.amend3arrayI(x, y.I(), f)
	}
	switch yv := y.value.(type) {
	case *AI:
		for _, yi := range yv.Slice {
			ax := ctx.amend3arrayI(x, yi, f)
			if ax.IsPanic() {
				return ax
			}
			x = ax.value.(array)
		}
		return NewV(x)
	case *AV:
		for _, yi := range yv.Slice {
			ax := ctx.amend3array(x, yi, f)
			if ax.IsPanic() {
				return ax
			}
			x = ax.value.(array)
		}
		return NewV(x)
	default:
		return panicType("@[x;y;f]", "y", y)
	}
}

// amend4 implements @[x;y;f;z].
func (ctx *Context) amend4(x, y, f, z V) V {
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return y
		}
		return Canonical(ctx.amend4array(cloneShallowArray(xv), y, f, z))
	default:
		return panicType("@[x;y;f;z]", "x", x)
	}
}

func (ctx *Context) amend4arrayI(x array, y int64, f, z V) V {
	if y < 0 || y >= int64(x.Len()) {
		return Panicf("@[x;y;f;z] : x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply2(f, xy, z)
	if repl.IsPanic() {
		return Panicf("f call in @[x;y;f;z] : %v", repl)
	}
	if isEltType(x, repl) {
		x.set(int(y), repl)
		return NewV(x)
	}
	a := make([]V, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a[y] = repl
	return NewAV(a)
}

func (ctx *Context) amend4array(x array, y, f, z V) V {
	if y.IsI() {
		switch z.value.(type) {
		case array:
			return panics("@[x;y;f;z] : shape mismatch between x and y")
		}
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	switch yv := y.value.(type) {
	case *AI:
		az, ok := z.value.(array)
		if !ok {
			for _, xi := range yv.Slice {
				ax := ctx.amend4arrayI(x, xi, f, z)
				if ax.IsPanic() {
					return ax
				}
				x = ax.value.(array)
			}
			return NewV(x)
		}
		if az.Len() != yv.Len() {
			return Panicf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				yv.Len(), az.Len())

		}
		for i, xi := range yv.Slice {
			ax := ctx.amend4arrayI(x, xi, f, az.at(i))
			if ax.IsPanic() {
				return ax
			}
			x = ax.value.(array)
		}
		return NewV(x)
	case *AV:
		az, ok := z.value.(array)
		if !ok {
			for _, xi := range yv.Slice {
				ax := ctx.amend4array(x, xi, f, z)
				if ax.IsPanic() {
					return ax
				}
				x = ax.value.(array)
			}
			return NewV(x)
		}
		if az.Len() != yv.Len() {
			return Panicf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				yv.Len(), az.Len())

		}
		for i, xi := range yv.Slice {
			ax := ctx.amend4array(x, xi, f, az.at(i))
			if ax.IsPanic() {
				return ax
			}
			x = ax.value.(array)
		}
		return NewV(x)
	default:
		return panicType("@[x;y;f;z]", "y", y)
	}
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).value.(array)
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	r := ctx.applyN(f1, av.Len())
	if r.IsPanic() {
		r.kind = valBoxed // we used the boxed value
		ctx.push(r)
		r = ctx.applyN(f2, 1)
		if r.IsPanic() {
			return Panicf("f2 call in .[f1;x;f2] : %v", r)
		}
	}
	return r
}
