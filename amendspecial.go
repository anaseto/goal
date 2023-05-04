package goal

import (
	"fmt"
)

// inBoundsV returns true if it contains only indexes within [0,l), and false
// otherwise, along with the out index. NOTE: we do not handle negative indices
// in amend for now.
func inBoundsV(y V, l int) (int64, bool) {
	if y.IsI() {
		return inBoundsI(y.I(), l)
	}
	if isStar(y) {
		return 0, true
	}
	switch yv := y.value.(type) {
	case *AB:
		return inBoundsBytes(yv.elts, l)
	case *AI:
		return inBoundsInts(yv.elts, l)
	case *AV:
		for _, yi := range yv.elts {
			i, ok := inBoundsV(yi, l)
			if !ok {
				return i, false
			}
		}
		return 0, true
	default:
		panic("inBoundsV")
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

func amend3NotV(x array, y V) (array, error) {
	yi, ok := inBoundsV(y, x.Len())
	if !ok {
		return x, fmt.Errorf("out of bounds index (%d)", yi)
	}
	if y.IsI() {
		return amend3NotI(x, y.I()), nil
	}
	switch yv := y.value.(type) {
	case *AB:
		return amend3NotIntegers(x, yv.elts), nil
	case *AI:
		return amend3NotIntegers(x, yv.elts), nil
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
	switch yv := y.value.(type) {
	case *AB:
		return amend3NegateIntegers(x, yv.elts), nil
	case *AI:
		return amend3NegateIntegers(x, yv.elts), nil
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
		return &AI{elts: r, rc: x.RC()}
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
		return &AI{elts: r, rc: x.RC()}
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
		return amend4RightIV(x, y.I(), z)
	}
	switch yv := y.value.(type) {
	case *AB:
		return amend4RightIntegers(x, yv.elts, z)
	case *AI:
		return amend4RightIntegers(x, yv.elts, z)
	default:
		panic("amend4Right")
	}
}

func amend4RightIV(x array, y int64, z V) (array, error) {
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
	rc := x.RC()
	z.InitWithRC(rc)
	r[y] = z
	return &AV{elts: r, rc: rc}, nil
}

func amend4RightIntegers[I integer](x array, y []I, z V) (array, error) {
	xlen := x.Len()
	for _, yi := range y {
		if outOfBounds(int64(yi), xlen) {
			return x, fmt.Errorf("out of bounds index (%d)", yi)
		}
	}
	za, ok := z.value.(array)
	if !ok {
		return amend4RightIsV(x, y, z)
	}
	if za.Len() != len(y) {
		return x, fmt.Errorf("length mismatch between y and z (%d vs %d)",
			len(y), za.Len())
	}
	if sameType(x, za) {
		amend4RightIntegersSlice(x, y, za)
		return x, nil
	}
	return amend4RightIntegersArrays(x, y, za)
}

func amend4RightIsV[I integer](x array, y []I, z V) (array, error) {
	if x.canSet(z) {
		amend4RightIntegersAtom(x, y, z)
		return x, nil
	}
	r := make([]V, x.Len())
	for i := range r {
		r[i] = x.at(i)
	}
	rc := x.RC()
	z.immutable()
	for _, yi := range y {
		r[yi] = z
	}
	return &AV{elts: r, rc: rc}, nil
}

func amend4RightIntegersAtom[I integer](x array, y []I, z V) {
	switch xv := x.(type) {
	case *AB:
		var zi byte
		if z.IsI() {
			zi = byte(z.I())
		} else {
			zi = byte(z.F())
		}
		amendSlice(xv.elts, y, zi)
	case *AI:
		var zi int64
		if z.IsI() {
			zi = z.I()
		} else {
			zi = int64(z.F())
		}
		amendSlice(xv.elts, y, zi)
	case *AF:
		var zf float64
		if z.IsI() {
			zf = float64(z.I())
		} else {
			zf = z.F()
		}
		amendSlice(xv.elts, y, zf)
	}
}

func amendSlice[I integer, T any](x []T, y []I, z T) {
	for _, yi := range y {
		x[yi] = z
	}
}

func amend4RightIntegersSlice[I integer](x array, y []I, za array) {
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
	}
}

func amend4RightSlices[I integer, T any](x []T, y []I, z []T) {
	for i, yi := range y {
		x[yi] = z[i]
	}
}

func amend4RightIntegersArrays[I integer](x array, y []I, z array) (array, error) {
	for i := range y {
		if !x.canSet(z.at(i)) {
			r := make([]V, x.Len())
			for i := range r {
				r[i] = x.at(i)
			}
			x = &AV{elts: r, rc: x.RC()}
			break
		}
	}
	for i, yi := range y {
		x.set(int(yi), z.at(i))
	}
	return x, nil
}
