#!/usr/bin/env perl

use strict;
use warnings;
use v5.28;

my %dyads = (
    Equal =>  {
        B_B => ["w == x", "B"],
        B_I => ["B2I(w) == x", "B"],
        B_F => ["B2F(w) == x", "B"],
        I_B => ["w == B2I(x)", "B"],
        I_I => ["w == x", "B"],
        I_F => ["F(w) == x", "B"],
        F_B => ["w == B2F(x)", "B"],
        F_I => ["w == F(x)", "B"],
        F_F => ["w == x", "B"],
        S_S => ["w == x", "B"],
    },
    NotEqual =>  {
        B_B => ["w != x", "B"],
        B_I => ["B2I(w) != x", "B"],
        B_F => ["B2F(w) != x", "B"],
        I_B => ["w != B2I(x)", "B"],
        I_I => ["w != x", "B"],
        I_F => ["F(w) != x", "B"],
        F_B => ["w != B2F(x)", "B"],
        F_I => ["w != F(x)", "B"],
        F_F => ["w != x", "B"],
        S_S => ["w != x", "B"],
    },
    Add =>  {
        B_B => ["B2I(w) + B2I(x)", "I"],
        B_I => ["B2I(w) + x", "I"],
        B_F => ["B2F(w) + x", "F"],
        I_B => ["w + B2I(x)", "I"],
        I_I => ["w + x", "I"],
        I_F => ["F(w) + x", "F"],
        F_B => ["w + B2F(x)", "F"],
        F_I => ["w + F(x)", "F"],
        F_F => ["w + x", "F"],
    },
    Subtract =>  {
        B_B => ["B2I(w) - B2I(x)", "I"],
        B_I => ["B2I(w) - x", "I"],
        B_F => ["B2F(w) - x", "F"],
        I_B => ["w - B2I(x)", "I"],
        I_I => ["w - x", "I"],
        I_F => ["F(w) - x", "F"],
        F_B => ["w - B2F(x)", "F"],
        F_I => ["w - F(x)", "F"],
        F_F => ["w - x", "F"],
    },
    Multiply =>  {
        B_B => ["B2I(w) * B2I(x)", "I"],
        B_I => ["B2I(w) * x", "I"],
        B_F => ["B2F(w) * x", "F"],
        I_B => ["w * B2I(x)", "I"],
        I_I => ["w * x", "I"],
        I_F => ["F(w) * x", "F"],
        F_B => ["w * B2F(x)", "F"],
        F_I => ["w * F(x)", "F"],
        F_F => ["w * x", "F"],
    },
    Divide =>  {
        B_B => ["divide(B2F(w), B2F(x))", "F"],
        B_I => ["divide(B2F(w), F(x))", "F"],
        B_F => ["divide(B2F(w), x)", "F"],
        I_B => ["divide(F(w), B2F(x))", "F"],
        I_I => ["divide(F(w), F(x))", "F"],
        I_F => ["divide(F(w), x)", "F"],
        F_B => ["divide(w, B2F(x))", "F"],
        F_I => ["divide(w, F(x))", "F"],
        F_F => ["divide(w, x)", "F"],
    },
    #Lesser => "<",
    #LesserEq => "<=",
    #Greater => ">",
    #GreaterEq => ">=",
);

print <<EOS;
package main

EOS

genOp("Equal", "=");
genOp("Add", "+");
genOp("Subtract", "-");
genOp("Multiply", "×");
genOp("Divide", "÷");

sub genOp {
    my ($name, $op) = @_;
    my $cases = $dyads{$name};
    my %types = map { $_=~/(\w)_/; $1 => 1 } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}(w, x Object) Object {
	switch w := w.(type) {
EOS
    for my $t (sort keys %types) {
        print $out <<EOS;
	case $t:
		return ${name}${t}O(w, x)
EOS
    }
    for my $t (sort keys %types) {
        print $out <<EOS;
	case A$t:
		return ${name}A${t}O(w, x)
EOS
    }
    print $out <<EOS;
	case AO:
		r := make(AO, len(w))
		for i := range r {
			v := ${name}(w[i], x)
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("${op}")
	}
}\n
EOS
    print $s;
    for my $t (sort keys %types) {
        genLeftExpanded($name, $op, $cases, $t);
    }
    for my $t (sort keys %types) {
        genLeftArrayExpanded($name, $op, $cases, $t);
    }
}

sub genLeftExpanded {
    my ($name, $op, $cases, $t) = @_;
    my %types = map { /_(\w)/; $1 => $cases->{"${t}_$1"}} grep { /${t}_(\w)/ } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}${t}O(w $t, x Object) Object {
	switch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        print $out <<EOS;
	case $tt:
		return $expr
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w", "x[i]");
        print $out <<EOS;
	case A$tt:
		r := make(A$type, len(x))
		for i := range r {
			r[i] = $iexpr
		}
		return r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case AO:
		r := make([]Object, len(x))
		for i := range r {
			v := ${name}${t}O(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("${op}")
	}
}\n
EOS
    print $s;
}

sub genLeftArrayExpanded {
    my ($name, $op, $cases, $t) = @_;
    my %types = map { /_(\w)/; $1 => $cases->{"${t}_$1"}} grep { /${t}_(\w)/ } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}A${t}O(w A$t, x Object) Object {
	switch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w[i]", "x");
        print $out <<EOS;
	case $tt:
		r := make(A$type, len(w))
		for i := range r {
			r[i] = $iexpr
		}
		return r
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w[i]", "x[i]");
        print $out <<EOS;
	case A$tt:
		r := make(A$type, len(x))
		for i := range r {
			r[i] = $iexpr
		}
		return r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case AO:
		r := make(AO, len(x))
		for i := range r {
			v := ${name}A${t}O(w, x[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return w
	default:
		return badtype("${op}")
	}
}\n
EOS
    print $s;
}

sub subst {
    my ($expr, $w, $x) = @_;
    $expr =~ s/\bw\b/$w/g;
    $expr =~ s/\bx\b/$x/g;
    return $expr
}
