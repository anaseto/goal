package main

import (
	"math"
	"sort"
	"strings"
)

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

func signF(x F) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

func signI(x I) I {
	switch {
	case x > 0:
		return I(1)
	case x < 0:
		return I(-1)
	default:
		return I(0)
	}
}

func Sign(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return signF(x)
	case I:
		return signI(x)
	case AB:
		return x
	case AF:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signF(x[i])
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = signI(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Sign(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("×")
	}
}

func Reciprocal(x O) O {
	switch x := x.(type) {
	case B:
		return divide(1, B2F(x))
	case F:
		return divide(1, x)
	case I:
		return divide(1, F(x))
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(1, F(x[i]))
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Reciprocal(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("÷")
	}
}

func Floor(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Floor(x)
	case I:
		return x
	case S:
		return strings.ToLower(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Floor(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToLower(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Floor(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌊")
	}
}

func Ceil(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Ceil(x)
	case I:
		return x
	case S:
		return strings.ToUpper(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Ceil(x[i])
		}
		return r
	case AI:
		return x
	case AS:
		r := make(AS, len(x))
		for i := range r {
			r[i] = strings.ToUpper(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Ceil(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("⌈")
	}
}

func Not(x O) O {
	switch x := x.(type) {
	case B:
		return !x
	case F:
		return 1 - x
	case I:
		return 1 - x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = !x[i]
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = 1 - x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Not(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}

func absI(x I) I {
	if x < 0 {
		return -x
	}
	return x
}

func Abs(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return math.Abs(x)
	case I:
		return absI(x)
	case AB:
		return x
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = math.Abs(x[i])
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = absI(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = Abs(x[i])
		}
		return r
	case E:
		return x
	default:
		return badtype("¬")
	}
}

func clone(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			r[i] = clone(x[i])
		}
		return r
	case E:
		return x
	default:
		return x
	}
}

func cloneShallow(x O) O {
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		r := make(AB, len(x))
		copy(r, x)
		return r
	case AF:
		r := make(AF, len(x))
		copy(r, x)
		return r
	case AI:
		r := make(AI, len(x))
		copy(r, x)
		return r
	case AS:
		r := make(AS, len(x))
		copy(r, x)
		return r
	case AO:
		r := make(AO, len(x))
		copy(r, x)
		return r
	case E:
		return x
	default:
		return x
	}
}

type BoolSliceUp []bool

func (bs BoolSliceUp) Len() int {
	return len(bs)
}

func (bs BoolSliceUp) Less(i, j int) bool {
	return bs[j] && !bs[i]
}

func (bs BoolSliceUp) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

//type AOUp []O

//func (bs AOUp) Len() int {
//return len(bs)
//}

//func (bs AOUp) Less(i, j int) bool {
//return less(bs[i], bs[j])
//}

//func (bs AOUp) Swap(i, j int) {
//bs[i], bs[j] = bs[j], bs[i]
//}

//func less(w, x O) bool {
//switch w := w.(type) {
//case B:
//return lessB(w, x)
//case F:
//return lessF(w, x)
//case I:
//return lessI(w, x)
//case S:
//return lessS(w, x)
//case AB:
//if len(w) == 0 {
//return true
//}
//return x
//case AF:
//return x
//case AI:
//return x
//case AS:
//return x
//case AO:
//// TODO:
////r := make(AO, len(x))
////for i := range r {
////r[i] = SortUp(x[i])
////}
////return r
//return x
//default:
//return false
//}
//}

func SortUp(x O) O {
	x = cloneShallow(x)
	switch x := x.(type) {
	case B:
		return x
	case F:
		return x
	case I:
		return x
	case S:
		return x
	case AB:
		sort.Stable(BoolSliceUp(x))
		return x
	case AF:
		sort.Stable(sort.Float64Slice(x))
		return x
	case AI:
		sort.Stable(sort.IntSlice(x))
		return x
	case AS:
		sort.Stable(sort.StringSlice(x))
		return x
	case AO:
		// TODO:
		//r := make(AO, len(x))
		//for i := range r {
		//r[i] = SortUp(x[i])
		//}
		//return r
		return x
	case E:
		return x
	default:
		return badtype("<")
	}
}
