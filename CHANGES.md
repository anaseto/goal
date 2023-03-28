# v? ?

+ New i!i range form.
+ New ""^s unicode-aware trim spaces form.
+ More permissive projection application: for example `+[][2;3]` is valid, even
  though `+[]` has rank 1 when used in adverbial contexts that make use of
  function rank.
+ For consistency, make run s form returns its standard output too.

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
+ New goal["time";s;n] and goal["time";f;x;n] forms for timing code.
+ New d[] form equivalent to .d (for dicts).
+ New help feature: extract help for specific op.
+ Fixed refcount bug in i f\y form.
+ More tests, and various refactorings, and documentation improvements.

# v0.8.0 2023-03-13

+ lists (expr1;expr2) are now evaluated from left to right. This avoids a list
  reversal in the implementation, and is also more intuitive in multi-line
  lists. Function arguments in binary and index application are still evaluated
  from right to left as always.  Now lists and sequences are similar, except
  that the first construct collects all values, and the second returns the last
  one only.
+ In rq// raw quoting introduced in last release, allow escaping delimiter by
  doubling it, as a simplification of previous rule. Note that the old
  backquoted raw strings behave the same as always and do not have any way of
  writing a literal backquote.
+ ,/() now returns ()
+ "n"$y returns 0n instead of error for incorrect syntax, as it is more
  convenient by ensuring a numerical result.
+ Various bug fixes and more op/ special code.

# v0.7.0 2023-03-02

Highlights:

+ String interpolation in "" strings, as well as the new qq// Perl-style
  quoting form with custom delimiter (potentially breaking if strings contained
  character `$`, which now needs to be escaped).
+ New rq// quoting form with custom delimiter for raw strings. Allow custom
  delimiter for rx// form too.
+ New x run s form, piping string x to command s, and returning its own stdout.
+ Optimize #'= so that icount can be finally removed.
+ Various small improvements and fixes in language's help.

# v0.6.0 2023-02-26

This release has various breaking changes in the core language.  They should be
more rare from now on (but still possible if important until we reach 1.0).
Some changes might still happen in the embedding API.

Highlights:

* General IO with open (including pipes), read, close, flush, and env builtins,
  as well as STDIN, STDOUT, STDERR globals containing the standard filehandles.
* slurp s became read s, complementing the new read h form.
* New “run” builtin for running commands without shell.
* Dict support is now implemented in all builtins where it makes sense (though
  the help only mentions most interesting cases).
* New forms `x_i` and `s_i` (delete)
* i$s is now pad, and i!x is the old i$s.
* Added special code for some op\ and op' forms.
* Many improvements in help.
* Added goal builtin for runtime state inspection and configuration, like
  floating point formatting precision. Also, the seed builtin was removed, and
  uses the goal builtin with a "seed" argument.
* Added aliases « and » for shift and shiftr.
