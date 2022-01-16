package main

// Object represents any kind of value.
type Object interface{}

type B bool    // B represents booleans (0 and 1 but less memory)
type I int64   // I represents integers.
type F float64 // F represents real numbers.
type S string  // S represents byte strings (can be utf8 or not).
//type C []rune represents strings of code points. TODO: decide whether it's worth or not.
type D func(Object, Object) Object // D represents dyadic operators
type M func(Object) Object         // M represents monadic functions
type AO []Object                   //
type AB []B
type AI []I
type AF []F
type AS []S

//type LC []C TODO
type E error // Errors
