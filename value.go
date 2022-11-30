package goal

//go:generate stringer -type=TokenType,astTokenType,opcode -output stringer.go

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// V represents a boxed or unboxed value.
type V struct {
	Kind  ValueKind // int, boxed
	Flags int8      // unused for now (for sorted)
	N     int       // refcount or unboxed integer value
	Value Value     // boxed value
}

// ValueKind represents the kinds of values.
type ValueKind int8

const (
	Nil ValueKind = iota
	Int
	IntVariadic
	IntLambda
	Boxed // boxed value (Value field)
)

// Value represents any kind of boxed value.
type Value interface {
	// Matches returns true if the value matches another (in the sense of
	// the ~ operator).
	Matches(x Value) bool
	// Sprint returns a prettified string representation of the value.
	Sprint(*Context) string
	// Type returns the name of the value's type.
	Type() string
}

// Type returns the name of the value's type.
func (v V) Type() string {
	switch v.Kind {
	case Int:
		return "n"
	case IntVariadic:
		return "v"
	case IntLambda:
		return "l"
	case Boxed:
		return v.Value.Type()
	default:
		return ""
	}
}

// Variadic retrieves the Variadic value from N field. It assumes Kind is
// IntVariadic.
func (v V) Variadic() Variadic {
	return Variadic(v.N)
}

// Variadic retrieves the Lambda value from N field. It assumes Kind is
// IntLambda.
func (v V) Lambda() Lambda {
	return Lambda(v.N)
}

// Sprint returns a prettified string representation of the value.
func (v V) Sprint(ctx *Context) string {
	switch v.Kind {
	case Int:
		return fmt.Sprintf("%d", v.N)
	case IntVariadic:
		return "v"
	case IntLambda:
		return "l"
	case Boxed:
		return v.Value.Sprint(ctx)
	default:
		return ""
	}
}

func (v V) Rank(ctx *Context) int {
	switch v.Kind {
	case IntVariadic:
		return 2
	case IntLambda:
		return ctx.lambdas[v.N].Rank
	case Boxed:
		if vf, ok := v.Value.(Function); ok {
			return vf.Rank(ctx)
		}
		return 0
	default:
		return 0
	}
}

// NewV returns a new boxed value.
func NewV(bv Value) V {
	return V{Kind: Boxed, Value: bv}
}

// NewVariadic returns a new int value.
func NewVariadic(v Variadic) V {
	return V{Kind: IntVariadic, N: int(v)}
}

// NewLambda returns a new int value.
func NewLambda(v Lambda) V {
	return V{Kind: IntLambda, N: int(v)}
}

// NewI returns a new int value.
func NewI(i int) V {
	return V{Kind: Boxed, Value: I(i)}
}

// NewF returns a new float64 value.
func NewF(f float64) V {
	return V{Kind: Boxed, Value: F(f)}
}

// NewS returns a new string value.
func NewS(s string) V {
	return V{Kind: Boxed, Value: S(s)}
}

// F represents real numbers.
type F float64

// I represents integers.
type I int

// S represents (immutable) strings of bytes.
type S string

// errV represents errors
type errV string

func isErr(x V) bool {
	_, ok := x.Value.(errV)
	return ok
}

func isFunction(x V) bool {
	switch x.Kind {
	case IntVariadic, IntLambda:
		return true
	case Boxed:
		_, ok := x.Value.(Function)
		return ok
	default:
		return false
	}
}

func (f F) Matches(y Value) bool {
	switch y := y.(type) {
	case I:
		return f == F(y)
	case F:
		return f == y
	default:
		return false
	}
}

func (i I) Matches(y Value) bool {
	switch y := y.(type) {
	case I:
		return i == y
	case F:
		return F(i) == y
	default:
		return false
	}
}

func (s S) Matches(y Value) bool {
	switch y := y.(type) {
	case S:
		return s == y
	default:
		return false
	}
}

// Type retuns "n" for numeric atoms.
func (f F) Type() string { return "n" }

// Type retuns "n" for numeric atoms.
func (i I) Type() string { return "n" }

// Type retuns "s" for string atoms.
func (s S) Type() string { return "s" }

func (f F) Sprint(ctx *Context) string { return fmt.Sprintf("%g", f) }
func (i I) Sprint(ctx *Context) string { return fmt.Sprintf("%d", i) }
func (s S) Sprint(ctx *Context) string { return strconv.Quote(string(s)) }

func (e errV) Matches(y Value) bool {
	err, ok := y.(errV)
	return ok && e == err
}

