package goal

//go:generate stringer -type=TokenType,astTokenType,astBlockType,opcode -output stringer.go

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// V represents any kind of value.
type V interface {
	Len() int
	Type() string
	Sprint(*Context) string
}

type F float64   // F represents real numbers.
type I int       // I represents integers.
type S string    // S represents (immutable) strings of bytes.
type errV string // E represents errors

func (f F) Len() int                      { return 1 }
func (i I) Len() int                      { return 1 }
func (s S) Len() int                      { return 1 }
func (e errV) Len() int                   { return 1 }
func (f F) Type() string                  { return "f" }
func (i I) Type() string                  { return "i" }
func (s S) Type() string                  { return "s" }
func (e errV) Type() string               { return "e" }
func (f F) Sprint(ctx *Context) string    { return fmt.Sprintf("%g", f) }
func (i I) Sprint(ctx *Context) string    { return fmt.Sprintf("%d", i) }
func (s S) Sprint(ctx *Context) string    { return strconv.Quote(string(s)) }
func (e errV) Sprint(ctx *Context) string { return fmt.Sprintf("'ERROR %s", e) }

func (e errV) Error() string { return string(e) }

type AV []V       // generic array
type AB []bool    // boolean array
type AF []float64 // real array
type AI []int     // integer array (TODO: optimization: add Range type?)
type AS []string  // string array

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
	vIn                       // in
	vList                     // (...;...;...)
	vEach                     // ' (adverb)
	vFold                     // / (adverb)
	vScan                     // \ (adverb)
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

func (v Variadic) Sprint(ctx *Context) string {
	return v.String()
}

// DerivedVerb represents values modified by an adverb.
type DerivedVerb struct {
	Fun Variadic
	Arg V
}

// Projection represents a partial application of a function. Because variadic
// verbs do not have a fixed arity, it is possible to produce a projection of
// arbitrary arity.
type Projection struct {
	Fun  Function
	Args AV
}

// ProjectionOne represents a projection with one argument.
type ProjectionOne struct {
	Fun Function
	Arg V
}

// Composition represents a composition of two functions. The left one is
// always called monadically. NOTE: unused for now.
type Composition struct {
	Left  Function
	Right Function
}

// Lambda represents an user defined function by ID.
type Lambda int32

func (v Variadic) Len() int      { return 1 }
func (r DerivedVerb) Len() int   { return 1 }
func (p Projection) Len() int    { return 1 }
func (p ProjectionOne) Len() int { return 1 }
func (q Composition) Len() int   { return 1 }
func (l Lambda) Len() int        { return 1 }

func (v Variadic) Type() string      { return "v" }
func (p Projection) Type() string    { return "p" }
func (p ProjectionOne) Type() string { return "p" }
func (q Composition) Type() string   { return "q" }
func (r DerivedVerb) Type() string   { return "r" }
func (l Lambda) Type() string        { return "l" }

func (p Projection) Sprint(ctx *Context) string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s", p.Fun.Sprint(ctx))
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg != nil {
			fmt.Fprintf(sb, "%s", arg.Sprint(ctx))
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p ProjectionOne) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s[%s;]", p.Fun.Sprint(ctx), p.Arg.Sprint(ctx))
}

func (q Composition) Sprint(ctx *Context) string {
	return fmt.Sprintf("%s %s", q.Left.Sprint(ctx), q.Right.Sprint(ctx))
}

func (r DerivedVerb) Sprint(ctx *Context) string {
	switch arg := r.Arg.(type) {
	case Composition:
		return fmt.Sprintf("(%s)%s", arg.Sprint(ctx), r.Fun.Sprint(ctx))
	default:
		return fmt.Sprintf("%s%s", arg.Sprint(ctx), r.Fun.Sprint(ctx))
	}
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
	V
	At(i int) V           // x[i]
	Slice(i, j int) array // x[i:j]
	Select(y AI) V        // x[y] (goal code)
}

func (x AV) Len() int { return len(x) }
func (x AB) Len() int { return len(x) }
func (x AI) Len() int { return len(x) }
func (x AF) Len() int { return len(x) }
func (x AS) Len() int { return len(x) }

func (x AV) Type() string { return "A" }
func (x AB) Type() string { return "B" }
func (x AI) Type() string { return "I" }
func (x AF) Type() string { return "F" }
func (x AS) Type() string { return "S" }

func (x AV) At(i int) V { return x[i] }
func (x AB) At(i int) V { return B2I(x[i]) }
func (x AI) At(i int) V { return I(x[i]) }
func (x AF) At(i int) V { return F(x[i]) }
func (x AS) At(i int) V { return S(x[i]) }

func (x AV) Slice(i, j int) array { return x[i:j] }
func (x AB) Slice(i, j int) array { return x[i:j] }
func (x AI) Slice(i, j int) array { return x[i:j] }
func (x AF) Slice(i, j int) array { return x[i:j] }
func (x AS) Slice(i, j int) array { return x[i:j] }

// sprintV returns a string for a V deep in an AV.
func sprintV(ctx *Context, v V) string {
	av, ok := v.(AV)
	if ok {
		return av.sprint(ctx, true)
	}
	return v.Sprint(ctx)
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
	for i, v := range x {
		if v != nil {
			fmt.Fprintf(sb, "%s", sprintV(ctx, v))
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
	for i, v := range x {
		fmt.Fprintf(sb, "%d", B2I(v))
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
	for i, v := range x {
		fmt.Fprintf(sb, "%d", v)
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
	for i, v := range x {
		fmt.Fprintf(sb, "%g", v)
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
	for i, v := range x {
		fmt.Fprintf(sb, "%q", v)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

// Function interface is satisfied by the different kind of functions. A
// function is a value thas has a default rank. Note that arrays do have a
// “rank” but do not implement this interface.
type Function interface {
	V
	Rank(ctx *Context) int
}

func (v Variadic) Rank(ctx *Context) int      { return 2 }
func (r DerivedVerb) Rank(ctx *Context) int   { return 2 }
func (p Projection) Rank(ctx *Context) int    { return countNils(p.Args) }
func (p ProjectionOne) Rank(ctx *Context) int { return 1 }
func (q Composition) Rank(ctx *Context) int   { return q.Right.Rank(ctx) }
func (l Lambda) Rank(ctx *Context) int        { return ctx.lambdas[l].Rank }

type zeroFun interface {
	Function
	zero() V
}

func (v Variadic) zero() V {
	switch v {
	case vAdd, vSubtract:
		return I(0)
	case vMultiply:
		return I(1)
	case vMin:
		return I(math.MinInt)
	case vMax:
		return I(math.MaxInt)
	}
	return nil
}

func (v Variadic) Matches(x V) bool {
	xv, ok := x.(Variadic)
	return ok && v == xv
}

func (r DerivedVerb) Matches(x V) bool {
	xr, ok := x.(DerivedVerb)
	return ok && r.Fun == xr.Fun && Match(r.Arg, xr.Arg)
}

func (p Projection) Matches(x V) bool {
	xp, ok := x.(Projection)
	return ok && Match(p.Fun, xp.Fun) && Match(p.Args, xp.Args)
}

func (p ProjectionOne) Matches(x V) bool {
	xp, ok := x.(ProjectionOne)
	return ok && Match(p.Fun, xp.Fun) && Match(p.Arg, xp.Arg)
}

func (q Composition) Matches(x V) bool {
	xq, ok := x.(Composition)
	return ok && Match(q.Left, xq.Left) && Match(q.Right, xq.Right)
}

func (l Lambda) Matches(x V) bool {
	xl, ok := x.(Lambda)
	return ok && l == xl
}
