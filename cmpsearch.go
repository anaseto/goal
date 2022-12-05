package goal

import "strings"

// Matcher is implemented by types that can be matched againts other objects
// (typically a struct of the same type with fields that match).
type Matcher interface {
	Matches(x V) bool
}

// Match returns true if the two values match like in x~y.
func Match(x, y V) bool {
	switch x.Kind {
	case Nil:
		return y.Kind == Nil
	case Int:
		if y.Kind == Int {
			return x.N == y.N
		}
		yf, ok := y.Value.(F)
		return ok && F(x.N) == yf
	case Variadic:
		return y.Kind == Variadic && x.N == y.N
	case Lambda:
		return y.Kind == Lambda && x.N == y.N
	default:
		if y.Kind == Int {
			xf, ok := x.Value.(F)
			return ok && F(y.N) == xf
		}
		return y.Kind == Boxed && x.Value.Matches(y.Value)
	}
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
	switch xv := x.(type) {
	case *AB:
		switch yv := y.(type) {
		case *AB:
			return matchAB(xv, yv)
		case *AI:
			return matchABAI(xv, yv)
		case *AF:
			return matchABAF(xv, yv)
		}
	case *AI:
		switch yv := y.(type) {
		case *AB:
			return matchABAI(yv, xv)
		case *AI:
			return matchAI(xv, yv)
		case *AF:
			return matchAIAF(xv, yv)
		}
	case *AF:
		switch yv := y.(type) {
		case *AB:
			return matchABAF(yv, xv)
		case *AI:
			return matchAIAF(yv, xv)
		case *AF:
			return matchAF(xv, yv)
		}
	case *AS:
		yv, ok := y.(*AS)
		if !ok {
			break
		}
		for i, yi := range yv.Slice {
			if yi != xv.At(i) {
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

func matchAB(x, y *AB) bool {
	for i, yi := range y.Slice {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

func matchABAI(x *AB, y *AI) bool {
	for i, yi := range y.Slice {
		if yi != B2I(x.At(i)) {
			return false
		}
	}
	return true
}

func matchABAF(x *AB, y *AF) bool {
	for i, yi := range y.Slice {
		if F(yi) != B2F(x.At(i)) {
			return false
		}
	}
	return true
}

func matchAI(x, y *AI) bool {
	for i, yi := range y.Slice {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

func matchAIAF(x *AI, y *AF) bool {
	for i, yi := range y.Slice {
		if F(yi) != F(x.At(i)) {
			return false
		}
	}
	return true
}

func matchAF(x, y *AF) bool {
	for i, yi := range y.Slice {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

// classify returns %x.
func classify(x V) V {
	if Length(x) == 0 {
		return NewAB([]bool{})
	}
	//assertCanonical(x)
	switch xv := x.Value.(type) {
	case *AB:
		if !xv.At(0) {
			return NewV(xv)
		}
		return not(x)
	case *AF:
		r := make([]int64, xv.Len())
		m := map[float64]int64{}
		n := int64(0)
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewAI(r)
	case *AI:
		r := make([]int64, xv.Len())
		m := map[int64]int64{}
		n := int64(0)
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewAI(r)
	case *AS:
		r := make([]int64, xv.Len())
		m := map[string]int64{}
		n := int64(0)
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				r[i] = n
				m[xi] = n
				n++
				continue
			}
			r[i] = c
		}
		return NewAI(r)
	case *AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make([]int64, xv.Len())
		n := int64(0)
	loop:
		for i, xi := range xv.Slice {
			for j := range xv.Slice[:i] {
				if Match(xi, xv.At(j)) {
					r[i] = r[j]
					continue loop
				}
			}
			r[i] = n
			n++
		}
		return NewAI(r)
	default:
		return errf("%%x : x not an array (%s)", x.Type())
	}
}

// uniq returns ?x.
func uniq(x V) V {
	if Length(x) == 0 {
		return x
	}
	//assertCanonical(xv)
	switch xv := x.Value.(type) {
	case *AB:
		if xv.Len() == 0 {
			return NewV(xv)
		}
		b := xv.At(0)
		for i := 1; i < xv.Len(); i++ {
			if xv.At(i) != b {
				return NewAB([]bool{b, xv.At(i)})
			}
		}
		return NewAB([]bool{b})
	case *AF:
		r := []float64{}
		m := map[float64]struct{}{}
		for _, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
			}
		}
		return NewAF(r)
	case *AI:
		r := []int64{}
		m := map[int64]struct{}{}
		for _, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
				continue
			}
		}
		return NewAI(r)
	case *AS:
		r := []string{}
		m := map[string]struct{}{}
		for _, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r = append(r, xi)
				m[xi] = struct{}{}
				continue
			}
		}
		return NewAS(r)
	case *AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make([]V, xv.Len())
	loop:
		for i, xi := range xv.Slice {
			for j := range xv.Slice[:i] {
				if Match(xi, xv.At(j)) {
					continue loop
				}
			}
			r = append(r, xi)
		}
		return canonicalV(NewAV(r))
	default:
		// NOTE: ?atom could be used for something.
		return errf("?x : x not an array (%s)", x.Type())
	}
}

// Mark Firsts returns ∊x. XXX unused for now
func markFirsts(x V) V {
	if Length(x) == 0 {
		return NewAB([]bool{})
	}
	//assertCanonical(xv)
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		r[0] = true
		x0 := xv.At(0)
		for i := 1; i < xv.Len(); i++ {
			if xv.At(i) != x0 {
				r[i] = true
				break
			}
		}
		return NewAB(r)
	case *AF:
		r := make([]bool, xv.Len())
		m := map[float64]struct{}{}
		for i, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewAB(r)
	case *AI:
		r := make([]bool, xv.Len())
		m := map[int64]struct{}{}
		for i, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewAB(r)
	case *AS:
		r := make([]bool, xv.Len())
		m := map[string]struct{}{}
		for i, xi := range xv.Slice {
			_, ok := m[xi]
			if !ok {
				r[i] = true
				m[xi] = struct{}{}
				continue
			}
		}
		return NewAB(r)
	case *AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make([]bool, xv.Len())
	loop:
		for i, xi := range xv.Slice {
			for j := range xv.Slice[:i] {
				if Match(xi, xv.At(j)) {
					continue loop
				}
			}
			r[i] = true
		}
		return NewAB(r)
	default:
		return errf("∊x : x not an array (%s)", x.Type())
	}
}

// memberOf returns x in y.
func memberOf(x, y V) V {
	if Length(y) == 0 {
		switch xv := x.Value.(type) {
		case array:
			r := make([]bool, xv.Len())
			return NewAB(r)
		default:
			return NewI(B2I(false))
		}
	}
	if Length(x) == 0 {
		return NewAB([]bool{})
	}
	//assertCanonical(x)
	//assertCanonical(yv)
	switch yv := y.Value.(type) {
	case *AB:
		return memberOfAB(x, yv)
	case *AF:
		return memberOfAF(x, yv)
	case *AI:
		return memberOfAI(x, yv)
	case *AS:
		return memberOfAS(x, yv)
	case *AV:
		return memberOfAV(x, yv)
	default:
		return errf("x in y : y not an array (%s)", y.Type())
	}
}

func memberOfAB(x V, y *AB) V {
	var t, f bool
	for _, yi := range y.Slice {
		if t && f {
			break
		}
		t, f = t || yi, f || !yi
	}
	if t && f {
		switch xv := x.Value.(type) {
		case array:
			r := make([]bool, xv.Len())
			for i := range r {
				r[i] = true
			}
			return NewAB(r)
		default:
			return NewI(B2I(true))
		}
	}
	if t {
		return equal(x, NewI(B2I(true)))
	}
	return equal(x, NewI(B2I(false)))
}

func memberOfAF(x V, y *AF) V {
	m := map[F]struct{}{}
	for _, yi := range y.Slice {
		_, ok := m[F(yi)]
		if !ok {
			m[F(yi)] = struct{}{}
			continue
		}
	}
	if x.IsInt() {
		_, ok := m[F(x.I())]
		return NewI(B2I(ok))
	}
	switch xv := x.Value.(type) {
	case F:
		_, ok := m[xv]
		return NewI(B2I(ok))
	case *AB:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[B2F(xi)]
		}
		return NewAB(r)
	case *AI:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[F(xi)]
		}
		return NewAB(r)
	case *AF:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[F(xi)]
		}
		return NewAB(r)
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfAI(x V, y *AI) V {
	m := map[int64]struct{}{}
	for _, yi := range y.Slice {
		_, ok := m[yi]
		if !ok {
			m[yi] = struct{}{}
			continue
		}
	}
	if x.IsInt() {
		_, ok := m[x.I()]
		return NewI(B2I(ok))
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return NewI(B2I(false))
		}
		_, ok := m[int64(xv)]
		return NewI(B2I(ok))
	case *AB:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[B2I(xi)]
		}
		return NewAB(r)
	case *AI:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[xi]
		}
		return NewAB(r)
	case *AF:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			if !isI(F(xi)) {
				continue
			}
			_, r[i] = m[int64(xi)]
		}
		return NewAB(r)
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfAS(x V, y *AS) V {
	m := map[string]struct{}{}
	for _, yi := range y.Slice {
		_, ok := m[yi]
		if !ok {
			m[yi] = struct{}{}
			continue
		}
	}
	switch xv := x.Value.(type) {
	case S:
		_, ok := m[string(xv)]
		return NewI(B2I(ok))
	case *AS:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			_, r[i] = m[xi]
		}
		return NewAB(r)
	case array:
		return memberOfArray(xv, y)
	default:
		return NewI(0)
	}
}

