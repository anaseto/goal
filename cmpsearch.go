package goal

import "strings"

// Matcher is implemented by types that can be matched againts other objects
// (typically a struct of the same type with fields that match).
type Matcher interface {
	Matches(x V) bool
}

// Match returns true if the two values match like in x~y.
func Match(x, y V) bool {
	return x.Value != nil && x.Value.Matches(y.Value) || x.Value == nil && y.Value == nil
}

func matchArray(x array, y Value) bool {
	ya, ok := y.(array)
	if !ok {
		return false
	}
	l := x.Len()
	if l != ya.Len() {
		return false
	}
	switch x := x.(type) {
	case AB:
		switch ya := ya.(type) {
		case AB:
			return matchAB(x, ya)
		case AI:
			return matchABAI(x, ya)
		case AF:
			return matchABAF(x, ya)
		}
	case AI:
		switch ya := ya.(type) {
		case AB:
			return matchABAI(ya, x)
		case AI:
			return matchAI(x, ya)
		case AF:
			return matchAIAF(x, ya)
		}
	case AF:
		switch ya := ya.(type) {
		case AB:
			return matchABAF(ya, x)
		case AI:
			return matchAIAF(ya, x)
		case AF:
			return matchAF(x, ya)
		}
	case AS:
		ya, ok := ya.(AS)
		if !ok {
			break
		}
		for i, yi := range ya {
			if yi != x[i] {
				return false
			}
		}
		return true
	}
	for i := 0; i < l; i++ {
		if !Match(x.at(i), ya.at(i)) {
			return false
		}
	}
	return true
}

func matchAB(x, y AB) bool {
	for i, yi := range y {
		if yi != x[i] {
			return false
		}
	}
	return true
}

func matchABAI(x AB, y AI) bool {
	for i, yi := range y {
		if yi != int(B2I(x[i])) {
			return false
		}
	}
	return true
}

func matchABAF(x AB, y AF) bool {
	for i, yi := range y {
		if F(yi) != B2F(x[i]) {
			return false
		}
	}
	return true
}

func matchAI(x, y AI) bool {
	for i, yi := range y {
		if yi != x[i] {
			return false
		}
	}
	return true
}

func matchAIAF(x AI, y AF) bool {
	for i, yi := range y {
		if F(yi) != F(x[i]) {
			return false
		}
	}
	return true
}

func matchAF(x, y AF) bool {
	for i, yi := range y {
		if yi != x[i] {
			return false
		}
	}
	return true
}

// classify returns %x.
func classify(x V) V {
	if Length(x) == 0 {
		return NewV(AB{})
	}
	//assertCanonical(x)
	switch xv := x.Value.(type) {
	case F, I, S:
		return errf("%%x : x not an array (%s)", x.Type())
	case AB:
		if !xv[0] {
			return NewV(xv)
		}
		return not(x)
	case AF:
		r := make(AI, xv.Len())
		m := map[float64]int{}
		n := 0
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewV(r)
	case AI:
		r := make(AI, xv.Len())
		m := map[int]int{}
		n := 0
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewV(r)
	case AS:
		r := make(AI, xv.Len())
		m := map[string]int{}
		n := 0
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewV(r)
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make(AI, xv.Len())
		n := 0
	loop:
		for i, xi := range xv {
			for j := range xv[:i] {
				if Match(xi, xv[j]) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return NewV(r)
	default:
		return errf("%%x : x not an array (%s)", x.Type())
	}
}

// uniq returns ?x.
func uniq(x V) V {
	if Length(x) == 0 {
		return x
	}
	//assertCanonical(x)
	switch x := x.Value.(type) {
	case F, I, S:
		// NOTE: ?atom could be used for something.
		return errf("?x : x not an array (%s)", x.Type())
	case AB:
		if x.Len() == 0 {
			return NewV(x)
		}
		b := x[0]
		for i := 1; i < x.Len(); i++ {
			if x[i] != b {
				return NewV(AB{b, x[i]})
			}
		}
		return NewV(AB{b})
	case AF:
		r := AF{}
		m := map[float64]struct{}{}
		for _, xi := range x {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
			}
		}
		return NewV(r)
	case AI:
		r := AI{}
		m := map[int]struct{}{}
		for _, xi := range x {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
				continue
			}
		}
		return NewV(r)
	case AS:
		r := AS{}
		m := map[string]struct{}{}
		for _, xi := range x {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
				continue
			}
		}
		return NewV(r)
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make(AV, x.Len())
	loop:
		for i, xi := range x {
			for j := range x[:i] {
				if Match(xi, x[j]) {
					continue loop
				}
			}
			r = append(r, xi)
		}
		return NewV(canonical(r))
	default:
		return errf("?x : x not an array (%s)", x.Type())
	}
}

