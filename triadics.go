package goal

//import "fmt"

// amend3 implements @[x;y;f].
func (ctx *Context) amend3(x, y, f V) V {
	switch x := x.Value.(type) {
	case array:
		y = toIndices(y)
		if y.IsErr() {
			return y
		}
		return canonicalV(ctx.amend3array(cloneShallowArray(x), y, f))
	default:
		return errType("@[x;y;f]", "x", x)
	}
}

func (ctx *Context) amend3arrayI(x array, y I, f V) V {
	if y < 0 || int(y) >= x.Len() {
		return errf("@[x;y;f] : x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply(f, xy)
	if repl.IsErr() {
		return errf("f call in @[x;y;f] : %v", repl)
	}
	if compatEltType(x, repl) {
		x.set(int(y), repl)
		return NewV(x)
	}
	a := make(AV, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a.set(int(y), repl)
	return NewV(a)
}

func (ctx *Context) amend3array(x array, y, f V) V {
	switch y := y.Value.(type) {
	case I:
		return ctx.amend3arrayI(x, y, f)
	case AI:
		for _, yi := range y {
			ax := ctx.amend3arrayI(x, I(yi), f)
			if ax.IsErr() {
				return ax
			}
			x = ax.Value.(array)
		}
		return NewV(x)
	case AV:
		for _, yi := range y {
			ax := ctx.amend3array(x, yi, f)
			if ax.IsErr() {
				return ax
			}
			x = ax.Value.(array)
		}
		return NewV(x)
	default:
		return errType("@[x;y;f]", "y", y)
	}
}

// amend4 implements @[x;y;f;z].
func (ctx *Context) amend4(x, y, f, z V) V {
	switch x := x.Value.(type) {
	case array:
		y = toIndices(y)
		if y.IsErr() {
			return y
		}
		return canonicalV(ctx.amend4array(cloneShallowArray(x), y, f, z))
	default:
		return errType("@[x;y;f;z]", "x", x)
	}
}

func (ctx *Context) amend4arrayI(x array, y I, f, z V) V {
	if y < 0 || int(y) >= x.Len() {
		return errf("@[x;y;f;z] : x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply2(f, xy, z)
	if repl.IsErr() {
		return errf("f call in @[x;y;f;z] : %v", repl)
	}
	if compatEltType(x, repl) {
		x.set(int(y), repl)
		return NewV(x)
	}
	a := make(AV, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a.set(int(y), repl)
	return NewV(a)
}

func (ctx *Context) amend4array(x array, y, f, z V) V {
	switch y := y.Value.(type) {
	case I:
		switch z.Value.(type) {
		case array:
			return errs("@[x;y;f;z] : shape mismatch between x and y")
		}
		return ctx.amend4arrayI(x, y, f, z)
	case AI:
		az, ok := z.Value.(array)
		if !ok {
			for _, xi := range y {
				ax := ctx.amend4arrayI(x, I(xi), f, z)
				if ax.IsErr() {
					return ax
				}
				x = ax.Value.(array)
			}
			return NewV(x)
		}
		if az.Len() != y.Len() {
			return errf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				y.Len(), az.Len())
		}
		for i, xi := range y {
			ax := ctx.amend4arrayI(x, I(xi), f, az.at(i))
			if ax.IsErr() {
				return ax
			}
			x = ax.Value.(array)
		}
		return NewV(x)
	case AV:
		az, ok := z.Value.(array)
		if !ok {
			for _, xi := range y {
				ax := ctx.amend4array(x, xi, f, z)
				if ax.IsErr() {
					return ax
				}
				x = ax.Value.(array)
			}
			return NewV(x)
		}
		if az.Len() != y.Len() {
			return errf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				y.Len(), az.Len())
		}
		for i, xi := range y {
			ax := ctx.amend4array(x, xi, f, az.at(i))
			if ax.IsErr() {
				return ax
			}
			x = ax.Value.(array)
		}
		return NewV(x)
	default:
		return errType("@[x;y;f;z]", "y", y)
	}
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).Value.(array)
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	r := ctx.applyN(f1, av.Len())
	if err, ok := r.Value.(errV); ok {
		ctx.push(NewS(string(err)))
		r = ctx.applyN(f2, 1)
		if r.IsErr() {
			return errf("f2 call in .[f1;x;f2] : %v", r)
		}
	}
	return r
}