func memberOfAV(x V, y *AV) V {
	switch xv := x.Value.(type) {
	case array:
		return memberOfArray(xv, y)
	default:
		for _, yi := range y.Slice {
			if Match(x, yi) {
				return NewI(B2I(true))
			}
		}
		return NewI(B2I(false))
	}
}

func memberOfArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make([]bool, x.Len())
	for i := 0; i < x.Len(); i++ {
		for j := 0; j < y.Len(); j++ {
			if Match(x.at(i), y.at(j)) {
				r[i] = true
				break
			}
		}
	}
	return NewAB(r)
}

// OccurrenceCount returns ⊒x.
func occurrenceCount(x V) V {
	if Length(x) == 0 {
		return NewAB([]bool{})
	}
	//assertCanonical(x)
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]int64, xv.Len())
		var f, t int64
		for i, xi := range xv.Slice {
			if xi {
				r[i] = t
				t++
				continue
			}
			r[i] = f
			f++
		}
		return NewAI(r)
	case *AF:
		r := make([]int64, xv.Len())
		m := map[float64]int64{}
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewAI(r)
	case *AI:
		r := make([]int64, xv.Len())
		m := map[int64]int64{}
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewAI(r)
	case *AS:
		r := make([]int64, xv.Len())
		m := map[string]int64{}
		for i, xi := range xv.Slice {
			c, ok := m[xi]
			if !ok {
				m[xi] = 0
				continue
			}
			m[xi] = c + 1
			r[i] = c + 1
		}
		return NewAI(r)
	case *AV:
		// NOTE: quadratic algorithm, worst case complexity could be
		// improved by sorting or string hashing.
		r := make([]int64, xv.Len())
	loop:
		for i, xi := range xv.Slice {
			for j := i - 1; j >= 0; j-- {
				if Match(xi, xv.At(j)) {
					r[i] = r[j] + 1
					continue loop
				}
			}
		}
		return NewAI(r)
	default:
		return errf("⊒x : x not an array (%s)", x.Type())
	}
}

