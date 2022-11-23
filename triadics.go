package goal

//import "fmt"

func (ctx *Context) amend3(x, y, f V) V {
	switch x := x.(type) {
	case array:
		y = toIndices(y)
		if err, ok := y.(errV); ok {
			return err
		}
		return canonical(ctx.amend3array(cloneShallow(x).(array), y, f))
	default:
		return errType("@[x;y;f]", "x", x)
	}
}

func (ctx *Context) amend3arrayI(x array, y I, f V) V {
	if y < 0 || int(y) >= x.Len() {
		return errf("@[x;y;f] : x out of bounds (%d)", y)
	}
	z := x.at(int(y))
	repl := ctx.Apply(f, z)
	if err, ok := repl.(errV); ok {
		return err
	}
	if compatEltType(x, repl) {
		x.set(int(y), repl)
		return x
	}
	a := make(AV, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a.set(int(y), repl)
	return a
}

func (ctx *Context) amend3array(x array, y, f V) V {
	switch y := y.(type) {
	case I:
		return ctx.amend3arrayI(x, y, f)
	case AI:
		for _, idx := range y {
			x = ctx.amend3array(x, I(idx), f).(array)
		}
		return x
	case AV:
		for _, z := range y {
			x = ctx.amend3array(x, z, f).(array)
		}
		return x
	default:
		return errType("@[x;y;f]", "y", y)
	}
}

func (ctx *Context) amend4(x, y, f, z V) V {
	switch x := x.(type) {
	case array:
		y = toIndices(y)
		if err, ok := y.(errV); ok {
			return err
		}
		return canonical(ctx.amend4array(cloneShallow(x).(array), y, f, z))
	default:
		return errType("@[x;y;f]", "x", x)
	}
}

func (ctx *Context) amend4arrayI(x array, y I, f, z V) V {
	if y < 0 || int(y) >= x.Len() {
		return errf("@[x;y;f;z] : x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply2(f, xy, z)
	if err, ok := repl.(errV); ok {
		return err
	}
	if compatEltType(x, repl) {
		x.set(int(y), repl)
		return x
	}
	a := make(AV, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a.set(int(y), repl)
	return a
}

func (ctx *Context) amend4array(x array, y, f, z V) V {
	switch y := y.(type) {
	case I:
		switch z.(type) {
		case array:
			return errs("@[x;y;f;z] : shape mismatch between x and y")
		}
		return ctx.amend4arrayI(x, y, f, z)
	case AI:
		ay, ok := z.(array)
		if !ok {
			for _, xi := range y {
				x = ctx.amend4arrayI(x, I(xi), f, z).(array)
			}
			return x
		}
		if ay.Len() != y.Len() {
			return errf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				y.Len(), ay.Len())
		}
		for i, xi := range y {
			x = ctx.amend4arrayI(x, I(xi), f, ay.at(i)).(array)
		}
		return x
	case AV:
		ay, ok := z.(array)
		if !ok {
			for _, xi := range y {
				x = ctx.amend4array(x, xi, f, z).(array)
			}
			return x
		}
		if ay.Len() != y.Len() {
			return errf("@[x;y;f;z] : length mismatch between x and y (%d vs %d)",
				y.Len(), ay.Len())
		}
		for i, xi := range y {
			x = ctx.amend4array(x, xi, f, ay.at(i)).(array)
		}
		return x
	default:
		return errType("@[x;y;f]", "y", y)
	}
}
