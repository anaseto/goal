# v0.6.0 2023-02-26

This release should be the last with various breaking changes in the core
language.  They should be more rare from now on (but still possible if
important until we reach 1.0).  Some changes might still happen in the
embedding API.

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
