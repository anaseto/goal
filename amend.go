package goal

import (
	"errors"
	"fmt"
)

// amend3 implements @[X;i;f].
func (ctx *Context) amend3(x, y, f V) V {
	switch xv := x.bv.(type) {
	case *D:
		return amend3Dict(ctx, xv, y, f)
	case Array:
		xv = xv.sclone()
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[X;i;f] : ", y)
		}
		r, err := ctx.amend3array(xv, y, f)
		if err != nil {
			return Panicf("@[X;i;f] : %v", err)
		}
		return NewV(canonicalArray(r))
	default:
		return panicType("@[X;i;f]", "X", x)
	}
}

func (ctx *Context) amend3array(x Array, y, f V) (Array, error) {
	if f.kind == valVariadic && x.numeric() {
		switch f.variadic() {
		case vMatch:
			return amend3NotV(x, y)
		case vSubtract:
			return amend3NegateV(x, y)
		}
	}
	return ctx.amend3arrayGeneric(x, y, f)
}

func amend3Dict(ctx *Context, d *D, y, f V) V {
	switch yv := y.bv.(type) {
	case Array:
		keys, values, ky := dictAmendKVI(d, yv)
		r, err := ctx.amend3array(values, ky, f)
		if err != nil {
			return Panicf("@[d;y;f] : %v", err)
		}
		return NewV(&D{keys: keys, values: canonicalArray(r)})
	default:
		keys, values := d.keys, d.values.sclone()
		ky := findArray(keys, y)
		if ky.I() == int64(keys.Len()) {
			keys = join(NewV(keys), y).bv.(Array)
			values = padArrayMut(1, values)
		}
		r, err := ctx.amend3arrayI(values, ky.I(), f)
		if err != nil {
			return Panicf("@[d;y;f] : %v", err)
		}
		return NewV(&D{keys: keys, values: canonicalArray(r)})
	}
}

func outOfBounds(y int64, l int) bool {
	return y < 0 || y >= int64(l)
}

func padArrayMut(n int, x Array) Array {
	switch xv := x.(type) {
	case *AB:
		for i := 0; i < n; i++ {
			xv.elts = append(xv.elts, 0)
		}
	case *AI:
		for i := 0; i < n; i++ {
			xv.elts = append(xv.elts, 0)
		}
	case *AF:
		for i := 0; i < n; i++ {
			xv.elts = append(xv.elts, 0)
		}
	case *AS:
		for i := 0; i < n; i++ {
			xv.elts = append(xv.elts, "")
		}
	case *AV:
		pad := proto(xv.elts)
		pad.MarkImmutable()
		for i := 0; i < n; i++ {
			xv.elts = append(xv.elts, pad)
		}
	}
	return x
}

func amendArrayAt(x Array, y int, z V) Array {
	if x.canSet(z) {
		x.set(y, z)
		return x
	}
	a := make([]V, x.Len())
	for i := range a {
		a[i] = x.VAt(i)
	}
	z.MarkImmutable()
	a[y] = z
	return &AV{elts: a}
}

func (ctx *Context) amend3arrayI(x Array, y int64, f V) (Array, error) {
	if outOfBounds(y, x.Len()) {
		return x, fmt.Errorf("y out of bounds (%d)", y)
	}
	xy := x.VAt(int(y))
	repl := ctx.Apply(f, xy)
	if repl.IsPanic() {
		return x, fmt.Errorf("f call %v", repl.bv.(panicV))
	}
	return amendArrayAt(x, int(y), repl), nil
}