// Mark Firsts returns ∊x. XXX unused for now
func markFirsts(x V) V {
	if Length(x) == 0 {
		return NewV(AB{})
	}
	//assertCanonical(x)
	switch x := x.Value.(type) {
	case F, I, S:
		return errf("∊x : x not an array (%s)", x.Type())
	case AB:
		r := make(AB, x.Len())
		r[0] = true
		x0 := x[0]
		for i := 1; i < x.Len(); i++ {
			if x[i] != x0 {
				r[i] = true
				break
			}
		}
		return NewV(r)
	case AF:
		r := make(AB, x.Len())
		m := map[float64]struct{}{}
		for i, xi := range x {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewV(r)
	case AI:
		r := make(AB, x.Len())
		m := map[int]struct{}{}
		for i, xi := range x {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewV(r)
	case AS:
		r := make(AB, x.Len())
		m := map[string]struct{}{}
		for i, xi := range x {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewV(r)
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make(AB, x.Len())
	loop:
		for i, xi := range x {
			for j := range x[:i] {
				if Match(xi, x[j]) {
					continue loop
				}
			}
			r[i] = true
		}
		return NewV(r)
	default:
		return errf("∊x : x not an array (%s)", x.Type())
	}
}

// memberOf returns x in y.
func memberOf(x, y V) V {
	if Length(y) == 0 {
		switch x := x.Value.(type) {
		case array:
			r := make(AB, x.Len())
			return NewV(r)
		default:
			return NewV(B2I(false))
		}
	}
	if Length(x) == 0 {
		return NewV(AB{})
	}
	//assertCanonical(x)
	//assertCanonical(y)
	switch y := y.Value.(type) {
	case AB:
		return memberOfAB(x, y)
	case AF:
		return memberOfAF(x, y)
	case AI:
		return memberOfAI(x, y)
	case AS:
		return memberOfAS(x, y)
	case AV:
		return memberOfAV(x, y)
	default:
		return errf("x in y : y not an array (%s)", y.Type())
	}
}

func memberOfAB(x V, y AB) V {
	var t, f bool
	for _, yi := range y {
		if t && f {
			break
		}
		t, f = t || yi, f || !yi
	}
	if t && f {
		switch x := x.Value.(type) {
		case array:
			r := make(AB, x.Len())
			for i := range r {
				r[i] = true
			}
			return NewV(r)
		default:
			return NewV(B2I(true))
		}
	}
	if t {
		return equal(x, NewV(B2I(true)))
	}
	return equal(x, NewV(B2I(false)))
}

func memberOfAF(x V, y AF) V {
	m := map[F]struct{}{}
	for _, yi := range y {
		_, ok := m[F(yi)]
		if !ok {
			m[F(yi)] = struct{}{}
			continue
		}
	}
	switch x := x.Value.(type) {
	case I:
		_, ok := m[F(x)]
		return NewV(B2I(ok))
	case F:
		_, ok := m[x]
		return NewV(B2I(ok))
	case AB:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[B2F(xi)]
		}
		return NewV(r)
	case AI:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[F(xi)]
		}
		return NewV(r)
	case AF:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[F(xi)]
		}
		return NewV(r)
	case array:
		return memberOfArray(x, y)
	default:
		return NewV(I(0))
	}
}

