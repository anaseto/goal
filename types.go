package main

type O interface{}  // O represents any kind of value.
type B bool         // B represents booleans (0 and 1 but less memory)
type F float64      // F represents real numbers.
type I int64        // I represents integers.
type S string       // S represents (immutable) strings of bytes.
type E error        // Errors (TODO: think about it)
type D func(O, O) O // D represents dyadic operators (TODO: distinct type for more arguments?)
type M func(O) O    // M represents monadic functions
type AO []O         // generic array
type AB []B         // boolean array
type AF []F         // real array
type AI []I         // integer array
type AS []S         // string array
