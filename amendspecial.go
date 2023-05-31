package goal

import (
	"fmt"
)

// inBoundsV returns true if it contains only indexes within [0,l), and false
// otherwise, along with the out index. NOTE: We do not handle negative indices
// in amend for now. Also, it doesn't do recursion in generic arrays.
func inBoundsV(y V, l int) (int64, bool) {
	if y.IsI() {
		return inBoundsI(y.I(), l)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return inBoundsBytes(yv.elts, l)
	case *AI:
		return inBoundsInt64s(yv.elts, l)
	default:
		return 0, true
	}
}

func inBoundsBytes(y []byte, l int) (int64, bool) {
	for _, yi := range y {
		if int(yi) >= l {
			return int64(yi), false
		}
	}
	return 0, true
}

func inBoundsInt64s(y []int64, l int) (int64, bool) {
	for _, yi := range y {
		if outOfBounds(yi, l) {
			return yi, false
		}
	}
	return 0, true
}

func inBoundsI(y int64, l int) (int64, bool) {
	if outOfBounds(y, l) {
		return y, false
	}
	return 0, true
}

func arithmAmend3AV(x Array, y *AV, f func(Array, V) (Array, error)) (Array, error) {
	var err error
	for _, yi := range y.elts {
		x, err = f(x, yi)
		if err != nil {
			return x, err
		}
	}
	return x, err
}

func amend3NotV(x Array, y V) (Array, error) {
	yi, ok := inBoundsV(y, x.Len())
	if !ok {
		return x, fmt.Errorf("out of bounds index (%d)", yi)
	}
	if y.IsI() {
		return amend3NotI(x, y.I()), nil
	}
	if isStar(y) {
		return amend3NotV(x, enumI(int64(x.Len())))
	}
	switch yv := y.bv.(type) {
	case *AB:
		return amend3NotIs(x, yv.elts), nil
	case *AI:
		return amend3NotIs(x, yv.elts), nil
	case *AV:
		return arithmAmend3AV(x, yv, amend3NotV)
	default:
		panic("amend3NotV")
	}
}

func amend3NotI(x Array, y int64) Array {
	switch xv := x.(type) {
	case *AB:
		xv.elts[y] = b2B(xv.elts[y] == 0)
		return x
	case *AI:
		xv.elts[y] = b2I(xv.elts[y] == 0)
		return x
	case *AF:
		xv.elts[y] = b2F(xv.elts[y] == 0)
		return x
	default:
		panic("amend3NotI")
	}
}

func amend3NotIs[I integer](x Array, y []I) Array {
	switch xv := x.(type) {
	case *AB:
		for _, yi := range y {
			xv.elts[yi] = b2B(xv.elts[yi] == 0)
		}
		return x
	case *AI:
		for _, yi := range y {
			xv.elts[yi] = b2I(xv.elts[yi] == 0)
		}
		return x
	case *AF:
		for _, yi := range y {
			xv.elts[yi] = b2F(xv.elts[yi] == 0)
		}
		return x
	default:
		panic("amend3NotIs")
	}
}

func amend3NegateV(x Array, y V) (Array, error) {
	yi, ok := inBoundsV(y, x.Len())
	if !ok {
		return x, fmt.Errorf("out of bounds index (%d)", yi)
	}
	if y.IsI() {
		return amend3NegateI(x, y.I()), nil
	}
	if isStar(y) {
		return amend3NegateV(x, enumI(int64(x.Len())))
	}
	switch yv := y.bv.(type) {
	case *AB:
		return amend3NegateIs(x, yv.elts), nil
	case *AI:
		return amend3NegateIs(x, yv.elts), nil
	case *AV:
		return arithmAmend3AV(x, yv, amend3NegateV)
	default:
		panic("amend3NegateV")
	}
}

func amend3NegateI(x Array, y int64) Array {
	switch xv := x.(type) {
	case *AB:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
			r[i] = int64(xi)
		}
		r[y] = -r[y]
		return &AI{elts: r}
	case *AI:
		xv.elts[y] = -xv.elts[y]
		return x
	case *AF:
		xv.elts[y] = -xv.elts[y]
		return x
	default:
		panic("amend3NegateI")
	}
}

func amend3NegateIs[I integer](x Array, y []I) Array {
	switch xv := x.(type) {
	case *AB:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
			r[i] = int64(xi)
		}
		for _, yi := range y {
			r[yi] = -r[yi]
		}
		return &AI{elts: r}
	case *AI:
		for _, yi := range y {
			xv.elts[yi] = -xv.elts[yi]
		}
		return x
	case *AF:
		for _, yi := range y {
			xv.elts[yi] = -xv.elts[yi]
		}
		return x
	default:
		panic("amend3NegateIs")
	}
}

func amend4Right(x Array, y, z V) (Array, error) {
	if y.IsI() {
		return amend4RightI(x, y.I(), z)
	}
	if isStar(y) {
		return amend4Right(x, enumI(int64(x.Len())), z)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return amend4RightIs(x, yv.elts, z)
	case *AI:
		return amend4RightIs(x, yv.elts, z)
	case *AV:
		return amend4RightAV(x, yv, z)
	default:
		panic("amend4Right")
	}
}

func amend4RightI(x Array, y int64, z V) (Array, error) {
	if outOfBounds(y, x.Len()) {
		return x, fmt.Errorf("out of bounds index (%d)", y)
	}
	if x.canSet(z) {
		x.set(int(y), z)
		return x, nil
	}
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.VAt(i)
	}
	z.MarkImmutable()
	r[y] = z
	return &AV{elts: r}, nil
}

