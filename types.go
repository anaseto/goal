package main

// Object represents any kind of value.
type Object interface{}

type B bool                        // B represents booleans (0 and 1 but less memory)
type F float64                     // F represents real numbers.
type I int64                       // I represents integers.
type R rune                        // R represents runes.
type S string                      // S represents (immutable) strings of bytes (blobs).
type E error                       // Errors
type D func(Object, Object) Object // D represents dyadic operators
type M func(Object) Object         // M represents monadic functions
type AB []B                        // boolean array
type AF []F                        // real array
type AI []I                        // integer array
type AR []R                        // text
type AS []S                        // blob array
type AO []Object                   // generic array
