package goal

import (
	"fmt"
	"math"
	"reflect"
)

// b2I converts a boolean to a 64-bit integer.
func b2I(b bool) int64 {
	var i int64
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// b2B converts a boolean to a byte.
func b2B(b bool) byte {
	var i byte
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// b2F converts a boolean to a float.
func b2F(b bool) float64 {
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

// modB returns y % x or y if x is zero
func modB(x, y byte) byte {
	return y % x
}

// modI returns y % x or y if x is zero
func modI(x, y int64) int64 {
	y = y % x
	if y < 0 {
		y += x
	}
	return y
}

// modF returns y % x or y if x is zero
func modF(x, y float64) float64 {
	y = math.Mod(float64(y), float64(x))
	if y < 0 {
		y += x
	}
	return y
}

func divI(x, y int64) int64 {
	if y >= 0 {
		return y / x
	}
	if y%x == 0 {
		return y / x
	}
	return (y / x) - 1
}

func divF(x, y float64) float64 {
	return math.Floor(y / x)
}

func minI(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func minB(x, y byte) byte {
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

func maxB(x, y byte) byte {
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

func minMaxAI(x *AI) (min, max int64) {
	return minMaxIs(x.elts)
}

func minMaxIs[I integer](x []I) (min, max I) {
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
	switch xv := x.bv.(type) {
	case *AB:
		return true
	case *AI:
		return true
	case *AV:
		for _, xi := range xv.elts {
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
	switch xv := x.bv.(type) {
	case *AF:
		return toAI(xv)
	case *AV:
		return cmapAV(xv, func(xi V) V { return toIndicesRec(xi) })
	case *AS:
		return Panicf("bad type \"%s\" as index", x.Type())
	default:
		return Panicf("bad type \"%s\" as index", x.Type())
	}
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	if x.IsI() {
		if isBI(x.I()) {
			r := &AB{elts: []byte{byte(x.I())}}
			if isbI(x.I()) {
				r.flags |= flagBool
			}
			return NewV(r)
		}
		r := &AI{elts: []int64{x.I()}}
		return NewV(r)
	}
	if x.IsF() {
		r := &AF{elts: []float64{float64(x.F())}}
		return NewV(r)
	}
	switch xv := x.bv.(type) {
	case S:
		r := &AS{elts: []string{string(xv)}}
		return NewV(r)
	case Array:
		return x
	default:
		x.MarkImmutable()
		r := &AV{elts: []V{x}}
		return NewV(r)
	}
}

// toAI converts AF into AI if possible.
func toAI(x *AF) V {
	r := make([]int64, x.Len())
	for i, xi := range x.elts {
		if !isI(xi) {
			return Panicf("contains non-integer (%g)", xi)
		}
		r[i] = int64(xi)
	}
	return NewAI(r)
}

// castToAI casts AF into AI.
func castToAI(x *AF) V {
	r := make([]int64, x.Len())
	for i, xi := range x.elts {
		r[i] = int64(xi)
	}
	return NewAI(r)
}

// toAF converts AI into AF.
func toAF(x *AI) V {
	r := make([]float64, x.Len())
	for i, xi := range x.elts {
		r[i] = float64(xi)
	}
	return NewAF(r)
}

// fromABtoAF converts AB into AF.
func fromABtoAF(x *AB) V {
	r := make([]float64, x.Len())
	for i, xi := range x.elts {
		r[i] = float64(xi)
	}
	return NewAF(r)
}

// IsFalse returns true for false values, that is zero numbers, empty strings,
// zero-length values, and errors.
func (x V) IsFalse() bool {
	if x.IsI() {
		return x.I() == 0
	}
	if x.IsF() {
		return x.F() == 0
	}
	switch xv := x.bv.(type) {
	case S:
		return xv == ""
	case *errV:
		return true
	default:
		return x.Len() == 0
	}
}

// IsTrue returns true for true values, that is non-zero numbers, non-empty
// strings, and non-zero length values that are not errors.
func (x V) IsTrue() bool {
	if x.IsI() {
		return x.I() != 0
	}
	if x.IsF() {
		return x.F() != 0
	}
	switch xv := x.bv.(type) {
	case S:
		return xv != ""
	case *errV:
		return false
	default:
		return x.Len() > 0
	}
}

// vType represents information about value types.
type vType int32

const (
	tV  vType = 0b000000
	tb  vType = 0b001111
	tB  vType = 0b000111
	tI  vType = 0b000011
	tF  vType = 0b000001
	tS  vType = 0b010000
	tAb vType = 0b101111
	tAB vType = 0b100111
	tAI vType = 0b100011
	tAF vType = 0b100001
	tAS vType = 0b110000
	tAV vType = 0b100000
)

// getType returns the vType of x.
func getType(x V) vType {
	if x.IsI() {
		switch {
		case x.uv >= 0 && x.uv < 256:
			if x.uv < 2 {
				return tb
			}
			return tB
		default:
			return tI
		}
	}
	if x.IsF() {
		return tF
	}
	switch xv := x.bv.(type) {
	case S:
		return tS
	case *AB:
		if xv.IsBoolean() {
			return tAb
		}
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

// getAtomType returns the vType of x, returning tV for non-atoms or atoms that
// cannot be packed in an unboxed array.
func getAtomType(x V) vType {
	if x.IsI() {
		switch {
		case x.uv >= 0 && x.uv < 256:
			if x.uv < 2 {
				return tb
			}
			return tB
		default:
			return tI
		}
	}
	if x.IsF() {
		return tF
	}
	switch x.bv.(type) {
	case S:
		return tS
	default:
		return tV
	}
}

// getAtomTypeFast is like getAtomType, but returns tI for tB. It's meant to be
// used in cases where we need to quickly get a type for canonical form, and tB
// is not an expected outcome.
func getAtomTypeFast(x V) vType {
	if x.IsI() {
		return tI
	}
	if x.IsF() {
		return tF
	}
	switch x.bv.(type) {
	case S:
		return tS
	default:
		return tV
	}
}

// aType returns the most specific atom type common to the all elements, or tV
// for a generic array. The returned value can be used to normalize a generic
// array.
func aType(x *AV) vType {
	if x.Len() == 0 {
		return tV
	}
	t := getAtomType(x.elts[0])
	if t == tV {
		return t
	}
	if t != tB && t != tb {
		for _, xi := range x.elts[1:] {
			t &= getAtomTypeFast(xi)
			if t == tV {
				return tV
			}
		}
		return t
	}
	for _, xi := range x.elts[1:] {
		t &= getAtomType(xi)
		if t == tV {
			return tV
		}
	}
	return t
}

// eType returns the most specific element type common to the the elements of a
// generic array. For example, eType (1;2 3) returns tI, obtained from merging
// tI and tAI.
func eType(x *AV) vType {
	if x.Len() == 0 {
		return tV
	}
	t := getType(x.elts[0])
	for _, xi := range x.elts[1:] {
		t &= getType(xi)
		if t == tV {
			return tV
		}
	}
	return t
}

// sameType returns true if two arrays have same type.
func sameType(x, y Array) bool {
	return reflect.TypeOf(x) == reflect.TypeOf(y)
}

func isI(x float64) bool {
	return x == float64(int64(x))
}

func isBI(x int64) bool {
	return x >= 0 && x < 256
}

func isbI(x int64) bool {
	return x == 0 || x == 1
}

func isCanonical(x V) bool {
	switch xv := x.bv.(type) {
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
	case tb, tB, tI, tF, tS:
		return t, false
	case tV, tAV:
		for _, xi := range x.elts {
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
	switch xv := x.bv.(type) {
	case *AV:
		_, ok := isCanonicalAV(xv)
		if !ok {
			panic(fmt.Sprintf("not canonical: %#v: %s", xv.elts, x.Sprint(ctx)))
		}
	}
}

// normalize returns a canonical form of an AV array, assuming it's
// elements themselves are canonical. It returns true if a shallow clone was
// made, in other words, if the returned array is not generic.
func normalize(x *AV, t vType) (Array, bool) {
	switch t {
	case tb, tB:
		r := make([]byte, x.Len())
		for i, xi := range x.elts {
			r[i] = byte(xi.I())
		}
		fl := x.flags
		if t == tb {
			fl |= flagBool
		}
		return &AB{elts: r, flags: fl &^ flagImmutable}, true
	case tI:
		r := make([]int64, x.Len())
		for i, xi := range x.elts {
			r[i] = xi.I()
		}
		return &AI{elts: r, flags: x.flags &^ flagImmutable}, true
	case tF:
		r := make([]float64, x.Len())
		for i, xi := range x.elts {
			if xi.IsI() {
				r[i] = float64(xi.I())
			} else {
				r[i] = float64(xi.F())
			}
		}
		return &AF{elts: r, flags: x.flags &^ flagImmutable}, true
	case tS:
		r := make([]string, x.Len())
		for i, xi := range x.elts {
			r[i] = string(xi.bv.(S))
		}
		return &AS{elts: r, flags: x.flags &^ flagImmutable}, true
	default:
		return x, false
	}
}

// canonicalRec returns the canonical form of a given value, that is the most
// specialized form. In practice, if the value is a generic array, but a more
// specialized version could represent the value, it returns the specialized
// value. All variadic functions have to return results in canonical form, so
// this function can be used to ensure that when defining new ones.
func canonicalRec(x V) V {
	switch xv := x.bv.(type) {
	case *AV:
		t := aType(xv)
		switch t {
		case tV, tAV:
			for i, xi := range xv.elts {
				cxi := canonicalRec(xi)
				cxi.MarkImmutable()
				xv.elts[i] = cxi
			}
			return x
		default:
			r, b := normalize(xv, t)
			if b {
				x.bv = r
			}
			return x
		}
	default:
		return x
	}
}

// canonicalArrayAV returns the canonical form of a given generic array.
func canonicalArrayAV(x *AV) Array {
	r, _ := normalize(x, aType(x))
	return r
}

// canonicalAV returns the canonical form of a given generic array.
func canonicalAV(x *AV) V {
	r, _ := normalize(x, aType(x))
	return NewV(r)
}

// canonicalAV returns the canonical form of a given generic array.
func canonicalAVImmut(x *AV) V {
	r, _ := normalize(x, aType(x))
	r.setFlags(x.flags)
	return NewV(r)
}

// canonicalArray returns the canonical form of a given generic array.
func canonicalArray(x Array) Array {
	switch xv := x.(type) {
	case *AV:
		r, _ := normalize(xv, aType(xv))
		return r
	default:
		return x
	}
}

func canonicalVs(r []V) V {
	x := &AV{elts: r}
	ra, ok := normalize(x, aType(x))
	if !ok {
		newAVu(r)
	}
	return NewV(ra)
}

func canonicalArrayVs(r []V) Array {
	x := &AV{elts: r}
	ra, _ := normalize(x, aType(x))
	return ra
}

// Canonical returns the canonical form of a given value, that is the
// most specialized form, assuming it's already canonical at depth > 1. In
// practice, if the value is a generic array, but a more specialized version
// could represent the value, it returns the specialized value. All variadic
// functions have to return results in canonical form, so this function can be
// used to ensure that when defining new ones.
func Canonical(x V) V {
	switch xv := x.bv.(type) {
	case *AV:
		r, b := normalize(xv, aType(xv))
		if b {
			x.bv = r
		}
		return x
	default:
		return x
	}
}

// canonicalImmut is like Canonical, but preserve immutable flag too.
func canonicalImmut(x V) V {
	switch xv := x.bv.(type) {
	case *AV:
		r, b := normalize(xv, aType(xv))
		if b {
			r.setFlags(xv.flags)
			x.bv = r
		}
		return x
	default:
		return x
	}
}

func protoAV() V {
	return NewV(&AV{flags: flagImmutable})
}

func proto(x []V) V {
	if len(x) == 0 {
		return protoAV()
	}
	return protoV(x[0])
}

func protoV(x V) V {
	if x.IsI() {
		return NewI(0)
	}
	if x.IsF() {
		return NewF(0)
	}
	switch xv := x.bv.(type) {
	case S:
		return NewS("")
	case *AB:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AI:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AF:
		return NewV(&AF{flags: flagImmutable})
	case *AS:
		return NewV(&AS{flags: flagImmutable})
	case *AV:
		return protoAV()
	case *D:
		return NewD(protoArray(xv.keys), protoArray(xv.values))
	default:
		return newVariadic(vRight)
	}
}

func protoArray(x Array) V {
	switch x.(type) {
	case *AB:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AI:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AF:
		return NewV(&AF{flags: flagImmutable})
	case *AS:
		return NewV(&AS{flags: flagImmutable})
	case *AV:
		return protoAV()
	default:
		panic("protoArray")
	}
}

func arrayProtoV(x Array) V {
	switch xv := x.(type) {
	case *AB:
		return NewI(0)
	case *AI:
		return NewI(0)
	case *AF:
		return NewF(0)
	case *AS:
		return NewS("")
	case *AV:
		return proto(xv.elts)
	default:
		panic("protoArray")
	}
}

func protoArrayForV(x V) V {
	if x.IsI() {
		return newABb(nil)
	}
	if x.IsF() {
		return NewAF(nil)
	}
	switch x.bv.(type) {
	case S:
		return NewV(&AS{flags: flagImmutable})
	case *AB:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AI:
		return NewV(&AB{flags: flagBool | flagImmutable})
	case *AF:
		return NewV(&AF{flags: flagImmutable})
	case *AS:
		return NewV(&AS{flags: flagImmutable})
	default:
		return protoAV()
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

// maxIndices returns the maximum index, assuming V is an array of indices.
func maxIndices(x V) int64 {
	switch xv := x.bv.(type) {
	case *AB:
		if xv.Len() == 0 {
			return math.MinInt64
		}
		return int64(maxBytes(xv.elts))
	case *AI:
		return maxIs(xv.elts)
	default:
		panic("maxIndices")
	}
}

func maxBytes(x []byte) byte {
	var max byte
	for _, xi := range x {
		if xi > max {
			max = xi
		}
	}
	return max
}

// numeric returns true for atomic and flat array numeric values.
func (x V) numeric() bool {
	if x.IsI() {
		return true
	}
	if x.IsF() {
		return true
	}
	switch xv := x.bv.(type) {
	case Array:
		return xv.numeric()
	default:
		return false
	}
}

func mapAV(x *AV, f func(V) V) V {
	r := x.reuse()
	for i, xi := range x.elts {
		ri := f(xi)
		if ri.IsPanic() {
			return ri
		}
		ri.MarkImmutable()
		r.elts[i] = ri
	}
	return NewV(r)
}

func imapAV(x *AV, f func(int, V) V) V {
	r := x.reuse()
	for i, xi := range x.elts {
		ri := f(i, xi)
		if ri.IsPanic() {
			return ri
		}
		ri.MarkImmutable()
		r.elts[i] = ri
	}
	return NewV(r)
}

func cmapAV(x *AV, f func(V) V) V {
	r := x.reuse()
	for i, xi := range x.elts {
		ri := f(xi)
		if ri.IsPanic() {
			return ri
		}
		ri.MarkImmutable()
		r.elts[i] = ri
	}
	return canonicalAV(r)
}

func cdoN(n int, f func(int) V) V {
	r := make([]V, n)
	for i := 0; i < n; i++ {
		ri := f(i)
		if ri.IsPanic() {
			return ri
		}
		ri.MarkImmutable()
		r[i] = ri
	}
	return canonicalVs(r)
}

func doN(n int, f func(int) V) V {
	r := make([]V, n)
	for i := 0; i < n; i++ {
		ri := f(i)
		if ri.IsPanic() {
			return ri
		}
		ri.MarkImmutable()
		r[i] = ri
	}
	return newAVu(r)
}

func sumIs[I integer](x []I) int64 {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	return n
}

func isFlat(x []V) bool {
	for i, xi := range x {
		if xi.kind != valBoxed {
			continue
		}
		switch xi.bv.(type) {
		case *D:
			if i == 0 {
				return false
			}
		case Array:
			return false
		}
	}
	return true
}

func ascFirstLast[T ordered](x []T) bool {
	if len(x) < 2 {
		return true
	}
	return x[0] < x[len(x)-1]
}
