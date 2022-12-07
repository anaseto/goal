package goal

//go:generate stringer -type=TokenType,astTokenType,opcode -output stringer.go

import (
	"fmt"
	"unsafe"
)

// V represents a boxed or unboxed value.
type V struct {
	Kind  ValueKind // int, boxed
	N     int64     // refcount or unboxed integer value
	Value Value     // boxed value
}

// ValueKind represents the kinds of values.
type ValueKind int8

const (
	Nil      ValueKind = iota
	Int                // unboxed int64 (N field)
	Float              // unboxed float64 (N field)
	Variadic           // unboxed int32 (N field)
	Lambda             // unboxed int32 (N field)
	Boxed              // boxed value (Value field)
	Panic              // boxed value (Value field)
)

// Value represents any kind of boxed value.
type Value interface {
	fmt.Stringer
	// Matches returns true if the value matches another (in the sense of
	// the ~ operator).
	Matches(x Value) bool
	// Sprint returns a prettified string representation of the value.
	Sprint(*Context) string
	// Type returns the name of the value's type.
	Type() string
}

// newVariadic returns a new variadic value.
func newVariadic(v variadic) V {
	return V{Kind: Variadic, N: int64(v)}
}

// lambda represents an user defined function by ID.
type lambda int32

// newLambda returns a new lambda value.
func newLambda(v lambda) V {
	return V{Kind: Lambda, N: int64(v)}
}

// NewError returns a new recoverable error value.
func NewError(x V) V {
	return V{Kind: Boxed, Value: &errV{V: x}}
}

// NewI returns a new int64 value.
func NewI(i int64) V {
	return V{Kind: Int, N: i}
}

// NewF returns a new float64 value.
func NewF(f float64) V {
	i := *(*int64)(unsafe.Pointer(&f))
	return V{Kind: Float, N: i}
}

// NewS returns a new string value.
func NewS(s string) V {
	return V{Kind: Boxed, Value: S(s)}
}

// NewV returns a new boxed value.
func NewV(bv Value) V {
	return V{Kind: Boxed, Value: bv}
}

// variadic retrieves the variadic value from N field. It assumes Kind is
// IntVariadic.
func (v V) variadic() variadic {
	return variadic(v.N)
}

// Variadic retrieves the lambda value from N field. It assumes Kind is
// IntLambda.
func (v V) lambda() lambda {
	return lambda(v.N)
}

// Error retrieves the error value. It assumes IsError(v).
func (v V) Error() V {
	return v.Value.(errV).V
}

// I retrieves the integer value from N field. It assumes IsI(v).
func (v V) I() int64 {
	return v.N
}

// F retrieves the float64 value. It assumes Kind isF(v).
func (v V) F() float64 {
	i := v.N
	f := *(*float64)(unsafe.Pointer(&i))
	return f
}

// S retrieves the S value. It assumes Value type is S.
func (v V) S() S {
	return v.Value.(S)
}

// AB retrieves the *AB value. It assumes Value type is *AB.
func (v V) AB() *AB {
	return v.Value.(*AB)
}

// AI retrieves the *AI value. It assumes Value type is *AI.
func (v V) AI() *AI {
	return v.Value.(*AI)
}

// AF retrieves the *AF value. It assumes Value type is *AF.
func (v V) AF() *AF {
	return v.Value.(*AF)
}

// AS retrieves the *AS value. It assumes Value type is *AS.
func (v V) AS() *AS {
	return v.Value.(*AS)
}

// AV retrieves the *AV value. It assumes Value type is *AV.
func (v V) AV() *AV {
	return v.Value.(*AV)
}

// Type returns the name of the value's type.
func (v V) Type() string {
	switch v.Kind {
	case Nil:
		return "nil"
	case Int:
		return "n"
	case Float:
		return "n"
	case Variadic:
		return "f"
	case Lambda:
		return "f"
	case Boxed:
		return v.Value.Type()
	default:
		return ""
	}
}

// IsI returns true if the value is an integer.
func (x V) IsI() bool {
	return x.Kind == Int
}

// IsF returns true if the value is a float.
func (x V) IsF() bool {
	return x.Kind == Float
}

// isPanic returns true if the value is a fatal error.
func (x V) isPanic() bool {
	return x.Kind == Panic
}

