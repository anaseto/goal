package goal

// Matcher is implemented by types that can be matched againts other objects
// (typically a struct of the same type with fields that match).
type Matcher interface {
	Matches(x V) bool
}

// Match returns true if the two values match like in x~y.
func Match(x, y V) bool {
	switch x := x.(type) {
	case F:
		switch y := y.(type) {
		case I:
			return x == F(y)
		case F:
			return x == y
		default:
			return false
		}
	case I:
		switch y := y.(type) {
		case I:
			return x == y
		case F:
			return F(x) == y
		default:
			return false
		}
	case S:
		switch y := y.(type) {
		case S:
			return x == y
		default:
			return false
		}
	case Array:
		y, ok := y.(Array)
		if !ok {
			return false
		}
		l := x.Len()
		if l != y.Len() {
			return false
		}
		switch x := x.(type) {
		case AB:
			switch y := y.(type) {
			case AB:
				return matchAB(x, y)
			case AI:
				return matchABAI(x, y)
			case AF:
				return matchABAF(x, y)
			}
		case AI:
			switch y := y.(type) {
			case AB:
				return matchABAI(y, x)
			case AI:
				return matchAI(x, y)
			case AF:
				return matchAIAF(x, y)
			}
		case AF:
			switch y := y.(type) {
			case AB:
				return matchABAF(y, x)
			case AI:
				return matchAIAF(y, x)
			case AF:
				return matchAF(x, y)
			}
		case AS:
			y, ok := y.(AS)
			if !ok {
				break
			}
			for i, v := range y {
				if v != x[i] {
					return false
				}
			}
			return true
		}
		for i := 0; i < l; i++ {
			if !Match(x.At(i), y.At(i)) {
				return false
			}
		}
		return true
	case Matcher:
		return x.Matches(y)
	default:
		return x == y
	}
}

func matchAB(x, y AB) bool {
	for i, v := range y {
		if v != x[i] {
			return false
		}
	}
	return true
}

func matchABAI(x AB, y AI) bool {
	for i, v := range y {
		if v != int(B2I(x[i])) {
			return false
		}
	}
	return true
}

func matchABAF(x AB, y AF) bool {
	for i, v := range y {
		if F(v) != B2F(x[i]) {
			return false
		}
	}
	return true
}

func matchAI(x, y AI) bool {
	for i, v := range y {
		if v != x[i] {
			return false
		}
	}
	return true
}

func matchAIAF(x AI, y AF) bool {
	for i, v := range y {
		if F(v) != F(x[i]) {
			return false
		}
	}
	return true
}

func matchAF(x, y AF) bool {
	for i, v := range y {
		if v != x[i] {
			return false
		}
	}
	return true
}

// classify returns %x.
func classify(x V) V {
	if length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errf("%%x : x not an array (%s)", x.Type())
	case AB:
		v := x[0]
		if !v {
			return x
		}
		return not(x)
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
				if Match(v, x[j]) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return r
	default:
		return errf("%%x : x not an array (%s)", x.Type())
	}
}

// uniq returns ?x.
func uniq(x V) V {
	if length(x) == 0 {
		return x
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		// NOTE: ?atom could be used for something.
		return errf("?x : x not an array (%s)", x.Type())
	case AB:
		if len(x) == 0 {
			return x
		}
		b := x[0]
		for i := 1; i < len(x); i++ {
			if x[i] != b {
				return AB{b, x[i]}
			}
		}
		return AB{b}
	case AF:
		r := AF{}
		m := map[float64]struct{}{}
		for _, v := range x {
			_, ok := m[v]
			if !ok {
				r = append(r, v)
				m[v] = struct{}{}
			}
		}
		return r
	case AI:
		r := AI{}
		m := map[int]struct{}{}
		for _, v := range x {
			_, ok := m[v]
			if !ok {
				r = append(r, v)
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AS:
		r := AS{}
		m := map[string]struct{}{}
		for _, v := range x {
			_, ok := m[v]
			if !ok {
				r = append(r, v)
				m[v] = struct{}{}
				continue
			}
		}
		return r
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing, but that would be
		// quite bad for short lengths.
		r := make(AV, len(x))
	loop:
		for i, v := range x {
			for j := range x[:i] {
				if Match(v, x[j]) {
					continue loop
				}
			}
			r = append(r, v)
		}
		return r
	default:
		return errf("?x : x not an array (%s)", x.Type())
	}
}

// Mark Firsts returns ∊x. XXX unused for now
func markFirsts(x V) V {
	if length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errf("∊x : x not an array (%s)", x.Type())
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
				if Match(v, x[j]) {
					continue loop
				}
			}
			r[i] = true
		}
		return r
	default:
		return errf("∊x : x not an array (%s)", x.Type())
	}
}

