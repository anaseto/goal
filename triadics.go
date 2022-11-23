package goal

//import "fmt"

func (ctx *Context) amend3(v, x, f V) V {
	switch v := v.(type) {
	case array:
		x = toIndices(x)
		if err, ok := x.(errV); ok {
			return err
		}
		return canonical(ctx.amend3array(cloneShallow(v).(array), x, f))
	default:
		return errType("@[v;x;f]", "v", v)
	}
}

func (ctx *Context) amend3arrayI(v array, x I, f V) V {
	if x < 0 || int(x) >= v.Len() {
		return errf("@[v;x;f] : x out of bounds (%d)", x)
	}
	z := v.at(int(x))
	repl := ctx.Apply(f, z)
	if err, ok := repl.(errV); ok {
		return err
	}
	if compatEltType(v, repl) {
		v.set(int(x), repl)
		return v
	}
	a := make(AV, v.Len())
	for i := range a {
		a[i] = v.at(i)
	}
	a.set(int(x), repl)
	return a
}

func (ctx *Context) amend3array(v array, x, f V) V {
	switch x := x.(type) {
	case I:
		return ctx.amend3arrayI(v, x, f)
	case AI:
		for _, idx := range x {
			v = ctx.amend3array(v, I(idx), f).(array)
		}
		return v
	case AV:
		for _, z := range x {
			v = ctx.amend3array(v, z, f).(array)
		}
		return v
	default:
		return errType("@[v;x;f]", "x", x)
	}
}

func (ctx *Context) amend4(v, x, f, y V) V {
	switch v := v.(type) {
	case array:
		x = toIndices(x)
		if err, ok := x.(errV); ok {
			return err
		}
		return canonical(ctx.amend4array(cloneShallow(v).(array), x, f, y))
	default:
		return errType("@[v;x;f]", "v", v)
	}
}

func (ctx *Context) amend4arrayI(v array, x I, f, y V) V {
	if x < 0 || int(x) >= v.Len() {
		return errf("@[v;x;f;y] : x out of bounds (%d)", x)
	}
	z := v.at(int(x))
	repl := ctx.Apply2(f, z, y)
	if err, ok := repl.(errV); ok {
		return err
	}
	if compatEltType(v, repl) {
		v.set(int(x), repl)
		return v
	}
	a := make(AV, v.Len())
	for i := range a {
		a[i] = v.at(i)
	}
	a.set(int(x), repl)
	return a
}

func (ctx *Context) amend4array(v array, x, f, y V) V {
	switch x := x.(type) {
	case I:
		switch y.(type) {
		case array:
			return errs("@[v;x;f;y] : shape mismatch between x and y")
		}
		return ctx.amend4arrayI(v, x, f, y)
	case AI:
		ay, ok := y.(array)
		if !ok {
			for _, idx := range x {
				v = ctx.amend4arrayI(v, I(idx), f, y).(array)
			}
			return v
		}
		if ay.Len() != x.Len() {
			return errf("@[v;x;f;y] : length mismatch between x and y (%d vs %d)",
				x.Len(), ay.Len())
		}
		for i, idx := range x {
			v = ctx.amend4arrayI(v, I(idx), f, ay.at(i)).(array)
		}
		return v
	case AV:
		ay, ok := y.(array)
		if !ok {
			for _, z := range x {
				v = ctx.amend4array(v, z, f, y).(array)
			}
			return v
		}
		if ay.Len() != x.Len() {
			return errf("@[v;x;f;y] : length mismatch between x and y (%d vs %d)",
				x.Len(), ay.Len())
		}
		for i, z := range x {
			v = ctx.amend4array(v, z, f, ay.at(i)).(array)
		}
		return v
	default:
		return errType("@[v;x;f]", "x", x)
	}
}
