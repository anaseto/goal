package goal

import (
	"errors"
	"fmt"
)

// amend3 implements @[x;y;f].
func (ctx *Context) amend3(x, y, f V) V {
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[x;y;f] : y ", y)
		}
		r, err := ctx.amend3array(cloneShallowArray(xv), y, f)
		if err != nil {
			return Panicf("@[x;y;f] : %v", err)
		}
		return Canonical(NewV(r))
	default:
		return panicType("@[x;y;f]", "x", x)
	}
}

func amendArrayAt(x array, y int, z V) array {
	if isEltType(x, z) {
		x.set(y, z)
		return x
	}
	a := make([]V, x.Len())
	for i := range a {
		a[i] = x.at(i)
	}
	a[y] = z
	return &AV{Slice: a}
}

func (ctx *Context) amend3arrayI(x array, y int64, f V) (array, error) {
	if outOfBounds(y, x.Len()) {
		return x, fmt.Errorf("y out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply(f, xy)
	if repl.IsPanic() {
		return x, fmt.Errorf("f call: %v", repl)
	}
	return amendArrayAt(x, int(y), repl), nil
}

func (ctx *Context) amend3array(x array, y, f V) (array, error) {
	if y.IsI() {
		return ctx.amend3arrayI(x, y.I(), f)
	}
	switch yv := y.value.(type) {
	case *AI:
		var err error
		for _, yi := range yv.Slice {
			x, err = ctx.amend3arrayI(x, yi, f)
			if err != nil {
				return x, err
			}
		}
		return x, nil
	case *AV:
		var err error
		for _, yi := range yv.Slice {
			x, err = ctx.amend3array(x, yi, f)
			if err != nil {
				return x, err
			}
		}
		return x, nil
	default:
		panic("amend3array: y bad type")
	}
}

// amend4 implements @[x;y;f;z].
func (ctx *Context) amend4(x, y, f, z V) V {
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[x;y;f;z] : y ", y)
		}
		if f.kind == valVariadic && variadic(f.n) == vRight {
			r, err := amendr(cloneShallowArray(xv), y, z)
			if err != nil {
				return Panicf("@[x;y;:;z] : %v", err)
			}
			return Canonical(NewV(r))
		}
		r, err := ctx.amend4array(cloneShallowArray(xv), y, f, z)
		if err != nil {
			return Panicf("@[x;y;f;z] : %v", err)
		}
		return Canonical(NewV(r))
	default:
		return panicType("@[x;y;f;z]", "x", x)
	}
}

func (ctx *Context) amend4arrayI(x array, y int64, f, z V) (array, error) {
	if y < 0 || y >= int64(x.Len()) {
		return x, fmt.Errorf("x out of bounds (%d)", y)
	}
	xy := x.at(int(y))
	repl := ctx.Apply2(f, xy, z)
	if repl.IsPanic() {
		return x, fmt.Errorf("f call: %v", repl)
	}
	return amendArrayAt(x, int(y), repl), nil
}

func (ctx *Context) amend4array(x array, y, f, z V) (array, error) {
	if y.IsI() {
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	switch yv := y.value.(type) {
	case *AI:
		var err error
		za, ok := z.value.(array)
		if !ok {
			for _, yi := range yv.Slice {
				x, err = ctx.amend4arrayI(x, yi, f, z)
				if err != nil {
					return x, err
				}
			}
			return x, nil
		}
		if za.Len() != yv.Len() {
			return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
				yv.Len(), za.Len())

		}
		for i, yi := range yv.Slice {
			x, err = ctx.amend4arrayI(x, yi, f, za.at(i))
			if err != nil {
				return x, err
			}
		}
		return x, nil
	case *AV:
		var err error
		za, ok := z.value.(array)
		if !ok {
			for _, yi := range yv.Slice {
				x, err = ctx.amend4array(x, yi, f, z)
				if err != nil {
					return x, err
				}
			}
			return x, nil
		}
		if za.Len() != yv.Len() {
			return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
				yv.Len(), za.Len())

		}
		for i, yi := range yv.Slice {
			x, err = ctx.amend4array(x, yi, f, za.at(i))
			if err != nil {
				return x, err
			}
		}
		return x, nil
	default:
		panic("amend4array: y bad type")
	}
}

