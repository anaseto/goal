package main

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
	case Array:
		return true
	default:
		return false
	}
}

func sign(x F) (sign int) {
	if x > 0 {
		sign = 1
	} else if x < 0 {
		sign = -1
	}
	return sign
}

func divide(w, x F) F {
	if x == 0 {
		return F(math.Inf(sign(w)))
	}
	return w / x
}

func modI(w, x I) I {
	if x == 0 {
		return x
	}
	return x % w
}

func modF(w, x F) F {
	if x == 0 {
		return x
	}
	return F(math.Mod(float64(x), float64(w)))
}

func minI(w, x I) I {
	if w < x {
		return w
	}
	return x
}

func maxI(w, x I) I {
	if w < x {
		return x
	}
	return w
}

func minS(w, x S) S {
	if w < x {
		return w
	}
	return x
}

func maxS(w, x S) S {
	if w < x {
		return x
	}
	return w
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
	case E:
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
	case E:
		return AV{x}
	default:
		return x
	}
}

func growArray(x V, n int) V {
	l := int(Length(x))
	if l >= n && n >= -l {
		return x
	}
	i := 0
	if n < 0 {
		i = l + n
		n = -n
		if i < 0 {
			i = 0
		}
	}
	switch x := x.(type) {
	case AB:
		r := make(AB, n)
		copy(r[i:], x)
		return r
	case AF:
		r := make(AF, n)
		copy(r[i:], x)
		return r
	case AI:
		r := make(AI, n)
		copy(r[i:], x)
		return r
	case AS:
		r := make(AS, n)
		copy(r[i:], x)
		return r
	case AV:
		r := make(AV, n)
		copy(r[i:], x)
		return r
	default:
		return x
	}
}

func isFalse(x V) bool {
	switch x := x.(type) {
	case F:
		return x == 0
	case I:
		return x == 0
	case S:
		return x == ""
	case AB:
		return len(x) == 0
	case AF:
		return len(x) == 0
	case AI:
		return len(x) == 0
	case AS:
		return len(x) == 0
	case AV:
		return len(x) == 0
	case E:
		return true
	default:
		// TODO: Interface for other objects?
		return false
	}
}

// eltype represents distinct kinds of elements for specialized
// arrays.
type eltype int

const (
	tO  eltype = 0b00000
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
	return tO
}

// eType returns the eltype of x.
func eType(x V) eltype {
	switch x.(type) {
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
		return tO
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
		return tO
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
		return tO
	}
	t := eType(x[0])
	for i := 1; i < len(x); i++ {
		t = mergeTypes(t, eType(x[i]))
	}
	return t
}

func isI(x F) bool {
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
		return
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
				r[i] = v.(I) == 1
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
				r[i] = float64(v.(F))
			}
			return r
		case tS:
			r := make(AS, len(y))
			for i, v := range y {
				r[i] = string(v.(S))
			}
			return r
		case tO:
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
