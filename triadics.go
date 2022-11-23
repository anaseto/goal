package goal

func (ctx *Context) amend3(v, x, f V) V {
	x = toIndices(x)
	if err, ok := x.(errV); ok {
		return err
	}
	switch v := v.(type) {
	case array:
		return ctx.amend3array(v, x, f)
	default:
		return errType("@[v;x;f]", "v", v)
	}
}

func (ctx *Context) amend3array(v array, x, f V) V {
	switch x := x.(type) {
	case I:
		if x < 0 || int(x) >= v.Len() {
			return errf("@[v;x;f] : x out of bounds (%d)", x)
		}
		z := v.at(int(x))
		repl := ctx.Apply(f, z)
		if err, ok := repl.(errV); ok {
			return err
		}
		if sameEltType(v, repl) {
			v.set(int(x), repl)
			return v
		}
		a := make(AV, v.Len())
		for i := range a {
			a[i] = v.at(i)
		}
		a.set(int(x), repl)
		return canonical(a)
	case AI:
		z := v.atIndices(x)
		if err, ok := z.(errV); ok {
			return err
		}
		repl := ctx.Apply(f, z)
		if err, ok := repl.(errV); ok {
			return err
		}
		repl = toArray(repl)
		if Length(repl) != Length(z) {
			return errs("@[v;x;f] : x and f[x] should have same length")
		}
		if sameType(v, repl) {
			v.setIndices(x, repl)
			return v
		}
		a := make(AV, v.Len())
		for i := range a {
			a[i] = v.at(i)
		}
		a.setIndices(x, repl)
		return canonical(a)
	case AV:
		a := make(AV, v.Len())
		for i, z := range x {
			a[i] = ctx.amend3array(v, z, f)
			if err, ok := a[i].(errV); ok {
				return err
			}
		}
		return canonical(a)
	default:
		return errType("@[v;x;f]", "x", x)
	}
}

func (ctx *Context) amend4(v, x, f, y V) V {
	return errNYI("@[v;x;f;y]")
}
