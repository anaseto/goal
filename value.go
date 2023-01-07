package goal

//go:generate stringer -type=TokenType,astTokenType,opcode -output stringer.go

import (
	"io"
	"unsafe"
)

// V contains a boxed or unboxed value.
type V struct {
	kind  valueKind // valInt, valFloat, valBoxed, ...
	n     int64     // unboxed integer or float value
	value Value     // boxed value
}

// valueKind represents the kinds of values.
type valueKind int8

const (
	valNil      valueKind = iota
	valInt                // unboxed int64 (n field)
	valFloat              // unboxed float64 (n field)
	valVariadic           // unboxed int32 (n field)
	valLambda             // unboxed int32 (n field)
	valBoxed              // boxed value (value field)
	valPanic              // boxed value (value field)
)

// ValueWriter is the interface used when formatting values with Fprint.
type ValueWriter interface {
	io.Writer
	io.ByteWriter
	io.StringWriter
}

// Value is the interface satisfied by all boxed values.
type Value interface {
	// Matches returns true if the value matches another (in the sense of
	// the ~ operator).
	Matches(x Value) bool
	// Fprint writes a unique program string representation of the value.
	Fprint(*Context, ValueWriter) (n int, err error)
	// Type returns the name of the value's type.
	Type() string
}

// newVariadic returns a new variadic value.
func newVariadic(v variadic) V {
	return V{kind: valVariadic, n: int64(v)}
}

// lambda represents an user defined function by ID.
type lambda int32

// newLambda returns a new lambda value.
func newLambda(v lambda) V {
	return V{kind: valLambda, n: int64(v)}
}

// NewError returns a new recoverable error value.
func NewError(x V) V {
	return V{kind: valBoxed, value: &errV{V: x}}
}

// NewI returns a new int64 value.
func NewI(i int64) V {
	return V{kind: valInt, n: i}
}

// NewF returns a new float64 value.
func NewF(f float64) V {
	i := *(*int64)(unsafe.Pointer(&f))
	return V{kind: valFloat, n: i}
}

// NewS returns a new string value.
func NewS(s string) V {
	return V{kind: valBoxed, value: S(s)}
}

// NewV returns a new boxed value.
func NewV(bv Value) V {
	return V{kind: valBoxed, value: bv}
}

// variadic retrieves the variadic value from N field. It assumes Kind is
// IntVariadic.
func (x V) variadic() variadic {
	return variadic(x.n)
}

// Variadic retrieves the lambda value from N field. It assumes Kind is
// IntLambda.
func (x V) lambda() lambda {
	return lambda(x.n)
}

// Error retrieves the error value. It assumes IsError(v).
func (x V) Error() V {
	return x.value.(*errV).V
}

// I retrieves the unboxed integer value from N field. It assumes IsI(v).
func (x V) I() int64 {
	return x.n
}

// F retrieves the unboxed float64 value. It assumes isF(v).
func (x V) F() float64 {
	i := x.n
	f := *(*float64)(unsafe.Pointer(&i))
	return f
}

// Value retrieves the boxed value, or nil if the value is not boxed. You can
// check whether the value is boxed with IsValue(v).
func (x V) Value() Value {
	return x.value
}

// Type returns the name of the value's type.
func (x V) Type() string {
	switch x.kind {
	case valNil:
		return "nil"
	case valInt:
		return "n"
	case valFloat:
		return "n"
	case valVariadic:
		return "f"
	case valLambda:
		return "f"
	case valBoxed:
		return x.value.Type()
	default:
		return ""
	}
}

// IsI returns true if the value is an integer.
func (x V) IsI() bool {
	return x.kind == valInt
}

// IsF returns true if the value is a float.
func (x V) IsF() bool {
	return x.kind == valFloat
}

// IsPanic returns true if the value is a fatal error.
func (x V) IsPanic() bool {
	return x.kind == valPanic
}

// IsError returns true if the value is a recoverable error.
func (x V) IsError() bool {
	if x.kind != valBoxed {
		return false
	}
	_, ok := x.value.(*errV)
	return ok
}

// IsValue returns true if the value is a boxed value satisfying the Value
// interface. You can then get the value with the Value method.
func (x V) IsValue() bool {
	return x.kind == valBoxed
}

// IsFunction returns true if the value is some kind of function.
func (x V) IsFunction() bool {
	switch x.kind {
	case valVariadic, valLambda:
		return true
	case valBoxed:
		_, ok := x.value.(function)
		return ok
	default:
		return false
	}
}

