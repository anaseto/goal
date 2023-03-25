# FAQ

## Which builtins support dictionaries?

Builtins support dictionary values whenever it makes sense (like for arithmetic
builtins), and just apply to their value arrays like they would for a normal
array (but returning the keys along when sensible, and using upsert semantics
for uncommon keys). The help only mentions cases that have some special
dictionary-specific semantics.

## Which builtins generalize operations to arrays element-wise?

This is the case of arithmetic builtins, but also of most other builtins when
the generalization is useful and does not conflict with other usages (a few
builtins where it's not very useful have yet to be generalized).

## Why does 0n not match 0n?

This follows the usual floating-point arithmetic conventions. Use the `nan`
builtin to handle NaN values.

## Why does division by zero return 0w?

Because traditionally array languages return infinity in such case.