func (e errV) Type() string               { return "e" }
func (e errV) Sprint(ctx *Context) string { return fmt.Sprintf("'ERROR %s", e) }
func (e errV) Error() string              { return string(e) }

// AV represents a generic array.
type AV []V

// AB represents an array of booleans.
type AB []bool

// AF represents an array of reals.
type AF []float64

// AI represents an array of integers.
type AI []int

// AS represents an array of strings.
type AS []string // string array

// Variadic represents a built-in function.
type Variadic int32

const (
	vRight    Variadic = iota // :
	vAdd                      // +
	vSubtract                 // -
	vMultiply                 // *
	vDivide                   // %
	vMod                      // !
	vMin                      // &
	vMax                      // |
	vLess                     // <
	vMore                     // >
	vEqual                    // =
	vMatch                    // ~
	vJoin                     // ,
	vWithout                  // ^
	vTake                     // #
	vDrop                     // _
	vCast                     // $
	vFind                     // ?
	vApply                    // @
	vApplyN                   // .
	vList                     // (...;...;...)
	vEach                     // ' (adverb)
	vFold                     // / (adverb)
	vScan                     // \ (adverb)
	vIn                       // in
	vSign                     // sign
	vOCount                   // ocount (occurrence count)
	vICount                   // icount (index count)
	vBytes                    // bytes (byte count)
	vAnd                      // and
	vOr                       // or
)

var vStrings = [...]string{
	vRight:    ":",
	vAdd:      "+",
	vSubtract: "-",
	vMultiply: "*",
	vDivide:   "%",
	vMod:      "!",
	vMin:      "&",
	vMax:      "|",
	vLess:     "<",
	vMore:     ">",
	vEqual:    "=",
	vMatch:    "~",
	vJoin:     ",",
	vWithout:  "^",
	vTake:     "#",
	vDrop:     "_",
	vCast:     "$",
	vFind:     "?",
	vApply:    "@",
	vApplyN:   ".",
	vIn:       "in",
	vList:     "list",
	vEach:     "'",
	vFold:     "/",
	vScan:     "\\",
}

func (v Variadic) String() string {
	if v <= vScan {
		return vStrings[v]
	}
	return fmt.Sprintf("{Variadic %d}", v)
}

// DerivedVerb represents values modified by an adverb. This kind value is not
// manipulable within the program, as it is only produced as an intermediary
// value in adverb trains and only appears as an adverb argument.
type DerivedVerb struct {
	Fun Variadic
	Arg V
}

// Projection represents a partial application of a function. Because variadic
// verbs do not have a fixed arity, it is possible to produce a projection of
// arbitrary arity.
type Projection struct {
	Fun  V
	Args AV
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

// Lambda represents an user defined function by ID.
type Lambda int32

func (p Projection) Type() string      { return "p" }
func (p ProjectionFirst) Type() string { return "p" }
func (p ProjectionMonad) Type() string { return "p" }
func (r DerivedVerb) Type() string     { return "r" }
func (l Lambda) Type() string          { return "l" }

func (p Projection) Sprint(ctx *Context) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s", p.Fun.Sprint(ctx))
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg != (V{}) {
			fmt.Fprintf(sb, "%s", arg.Sprint(ctx))
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p ProjectionFirst) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[%s;]", p.Fun.Sprint(ctx), p.Arg.Sprint(ctx))
}

func (p ProjectionMonad) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[]", p.Fun.Sprint(ctx))
}

func (r DerivedVerb) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s%s", r.Arg.Sprint(ctx), r.Fun.String())
}

func (l Lambda) Sprint(ctx *Context) string {
	if l < 0 || int(l) >= len(ctx.lambdas) {
		return fmt.Sprintf("{Lambda %d}", l)
	}
	return ctx.lambdas[l].Source
}

// array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type array interface {
	Value
	Len() int
	at(i int) V           // x[i]
	slice(i, j int) array // x[i:j]
	atIndices(y AI) V     // x[y] (goal code)
	set(i int, y V)
	//setIndices(y AI, z V) error
}

func (x AV) Matches(y Value) bool { return matchArray(x, y) }
func (x AB) Matches(y Value) bool { return matchArray(x, y) }
func (x AI) Matches(y Value) bool { return matchArray(x, y) }
func (x AF) Matches(y Value) bool { return matchArray(x, y) }
func (x AS) Matches(y Value) bool { return matchArray(x, y) }

// Len returns the length of the array.
func (x AV) Len() int { return len(x) }

// Len returns the length of the array.
func (x AB) Len() int { return len(x) }

