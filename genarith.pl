#!/usr/bin/env perl

use strict;
use warnings;
use v5.28;
use utf8;
use open qw(:std :utf8);

my %dyads = (
    Add => "+",
    Subtract => "-",
    Multiply => "*",
    Divide => "/", # TODO: catch by zero division (Inf)
);

print <<EOS;
package main
EOS

for my $k (sort keys %dyads) {
    doFun($k, $dyads{$k})
}

sub doFun {
    my ($name, $op) = @_;
    print <<EOS;

// ${name} returns w${op}x.
func ${name}(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return ${name}I(w, x)
	case F:
		return ${name}F(w, x)
	case AI:
		return ${name}AI(w, x)
	case AF:
		return ${name}AF(w, x)
	case AO:
		return ${name}AO(w, x)
	}
	// TODO
	return badtype("${op}")
}

func ${name}I(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i ${op} x
	case F:
		return F(i) ${op} x
	case AI:
		return ${name}AII(x, i)
	case AF:
		return ${name}AFI(x, i)
	}
	// TODO
	return badtype("${op}")
}

func ${name}F(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f ${op} F(x)
	case F:
		return f ${op} x
	case AI:
		return ${name}AIF(x, f)
	case AF:
		return ${name}AFF(x, f)
	}
	// TODO
	return badtype("${op}")
}

func ${name}AI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return ${name}AII(a, x)
	case F:
		return ${name}AIF(a, x)
	case AI:
		return ${name}AIAI(a, x)
	case AF:
		return ${name}AIAF(a, x)
	case AO:
		return ${name}AOAI(x, a)
	}
	// TODO
	return badtype("${op}")
}

func ${name}AF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return ${name}AFI(a, x)
	case F:
		return ${name}AFF(a, x)
	case AI:
		return ${name}AIAF(x, a)
	case AF:
		return ${name}AFAF(a, x)
	case AO:
		return ${name}AOAF(x, a)
	}
	// TODO
	return badtype("${op}")
}

func ${name}AO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return ${name}AOI(a, x)
	case F:
		return ${name}AOF(a, x)
	case AI:
		return ${name}AOAI(a, x)
	case AF:
		return ${name}AOAF(a, x)
	case AO:
		return ${name}AOAO(a, x)
	}
	// TODO
	return badtype("${op}")
}

func ${name}AII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] ${op} i
	}
	return r
}

func ${name}AIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) ${op} f
	}
	return r
}

func ${name}AFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] ${op} F(i)
	}
	return r
}

func ${name}AFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] ${op} f
	}
	return r
}

func ${name}AIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] ${op} a2[j]
	}
	return r
}

func ${name}AIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) ${op} a2[j]
	}
	return r
}

func ${name}AFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] ${op} a2[j]
	}
	return r
}

func ${name}AOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = ${name}I(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func ${name}AOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = ${name}F(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func ${name}AOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = ${name}I(a2[j], a1[j])
	}
	return r
}

func ${name}AOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = ${name}F(a2[j], a1[j])
	}
	return r
}

func ${name}AOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("${op}")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = ${name}(a2[j], a1[j])
	}
	return r
}
EOS
}
