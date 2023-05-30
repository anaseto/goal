package goal

import "fmt"

// Dict represents a dictionary.
type Dict struct {
	keys   array
	values array
}

// NewDict returns a dictionary. Both keys and values should be arrays, and
// they should have the same length.
func NewDict(keys, values V) V {
	xv, ok := keys.bv.(array)
	if !ok {
		panic(fmt.Sprintf("NewDict(keys, values) : keys not an array (%s)", keys.Type()))
	}
	yv, ok := values.bv.(array)
	if !ok {
		panic(fmt.Sprintf("NewDict(keys, values) : values not an array (%s)", values.Type()))
	}
	if xv.Len() != yv.Len() {
		panic(fmt.Sprintf("NewDict(keys, values) : length mismatch (%d vs %d)", xv.Len(), yv.Len()))
	}
	return NewV(&Dict{keys: xv, values: yv})
}

func newDictValues(keys array, values V) V {
	if values.IsPanic() {
		return values
	}
	v := values.bv.(array)
	return NewV(&Dict{keys: keys, values: v})
}

// Keys returns the keys of the dictionary.
func (d *Dict) Keys() V {
	return NewV(d.keys)
}

// Values returns the values of the dictionary.
func (d *Dict) Values() V {
	return NewV(d.values)
}

// Matches returns true if the two values match like in x~y.
func (d *Dict) Matches(y Value) bool {
	switch yv := y.(type) {
	case *Dict:
		return d.keys.Matches(yv.keys) && d.values.Matches(yv.values)
	default:
		return false
	}
}

// Type returns the name of the value's type.
func (d *Dict) Type() string {
	return "d"
}

// Len returns the length of the dictionary, that is the common length to its
// key and value arrays.
func (d *Dict) Len() int {
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
	return NewV(&Dict{keys: xv, values: yv})
}

func dictAmendKVI(xd *Dict, yk array) (array, array, V) {
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

func dictMerge(xd, yd *Dict) V {
	keys, values, ky := dictAmendKVI(xd, yd.keys)
	var r array
	switch kyv := ky.bv.(type) {
	case *AB:
		r = mergeAtIntegers(values, kyv.elts, yd.values)
	case *AI:
		r = mergeAtIntegers(values, kyv.elts, yd.values)
	}
	return NewV(&Dict{keys: keys, values: canonicalArray(r)})
}

func mergeAtIntegers[I integer](x array, y []I, z array) array {
	if sameType(x, z) {
		return amend4RightIntegersSlice(x, y, z)
	}
	return amend4RightIntegersArrays(x, y, z)
}

func dictArith(xd, yd *Dict, f func(V, V) V) V {
	keys, values, ky := dictAmendKVI(xd, yd.keys)
	var err error
	var r array
	switch kyv := ky.bv.(type) {
	case *AB:
		r, err = arithAmendIntegersArray(values, kyv.elts, f, yd.values)
	case *AI:
		r, err = arithAmendIntegersArray(values, kyv.elts, f, yd.values)
	}
	if err != nil {
		return Panicf("%v", err)
	}
	return NewV(&Dict{keys: keys, values: canonicalArray(r)})
}
