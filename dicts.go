package goal

import "fmt"

// D represents a dictionary.
type D struct {
	keys   array
	values array
}

// NewD returns a dictionary. Both keys and values should be arrays, and
// they should have the same length.
func NewD(keys, values V) V {
	xv, ok := keys.bv.(array)
	if !ok {
		panic(fmt.Sprintf("NewD(keys, values) : keys not an array (%s)", keys.Type()))
	}
	yv, ok := values.bv.(array)
	if !ok {
		panic(fmt.Sprintf("NewD(keys, values) : values not an array (%s)", values.Type()))
	}
	if xv.Len() != yv.Len() {
		panic(fmt.Sprintf("NewD(keys, values) : length mismatch (%d vs %d)", xv.Len(), yv.Len()))
	}
	return NewV(&D{keys: xv, values: yv})
}

func newDictValues(keys array, values V) V {
	if values.IsPanic() {
		return values
	}
	v := values.bv.(array)
	return NewV(&D{keys: keys, values: v})
}

// Keys returns the keys of the dictionary.
func (d *D) Keys() V {
	return NewV(d.keys)
}

// Values returns the values of the dictionary.
func (d *D) Values() V {
	return NewV(d.values)
}

// Matches returns true if the two values match like in x~y.
func (d *D) Matches(y BV) bool {
	switch yv := y.(type) {
	case *D:
		return d.keys.Matches(yv.keys) && d.values.Matches(yv.values)
	default:
		return false
	}
}

// Type returns the name of the value's type.
func (d *D) Type() string {
	return "d"
}

// Len returns the length of the dictionary, that is the common length to its
// key and value arrays.
func (d *D) Len() int {
	return d.keys.Len()
}

func dict(x, y V) V {
	if x.IsI() {
		return moddivpad(x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i!y : non-integer i (%g)", x.F())
		}
		return moddivpad(int64(x.F()), y)
	}
	xv, ok := x.bv.(array)
	if !ok {
		return panicType("X!Y", "X", x)
	}
	yv, ok := y.bv.(array)
	if !ok {
		return panicType("X!Y", "X", x)
	}
	if xv.Len() != yv.Len() {
		return panicLength("X!Y", xv.Len(), yv.Len())
	}
	return NewV(&D{keys: xv, values: yv})
}

func dictAmendKVI(xd *D, yk array) (array, array, V) {
	keys, values := xd.keys, xd.values.sclone()
	ykv := NewV(yk)
	yk.IncrRC()
	ky := findArray(keys, ykv)
	nkeys := keys.Len()
	max := maxIndices(ky)
	if max == int64(nkeys) {
		b := equalIV(max, ky)
		flags := keys.getFlags() & flagDistinct
		keys = join(NewV(keys), distinct(replicate(b, ykv))).bv.(array)
		keys.setFlags(flags)
		values = padArrayMut(keys.Len()-nkeys, values)
		ky = findArray(keys, ykv)
	}
	yk.DecrRC()
	return keys, values, ky
}

func dictMerge(xd, yd *D) V {
	keys, values, ky := dictAmendKVI(xd, yd.keys)
	var r array
	switch kyv := ky.bv.(type) {
	case *AB:
		r = mergeAtIs(values, kyv.elts, yd.values)
	case *AI:
		r = mergeAtIs(values, kyv.elts, yd.values)
	}
	return NewV(&D{keys: keys, values: canonicalArray(r)})
}

func mergeAtIs[I integer](x array, y []I, z array) array {
	if sameType(x, z) {
		return amend4RightIsATs(x, y, z)
	}
	return amend4RightIsArrays(x, y, z)
}

func dictArith(xd, yd *D, f func(V, V) V) V {
	keys, values, ky := dictAmendKVI(xd, yd.keys)
	var err error
	var r array
	switch kyv := ky.bv.(type) {
	case *AB:
		r, err = arithAmendIsArray(values, kyv.elts, f, yd.values)
	case *AI:
		r, err = arithAmendIsArray(values, kyv.elts, f, yd.values)
	}
	if err != nil {
		return Panicf("%v", err)
	}
	return NewV(&D{keys: keys, values: canonicalArray(r)})
}
