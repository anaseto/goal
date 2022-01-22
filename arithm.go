package main

// Conjugate returns +x.
func Conjugate(x O) O {
	switch x := x.(type) {
	case I, F:
		return x
	}
	// TODO: complex values (conjugate)
	return badtype("+")
}
