package main

//go:generate stringer -type=Monad,Dyad,Adverb,TokenType,ppTokenType,ppBlockType,opcode -output stringer.go

// V represents any kind of value.
type V interface {
	Len() int
	Type() string
}

type B bool    // B represents booleans (0 and 1 but less memory)
type F float64 // F represents real numbers.
type I int     // I represents integers.
type S string  // S represents (immutable) strings of bytes.
type E string  // E represents errors

func (b B) Len() int      { return 1 }
func (f F) Len() int      { return 1 }
func (i I) Len() int      { return 1 }
func (s S) Len() int      { return 1 }
func (e E) Len() int      { return 1 }
func (b B) Type() string  { return "b" }
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

// Array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type Array interface {
	V
	At(i int) V           // x[i]
	Slice(i, j int) Array // x[i:j]
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
func (x AB) At(i int) V { return B(x[i]) }
func (x AI) At(i int) V { return I(x[i]) }
func (x AF) At(i int) V { return F(x[i]) }
func (x AS) At(i int) V { return S(x[i]) }

func (x AV) Slice(i, j int) Array { return x[i:j] }
func (x AB) Slice(i, j int) Array { return x[i:j] }
func (x AI) Slice(i, j int) Array { return x[i:j] }
func (x AF) Slice(i, j int) Array { return x[i:j] }
func (x AS) Slice(i, j int) Array { return x[i:j] }

// Monad represents built-in 1-symbol unary operators.
type Monad int32

const (
	VReturn    Monad = iota // :
	VFlip                   // +
	VNegate                 // -
	VFirst                  // *
	VClassify               // %
	VRange                  // !
	VWhere                  // &
	VReverse                // |
	VGradeUp                // <
	VGradeDown              // >
	VGroup                  // =
	VNot                    // ~
	VEnlist                 // ,
	VSort                   // ^
	VLen                    // #
	VFloor                  // _
	VString                 // $
	VNub                    // ?
	VType                   // @
	VEval                   // .
)

// Dyad represents built-in 1-symbol binary operators.
type Dyad int32

const (
	VRight    Dyad = iota // :
	VAdd                  // +
	VSubtract             // -
	VMultiply             // *
	VDivide               // %
	VMod                  // !
	VMin                  // &
	VMax                  // |
	VLess                 // <
	VMore                 // >
	VEqual                // =
	VMatch                // ~
	VJoin                 // ,
	VCut                  // ^
	VTake                 // #
	VDrop                 // _
	VCast                 // $
	VFind                 // ?
	VApply                // @
	VApplyN               // .
)

// Adverb represents verb modifiers.
type Adverb int32

const (
	AEach Adverb = iota // '
	AFold               // /
	AScan               // \
)

// Variadic represents verbs with variable arity > 2.
type Variadic int32

const (
	VList Variadic = iota
	VAmend
)

// DerivedVerb represents a value modified by an adverb.
type DerivedVerb struct {
	Adverb Adverb
	Value  V
}

// Projection represents a partial application of a function. Because many
// functions do not have a fixed arity, the number of provided arguments can be
// arbitrary.
type Projection struct {
	Fun  Function
	Args AV
}

// Composition represents a composition of several functions. All except the
// last will be called monadically. XXX: not really any Function.
type Composition struct {
	Funs []Function
}

// Lambda represents an user defined function by ID.
type Lambda int32

// Function represents any kind of callable value that can be projected.
type Function interface {
	V
	Project(AV) Projection
}

func (u Monad) Len() int       { return 1 }
func (v Dyad) Len() int        { return 1 }
func (vv Variadic) Len() int   { return 1 }
func (w Adverb) Len() int      { return 1 }
func (r DerivedVerb) Len() int { return 1 }
func (p Projection) Len() int  { return 1 }
func (c Composition) Len() int { return 1 }
func (l Lambda) Len() int      { return 1 }

func (u Monad) Type() string       { return "u" }
func (v Dyad) Type() string        { return "v" }
func (vv Variadic) Type() string   { return "V" }
func (w Adverb) Type() string      { return "w" }
func (r DerivedVerb) Type() string { return "r" }
func (p Projection) Type() string  { return "p" }
func (c Composition) Type() string { return "c" }
func (l Lambda) Type() string      { return "l" }

func (u Monad) Project(vs AV) Projection       { return Projection{Fun: u, Args: vs} }
func (v Dyad) Project(vs AV) Projection        { return Projection{Fun: v, Args: vs} }
func (w Adverb) Project(vs AV) Projection      { return Projection{Fun: w, Args: vs} }
func (r DerivedVerb) Project(vs AV) Projection { return Projection{Fun: r, Args: vs} }
func (p Projection) Project(vs AV) Projection  { return Projection{Fun: p, Args: vs} }
func (c Composition) Project(vs AV) Projection { return Projection{Fun: c, Args: vs} }
func (l Lambda) Project(vs AV) Projection      { return Projection{Fun: l, Args: vs} }

// vReturn represents a return.
type vReturn struct{}

func (rv vReturn) Len() int     { return 1 }
func (rv vReturn) Type() string { return "return" }
