package main

type O interface{}  // O represents any kind of value.
type B = bool       // B represents booleans (0 and 1 but less memory)
type F = float64    // F represents real numbers.
type I = int        // I represents integers.
type S = string     // S represents (immutable) strings of bytes.
type E = error      // Errors (TODO: think about it)
type M func(O) O    // M represents monadic functions
type D func(O, O) O // D represents dyadic operators
// V represents a variadic function with more than two arguments
type V struct {
	Arity int          // Number of arguments > 2
	Fun   func(...O) O // Function
}
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
