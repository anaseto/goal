#!/usr/bin/env perl

use strict;
use warnings;
use v5.28;

my %dyads = (
    equal =>  {
        B_B => ["x == y", "B"],
        B_I => ["B2I(x) == y", "B"],
        B_F => ["B2F(x) == y", "B"],
        I_B => ["x == B2I(y)", "B"],
        I_I => ["x == y", "B"],
        I_F => ["F(x) == y", "B"],
        F_B => ["x == B2F(y)", "B"],
        F_I => ["x == F(y)", "B"],
        F_F => ["x == y", "B"],
        S_S => ["x == y", "B"],
    },
    #NotEqual =>  {
        #B_B => ["x != y", "B"],
        #B_I => ["B2I(x) != y", "B"],
        #B_F => ["B2F(x) != y", "B"],
        #I_B => ["x != B2I(y)", "B"],
        #I_I => ["x != y", "B"],
        #I_F => ["F(x) != y", "B"],
        #F_B => ["x != B2F(y)", "B"],
        #F_I => ["x != F(y)", "B"],
        #F_F => ["x != y", "B"],
        #S_S => ["x != y", "B"],
    #},
    lesser =>  {
        B_B => ["!x && y", "B"],
        B_I => ["B2I(x) < y", "B"],
        B_F => ["B2F(x) < y", "B"],
        I_B => ["x < B2I(y)", "B"],
        I_I => ["x < y", "B"],
        I_F => ["F(x) < y", "B"],
        F_B => ["x < B2F(y)", "B"],
        F_I => ["x < F(y)", "B"],
        F_F => ["x < y", "B"],
        S_S => ["x < y", "B"],
    },
    #LesserEq =>  {
        #B_B => ["y || !x", "B"],
        #B_I => ["B2I(x) <= y", "B"],
        #B_F => ["B2F(x) <= y", "B"],
        #I_B => ["x <= B2I(y)", "B"],
        #I_I => ["x <= y", "B"],
        #I_F => ["F(x) <= y", "B"],
        #F_B => ["x <= B2F(y)", "B"],
        #F_I => ["x <= F(y)", "B"],
        #F_F => ["x <= y", "B"],
        #S_S => ["x <= y", "B"],
    #},
    greater =>  {
        B_B => ["x && !y", "B"],
        B_I => ["B2I(x) > y", "B"],
        B_F => ["B2F(x) > y", "B"],
        I_B => ["x > B2I(y)", "B"],
        I_I => ["x > y", "B"],
        I_F => ["F(x) > y", "B"],
        F_B => ["x > B2F(y)", "B"],
        F_I => ["x > F(y)", "B"],
        F_F => ["x > y", "B"],
        S_S => ["x > y", "B"],
    },
    #GreaterEq =>  {
        #B_B => ["x || !y", "B"],
        #B_I => ["B2I(x) >= y", "B"],
        #B_F => ["B2F(x) >= y", "B"],
        #I_B => ["x >= B2I(y)", "B"],
        #I_I => ["x >= y", "B"],
        #I_F => ["F(x) >= y", "B"],
        #F_B => ["x >= B2F(y)", "B"],
        #F_I => ["x >= F(y)", "B"],
        #F_F => ["x >= y", "B"],
        #S_S => ["x >= y", "B"],
    #},
    add =>  {
        B_B => ["B2I(x) + B2I(y)", "I"],
        B_I => ["B2I(x) + y", "I"],
        B_F => ["B2F(x) + y", "F"],
        I_B => ["x + B2I(y)", "I"],
        I_I => ["x + y", "I"],
        I_F => ["F(x) + y", "F"],
        F_B => ["x + B2F(y)", "F"],
        F_I => ["x + F(y)", "F"],
        F_F => ["x + y", "F"],
        S_S => ["x + y", "S"],
    },
    subtract =>  {
        B_B => ["B2I(x) - B2I(y)", "I"],
        B_I => ["B2I(x) - y", "I"],
        B_F => ["B2F(x) - y", "F"],
        I_B => ["x - B2I(y)", "I"],
        I_I => ["x - y", "I"],
        I_F => ["F(x) - y", "F"],
        F_B => ["x - B2F(y)", "F"],
        F_I => ["x - F(y)", "F"],
        F_F => ["x - y", "F"],
        S_S => ["strings.TrimSuffix(string(x), string(y))", "S"],
    },
    #Span =>  {
        #B_B => ["1+ B2I(x) - B2I(y)", "I"],
        #B_I => ["1 + B2I(x) - y", "I"],
        #B_F => ["1 + B2F(x) - y", "F"],
        #I_B => ["1 + x - B2I(y)", "I"],
        #I_I => ["1 + x - y", "I"],
        #I_F => ["1 + F(x) - y", "F"],
        #F_B => ["1 + x - B2F(y)", "F"],
        #F_I => ["1 + x - F(y)", "F"],
        #F_F => ["1 + x - y", "F"],
    #},
    multiply =>  {
        B_B => ["x && y", "B"],
        B_I => ["B2I(x) * y", "I"],
        B_F => ["B2F(x) * y", "F"],
        B_S => ["strings.Repeat(string(y), int(B2I(x)))", "S"],
        I_B => ["x * B2I(y)", "I"],
        I_I => ["x * y", "I"],
        I_F => ["F(x) * y", "F"],
        I_S => ["strings.Repeat(string(y), int(x))", "S"],
        F_B => ["x * B2F(y)", "F"],
        F_I => ["x * F(y)", "F"],
        F_F => ["x * y", "F"],
        F_S => ["strings.Repeat(string(y), int(float64(x)))", "S"],
        S_B => ["strings.Repeat(string(x), int(B2I(y)))", "S"],
        S_I => ["strings.Repeat(string(x), int(y))", "S"],
        S_F => ["strings.Repeat(string(x), int(float64(y)))", "S"],
    },
    divide =>  {
        B_B => ["divideF(B2F(x), B2F(y))", "F"],
        B_I => ["divideF(B2F(x), F(y))", "F"],
        B_F => ["divideF(B2F(x), y)", "F"],
        I_B => ["divideF(F(x), B2F(y))", "F"],
        I_I => ["divideF(F(x), F(y))", "F"],
        I_F => ["divideF(F(x), y)", "F"],
        F_B => ["divideF(x, B2F(y))", "F"],
        F_I => ["divideF(x, F(y))", "F"],
        F_F => ["divideF(x, y)", "F"],
    },
    minimum =>  {
        B_B => ["x && y", "B"],
        B_I => ["minI(B2I(x), y)", "I"],
        B_F => ["F(math.Min(float64(B2F(x)), float64(y)))", "F"],
        I_B => ["minI(x, B2I(y))", "I"],
        I_I => ["minI(x, y)", "I"],
        I_F => ["F(math.Min(float64(x), float64(y)))", "F"],
        F_B => ["F(math.Min(float64(x), float64(B2F(y))))", "F"],
        F_I => ["F(math.Min(float64(x), float64(y)))", "F"],
        F_F => ["F(math.Min(float64(x), float64(y)))", "F"],
        S_S => ["minS(x, y)", "S"],
    },
    maximum =>  {
        B_B => ["x || y", "B"],
        B_I => ["maxI(B2I(x), y)", "I"],
        B_F => ["F(math.Max(float64(B2F(x)), float64(y)))", "F"],
        I_B => ["maxI(x, B2I(y))", "I"],
        I_I => ["maxI(x, y)", "I"],
        I_F => ["F(math.Max(float64(x), float64(y)))", "F"],
        F_B => ["F(math.Max(float64(x), float64(B2F(y))))", "F"],
        F_I => ["F(math.Max(float64(x), float64(y)))", "F"],
        F_F => ["F(math.Max(float64(x), float64(y)))", "F"],
        S_S => ["maxS(x, y)", "S"],
    },
    #Or =>  {
        #B_B => ["x || y", "B"],
        #B_I => ["1-((1-B2I(x)) * (1-y))", "I"],
        #B_F => ["1-((1-B2F(x)) * (1-y))", "F"],
        #I_B => ["1-((1-x) * (1-B2I(y)))", "I"],
        #I_I => ["1-((1-x) * (1-y))", "I"],
        #I_F => ["1-((1-F(x)) * (1-y))", "F"],
        #F_B => ["1-((1-x) * (1-B2F(y)))", "F"],
        #F_I => ["1-((1-x) * F(1-y))", "F"],
        #F_F => ["1-((1-x) * (1-y))", "F"],
    #},
    #And =>  {
        #B_B => ["x && y", "B"],
        #B_I => ["B2I(x) * y", "I"],
        #B_F => ["B2F(x) * y", "F"],
        #I_B => ["x * B2I(y)", "I"],
        #I_I => ["x * y", "I"],
        #I_F => ["F(x) * y", "F"],
        #F_B => ["x * B2F(y)", "F"],
        #F_I => ["x * F(y)", "F"],
        #F_F => ["x * y", "F"],
    #},
    modulus =>  {
        B_B => ["modI(B2I(x), B2I(y))", "I"],
        B_I => ["modI(B2I(x), y)", "I"],
        B_F => ["modF(F(B2I(x)), y)", "I"],
        I_B => ["modI(x, B2I(y))", "I"],
        I_I => ["modI(x, y)", "I"],
        I_F => ["modF(F(x), y)", "I"],
        F_B => ["modF(x, F(B2I(y)))", "I"],
        F_I => ["modF(x, F(y))", "I"],
        F_F => ["modF(x, y)", "I"],
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

package goal

import (
    "math"
    "strings"
)

EOS

genOp("equal", "=");
#genOp("NotEqual", "≠");
genOp("lesser", "<");
#genOp("LesserEq", "≤");
genOp("greater", ">");
#genOp("GreaterEq", "≥");
genOp("add", "+");
genOp("subtract", "-");
#genOp("Span", "¬");
genOp("multiply", "*");
genOp("divide", "%");
genOp("minimum", "&");
genOp("maximum", "|");
#genOp("And", "∧"); # identical to Multiply
#genOp("Or", "∨"); # Multiply under Not
genOp("modulus", " mod ");

sub genOp {
    my ($name, $op) = @_;
    my $errOp = $op;
    $errOp .= "%" if $op eq "%";
    my $cases = $dyads{$name};
    my %types = map { $_=~/(\w)_/; $1 => 1 } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    my $namelc = lc($name);
    print $out <<EOS;
// ${name} returns x${op}y.
func ${name}(x, y V) V {
	switch x := x.(type) {
EOS
    for my $t (sort keys %types) {
        next if $t eq "B";
        print $out <<EOS;
	case $t:
		return ${namelc}${t}V(x, y)
EOS
    }
    for my $t (sort keys %types) {
        print $out <<EOS;
	case A$t:
		return ${namelc}A${t}V(x, y)
EOS
    }
    print $out <<EOS;
	case AV:
                switch y := y.(type) {
                case Array:
                        if y.Len() != len(x) {
                                return errf("x${errOp}y : length mismatch: %d vs %d", len(x), y.Len())
                        }
                        r := make(AV, len(x))
                        for i := range r {
                                v := ${name}(x[i], y.At(i))
                                e, ok := v.(E)
                                if ok {
                                        return e
                                }
                                r[i] = v
                        }
                        return r
                }
                r := make(AV, len(x))
                for i := range r {
                        v := ${name}(x[i], y)
                        e, ok := v.(E)
                        if ok {
                                return e
                        }
                        r[i] = v
                }
                return r
	case E:
		return x
	default:
		return errf("x${errOp}y : bad type `%s for x", x.Type())
	}
}\n
EOS
    print $s;
    for my $t (sort keys %types) {
        next if $t eq "B";
        genLeftExpanded($namelc, $cases, $t, $errOp);
    }
    for my $t (sort keys %types) {
        genLeftArrayExpanded($namelc, $cases, $t, $errOp);
    }
}

sub genLeftExpanded {
    my ($name, $cases, $t, $errOp) = @_;
    my %types = map { /_(\w)/; $1 => $cases->{"${t}_$1"}} grep { /${t}_(\w)/ } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}${t}V(x $t, y V) V {
	switch y := y.(type) {
EOS
    for my $tt (sort keys %types) {
        next if $tt eq "B";
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        $type = "B2I" if $type eq "B";
        print $out <<EOS;
	case $tt:
		return $type($expr)
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x", "y[i]");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case A$tt:
		r := make(A$type, len(y))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case AV:
		r := make(AV, len(y))
		for i := range r {
			v := ${name}${t}V($t(x), y[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return x
	default:
		return errf("x${errOp}y : bad type `%s for y", y.Type())
	}
}\n
EOS
    print $s;
}

sub genLeftArrayExpanded {
    my ($name, $cases, $t, $errOp) = @_;
    my %types = map { /_(\w)/; $1 => $cases->{"${t}_$1"}} grep { /${t}_(\w)/ } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}A${t}V(x A$t, y V) V {
	switch y := y.(type) {
EOS
    for my $tt (sort keys %types) {
        next if $tt eq "B";
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x[i]", "y");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case $tt:
		r := make(A$type, len(x))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x[i]", "y[i]");
        my $rtype = $atypes{$type};
        print $out <<EOS;
	case A$tt:
                if len(x) != len(y) {
                        return errf("x${errOp}y : length mismatch: %d vs %d", len(x), len(y))
                }
		r := make(A$type, len(y))
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return r
EOS
    }
    my $tt = $t;
    if ($t eq "B") {
        $t = "I";
        $tt = "B2I";
    }
    print $out <<EOS if $t !~ /^A/;
	case AV:
                if len(x) != len(y) {
                        return errf("x${errOp}y : length mismatch: %d vs %d", len(x), len(y))
                }
		r := make(AV, len(y))
		for i := range r {
			v := ${name}${t}V($tt(x[i]), y[i])
			e, ok := v.(E)
			if ok {
				return e
			}
			r[i] = v
		}
		return r
	case E:
		return x
	default:
		return errf("x${errOp}y : bad type `%s for y", y.Type())
	}
}\n
EOS
    print $s;
}

sub subst {
    my ($expr, $t, $tt, $x, $y) = @_;
    $expr =~ s/(!x|\bB2[IF]\(x\)|\bx)\b/$t($1)/g unless $t eq "B";
    $expr =~ s/(!y|\bB2[IF]\(y\)|\by)\b/$tt($1)/g unless $tt eq "B";
    $expr =~ s/\bx\b/$x/g;
    $expr =~ s/\by\b/$y/g;
    return $expr
}
