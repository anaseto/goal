package main

// V represents any kind of value.
type V interface {
	Len() int
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
func (e E) Error() string { return string(e) }

// Verb represents built-in 1-symbol operators.
type Verb int

const (
	VRight    Verb = iota // :
	VAdd                  // +
	VSubtract             // -
	VMultiply             // *
	VDivide               // %
	VMod                  // !
	VAnd                  // &
	VOr                   // |
	VLess                 // <
	VMore                 // >
	VEqual                // =
	VMatch                // ~
	VConcat               // ,
	VWithout              // ^
	VTake                 // #
	VDrop                 // _
	VCast                 // $
	VFind                 // ?
	VApply                // @
	VApplyN               // .
)

// Adverb represents verb modifiers. They are not values by themselves.
type Adverb int

const (
	AEach Adverb = iota // '
	AFold               // /
	AScan               // \
)

type AV []V       // generic array
type AB []bool    // boolean array
type AF []float64 // real array
type AI []int     // integer array (TODO: optimization: add Range type)
type AS []string  // string array

// Array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type Array interface {
	V
	At(i int) V           // x[i]
	Slice(i, j int) Array // x[i:j]
}

func (x AV) At(i int) V {
	return x[i]
}

func (x AV) Len() int {
	return len(x)
}

func (x AV) Slice(i, j int) Array {
	return x[i:j]
}

func (x AB) At(i int) V {
	return B(x[i])
}

func (x AB) Len() int {
	return len(x)
}

func (x AB) Slice(i, j int) Array {
	return x[i:j]
}

func (x AI) At(i int) V {
	return I(x[i])
}

func (x AI) Len() int {
	return len(x)
}

func (x AI) Slice(i, j int) Array {
	return x[i:j]
}

func (x AF) At(i int) V {
	return F(x[i])
}

func (x AF) Len() int {
	return len(x)
}

func (x AF) Slice(i, j int) Array {
	return x[i:j]
}

func (x AS) At(i int) V {
	return S(x[i])
}

func (x AS) Len() int {
	return len(x)
}

func (x AS) Slice(i, j int) Array {
	return x[i:j]
}
