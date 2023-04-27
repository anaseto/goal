# v? ?

+ New `"b"$s` for converting a string to and from array of bytes.
+ New `"c"$s` for converting a string to and from array of code points, and make
  `"i"$s` now be parse int, and "n"$s only parse number (floats).
+ New `"s"$y` form that formats in a default way non-strings elements.
+ New `0i` value, representing the minimum representable integer.
+ Return `"i"` for `@i`, not `"n"`, because they're actually different types, even
  though implicit conversions are made when possible.
+ `$X` returns now stranding representation for mixed lists of strings and
  numbers, without parens.
+ Add ¿ as symbol alternative for in and firsts, the boolean primitives
  counterparts of ?.
+ Rename atan2 into atan, which is now dyadic and accepts either one or two
  arguments (same simplification as Lua 5.3 did). (breaking change)
+ Improvements in sorting of integers, depending on the range, using either
  counting sort (`^I` for small-range) or radix sort (for `^I`, `<I`, and
  `>I` when `I` fits into a `[]int8`, `[]int16` or `[]int32` slice).
+ Recognize `x:op x` assignement as potential in-place operation for global
  variables too (previously only `x op:y` form was recognized for global
  variables).
+ Fix missing integer overflow check in odometer.

# v0.15.0 2023-04-21

This release makes quite a few significant changes and improvements.

+ Rename `bytes s` into `&s`, because getting the length in bytes of a string
  is quite a fundamental operation, even if it's not used that often (breaking
  change).
+ New take/repeat `i@y` equivalent to old `i#y`, that now is take/pad. Both give
  same results unless `i>#y` or `(-i)>#y`. A simple way to pad arrays with
  zero-values (now properly based on first element type, or default type) was
  missing.  `@` in kind of a circle, which is a good mnemonic for the cyclic
  behavior (which also feels to me more like an apply action than just padding
  would). Also, because new `X#d` does padding for new keys, it's expected that
  `i#y` does too for out of bound indices. (breaking change)
+ Make `i^Y`, `i^s`, `i!s`, `i!Y` use `i` as the length of the result, and use
  `-i` for the old behavior (breaking change). This was suggested by @Marshall
  on the aplfarm matrix channel.
+ Make `=I` return index-counting `#'=I`, because the former is not used often
  in this form, and can be obtained with group keys `=(!#I)!I` and group by
  `{!#x}=I`. (minor breaking change)
+ Improve state of things with fill/pad values of generic arrays: now we use
  the type of the first element to determine the zero value if it's not empty,
  and we use () otherwise (which is often the desired default for generic
  arrays). Previously fills where only really usable in numeric and string
  arrays.
