# FAQ

## Which builtins support dictionaries?

Builtins support dictionary values whenever it makes sense (like for arithmetic
builtins), and just apply to their value arrays like they would for a normal
array (but returning the keys along when sensible using upsert semantics, and
appending uncommon keys as-is). The help only mentions cases that have some
special dictionary-specific semantics.

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

## Implicit numeric type conversions and overflow

Goal has both 64-bit integer and floating point numbers, whoses types are `"i"`
and `"n"` as returned by `@` respectively. Builtins convert from one to another
whenever possible, so most applications do not have to care about this
distinction.

Conversion from integer to float means that big integers might be approximated.
From float to integer, if the float is too big to be represented or is NaN, it
will not be considered an integer by primitives that want an integer. Also,
operations on integer operands can overflow, as defined by two's complement
integer overflow.
