package main

func Equal(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return EqualBO(w, x)
	case F:
		return EqualFO(w, x)
	case I:
		return EqualIO(w, x)
	case S:
		return EqualSO(w, x)
	case AB:
		return EqualABO(w, x)
	case AF:
		return EqualAFO(w, x)
	case AI:
		return EqualAIO(w, x)
	case AS:
		return EqualASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Equal(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return w == x
	case F:
		return B2F(w) == x
	case I:
		return B2I(w) == x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == x[i]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) == x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := EqualBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w == B2F(x)
	case F:
		return w == x
	case I:
		return w == F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := EqualFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w == B2I(x)
	case F:
		return F(w) == x
	case I:
		return w == x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := EqualIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w == x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w == x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := EqualSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == x
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) == x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) == x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == x[i]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) == x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := EqualABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := EqualAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) == x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) == x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := EqualAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func EqualASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] == x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] == x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := EqualASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("=")
	}
}

func NotEqual(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return NotEqualBO(w, x)
	case F:
		return NotEqualFO(w, x)
	case I:
		return NotEqualIO(w, x)
	case S:
		return NotEqualSO(w, x)
	case AB:
		return NotEqualABO(w, x)
	case AF:
		return NotEqualAFO(w, x)
	case AI:
		return NotEqualAIO(w, x)
	case AS:
		return NotEqualASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := NotEqual(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return w != x
	case F:
		return B2F(w) != x
	case I:
		return B2I(w) != x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != x[i]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) != x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := NotEqualBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w != B2F(x)
	case F:
		return w != x
	case I:
		return w != F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := NotEqualFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w != B2I(x)
	case F:
		return F(w) != x
	case I:
		return w != x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := NotEqualIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w != x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w != x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := NotEqualSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != x
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) != x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) != x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != x[i]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) != x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := NotEqualABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := NotEqualAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) != x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) != x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := NotEqualAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func NotEqualASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] != x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] != x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := NotEqualASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≠")
	}
}

func Lesser(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return LesserBO(w, x)
	case F:
		return LesserFO(w, x)
	case I:
		return LesserIO(w, x)
	case S:
		return LesserSO(w, x)
	case AB:
		return LesserABO(w, x)
	case AF:
		return LesserAFO(w, x)
	case AI:
		return LesserAIO(w, x)
	case AS:
		return LesserASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Lesser(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) < B2I(x)
	case F:
		return B2F(w) < x
	case I:
		return B2I(w) < x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) < B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) < x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w < B2F(x)
	case F:
		return w < x
	case I:
		return w < F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w < B2I(x)
	case F:
		return F(w) < x
	case I:
		return w < x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w < x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w < x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) < B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) < x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) < x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) < B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) < x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) < x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) < x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] < x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] < x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("<")
	}
}

func LesserEq(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return LesserEqBO(w, x)
	case F:
		return LesserEqFO(w, x)
	case I:
		return LesserEqIO(w, x)
	case S:
		return LesserEqSO(w, x)
	case AB:
		return LesserEqABO(w, x)
	case AF:
		return LesserEqAFO(w, x)
	case AI:
		return LesserEqAIO(w, x)
	case AS:
		return LesserEqASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := LesserEq(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) <= B2I(x)
	case F:
		return B2F(w) <= x
	case I:
		return B2I(w) <= x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) <= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) <= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserEqBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w <= B2F(x)
	case F:
		return w <= x
	case I:
		return w <= F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserEqFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w <= B2I(x)
	case F:
		return F(w) <= x
	case I:
		return w <= x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserEqIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w <= x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w <= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := LesserEqSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) <= B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) <= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) <= x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) <= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) <= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserEqABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserEqAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) <= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) <= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserEqAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func LesserEqASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] <= x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] <= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := LesserEqASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≤")
	}
}

