package main

// Matcher is implemented by types that can be matched againts other objects
// (typically a struct of the same type with fields that match).
type Matcher interface {
	Matches(x V) bool
}

// Match returns w≡x.
func Match(w, x V) V {
	return B2I(match(w, x))
}

func match(w, x V) bool {
	switch w := w.(type) {
	case F:
		switch x := x.(type) {
		case I:
			return w == F(x)
		case F:
			return w == x
		default:
			return false
		}
	case I:
		switch x := x.(type) {
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
		if v != int(B2I(w[i])) {
			return false
		}
	}
	return true
}

func matchABAF(w AB, x AF) bool {
	for i, v := range x {
		if F(v) != B2F(w[i]) {
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
		if F(v) != F(w[i]) {
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

// Classify returns ⊐x.
func Classify(x V) V {
	if Length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errs("not an array")
	case AB:
		v := x[0]
		if !v {
			return x
		}
		return Not(x)
	case AF:
		r := make(AI, len(x))
		m := map[float64]int{}
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
		m := map[int]int{}
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
		m := map[string]int{}
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
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing, but that would be
		// quite bad for short lengths.
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
		return errs("not an array")
	}
}

// Mark Firsts returns ∊x.
func MarkFirsts(x V) V {
	if Length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errs("not an array")
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
		m := map[float64]struct{}{}
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
		m := map[int]struct{}{}
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
		m := map[string]struct{}{}
		for i, v := range x {
			_, ok := m[v]
			if !ok {
				r[i] = true
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing, but that would be
		// quite bad for short lengths.
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
		return errs("not an array")
	}
}

// MemberOf returns w∊x.
func MemberOf(w, x V) V {
	if Length(x) == 0 || Length(w) == 0 {
		switch x.(type) {
		case Array:
			switch w := w.(type) {
			case Array:
				r := make(AB, Length(w))
				return r
			default:
				return B2I(false)
			}
		default:
			return errs("not an array")
		}
	}
	x = canonical(x)
	w = canonical(w)
	switch x := x.(type) {
	case AB:
		return memberOfAB(w, x)
	case AF:
		return memberOfAF(w, x)
	case AI:
		return memberOfAI(w, x)
	case AS:
		return memberOfAS(w, x)
	case AV:
		return memberOfAO(w, x)
	default:
		return errs("not an array")
	}
}

func memberOfAB(w V, x AB) V {
	var t, f bool
	for _, v := range x {
		if t && f {
			break
		}
		t, f = t || v, f || !v
	}
	if t && f {
		r := make(AB, Length(w))
		for i := range r {
			r[i] = true
		}
		return r
	}
	if t {
		return Equal(w, B2I(true))
	}
	return Equal(w, B2I(false))
}

func memberOfAF(w V, x AF) V {
	m := map[F]struct{}{}
	for _, v := range x {
		_, ok := m[F(v)]
		if !ok {
			m[F(v)] = struct{}{}
			continue
		}
	}
	switch w := w.(type) {
	case I:
		_, ok := m[F(w)]
		return B2I(ok)
	case F:
		_, ok := m[w]
		return B2I(ok)
	case AB:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[B2F(v)]
		}
		return r
	case AI:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[F(v)]
		}
		return r
	case AF:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[F(v)]
		}
		return r
	default:
		return make(AB, Length(w))
	}
}

func memberOfAI(w V, x AI) V {
	m := map[int]struct{}{}
	for _, v := range x {
		_, ok := m[v]
		if !ok {
			m[v] = struct{}{}
			continue
		}
	}
	switch w := w.(type) {
	case I:
		_, ok := m[int(w)]
		return B2I(ok)
	case F:
		if !isI(w) {
			return B2I(false)
		}
		_, ok := m[int(w)]
		return B2I(ok)
	case AB:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[int(B2I(v))]
		}
		return r
	case AI:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[v]
		}
		return r
	case AF:
		r := make(AB, len(w))
		for i, v := range w {
			if !isI(F(v)) {
				continue
			}
			_, r[i] = m[int(v)]
		}
		return r
	default:
		return make(AB, Length(w))
	}
}

func memberOfAS(w V, x AS) V {
	m := map[string]struct{}{}
	for _, v := range x {
		_, ok := m[v]
		if !ok {
			m[v] = struct{}{}
			continue
		}
	}
	switch w := w.(type) {
	case S:
		_, ok := m[string(w)]
		return B2I(ok)
	case AS:
		r := make(AB, len(w))
		for i, v := range w {
			_, r[i] = m[v]
		}
		return r
	default:
		return make(AB, Length(w))
	}
}

func memberOfAO(w V, x AV) V {
	switch w := w.(type) {
	case Array:
		// NOTE: quadratic algorithm
		r := make(AB, w.Len())
		for i := 0; i < w.Len(); i++ {
			for _, v := range x {
				if match(w.At(i), v) {
					r[i] = true
					break
				}
			}
		}
		return r
	default:
		for _, v := range x {
			if match(w, v) {
				return B2I(true)
			}
		}
		return B2I(false)
	}
}

// OccurrenceCount returns ⊒x.
func OccurrenceCount(x V) V {
	if Length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errs("not an array")
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
		m := map[float64]int{}
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
		m := map[int]int{}
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
		m := map[string]int{}
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
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing, but that would be
		// quite bad for short lengths.
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
		return errs("not an array")
	}
}
