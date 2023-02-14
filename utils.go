package goal

import (
	"fmt"
	"math"
	"reflect"
)

// b2i converts a boolean to an integer.
func b2i(b bool) int64 {
	var i int64
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// b2i converts a boolean to a float.
func b2f(b bool) float64 {
	var f float64
	if b {
		f = 1
	} else {
		f = 0
	}
	return f
}

// divideF divides two floats, returning infinity with appropriate sign when
// dividing by zero.
func divideF(x, y float64) float64 {
	// NOTE: Go's standard says it could panic, but current implementation
	// seems to provide the desired behaviour.
	return x / y
}

// modI returns y % x or y if x is zero
func modI(x, y int64) int64 {
	if x == 0 {
		return y
	}
	return y % x
}

// modF returns y % x or y if x is zero
func modF(x, y float64) float64 {
	if x == 0 {
		return y
	}
	return math.Mod(float64(y), float64(x))
}

func minI(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func maxI(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func maxInt(x, y int) int {
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

func minMax(x *AI) (min, max int64) {
	if x.Len() == 0 {
		return
	}
	min = x.At(0)
	max = min
	for _, xi := range x.Slice[1:] {
		switch {
		case xi > max:
			max = xi
		case xi < min:
			min = xi
		}
	}
	return
}

func maxAI(x *AI) int64 {
	max := int64(math.MinInt64)
	if x.Len() == 0 {
		return max
	}
	for _, xi := range x.Slice {
		if xi > max {
			max = xi
		}
	}
	return max
}

func minAI(x *AI) int64 {
	min := int64(math.MaxInt64)
	if x.Len() == 0 {
		return min
	}
	for _, xi := range x.Slice {
		if xi < min {
			min = xi
		}
	}
	return min
}

func maxAF(x *AF) float64 {
	max := math.Inf(-1)
	if x.Len() == 0 {
		return max
	}
	for _, xi := range x.Slice {
		if xi > max {
			max = xi
		}
	}
	return max
}

func minAF(x *AF) float64 {
	min := math.Inf(1)
	if x.Len() == 0 {
		return min
	}
	for _, xi := range x.Slice {
		if xi < min {
			min = xi
		}
	}
	return min
}

func isStar(x V) bool {
	return x.kind == valVariadic && x.variadic() == vMultiply
}

// isIndices returns true if we have indices in canonical form, that is,
// using types I, AI and AV of thoses.
func isIndices(x V) bool {
	if x.IsI() {
		return true
	}
	if isStar(x) {
		return true
	}
	switch xv := x.value.(type) {
	case *AI:
		return true
	case *AV:
		for _, xi := range xv.Slice {
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
	return CanonicalRec(toIndicesRec(x))
}

func toIndicesRec(x V) V {
	if x.IsI() {
		return x
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("non-integer index (%g)", x.F())
		}
		return NewI(int64(x.F()))
	}
	if isStar(x) {
		return x
	}
	switch xv := x.value.(type) {
	case *AB:
		return fromABtoAI(xv)
	case *AF:
		return toAI(xv)
	case *AV:
		r := make([]V, xv.Len())
		for i, z := range xv.Slice {
			r[i] = toIndicesRec(z)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	case array:
		return panics("non-integer indices")
	default:
		return Panicf("non-integer index (type %s)", x.Type())
	}
}

func indicesInBounds(x *AI, l int) (int64, bool) {
	for _, xi := range x.Slice {
		if xi < 0 {
			xi += int64(l)
		}
		if xi < 0 || xi >= int64(l) {
			return xi, false
		}
	}
	return 0, true
}

func inBoundsInfo(x *AI, l int) (int64, int, bool) {
	for i, xi := range x.Slice {
		if xi < 0 {
			xi += int64(l)
		}
		if xi < 0 || xi >= int64(l) {
			return xi, i, false
		}
	}
	return 0, 0, true
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	if x.IsI() {
		var n int
		r := &AI{Slice: []int64{x.I()}, rc: &n}
		return NewV(r)
	}
	if x.IsF() {
		var n int
		r := &AF{Slice: []float64{float64(x.F())}, rc: &n}
		return NewV(r)
	}
	switch xv := x.value.(type) {
	case S:
		var n int
		r := &AS{Slice: []string{string(xv)}, rc: &n}
		return NewV(r)
	case array:
		return x
	default:
		var n int
		r := &AV{Slice: []V{x}, rc: &n}
		return NewV(r)
	}
}

// toAI converts AF into AI if possible.
func toAI(x *AF) V {
	r := make([]int64, x.Len())
	for i, xi := range x.Slice {
		if !isI(xi) {
			return Panicf("contains non-integer (%g)", xi)
		}
		r[i] = int64(xi)
	}
	return NewAIWithRC(r, reuseRCp(x.rc))
}

// toAF converts AI into AF.
func toAF(x *AI) V {
	r := make([]float64, x.Len())
	for i, xi := range x.Slice {
		r[i] = float64(xi)
	}
	return NewAFWithRC(r, reuseRCp(x.rc))
}

// fromABtoAF converts AB into AF.
func fromABtoAF(x *AB) V {
	r := make([]float64, x.Len())
	for i, xi := range x.Slice {
		r[i] = float64(b2i(xi))
	}
	return NewAFWithRC(r, reuseRCp(x.rc))
}

// fromABtoAI converts AB into AI (for simplifying code, used only for
// unfrequent code).
func fromABtoAI(x *AB) V {
	r := make([]int64, x.Len())
	for i := range r {
		r[i] = b2i(x.At(i))
	}
	return NewAIWithRC(r, reuseRCp(x.rc))
}

// isFalse returns true for false values.
func isFalse(x V) bool {
	if x.IsI() {
		return x.I() == 0
	}
	if x.IsF() {
		return x.F() == 0
	}
	switch xv := x.value.(type) {
	case S:
		return xv == ""
	case *errV:
		return true
	default:
		return Length(x) == 0
	}
}

// isTrue returns true for true values.
func isTrue(x V) bool {
	if x.IsI() {
		return x.I() != 0
	}
	if x.IsF() {
		return x.F() != 0
	}
	switch xv := x.value.(type) {
	case S:
		return xv != ""
	case *errV:
		return false
	default:
		return Length(x) > 0
	}
}

// vType represents information about value types.
type vType int32

const (
	tV  vType = 0b00000
	tB  vType = 0b00111
	tI  vType = 0b00011
	tF  vType = 0b00001
	tS  vType = 0b01000
	tAB vType = 0b10111
	tAI vType = 0b10011
	tAF vType = 0b10001
	tAS vType = 0b11000
	tAV vType = 0b10000
)

// getType returns the vType of x.
func getType(x V) vType {
	if x.IsI() {
		switch x.I() {
		case 0, 1:
			return tB
		default:
			return tI
		}
	}
	if x.IsF() {
		return tF
	}
	switch x.value.(type) {
	case S:
		return tS
	case *AB:
		return tAB
	case *AF:
		return tAF
	case *AI:
		return tAI
	case *AS:
		return tAS
	case *AV:
		return tAV
	default:
		return tV
	}
}

// aType returns the most specific type common to the elements of a generic
// array. For example, aType (1 2;2 3) returns tAI, but aType (1;2 3) returns
// tV.
func aType(x *AV) vType {
	if x.Len() == 0 {
		return tV
	}
	t := getType(x.Slice[0])
	for _, xi := range x.Slice[1:] {
		s := getType(xi)
		if t&tAV != s&tAV {
			return tV
		}
		t &= s
	}
	return t
}

// eType returns the most specific element type common to the the elements of a
// generic array. For example, eType (1;2 3) returns tI.
func eType(x *AV) vType {
	if x.Len() == 0 {
		return tV
	}
	t := getType(x.Slice[0])
	for _, xi := range x.Slice[1:] {
		t &= getType(xi)
		if t == tV {
			return tV
		}
	}
	return t
}

// sameType returns true if two arrays have same type.
func sameType(x, y array) bool {
	return reflect.TypeOf(x) == reflect.TypeOf(y)
}

// isEltType returns true if the type of y is compatible with the type of x
// elements.
func isEltType(x array, y V) bool {
	switch x.(type) {
	case *AB:
		return y.IsI() && (y.n == 0 || y.n == 1) || y.IsF() &&
			(y.F() == 0 || y.F() == 1)
	case *AI:
		return y.IsI() || y.IsF() && isI(y.F())
	case *AF:
		return y.IsF()
	case *AS:
		_, ok := y.value.(S)
		return ok
	case *AV:
		return true
	default:
		return false
	}
}

func isI(x float64) bool {
	return x == float64(int64(x))
}

func isBI(x int64) bool {
	return x == 0 || x == 1
}

func isBF(x float64) bool {
	return x == 0 || x == 1
}

func isCanonical(x V) bool {
	switch xv := x.value.(type) {
	case *AV:
		_, ok := isCanonicalAV(xv)
		return ok
	default:
		return true
	}
}

// isCanonicalAV returns true if the given generic array is in canonical form,
// that is, it uses the most specialized representation.
func isCanonicalAV(x *AV) (vType, bool) {
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		return t, false
	case tV, tAV:
		for _, xi := range x.Slice {
			if !isCanonical(xi) {
				return t, false
			}
		}
		return t, true
	default:
		return t, true
	}
}

func (ctx *Context) assertCanonical(x V) {
	switch xv := x.value.(type) {
	case *AV:
		_, ok := isCanonicalAV(xv)
		if !ok {
			panic(fmt.Sprintf("not canonical: %#v: %s", xv.Slice, x.Sprint(ctx)))
		}
	}
}

// normalizeRec returns a canonical form of an AV array. It returns true if a
// shallow clone was made.
func normalizeRec(x *AV) (array, bool) {
	t := aType(x)
	switch t {
	case tB:
		r := make([]bool, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I() != 0
		}
		return &AB{Slice: r, rc: x.rc}, true
	case tI:
		r := make([]int64, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I()
		}
		return &AI{Slice: r, rc: x.rc}, true
	case tF:
		r := make([]float64, x.Len())
		for i, xi := range x.Slice {
			if xi.IsI() {
				r[i] = float64(xi.I())
			} else {
				r[i] = float64(xi.F())
			}
		}
		return &AF{Slice: r, rc: x.rc}, true
	case tS:
		r := make([]string, x.Len())
		for i, xi := range x.Slice {
			r[i] = string(xi.value.(S))
		}
		return &AS{Slice: r, rc: x.rc}, true
	case tV, tAV:
		for i, xi := range x.Slice {
			x.Slice[i] = CanonicalRec(xi)
		}
		return x, false
	default:
		return x, false
	}
}

// normalize returns a canonical form of an AV array, assuming it's
// elements themselves are canonical. It returns true if a shallow clone was
// made.
func normalize(x *AV) (array, bool) {
	t := aType(x)
	switch t {
	case tB:
		r := make([]bool, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I() != 0
		}
		return &AB{Slice: r, rc: reuseRCp(x.rc)}, true
	case tI:
		r := make([]int64, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I()
		}
		return &AI{Slice: r, rc: reuseRCp(x.rc)}, true
	case tF:
		r := make([]float64, x.Len())
		for i, xi := range x.Slice {
			if xi.IsI() {
				r[i] = float64(xi.I())
			} else {
				r[i] = float64(xi.F())
			}
		}
		return &AF{Slice: r, rc: reuseRCp(x.rc)}, true
	case tS:
		r := make([]string, x.Len())
		for i, xi := range x.Slice {
			r[i] = string(xi.value.(S))
		}
		return &AS{Slice: r, rc: reuseRCp(x.rc)}, true
	default:
		return x, false
	}
}

// CanonicalRec returns the canonical form of a given value, that is the most
// specialized form. In practice, if the value is a generic array, but a more
// specialized version could represent the value, it returns the specialized
// value. All goal variadic functions have to return results in canonical
// form, so this function can be used to ensure that.
func CanonicalRec(x V) V {
	switch xv := x.value.(type) {
	case *AV:
		r, b := normalizeRec(xv)
		if b {
			x.value = r
		}
		return x
	default:
		return x
	}
}

// canonicalAV returns the canonical form of a given generic array.
func canonicalAV(x *AV) Value {
	r, _ := normalize(x)
	return r
}

// canonicalArray returns the canonical form of a given generic array.
func canonicalArray(x array) array {
	switch xv := x.(type) {
	case *AV:
		r, _ := normalize(xv)
		return r
	default:
		return x
	}
}

// Canonical returns the canonical form of a given value, that is the
// most specialized form, assuming it's already canonical at depth > 1. In
// practice, if the value is a generic array, but a more specialized version
// could represent the value, it returns the specialized value. All goal
// variadic functions have to return results in canonical form, so this
// function can be used to ensure that.
func Canonical(x V) V {
	switch xv := x.value.(type) {
	case *AV:
		r, b := normalize(xv)
		if b {
			x.value = r
		}
		return x
	default:
		return x
	}
}

// hasNil returns true if there is a nil value in the given array.
func hasNil(a []V) bool {
	for _, x := range a {
		if x.kind == valNil {
			return true
		}
	}
	return false
}

// countNils returns the number of nil values in the given array.
func countNils(a []V) int {
	n := 0
	for _, ai := range a {
		if ai.kind == valNil {
			n++
		}
	}
	return n
}

func cloneArgs(a []V) []V {
	args := make([]V, len(a))
	copy(args, a)
	return args
}

func sumAB(x *AB) int64 {
	n := int64(0)
	for _, xi := range x.Slice {
		if xi {
			n++
		}
	}
	return n
}