func Greater(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return GreaterBO(w, x)
	case F:
		return GreaterFO(w, x)
	case I:
		return GreaterIO(w, x)
	case S:
		return GreaterSO(w, x)
	case AB:
		return GreaterABO(w, x)
	case AF:
		return GreaterAFO(w, x)
	case AI:
		return GreaterAIO(w, x)
	case AS:
		return GreaterASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Greater(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) > B2I(x)
	case F:
		return B2F(w) > x
	case I:
		return B2I(w) > x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) > B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) > x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w > B2F(x)
	case F:
		return w > x
	case I:
		return w > F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w > B2I(x)
	case F:
		return F(w) > x
	case I:
		return w > x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w > x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w > x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) > B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) > x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) > x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) > B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) > x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) > x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) > x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] > x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] > x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype(">")
	}
}

func GreaterEq(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return GreaterEqBO(w, x)
	case F:
		return GreaterEqFO(w, x)
	case I:
		return GreaterEqIO(w, x)
	case S:
		return GreaterEqSO(w, x)
	case AB:
		return GreaterEqABO(w, x)
	case AF:
		return GreaterEqAFO(w, x)
	case AI:
		return GreaterEqAIO(w, x)
	case AS:
		return GreaterEqASO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := GreaterEq(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) >= B2I(x)
	case F:
		return B2F(w) >= x
	case I:
		return B2I(w) >= x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) >= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w) >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w) >= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterEqBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w >= B2F(x)
	case F:
		return w >= x
	case I:
		return w >= F(x)
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterEqFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w >= B2I(x)
	case F:
		return F(w) >= x
	case I:
		return w >= x
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w) >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterEqIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqSO(w S, x Object) Object {
	switch x := x.(type) {
	case S:
		return w >= x
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w >= x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := GreaterEqSO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) >= B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2F(w[i]) >= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = B2I(w[i]) >= x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) >= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2F(w[i]) >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = B2I(w[i]) >= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterEqABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= B2F(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= F(x)
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= B2F(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterEqAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= B2I(x)
		}
		return r
	case F:
		r := make(AB, len(w))
		for i := range r {
			r[i] = F(w[i]) >= x
		}
		return r
	case I:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= x
		}
		return r
	case AB:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= B2I(x[i])
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i := range r {
			r[i] = F(w[i]) >= x[i]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterEqAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func GreaterEqASO(w AS, x Object) Object {
	switch x := x.(type) {
	case S:
		r := make(AB, len(w))
		for i := range r {
			r[i] = w[i] >= x
		}
		return r
	case AS:
		r := make(AB, len(x))
		for i := range r {
			r[i] = w[i] >= x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := GreaterEqASO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("≥")
	}
}

func Add(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return AddBO(w, x)
	case F:
		return AddFO(w, x)
	case I:
		return AddIO(w, x)
	case AB:
		return AddABO(w, x)
	case AF:
		return AddAFO(w, x)
	case AI:
		return AddAIO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Add(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) + B2I(x)
	case F:
		return B2F(w) + x
	case I:
		return B2I(w) + x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) + B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w) + x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) + x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := AddBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w + B2F(x)
	case F:
		return w + x
	case I:
		return w + F(x)
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w + B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w + x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w + F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := AddFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w + B2I(x)
	case F:
		return F(w) + x
	case I:
		return w + x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w + B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w) + x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w + x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := AddIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) + B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = B2F(w[i]) + x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) + x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) + B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w[i]) + x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) + x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := AddABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] + B2F(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] + x
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] + F(x)
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] + B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] + x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] + F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := AddAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func AddAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] + B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = F(w[i]) + x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] + x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] + B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w[i]) + x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] + x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := AddAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("+")
	}
}

func Subtract(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return SubtractBO(w, x)
	case F:
		return SubtractFO(w, x)
	case I:
		return SubtractIO(w, x)
	case AB:
		return SubtractABO(w, x)
	case AF:
		return SubtractAFO(w, x)
	case AI:
		return SubtractAIO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Subtract(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) - B2I(x)
	case F:
		return B2F(w) - x
	case I:
		return B2I(w) - x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) - B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w) - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) - x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := SubtractBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w - B2F(x)
	case F:
		return w - x
	case I:
		return w - F(x)
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w - B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w - x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w - F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := SubtractFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w - B2I(x)
	case F:
		return F(w) - x
	case I:
		return w - x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w - B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w) - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w - x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := SubtractIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) - B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = B2F(w[i]) - x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) - x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) - B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w[i]) - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) - x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := SubtractABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] - B2F(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] - x
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] - F(x)
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] - B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] - x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] - F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := SubtractAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func SubtractAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] - B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = F(w[i]) - x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] - x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] - B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w[i]) - x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] - x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := SubtractAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("-")
	}
}

