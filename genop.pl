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
    #NotEqual =>  {
        #B_B => ["w != x", "B"],
        #B_I => ["B2I(w) != x", "B"],
        #B_F => ["B2F(w) != x", "B"],
        #I_B => ["w != B2I(x)", "B"],
        #I_I => ["w != x", "B"],
        #I_F => ["F(w) != x", "B"],
        #F_B => ["w != B2F(x)", "B"],
        #F_I => ["w != F(x)", "B"],
        #F_F => ["w != x", "B"],
        #S_S => ["w != x", "B"],
    #},
    Lesser =>  {
        B_B => ["!w && x", "B"],
        B_I => ["B2I(w) < x", "B"],
        B_F => ["B2F(w) < x", "B"],
        I_B => ["w < B2I(x)", "B"],
        I_I => ["w < x", "B"],
        I_F => ["F(w) < x", "B"],
        F_B => ["w < B2F(x)", "B"],
        F_I => ["w < F(x)", "B"],
        F_F => ["w < x", "B"],
        S_S => ["w < x", "B"],
    },
    #LesserEq =>  {
        #B_B => ["x || !w", "B"],
        #B_I => ["B2I(w) <= x", "B"],
        #B_F => ["B2F(w) <= x", "B"],
        #I_B => ["w <= B2I(x)", "B"],
        #I_I => ["w <= x", "B"],
        #I_F => ["F(w) <= x", "B"],
        #F_B => ["w <= B2F(x)", "B"],
        #F_I => ["w <= F(x)", "B"],
        #F_F => ["w <= x", "B"],
        #S_S => ["w <= x", "B"],
    #},
    Greater =>  {
        B_B => ["w && !x", "B"],
        B_I => ["B2I(w) > x", "B"],
        B_F => ["B2F(w) > x", "B"],
        I_B => ["w > B2I(x)", "B"],
        I_I => ["w > x", "B"],
        I_F => ["F(w) > x", "B"],
        F_B => ["w > B2F(x)", "B"],
        F_I => ["w > F(x)", "B"],
        F_F => ["w > x", "B"],
        S_S => ["w > x", "B"],
    },
    #GreaterEq =>  {
        #B_B => ["w || !x", "B"],
        #B_I => ["B2I(w) >= x", "B"],
        #B_F => ["B2F(w) >= x", "B"],
        #I_B => ["w >= B2I(x)", "B"],
        #I_I => ["w >= x", "B"],
        #I_F => ["F(w) >= x", "B"],
        #F_B => ["w >= B2F(x)", "B"],
        #F_I => ["w >= F(x)", "B"],
        #F_F => ["w >= x", "B"],
        #S_S => ["w >= x", "B"],
    #},
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
        S_S => ["w + x", "S"],
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
        S_S => ["strings.TrimSuffix(string(w), string(x))", "S"],
    },
    #Span =>  {
        #B_B => ["1+ B2I(w) - B2I(x)", "I"],
        #B_I => ["1 + B2I(w) - x", "I"],
        #B_F => ["1 + B2F(w) - x", "F"],
        #I_B => ["1 + w - B2I(x)", "I"],
        #I_I => ["1 + w - x", "I"],
        #I_F => ["1 + F(w) - x", "F"],
        #F_B => ["1 + w - B2F(x)", "F"],
        #F_I => ["1 + w - F(x)", "F"],
        #F_F => ["1 + w - x", "F"],
    #},
    Multiply =>  {
        B_B => ["w && x", "B"],
        B_I => ["B2I(w) * x", "I"],
        B_F => ["B2F(w) * x", "F"],
        B_S => ["strings.Repeat(string(x), int(B2I(w)))", "S"],
        I_B => ["w * B2I(x)", "I"],
        I_I => ["w * x", "I"],
        I_F => ["F(w) * x", "F"],
        I_S => ["strings.Repeat(string(x), int(w))", "S"],
        F_B => ["w * B2F(x)", "F"],
        F_I => ["w * F(x)", "F"],
        F_F => ["w * x", "F"],
        F_S => ["strings.Repeat(string(x), int(math.Round(float64(w))))", "S"],
        S_B => ["strings.Repeat(string(w), int(B2I(x)))", "S"],
        S_I => ["strings.Repeat(string(w), int(x))", "S"],
        S_F => ["strings.Repeat(string(w), int(math.Round(float64(x))))", "S"],
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
    Minimum =>  {
        B_B => ["w && x", "B"],
        B_I => ["minI(B2I(w), x)", "I"],
        B_F => ["F(math.Min(float64(B2F(w)), float64(x)))", "F"],
        I_B => ["minI(w, B2I(x))", "I"],
        I_I => ["minI(w, x)", "I"],
        I_F => ["F(math.Min(float64(w), float64(x)))", "F"],
        F_B => ["F(math.Min(float64(w), float64(B2F(x))))", "F"],
        F_I => ["F(math.Min(float64(w), float64(x)))", "F"],
        F_F => ["F(math.Min(float64(w), float64(x)))", "F"],
        S_S => ["minS(w, x)", "S"],
    },
    Maximum =>  {
        B_B => ["w || x", "B"],
        B_I => ["maxI(B2I(w), x)", "I"],
        B_F => ["F(math.Max(float64(B2F(w)), float64(x)))", "F"],
        I_B => ["maxI(w, B2I(x))", "I"],
        I_I => ["maxI(w, x)", "I"],
        I_F => ["F(math.Max(float64(w), float64(x)))", "F"],
        F_B => ["F(math.Max(float64(w), float64(B2F(x))))", "F"],
        F_I => ["F(math.Max(float64(w), float64(x)))", "F"],
        F_F => ["F(math.Max(float64(w), float64(x)))", "F"],
        S_S => ["maxS(w, x)", "S"],
    },
    #Or =>  {
        #B_B => ["w || x", "B"],
        #B_I => ["1-((1-B2I(w)) * (1-x))", "I"],
        #B_F => ["1-((1-B2F(w)) * (1-x))", "F"],
        #I_B => ["1-((1-w) * (1-B2I(x)))", "I"],
        #I_I => ["1-((1-w) * (1-x))", "I"],
        #I_F => ["1-((1-F(w)) * (1-x))", "F"],
        #F_B => ["1-((1-w) * (1-B2F(x)))", "F"],
        #F_I => ["1-((1-w) * F(1-x))", "F"],
        #F_F => ["1-((1-w) * (1-x))", "F"],
    #},
    #And =>  {
        #B_B => ["w && x", "B"],
        #B_I => ["B2I(w) * x", "I"],
        #B_F => ["B2F(w) * x", "F"],
        #I_B => ["w * B2I(x)", "I"],
        #I_I => ["w * x", "I"],
        #I_F => ["F(w) * x", "F"],
        #F_B => ["w * B2F(x)", "F"],
        #F_I => ["w * F(x)", "F"],
        #F_F => ["w * x", "F"],
    #},
    Modulus =>  {
        B_B => ["modI(B2I(w), B2I(x))", "I"],
        B_I => ["modI(B2I(w), x)", "I"],
        B_F => ["modF(F(B2I(w)), x)", "I"],
        I_B => ["modI(w, B2I(x))", "I"],
        I_I => ["modI(w, x)", "I"],
        I_F => ["modF(F(w), x)", "I"],
        F_B => ["modF(w, F(B2I(x)))", "I"],
        F_I => ["modF(w, F(x))", "I"],
        F_F => ["modF(w, x)", "I"],
    },
);