func (ctx *Context) amend3arrayGeneric(x Array, y, f V) (Array, error) {
	if y.IsI() {
		return ctx.amend3arrayI(x, y.I(), f)
	}
	if isStar(y) {
		return ctx.amend3array(x, enumI(int64(x.Len())), f)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return amend3arrayAtIs(ctx, x, yv.elts, f)
	case *AI:
		return amend3arrayAtIs(ctx, x, yv.elts, f)
	case *AV:
		var err error
		for _, yi := range yv.elts {
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

func amend3arrayAtIs[I integer](ctx *Context, x Array, y []I, f V) (Array, error) {
	var err error
	for _, yi := range y {
		x, err = ctx.amend3arrayI(x, int64(yi), f)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// amend4 implements @[X;i;f;z].
func (ctx *Context) amend4(x, y, f, z V) V {
	switch xv := x.bv.(type) {
	case *D:
		return amend4Dict(ctx, xv, y, f, z)
	case Array:
		xv = xv.sclone()
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic("@[X;i;f;z] : ", y)
		}
		r, err := ctx.amend4array(xv, y, f, z)
		if err != nil {
			return Panicf("@[X;i;f;z] : %v", err)
		}
		return NewV(canonicalArray(r))
	default:
		return panicType("@[X;i;f;z]", "X", x)
	}
}

func (ctx *Context) amend4array(x Array, y, f, z V) (Array, error) {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vRight:
			return amend4Right(x, y, z)
		case vAdd:
			return amend4Arith(x, y, add, z)
		case vSubtract:
			return amend4Arith(x, y, subtract, z)
		case vMultiply:
			return amend4Arith(x, y, multiply, z)
		case vDivide:
			return amend4Arith(x, y, divide, z)
		case vMax:
			return amend4Arith(x, y, maximum, z)
		case vMin:
			return amend4Arith(x, y, minimum, z)
		}
	}
	return ctx.amend4arrayGeneric(x, y, f, z)
}

func amend4Dict(ctx *Context, d *D, y, f, z V) V {
	switch yv := y.bv.(type) {
	case Array:
		keys, values, ky := dictAmendKVI(d, yv)
		r, err := ctx.amend4array(values, ky, f, z)
		if err != nil {
			return Panicf("@[d;y;f;z] : %v", err)
		}
		return NewV(&D{keys: keys, values: canonicalArray(r)})
	default:
		keys, values := d.keys, d.values.sclone()
		ky := findArray(keys, y)
		if ky.I() == int64(keys.Len()) {
			keys = join(NewV(keys), y).bv.(Array)
			values = padArrayMut(1, values)
		}
		if f.kind == valVariadic && variadic(f.uv) == vRight {
			r, err := amend4Right(values, ky, z)
			if err != nil {
				// never happens because of key padding
				return Panicf("@[d;y;f;z] : %v", err)
			}
			return NewV(&D{keys: keys, values: canonicalArray(r)})
		}
		r, err := ctx.amend4arrayI(values, ky.I(), f, z)
		if err != nil {
			return Panicf("@[d;y;f;z] : %v", err)
		}
		return NewV(&D{keys: keys, values: canonicalArray(r)})
	}
}

func (ctx *Context) amend4arrayI(x Array, y int64, f, z V) (Array, error) {
	if y < 0 || y >= int64(x.Len()) {
		return x, fmt.Errorf("y out of bounds (%d)", y)
	}
	xy := x.VAt(int(y))
	repl := ctx.Apply2(f, xy, z)
	if repl.IsPanic() {
		return x, fmt.Errorf("f call %v", repl.bv.(panicV))
	}
	return amendArrayAt(x, int(y), repl), nil
}

func (ctx *Context) amend4arrayGeneric(x Array, y, f, z V) (Array, error) {
	if y.IsI() {
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	if isStar(y) {
		return ctx.amend4array(x, enumI(int64(x.Len())), f, z)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return amend4arrayAtIs(ctx, x, yv.elts, f, z)
	case *AI:
		return amend4arrayAtIs(ctx, x, yv.elts, f, z)
	case *AV:
		var err error
		za, ok := z.bv.(Array)
		if !ok {
			for _, yi := range yv.elts {
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
		for i, yi := range yv.elts {
			x, err = ctx.amend4array(x, yi, f, za.VAt(i))
			if err != nil {
				return x, err
			}
		}
		return x, nil
	default:
		panic("amend4array: y bad type")
	}
}

func amend4arrayAtIs[I integer](ctx *Context, x Array, y []I, f, z V) (Array, error) {
	var err error
	za, ok := z.bv.(Array)
	if !ok {
		for _, yi := range y {
			x, err = ctx.amend4arrayI(x, int64(yi), f, z)
			if err != nil {
				return x, err
			}
		}
		return x, nil
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())

	}
	for i, yi := range y {
		x, err = ctx.amend4arrayI(x, int64(yi), f, za.VAt(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// deepAmend3 implements .[X;y;f].
func (ctx *Context) deepAmend3(x, y, f V) V {
	x = x.Clone()
	switch xv := x.bv.(type) {
	case Array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic(".[X;y;f] : ", y)
		}
		x, err := ctx.deepAmend3array(xv, y, f)
		if err != nil {
			return Panicf(".[X;y;f] : %v", err)
		}
		return canonicalRec(NewV(x))
	default:
		return panicType(".[X;y;f]", "x", x)
	}
}

func (ctx *Context) deepAmend3array(x Array, y, f V) (Array, error) {
	if y.IsI() {
		return ctx.amend3arrayI(x, y.I(), f)
	}
	if isStar(y) {
		return ctx.amend3array(x, enumI(int64(x.Len())), f)
	}
	yv := y.bv.(Array)
	if yv.Len() == 0 {
		return ctx.amend3array(x, enumI(int64(x.Len())), f)
	}
	return ctx.deepAmend3rec(x, yv.VAt(0), yv.slice(1, yv.Len()), f)
}

func (ctx *Context) deepAmend3rec(x Array, y0 V, y Array, f V) (Array, error) {
	var err error
	if isStar(y0) {
		return ctx.deepAmend3rec(x, enumI(int64(x.Len())), y, f)
	}
	if y.Len() == 0 {
		return ctx.amend3array(x, y0, f)
	}
	if y0.IsI() {
		if outOfBounds(y0.I(), x.Len()) {
			return x, fmt.Errorf("y out of bounds (%d)", y0.I())
		}
		xy0 := x.VAt(int(y0.I()))
		xy0v, ok := xy0.bv.(Array)
		if !ok {
			return x, errors.New("y out of depth")
		}
		repl, err := ctx.deepAmend3rec(xy0v, y.VAt(0), y.slice(1, y.Len()), f)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	y0v := y0.bv.(Array)
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.VAt(i)
		x, err = ctx.deepAmend3rec(x, y0i, y, f)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

// deepAmend4 implements .[X;y;f;z].
func (ctx *Context) deepAmend4(x, y, f, z V) V {
	x = x.Clone()
	switch xv := x.bv.(type) {
	case Array:
		y = toIndices(y)
		if y.IsPanic() {
			return ppanic(".[X;y;f] : ", y)
		}
		x, err := ctx.deepAmend4array(xv, y, f, z)
		if err != nil {
			return Panicf(".[X;y;f] : %v", err)
		}
		return canonicalRec(NewV(x))
	default:
		return panicType(".[X;y;f]", "x", x)
	}
}

func (ctx *Context) deepAmend4array(x Array, y, f, z V) (Array, error) {
	if y.IsI() {
		return ctx.amend4arrayI(x, y.I(), f, z)
	}
	if isStar(y) {
		return ctx.amend4array(x, enumI(int64(x.Len())), f, z)
	}
	yv := y.bv.(Array)
	if yv.Len() == 0 {
		return ctx.amend4array(x, enumI(int64(x.Len())), f, z)
	}
	return ctx.deepAmend4rec(x, yv.VAt(0), yv.slice(1, yv.Len()), f, z)
}

func (ctx *Context) deepAmend4rec(x Array, y0 V, y Array, f, z V) (Array, error) {
	var err error
	if isStar(y0) {
		return ctx.deepAmend4rec(x, enumI(int64(x.Len())), y, f, z)
	}
	if y.Len() == 0 {
		return ctx.amend4array(x, y0, f, z)
	}
	if y0.IsI() {
		if outOfBounds(y0.I(), x.Len()) {
			return x, fmt.Errorf("y out of bounds (%d)", y0.I())
		}
		xy0 := x.VAt(int(y0.I()))
		xy0v, ok := xy0.bv.(Array)
		if !ok {
			return x, errors.New("y out of depth")
		}
		repl, err := ctx.deepAmend4rec(xy0v, y.VAt(0), y.slice(1, y.Len()), f, z)
		if err != nil {
			return x, err
		}
		return amendArrayAt(x, int(y0.I()), NewV(repl)), nil
	}
	y0v := y0.bv.(Array)
	for i := 0; i < y0v.Len(); i++ {
		y0i := y0v.VAt(i)
		x, err = ctx.deepAmend4rec(x, y0i, y, f, z)
		if err != nil {
			return x, err
		}
	}
	return x, nil
}
