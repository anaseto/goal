package goal

import (
	"fmt"
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
	switch x := x.BV.(type) {
	case I:
		n = x
	case F:
		n = I(x)
	}
	// x is assumed to be a number.
	return n
}

func isNum(x V) bool {
	switch x.BV.(type) {
	case I, F:
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
	switch x := x.BV.(type) {
	case F:
		return newBV(x)
	case I:
		return newBV(x)
	case S:
		return newBV(x)
	case AB:
		r := make(AB, x.Len())
		copy(r, x)
		return newBV(r)
	case AF:
		r := make(AF, x.Len())
		copy(r, x)
		return newBV(r)
	case AI:
		r := make(AI, x.Len())
		copy(r, x)
		return newBV(r)
	case AS:
		r := make(AS, x.Len())
		copy(r, x)
		return newBV(r)
	case AV:
		r := make(AV, x.Len())
		for i := range r {
			r[i] = clone(x[i])
		}
		return newBV(r)
	case errV:
		return newBV(x)
	default:
		return newBV(x)
	}
}

func cloneShallow(x V) V {
	switch xv := x.BV.(type) {
	case AB:
		r := make(AB, xv.Len())
		copy(r, xv)
		return newBV(r)
	case AF:
		r := make(AF, xv.Len())
		copy(r, xv)
		return newBV(r)
	case AI:
		r := make(AI, xv.Len())
		copy(r, xv)
		return newBV(r)
	case AS:
		r := make(AS, xv.Len())
		copy(r, xv)
		return newBV(r)
	case AV:
		r := make(AV, xv.Len())
		copy(r, xv)
		return newBV(r)
	default:
		return x
	}
}

func cloneShallowArray(x array) array {
	switch xv := x.(type) {
	case AB:
		r := make(AB, xv.Len())
		copy(r, xv)
		return r
	case AF:
		r := make(AF, xv.Len())
		copy(r, xv)
		return r
	case AI:
		r := make(AI, xv.Len())
		copy(r, xv)
		return r
	case AS:
		r := make(AS, xv.Len())
		copy(r, xv)
		return r
	case AV:
		r := make(AV, xv.Len())
		copy(r, xv)
		return r
	default:
		return x
	}
}

// isIndices returns true if we have indices in canonical form, that is,
// using types I, AI and AV of thoses.
func isIndices(x V) bool {
	switch xv := x.BV.(type) {
	case I:
		return true
	case AI:
		return true
	case AV:
		for _, xi := range xv {
			if !isIndices(xi) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func toIndices(x V) V {
	if isIndices(x) {
		return x
	}
	return toIndicesRec(x)
}

func toIndicesRec(x V) V {
	//assertCanonical(x)
	switch xv := x.BV.(type) {
	case F:
		if !isI(xv) {
			return errf("non-integer index (%g)", xv)
		}
		return newBV(I(xv))
	case I:
		return newBV(xv)
	case AB:
		return fromABtoAI(xv)
	case AF:
		return toAI(xv)
	case AV:
		r := make(AV, xv.Len())
		for i, z := range xv {
			r[i] = toIndicesRec(z)
			if isErr(r[i]) {
				return r[i]
			}
		}
		return newBV(canonical(r))
	default:
		return errs("not an indices array")
	}
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	switch xv := x.BV.(type) {
	case F:
		return newBV(AF{float64(xv)})
	case I:
		if xv == 0 || xv == 1 {
			return newBV(AB{xv == 1})
		}
		return newBV(AI{int(xv)})
	case S:
		return newBV(AS{string(xv)})
	case array:
		return newBV(xv)
	default:
		return newBV(AV{x})
	}
}

// toAI converts AF into AI if possible.
func toAI(x AF) V {
	r := make(AI, x.Len())
	for i, xi := range x {
		if !isI(F(xi)) {
			return errf("contains non-integer (%g)", xi)
		}
		r[i] = int(xi)
	}
	return newBV(r)
}

// fromABtoAI converts AB into AI (for simplifying code, used only for
// unfrequent code).
func fromABtoAI(x AB) V {
	r := make(AI, x.Len())
	for i := range r {
		r[i] = int(B2I(x[i]))
	}
	return newBV(r)
}

func isFalse(x V) bool {
	switch xv := x.BV.(type) {
	case F:
		return xv == 0
	case I:
		return xv == 0
	case S:
		return xv == ""
	default:
		return x == V{} || Length(x) == 0
	}
}

func isTrue(x V) bool {
	switch xv := x.BV.(type) {
	case F:
		return xv != 0
	case I:
		return xv != 0
	case S:
		return xv != ""
	default:
		return x != V{} && Length(x) > 0
	}
}

// eltype represents distinct kinds of elements for specialized
// arrays.
type eltype int32

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
	switch x := x.BV.(type) {
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
	switch x := x.BV.(type) {
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

func sameType(x, y V) bool {
	switch x.BV.(type) {
	case I:
		_, ok := y.BV.(I)
		return ok
	case F:
		_, ok := y.BV.(F)
		return ok
	case AB:
		_, ok := y.BV.(AB)
		return ok
	case AI:
		_, ok := y.BV.(AI)
		return ok
	case AF:
		_, ok := y.BV.(AF)
		return ok
	case AS:
		_, ok := y.BV.(AS)
		return ok
	case AV:
		_, ok := y.BV.(AV)
		return ok
	default:
		// TODO: sameType, handle other cases (unused for now)
		return false
	}
}

func compatEltType(x array, y V) bool {
	switch x.(type) {
	case AI:
		_, ok := y.BV.(I)
		return ok
	case AF:
		_, ok := y.BV.(F)
		return ok
	case AS:
		_, ok := y.BV.(S)
		return ok
	case AV:
		return true
	default:
		return false
	}
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
	for _, xi := range x[1:] {
		switch {
		case xi > max:
			max = xi
		case xi < min:
			min = xi
		}
	}
	return
}

func maxAI(x AI) int {
	max := math.MinInt
	if len(x) == 0 {
		return max
	}
	for _, xi := range x {
		if xi > max {
			max = xi
		}
	}
	return max
}

func minMaxB(x AB) (I, I) {
	if len(x) == 0 {
		return 0, 0
	}
	min := true
	max := false
	for _, xi := range x {
		max, min = max || xi, min && !xi
		if max && !min {
			break
		}
	}
	return B2I(min), B2I(max)
}

func maxAB(x AB) bool {
	for _, xi := range x {
		if xi {
			return true
		}
	}
	return false
}

func isCanonicalV(x V) bool {
	switch xv := x.BV.(type) {
	case AV:
		_, ok := isCanonical(xv)
		return ok
	default:
		return true
	}
}

// isCanonical returns true if the array is in canonical form, that is, it uses
// the most specialized representation. For example AV{I(2), I(3)} is not
// canonical, but AI{2, 3} is.
func isCanonical(x AV) (eltype, bool) {
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		return t, false
	case tV:
		for _, xi := range x {
			if isCanonicalV(xi) {
				return t, false
			}
		}
		return t, true
	default:
		return t, true
	}
}

func assertCanonical(x AV) {
	_, ok := isCanonical(x)
	if !ok {
		panic(fmt.Sprintf("not canonical: %#v", x))
	}
}

// normalize returns a canonical form of an AV array.
func normalize(x AV, t eltype) Value {
	switch t {
	case tB:
		r := make(AB, x.Len())
		for i, xi := range x {
			r[i] = xi.BV.(I) != 0
		}
		return r
	case tI:
		r := make(AI, x.Len())
		for i, xi := range x {
			r[i] = int(xi.BV.(I))
		}
		return r
	case tF:
		r := make(AF, x.Len())
		for i, xi := range x {
			switch xi := xi.BV.(type) {
			case F:
				r[i] = float64(xi)
			case I:
				r[i] = float64(xi)
			}
		}
		return r
	case tS:
		r := make(AS, x.Len())
		for i, xi := range x {
			r[i] = string(xi.BV.(S))
		}
		return r
	case tV:
		for i, xi := range x {
			x[i] = canonicalV(xi)
		}
		return x
	default:
		// should not happen
		return x
	}
}

// canonical returns the canonical form of a given value.
func canonicalV(x V) V {
	switch xv := x.BV.(type) {
	case AV:
		t, ok := isCanonical(xv)
		if ok {
			return x
		}
		return newBV(normalize(xv, t))
	default:
		return x
	}
}

// canonical returns the canonical form of a given value.
func canonical(x AV) Value {
	t, ok := isCanonical(x)
	if ok {
		return x
	}
	return normalize(x, t)
}

// canonical returns the canonical form of a given value.
func canonicalArray(x array) Value {
	if xv, ok := x.(AV); ok {
		return canonical(xv)
	}
	return x
}

// toCanonical returns the canonical form of a given value, and false if it was
// already canonical.
func toCanonical(x AV) (Value, bool) {
	t, ok := isCanonical(x)
	if ok {
		return x, false
	}
	return normalize(x, t), true
}

// hasNil returns true if there is a nil value in the given array.
func hasNil(a []V) bool {
	for _, x := range a {
		if x == (V{}) {
			return true
		}
	}
	return false
}

// countNils returns the number of nil values in the given array.
func countNils(a []V) int {
	n := 0
	for _, ai := range a {
		if ai == (V{}) {
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
