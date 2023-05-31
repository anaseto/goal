*Last updated: 2023-05-31*

# Implementation notes

This document talks about Goal's implementation. It's intended to be a kind of
starting point before digging into the code or extending it, while telling as
well how some things ended up in a way or another.

# Choosing the implementation language

To be honest, among suitable languages, I chose Go because I know it very well
and feel comfortable with it. It might not be the usual go-to language for
implementing interpreters, but it has quite a few good points for such task:

* A comprehensive standard library, without requiring more external
  dependencies.
* Higher-level language for embedding than what C provides, so that it's easier
  to extend.
* Low-level enough that things go fast.
* Fast compilation.

Among languages I know well, OCaml mostly fits too, except for the first point.
Of course, given current trends, I did think of languages like Zig (built-in
SIMD) and Rust (with it's excellent regexp library), but I only know the first
on the surface, I don't feel comfortable with the second, and SIMD comes at
quite a cost in code complexity and portability anyway.

Also, Go gives us excellent garbage collection out of the box, which is quite a
good thing, as my knowledge about garbage collection implementation is limited.
I know there are GC libraries for most non-GC languages, but it's still one
less thing to worry about. As a tradeoff, we cannot catch out of memory errors
reliably in programs.

Even counting the recent generics addition, a downside of Go for implementing
an interpreter might be the lack of macros, but code generation helps, and
given Goal's somewhat minimalist design, it's not an issue for me.

Interestingly, while Go and Goal are quite the opposite in terms of conciseness
due to the gap between scalar and array paradigms, they both encourage idioms
over abstraction, and writing executable code over writing types: this might
explain why this in at least the third project for an array language in Go.

# Context

The `Context` struct type in `context.go` represents the state of the
interpreter.  It's the first type that you will encounter if you want to embed
Goal in a program. The `Context` type records all the information needed for
execution, including arrays of constants, globals, lambdas, extra built-ins, or
error locations, so it's the natural entry point for embedded usage. From an
internal's perspective, there's nothing unusual there, so we'll rather talk
about program representation next.

# Program representation

The compiler chain is quite typical: scanning goes first (`scanner.go`), then
we parse into an AST (`parser.go`), which is then compiled into bytecode
(`compiler.go`), and then finally run by a VM (`vm.go`).

One thing to note, is that in Goal, variables are resolved during compilation,
and still textual in the AST. The compiler is mainly one-pass, except for
lambdas, which have a variable resolution pass followed by a fast simple
variable liveness analysis that works well for typical array code with limited
branching, and is used to conservatively reuse dead immutable arrays.

Our “bytecode” opcodes are actually of type `int32`. It's originally inspired
from how it's done in [GoAWK](https://benhoyt.com/writings/goawk/). Though
values in GoAWK are a bit different, they are the same size as Goal's (four
words), so stack-handling in both is quite close. 

# Value representation

