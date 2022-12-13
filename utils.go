package goal

import (
	"fmt"
	"math"
	//"reflect"
)

// b2i converts a boolean to an integer.
func b2i(b bool) (i int64) {
	if b {
		i = 1
	}
	return
}

// b2i converts a boolean to a float.
func b2f(b bool) (f float64) {
	if b {
		f = 1
	}
	return
}

// divideF divides two floats, returning infinity with appropriate sign when
// dividing by zero.
func divideF(x, y float64) float64 {
	if y == 0 {
		return math.Inf(signF(x))
	}
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

// clone creates an identical deep copy of a value, or the value itself if it
// is reusable.
func clone(x V) V {
	x = cloneShallow(x)
	if xv, ok := x.value.(*AV); ok {
		for i, xi := range xv.Slice {
			xv.Slice[i] = clone(xi)
		}
	}
	return x
}

// cloneShallow creates an identical shallow copy of a value, or the value
// itself if it is reusable.
func cloneShallow(x V) V {
	if xv, ok := x.value.(array); ok {
		x.value = cloneShallowArray(xv)
	}
	return x
}

// clone creates an identical shallow copy of an array, or the value itself if
// it is reusable.
func cloneShallowArray(x array) array {
	if x.reusable() {
		return x
	}
	switch xv := x.(type) {
	case *AB:
		r := &AB{Slice: make([]bool, xv.Len())}
		copy(r.Slice, xv.Slice)
		return r
	case *AI:
		r := &AI{Slice: make([]int64, xv.Len())}
		copy(r.Slice, xv.Slice)
		return r
	case *AF:
		r := &AF{Slice: make([]float64, xv.Len())}
		copy(r.Slice, xv.Slice)
		return r
	case *AS:
		r := &AS{Slice: make([]string, xv.Len())}
		copy(r.Slice, xv.Slice)
		return r
	case *AV:
		r := &AV{Slice: make([]V, xv.Len())}
		copy(r.Slice, xv.Slice)
		return r
	default:
		// should not happen
		panic("cloneShallowArray: x not a clonable array")
	}
}

// isIndices returns true if we have indices in canonical form, that is,
// using types I, AI and AV of thoses.
func isIndices(x V) bool {
	if x.IsI() {
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
	return toIndicesRec(x)
}

func toIndicesRec(x V) V {
	//assertCanonical(x)
	if x.IsI() {
		return x
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("non-integer index (%g)", x.F())
		}
		return NewI(int64(x.F()))
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
		return Canonical(NewAV(r))
	default:
		return panics("not an indices array")
	}
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	if x.IsI() {
		switch x.I() {
		case 0, 1:
			return NewAB([]bool{x.I() == 1})
		default:
			return NewAI([]int64{x.I()})
		}
	}
	if x.IsF() {
		return NewAF([]float64{float64(x.F())})
	}
	switch xv := x.value.(type) {
	case S:
		return NewAS([]string{string(xv)})
	case array:
		return x
	default:
		return NewAV([]V{x})
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
	return NewAI(r)
}

// toAF converts AI into AF.
func toAF(x *AI) V {
	r := make([]float64, x.Len())
	for i, xi := range x.Slice {
		r[i] = float64(xi)
	}
	return NewAF(r)
}

// fromABtoAF converts AB into AF.
func fromABtoAF(x *AB) V {
	r := make([]float64, x.Len())
	for i, xi := range x.Slice {
		r[i] = float64(b2i(xi))
	}
	return NewAF(r)
}

// fromABtoAI converts AB into AI (for simplifying code, used only for
// unfrequent code).
func fromABtoAI(x *AB) V {
	r := make([]int64, x.Len())
	for i := range r {
		r[i] = b2i(x.At(i))
	}
	return NewAI(r)
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

// mergeArrayTypes returns the most specialized type that can represent
// both array types, or tV if any of them is not an array type. For example,
// merging 1 2 and 2 3 gives tAI, but merging 4 and 2 3 gives tV.
func mergeArrayTypes(t, s vType) vType {
	if t&tAV == s&tAV {
		return t & s
	}
	return tV
}

// mergeEltTypes returns the most specialized type that can represent both
// types elements (for arrays) or themselves (for atoms). For example, merging
// 4 and 2 3 gives tI. It is identical to mergeArrayTypes if both are array
// types.
func mergeEltTypes(t, s vType) vType {
	return t & s
}

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
	for i := 1; i < x.Len(); i++ {
		t = mergeArrayTypes(t, getType(x.At(i)))
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
	for i := 1; i < x.Len(); i++ {
		t = mergeEltTypes(t, getType(x.At(i)))
	}
	return t
}

//// sameType returns true if two (non-Panic) values have same type.
//func sameType(x, y V) bool {
//return x.kind != valBoxed && x.kind == y.kind ||
//reflect.TypeOf(x.value) == reflect.TypeOf(y.value)
//}

// isEltType returns true if the type of y is compatible with the type of x
// elements.
func isEltType(x array, y V) bool {
	switch x.(type) {
	case *AI:
		return y.IsI()
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
	// NOTE: We assume no NaN or Inf: handling those special cases is left
	// to the program.
	return math.Floor(float64(x)) == float64(x)
}

func isBI(x int64) bool {
	return x == 0 || x == 1
}

func isBF(x float64) bool {
	return x == 0 || x == 1
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

func minMaxB(x *AB) (int64, int64) {
	if x.Len() == 0 {
		return 0, 0
	}
	min := true
	max := false
	for _, xi := range x.Slice {
		max, min = max || xi, min && !xi
		if max && !min {
			break
		}
	}
	return b2i(min), b2i(max)
}

func maxAB(x AB) bool {
	for _, xi := range x.Slice {
		if xi {
			return true
		}
	}
	return false
}

func isCanonicalV(x V) bool {
	switch xv := x.value.(type) {
	case *AV:
		_, ok := isCanonical(xv)
		return ok
	default:
		return true
	}
}

// isCanonical returns true if the given generic array is in canonical form,
// that is, it uses the most specialized representation.
func isCanonical(x *AV) (vType, bool) {
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		return t, false
	case tV:
		for _, xi := range x.Slice {
			if isCanonicalV(xi) {
				return t, false
			}
		}
		return t, true
	default:
		return t, true
	}
}

func assertCanonical(x V) {
	switch xv := x.value.(type) {
	case *AV:
		_, ok := isCanonical(xv)
		if !ok {
			panic(fmt.Sprintf("not canonical: %#v", x))
		}
	}
}

// normalize returns a canonical form of an AV array. It returns true if a
// shallow clone was made.
func normalize(x *AV) (array, bool) {
	t := aType(x)
	switch t {
	case tB:
		r := make([]bool, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I() != 0
		}
		return &AB{Slice: r}, true
	case tI:
		r := make([]int64, x.Len())
		for i, xi := range x.Slice {
			r[i] = xi.I()
		}
		return &AI{Slice: r}, true
	case tF:
		r := make([]float64, x.Len())
		for i, xi := range x.Slice {
			if xi.IsI() {
				r[i] = float64(xi.I())
			} else {
				r[i] = float64(xi.F())
			}
		}
		return &AF{Slice: r}, true
	case tS:
		r := make([]string, x.Len())
		for i, xi := range x.Slice {
			r[i] = string(xi.value.(S))
		}
		return &AS{Slice: r}, true
	case tV:
		for i, xi := range x.Slice {
			x.Slice[i] = Canonical(xi)
		}
		return x, false
	default:
		// should not happen
		return x, false
	}
}

// canonicalAV returns the canonicalAV form of a given generic array.
func canonicalAV(x *AV) Value {
	r, _ := normalize(x)
	return r
}

// Canonical returns the canonical form of a given value, that is the most
// specialized form. In practice, if the value is a generic array, but a more
// specialized version could represent the value, it returns the specialized
// value. All goal variadic functions have to return results in canonical
// form, so this function can be used to ensure that.
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

func sumAB(x *AB) int64 {
	n := int64(0)
	for _, xi := range x.Slice {
		if xi {
			n++
		}
	}
	return n
}
