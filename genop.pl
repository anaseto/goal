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
        B_F => ["modF(F(B2I(x)), y)", "F"],
        I_B => ["modI(x, B2I(y))", "I"],
        I_I => ["modI(x, y)", "I"],
        I_F => ["modF(F(x), y)", "F"],
        F_B => ["modF(x, F(B2I(y)))", "F"],
        F_I => ["modF(x, F(y))", "F"],
        F_F => ["modF(x, y)", "F"],
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
EOS
    if ($types{"I"}) {
        print $out <<EOS;
        if x.IsInt() {
		return ${namelc}IV(x.Int(), y)
        }
EOS
    }
        print $out <<EOS;
	switch x := x.Value.(type) {
EOS
    for my $t (sort keys %types) {
        next if $t eq "B" or $t eq "I";
        print $out <<EOS;
	case $t:
		return ${namelc}${t}V(x, y)
EOS
    }
    for my $t (sort keys %types) {
        print $out <<EOS;
	case *A$t:
		return ${namelc}A${t}V(x, y)
EOS
    }
    print $out <<EOS;
	case *AV:
                switch y := y.Value.(type) {
                case array:
                        if y.Len() != x.Len() {
                                return errf("x${errOp}y : length mismatch: %d vs %d", x.Len(), y.Len())
                        }
                        r := x.reuse()
                        for i, xi := range x.Slice {
                                ri := ${name}(xi, y.at(i))
                                if ri.IsErr() {
                                        return ri
                                }
                                r.Slice[i] = ri
                        }
                        return NewV(r) 
                }
                r := x.reuse()
                for i, xi := range x.Slice {
                        ri := ${name}(xi, y)
                        if ri.IsErr() {
                                return ri
                        }
                        r.Slice[i] = ri
                }
                return NewV(r)
	default:
		return errType("x${errOp}y", "x", x)
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
    my $xt = $t;
    if ($xt eq "I") {
        $xt = "int"
    }
    print $out <<EOS;
func ${name}${t}V(x $xt, y V) V {
EOS
    if ($types{"I"}) {
        my $expr = $cases->{"${t}_I"}->[0];
        my $type = $cases->{"${t}_I"}->[1];
        $expr = "B2I($expr)" if $type eq "B";
        $expr =~ s/\by\b/y.Int()/g;
        $type = "I" if $type eq "B";
        my $ret = "NewV($type($expr))";
        if ($type eq "I") {
            $ret = "NewI($expr)";
        }
        print $out <<EOS;
        if y.IsInt() {
            return $ret;
        }
EOS
    }
    print $out <<EOS;
	switch y := y.Value.(type) {
EOS
    for my $tt (sort keys %types) {
        next if $tt eq "B" or $tt eq "I";
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $nt = "V";
        $nt = "I" if $type eq "B" or $type eq "I";
        $expr = "B2I($expr)" if $type eq "B";
        $type = "int" if $type eq "B" or $type eq "I";
        #$type = "int" if $type eq "B";
		#return New${nt}($type($expr))
        print $out <<EOS;
	case $tt:
		return New${nt}($type($expr))
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x", "y.At(i)");
        my $rtype = $atypes{$type};
        if ($tt eq $type) {
            print $out <<EOS;
	case *A$tt:
		r := y.reuse()
		for i := range r.Slice {
			r.Slice[i] = $rtype($iexpr)
		}
		return NewV(r)
EOS
        } else {
            print $out <<EOS;
	case *A$tt:
		r := make([]$rtype, y.Len())
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return NewA$type(r)
EOS
        }
    }
    print $out <<EOS if $t !~ /^A/;
	case *AV:
		r := y.reuse()
		for i, yi := range y.Slice {
			ri := ${name}${t}V(x, yi)
                        if ri.IsErr() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return errType("x${errOp}y", "y", y)
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
func ${name}A${t}V(x *A$t, y V) V {
EOS
    if ($types{"I"}) {
        my $expr = $cases->{"${t}_I"}->[0];
        my $type = $cases->{"${t}_I"}->[1];
        my $iexpr = subst($expr, $t, "int", "x.At(i)", "y.Int()");
        my $rtype = $atypes{$type};
        if ($t eq $type) {
            print $out <<EOS;
        if y.IsInt() {
            r := x.reuse()
            for i := range r.Slice {
                    r.Slice[i] = $rtype($iexpr)
            }
            return NewV(r)
        }
EOS
        } else {
            print $out <<EOS;
        if y.IsInt() {
            r := make([]$rtype, x.Len())
            for i := range r {
                    r[i] = $rtype($iexpr)
            }
            return NewA$type(r)
        }
EOS
        }
    }
    print $out <<EOS;
	switch y := y.Value.(type) {
EOS
    for my $tt (sort keys %types) {
        next if $tt eq "B" or $tt eq "I";
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x.At(i)", "y");
        my $rtype = $atypes{$type};
        if ($t eq $type) {
            print $out <<EOS;
	case $tt:
		r := x.reuse()
		for i := range r.Slice {
			r.Slice[i] = $rtype($iexpr)
		}
		return NewV(r)
EOS
        } else {
            print $out <<EOS;
	case $tt:
		r := make([]$rtype, x.Len())
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return NewA$type(r)
EOS
        }
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, $t, $tt, "x.At(i)", "y.At(i)");
        my $rtype = $atypes{$type};
        if ($t eq $type) {
            print $out <<EOS;
	case *A$tt:
                if x.Len() != y.Len() {
                        return errf("x${errOp}y : length mismatch: %d vs %d", x.Len(), y.Len())
                }
                r := x.reuse()
		for i := range r.Slice {
			r.Slice[i] = $rtype($iexpr)
		}
		return NewV(r)
EOS
        } elsif ($tt eq $type) {
            print $out <<EOS;
	case *A$tt:
                if x.Len() != y.Len() {
                        return errf("x${errOp}y : length mismatch: %d vs %d", x.Len(), y.Len())
                }
                r := y.reuse()
		for i := range r.Slice {
			r.Slice[i] = $rtype($iexpr)
		}
		return NewV(r)
EOS
        } else {
            print $out <<EOS;
	case *A$tt:
                if x.Len() != y.Len() {
                        return errf("x${errOp}y : length mismatch: %d vs %d", x.Len(), y.Len())
                }
		r := make([]$rtype, y.Len())
		for i := range r {
			r[i] = $rtype($iexpr)
		}
		return NewA$type(r)
EOS
        }
    }
    my $tt = $t;
    if ($t eq "B") {
        $t = "I";
        $tt = "B2I";
    } elsif ($t eq "I") {
        $tt = "int";
    }
    my $reuse;
    if ($t eq "V") {
        $reuse = <<EOS;
                var r *AV
                if x.reusable() {
                    r = x.reuse()
                } else {
                    r = y.reuse()
                }
EOS
    } else {
        $reuse = <<EOS;
                r := y.reuse()
EOS
    }
    print $out <<EOS if $t !~ /^A/;
	case *AV:
                if x.Len() != y.Len() {
                        return errf("x${errOp}y : length mismatch: %d vs %d", x.Len(), y.Len())
                }
                $reuse
		for i := range r.Slice {
			ri := ${name}${t}V($tt(x.At(i)), y.At(i))
                        if ri.IsErr() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return errType("x${errOp}y", "y", y)
	}
}\n
EOS
    print $s;
}

sub subst {
    my ($expr, $t, $tt, $x, $y) = @_;
    $expr =~ s/(!x|\bB2[IF]\(x\)|\bx)\b/$t($1)/g unless $t eq "B" or $t eq "I";
    $expr =~ s/(!y|\bB2[IF]\(y\)|\by)\b/$tt($1)/g unless $tt eq "B" or $tt eq "I";
    $expr =~ s/\bx\b/$x/g;
    $expr =~ s/\by\b/$y/g;
    return $expr
}
