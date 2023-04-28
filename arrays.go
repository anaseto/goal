package goal

import "sort"

type countable interface {
	Len() int
}

// array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type array interface {
	RefCountHolder
	countable
	sort.Interface
	at(i int) V           // x[i]
	slice(i, j int) array // x[i:j]
	getFlags() flags      // get the array's flags
	setFlags(flags)       // set the array's flags
	set(i int, y V)       // puts y at indix i, assuming compatibility
	atIndices(y *AI) V    // x[y] (like in goal code)
	atInts([]int64) array // like x[y] but assumes valid positive indices
	shallowClone() array  // shallow clone, erases flags
}

type flags int32

const (
	flagNone      flags = 0b000
	flagAscending flags = 0b001
	flagUnique    flags = 0b010
	flagBool      flags = 0b100
)

func (f flags) Has(ff flags) bool {
	return f&ff != 0
}

// AB represents an array of bytes. From Goal's perspective, it's the same as
// AI. It's used as an optimization to save space for small-integers, in
// particular for arrays of booleans (0s and 1s).
type AB struct {
	flags flags
	rc    *int
	elts  []byte
}

// NewAB returns a new byte array. It does not initialize the reference
// counter.
func NewAB(x []byte) V {
	return NewV(&AB{elts: x})
}

// newABb returns a new byte array with boolean flag. It does not initialize
// the reference counter.
func newABb(x []byte) V {
	return NewV(&AB{elts: x, flags: flagBool})
}

// IsBoolean returns true when the array of bytes is known to contain only 1s
// and 0s.
func (x *AB) IsBoolean() bool {
	return x.flags.Has(flagBool)
}

// Slice returns the underlying immutable slice of values. It should not be
// modified unless the value's refcount pointer is reusable, and even then, you
// should normally return a new array with the modified slice.
func (x *AB) Slice() []byte {
	return x.elts
}

// NewABWithRC returns a new byte array.
func NewABWithRC(x []byte, rc *int) V {
	return NewV(&AB{elts: x, rc: rc})
}

// newABbWithRC returns a new byte array.
func newABbWithRC(x []byte, rc *int) V {
	return NewV(&AB{elts: x, rc: rc, flags: flagBool})
}

// AI represents an array of integers.
type AI struct {
	flags flags
	rc    *int
	elts  []int64
}

