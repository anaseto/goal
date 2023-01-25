package goal

import (
	"errors"
	"fmt"
)

// amend3 implements @[x;y;f].
func (ctx *Context) amend3(x, y, f V) V {
	switch xv := x.value.(type) {
	case array:
		xv = xv.shallowClone()
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[x;y;f] : y ", y)
		}
		r, err := ctx.amend3array(xv, y, f)
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
	rc := x.RC()
	z.InitWithRC(rc)
	a[y] = z
	return &AV{Slice: a, rc: rc}
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
	if isStar(y) {
		var err error
		for i := 0; i < x.Len(); i++ {
			x, err = ctx.amend3arrayI(x, int64(i), f)
			if err != nil {
				return x, err
			}
		}
		return x, nil
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
		xv = xv.shallowClone()
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[x;y;f;z] : y ", y)
		}
		if f.kind == valVariadic && variadic(f.n) == vRight {
			r, err := amendr(xv, y, z)
			if err != nil {
				return Panicf("@[x;y;:;z] : %v", err)
			}
			return Canonical(NewV(r))
		}
		r, err := ctx.amend4array(xv, y, f, z)
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
	if isStar(y) {
		y = rangeI(int64(x.Len()))
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
		rc := x.RC()
		z.InitWithRC(rc)
		r[y.I()] = z
		return &AV{Slice: r, rc: rc}, nil
	}
	if isStar(y) {
		y = rangeI(int64(x.Len()))
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
		rc := x.RC()
		z.InitWithRC(rc)
		for _, yi := range yv.Slice {
			r[yi] = z
		}
		return &AV{Slice: r, rc: rc}, nil
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
			x = &AV{Slice: r, rc: x.RC()}
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
		rc := x.RC()
		z.InitWithRC(rc)
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
		rc := x.RC()
		for i, yi := range yv.Slice {
			zi := zv.Slice[i]
			zi.InitWithRC(rc)
			xv.Slice[yi] = zi
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

// deepAmend3 implements .[x;y;f].
func (ctx *Context) deepAmend3(x, y, f V) V {
	x = x.Clone()
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
		return CanonicalRec(NewV(x))
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
		return ctx.amend3array(x, rangeI(int64(x.Len())), f)
	}
	return ctx.deepAmend3rec(x, yv.at(0), yv.slice(1, yv.Len()), f)
}

func (ctx *Context) deepAmend3rec(x array, y0 V, y array, f V) (array, error) {
	var err error
	if isStar(y0) {
		for i := 0; i < x.Len(); i++ {
			x, err = ctx.deepAmend3rec(x, NewI(int64(i)), y, f)
			if err != nil {
				return x, err
			}
		}
		return x, nil
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
		repl, err := ctx.deepAmend3rec(xy0v, y.at(0), y.slice(1, y.Len()), f)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	y0v := y0.value.(array)
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.at(i)
		x, err = ctx.deepAmend3rec(x, y0i, y, f)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// deepAmend4 implements .[x;y;f].
func (ctx *Context) deepAmend4(x, y, f, z V) V {
	x = x.Clone()
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
		return CanonicalRec(NewV(x))
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
		return ctx.amend4array(x, rangeI(int64(x.Len())), f, z)
	}
	return ctx.deepAmend4rec(x, yv.at(0), yv.slice(1, yv.Len()), f, z)
}

func (ctx *Context) deepAmend4rec(x array, y0 V, y array, f, z V) (array, error) {
	var err error
	if isStar(y0) {
		for i := 0; i < x.Len(); i++ {
			x, err = ctx.deepAmend4rec(x, NewI(int64(i)), y, f, z)
			if err != nil {
				return x, err
			}
		}
		return x, nil
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
		repl, err := ctx.deepAmend4rec(xy0v, y.at(0), y.slice(1, y.Len()), f, z)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	y0v := y0.value.(array)
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.at(i)
		x, err = ctx.deepAmend4rec(x, y0i, y, f, z)
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
