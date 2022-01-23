package main

func Negate(x O) O {
	switch x := x.(type) {
	case B:
		return -B2I(x)
	case F:
		return -x
	case I:
		return -x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = -x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Negate(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("=")
	}
}
