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
	Arity int          // Number of arguments
	Fun   func(...O) O // Function
}
type AO []O // generic array
type AB []B // boolean array
type AF []F // real array
type AI []I // integer array (TODO: optimization: add Range type)
type AS []S // string array
