# Goal

[![pkg.go.dev](https://pkg.go.dev/badge/codeberg.org/anaseto/goal.svg)](https://pkg.go.dev/codeberg.org/anaseto/goal)
[![godocs.io](https://godocs.io/codeberg.org/anaseto/goal?status.svg)](https://godocs.io/codeberg.org/anaseto/goal)

Goal is an embeddable array programming language with a bytecode interpreter,
written in Go. It provides both a command line intepreter (that can be used in
the REPL), and a library interface. The core features are mostly there and
tested, so Goal is usable both for writing useful short scripts and playing
with the REPL. User testing and bug reports are welcome!

Like in most array programming languages, Goal's builtins vectorize operations
on immutable arrays, and encourage a functional style for control and data
transformations, supported by a simple dynamic type system with little
abstraction, and mutable variables (but no mutable values).

It's main distinctive features are as follows:

* Syntax inspired mainly from the K language, but with quite a few deviations.
  For example, backquotes produce Go-like raw strings instead of symbols,
  `rx/\s+/` is a compile-time regular expression literal, and there is
  Perl-style string interpolation. On the other side, there are no tacit
  compositions, and digraph operator verbs and adverbs are gone or done
  differently (except for global assignment with ::).
* Primitive semantics are both inspired from the
  [ngn/k](https://codeberg.org/ngn/k) variant of the K language and
  [BQN](https://mlochbaum.github.io/BQN/index.html). For example, group by,
  classify, shifts, windows, binary search and occurrence count take after
  BQN's semantics, but free-form immutable arrays, dictionaries and adverbs
  take after K.
* Unlike in typical array languages, strings are atoms, and common string
  handling functions have been integrated into the primitives, including
  regular expression functions. Primitives acting on whole strings are
  Unicode-aware (like case-folding or Unicode properties in regexps).
* Error handling makes a distinction between fatal errors (panics) and
  recoverable errors which are handled as values.
* Integrated support for csv, json, time handling, and basic math.
* Simple IO: read/write files, run commands/pipes, open filehandles.
* Easily embeddable and extensible in Go.
* Array performance is quite decent, with specialized algorithms depending on
  inputs (type, size, range), and variable liveness analysis that reduces
  cloning by reusing dead immutable arrays (in code with limited branching).
  However, it is not a goal to reach state-of-the-art (no SIMD, and no bit
  booleans, fitting integers in arrays using either uint8 or int64 elements).
+ Scalar performance is typical for a bytecode-compiled interpreter (without
  JIT), somewhat slower than a C bytecode interpreter: value representation is
  less compact than how it could be done in C, but Goal does have unboxed
  integers and floats.

If this list is not enough to satisfy your curiosity, there's also a
[Why.md](docs/Why.md) text for you. You can also read the [Credits.md](Credits.md)
to know about main inspiration sources for the language. Last, but not least,
there are some [implementation notes](docs/Implementation.md) too.

# Install

To install the command line interpreter, first do the following:

* Install the [go compiler](https://golang.org/).
* Add `$(go env GOPATH)/bin` to your `$PATH` (for example `export PATH="$PATH:$(go env GOPATH)/bin"`).

Then you can build the intepreter with:

	go install ./cmd/goal

Alternatively, you may type `go build -o /path/to/bin/goal ./cmd/goal` to put
the resulting binary in a custom location in your $PATH.

The `goal` command should now be available. Type `goal --help` for command-line
usage.

Typing just `goal` opens the REPL. For a better experience using the REPL (to
get typical keyboard shortcuts), you can install the readline wrapper `rlwrap`
program (available as a package in most systems), and then use instead `rlwrap
goal`.

# Links

* [Goal docs](https://anaseto.codeberg.page/goal/) : work-in-progress
  documentation for Goal.
* [vim-goal](https://codeberg.org/anaseto/vim-goal) : vim files for Goal.
* [APL Farm](https://matrix.to/#/#aplfarm:matrix.org) : chat about array
  languages.

# Documentation

In addition to the work-in-progress [documentation
website](https://anaseto.codeberg.page/goal/), you might be interested in the
[Changelog](Changes.md) changes between releases. The REPL help is also
available in text form here at [docs/help.txt](docs/help.txt).

# Examples

A few short examples can be found in the `examples` and `testdata/scripts`
directory. Because the latter are used for testing, they come along an expected
result after a `/RESULT:` comment.

Also, various code generation scripts in the toplevel `scripts` directory are
written in Goal.
