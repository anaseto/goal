package main

// Conjugate returns +x.
func Conjugate(x Object) Object {
	switch x := x.(type) {
	case I, F:
		return x
	}
	// TODO: complex values (conjugate)
	return badtype("+")
}
