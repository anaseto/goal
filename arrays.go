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
	getFlags() flags
	setFlags(flags)
	set(i int, y V)
	atIndices(y *AI) array // x[y] (goal code)
	shallowClone() array
}

type flags int32

const (
	flagNone      flags = 0b00
	flagAscending flags = 0b01
	flagUnique    flags = 0b10 // unused for now
)

func (f flags) Has(ff flags) bool {
	return f&ff != 0
}

// AB represents an array of booleans.
type AB struct {
	flags flags
	rc    *int
	Slice []bool
}

// NewAB returns a new boolean array. It does not initialize the reference
// counter.
func NewAB(x []bool) V {
	return NewV(&AB{Slice: x})
}

// NewABWithRC returns a new boolean array.
func NewABWithRC(x []bool, rc *int) V {
	return NewV(&AB{Slice: x, rc: rc})
}

// AI represents an array of integers.
type AI struct {
	flags flags
	rc    *int
	Slice []int64
}

// NewAI returns a new int array. It does not initialize the reference
// counter.
func NewAI(x []int64) V {
	return NewV(&AI{Slice: x})
}

// NewAIWithRC returns a new int array.
func NewAIWithRC(x []int64, rc *int) V {
	return NewV(&AI{Slice: x, rc: rc})
}

// AF represents an array of reals.
type AF struct {
	flags flags
	rc    *int
	Slice []float64
}

// NewAF returns a new array of reals. It does not initialize the reference
// counter.
func NewAF(x []float64) V {
	return NewV(&AF{Slice: x})
}

// NewAFWithRC returns a new array of reals.
func NewAFWithRC(x []float64, rc *int) V {
	return NewV(&AF{Slice: x, rc: rc})
}

// AS represents an array of strings.
type AS struct {
	flags flags
	rc    *int
	Slice []string // string array
}

// NewAS returns a new array of strings. It does not initialize the reference
// counter.
func NewAS(x []string) V {
	return NewV(&AS{Slice: x})
}

// NewASWithRC returns a new array of strings.
func NewASWithRC(x []string, rc *int) V {
	return NewV(&AS{Slice: x, rc: rc})
}

// AV represents a generic array.
type AV struct {
	flags flags
	rc    *int
	Slice []V
}

// NewAV returns a new generic array. It does not initialize the reference
// counter.
func NewAV(x []V) V {
	return NewV(&AV{Slice: x})
}

// NewAVWithRC returns a new generic array.
func NewAVWithRC(x []V, rc *int) V {
	return NewV(&AV{Slice: x, rc: rc})
}

// Type returns a string representation of the array's type.
func (x *AB) Type() string { return "N" }

// Type returns a string representation of the array's type.
func (x *AI) Type() string { return "N" }

// Type returns a string representation of the array's type.
func (x *AF) Type() string { return "N" }

// Type returns a string representation of the array's type.
func (x *AS) Type() string { return "S" }

// Type returns a string representation of the array's type.
func (x *AV) Type() string { return "A" }

// Len returns the length of the array.
func (x *AB) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AI) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AF) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AS) Len() int { return len(x.Slice) }

// Len returns the length of the array.
func (x *AV) Len() int { return len(x.Slice) }

func (x *AB) at(i int) V { return NewI(b2i(x.Slice[i])) }
func (x *AI) at(i int) V { return NewI(x.Slice[i]) }
func (x *AF) at(i int) V { return NewF(x.Slice[i]) }
func (x *AS) at(i int) V { return NewS(x.Slice[i]) }
func (x *AV) at(i int) V { return x.Slice[i] }

// At returns array value at the given index.
func (x *AB) At(i int) bool { return x.Slice[i] }

// At returns array value at the given index.
func (x *AI) At(i int) int64 { return x.Slice[i] }

// At returns array value at the given index.
func (x *AF) At(i int) float64 { return x.Slice[i] }

// At returns array value at the given index.
func (x *AS) At(i int) string { return x.Slice[i] }

