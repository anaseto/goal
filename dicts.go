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
	xv, ok := keys.value.(array)
	if !ok {
		panic(fmt.Sprintf("NewDict(keys, values) : keys not an array (%s)", keys.Type()))
	}
	yv, ok := values.value.(array)
	if !ok {
		panic(fmt.Sprintf("NewDict(keys, values) : values not an array (%s)", values.Type()))
	}
	if xv.Len() != yv.Len() {
		panic(fmt.Sprintf("NewDict(keys, values) : length mismatch (%d vs %d)", xv.Len(), yv.Len()))
	}
	initRC(xv)
	initRC(yv)
	return NewV(&Dict{keys: xv, values: yv})
}

func newDictValues(keys array, values V) V {
	if values.IsPanic() {
		return values
	}
	v := values.value.(array)
	initRC(v)
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
	xv, ok := x.value.(array)
	if !ok {
		return panicType("X!Y", "X", x)
	}
	yv, ok := y.value.(array)
	if !ok {
		return panicType("X!Y", "X", x)
	}
	if xv.Len() != yv.Len() {
		return panicLength("X!Y", xv.Len(), yv.Len())
	}
	return NewV(&Dict{keys: xv, values: yv})
}

func dictArith(xd, yd *Dict, f func(V, V) V) V {
	xd = xd.clone()
	xk, xv := xd.keys, xd.values
	yk, yv := yd.keys, yd.values
	ky := findArray(xk, NewV(yk))
	nkeys := xk.Len()
	max := maxIndices(ky) // *AB or *AI -> I
	if max == int64(nkeys) {
		bnk := memberOf(NewV(yk), NewV(xk))
		bnk.InitRC()
		bnk.incrRC2()
		notbnk := not(bnk)
		bnk.decrRC2()
		nk := replicate(notbnk, NewV(yk))
		nv := replicate(notbnk, NewV(yv))
		yv = replicate(bnk, NewV(yv)).value.(array)
		xk = joinTo(NewV(xk), nk).value.(array)
		initRC(xk)
		xv = joinTo(NewV(xv), nv).value.(array)
		ky = without(NewAI([]int64{int64(nkeys)}), ky)
	}
	r, err := dictArithAmend(xv, ky.value.(array), f, yv)
	if err != nil {
		return Panicf("%v", err)
	}
	initRC(r)
	return NewV(&Dict{keys: xk, values: canonicalArray(r)})
}

func dictArithAmendI(x array, y int64, f func(V, V) V, z V) (array, error) {
	xy := x.at(int(y))
	repl := f(xy, z)
	if repl.IsPanic() {
		return x, newExecError(repl)
	}
	return amendArrayAt(x, int(y), repl), nil
}

func dictArithAmend(x array, y array, f func(V, V) V, z array) (array, error) {
	switch yv := y.(type) {
	case *AB:
		var err error
		for i, yi := range yv.elts {
			x, err = dictArithAmendI(x, int64(yi), f, z.at(i))
			if err != nil {
				return x, err
			}
		}
		return x, nil
	case *AI:
		var err error
		for i, yi := range yv.elts {
			x, err = dictArithAmendI(x, yi, f, z.at(i))
			if err != nil {
				return x, err
			}
		}
		return x, nil
	default:
		panic("dictArithAmend")
	}
}

// withKeys implements X#d.
func withKeys(x array, y *Dict) V {
	r := memberOf(NewV(y.keys), NewV(x))
	return NewDict(replicate(r, NewV(y.keys)), replicate(r, NewV(y.values)))
}
