package main

type O interface{} // O represents any kind of value.
type B = bool      // B represents booleans (0 and 1 but less memory)
type F = float64   // F represents real numbers.
type I = int       // I represents integers.
type S = string    // S represents (immutable) strings of bytes.
type E = error     // E represents errors (TODO: think about it)

// Q represents a verb composition.
type Q struct {
	Verbs []O
}

// R represents a derived verb.
type R struct {
}

type U int // U represents monadic verbs.

const (
	USelf     U = iota // ::
	UFlip              // +:
	UNegate            // -:
	UFirst             // *:
	UClassify          // %: (classify instead of sqrt? or abs?)
	UEnum              // !:
	UWhere             // &:
	UReverse           // |:
	UAscend            // <:
	UDescend           // >:
	UGroup             // =:
	UNot               // ~:
	UEnlist            // ,:
	UNull              // ^: (maybe change)
	ULength            // #:
	UFloor             // _:
	UString            // $:
	UUniq              // ?:
	UType              // @:
	UEval              // .:
)

type V int // dyadic verbs

const (
	VRight    V = iota // :
	VAdd               // +
	VSubtract          // -
	VMultiply          // *
	VDivide            // %
	VMod               // !
	VAnd               // &
	VOr                // |
	VLess              // <
	VMore              // >
	VEqual             // =
	VMatch             // ~
	VConcat            // ,
	VWithout           // ^
	VTake              // #
	VDrop              // _
	VCast              // $
	VFind              // ?
	VApply             // @
	VApplyN            // .
)

// W represents adverbs.
type W int

const (
	WEach      W = iota // '
	WEachPrior          // ':
	WFold               // /
	WScan               // \
	WEachRight          // /:
	WEachLeft           // \:
)

type AO []O // generic array
type AB []B // boolean array
type AF []F // real array
type AI []I // integer array (TODO: optimization: add Range type)
type AS []S // string array

// Array interface is satisfied by the different kind of supported arrays.
// Typical implementation is given in comments.
type Array interface {
	At(i I) O           // x[i]
	Len() I             // len(x)
	Slice(i, j I) Array // x[i:j]
}

func (x AO) At(i I) O {
	return x[i]
}

func (x AO) Len() I {
	return len(x)
}

func (x AO) Slice(i, j I) Array {
	return x[i:j]
}

func (x AB) At(i I) O {
	return x[i]
}

func (x AB) Len() I {
	return len(x)
}

func (x AB) Slice(i, j I) Array {
	return x[i:j]
}

func (x AI) At(i I) O {
	return x[i]
}

func (x AI) Len() I {
	return len(x)
}

func (x AI) Slice(i, j I) Array {
	return x[i:j]
}

func (x AF) At(i I) O {
	return x[i]
}

func (x AF) Len() I {
	return len(x)
}

func (x AF) Slice(i, j I) Array {
	return x[i:j]
}

func (x AS) At(i I) O {
	return x[i]
}

func (x AS) Len() I {
	return len(x)
}

func (x AS) Slice(i, j I) Array {
	return x[i:j]
}