func outOfBounds(y int64, l int) bool {
	return y < 0 || y >= int64(l)
}

func amendr(x array, y, z V) (array, error) {
	if y.IsI() {
		if outOfBounds(y.I(), x.Len()) {
			return x, fmt.Errorf("y out of bounds (%d)", y.I())
		}
		if isEltType(x, z) {
			x.set(int(y.I()), z)
			return x, nil
		}
		r := make([]V, x.Len())
		for i := range r {
			r[i] = x.at(i)
		}
		r[y.I()] = z
		return &AV{Slice: r}, nil
	}
	switch yv := y.value.(type) {
	case *AI:
		return amendrAI(x, yv, z)
	case *AV:
		return amendrAV(x, yv, z)
	default:
		panic("amendr: y bad type")
	}
}

func amendrAI(x array, yv *AI, z V) (array, error) {
	xlen := x.Len()
	for _, yi := range yv.Slice {
		if outOfBounds(yi, xlen) {
			return x, fmt.Errorf("out of bounds index (%d)", yi)
		}
	}
	za, ok := z.value.(array)
	if !ok {
		if isEltType(x, z) {
			amendrAIatomMut(x, yv, z)
			return x, nil
		}
		r := make([]V, xlen)
		for i := range r {
			r[i] = x.at(i)
		}
		for _, yi := range yv.Slice {
			r[yi] = z
		}
		return &AV{Slice: r}, nil
	}
	if za.Len() != yv.Len() {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			yv.Len(), za.Len())
	}
	if sameType(x, za) {
		amendrAIarrayMut(x, yv, za)
		return x, nil
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
	return x, nil
}

func amendrAIatomMut(x array, yv *AI, z V) {
	switch xv := x.(type) {
	case *AB:
		var zb bool
		if z.IsI() {
			zb = z.I() != 0
		} else {
			zb = z.F() != 0
		}
		for _, yi := range yv.Slice {
			xv.Slice[int(yi)] = zb
		}
	case *AI:
		var zi int64
		if z.IsI() {
			zi = z.I()
		} else {
			zi = int64(z.F())
		}
		for _, yi := range yv.Slice {
			xv.Slice[yi] = zi
		}
	case *AF:
		var zf float64
		if z.IsI() {
			zf = float64(z.I())
		} else {
			zf = z.F()
		}
		for _, yi := range yv.Slice {
			xv.Slice[yi] = zf
		}
	case *AS:
		zs := string(z.value.(S))
		for _, yi := range yv.Slice {
			xv.Slice[yi] = zs
		}
	case *AV:
		for _, yi := range yv.Slice {
			xv.Slice[yi] = z
		}
	}
}

func amendrAIarrayMut(x array, yv *AI, za array) {
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
}

func amendrAV(x array, yv *AV, z V) (array, error) {
	var err error
	za, ok := z.value.(array)
	if !ok {
		for _, yi := range yv.Slice {
			x, err = amendr(x, yi, z)
			if err != nil {
				return x, err
			}
		}
		return x, nil
	}
	if za.Len() != yv.Len() {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			yv.Len(), za.Len())

	}
	for i, yi := range yv.Slice {
		x, err = amendr(x, yi, za.at(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	x.Slice[i] = y
}

// set changes x at i with y (in place).
func (x *AB) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n != 0
	} else {
		x.Slice[i] = y.F() != 0
	}
}

// set changes x at i with y (in place).
func (x *AI) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n
	} else {
		x.Slice[i] = int64(y.F())
	}
}

// set changes x at i with y (in place).
func (x *AF) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = float64(y.I())
	} else {
		x.Slice[i] = y.F()
	}
}

// set changes x at i with y (in place).
func (x *AS) set(i int, y V) {
	x.Slice[i] = string(y.value.(S))
}

// deepAmend3 implements .[x;y;f].
func (ctx *Context) deepAmend3(x, y, f V) V {
	x = clone(x)
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic(".[x;y;f] : y ", y)
		}
		x, err := ctx.deepAmend3array(xv, y, f)
		if err != nil {
			return Panicf(".[x;y;f] : %v", err)
		}
		return Canonical(NewV(x))
	default:
		return panicType(".[x;y;f]", "x", x)
	}
}

