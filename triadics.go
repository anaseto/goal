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
			xa := ctx.amend3arrayI(x, yi, f)
			if xa.IsPanic() {
				return xa
			}
			x = xa.value.(array)
		}
		return NewV(x)
	case *AV:
		for _, yi := range yv.Slice {
			xa := ctx.amend3array(x, yi, f)
			if xa.IsPanic() {
				return xa
			}
			x = xa.value.(array)
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
		if f.kind == valVariadic && variadic(f.n) == vRight {
			return Canonical(amendr(cloneShallowArray(xv), y, z))
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
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.at(i)
	}
	r[y] = repl
	return NewAV(r)
}

func (ctx *Context) amend4array(x array, y, f, z V) V {
	if y.IsI() {
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	switch yv := y.value.(type) {
	case *AI:
		za, ok := z.value.(array)
		if !ok {
			for _, yi := range yv.Slice {
				xa := ctx.amend4arrayI(x, yi, f, z)
				if xa.IsPanic() {
					return xa
				}
				x = xa.value.(array)
			}
			return NewV(x)
		}
		if za.Len() != yv.Len() {
			return Panicf("@[x;y;f;z] : length mismatch between y and z (%d vs %d)",
				yv.Len(), za.Len())

		}
		for i, yi := range yv.Slice {
			xa := ctx.amend4arrayI(x, yi, f, za.at(i))
			if xa.IsPanic() {
				return xa
			}
			x = xa.value.(array)
		}
		return NewV(x)
	case *AV:
		za, ok := z.value.(array)
		if !ok {
			for _, yi := range yv.Slice {
				xa := ctx.amend4array(x, yi, f, z)
				if xa.IsPanic() {
					return xa
				}
				x = xa.value.(array)
			}
			return NewV(x)
		}
		if za.Len() != yv.Len() {
			return Panicf("@[x;y;f;z] : length mismatch between y and z (%d vs %d)",
				yv.Len(), za.Len())

		}
		for i, yi := range yv.Slice {
			xa := ctx.amend4array(x, yi, f, za.at(i))
			if xa.IsPanic() {
				return xa
			}
			x = xa.value.(array)
		}
		return NewV(x)
	default:
		return panicType("@[x;y;f;z]", "y", y)
	}
}

func outOfBounds(y int64, l int) bool {
	return y < 0 || y >= int64(l)
}

func amendr(x array, y, z V) V {
	if y.IsI() {
		if outOfBounds(y.I(), x.Len()) {
			return Panicf("@[x;y;:;z] : y out of bounds (%d)", y.I())
		}
		if isEltType(x, z) {
			x.set(int(y.I()), z)
			return NewV(x)
		}
		r := make([]V, x.Len())
		for i := range r {
			r[i] = x.at(i)
		}
		r[y.I()] = z
		return NewAV(r)
	}
	switch yv := y.value.(type) {
	case *AI:
		return amendrAI(x, yv, z)
	case *AV:
		return amendrAV(x, yv, z)
	default:
		return panicType("@[x;y;:;z]", "y", y)
	}
}

func amendrAI(x array, yv *AI, z V) V {
	xlen := x.Len()
	for _, yi := range yv.Slice {
		if outOfBounds(yi, xlen) {
			return Panicf("@[x;y;:;z] : out of bounds index (%d)", yi)
		}
	}
	za, ok := z.value.(array)
	if !ok {
		if isEltType(x, z) {
			for _, yi := range yv.Slice {
				x.set(int(yi), z)
			}
			return NewV(x)
		}
		r := make([]V, xlen)
		for i := range r {
			r[i] = x.at(i)
		}
		for _, yi := range yv.Slice {
			r[yi] = z
		}
		return NewAV(r)
	}
	if za.Len() != yv.Len() {
		return Panicf("@[x;y;:;z] : length mismatch between y and z (%d vs %d)",
			yv.Len(), za.Len())
	}
	if sameType(x, za) {
		switch xv := x.(type) {
		case *AB:
			zv := za.(*AB)
			for i, yi := range yv.Slice {
				xv.Slice[yi] = zv.Slice[i]
			}
		case *AI:
			zv := za.(*AI)
			for i, yi := range yv.Slice {
				xv.Slice[yi] = zv.Slice[i]
			}
		case *AF:
			zv := za.(*AF)
			for i, yi := range yv.Slice {
				xv.Slice[yi] = zv.Slice[i]
			}
		case *AS:
			zv := za.(*AS)
			for i, yi := range yv.Slice {
				xv.Slice[yi] = zv.Slice[i]
			}
		case *AV:
			zv := za.(*AV)
			for i, yi := range yv.Slice {
				xv.Slice[yi] = zv.Slice[i]
			}
		}
		return NewV(x)
	}
	for i := range yv.Slice {
		if !isEltType(x, za.at(i)) {
			r := make([]V, xlen)
			for i := range r {
				r[i] = x.at(i)
			}
			x = &AV{Slice: r}
			break
		}
	}
	for i, yi := range yv.Slice {
		x.set(int(yi), za.at(i))
	}
	return NewV(x)
}

func amendrAV(x array, yv *AV, z V) V {
	za, ok := z.value.(array)
	if !ok {
		for _, yi := range yv.Slice {
			xa := amendr(x, yi, z)
			if xa.IsPanic() {
				return xa
			}
			x = xa.value.(array)
		}
		return NewV(x)
	}
	if za.Len() != yv.Len() {
		return Panicf("@[x;y;:;z] : length mismatch between y and z (%d vs %d)",
			yv.Len(), za.Len())

	}
	for i, yi := range yv.Slice {
		xa := amendr(x, yi, za.at(i))
		if xa.IsPanic() {
			return xa
		}
		x = xa.value.(array)
	}
	return NewV(x)
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).value.(array)
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	r := ctx.applyN(f1, av.Len())
	if r.IsPanic() {
		r = NewS(string(r.value.(panicV)))
		ctx.push(r)
		r = ctx.applyN(f2, 1)
		if r.IsPanic() {
			return Panicf("f2 call in .[f1;x;f2] : %v", r)
		}
	}
	return r
}
