# v? ?

* Format floating point integer values with a trailing `.0`, so that the
  numeric type can be retrieved if parsed again.
* Fix stability of `>X` and `>d`.
* Fix ascending flag preservation for `I+i`, `i+I`, and `I-i` in case of
  integer overflow by checking that first element is still smaller than last.

# v0.19.0 2023-06-01

* Flip `+x` now extends array elements (like in take `i#`) to maximum length if
  needed, so that it works with any kind of ragged array.
* There were a few changes in type names in the Go library interface, for
  better consistency, like renaming `Value` into `BV` (boxed value) for
  consistency with `V` (unboxed or boxed value), and `D` for dicts. Also, the
  Array interface is now public.
* Simplification of reference count system, so that it's less error-prone and
  easier to maintain. Because of the GC, we use it only for optimization
  purposes, so we can take advantage of that. We limit reference counting to
  the first level of nesting of generic arrays, marking deeply nested data as
  definitively immutable. This avoids the need for pointers, and simplifies
  handling of shared data. It makes incrementing and decrementing reference
  counts cheaper in the common cases, and most of the time doesn't affect
  negatively performance, except when making many small amends to nested data.
* Various other code refactorings that should facilitate maintenance in the
  long run.

# v0.18.0 2023-05-21

* Improve `d op d` semantics for uncommon or duplicate keys, so that it's in
  line with amend dict.
* Support regexp projections (like for any function).
* Fix regression refcount bug in converges `f\x`.
* Fix refcount bug for out of bounds dict indexing.
* Optimize seeded `x op/y` and `x op\y` for various common op (only non-seeded
  variants had special code before), as well as `~'` and `,//`.
* Optimize small-range computation for sorted inputs in search functions.
* Added some debug and dev purpose tools for refcounting issues and stack
  handling.

# v0.17.1 2023-05-16

* Fix CRLF handling in Goal scripts.

# v0.17.0 2023-05-14

* Rename weed out to `f^y`, related to `X^y`, in the same way that `f#y` is
  related to `X#y`. (breaking change) Now `f_y` is “cut where”, same as `(&f
  y)_y`, but `f` has to return an array of integers.
* Rename `ceil` into `uc`, which can be read both as upper/ceil (for both of
  its meanings on strings and numbers), or just upper case (like `uc` in Perl).
  It felt unnatural to have a name representing only one of its meanings, but
  it was natural to have a single built-in for both (like for `_`), hence a
  name with double-reading. (breaking change)