func amend4RightIs[I integer](x Array, y []I, z V) (Array, error) {
	xlen := x.Len()
	for _, yi := range y {
		if outOfBounds(int64(yi), xlen) {
			return x, fmt.Errorf("out of bounds index (%d)", yi)
		}
	}
	za, ok := z.bv.(Array)
	if !ok {
		return amend4RightIsV(x, y, z)
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())
	}
	if sameType(x, za) {
		return amend4RightIsATs(x, y, za), nil
	}
	return amend4RightIsArrays(x, y, za), nil
}

func amend4RightIsV[I integer](x Array, y []I, z V) (Array, error) {
	if x.canSet(z) {
		// NOTE: we could optimize some float and integer mix cases.
		amend4RightIsAtom(x, y, z)
		return x, nil
	}
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.VAt(i)
	}
	z.MarkImmutable()
	for _, yi := range y {
		r[yi] = z
	}
	return &AV{elts: r}, nil
}

func amend4RightIsAtom[I integer](x Array, y []I, z V) {
	switch xv := x.(type) {
	case *AB:
		amendSlice(xv.elts, y, byte(z.I()))
	case *AI:
		amendSlice(xv.elts, y, z.I())
	case *AF:
		amendSlice(xv.elts, y, z.F())
	case *AS:
		zs := string(z.bv.(S))
		amendSlice(xv.elts, y, zs)
	case *AV:
		z.MarkImmutable()
		amendSlice(xv.elts, y, z)
	}
}

func amendSlice[I integer, T any](x []T, y []I, z T) {
	for _, yi := range y {
		x[yi] = z
	}
}

func amend4RightIsATs[I integer](x Array, y []I, za Array) Array {
	switch xv := x.(type) {
	case *AB:
		zv := za.(*AB)
		amend4RightSlices(xv.elts, y, zv.elts)
	case *AI:
		zv := za.(*AI)
		amend4RightSlices(xv.elts, y, zv.elts)
	case *AF:
		zv := za.(*AF)
		amend4RightSlices(xv.elts, y, zv.elts)
	case *AS:
		zv := za.(*AS)
		amend4RightSlices(xv.elts, y, zv.elts)
	case *AV:
		zv := za.(*AV)
		amend4RightSlices(xv.elts, y, zv.elts)
	}
	return x
}

func amend4RightSlices[I integer, T any](x []T, y []I, z []T) {
	for i, yi := range y {
		x[yi] = z[i]
	}
}

func amend4RightIsArrays[I integer](x Array, y []I, z Array) Array {
	for i := range y {
		if !x.canSet(z.VAt(i)) {
			r := make([]V, x.Len())
			for i := range r {
				r[i] = x.VAt(i)
			}
			x = &AV{elts: r}
			break
		}
	}
	for i, yi := range y {
		x.set(int(yi), z.VAt(i))
	}
	return x
}

func amend4RightAV(x Array, yv *AV, z V) (Array, error) {
	var err error
	za, ok := z.bv.(Array)
	if !ok {
		for _, yi := range yv.elts {
			x, err = amend4Right(x, yi, z)
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
		x, err = amend4Right(x, yi, za.VAt(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

func amend4Arith(x Array, y V, f func(V, V) V, z V) (Array, error) {
	yi, ok := inBoundsV(y, x.Len())
	if !ok {
		return x, fmt.Errorf("out of bounds index (%d)", yi)
	}
	if y.IsI() {
		return arithAmendI(x, int(y.I()), f, z)
	}
	if isStar(y) {
		return amend4Arith(x, enumI(int64(x.Len())), f, z)
	}
	switch yv := y.bv.(type) {
	case *AB:
		return arithAmendIsV(x, yv.elts, f, z)
	case *AI:
		return arithAmendIsV(x, yv.elts, f, z)
	case *AV:
		return arithAmend4AV(x, yv.elts, f, z)
	default:
		panic("amend4Arith")
	}
}

func arithAmendI(x Array, y int, f func(V, V) V, z V) (Array, error) {
	xy := x.VAt(y)
	repl := f(xy, z)
	if repl.IsPanic() {
		return x, newExecError(repl)
	}
	return amendArrayAt(x, y, repl), nil
}

func arithAmendIsV[I integer](x Array, y []I, f func(V, V) V, z V) (Array, error) {
	za, ok := z.bv.(Array)
	if !ok {
		return arithAmendIsAtom(x, y, f, z)
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())
	}
	return arithAmendIsArray(x, y, f, za)
}

func arithAmendIsArray[I integer](x Array, y []I, f func(V, V) V, z Array) (Array, error) {
	var err error
	for i, yi := range y {
		x, err = arithAmendI(x, int(yi), f, z.VAt(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

func arithAmendIsAtom[I integer](x Array, y []I, f func(V, V) V, z V) (Array, error) {
	var err error
	z.incrRC2()
	for _, yi := range y {
		x, err = arithAmendI(x, int(yi), f, z)
		if err != nil {
			z.decrRC2()
			return x, err
		}
	}
	z.decrRC2()
	return x, nil
}

func arithAmend4AV(x Array, y []V, f func(V, V) V, z V) (Array, error) {
	var err error
	za, ok := z.bv.(Array)
	if !ok {
		for _, yi := range y {
			x, err = amend4Arith(x, yi, f, z)
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
		x, err = amend4Arith(x, yi, f, za.VAt(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}
