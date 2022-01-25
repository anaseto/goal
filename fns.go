package main

func at(x O, i int) O {
	if Length(x) <= i {
		return badlen("at")
	}
	switch x := x.(type) {
	case AB:
		return x[i]
	case AF:
		return x[i]
	case AI:
		return x[i]
	case AS:
		return x[i]
	case AO:
		return x[i]
	default:
		return badtype("at")
	}
}