// At returns array value at the given index.
func (x *AV) At(i int) V { return x.Slice[i] }

func (x *AB) slice(i, j int) array { return &AB{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AI) slice(i, j int) array { return &AI{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AF) slice(i, j int) array { return &AF{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AS) slice(i, j int) array { return &AS{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }
func (x *AV) slice(i, j int) array { return &AV{rc: x.rc, flags: x.flags, Slice: x.Slice[i:j]} }

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

// set changes x at i with y (in place).
func (x *AB) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n != 0
	} else {
		x.Slice[i] = y.F() != 0
	}
}

// set changes x at i with y (in place).
func (x *AI) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = y.n
	} else {
		x.Slice[i] = int64(y.F())
	}
}

// set changes x at i with y (in place).
func (x *AF) set(i int, y V) {
	if y.IsI() {
		x.Slice[i] = float64(y.I())
	} else {
		x.Slice[i] = y.F()
	}
}

// set changes x at i with y (in place).
func (x *AS) set(i int, y V) {
	x.Slice[i] = string(y.value.(S))
}

// set changes x at i with y (in place).
func (x *AV) set(i int, y V) {
	y.InitWithRC(x.rc)
	x.Slice[i] = y
}

func (x *AB) atIndices(y *AI) array {
	r := make([]bool, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.Slice {
		if yi < 0 {
			yi += xlen
		}
		r[i] = x.At(int(yi))
	}
	return &AB{Slice: r}
}

func (x *AI) atIndices(y *AI) array {
	r := make([]int64, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.Slice {
		if yi < 0 {
			yi += xlen
		}
		r[i] = x.At(int(yi))
	}
	return &AI{Slice: r}
}

func (x *AF) atIndices(y *AI) array {
	r := make([]float64, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.Slice {
		if yi < 0 {
			yi += xlen
		}
		r[i] = x.At(int(yi))
	}
	return &AF{Slice: r}
}

func (x *AS) atIndices(y *AI) array {
	r := make([]string, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.Slice {
		if yi < 0 {
			yi += xlen
		}
		r[i] = x.At(int(yi))
	}
	return &AS{Slice: r}
}

func (x *AV) atIndices(y *AI) array {
	r := make([]V, y.Len())
	xlen := int64(x.Len())
	for i, yi := range y.Slice {
		if yi < 0 {
			yi += xlen
		}
		r[i] = x.At(int(yi))
	}
	nr := &AV{Slice: r}
	var p *int
	if !reusableRCp(p) {
		var n int
		p = &n
	} else {
		p = x.rc
		*p++
	}
	nr.InitWithRC(p)
	a, _ := normalize(nr)
	return a
}

func (x *AB) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AB{Slice: make([]bool, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AI) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AI{Slice: make([]int64, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AF) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AF{Slice: make([]float64, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AS) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AS{Slice: make([]string, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AV) shallowClone() array {
	if reusableRCp(x.rc) {
		x.setFlags(flagNone)
		return x
	}
	var n int
	r := &AV{Slice: make([]V, x.Len()), rc: &n}
	copy(r.Slice, x.Slice)
	return r
}

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
	for i, yi := range yv.Slice {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

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
	for i, yi := range yv.Slice {
		if !Match(yi, x.At(i)) {
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
	for i, yi := range y.Slice {
		if yi != x.At(i) {
			return false
		}
	}
	return true
}

func matchABAI(x *AB, y *AI) bool {
	for i, yi := range y.Slice {
		if yi != b2i(x.At(i)) {
			return false
		}
	}
	return true
}

func matchABAF(x *AB, y *AF) bool {
	for i, yi := range y.Slice {
		if yi != b2f(x.At(i)) {
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
		if yi != float64(x.At(i)) {
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

// initArrayFlags sets Ascending flag if x is non-generic sorted array. It is
// used to set the flag on constants arrays.
func initArrayFlags(x array) {
	flags := x.getFlags()
	if !flags.Has(flagAscending) && sort.IsSorted(x) {
		x.setFlags(flags | flagAscending)
	}
}
