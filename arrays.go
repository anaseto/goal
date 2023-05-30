package goal

import "sort"

type countable interface {
	Len() int
}

// array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type array interface {
	RefCounter
	countable
	sort.Interface
	at(i int) V           // x[i]
	slice(i, j int) array // x[i:j]
	getFlags() flags      // get the array's flags
	setFlags(flags)       // set the array's flags
	set(i int, y V)       // puts y at indix i, assuming compatibility
	vAtInts(y *AI) V      // x[y] (like in goal code)
	vAtBytes(y *AB) V     // x[y] (like in goal code)
	atInts([]int64) array // like x[y] but assumes valid positive indices
	atBytes([]byte) array // like x[y] but assumes valid positive indices
	sclone() array        // shallow clone, erases flags
	numeric() bool        // flat numeric array
	canSet(y V) bool      // compatible type for set
	reusable() bool       // reusable array
}

type flags uint32

const (
	flagNone      flags = 0b0000
	flagImmutable flags = 0b0001
	flagAscending flags = 0b0010
	flagDistinct  flags = 0b0100
	flagBool      flags = 0b1000
)

func (f flags) Has(ff flags) bool {
	return f&ff != 0
}

// A is a generic type used to represent arrays. Only specific instantiations
// implement the Value interface.
type A[T any] struct {
	flags flags
	rc    int32
	elts  []T
}

func newA[T any](x []T) A[T] {
	return A[T]{elts: x}
}

// AB represents an array of bytes. From Goal's perspective, it's the same as
// AI. It's used as an optimization to save space for small-integers, in
// particular for arrays of booleans (0s and 1s).
type AB A[byte]

// NewAB returns a new byte array.
func NewAB(x []byte) V {
	return NewV(&AB{elts: x})
}

// newABb returns a new byte array with boolean flag.
func newABb(x []byte) V {
	return NewV(&AB{elts: x, flags: flagBool})
}

// IsBoolean returns true when the array of bytes is known to contain only 1s
// and 0s.
func (x *AB) IsBoolean() bool {
	return x.flags.Has(flagBool)
}

// Slice returns the underlying immutable slice of values. It should not be
// modified.
func (x *AB) Slice() []byte {
	return x.elts
}

// AI represents an array of integers.
type AI A[int64]

