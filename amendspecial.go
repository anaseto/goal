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
		return inBoundsInts(yv.elts, l)
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

func inBoundsInts(y []int64, l int) (int64, bool) {
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

func arithmAmend3AV(x array, y *AV, f func(array, V) (array, error)) (array, error) {
	var err error
	for _, yi := range y.elts {
		x, err = f(x, yi)
		if err != nil {
			return x, err
		}
	}
	return x, err
}

func amend3NotV(x array, y V) (array, error) {
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
		return amend3NotIntegers(x, yv.elts), nil
	case *AI:
		return amend3NotIntegers(x, yv.elts), nil
	case *AV:
		return arithmAmend3AV(x, yv, amend3NotV)
	default:
		panic("amend3NotV")
	}
}

func amend3NotI(x array, y int64) array {
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

func amend3NotIntegers[I integer](x array, y []I) array {
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
		panic("amend3NotIntegers")
	}
}

func amend3NegateV(x array, y V) (array, error) {
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
		return amend3NegateIntegers(x, yv.elts), nil
	case *AI:
		return amend3NegateIntegers(x, yv.elts), nil
	case *AV:
		return arithmAmend3AV(x, yv, amend3NegateV)
	default:
		panic("amend3NegateV")
	}
}

func amend3NegateI(x array, y int64) array {
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

func amend3NegateIntegers[I integer](x array, y []I) array {
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
		panic("amend3NegateIntegers")
	}
}

func amend4Right(x array, y, z V) (array, error) {
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

func amend4RightI(x array, y int64, z V) (array, error) {
	if outOfBounds(y, x.Len()) {
		return x, fmt.Errorf("out of bounds index (%d)", y)
	}
	if x.canSet(z) {
		x.set(int(y), z)
		return x, nil
	}
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.at(i)
	}
	z.MarkImmutable()
	r[y] = z
	return &AV{elts: r}, nil
}

func amend4RightIs[I integer](x array, y []I, z V) (array, error) {
	xlen := x.Len()
	for _, yi := range y {
		if outOfBounds(int64(yi), xlen) {
			return x, fmt.Errorf("out of bounds index (%d)", yi)
		}
	}
	za, ok := z.bv.(array)
	if !ok {
		return amend4RightIsV(x, y, z)
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())
	}
	if sameType(x, za) {
		return amend4RightIntegersSlice(x, y, za), nil
	}
	return amend4RightIntegersArrays(x, y, za), nil
}

func amend4RightIsV[I integer](x array, y []I, z V) (array, error) {
	if x.canSet(z) {
		// NOTE: we could optimize some float and integer mix cases.
		amend4RightIntegersAtom(x, y, z)
		return x, nil
	}
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.at(i)
	}
	z.MarkImmutable()
	for _, yi := range y {
		r[yi] = z
	}
	return &AV{elts: r}, nil
}

func amend4RightIntegersAtom[I integer](x array, y []I, z V) {
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
		z.immutable()
		amendSlice(xv.elts, y, z)
	}
}

func amendSlice[I integer, T any](x []T, y []I, z T) {
	for _, yi := range y {
		x[yi] = z
	}
}

func amend4RightIntegersSlice[I integer](x array, y []I, za array) array {
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

func amend4RightIntegersArrays[I integer](x array, y []I, z array) array {
	for i := range y {
		if !x.canSet(z.at(i)) {
			r := make([]V, x.Len())
			for i := range r {
				r[i] = x.at(i)
			}
			x = &AV{elts: r}
			break
		}
	}
	for i, yi := range y {
		x.set(int(yi), z.at(i))
	}
	return x
}

func amend4RightAV(x array, yv *AV, z V) (array, error) {
	var err error
	za, ok := z.bv.(array)
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
		x, err = amend4Right(x, yi, za.at(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

func amend4Arith(x array, y V, f func(V, V) V, z V) (array, error) {
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
		return arithAmendIntegersV(x, yv.elts, f, z)
	case *AI:
		return arithAmendIntegersV(x, yv.elts, f, z)
	case *AV:
		return arithAmend4AV(x, yv.elts, f, z)
	default:
		panic("amend4Arith")
	}
}

func arithAmendI(x array, y int, f func(V, V) V, z V) (array, error) {
	xy := x.at(y)
	repl := f(xy, z)
	if repl.IsPanic() {
		return x, newExecError(repl)
	}
	return amendArrayAt(x, y, repl), nil
}

func arithAmendIntegersV[I integer](x array, y []I, f func(V, V) V, z V) (array, error) {
	za, ok := z.bv.(array)
	if !ok {
		return arithAmendIntegersAtom(x, y, f, z)
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())
	}
	return arithAmendIntegersArray(x, y, f, za)
}

func arithAmendIntegersArray[I integer](x array, y []I, f func(V, V) V, z array) (array, error) {
	var err error
	for i, yi := range y {
		x, err = arithAmendI(x, int(yi), f, z.at(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}

func arithAmendIntegersAtom[I integer](x array, y []I, f func(V, V) V, z V) (array, error) {
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

func arithAmend4AV(x array, y []V, f func(V, V) V, z V) (array, error) {
	var err error
	za, ok := z.bv.(array)
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
		x, err = amend4Arith(x, yi, f, za.at(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}
