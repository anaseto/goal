package main

//go:generate stringer -type=TokenType,ppTokenType,ppBlockType,opcode -output stringer.go

import (
	"fmt"
	"strings"
)

// V represents any kind of value.
type V interface {
	Len() int
	Type() string
}

type F float64 // F represents real numbers.
type I int     // I represents integers.
type S string  // S represents (immutable) strings of bytes.
type E string  // E represents errors

func (f F) Len() int      { return 1 }
func (i I) Len() int      { return 1 }
func (s S) Len() int      { return 1 }
func (e E) Len() int      { return 1 }
func (f F) Type() string  { return "f" }
func (i I) Type() string  { return "i" }
func (s S) Type() string  { return "s" }
func (e E) Type() string  { return "e" }
func (e E) Error() string { return string(e) }

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
	vCut                      // ^
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
	vCut:      "^",
	vTake:     "#",
	vDrop:     "_",
	vCast:     "$",
	vFind:     "?",
	vApply:    "@",
	vApplyN:   ".",
	vList:     "List",
	vEach:     "'",
	vFold:     "/",
	vScan:     "\\",
}

func (v Variadic) String() string {
	return vStrings[v]
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
// always called monadically. XXX: not used for now.
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
func (c Composition) Len() int   { return 1 }
func (l Lambda) Len() int        { return 1 }

func (v Variadic) Type() string      { return "v" }
func (p Projection) Type() string    { return "p" }
func (p ProjectionOne) Type() string { return "p" }
func (c Composition) Type() string   { return "q" }
func (r DerivedVerb) Type() string   { return "r" }
func (l Lambda) Type() string        { return "l" }

func (p Projection) String() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%v", p.Fun)
	sb.WriteRune('[')
	for i := len(p.Args) - 1; i >= 0; i-- {
		arg := p.Args[i]
		if arg != nil {
			fmt.Fprintf(sb, "%v", arg)
		}
		if i > 0 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(']')
	return sb.String()
}

func (p ProjectionOne) String() string {
	return fmt.Sprintf("%v[%v;]", p.Fun, p.Arg)
}

func (c Composition) String() string { return fmt.Sprintf("%v %v", c.Left, c.Right) }

func (r DerivedVerb) String() string {
	switch arg := r.Arg.(type) {
	case Composition:
		return fmt.Sprintf("(%v)%v", arg, r.Fun)
	default:
		return fmt.Sprintf("%v%v", arg, r.Fun)
	}
}

func (l Lambda) String() string { return fmt.Sprintf("{Lambda %d}", l) }

// Array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type Array interface {
	V
	At(i int) V           // x[i]
	Slice(i, j int) Array // x[i:j]
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

func (x AV) Slice(i, j int) Array { return x[i:j] }
func (x AB) Slice(i, j int) Array { return x[i:j] }
func (x AI) Slice(i, j int) Array { return x[i:j] }
func (x AF) Slice(i, j int) Array { return x[i:j] }
func (x AS) Slice(i, j int) Array { return x[i:j] }

func (x AV) String() string {
	sb := &strings.Builder{}
	sb.WriteRune('(')
	for i, v := range x {
		if v != nil {
			fmt.Fprintf(sb, "%v", v)
		}
		if i < len(x)-1 {
			sb.WriteRune(';')
		}
	}
	sb.WriteRune(')')
	return sb.String()
}
func (x AB) String() string {
	sb := &strings.Builder{}
	for i, v := range x {
		fmt.Fprintf(sb, "%d", B2I(v))
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}
func (x AI) String() string {
	sb := &strings.Builder{}
	for i, v := range x {
		fmt.Fprintf(sb, "%d", v)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x AF) String() string {
	sb := &strings.Builder{}
	for i, v := range x {
		fmt.Fprintf(sb, "%g", v)
		if i < len(x)-1 {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (x AS) String() string {
	sb := &strings.Builder{}
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
func (c Composition) Rank(ctx *Context) int   { return c.Right.Rank(ctx) }
func (l Lambda) Rank(ctx *Context) int        { return ctx.prog.Lambdas[l].Rank }