* Now where `&` allows negative inputs, and simply treats them as zero.
* Allow referencing main namespace from an imported file using `main.` as
  prefix (still thinking about alternative prefixes, like just `m.`, but this
  shouldn't come often so a longer form might be better).
* Implement dyadic version for shell too, and make it inherit STDIN too. It
  works now exactly like run, but through /bin/sh.
* Optimize group by for sorted indices, and `X?Y` (find) and `X¿Y` (in) for
  long generic arrays too (hashing).
* Make `f'[d;y;…]` and `f\[x;d;z;…]` return dictionary too in n-arg case.
* Fix syntax regression when projecting `@` due to missing check for dyadic
  case when optimizing away `@` and replacing it with a single opcode.
* Fix bug in `s?r` when there is no match: it returned
  (start-offset;end-offset) instead of (offset;len) in such case.
* Fix refcount bug in each and scan adverbs when function returned a mutated
  global.

# v0.16.0 2023-05-05

* Make out-indexing valid and return the zero value of the array or dict, in
  the same way as padding does (0s for numeric arrays, empty strings for string
  arrays, and zero value based on first element type for generic arrays, or
  an empty generic array if it's empty). In particular, outdexing results
  returned by group will return an empty array as expected.
* Now that `X@i` does pad on outdexing, invert `i@Y` and `i#Y` swap from
  previous realease to keep more compatibility with K. It's also now quite
  natural that `i@Y` is the same as `Y@!i` (except for being more efficient by
  not actually generating indices). Also, new mnemonic: `@` looks like a bit
  like a zero, so it pads.
* Make `i!n` mod/div, and remove `mod` keyword, both for better compatibility
  with ngn/k, and because only `i!I` is easy to vectorize (for example for
  powers of two, and to produce a result with smaller type for small i). As a
  result, cut shape is now `i$y` (used in `J` too for reshape), and string
  padding is done with `i!s` (though it can be done too with new sprintf-like
  format `s$y`).  Also, span `i$i` replaces range `i!i`, which was not very
  composable (albeit convenient). (breaking changes)
* New format `s$y` form that provides classic sprintf-like functionality, for
  when more advanced formatting than just space padding is useful. Formatting
  works element-wise when `s` contains a single formatting word, like `"%03d"`,
  and list-wise when it contains more, like `"%s=%.4f"` (expecting `2~#y`).
* New `"b"$s` for converting a string to and from array of bytes.
* New `"c"$s` for converting a string to and from array of code points, and make
  old `"i"$s` now just be parse/cast to int, and `"n"$s` only parse/cast to
  number (floats).
* New `"s"$y` form that formats in a default way non-strings element-wise,
  instead of just the whole `y` like in `$y`. Also, it leaves strings as-is,
  unlike `$y` which quotes them with escapes.
* Return `"i"` for `@i`, not `"n"`, because integers and floats are different
  types, even though implicit conversions are made when possible.
* New `0i` value, representing the minimum representable integer. It's used in
  `"i"$s` when parsing an integer with invalid syntax, as well as in the
  special shortcut `0i?Y` for doing a shuffle. We use `0i` for consistency,
  instead of `0N`, because `i` is the name of the integer type.
* `$X` returns now stranding representation for mixed lists of strings and
  numbers, without parens.
* Add `¿` as symbol alternative for in and firsts, the boolean primitive
  counterparts of `?`.
* Rename `atan2` into `atan`, which now accepts either one or two arguments
  (same simplification as Lua 5.3 did). (breaking change)
* More consistent results in `op/x` for empty list `x`. Now we use the default
  zero value for the given list type, except for a few special reductions where
  a different neutral element is clear (like `*/`, `&/`, and `|/` for
  non-generic numeric lists).
* Make `X#d` follow same semantics as `X^d` (except for being the negation
  of it), returning a dictionary derived from `d`, preserving duplicate keys,
  but with keys not in `X` removed. In particular, it does not add new keys to
  `d` anymore, so dictionary padding should either be done with merge or
  individually to its arrays.
* Implement !i for negative integers too.
* Improvements in sorting of integers, depending on the range, using either
  counting sort (`^I` for small-range) or radix sort (for `^I`, `<I`, and
  `>I` when `I` fits into a `[]int8`, `[]int16` or `[]int32` slice).
* Arrays of small integers (0-255) are now represented as arrays of bytes,
  saving memory (in particular when converting a string to bytes, or even
  codepoints if the string is ASCII), and facilitating small-range
  optimizations for searching and sorting. Previously, only booleans where
  stored more compactly.
* Recognize `x:op x` assignement as potential in-place operation for global
  variables too (previously only `x op:y` form was recognized for global
  variables).
* Fix missing integer overflow check in odometer, a regression in boolean
  grade (maybe posterior to 0.15.0), as well as other minor issues.

# v0.15.0 2023-04-21

This release makes quite a few significant changes and improvements.

* Rename `bytes s` into `&s`, because getting the length in bytes of a string
  is quite a fundamental operation, even if it's not used that often (breaking
  change).
* New take/repeat `i@y` equivalent to old `i#y`, that now is take/pad. Both give
  same results unless `i>#y` or `(-i)>#y`. A simple way to pad arrays with
  zero-values (now properly based on first element type, or default type) was
  missing.  `@` in kind of a circle, which is a good mnemonic for the cyclic
  behavior (which also feels to me more like an apply action than just padding
  would). Also, because new `X#d` does padding for new keys, it's expected that
  `i#y` does too for out of bound indices. (breaking change: Actually `i#y`
  swapped to K normal behavior in next release, so that `i@y` does padding)
* Make `i^Y`, `i^s`, `i!s`, `i!Y` use `i` as the length of the result, and use
  `-i` for the old behavior (breaking change). This was suggested by @Marshall
  on the aplfarm matrix channel.
* Make `=I` return index-counting `#'=I`, because the former is not used often
  in this form, and can be obtained with group keys `=(!#I)!I` and group by
  `{!#x}=I`. (minor breaking change)
* Improve state of things with fill/pad values of generic arrays: now we use
  the type of the first element to determine the zero value if it's not empty,
  and we use () otherwise (which is often the desired default for generic
  arrays). Previously fills where only really usable in numeric and string
  arrays.
* Rename fields from v0.13 into `!s`, and add new `=s` lines (like `"\n"\` but
  handles `"\r\n"` endings too). The mnemonics for the new name is that fields
  is mainly used to break text vertically, while `=s` only breaks it
  horizontally (and `=` represents two lines).  (breaking change)
* Rename utf8.valid into utf8, and remove utf8.rcount, because it's the same as
  `-1+""#x` with same underlying counting code. (breaking change)
* Remove delete `X_i` and `s_i`, because they don't follow the common pattern,
  are not that useful (outside of golfing maybe), and `_` is already quite
  polysemic. (minor breaking change)
* Implement w/o keys `X^d` and w/ keys `X#d`.
* Implement `+d` as swap of keys and values. The reason for this is that Goal
  does not have tables, and they are not planned, as making them useful would
  require a lot of work (including some kind of query methods).
* Implement `d.y` for `1<#y`, and `X@d`.
* New `.X` self-dict form, equivalent to `{x!x}`.
* New `-s` form triming trailing spaces, as defined by Unicode.
* Implement `@[f1;x;f2]`, like `.[f1;y;f2]` but doing `f1@x` instead of `f1 .
  y`.
* Improve at depth indexing of mixed dicts and arrays.
* `<d` and `>d` return now sorted dicts, instead of just the keys, and `^d`
  sorts `d` by its keys.
* Fix default rank of `s/` `s\` `I/` `I\` (used in the case they would be
  followed by a fold or scan, which would be quite rare).
* When an in-place assignement operation panics, clear the variable (instead of
  producing random garbage).

# v0.14.0 2023-04-08

* New `?(-i)` form returning normal distribution.
* Make rt.ofs and rt.prec return the previous value.
* New -q command-line option disabling echo.
* New examples/ directory with some example scripts.
* csv now accepts non-string fields (it stringifies fields if necessary).
* Do not search for script filename when -e option is provided.
* Fix possible refcount bug with shallowClone and join of generic array.

# v0.13.0 2023-04-03

* New rt.ofs runtime builtin for setting the output field separator for string
  lists used in print S and "$S", defaulting to space.
* Export OFS and Prec parameters in Context API.
* Allow use of 3-arg ? as function too, by putting parens around, in which
  case, the arguments are evaluated like for any function (this can be useful
  in special cases where avoiding branching improves liveness analysis and
  performance).
* Make the new `=s` form of last release work with the `#'=x` special case.
* Fixed incorrect newline termination in x say S form.

# v0.12.0 2023-04-02

* New `=s` fields form, similar to `rx/[\n \t]+/\` but splits on any kind of unicode
  space.
* New "zone" command for time.
* Simplify math builtins: atan2 in place of acos, asin, atan, and remove tan
  (provide same functions as AWK).

# v0.11.0 2023-03-30

* New `i!i` range form.
* New `""^s` unicode-aware trim spaces form.
* New `i?Y` and `(-i)?Y` forms.
* New `x utf8.valid s` form for replacing invalid byte sequences.
* New `i?Y` roll array and `(-i)?Y` deal array forms.
* More permissive projection application: for example `+[][2;3]` is valid, even
  though `+[]` has rank 1 when used in adverbial contexts that make use of
  function rank.
* Disallow space before adverb application. For `'` this means that usage is
  reserved for early return on error, and for `\` it has no meaning yet.
* For consistency, make run s form return its standard output too, and return
  a dictionary in error case containing exit code, message and output.
* More consistent error messages.
* Fix some unhandled cases of `.s` with lambda return.

# v0.10.0 2023-03-26

* Renamings and cleanups in the library interface, preparing for a stable
  and documented API.
* Rename goal builtin into separate rt.CMD builtins.
* Improvements in import builtin supporting GOALLIB environment variable,
  both use of extension or not, and importing several files at once.
* Improvements and fixes in help, with better examples and a new short FAQ.
* Improvements in REPL handling of multi-line quoting constructs.
* The json builtin can now handle several JSON strings at once.
* Error location improvements in value application.
* Fix panic-case handling in eval, in particular error location.

# v0.9.0 2023-03-23

* New json builtin for parsing JSON strings.
* New chdir builtin for changing working directory.
* Implement i!s colsplit form for strings.
* New `goal["time";s;n]` and `goal["time";f;x;n]` forms for timing code.
* New `d[]` form equivalent to `.d` (for dicts).
* New help feature: extract help for specific op.
* Fixed refcount bug in `i f\y` form.
* More tests, and various refactorings, and documentation improvements.

# v0.8.0 2023-03-13

* lists `(expr1;expr2)` are now evaluated from left to right. This avoids a list
  reversal in the implementation, and is also more intuitive in multi-line
  lists. Function arguments in binary and index application are still evaluated
  from right to left as always.  Now lists and sequences are similar, except
  that the first construct collects all values, and the second returns the last
  one only.
* In `rq//` raw quoting introduced in last release, allow escaping delimiter by
  doubling it, as a simplification of previous rule. Note that the old
  backquoted raw strings behave the same as always and do not have any way of
  writing a literal backquote.
* `,/()` now returns ()
* `"n"$y` returns 0n instead of error for incorrect syntax, as it is more
  convenient by ensuring a numerical result.
* Various bug fixes and more op/ special code.

# v0.7.0 2023-03-02

Highlights:

* String interpolation in "" strings, as well as the new `qq//` Perl-style
  quoting form with custom delimiter (potentially breaking if strings contained
  character `$`, which now needs to be escaped).
* New `rq//` quoting form with custom delimiter for raw strings. Allow custom
  delimiter for `rx//` form too.
* New x run s form, piping string x to command s, and returning its own stdout.
* Optimize `#'=` so that icount can be finally removed.
* Various small improvements and fixes in language's help.

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