// Rank returns the default rank of the value, that is the number of arguments
// it normally takes. It returns 0 for non-function values. This default rank
// is used when a function is used in an adverbial expression that has
// different semantics depending on the function arity. Currently, ranks are as
// follows:
//
//	variadic	2
//	projections	number of nils
//	lambda		number of arguments
//	derived verb	1
func (x V) Rank(ctx *Context) int {
	switch x.kind {
	case valVariadic:
		return 2
	case valLambda:
		return ctx.lambdas[x.n].Rank
	case valBoxed:
		if xf, ok := x.value.(function); ok {
			return xf.rank(ctx)
		}
		return 0
	default:
		return 0
	}
}

// IncrRC increments the value reference count (if it has any).
func (x V) IncrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.IncrRC()
	}
}

// IncrRC increments the value reference count (if it has any).
func (x V) DecrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

func (x V) rcdecrRefCounter() {
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

// errV represents a recoverable error. It may contain some goal value of any
// kind.
type errV struct {
	V V
}

func (e *errV) Matches(y Value) bool {
	switch yv := y.(type) {
	case *errV:
		return Match(e.V, yv.V)
	default:
		return false
	}
}

func (e *errV) Type() string { return "e" }

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

func (e panicV) Type() string { return "panic" }

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
	rc    int32
	flags flags
	Slice []bool
}

// NewAB returns a new boolean array.
func NewAB(x []bool) V {
	return NewV(&AB{Slice: x})
}

// AI represents an array of integers.
type AI struct {
	rc    int32
	flags flags
	Slice []int64
}

// NewAI returns a new int array.
func NewAI(x []int64) V {
	return NewV(&AI{Slice: x})
}

// newAscUniqAI returns a new sorted int array.
func newAscUniqAI(x []int64) V {
	return NewV(&AI{Slice: x, flags: flagAscending | flagUnique})
}

// AF represents an array of reals.
type AF struct {
	rc    int32
	flags flags
	Slice []float64
}

// NewAF returns a new array of reals.
func NewAF(x []float64) V {
	return NewV(&AF{Slice: x})
}

// AS represents an array of strings.
type AS struct {
	rc    int32
	flags flags
	Slice []string // string array
}

// NewAS returns a new array of strings.
func NewAS(x []string) V {
	return NewV(&AS{Slice: x})
}

