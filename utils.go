package goal

import (
	"math"
)

func B2I(b bool) (i I) {
	if b {
		i = 1
	}
	return
}

func B2F(b bool) (f F) {
	if b {
		f = 1
	}
	return
}

func num2I(x V) (n I) {
	switch x := x.(type) {
	case I:
		n = x
	case F:
		n = I(x)
	}
	// x is assumed to be a number.
	return n
}

func isNum(x V) bool {
	switch x.(type) {
	case I, F:
		return true
	default:
		return false
	}
}

func isArray(x V) bool {
	switch x.(type) {
	case array:
		return true
	default:
		return false
	}
}

func divideF(x, y F) F {
	if y == 0 {
		return F(math.Inf(int(signF(x))))
	}
	return x / y
}

func modI(x, y I) I {
	if y == 0 {
		return y
	}
	return y % x
}

func modF(x, y F) F {
	if y == 0 {
		return y
	}
	return F(math.Mod(float64(y), float64(x)))
}

func minI(x, y I) I {
	if x < y {
		return x
	}
	return y
}

func maxI(x, y I) I {
	if x < y {
		return y
	}
	return x
}

func minS(x, y S) S {
	if x < y {
		return x
	}
	return y
}

func maxS(x, y S) S {
	if x < y {
		return y
	}
	return x
}

func clone(x V) V {
	switch x := x.(type) {
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
	case AV:
		r := make(AV, len(x))
		for i := range r {
			r[i] = clone(x[i])
		}
		return r
	case errV:
		return x
	default:
		return x
	}
}

func cloneShallow(x V) V {
	switch x := x.(type) {
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
	case AV:
		r := make(AV, len(x))
		copy(r, x)
		return r
	default:
		return x
	}
}

func toIndices(x V, l int) V {
	x = canonical(x)
	switch x := x.(type) {
	case AB:
		res := make(AI, len(x))
		for i := range res {
			res[i] = int(B2I(x[i]))
		}
		return res
	case AF:
		res := make(AI, len(x))
		for i := range res {
			if !isI(F(x[i])) {
				return errs("x[y] : non-integer indices array")
			}
			res[i] = int(x[i])
		}
		return res
	case AI:
		return x
	default:
		return errs("x[y] : not an indices array")
	}
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	switch x := x.(type) {
	case F:
		return AF{float64(x)}
	case I:
		if x == 0 || x == 1 {
			return AB{x == 1}
		}
		return AI{int(x)}
	case S:
		return AS{string(x)}
	case errV:
		return AV{x}
	case array:
		return x
	default:
		return AV{x}
	}
}

// toAI converts AF into AI if possible.
func toAI(x AF) V {
	for _, v := range x {
		if !isI(F(v)) {
			return errf("contains non-integer (%g)", v)
		}
	}
	r := make(AI, len(x))
	for i := range r {
		r[i] = int(x[i])
	}
	return r
}

// fromABtoAI converts AB into AI (for simplifying code, used only for
// unfrequent code).
func fromABtoAI(x AB) V {
	r := make(AI, len(x))
	for i := range r {
		r[i] = int(B2I(x[i]))
	}
	return r
}

func isFalse(x V) bool {
	switch x := x.(type) {
	case F:
		return x == 0
	case I:
		return x == 0
	case S:
		return x == ""
	default:
		return x == nil || x.Len() == 0
	}
}

func isTrue(x V) bool {
	switch x := x.(type) {
	case F:
		return x != 0
	case I:
		return x != 0
	case S:
		return x != ""
	default:
		return x != nil && x.Len() > 0
	}
}

// eltype represents distinct kinds of elements for specialized
// arrays.
type eltype int

const (
	tV  eltype = 0b00000
	tB  eltype = 0b00111
	tF  eltype = 0b00001
	tI  eltype = 0b00011
	tS  eltype = 0b01000
	tAB eltype = 0b10111
	tAF eltype = 0b10001
	tAI eltype = 0b10011
	tAS eltype = 0b11000
	tAO eltype = 0b10000
)