// Len returns the length of the array.
func (x AI) Len() int { return len(x) }

// Len returns the length of the array.
func (x AF) Len() int { return len(x) }

// Len returns the length of the array.
func (x AS) Len() int { return len(x) }

func (x AV) Type() string { return "A" }
func (x AB) Type() string { return "B" }
func (x AI) Type() string { return "I" }
func (x AF) Type() string { return "F" }
func (x AS) Type() string { return "S" }

func (x AV) at(i int) V { return x[i] }
func (x AB) at(i int) V { return NewV(B2I(x[i])) }
func (x AI) at(i int) V { return NewI(x[i]) }
func (x AF) at(i int) V { return NewF(x[i]) }
func (x AS) at(i int) V { return NewS(x[i]) }

func (x AV) slice(i, j int) array { return x[i:j] }
func (x AB) slice(i, j int) array { return x[i:j] }
func (x AI) slice(i, j int) array { return x[i:j] }
func (x AF) slice(i, j int) array { return x[i:j] }
func (x AS) slice(i, j int) array { return x[i:j] }

// sprintV returns a string for a V deep in an AV.
func sprintV(ctx *Context, x V) string {
	avx, ok := x.Value.(AV)
	if ok {
		return avx.sprint(ctx, true)
	}
	return x.Sprint(ctx)
}

func (x AV) Sprint(ctx *Context) string {
	return x.sprint(ctx, false)
}

func (x AV) sprint(ctx *Context, deep bool) string {
	if len(x) == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if len(x) == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%s", x[0].Sprint(ctx))
		return sb.String()
	}
	sb.WriteRune('(')
	var sep string
	if deep {
		sep = ";"
	} else {
		sep = "\n "
	}
	t := aType(x)
	switch t {
	case tB, tI, tF, tS:
		sep = " "
	}
	for i, xi := range x {
		if xi != (V{}) {
			fmt.Fprintf(sb, "%s", sprintV(ctx, xi))
		}
		if i < len(x)-1 {
			sb.WriteString(sep)
		}
	}
	sb.WriteRune(')')
	return sb.String()
}

func (x AB) Sprint(ctx *Context) string {
	if len(x) == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if len(x) == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", B2I(x[0]))
		return sb.String()
	}
	for i, xi := range x {
		fmt.Fprintf(sb, "%d", B2I(xi))
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x AI) Sprint(ctx *Context) string {
	if len(x) == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if len(x) == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%d", x[0])
		return sb.String()
	}
	for i, xi := range x {
		fmt.Fprintf(sb, "%d", xi)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x AF) Sprint(ctx *Context) string {
	if len(x) == 0 {
		return `!0`
	}
	sb := &strings.Builder{}
	if len(x) == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%g", x[0])
		return sb.String()
	}
	for i, xi := range x {
		fmt.Fprintf(sb, "%g", xi)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x AS) Sprint(ctx *Context) string {
	if len(x) == 0 {
		return `0#""`
	}
	sb := &strings.Builder{}
	if len(x) == 1 {
		sb.WriteRune(',')
		fmt.Fprintf(sb, "%q", x[0])
		return sb.String()
	}
	for i, xi := range x {
		fmt.Fprintf(sb, "%q", xi)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

// Function interface is satisfied by the different kind of functions. A
// function is a value thas has a default rank. The default rank is used in
// situations where an adverb or function has different meanings depending on
// the arity of the function that is passed to it.
// Note that arrays do also have a “rank” but do not implement this interface.
type Function interface {
	Value
	Rank(ctx *Context) int
}

// Rank for a projection is the number of nil arguments.
func (p Projection) Rank(ctx *Context) int { return countNils(p.Args) }

// Rank for a 1-arg projection is 1.
func (p ProjectionFirst) Rank(ctx *Context) int { return 1 }

// Rank for a curryfied function is 1.
func (p ProjectionMonad) Rank(ctx *Context) int { return 1 }

// Rank returns 2 for derived verbs.
func (r DerivedVerb) Rank(ctx *Context) int { return 2 }

type zeroFun interface {
	Function
	zero() V
}

func (v Variadic) zero() V {
	switch v {
	case vAdd, vSubtract:
		return NewI(0)
	case vMultiply:
		return NewI(1)
	case vMin:
		return NewI(math.MinInt)
	case vMax:
		return NewI(math.MaxInt)
	}
	return V{}
}

func (p Projection) Matches(x Value) bool {
	xp, ok := x.(Projection)
	return ok && Match(p.Fun, xp.Fun) && p.Args.Matches(xp.Args)
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