func (ctx *Context) deepAmend3array(x array, y, f V) (array, error) {
	if y.IsI() {
		return ctx.amend3arrayI(x, y.I(), f)
	}
	yv := y.value.(array)
	if yv.Len() == 0 {
		return ctx.deepAmend3rec(x, rangeI(int64(x.Len())), yv, f, false)
	}
	return ctx.deepAmend3rec(x, yv.at(0), yv.slice(1, yv.Len()), f, false)
}

func (ctx *Context) deepAmend3rec(x array, y0 V, y array, f V, depth bool) (array, error) {
	y0v, ok := y0.value.(array)
	if ok && y0v.Len() == 0 && !depth {
		return ctx.deepAmend3rec(x, rangeI(int64(x.Len())), y, f, depth)
	}
	if y.Len() == 0 {
		return ctx.amend3array(x, y0, f)
	}
	if y0.IsI() {
		if outOfBounds(y0.I(), x.Len()) {
			return x, fmt.Errorf("y out of bounds (%d)", y0.I())
		}
		xy0 := x.at(int(y0.I()))
		xy0v, ok := xy0.value.(array)
		if !ok {
			return x, errors.New("y out of depth")
		}
		repl, err := ctx.deepAmend3rec(xy0v, y.at(0), y.slice(1, y.Len()), f, false)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	var err error
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.at(i)
		x, err = ctx.deepAmend3rec(x, y0i, y, f, true)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// deepAmend4 implements .[x;y;f].
func (ctx *Context) deepAmend4(x, y, f, z V) V {
	x = clone(x)
	switch xv := x.value.(type) {
	case array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic(".[x;y;f] : y ", y)
		}
		x, err := ctx.deepAmend4array(xv, y, f, z)
		if err != nil {
			return Panicf(".[x;y;f] : %v", err)
		}
		return Canonical(NewV(x))
	default:
		return panicType(".[x;y;f]", "x", x)
	}
}

func (ctx *Context) deepAmend4array(x array, y, f, z V) (array, error) {
	if y.IsI() {
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	yv := y.value.(array)
	if yv.Len() == 0 {
		return ctx.deepAmend4rec(x, rangeI(int64(x.Len())), yv, f, z, false)
	}
	return ctx.deepAmend4rec(x, yv.at(0), yv.slice(1, yv.Len()), f, z, false)
}

func (ctx *Context) deepAmend4rec(x array, y0 V, y array, f, z V, depth bool) (array, error) {
	y0v, ok := y0.value.(array)
	if ok && y0v.Len() == 0 && !depth {
		return ctx.deepAmend4rec(x, rangeI(int64(x.Len())), y, f, z, depth)
	}
	if y.Len() == 0 {
		return ctx.amend4array(x, y0, f, z)
	}
	if y0.IsI() {
		if outOfBounds(y0.I(), x.Len()) {
			return x, fmt.Errorf("y out of bounds (%d)", y0.I())
		}
		xy0 := x.at(int(y0.I()))
		xy0v, ok := xy0.value.(array)
		if !ok {
			return x, errors.New("y out of depth")
		}
		repl, err := ctx.deepAmend4rec(xy0v, y.at(0), y.slice(1, y.Len()), f, z, false)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	var err error
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.at(i)
		x, err = ctx.deepAmend4rec(x, y0i, y, f, z, true)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// try implements .[f1;x;f2].
func try(ctx *Context, f1, x, f2 V) V {
	av := toArray(x).value.(array)
	if av.Len() == 0 {
		return panics(".[f1;x;f2] : empty x")
	}
	for i := av.Len() - 1; i >= 0; i-- {
		ctx.push(av.at(i))
	}
	r := f1.applyN(ctx, av.Len())
	if r.IsPanic() {
		r = NewS(string(r.value.(panicV)))
		ctx.replaceTop(r)
		r = f2.applyN(ctx, 1)
		if r.IsPanic() {
			ctx.drop()
			return Panicf(".[f1;x;f2] : f2 call: %v", r)
		}
	}
	ctx.drop()
	return r
}