func memberOfAI(x V, y AI) V {
	m := map[int]struct{}{}
	for _, yi := range y {
		_, ok := m[yi]
		if !ok {
			m[yi] = struct{}{}
			continue
		}
	}
	switch x := x.Value.(type) {
	case I:
		_, ok := m[int(x)]
		return NewV(B2I(ok))
	case F:
		if !isI(x) {
			return NewV(B2I(false))
		}
		_, ok := m[int(x)]
		return NewV(B2I(ok))
	case AB:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[int(B2I(xi))]
		}
		return NewV(r)
	case AI:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[xi]
		}
		return NewV(r)
	case AF:
		r := make(AB, x.Len())
		for i, xi := range x {
			if !isI(F(xi)) {
				continue
			}
			_, r[i] = m[int(xi)]
		}
		return NewV(r)
	case array:
		return memberOfArray(x, y)
	default:
		return NewV(I(0))
	}
}

func memberOfAS(x V, y AS) V {
	m := map[string]struct{}{}
	for _, yi := range y {
		_, ok := m[yi]
		if !ok {
			m[yi] = struct{}{}
			continue
		}
	}
	switch x := x.Value.(type) {
	case S:
		_, ok := m[string(x)]
		return NewV(B2I(ok))
	case AS:
		r := make(AB, x.Len())
		for i, xi := range x {
			_, r[i] = m[xi]
		}
		return NewV(r)
	case array:
		return memberOfArray(x, y)
	default:
		return NewV(I(0))
	}
}

func memberOfAV(x V, y AV) V {
	switch xv := x.Value.(type) {
	case array:
		return memberOfArray(xv, y)
	default:
		for _, yi := range y {
			if Match(x, yi) {
				return NewV(B2I(true))
			}
		}
		return NewV(B2I(false))
	}
}

func memberOfArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make(AB, x.Len())
	for i := 0; i < x.Len(); i++ {
		for j := 0; j < y.Len(); j++ {
			if Match(x.at(i), y.at(j)) {
				r[i] = true
				break
			}
		}
	}
	return NewV(r)
}