my %atypes = (
    B => "bool",
    I => "int",
    F => "float64",
    S => "string",
);

print <<EOS;
// Code generated by genop.pl. DO NOT EDIT.

package main

import (
    "math"
    "strings"
)

EOS

genOp("Equal", "=");
#genOp("NotEqual", "≠");
genOp("Lesser", "<");
#genOp("LesserEq", "≤");
genOp("Greater", ">");
#genOp("GreaterEq", "≥");
genOp("Add", "+");
genOp("Subtract", "-");
#genOp("Span", "¬");
genOp("Multiply", "*");
genOp("Divide", "%");
genOp("Minimum", "&");
genOp("Maximum", "|");
#genOp("And", "∧"); # identical to Multiply
#genOp("Or", "∨"); # Multiply under Not
genOp("Modulus", "!");

sub genOp {
    my ($name, $op) = @_;
    my $cases = $dyads{$name};
    my %types = map { $_=~/(\w)_/; $1 => 1 } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
// ${name} returns w${op}x.
func ${name}(w, x V) V {
	switch w := w.(type) {
EOS
    for my $t (sort keys %types) {
        print $out <<EOS;
	case $t:
		return ${name}${t}V(w, x)
EOS
    }
    for my $t (sort keys %types) {
        print $out <<EOS;
	case A$t:
		return ${name}A${t}V(w, x)
EOS
    }
    print $out <<EOS;
	case AV:
                switch x := x.(type) {
                case Array:
                        if x.Len() != len(w) {
                                return badlen("$op")
                        }
                        r := make(AV, len(w))
                        for i := range r {
                                v := ${name}(w[i], x.At(i))
                                e, ok := v.(E)
                                if ok {
                                        return e
                                }
                                r[i] = v
                        }
                        return r
                }
                r := make(AV, len(w))
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
func ${name}${t}V(w $t, x V) V {
	switch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        print $out <<EOS;
	case $tt:
		return $type($expr)
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "w", "x[i]");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case A$tt:
		r := make(A$type, len(x))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case AV:
		r := make(AV, len(x))
		for i := range r {
			v := ${name}${t}V($t(w), x[i])
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
func ${name}A${t}V(w A$t, x V) V {
	switch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "w[i]", "x");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case $tt:
		r := make(A$type, len(w))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "w[i]", "x[i]");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case A$tt:
                if len(w) != len(x) {
                        return badlen("$op")
                }
		r := make(A$type, len(x))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case AV:
                if len(w) != len(x) {
                        return badlen("$op")
                }
		r := make(AV, len(x))
		for i := range r {
			v := ${name}${t}V($t(w[i]), x[i])
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
    my ($expr, $t, $tt, $w, $x) = @_;
    $expr =~ s/(!w|\bB2[IF]\(w\)|\bw)\b/$t($1)/g;
    $expr =~ s/(!x|\bB2[IF]\(x\)|\bx)\b/$tt($1)/g;
    $expr =~ s/\bw\b/$w/g;
    $expr =~ s/\bx\b/$x/g;
    return $expr
}