// AV represents a generic array.
type AV struct {
	rc    int32
	flags flags
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
	RefCounter
	reusable() bool
	Len() int
	at(i int) V            // x[i]
	slice(i, j int) array  // x[i:j]
	atIndices(y []int64) V // x[y] (goal code)
	set(i int, y V)
	getFlags() flags
	setFlags(flags)
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

func (x *AB) slice(i, j int) array { return &AB{rc: x.rc, Slice: x.Slice[i:j]} }
func (x *AI) slice(i, j int) array { return &AI{rc: x.rc, Slice: x.Slice[i:j]} }
func (x *AF) slice(i, j int) array { return &AF{rc: x.rc, Slice: x.Slice[i:j]} }
func (x *AS) slice(i, j int) array { return &AS{rc: x.rc, Slice: x.Slice[i:j]} }
func (x *AV) slice(i, j int) array { return &AV{rc: x.rc, Slice: x.Slice[i:j]} }

func (x *AB) reusable() bool { return x.rc <= 1 }
func (x *AI) reusable() bool { return x.rc <= 1 }
func (x *AF) reusable() bool { return x.rc <= 1 }
func (x *AS) reusable() bool { return x.rc <= 1 }
func (x *AV) reusable() bool { return x.rc <= 1 }

func (x *AB) reuse() *AB {
	if x.rc <= 1 {
		x.flags = flagNone
		return x
	}
	return &AB{Slice: make([]bool, x.Len())}
}

func (x *AI) reuse() *AI {
	if x.rc <= 1 {
		x.flags = flagNone
		return x
	}
	return &AI{Slice: make([]int64, x.Len())}
}

func (x *AF) reuse() *AF {
	if x.rc <= 1 {
		x.flags = flagNone
		return x
	}
	return &AF{Slice: make([]float64, x.Len())}
}

func (x *AS) reuse() *AS {
	if x.rc <= 1 {
		x.flags = flagNone
		return x
	}
	return &AS{Slice: make([]string, x.Len())}
}

func (x *AV) reuse() *AV {
	if x.rc <= 1 {
		x.flags = flagNone
		return x
	}
	return &AV{Slice: make([]V, x.Len())}
}

// derivedVerb represents values modified by an adverb. This kind value is not
// manipulable within the program, as it is only produced as an intermediary
// value in adverb trains and only appears as an adverb argument.
type derivedVerb struct {
	Fun variadic
	Arg V
}

// projection represents a partial application of a function. Because variadic
// verbs do not have a fixed arity, it is possible to produce a projection of
// arbitrary arity.
type projection struct {
	Fun  V
	Args []V
}

// projectionFirst represents a monadic projection fixing the first argument of
// a function with rank greater than 2.
type projectionFirst struct {
	Fun V // function with rank >= 2
	Arg V // first argument x
}

// projectionMonad represents a monadic projection of a function of any rank.
type projectionMonad struct {
	Fun V
}

func (p *projection) Type() string      { return "f" }
func (p *projectionFirst) Type() string { return "f" }
func (p *projectionMonad) Type() string { return "f" }
func (r *derivedVerb) Type() string     { return "f" }

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
func (p *projection) rank(ctx *Context) int { return countNils(p.Args) }

// Rank for a 1-arg projection is 1.
func (p *projectionFirst) rank(ctx *Context) int { return 1 }

// Rank for a curryfied function is 1.
func (p *projectionMonad) rank(ctx *Context) int { return 1 }

// Rank returns 1 for derived verbs.
func (r *derivedVerb) rank(ctx *Context) int { return 1 }

func (p *projection) Matches(x Value) bool {
	xp, ok := x.(*projection)
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

func (p *projectionFirst) Matches(x Value) bool {
	xp, ok := x.(*projectionFirst)
	return ok && Match(p.Fun, xp.Fun) && Match(p.Arg, xp.Arg)
}

func (p *projectionMonad) Matches(x Value) bool {
	xp, ok := x.(*projectionMonad)
	return ok && Match(p.Fun, xp.Fun)
}

func (r *derivedVerb) Matches(x Value) bool {
	xr, ok := x.(*derivedVerb)
	return ok && r.Fun == xr.Fun && Match(r.Arg, xr.Arg)
}

// RefCounter is implemented by values that use a reference count. In goal the
// refcount is not used for memory management, but only for optimization of
// memory allocations.  Refcount is increased by each assignement, and each use
// in an operation. It is reduced after each operation, and for each last use
// of a variable (as approximated conservatively). If refcount is equal or less
// than one, then the value is considered reusable.
//
// When defining a new type implementing the Value interface, it is only
// necessary to also implement RefCounter if the type definition contains makes
// use of a type implementing it (for example an array type or a generic V).
type RefCounter interface {
	IncrRC() // IncrRC increments the reference count by one.
	DecrRC() // DecrRC decrements the reference count by one.
}

func (e *errV) IncrRC()       { e.V.IncrRC() }
func (e *errV) DecrRC()       { e.V.DecrRC() }
func (r *replacer) IncrRC()   { r.oldnew.IncrRC() }
func (r *replacer) DecrRC()   { r.oldnew.DecrRC() }
func (r *rxReplacer) IncrRC() { r.repl.IncrRC() }
func (r *rxReplacer) DecrRC() { r.repl.DecrRC() }

func (x *AB) IncrRC() { x.rc++ }
func (x *AI) IncrRC() { x.rc++ }
func (x *AF) IncrRC() { x.rc++ }
func (x *AS) IncrRC() { x.rc++ }
func (x *AV) IncrRC() {
	x.rc++
	for _, xi := range x.Slice {
		xi.IncrRC()
	}
}

func (r *derivedVerb) IncrRC() {
	r.Arg.IncrRC()
}

func (p *projection) IncrRC() {
	p.Fun.IncrRC()
	for _, arg := range p.Args {
		arg.IncrRC()
	}
}

func (p *projectionFirst) IncrRC() {
	p.Fun.IncrRC()
	p.Arg.IncrRC()
}

func (p *projectionMonad) IncrRC() {
	p.Fun.IncrRC()
}

func (x *AB) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AI) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AF) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AS) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

func (x *AV) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
	for _, xi := range x.Slice {
		xi.DecrRC()
	}
}

func (r *derivedVerb) DecrRC() {
	r.Arg.DecrRC()
}

func (p *projection) DecrRC() {
	p.Fun.DecrRC()
	for _, arg := range p.Args {
		arg.DecrRC()
	}
}

func (p *projectionFirst) DecrRC() {
	p.Fun.DecrRC()
	p.Arg.DecrRC()
}

func (p *projectionMonad) DecrRC() {
	p.Fun.DecrRC()
}
