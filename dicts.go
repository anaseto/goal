package goal

// Dict represents a dictionnary.
type Dict struct {
	keys   array
	values array
}

// NewDict returns a dictionnary. Both keys and values should be arrays, and
// they should have the same length.
func NewDict(keys, values V) V {
	return key(keys, values)
}

func newDict(keys array, values V) V {
	values.InitRC()
	return NewV(&Dict{keys: keys, values: values.value.(array)})
}

// Keys returns the keys of the dictionnary.
func (d *Dict) Keys() V {
	return NewV(d.keys)
}

// Values returns the values of the dictionnary.
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

func key(x, y V) V {
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
