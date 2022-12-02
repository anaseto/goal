package goal

import (
	"fmt"
	"math"
)

func B2I(b bool) (i int) {
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

func num2I(x V) (n int) {
	if x.IsInt() {
		return x.Int()
	}
	switch x := x.Value.(type) {
	case F:
		n = int(x)
	}
	// x is assumed to be a number.
	return n
}

func isNum(x V) bool {
	if x.IsInt() {
		return true
	}
	_, ok := x.Value.(F)
	return ok
}

func divideF(x, y F) F {
	if y == 0 {
		return F(math.Inf(int(signF(x))))
	}
	return x / y
}

func modI(x, y int) int {
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

func minI(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func maxI(x, y int) int {
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
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r, xv.Slice)
		return NewAB(r)
	case *AF:
		r := make([]float64, xv.Len())
		copy(r, xv.Slice)
		return NewAF(r)
	case *AI:
		r := make([]int, xv.Len())
		copy(r, xv.Slice)
		return NewAI(r)
	case *AS:
		r := make([]string, xv.Len())
		copy(r, xv.Slice)
		return NewAS(r)
	case *AV:
		r := make([]V, xv.Len())
		for i := range r {
			r[i] = clone(xv.At(i))
		}
		return NewAV(r)
	default:
		return x
	}
}

func cloneShallow(x V) V {
	switch xv := x.Value.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r, xv.Slice)
		return NewAB(r)
	case *AF:
		r := make([]float64, xv.Len())
		copy(r, xv.Slice)
		return NewAF(r)
	case *AI:
		r := make([]int, xv.Len())
		copy(r, xv.Slice)
		return NewAI(r)
	case *AS:
		r := make([]string, xv.Len())
		copy(r, xv.Slice)
		return NewAS(r)
	case *AV:
		r := make([]V, xv.Len())
		copy(r, xv.Slice)
		return NewAV(r)
	default:
		return x
	}
}

func cloneShallowArray(x array) array {
	switch xv := x.(type) {
	case *AB:
		r := make([]bool, xv.Len())
		copy(r, xv.Slice)
		return &AB{Slice: r}
	case *AF:
		r := make([]float64, xv.Len())
		copy(r, xv.Slice)
		return &AF{Slice: r}
	case *AI:
		r := make([]int, xv.Len())
		copy(r, xv.Slice)
		return &AI{Slice: r}
	case *AS:
		r := make([]string, xv.Len())
		copy(r, xv.Slice)
		return &AS{Slice: r}
	case *AV:
		r := make([]V, xv.Len())
		copy(r, xv.Slice)
		return &AV{Slice: r}
	default:
		return x
	}
}