Go's interface types are a quite good fit for representing the various kinds of
values in a dynamic language like Goal. Most interpreters written in Go have a
`value.go` file where a `Value` interface type is defined. This Value interface
is satisfied by all kinds of values supported by the language (of course
sometimes it's called differently, like Object or whatever :-). Having values
as an interface, as opposed to, say, a sum-type or an union-type in other
languages, has the advantage of allowing easy extensibility, as adding new
kinds of native values just means implementing an interface for a new type.

``` go
// Value is the interface satisfied by all boxed values.
type Value interface {
	// Matches returns true if the value matches another (in the sense of
	// the ~ operator).
	Matches(Value) bool
	// Append appends a unique program representation of the value to dst,
	// and returns the extended buffer.
	Append(ctx *Context, dst []byte) []byte
	// Type returns the name of the value's type. It may be used by Less to
	// sort non-comparable values using lexicographic order.  This means
	// Type should return different values for non-comparable values.
	Type() string
	// LessT returns true if the value should be orderer before the given
	// one. It is used for sorting values, but not for element-wise
	// comparison with < and >. It should produce a strict total order,
	// that is, irreflexive (~x<x), asymmetric (if x<y then ~y<x),
	// transitive, connected (different values are comparable, except
	// NaNs).
	LessT(Value) bool
}
```

That said, Goal's value representation is a bit more complicated than that: the
expected Value interface does exist and is satisfied by all *boxed* values, but
values themselves are represented by the following struct type:

``` go
// V contains a boxed or unboxed value.
type V struct {
	kind valueKind // valInt, valFloat, valBoxed, ...
	uv   int64     // unboxed integer or float value
	bv   Value     // boxed value
}
```

As a result, Goal has unboxed integer and floating point values (by
interpreting the `uv` field as one type or the other depending on the kind
field), as well as unboxed types for built-in functions and lambdas. This makes
scalar code faster and more memory-friendly by keeping numeric atoms in the
stack.  Although the V struct is less compact than a union struct in C (we need
four words), it does perform quite well in practice, as Go is good at
efficiently copying small types up to a few words.

# Primitives

There's nothing really surprising in the way primitives are implemented: they
are represented by the `VariadicFun` type as follows:

``` go
// VariadicFun represents a variadic function.
type VariadicFun func(*Context, []V) V
```

In other words, they take a context and a list of arguments, and they return a
new value. Each variadic function inspects dynamically its arguments types, and
process them accordingly. One gotcha, though: arguments are in reverse order,
because the slice comes from the VM's stack.

Array languages face a major difficulty with primitives, in that most of them
have to handle various different specialized array types, for performance
reasons.

For example, there's a type for arrays of bytes (AB), another for arrays of
64-bit integers (AI), another for arrays of 64-bit floats (AF), and one for
generic arrays (AV), which can contain any values (like arrays of arrays, and
mixed values of any kind).  Goal's user does not have to think about this, and
can treat booleans as 0 and 1 without worring whether they're stored compactly
as a byte or an int64. One can also use 4.0 as an integer (but not 4.5), and
generally perform most operations on numeric types without doing any manual
conversions.  The implementation, though, must handle each case for each
operand, which can quickly become daunting. And there's no really any way,
except for heavy macro usage, to avoid handling all the cases: generics do help
to some extent to abstract some algorithms, but not much to deal with many type
combinations simultaneously, and often specific types have better algorithms
than a generic one (for example sort and search for arrays of byte-sized
integers).

For simple unary functions, a variadic function looks like this:

``` go
// vfSubtract implements the - variadic verb.
func vfSubtract(ctx *Context, args []V) V {
	switch len(args) {
	case 1:
		return negate(args[0])
	case 2:
		return subtract(args[1], args[0])
	default:
		return panicRank("-")
	}
}

// negate returns -x.
func negate(x V) V {
	if x.IsI() {
		return NewI(-x.I())
	}
	if x.IsF() {
		return NewF(-x.F())
	}
	switch xv := x.bv.(type) {
	case S:
		return NewS(strings.TrimRightFunc(string(xv), unicode.IsSpace))
	case *AB:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
			r[i] = -int64(xi)
		}
		return NewAI(r)
	case *AI:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = -xi
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = -xi
		}
		return NewV(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = strings.TrimRightFunc(xi, unicode.IsSpace)
		}
		return NewV(r)
	case *AV:
		return mapAV(xv, negate)
	case *Dict:
		return newDictValues(xv.keys, negate(NewV(xv.values)))
	default:
		return panicType("-x", "x", x)
	}
}
```

The binary case is more intricate, as the number of cases to handle becomes
quadratic in the number of types (among those that can be combined).

Most of the function is quite self-explanatory, except for the `.reuse` method,
which is an internal optimization that returns an array value of same-length
and type which may reuse the original's value intact memory, if it's deemed
reusable by the reference count system for array values. Also, the `mapAV`
function is used for the generic array case, as it works the same in many
monadic primitives.

If performance is not a concern, like for example for the APL-like bignum
calculator [ivy](https://pkg.go.dev/robpike.io/ivy) (also written in Go!), it's
possible to handle this more simply: all arrays are slices of type `[]Value`
where `Value` is an interface, and operations are not vectorized for specific
types, so array operations are deduced easily from operations on atoms.

In our case, some things were even a bit more verbose than they would have been
in C, because in Go type conversions have to be explicit all the time, and
there are no macros, meaning more code duplication for the various numeric
types.  In the end, it was not as bad as expected, and for extreme cases, like
arithmetic primitives, I used some code generation (written in Goal), which is
similar to how they're usually done in C array language interpreters by using
macros.  The somewhat unsightly result is in `arithd.go`.

Other than that, some array primitives do require some algorithmic work, in
particular self-search functions, though there's no way I'm going to explain
this better than [BQN implementation
notes](https://mlochbaum.github.io/BQN/implementation/), so there you go. I did
not optimize as far as BQN, but I picked a few ideas from there.

# Performance

I talked about scalar performance above but did not give any numbers. There's
some benchmarks you can run with `go test -bench .`, and see how they perform
on your system, which is probably faster than my cheap computer.

I'm not sure micro-benchmarks for scalar code are that meaningful, as they
change a lot from one system to another and are not representative of real
world applications. That said, to give an order of magnitude, on my OpenBSD
machine, things like the naive fibonacci (fib 35) function completed faster
than Perl, more or less like Python, though I'm not going to celebrate because
on a Linux machine Python ran more than twice as fast for this same
micro-benchmark (but Perl ran like Goal).

Whatever, I would say that Goal's scalar performance is decent. Quite a few
array programming languages (like J) have done well without that. You're indeed
normally going to be using array primitives on performance sensitive parts, so
it's going to go fast, not like in SIMD or gonum fast in Goal's case, but fast
like in typical Go code. Also, talking about array performace, the refcount
system that allows for amortized constant time append and memory reuse in many
builtins while using immutable arrays has been the major source of bugs during
development: I know what people mean now when they say refcounting stuff is
hard!