// memberOf returns x∊y. XXX unused for now
func memberOf(x, y V) V {
	if length(y) == 0 || length(x) == 0 {
		switch y.(type) {
		case Array:
			switch x := x.(type) {
			case Array:
				r := make(AB, length(x))
				return r
			default:
				return B2I(false)
			}
		default:
			return errf("x∊y : y not an array (%s)", y.Type())
		}
	}
	y = canonical(y)
	x = canonical(x)
	switch y := y.(type) {
	case AB:
		return memberOfAB(x, y)
	case AF:
		return memberOfAF(x, y)
	case AI:
		return memberOfAI(x, y)
	case AS:
		return memberOfAS(x, y)
	case AV:
		return memberOfAO(x, y)
	default:
		return errf("x∊y : y not an array (%s)", y.Type())
	}
}

func memberOfAB(x V, y AB) V {
	var t, f bool
	for _, v := range y {
		if t && f {
			break
		}
		t, f = t || v, f || !v
	}
	if t && f {
		r := make(AB, length(x))
		for i := range r {
			r[i] = true
		}
		return r
	}
	if t {
		return equal(x, B2I(true))
	}
	return equal(x, B2I(false))
}

func memberOfAF(x V, y AF) V {
	m := map[F]struct{}{}
	for _, v := range y {
		_, ok := m[F(v)]
		if !ok {
			m[F(v)] = struct{}{}
			continue
		}
	}
	switch x := x.(type) {
	case I:
		_, ok := m[F(x)]
		return B2I(ok)
	case F:
		_, ok := m[x]
		return B2I(ok)
	case AB:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[B2F(v)]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[F(v)]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[F(v)]
		}
		return r
	default:
		return make(AB, length(x))
	}
}

func memberOfAI(x V, y AI) V {
	m := map[int]struct{}{}
	for _, v := range y {
		_, ok := m[v]
		if !ok {
			m[v] = struct{}{}
			continue
		}
	}
	switch x := x.(type) {
	case I:
		_, ok := m[int(x)]
		return B2I(ok)
	case F:
		if !isI(x) {
			return B2I(false)
		}
		_, ok := m[int(x)]
		return B2I(ok)
	case AB:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[int(B2I(v))]
		}
		return r
	case AI:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[v]
		}
		return r
	case AF:
		r := make(AB, len(x))
		for i, v := range x {
			if !isI(F(v)) {
				continue
			}
			_, r[i] = m[int(v)]
		}
		return r
	default:
		return make(AB, length(x))
	}
}

func memberOfAS(x V, y AS) V {
	m := map[string]struct{}{}
	for _, v := range y {
		_, ok := m[v]
		if !ok {
			m[v] = struct{}{}
			continue
		}
	}
	switch x := x.(type) {
	case S:
		_, ok := m[string(x)]
		return B2I(ok)
	case AS:
		r := make(AB, len(x))
		for i, v := range x {
			_, r[i] = m[v]
		}
		return r
	default:
		return make(AB, length(x))
	}
}

func memberOfAO(x V, y AV) V {
	switch x := x.(type) {
	case Array:
		// NOTE: quadratic algorithm
		r := make(AB, x.Len())
		for i := 0; i < x.Len(); i++ {
			for _, v := range y {
				if Match(x.At(i), v) {
					r[i] = true
					break
				}
			}
		}
		return r
	default:
		for _, v := range y {
			if Match(x, v) {
				return B2I(true)
			}
		}
		return B2I(false)
	}
}

// OccurrenceCount returns ⊒x. XXX unused for now
func occurrenceCount(x V) V {
	if length(x) == 0 {
		return AB{}
	}
	x = canonical(x)
	switch x := x.(type) {
	case F, I, S:
		return errf("⊒x : x not an array (%s)", x.Type())
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
				if Match(v, x[j]) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return r
	default:
		return errf("⊒x : x not an array (%s)", x.Type())
	}
}