// NewAI returns a new int array.
func NewAI(x []int64) V {
	return NewV(&AI{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified.
func (x *AI) Slice() []int64 {
	return x.elts
}

// AF represents an array of reals.
type AF A[float64]

// NewAF returns a new array of reals.
func NewAF(x []float64) V {
	return NewV(&AF{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified.
func (x *AF) Slice() []float64 {
	return x.elts
}

// AS represents an array of strings.
type AS A[string]

// NewAS returns a new array of strings.
func NewAS(x []string) V {
	return NewV(&AS{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified.
func (x *AS) Slice() []string {
	return x.elts
}

// AV represents a generic array. The elements of a generic array are marked as
// immutable, and they should not be representable together in a specialized
// array. In other words, it should be the canonical form of the array.
type AV A[V]

// NewAV returns a new array from a slice of generic values. The result value
// will be an array in canonical form.
func NewAV(x []V) V {
	xav := &AV{elts: x}
	xv, cloned := normalize(xav)
	if cloned {
		r := NewV(xv)
		return r
	}
	for _, x := range xav.elts {
		x.MarkImmutable()
	}
	return NewV(xav)
}

// newAVu returns a new generic array. It does not mark its elements as
// immutable, assuming they already were marked as such.
func newAVu(x []V) V {
	return NewV(&AV{elts: x})
}

// newAV returns a new generic array.
func newAV(x []V) *AV {
	for _, xi := range x {
		xi.MarkImmutable()
	}
	return &AV{elts: x}
}

// newAV returns a new generic array.
func newAVv(x []V) V {
	for _, xi := range x {
		xi.MarkImmutable()
	}
	return NewV(&AV{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified.
func (x *AV) Slice() []V {
	return x.elts
}

// Type returns the name of the value's type ("I").
func (x *AB) Type() string { return "I" }

// Type returns the name of the value's type ("I").
func (x *AI) Type() string { return "I" }

// Type returns the name of the value's type ("N").
func (x *AF) Type() string { return "N" }

// Type returns the name of the value's type ("S").
func (x *AS) Type() string { return "S" }

// Type returns the name of the value's type ("A").
func (x *AV) Type() string { return "A" }

// Len returns the length of the array.
func (x *AB) Len() int { return len(x.elts) }

// Len returns the length of the array.
func (x *AI) Len() int { return len(x.elts) }

// Len returns the length of the array.
func (x *AF) Len() int { return len(x.elts) }

// Len returns the length of the array.
func (x *AS) Len() int { return len(x.elts) }

// Len returns the length of the array.
func (x *AV) Len() int { return len(x.elts) }

func (x *AB) at(i int) V { return NewI(int64(x.elts[i])) }
func (x *AI) at(i int) V { return NewI(x.elts[i]) }
func (x *AF) at(i int) V { return NewF(x.elts[i]) }
func (x *AS) at(i int) V { return NewS(x.elts[i]) }
func (x *AV) at(i int) V { return x.elts[i] }

// At returns array value at the given index.
func (x *AB) At(i int) byte { return x.elts[i] }

// At returns array value at the given index.
func (x *AI) At(i int) int64 { return x.elts[i] }

// At returns array value at the given index.
func (x *AF) At(i int) float64 { return x.elts[i] }

// At returns array value at the given index.
func (x *AS) At(i int) string { return x.elts[i] }

// At returns array value at the given index.
func (x *AV) At(i int) V { return x.elts[i] }

func (x *AB) slice(i, j int) array {
	if !x.reusable() {
		x.flags |= flagImmutable
	}
	return &AB{flags: x.flags, elts: x.elts[i:j]}
}

func (x *AI) slice(i, j int) array {
	if !x.reusable() {
		x.flags |= flagImmutable
	}
	return &AI{flags: x.flags, elts: x.elts[i:j]}
}

func (x *AF) slice(i, j int) array {
	if !x.reusable() {
		x.flags |= flagImmutable
	}
	return &AF{flags: x.flags, elts: x.elts[i:j]}
}

func (x *AS) slice(i, j int) array {
	if !x.reusable() {
		x.flags |= flagImmutable
	}
	return &AS{flags: x.flags, elts: x.elts[i:j]}
}

func (x *AV) slice(i, j int) array {
	if !x.reusable() {
		x.flags |= flagImmutable
	}
	return canonicalArrayAV(&AV{flags: x.flags, elts: x.elts[i:j]})
}

func (x *AB) getFlags() flags { return x.flags }
func (x *AI) getFlags() flags { return x.flags }
func (x *AF) getFlags() flags { return x.flags }
func (x *AS) getFlags() flags { return x.flags }
func (x *AV) getFlags() flags { return x.flags }

func (x *AB) setFlags(f flags) { x.flags = f }
func (x *AI) setFlags(f flags) { x.flags = f }
func (x *AF) setFlags(f flags) { x.flags = f }
func (x *AS) setFlags(f flags) { x.flags = f }
func (x *AV) setFlags(f flags) { x.flags = f }

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AB) set(i int, y V) {
	x.elts[i] = byte(y.uv)
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AI) set(i int, y V) {
	x.elts[i] = y.uv
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AF) set(i int, y V) {
	x.elts[i] = y.F()
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AS) set(i int, y V) {
	x.elts[i] = string(y.bv.(S))
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	y.MarkImmutable()
	x.elts[i] = y
}

func selectNumsAtInts[N number](dst, x []N, y []int64) {
	xlen := int64(len(x))
	for i, yi := range y {
		if yi < 0 {
			yi += xlen
		}
		if yi >= 0 && yi < xlen {
			dst[i] = x[yi]
		} else {
			dst[i] = 0
		}
	}
}

func (x *AB) vAtInts(y *AI) V {
	r := &AB{elts: make([]byte, y.Len())}
	r.flags = x.flags & flagBool
	selectNumsAtInts(r.elts, x.elts, y.elts)
	return NewV(r)
}

func (x *AI) vAtInts(y *AI) V {
	r := make([]int64, y.Len())
	selectNumsAtInts(r, x.elts, y.elts)
	return NewAI(r)
}

func (x *AF) vAtInts(y *AI) V {
	r := make([]float64, y.Len())
	selectNumsAtInts(r, x.elts, y.elts)
	return NewAF(r)
}

func (x *AS) vAtInts(y *AI) V {
	r := make([]string, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi >= 0 && yi < xlen {
			r[i] = x.elts[yi]
		} else {
			r[i] = ""
		}
	}
	return NewAS(r)
}

func (x *AV) vAtInts(y *AI) V {
	r := make([]V, y.Len())
	xlen := int64(x.Len())
	var p V
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi >= 0 && yi < xlen {
			r[i] = x.elts[yi]
		} else {
			if p.kind == valNil {
				p = proto(x.elts)
			}
			r[i] = p
		}
	}
	return canonicalVs(r)
}

func selectNumsAtBytes[N number](dst, x []N, y []byte) {
	xlen := int(len(x))
	for i, yi := range y {
		if int(yi) < xlen {
			dst[i] = x[yi]
		} else {
			dst[i] = 0
		}
	}
}

func (x *AB) vAtBytes(y *AB) V {
	r := &AB{elts: make([]byte, y.Len())}
	r.flags = x.flags & flagBool
	selectNumsAtBytes(r.elts, x.elts, y.elts)
	return NewV(r)
}

func (x *AI) vAtBytes(y *AB) V {
	r := make([]int64, y.Len())
	selectNumsAtBytes(r, x.elts, y.elts)
	return NewAI(r)
}

func (x *AF) vAtBytes(y *AB) V {
	r := make([]float64, y.Len())
	selectNumsAtBytes(r, x.elts, y.elts)
	return NewAF(r)
}

func (x *AS) vAtBytes(y *AB) V {
	r := make([]string, y.Len())
	xlen := x.Len()
	for i, yi := range y.elts {
		if int(yi) < xlen {
			r[i] = x.elts[yi]
		} else {
			r[i] = ""
		}
	}
	return NewAS(r)
}

func (x *AV) vAtBytes(y *AB) V {
	r := make([]V, y.Len())
	xlen := x.Len()
	var p V
	for i, yi := range y.elts {
		if int(yi) < xlen {
			r[i] = x.elts[yi]
		} else {
			if p.kind == valNil {
				p = proto(x.elts)
			}
			r[i] = p
		}
	}
	return canonicalVs(r)
}

func (x *AB) atInts(y []int64) array {
	r := &AB{elts: make([]byte, len(y))}
	if x.IsBoolean() {
		r.flags |= flagBool
	}
	for i, yi := range y {
		r.elts[i] = x.elts[yi]
	}
	return r
}

func (x *AI) atInts(y []int64) array {
	r := make([]int64, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AI{elts: r}
}

func (x *AF) atInts(y []int64) array {
	r := make([]float64, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AF{elts: r}
}

func (x *AS) atInts(y []int64) array {
	r := make([]string, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AS{elts: r}
}

func (x *AV) atInts(y []int64) array {
	r := make([]V, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return canonicalArrayVs(r)
}

func (x *AB) atBytes(y []byte) array {
	r := &AB{elts: make([]byte, len(y))}
	if x.IsBoolean() {
		r.flags |= flagBool
	}
	for i, yi := range y {
		r.elts[i] = x.elts[yi]
	}
	return r
}

func (x *AI) atBytes(y []byte) array {
	r := make([]int64, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AI{elts: r}
}

func (x *AF) atBytes(y []byte) array {
	r := make([]float64, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AF{elts: r}
}

func (x *AS) atBytes(y []byte) array {
	r := make([]string, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return &AS{elts: r}
}

func (x *AV) atBytes(y []byte) array {
	r := make([]V, len(y))
	for i, yi := range y {
		r[i] = x.elts[yi]
	}
	return canonicalArrayVs(r)
}

func (x *AB) sclone() array {
	return (*AB)((*A[byte])(x).sclone())
}

func (x *AI) sclone() array {
	return (*AI)((*A[int64])(x).sclone())
}

func (x *AF) sclone() array {
	return (*AF)((*A[float64])(x).sclone())
}

func (x *AS) sclone() array {
	return (*AS)((*A[string])(x).sclone())
}

func (x *AV) sclone() array {
	return (*AV)((*A[V])(x).sclone())
}

func shallowCloneAB(x *AB) *AB {
	return (*AB)((*A[byte])(x).sclone())
}

func shallowCloneAI(x *AI) *AI {
	return (*AI)((*A[int64])(x).sclone())
}

// Matches returns true if the two values match like in x~y.
func (x *AB) Matches(y Value) bool {
	if !matchArrayLen(x, y) {
		return false
	}
	if x.Len() == 0 {
		return true
	}
	switch yv := y.(type) {
	case *AB:
		return matchAB(x, yv)
	case *AI:
		return matchABAI(x, yv)
	case *AF:
		return matchABAF(x, yv)
	default:
		return false
	}
}

// Matches returns true if the two values match like in x~y.
func (x *AI) Matches(y Value) bool {
	if !matchArrayLen(x, y) {
		return false
	}
	if x.Len() == 0 {
		return true
	}
	switch yv := y.(type) {
	case *AB:
		return matchABAI(yv, x)
	case *AI:
		return matchAI(x, yv)
	case *AF:
		return matchAIAF(x, yv)
	default:
		return false
	}
}

// Matches returns true if the two values match like in x~y.
func (x *AF) Matches(y Value) bool {
	if !matchArrayLen(x, y) {
		return false
	}
	if x.Len() == 0 {
		return true
	}
	switch yv := y.(type) {
	case *AB:
		return matchABAF(yv, x)
	case *AI:
		return matchAIAF(yv, x)
	case *AF:
		return matchAF(x, yv)
	default:
		return false
	}
}

// Matches returns true if the two values match like in x~y.
func (x *AS) Matches(y Value) bool {
	if !matchArrayLen(x, y) {
		return false
	}
	if x.Len() == 0 {
		return true
	}
	yv, ok := y.(*AS)
	if !ok {
		return false
	}
	for i, yi := range yv.elts {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

// Matches returns true if the two values match like in x~y.
func (x *AV) Matches(y Value) bool {
	if !matchArrayLen(x, y) {
		return false
	}
	if x.Len() == 0 {
		return true
	}
	yv, ok := y.(*AV)
	if !ok {
		return false
	}
	for i, yi := range yv.elts {
		if !yi.Matches(x.At(i)) {
			return false
		}
	}
	return true
}

// matchArrayLen returns true if y is an array of same length as x.
func matchArrayLen(x array, y Value) bool {
	ya, ok := y.(array)
	if !ok {
		return false
	}
	return x.Len() == ya.Len()
}

func matchAB(x, y *AB) bool {
	for i, yi := range y.elts {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

func matchABAI(x *AB, y *AI) bool {
	for i, yi := range y.elts {
		if yi != int64(x.At(i)) {
			return false
		}
	}
	return true
}

func matchABAF(x *AB, y *AF) bool {
	for i, yi := range y.elts {
		if yi != float64(x.At(i)) {
			return false
		}
	}
	return true
}

func matchAI(x, y *AI) bool {
	for i, yi := range y.elts {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

func matchAIAF(x *AI, y *AF) bool {
	for i, yi := range y.elts {
		if yi != float64(x.At(i)) {
			return false
		}
	}
	return true
}

func matchAF(x, y *AF) bool {
	for i, yi := range y.elts {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

// initArrayFlags sets Ascending flag if x is non-generic sorted array. It is
// used to set the flag on constants arrays.
func initArrayFlags(x array) {
	flags := x.getFlags()
	if !flags.Has(flagAscending) && sort.IsSorted(x) {
		x.setFlags(flags | flagAscending)
	}
}

func arrayAtV(x array, y V) array {
	switch yv := y.bv.(type) {
	case *AB:
		return x.atBytes(yv.elts)
	case *AI:
		return x.atInts(yv.elts)
	default:
		panic("arrayAtV")
	}
}

func vArrayAtV(x array, y V) V {
	switch yv := y.bv.(type) {
	case *AB:
		return x.vAtBytes(yv)
	case *AI:
		return x.vAtInts(yv)
	default:
		panic("arrayAtV")
	}
}

func (x *AB) numeric() bool { return true }
func (x *AI) numeric() bool { return true }
func (x *AF) numeric() bool { return true }
func (x *AS) numeric() bool { return false }
func (x *AV) numeric() bool { return false }

func (x *AB) canSet(y V) bool { return y.IsI() && y.uv >= 0 && y.uv < 256 }
func (x *AI) canSet(y V) bool { return y.IsI() }
func (x *AF) canSet(y V) bool { return y.IsF() }
func (x *AS) canSet(y V) bool { _, ok := y.bv.(S); return ok }
func (x *AV) canSet(y V) bool { return true }

func ascending(x array) bool {
	return x.getFlags().Has(flagAscending)
}
