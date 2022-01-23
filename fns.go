package main

func length(x O) int {
	switch x := x.(type) {
	case AB:
		return len(x)
	case AF:
		return len(x)
	case AI:
		return len(x)
	case AS:
		return len(x)
	case AO:
		return len(x)
	default:
		return 1
	}
}

func at(x O, i int) O {
	if length(x) <= i {
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