+ Rename fields from v0.13 into `!s`, and add new `=s` lines (like `"\n"\` but
  handles `"\r\n"` endings too). The mnemonics for the new name is that fields
  is mainly used to break text vertically, while `=s` only breaks it
  horizontally (and `=` represents two lines).  (breaking change)
+ Rename utf8.valid into utf8, and remove utf8.rcount, because it's the same as
  `-1+""#x` with same underlying counting code. (breaking change)
+ Remove delete `X_i` and `s_i`, because they don't follow the common pattern,
  are not that useful (outside of golfing maybe), and `_` is already quite
  polysemic. (minor breaking change)
+ Implement w/o keys `X^d` and w/ keys `X#d`.
+ Implement `+d` as swap of keys and values. The reason for this is that Goal
  does not have tables, and they are not planned, as making them useful would
  require a lot of work (including some kind of query methods).
+ Implement `d.y` for `1<#y`, and `X@d`.
+ New `.X` self-dict form, equivalent to `{x!x}`.
+ New `-s` form triming trailing spaces, as defined by Unicode.
+ Implement `@[f1;x;f2]`, like `.[f1;y;f2]` but doing `f1@x` instead of `f1 .
  y`.
+ Improve at depth indexing of mixed dicts and arrays.
+ `<d` and `>d` return now sorted dicts, instead of just the keys, and `^d`
  sorts `d` by its keys.
+ Fix default rank of `s/` `s\` `I/` `I\` (used in the case they would be
  followed by a fold or scan, which would be quite rare).
+ When an in-place assignement operation panics, clear the variable (instead of
  producing random garbage).

# v0.14.0 2023-04-08

+ New `?(-i)` form returning normal distribution.
+ Make rt.ofs and rt.prec return the previous value.
+ New -q command-line option disabling echo.
+ New examples/ directory with some example scripts.
+ csv now accepts non-string fields (it stringifies fields if necessary).
+ Do not search for script filename when -e option is provided.
+ Fix possible refcount bug with shallowClone and join of generic array.

# v0.13.0 2023-04-03

+ New rt.ofs runtime builtin for setting the output field separator for string
  lists used in print S and "$S", defaulting to space.
+ Export OFS and Prec parameters in Context API.
+ Allow use of 3-arg ? as function too, by putting parens around, in which
  case, the arguments are evaluated like for any function (this can be useful
  in special cases where avoiding branching improves liveness analysis and
  performance).
+ Make the new `=s` form of last release work with the `#'=x` special case.
+ Fixed incorrect newline termination in x say S form.

# v0.12.0 2023-04-02

+ New `=s` fields form, similar to `rx/[\n \t]+/\` but splits on any kind of unicode
  space.
+ New "zone" command for time.
+ Simplify math builtins: atan2 in place of acos, asin, atan, and remove tan
  (provide same functions as AWK).

# v0.11.0 2023-03-30

+ New `i!i` range form.
+ New `""^s` unicode-aware trim spaces form.
+ New `i?Y` and `(-i)?Y` forms.
+ New `x utf8.valid s` form for replacing invalid byte sequences.
+ New `i?Y` roll array and `(-i)?Y` deal array forms.
+ More permissive projection application: for example `+[][2;3]` is valid, even
  though `+[]` has rank 1 when used in adverbial contexts that make use of
  function rank.
+ Disallow space before adverb application. For `'` this means that usage is
  reserved for early return on error, and for `\` it has no meaning yet.
+ For consistency, make run s form return its standard output too, and return
  a dictionary in error case containing exit code, message and output.
+ More consistent error messages.
+ Fix some unhandled cases of `.s` with lambda return.

# v0.10.0 2023-03-26

+ Renamings and cleanups in the library interface, preparing for a stable
  and documented API.
+ Rename goal builtin into separate rt.CMD builtins.
+ Improvements in import builtin supporting GOALLIB environment variable,
  both use of extension or not, and importing several files at once.
+ Improvements and fixes in help, with better examples and a new short FAQ.
+ Improvements in REPL handling of multi-line quoting constructs.
+ The json builtin can now handle several JSON strings at once.
+ Error location improvements in value application.
+ Fix panic-case handling in eval, in particular error location.

# v0.9.0 2023-03-23

+ New json builtin for parsing JSON strings.
+ New chdir builtin for changing working directory.
+ Implement i!s colsplit form for strings.
+ New `goal["time";s;n]` and `goal["time";f;x;n]` forms for timing code.
+ New `d[]` form equivalent to `.d` (for dicts).
+ New help feature: extract help for specific op.
+ Fixed refcount bug in `i f\y` form.
+ More tests, and various refactorings, and documentation improvements.

# v0.8.0 2023-03-13

+ lists `(expr1;expr2)` are now evaluated from left to right. This avoids a list
  reversal in the implementation, and is also more intuitive in multi-line
  lists. Function arguments in binary and index application are still evaluated
  from right to left as always.  Now lists and sequences are similar, except
  that the first construct collects all values, and the second returns the last
  one only.
+ In `rq//` raw quoting introduced in last release, allow escaping delimiter by
  doubling it, as a simplification of previous rule. Note that the old
  backquoted raw strings behave the same as always and do not have any way of
  writing a literal backquote.
+ `,/()` now returns ()
+ `"n"$y` returns 0n instead of error for incorrect syntax, as it is more
  convenient by ensuring a numerical result.
+ Various bug fixes and more op/ special code.

# v0.7.0 2023-03-02

Highlights:

+ String interpolation in "" strings, as well as the new `qq//` Perl-style
  quoting form with custom delimiter (potentially breaking if strings contained
  character `$`, which now needs to be escaped).
+ New `rq//` quoting form with custom delimiter for raw strings. Allow custom
  delimiter for `rx//` form too.
+ New x run s form, piping string x to command s, and returning its own stdout.
+ Optimize `#'=` so that icount can be finally removed.
+ Various small improvements and fixes in language's help.

# v0.6.0 2023-02-26

This release has various breaking changes in the core language.  They should be
more rare from now on (but still possible if important until we reach 1.0), and
they will at least be documented.  Some changes might still happen in the
embedding API.

Highlights:

* General IO with open (including pipes), read, close, flush, and env builtins,
  as well as STDIN, STDOUT, STDERR globals containing the standard filehandles.
* slurp s became read s, complementing the new read h form.
* New “run” builtin for running commands without shell.
* Dict support is now implemented in all builtins where it makes sense (though
  the help only mentions most interesting cases).
* New forms `x_i` and `s_i` (delete)
* `i$s` is now pad, and `i!x` is the old `i$s`.
* Added special code for some `op\` and `op'` forms.
* Many improvements in help.
* Added goal builtin for runtime state inspection and configuration, like
  floating point formatting precision. Also, the seed builtin was removed, and
  uses the goal builtin with a "seed" argument.
* Added aliases « and » for shift and shiftr.
