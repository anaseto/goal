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
	Matches(Value) bool
	// Fprint writes a unique program string representation of the value.
	Fprint(*Context, ValueWriter) (n int, err error)
	// Type returns the name of the value's type. It may be used by Less to
	// sort non-comparable values using lexicographic order.  This means
	// Type should return different values for non-comparable values.
	Type() string
	// Less returns true if the value should be orderer before the given
	// one. It is used for sorting values, but not for element-wise
	// comparison with < and >. It should produce a strict total order, so,
	// in particular, if x < y, then we do not have y > x, and one of them
	// should hold unless both values match.
	Less(Value) bool
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

// Panic returns the panic string. It assumes IsPanic is true.
func (x V) Panic() string {
	if x.IsPanic() {
		return string(x.value.(panicV))
	}
	return ""
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
//	lambda		number of arguments
//	projections	number of gaps
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
	stype() string
}

// Rank for a projection is the number of nil arguments.
func (p *projection) rank(ctx *Context) int { return countNils(p.Args) }

// Rank for a 1-arg projection is 1.
func (p *projectionFirst) rank(ctx *Context) int { return 1 }

// Rank for a curryfied function is 1.
func (p *projectionMonad) rank(ctx *Context) int { return 1 }

// Rank returns the rank of a derived verb.
func (r *derivedVerb) rank(ctx *Context) int {
	switch r.Fun {
	case vEach:
		// f' has same rank as f
		return r.Arg.Rank(ctx)
	default:
		// f/ and f\ have rank derived from f, except that by default
		// it's one less, because we consider as default case the
		// non-seeded case.
		return maxInt(0, r.Arg.Rank(ctx)-1)
	}
}

func (p *projection) stype() string      { return "p" }
func (p *projectionFirst) stype() string { return "pf" }
func (p *projectionMonad) stype() string { return "pm" }
func (r *derivedVerb) stype() string     { return "r" }

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