// IsError returns true if the value is a recoverable error.
func (x V) IsError() bool {
	if x.Kind != Boxed {
		return false
	}
	_, ok := x.Value.(errV)
	return ok
}

// IsFunction returns true if the value is some kind of function.
func (x V) IsFunction() bool {
	switch x.Kind {
	case Variadic, Lambda:
		return true
	case Boxed:
		_, ok := x.Value.(function)
		return ok
	default:
		return false
	}
}

// Rank returns the default rank of the value, that is the number of arguments
// it normally takes. It returns 0 for non-function values.
func (v V) Rank(ctx *Context) int {
	switch v.Kind {
	case Variadic:
		return 2
	case Lambda:
		return ctx.lambdas[v.N].Rank
	case Boxed:
		if vf, ok := v.Value.(function); ok {
			return vf.rank(ctx)
		}
		return 0
	default:
		return 0
	}
}

func (v V) rcincr() {
	if v.Kind != Boxed {
		return
	}
	vrc, ok := v.Value.(refCounter)
	if ok {
		vrc.rcincr()
	}
}

func (v V) rcdecr() {
	if v.Kind != Boxed {
		return
	}
	vrc, ok := v.Value.(refCounter)
	if ok {
		vrc.rcdecr()
	}
}

// errV represents a recoverable error. It may contain some goal value of any
// kind.
type errV struct {
	V V
}

func (e errV) Matches(y Value) bool {
	switch yv := y.(type) {
	case errV:
		return Match(e.V, yv.V)
	default:
		return false
	}
}

func (e errV) Type() string { return "e" }

// panicV represents a fatal error string.
type panicV string

func (e panicV) Matches(y Value) bool {
	switch yv := y.(type) {
	case panicV:
		return e == yv
	default:
		return false
	}
}

func (e panicV) Type() string { return "e" }

// S represents (immutable) strings of bytes.
type S string

func (s S) Matches(y Value) bool {
	switch yv := y.(type) {
	case S:
		return s == yv
	default:
		return false
	}
}

// Type retuns "s" for string atoms.
func (s S) Type() string { return "s" }

type Flags int16

// AB represents an array of booleans.
type AB struct {
	rc    int16
	flags Flags
	Slice []bool
}

// NewAB returns a new boolean array.
func NewAB(x []bool) V {
	return NewV(&AB{Slice: x})
}

// AI represents an array of integers.
type AI struct {
	rc    int16
	flags Flags
	Slice []int64
}

// NewAI returns a new int array.
func NewAI(x []int64) V {
	return NewV(&AI{Slice: x})
}

// AF represents an array of reals.
type AF struct {
	rc    int16
	flags Flags
	Slice []float64
}

// NewAF returns a new array of reals.
func NewAF(x []float64) V {
	return NewV(&AF{Slice: x})
}

// AS represents an array of strings.
type AS struct {
	rc    int16
	flags Flags
	Slice []string // string array
}

// NewAS returns a new array of strings.
func NewAS(x []string) V {
	return NewV(&AS{Slice: x})
}

// AV represents a generic array.
type AV struct {
	rc    int16
	flags Flags
	Slice []V
}

// NewAV returns a new generic array.
func NewAV(x []V) V {
	return NewV(&AV{Slice: x})
}

func (x *AB) Matches(y Value) bool { return matchArray(x, y) }
func (x *AI) Matches(y Value) bool { return matchArray(x, y) }
func (x *AF) Matches(y Value) bool { return matchArray(x, y) }
func (x *AS) Matches(y Value) bool { return matchArray(x, y) }
func (x *AV) Matches(y Value) bool { return matchArray(x, y) }

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

// array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type array interface {
	Value
	refCounter
	Len() int
	at(i int) V            // x[i]
	slice(i, j int) array  // x[i:j]
	atIndices(y []int64) V // x[y] (goal code)
	set(i int, y V)
	//setIndices(y AI, z V) error
}

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

func (x *AB) at(i int) V { return NewI(B2I(x.Slice[i])) }
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

// DerivedVerb represents values modified by an adverb. This kind value is not
// manipulable within the program, as it is only produced as an intermediary
// value in adverb trains and only appears as an adverb argument.
type DerivedVerb struct {
	Fun variadic
	Arg V
}

// Projection represents a partial application of a function. Because variadic
// verbs do not have a fixed arity, it is possible to produce a projection of
// arbitrary arity.
type Projection struct {
	Fun  V
	Args []V
}

