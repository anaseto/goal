package main

// Match returns w≡x.
func Match(w, x O) O {
	return match(w, x)
}

func match(w, x O) bool {
	switch w := w.(type) {
	case B:
		switch x := x.(type) {
		case B:
			return w == x
		case I:
			return B2I(w) == x
		case F:
			return B2F(w) == x
		default:
			return false
		}
	case F:
		switch x := x.(type) {
		case B:
			return w == B2F(x)
		case I:
			return w == F(x)
		case F:
			return w == x
		default:
			return false
		}
	case I:
		switch x := x.(type) {
		case B:
			return w == B2I(x)
		case I:
			return w == x
		case F:
			return F(w) == x
		default:
			return false
		}
	case S:
		switch x := x.(type) {
		case S:
			return w == x
		default:
			return false
		}
	case Array:
		// TODO: optimize common cases
		switch x := x.(type) {
		case Array:
			l := Length(w)
			if l != Length(x) {
				return false
			}
			for i := 0; i < l; i++ {
				if !match(w.At(i), x.At(i)) {
					return false
				}
			}
			return true
		default:
			return false
		}
	default:
		// TODO: matching interface?
		return w == x
	}
}

// NotMatch returns w≢x.
func NotMatch(w, x O) O {
	return !match(w, x)
}

// Classify returns ⊐x.
func Classify(x O) O {
	if Length(x) == 0 {
		return AB{}
	}
	switch x := x.(type) {
	case B, F, I, S:
		return badtype("⊐ : expected array")
	case AB:
		v := x[0]
		if !v {
			return x
		}
		return Not(x)
	case AF:
		r := make(AI, len(x))
		m := map[F]I{}
		n := 0
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				r[i] = n
				m[v] = n
				n++
				continue
			}
			r[i] = c
		}
		return r
	case AI:
		r := make(AI, len(x))
		m := map[I]I{}
		n := 0
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				r[i] = n
				m[v] = n
				n++
				continue
			}
			r[i] = c
		}
		return r
	case AS:
		r := make(AI, len(x))
		m := map[S]I{}
		n := 0
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				r[i] = n
				m[v] = n
				n++
				continue
			}
			r[i] = c
		}
		return r
	case AO:
		// TODO: optimize common cases? (quadratic algorithm, worst
		// case complexity could be improved by sorting or string
		// hashing, but that would be quite bad for short lengths)
		r := make(AI, len(x))
		n := 0
	loop:
		for i := range r {
			v := x[i]
			for j := range x[:i] {
				if match(v, x[j]) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return r
	default:
		return badtype("⊐ : expected array")
	}
}
