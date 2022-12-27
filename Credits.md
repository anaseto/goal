Goal made use of many inspirations sources both for design and implementation.

Language design was greatly inspired by both K (for syntax and basic
primitives) and [BQN](https://mlochbaum.github.io/BQN/index.html) (quite a few
interesting primitives). I was thinking of Perl and Raku when adding regexp
literals (also slurp, say), even though the backing implementation is Go's.
There is some inspiration from the implementation language, Go: raw strings
using backquotes, as well as using the same semantics and syntax for number and
string literals.

I wrote the bytecode implementation after reading the one for
[goawk](https://benhoyt.com/writings/goawk/), and it still shows. I wrote the
scanner after reading [ivy's](https://pkg.go.dev/robpike.io/ivy). Vim syntax
highlighting is based on [ngn/k's](https://codeberg.org/ngn/k).

Be sure to check out all those great projects if you haven't yet!