func Multiply(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return MultiplyBO(w, x)
	case F:
		return MultiplyFO(w, x)
	case I:
		return MultiplyIO(w, x)
	case AB:
		return MultiplyABO(w, x)
	case AF:
		return MultiplyAFO(w, x)
	case AI:
		return MultiplyAIO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Multiply(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return B2I(w) * B2I(x)
	case F:
		return B2F(w) * x
	case I:
		return B2I(w) * x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) * B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w) * x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w) * x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := MultiplyBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return w * B2F(x)
	case F:
		return w * x
	case I:
		return w * F(x)
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w * B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w * x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w * F(x[i])
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := MultiplyFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return w * B2I(x)
	case F:
		return F(w) * x
	case I:
		return w * x
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w * B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w) * x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w * x[i]
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := MultiplyIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) * B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = B2F(w[i]) * x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = B2I(w[i]) * x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) * B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = B2F(w[i]) * x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = B2I(w[i]) * x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := MultiplyABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] * B2F(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] * x
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = w[i] * F(x)
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] * B2F(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] * x[i]
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = w[i] * F(x[i])
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := MultiplyAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func MultiplyAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] * B2I(x)
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = F(w[i]) * x
		}
		return r
	case I:
		r := make(AI, len(w))
		for i := range r {
			r[i] = w[i] * x
		}
		return r
	case AB:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] * B2I(x[i])
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = F(w[i]) * x[i]
		}
		return r
	case AI:
		r := make(AI, len(x))
		for i := range r {
			r[i] = w[i] * x[i]
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := MultiplyAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("×")
	}
}

func Divide(w, x Object) Object {
	switch w := w.(type) {
	case B:
		return DivideBO(w, x)
	case F:
		return DivideFO(w, x)
	case I:
		return DivideIO(w, x)
	case AB:
		return DivideABO(w, x)
	case AF:
		return DivideAFO(w, x)
	case AI:
		return DivideAIO(w, x)
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := Divide(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideBO(w B, x Object) Object {
	switch x := x.(type) {
	case B:
		return divide(B2F(w), B2F(x))
	case F:
		return divide(B2F(w), x)
	case I:
		return divide(B2F(w), F(x))
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w), B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w), x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w), F(x[i]))
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := DivideBO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideFO(w F, x Object) Object {
	switch x := x.(type) {
	case B:
		return divide(w, B2F(x))
	case F:
		return divide(w, x)
	case I:
		return divide(w, F(x))
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w, B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w, x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w, F(x[i]))
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := DivideFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideIO(w I, x Object) Object {
	switch x := x.(type) {
	case B:
		return divide(F(w), B2F(x))
	case F:
		return divide(F(w), x)
	case I:
		return divide(F(w), F(x))
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w), B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w), x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w), F(x[i]))
		}
		return r
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := DivideIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideABO(w AB, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(B2F(w[i]), B2F(x))
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(B2F(w[i]), x)
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(B2F(w[i]), F(x))
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w[i]), B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w[i]), x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(B2F(w[i]), F(x[i]))
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := DivideABO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideAFO(w AF, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(w[i], B2F(x))
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(w[i], x)
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(w[i], F(x))
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w[i], B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w[i], x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(w[i], F(x[i]))
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := DivideAFO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

func DivideAIO(w AI, x Object) Object {
	switch x := x.(type) {
	case B:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(F(w[i]), B2F(x))
		}
		return r
	case F:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(F(w[i]), x)
		}
		return r
	case I:
		r := make(AF, len(w))
		for i := range r {
			r[i] = divide(F(w[i]), F(x))
		}
		return r
	case AB:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w[i]), B2F(x[i]))
		}
		return r
	case AF:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w[i]), x[i])
		}
		return r
	case AI:
		r := make(AF, len(x))
		for i := range r {
			r[i] = divide(F(w[i]), F(x[i]))
		}
		return r
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := DivideAIO(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("÷")
	}
}

