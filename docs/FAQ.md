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

## Why does `0n` not match `0n`?

This follows the usual floating-point arithmetic conventions. Use the `nan`
builtin to handle NaN values.

## Why does division by zero return `0w`?

Because traditionally array languages return infinity in such case.

## Why do `-2+3` and `- 2+3` give different results?

In the first case, `-2` is parsed as a single token. In the second case, the
`-` represents the minus builtin.

# CAVEATS

## Number implicit conversions and overflow

Even though the type builtin `@` returns `"n"` for all numbers, Goal has both
64-bit integer and floating point numbers. Builtins convert from one to another
when possible, so most applications do not have to care about this.

From integer to float, this means that big integers might be approximated after
a conversion. From float to integer, if the float is too big to be represented
or is NaN, it will not be considered an integer. Also, operations on integer
operands can overflow, as defined by two's complement integer overflow.
