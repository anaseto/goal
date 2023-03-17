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

func (d *Dict) Matches(y Value) bool {
	switch yv := y.(type) {
	case *Dict:
		return d.keys.Matches(yv.keys) && d.values.Matches(yv.values)
	default:
		return false
	}
}

func (d *Dict) Type() string {
	return "d"
}

func (d *Dict) Less(y Value) bool {
	switch yv := y.(type) {
	case *Dict:
		return d.keys.Less(yv.keys) || d.keys.Matches(yv.keys) && d.values.Less(yv.values)
	default:
		return d.Type() < y.Type()
	}
}

func (d *Dict) Len() int {
	return d.keys.Len()
}

func dict(x, y V) V {
	xv, ok := x.value.(array)
	if !ok {
		return Panicf("x!y : not an array x (%s)", x.Type())
	}
	yv, ok := y.value.(array)
	if !ok {
		return Panicf("x!y : not an array y (%s)", y.Type())
	}
	if xv.Len() != yv.Len() {
		return Panicf("x!y : length mismatch (%d vs %d)", xv.Len(), yv.Len())
	}
	return NewV(&Dict{keys: xv, values: yv})
}

func dictArith(xd, yd *Dict, f func(V, V) V) V {
	xd = xd.clone()
	xk, xv := xd.keys, xd.values
	yk, yv := yd.keys, yd.values
	ky := findArray(xk, NewV(yk))
	kyv := ky.value.(*AI)
	nkeys := xk.Len()
	for _, kyi := range kyv.Slice {
		if kyi == int64(nkeys) {
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
			break
		}
	}
	r, err := dictArithAmend(xv, ky.value.(*AI), f, yv)
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

func dictArithAmend(x array, yv *AI, f func(V, V) V, z array) (array, error) {
	var err error
	for i, yi := range yv.Slice {
		x, err = dictArithAmendI(x, yi, f, z.at(i))
		if err != nil {
			return x, err
		}
	}
	return x, nil
}