// OccurrenceCount returns ⊒x.
func occurrenceCount(x V) V {
	if Length(x) == 0 {
		return NewV(AB{})
	}
	//assertCanonical(x)
	switch xv := x.Value.(type) {
	case F, I, S:
		return errf("⊒x : x not an array (%s)", xv.Type())
	case AB:
		r := make(AI, xv.Len())
		var f, t int
		for i, xi := range xv {
			if xi {
				r[i] = t
				t++
				continue
			}
			r[i] = f
			f++
		}
		return NewV(r)
	case AF:
		r := make(AI, xv.Len())
		m := map[float64]int{}
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewV(r)
	case AI:
		r := make(AI, xv.Len())
		m := map[int]int{}
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewV(r)
	case AS:
		r := make(AI, xv.Len())
		m := map[string]int{}
		for i, xi := range xv {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewV(r)
	case AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make(AI, xv.Len())
	loop:
		for i, xi := range xv {
			for j := i - 1; j >= 0; j-- {
				if Match(xi, xv[j]) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return NewV(r)
	default:
		return errf("⊒x : x not an array (%s)", x.Type())
	}
}

// without returns x^y.
func without(x, y V) V {
	switch xv := x.Value.(type) {
	case I:
		return windows(int(xv), y)
	case F:
		if !isI(xv) {
			return errf("i^y : i non-integer (%g)", xv)
		}
		return windows(int(xv), y)
	case S:
		return trim(xv, y)
	case array:
		y = toArray(y)
		r := memberOf(y, x)
		switch bres := r.Value.(type) {
		case I:
			r = NewV(I(1 - bres))
		case AB:
			for i, b := range bres {
				bres[i] = !b
			}
		}
		r = replicate(r, y)
		return r
	default:
		return errType("x^y", "x", xv)
	}
}

// find returns x?y.
func find(x, y V) V {
	//assertCanonical(y)
	//assertCanonical(x)
	switch x := x.Value.(type) {
	case S:
		return findS(x, y)
	case AB:
		return findAB(x, y)
	case AF:
		return findAF(x, y)
	case AI:
		return findAI(x, y)
	case AS:
		return findAS(x, y)
	case AV:
		return findAV(x, y)
	default:
		return errf("x?y : x not an array (%s)", x.Type())
	}
}

func findS(s S, y V) V {
	switch y := y.Value.(type) {
	case S:
		return NewV(I(strings.Index(string(s), string(y))))
	case AS:
		r := make(AI, y.Len())
		for i, ss := range y {
			r[i] = strings.Index(string(s), string(ss))
		}
		return NewV(r)
	case AV:
		r := make(AV, y.Len())
		for i, yi := range y {
			r[i] = findS(s, yi)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return NewV(r)
	default:
		return errType("s?y", "y", y)
	}
}

func imapAB(x AB) (m [2]int) {
	m[0] = x.Len()
	m[1] = x.Len()
	if x.Len() == 0 {
		return m
	}
	m[int(B2I(x[0]))] = 0
	for i, xi := range x[1:] {
		if xi != x[0] {
			m[int(B2I(xi))] = i + 1
			break
		}
	}
	return m
}

func imapAI(x AI) map[int]int {
	m := map[int]int{}
	for i, xi := range x {
		_, ok := m[xi]
		if !ok {
			m[xi] = i
			continue
		}
	}
	return m
}

func imapAF(x AF) map[float64]int {
	m := map[float64]int{}
	for i, xi := range x {
		_, ok := m[xi]
		if !ok {
			m[xi] = i
			continue
		}
	}
	return m
}

func imapAS(x AS) map[string]int {
	m := map[string]int{}
	for i, xi := range x {
		_, ok := m[xi]
		if !ok {
			m[xi] = i
			continue
		}
	}
	return m
}

func findAB(x AB, y V) V {
	switch y := y.Value.(type) {
	case I:
		for i, xi := range x {
			if B2I(xi) == y {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case F:
		if !isI(y) {
			return NewV(I(x.Len()))
		}
		return findAB(x, NewV(I(y)))
	case AB:
		m := imapAB(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			r[i] = m[B2I(yi)]
		}
		return NewV(r)
	case AI:
		m := imapAB(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			if yi != 0 && yi != 1 {
				r[i] = x.Len()
			} else {
				r[i] = m[yi]
			}
		}
		return NewV(r)
	case AF:
		m := imapAB(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			if yi != 0 && yi != 1 {
				r[i] = x.Len()
			} else {
				r[i] = m[int(yi)]
			}
		}
		return NewV(r)
	case array:
		// TODO: findArray may be redundant (canonical values)
		return findArray(x, y)
	default:
		return NewV(I(x.Len()))
	}
}

func findAF(x AF, y V) V {
	switch y := y.Value.(type) {
	case I:
		for i, xi := range x {
			if xi == float64(y) {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case F:
		for i, xi := range x {
			if F(xi) == y {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case AB:
		m := imapAF(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[float64(B2F(yi))]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case AI:
		m := imapAF(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[float64(yi)]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case AF:
		m := imapAF(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[yi]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case array:
		return findArray(x, y)
	default:
		return NewV(I(x.Len()))
	}
}

func findAI(x AI, y V) V {
	switch y := y.Value.(type) {
	case I:
		for i, xi := range x {
			if I(xi) == y {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case F:
		for i, xi := range x {
			if F(xi) == y {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case AB:
		m := imapAI(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[int(B2I(yi))]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case AI:
		m := imapAI(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[yi]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case AF:
		m := imapAI(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			if !isI(F(yi)) {
				r[i] = x.Len()
				continue
			}
			j, ok := m[int(yi)]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case array:
		return findArray(x, y)
	default:
		return NewV(I(x.Len()))
	}
}

func findAS(x AS, y V) V {
	switch y := y.Value.(type) {
	case S:
		for i, xi := range x {
			if S(xi) == y {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	case AS:
		m := imapAS(x)
		r := make(AI, y.Len())
		for i, yi := range y {
			j, ok := m[yi]
			if ok {
				r[i] = j
			} else {
				r[i] = x.Len()
			}
		}
		return NewV(r)
	case array:
		return findArray(x, y)
	default:
		return NewV(I(x.Len()))
	}
}

func findArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make(AI, y.Len())
	for i := range r {
		r[i] = x.Len()
	}
	for i := 0; i < y.Len(); i++ {
		for j := 0; j < x.Len(); j++ {
			if Match(y.at(i), x.at(j)) {
				r[i] = j
				break
			}
		}
	}
	return NewV(r)
}

func findAV(x AV, y V) V {
	switch yv := y.Value.(type) {
	case array:
		return findArray(x, yv)
	default:
		for i, xi := range x {
			if Match(y, xi) {
				return NewV(I(i))
			}
		}
		return NewV(I(x.Len()))
	}
}