// NewAI returns a new int array. It does not initialize the reference
// counter.
func NewAI(x []int64) V {
	return NewV(&AI{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified unless the value's refcount pointer is reusable, and even then, you
// should normally return a new array with the modified slice.
func (x *AI) Slice() []int64 {
	return x.elts
}

// NewAIWithRC returns a new int array.
func NewAIWithRC(x []int64, rc *int) V {
	return NewV(&AI{elts: x, rc: rc})
}

// AF represents an array of reals.
type AF struct {
	flags flags
	rc    *int
	elts  []float64
}

// NewAF returns a new array of reals. It does not initialize the reference
// counter.
func NewAF(x []float64) V {
	return NewV(&AF{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified unless the value's refcount pointer is reusable, and even then, you
// should normally return a new array with the modified slice.
func (x *AF) Slice() []float64 {
	return x.elts
}

// NewAFWithRC returns a new array of reals.
func NewAFWithRC(x []float64, rc *int) V {
	return NewV(&AF{elts: x, rc: rc})
}

// AS represents an array of strings.
type AS struct {
	flags flags
	rc    *int
	elts  []string // string array
}

// NewAS returns a new array of strings. It does not initialize the reference
// counter.
func NewAS(x []string) V {
	return NewV(&AS{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified unless the value's refcount pointer is reusable, and even then, you
// should normally return a new array with the modified slice.
func (x *AS) Slice() []string {
	return x.elts
}

// NewASWithRC returns a new array of strings.
func NewASWithRC(x []string, rc *int) V {
	return NewV(&AS{elts: x, rc: rc})
}

// AV represents a generic array.
type AV struct {
	flags flags
	rc    *int
	elts  []V
}

// NewAV returns a new generic array. It does not initialize the reference
// counter.
func NewAV(x []V) V {
	return NewV(&AV{elts: x})
}

// Slice returns the underlying immutable slice of values. It should not be
// modified unless the value's refcount pointer is reusable, and even then, you
// should normally return a new array with the modified slice.
func (x *AV) Slice() []V {
	return x.elts
}

// NewAVWithRC returns a new generic array.
func NewAVWithRC(x []V, rc *int) V {
	return NewV(&AV{elts: x, rc: rc})
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

func (x *AB) slice(i, j int) array { return &AB{rc: x.rc, flags: x.flags, elts: x.elts[i:j]} }
func (x *AI) slice(i, j int) array { return &AI{rc: x.rc, flags: x.flags, elts: x.elts[i:j]} }
func (x *AF) slice(i, j int) array { return &AF{rc: x.rc, flags: x.flags, elts: x.elts[i:j]} }
func (x *AS) slice(i, j int) array { return &AS{rc: x.rc, flags: x.flags, elts: x.elts[i:j]} }
func (x *AV) slice(i, j int) array { return &AV{rc: x.rc, flags: x.flags, elts: x.elts[i:j]} }

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
	if y.IsI() {
		x.elts[i] = byte(y.n)
	} else {
		x.elts[i] = byte(y.F())
	}
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AI) set(i int, y V) {
	if y.IsI() {
		x.elts[i] = y.n
	} else {
		x.elts[i] = int64(y.F())
	}
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AF) set(i int, y V) {
	if y.IsI() {
		x.elts[i] = float64(y.I())
	} else {
		x.elts[i] = y.F()
	}
}

// set changes x at i with y (in place), assuming the value is compatible.
func (x *AS) set(i int, y V) {
	x.elts[i] = string(y.value.(S))
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	y.InitWithRC(x.rc)
	x.elts[i] = y
}

func (x *AB) atIndices(y *AI) V {
	r := &AB{elts: make([]byte, y.Len())}
	if x.IsBoolean() {
		r.flags |= flagBool
	}
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("X@i : index out of bounds: %d (length %d)", yi, xlen)
		}
		r.elts[i] = x.At(int(yi))
	}
	return NewV(r)
}

func (x *AI) atIndices(y *AI) V {
	r := make([]int64, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("X@i : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAI(r)
}

func (x *AF) atIndices(y *AI) V {
	r := make([]float64, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("X@i : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAF(r)
}

func (x *AS) atIndices(y *AI) V {
	r := make([]string, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("X@i : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	return NewAS(r)
}

func (x *AV) atIndices(y *AI) V {
	r := make([]V, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.elts {
		if yi < 0 {
			yi += xlen
		}
		if yi < 0 || yi >= xlen {
			return Panicf("X@i : index out of bounds: %d (length %d)", yi, xlen)
		}
		r[i] = x.At(int(yi))
	}
	rv := &AV{elts: r}
	var n int
	rv.InitWithRC(&n)
	return NewV(canonicalAV(rv))
}

func (x *AB) atInts(y []int64) array {
	r := &AB{elts: make([]byte, len(y))}
	if x.IsBoolean() {
		r.flags |= flagBool
	}
	for i, yi := range y {
		r.elts[i] = x.At(int(yi))
	}
	return r
}

func (x *AI) atInts(y []int64) array {
	r := make([]int64, len(y))
	for i, yi := range y {
		r[i] = x.At(int(yi))
	}
	return &AI{elts: r}
}

func (x *AF) atInts(y []int64) array {
	r := make([]float64, len(y))
	for i, yi := range y {
		r[i] = x.At(int(yi))
	}
	return &AF{elts: r}
}

func (x *AS) atInts(y []int64) array {
	r := make([]string, len(y))
	for i, yi := range y {
		r[i] = x.At(int(yi))
	}
	return &AS{elts: r}
}

func (x *AV) atInts(y []int64) array {
	r := make([]V, len(y))
	for i, yi := range y {
		r[i] = x.At(int(yi))
	}
	nr := &AV{elts: r}
	var p *int
	if reusableRCp(x.rc) {
		p = x.rc
	} else {
		var n int
		p = &n
	}
	nr.InitWithRC(p)
	a, _ := normalize(nr)
	return a
}

func (x *AB) shallowClone() array {
	return shallowCloneAB(x)
}

func shallowCloneAB(x *AB) *AB {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AB{elts: make([]byte, x.Len()), rc: &n}
	copy(r.elts, x.elts)
	return r
}

func (x *AI) shallowClone() array {
	return shallowCloneAI(x)
}

func shallowCloneAI(x *AI) *AI {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AI{elts: make([]int64, x.Len()), rc: &n}
	copy(r.elts, x.elts)
	return r
}

func (x *AF) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AF{elts: make([]float64, x.Len()), rc: &n}
	copy(r.elts, x.elts)
	return r
}

func (x *AS) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AS{elts: make([]string, x.Len()), rc: &n}
	copy(r.elts, x.elts)
	return r
}

func (x *AV) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	r := &AV{elts: make([]V, x.Len())}
	copy(r.elts, x.elts)
	initRC(r)
	return r
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