func mergeTypes(t, s eltype) eltype {
	if t&tAO == s&tAO {
		return t & s
	}
	return tV
}

// eType returns the eltype of x.
func eType(x V) eltype {
	switch x := x.(type) {
	case F:
		return tF
	case I:
		if x == 0 || x == 1 {
			return tB
		}
		return tI
	case S:
		return tS
	case AB:
		return tAB
	case AF:
		return tAF
	case AI:
		return tAI
	case AS:
		return tAS
	case AV:
		return tAO
	default:
		return tV
	}
}

// cType returns the canonical eltype of x. XXX: unused.
func cType(x V) eltype {
	switch x := x.(type) {
	case AB:
		return tAB
	case F:
		return tF
	case AF:
		return tAF
	case I:
		if x == 0 || x == 1 {
			return tB
		}
		return tI
	case AI:
		return tAI
	case S:
		return tS
	case AS:
		return tAS
	case AV:
		return cTypeAO(x)
	default:
		return tV
	}
}

func cTypeAO(x AV) eltype {
	if x.Len() == 0 {
		return tAO
	}
	t := eType(x[0])
	for i := 1; i < len(x); i++ {
		t = mergeTypes(t, eType(x[i]))
	}
	switch t {
	case tB:
		return tAB
	case tF:
		return tAF
	case tI:
		return tAI
	case tS:
		return tAS
	default:
		return tAO
	}
}

// aType returns the most specific eltype of the elements of a generic array.
func aType(x AV) eltype {
	if x.Len() == 0 {
		return tV
	}
	t := eType(x[0])
	for i := 1; i < len(x); i++ {
		t = mergeTypes(t, eType(x[i]))
	}
	return t
}

func isI(x F) bool {
	// NOTE: We assume no NaN or Inf: handling those special cases is left
	// to the program.
	return math.Floor(float64(x)) == float64(x)
}

func isBI(x I) bool {
	return x == 0 || x == 1
}

func isBF(x F) bool {
	return x == 0 || x == 1
}

func minMax(x AI) (min, max int) {
	if len(x) == 0 {
		return
	}
	min = x[0]
	max = min
	for _, v := range x[1:] {
		switch {
		case v > max:
			max = v
		case v < min:
			min = v
		}
	}
	return
}

func minMaxB(x AB) (I, I) {
	if len(x) == 0 {
		return 0, 0
	}
	min := true
	max := false
	for _, v := range x {
		max, min = max || v, min && !v
		if max && !min {
			break
		}
	}
	return B2I(min), B2I(max)
}

func canonical(x V) V {
	switch y := x.(type) {
	case AV:
		t := aType(y)
		switch t {
		case tB:
			r := make(AB, len(y))
			for i, v := range y {
				r[i] = v.(I) != 0
			}
			return r
		case tI:
			r := make(AI, len(y))
			for i, v := range y {
				r[i] = int(v.(I))
			}
			return r
		case tF:
			r := make(AF, len(y))
			for i, v := range y {
				switch v := v.(type) {
				case F:
					r[i] = float64(v)
				case I:
					r[i] = float64(v)
				}
			}
			return r
		case tS:
			r := make(AS, len(y))
			for i, v := range y {
				r[i] = string(v.(S))
			}
			return r
		case tV:
			for i, v := range y {
				y[i] = canonical(v)
			}
			return y
		default:
			return x
		}
	default:
		return x
	}
}

func hasNil(a []V) bool {
	for _, x := range a {
		if x == nil {
			return true
		}
	}
	return false
}

func countNils(a []V) int {
	n := 0
	for _, v := range a {
		if v == nil {
			n++
		}
	}
	return n
}

func reverseArgs(a []V) {
	for i := 0; i < len(a)/2; i++ {
		a[i], a[len(a)-i-1] = a[len(a)-i-1], a[i]
	}
}

func cloneArgs(a []V) []V {
	args := make([]V, len(a))
	copy(args, a)
	return args
}