// ProjectionFirst represents a monadic projection fixing the first argument of
// a function with rank greater than 2.
type ProjectionFirst struct {
	Fun V // function with rank >= 2
	Arg V // first argument x
}

// ProjectionMonad represents a monadic projection of a function of any rank.
type ProjectionMonad struct {
	Fun V
}

func (p Projection) Type() string      { return "f" }
func (p ProjectionFirst) Type() string { return "f" }
func (p ProjectionMonad) Type() string { return "f" }
func (r DerivedVerb) Type() string     { return "f" }

// function interface is satisfied by the different kind of functions. A
// function is a value thas has a default rank. The default rank is used in
// situations where an adverb or function has different meanings depending on
// the arity of the function that is passed to it.
// Note that arrays do also have a “rank” but do not implement this interface.
type function interface {
	Value
	rank(ctx *Context) int
}

// Rank for a projection is the number of nil arguments.
func (p Projection) rank(ctx *Context) int { return countNils(p.Args) }

// Rank for a 1-arg projection is 1.
func (p ProjectionFirst) rank(ctx *Context) int { return 1 }

// Rank for a curryfied function is 1.
func (p ProjectionMonad) rank(ctx *Context) int { return 1 }

// Rank returns 2 for derived verbs.
func (r DerivedVerb) rank(ctx *Context) int { return 2 }

func (p Projection) Matches(x Value) bool {
	xp, ok := x.(Projection)
	if !ok || !Match(p.Fun, xp.Fun) {
		return false
	}
	if len(p.Args) != len(xp.Args) {
		return false
	}
	for i, arg := range p.Args {
		if !Match(arg, xp.Args[i]) {
			return false
		}
	}
	return true
}

func (p ProjectionFirst) Matches(x Value) bool {
	xp, ok := x.(ProjectionFirst)
	return ok && Match(p.Fun, xp.Fun) && Match(p.Arg, xp.Arg)
}

func (p ProjectionMonad) Matches(x Value) bool {
	xp, ok := x.(ProjectionMonad)
	return ok && Match(p.Fun, xp.Fun)
}

func (r DerivedVerb) Matches(x Value) bool {
	xr, ok := x.(DerivedVerb)
	return ok && r.Fun == xr.Fun && Match(r.Arg, xr.Arg)
}

// refCounter is implemented by values that use a reference count, allowing for
// memory reuse. Refcount is increased by each assignement, and each use in an
// operation. It is reduced after each operation, and for each last use of a
// variable (as approximated conservatively). If refcount is less than one,
// then the variable can be reused.
type refCounter interface {
	rcincr()
	rcdecr()
	reusable() bool
}

func (x *AB) reusable() bool { return x.rc <= 1 }
func (x *AI) reusable() bool { return x.rc <= 1 }
func (x *AF) reusable() bool { return x.rc <= 1 }
func (x *AS) reusable() bool { return x.rc <= 1 }
func (x *AV) reusable() bool { return x.rc <= 1 }

func (x *AB) rcincr() { x.rc++ }
func (x *AI) rcincr() { x.rc++ }
func (x *AF) rcincr() { x.rc++ }
func (x *AS) rcincr() { x.rc++ }
func (x *AV) rcincr() {
	x.rc++
	for _, xi := range x.Slice {
		xi.rcincr()
	}
}

func (x *AB) rcdecr() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AI) rcdecr() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AF) rcdecr() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AS) rcdecr() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AV) rcdecr() {
	if x.rc > 0 {
		x.rc--
	}
	for _, xi := range x.Slice {
		xi.rcdecr()
	}
}

func (x *AB) reuse() *AB {
	if x.rc <= 1 {
		return x
	}
	return &AB{Slice: make([]bool, x.Len())}
}

func (x *AI) reuse() *AI {
	if x.rc <= 1 {
		return x
	}
	return &AI{Slice: make([]int64, x.Len())}
}

func (x *AF) reuse() *AF {
	if x.rc <= 1 {
		return x
	}
	return &AF{Slice: make([]float64, x.Len())}
}

func (x *AS) reuse() *AS {
	if x.rc <= 1 {
		return x
	}
	return &AS{Slice: make([]string, x.Len())}
}

func (x *AV) reuse() *AV {
	if x.rc <= 1 {
		return x
	}
	return &AV{Slice: make([]V, x.Len())}
}
