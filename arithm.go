package main

// Conjugate returns +x.
func Conjugate(x V) V {
	switch x := x.(type) {
	case I, F:
		return x
	}
	// TODO: complex values (conjugate)
	return badtype("+")
}
