package main

// Matcher is implemented by types that can be matched againts other objects
// (typically a struct of the same type with fields that match).
type Matcher interface {
	Matches(x O) bool
}

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
		x, ok := x.(Array)
		if !ok {
			return false
		}
		l := w.Len()
		if l != x.Len() {
			return false
		}
		switch w := w.(type) {
		case AB:
			switch x := x.(type) {
			case AB:
				return matchAB(w, x)
			case AI:
				return matchABAI(w, x)
			case AF:
				return matchABAF(w, x)
			}
		case AI:
			switch x := x.(type) {
			case AB:
				return matchABAI(x, w)
			case AI:
				return matchAI(w, x)
			case AF:
				return matchAIAF(w, x)
			}
		case AF:
			switch x := x.(type) {
			case AB:
				return matchABAF(x, w)
			case AI:
				return matchAIAF(x, w)
			case AF:
				return matchAF(w, x)
			}
		case AS:
			x, ok := x.(AS)
			if !ok {
				break
			}
			for i, v := range x {
				if v != w[i] {
					return false
				}
			}
			return true
		}
		for i := 0; i < l; i++ {
			if !match(w.At(i), x.At(i)) {
				return false
			}
		}
		return true
	case Matcher:
		return w.Matches(x)
	default:
		return w == x
	}
}

func matchAB(w, x AB) bool {
	for i, v := range x {
		if v != w[i] {
			return false
		}
	}
	return true
}

func matchABAI(w AB, x AI) bool {
	for i, v := range x {
		if v != B2I(w[i]) {
			return false
		}
	}
	return true
}

func matchABAF(w AB, x AF) bool {
	for i, v := range x {
		if v != B2F(w[i]) {
			return false
		}
	}
	return true
}

func matchAI(w, x AI) bool {
	for i, v := range x {
		if v != w[i] {
			return false
		}
	}
	return true
}

func matchAIAF(w AI, x AF) bool {
	for i, v := range x {
		if v != F(w[i]) {
			return false
		}
	}
	return true
}

func matchAF(w, x AF) bool {
	for i, v := range x {
		if v != w[i] {
			return false
		}
	}
	return true
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
	x = canonical(x)
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
		for i, v := range x {
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

// Mark Firts returns ∊x.
func MarkFirts(x O) O {
	if Length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case B, F, I, S:
		return badtype("∊ : expected array")
	case AB:
		r := make(AB, len(x))
		r[0] = true
		v := x[0]
		for i := 1; i < len(x); i++ {
			if x[i] != v {
				r[i] = true
				break
			}
		}
		return r
	case AF:
		r := make(AB, len(x))
		m := map[F]struct{}{}
		for i, v := range x {
			_, ok := m[v]
			if !ok {
				r[i] = true
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AI:
		r := make(AB, len(x))
		m := map[I]struct{}{}
		for i, v := range x {
			_, ok := m[v]
			if !ok {
				r[i] = true
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AS:
		r := make(AB, len(x))
		m := map[S]struct{}{}
		for i, v := range x {
			_, ok := m[v]
			if !ok {
				r[i] = true
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AO:
		// TODO: optimize common cases? (quadratic algorithm, worst
		// case complexity could be improved by sorting or string
		// hashing, but that would be quite bad for short lengths)
		r := make(AB, len(x))
	loop:
		for i, v := range x {
			for j := range x[:i] {
				if match(v, x[j]) {
					continue loop
				}
			}
			r[i] = true
		}
		return r
	default:
		return badtype("⊐ : expected array")
	}
}

// OccurrenceCount returns ⊒x.
func OccurrenceCount(x O) O {
	if Length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case B, F, I, S:
		return badtype("⊒ : expected array")
	case AB:
		r := make(AI, len(x))
		var f, t int
		for i, v := range x {
			if v {
				r[i] = t
				t++
				continue
			}
			r[i] = f
			f++
		}
		return r
	case AF:
		r := make(AI, len(x))
		m := map[F]I{}
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				m[v] = 0
				continue
			}
			m[v] = c + 1
			r[i] = c + 1
		}
		return r
	case AI:
		r := make(AI, len(x))
		m := map[I]I{}
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				m[v] = 0
				continue
			}
			m[v] = c + 1
			r[i] = c + 1
		}
		return r
	case AS:
		r := make(AI, len(x))
		m := map[S]I{}
		for i, v := range x {
			c, ok := m[v]
			if !ok {
				m[v] = 0
				continue
			}
			m[v] = c + 1
			r[i] = c + 1
		}
		return r
	case AO:
		// TODO: optimize common cases? (quadratic algorithm, worst
		// case complexity could be improved by sorting or string
		// hashing, but that would be quite bad for short lengths)
		r := make(AI, len(x))
	loop:
		for i, v := range x {
			for j := i - 1; j >= 0; j-- {
				if match(v, x[j]) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return r
	default:
		return badtype("⊒ : expected array")
	}
}