// isIndices returns true if we have indices in canonical form, that is,
// using types I, AI and AV of thoses.
func isIndices(x V) bool {
	if x.IsInt() {
		return true
	}
	switch xv := x.Value.(type) {
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
	if x.IsInt() {
		return x
	}
	switch xv := x.Value.(type) {
	case F:
		if !isI(xv) {
			return errf("non-integer index (%g)", xv)
		}
		return NewI(int(xv))
	case *AB:
		return fromABtoAI(xv)
	case *AF:
		return toAI(xv)
	case *AV:
		r := make([]V, xv.Len())
		for i, z := range xv.Slice {
			r[i] = toIndicesRec(z)
			if r[i].IsErr() {
				return r[i]
			}
		}
		return canonicalV(NewAV(r))
	default:
		return errs("not an indices array")
	}
}

// toArray converts atoms into 1-length arrays. It returns arrays as-is.
func toArray(x V) V {
	if x.IsInt() {
		switch x.Int() {
		case 0, 1:
			return NewAB([]bool{x.Int() == 1})
		default:
			return NewAI([]int{int(x.Int())})
		}
	}
	switch xv := x.Value.(type) {
	case F:
		return NewAF([]float64{float64(xv)})
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
	r := make([]int, x.Len())
	for i, xi := range x.Slice {
		if !isI(F(xi)) {
			return errf("contains non-integer (%g)", xi)
		}
		r[i] = int(xi)
	}
	return NewAI(r)
}

// fromABtoAI converts AB into AI (for simplifying code, used only for
// unfrequent code).
func fromABtoAI(x *AB) V {
	r := make([]int, x.Len())
	for i := range r {
		r[i] = int(B2I(x.At(i)))
	}
	return NewAI(r)
}

func isFalse(x V) bool {
	if x.IsInt() {
		return x.Int() == 0
	}
	switch xv := x.Value.(type) {
	case F:
		return xv == 0
	case S:
		return xv == ""
	default:
		return x == V{} || Length(x) == 0
	}
}

func isTrue(x V) bool {
	if x.IsInt() {
		return x.Int() != 0
	}
	switch xv := x.Value.(type) {
	case F:
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
	if x.IsInt() {
		switch x.Int() {
		case 0, 1:
			return tB
		default:
			return tI
		}
	}
	switch x.Value.(type) {
	case F:
		return tF
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
		return tAO
	default:
		return tV
	}
}

// cType returns the canonical eltype of x. XXX: unused.
func cType(x V) eltype {
	if x.IsInt() {
		switch x.Int() {
		case 0, 1:
			return tB
		default:
			return tI
		}
	}
	switch x := x.Value.(type) {
	case *AB:
		return tAB
	case F:
		return tF
	case *AF:
		return tAF
	case *AI:
		return tAI
	case S:
		return tS
	case *AS:
		return tAS
	case *AV:
		return cTypeAO(x)
	default:
		return tV
	}
}

func cTypeAO(x *AV) eltype {
	if x.Len() == 0 {
		return tAO
	}
	t := eType(x.At(0))
	for _, xi := range x.Slice[1:] {
		t = mergeTypes(t, eType(xi))
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
func aType(x *AV) eltype {
	if x.Len() == 0 {
		return tV
	}
	t := eType(x.Slice[0])
	for i := 1; i < x.Len(); i++ {
		t = mergeTypes(t, eType(x.At(i)))
	}
	return t
}

func sameType(x, y V) bool {
	if x.IsInt() {
		return y.IsInt()
	}
	switch x.Value.(type) {
	case F:
		_, ok := y.Value.(F)
		return ok
	case *AB:
		_, ok := y.Value.(*AB)
		return ok
	case *AI:
		_, ok := y.Value.(*AI)
		return ok
	case *AF:
		_, ok := y.Value.(*AF)
		return ok
	case *AS:
		_, ok := y.Value.(*AS)
		return ok
	case *AV:
		_, ok := y.Value.(*AV)
		return ok
	default:
		// TODO: sameType, handle other cases (unused for now)
		return false
	}
}

func compatEltType(x array, y V) bool {
	switch x.(type) {
	case *AI:
		return y.IsInt()
	case *AF:
		_, ok := y.Value.(F)
		return ok
	case *AS:
		_, ok := y.Value.(S)
		return ok
	case *AV:
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

func isBI(x int) bool {
	return x == 0 || x == 1
}

func isBF(x F) bool {
	return x == 0 || x == 1
}

func minMax(x *AI) (min, max int) {
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

func maxAI(x *AI) int {
	max := math.MinInt
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

func minMaxB(x *AB) (int, int) {
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
	return B2I(min), B2I(max)
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
	switch xv := x.Value.(type) {
	case *AV:
		_, ok := isCanonical(xv)
		return ok
	default:
		return true
	}
}

// isCanonical returns true if the array is in canonical form, that is, it uses
// the most specialized representation. For example AV{I(2), I(3)} is not
// canonical, but AI{2, 3} is.
func isCanonical(x *AV) (eltype, bool) {
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

func assertCanonical(x *AV) {
	_, ok := isCanonical(x)
	if !ok {
		panic(fmt.Sprintf("not canonical: %#v", x))
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
			r[i] = xi.Int() != 0
		}
		return &AB{Slice: r}, true
	case tI:
		r := make([]int, x.Len())
		for i, xi := range x.Slice {
			r[i] = int(xi.Int())
		}
		return &AI{Slice: r}, true
	case tF:
		r := make([]float64, x.Len())
		for i, xi := range x.Slice {
			if xi.IsInt() {
				r[i] = float64(xi.Int())
			} else {
				r[i] = float64(xi.F())
			}
		}
		return &AF{Slice: r}, true
	case tS:
		r := make([]string, x.Len())
		for i, xi := range x.Slice {
			r[i] = string(xi.Value.(S))
		}
		return &AS{Slice: r}, true
	case tV:
		for i, xi := range x.Slice {
			x.Slice[i] = canonicalV(xi)
		}
		return x, false
	default:
		// should not happen
		return x, false
	}
}

// canonicalV returns the canonical form of a given value.
func canonicalV(x V) V {
	switch xv := x.Value.(type) {
	case *AV:
		r, b := normalize(xv)
		if b {
			return NewV(r)
		}
		return x
	default:
		return x
	}
}

// canonical returns the canonical form of a given generic array.
func canonical(x *AV) Value {
	r, _ := normalize(x)
	return r
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

func sumAB(x *AB) int {
	n := 0
	for _, xi := range x.Slice {
		if xi {
			n++
		}
	}
	return n
}