// without returns x^y.
func without(x, y V) V {
	if x.IsInt() {
		return windows(x.I(), y)
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("i^y : i non-integer (%g)", xv)
		}
		return windows(int64(xv), y)
	case S:
		return trim(xv, y)
	case array:
		y = toArray(y)
		r := memberOf(y, x)
		if r.IsInt() {
			r = NewI(1 - r.I())
		}
		switch bres := r.Value.(type) {
		case *AB:
			for i, b := range bres.Slice {
				bres.Slice[i] = !b
			}
		}
		r = replicate(r, y)
		return r
	default:
		return errType("x^y", "x", x)
	}
}

// find returns x?y.
func find(x, y V) V {
	//assertCanonical(y)
	//assertCanonical(xv)
	switch xv := x.Value.(type) {
	case S:
		return findS(xv, y)
	case *AB:
		return findAB(xv, y)
	case *AF:
		return findAF(xv, y)
	case *AI:
		return findAI(xv, y)
	case *AS:
		return findAS(xv, y)
	case *AV:
		return findAV(xv, y)
	default:
		return errf("x?y : x not an array (%s)", x.Type())
	}
}

func findS(s S, y V) V {
	switch yv := y.Value.(type) {
	case S:
		return NewI(int64(strings.Index(string(s), string(yv))))
	case *AS:
		r := make([]int64, yv.Len())
		for i, ss := range yv.Slice {
			r[i] = int64(strings.Index(string(s), string(ss)))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = findS(s, yi)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return errType("s?y", "y", y)
	}
}

func imapAB(x *AB) (m [2]int64) {
	m[0] = int64(x.Len())
	m[1] = int64(x.Len())
	if x.Len() == 0 {
		return m
	}
	m[B2I(x.At(0))] = 0
	for i, xi := range x.Slice[1:] {
		if xi != x.At(0) {
			m[B2I(xi)] = int64(i) + 1
			break
		}
	}
	return m
}

func imapAI(x *AI) map[int64]int64 {
	m := map[int64]int64{}
	for i, xi := range x.Slice {
		_, ok := m[xi]
		if !ok {
			m[xi] = int64(i)
			continue
		}
	}
	return m
}

func imapAF(x *AF) map[float64]int64 {
	m := map[float64]int64{}
	for i, xi := range x.Slice {
		_, ok := m[xi]
		if !ok {
			m[xi] = int64(i)
			continue
		}
	}
	return m
}

func imapAS(x *AS) map[string]int {
	m := map[string]int{}
	for i, xi := range x.Slice {
		_, ok := m[xi]
		if !ok {
			m[xi] = i
			continue
		}
	}
	return m
}

func findAB(x *AB, y V) V {
	if y.IsInt() {
		for i, xi := range x.Slice {
			if B2I(xi) == y.I() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
	switch yv := y.Value.(type) {
	case F:
		if !isI(yv) {
			return NewI(int64(x.Len()))
		}
		return findAB(x, NewI(int64(yv)))
	case *AB:
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = m[B2I(yi)]
		}
		return NewAI(r)
	case *AI:
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi != 0 && yi != 1 {
				r[i] = int64(x.Len())
			} else {
				r[i] = m[yi]
			}
		}
		return NewAI(r)
	case *AF:
		m := imapAB(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if yi != 0 && yi != 1 {
				r[i] = int64(x.Len())
			} else {
				r[i] = m[int(yi)]
			}
		}
		return NewAI(r)
	case array:
		return findArray(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findAF(x *AF, y V) V {
	if y.IsInt() {
		for i, xi := range x.Slice {
			if xi == float64(y.I()) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	switch yv := y.Value.(type) {
	case F:
		for i, xi := range x.Slice {
			if F(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AB:
		m := imapAF(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[float64(B2F(yi))]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case *AI:
		m := imapAF(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[float64(yi)]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case *AF:
		m := imapAF(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[yi]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case array:
		return findArray(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findAI(x *AI, y V) V {
	if y.IsInt() {
		for i, xi := range x.Slice {
			if xi == y.I() {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))

	}
	switch yv := y.Value.(type) {
	case F:
		for i, xi := range x.Slice {
			if F(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AB:
		m := imapAI(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[B2I(yi)]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case *AI:
		m := imapAI(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[yi]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case *AF:
		m := imapAI(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			if !isI(F(yi)) {
				r[i] = int64(x.Len())
				continue
			}
			j, ok := m[int64(yi)]
			if ok {
				r[i] = j
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case array:
		return findArray(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findAS(x *AS, y V) V {
	switch yv := y.Value.(type) {
	case S:
		for i, xi := range x.Slice {
			if S(xi) == yv {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	case *AS:
		m := imapAS(x)
		r := make([]int64, yv.Len())
		for i, yi := range yv.Slice {
			j, ok := m[yi]
			if ok {
				r[i] = int64(j)
			} else {
				r[i] = int64(x.Len())
			}
		}
		return NewAI(r)
	case array:
		return findArray(x, yv)
	default:
		return NewI(int64(x.Len()))
	}
}

func findArray(x, y array) V {
	// NOTE: quadratic algorithm, worst case complexity could be
	// improved by sorting or string hashing.
	r := make([]int64, y.Len())
	for i := range r {
		r[i] = int64(x.Len())
	}
	for i := 0; i < y.Len(); i++ {
		for j := 0; j < x.Len(); j++ {
			if Match(y.at(i), x.at(j)) {
				r[i] = int64(j)
				break
			}
		}
	}
	return NewAI(r)
}

func findAV(x *AV, y V) V {
	switch yv := y.Value.(type) {
	case array:
		return findArray(x, yv)
	default:
		for i, xi := range x.Slice {
			if Match(y, xi) {
				return NewI(int64(i))
			}
		}
		return NewI(int64(x.Len()))
	}
}
