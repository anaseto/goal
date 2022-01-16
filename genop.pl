#!/usr/bin/env perl

use strict;
use warnings;
use v5.28;
use utf8;
use open qw(:std :utf8);

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

sub genOp {
    my ($name, $op) = @_;
    my $cases = $dyads{$name};
    my %types = map { $_=~/(\w)_/; $1 => 1 } keys $cases->%*;
    my $s = "";
    open my $out, '>', \$s;
    print $out <<EOS;
func ${name}(w, x Object) Object {
\tswitch w := w.(type) {
EOS
    for my $t (sort keys %types) {
        print $out <<EOS;
\tcase $t:
\t\treturn ${name}${t}O(w, x)
EOS
    }
    for my $t (sort keys %types) {
        print $out <<EOS;
\tcase A$t:
\t\treturn ${name}A${t}O(w, x)
EOS
    }
    print $out <<EOS;
\tcase AO:
\t\tr := make(AO, len(w))
\t\tfor i := range r {
\t\t\tv := ${name}(w[i], x)
\t\t\te, ok := v.(E)
\t\t\tif ok {
\t\t\t\treturn e
\t\t\t}
\t\t\tr[i] = v
\t\t}
\t\treturn r
\tcase E:
\t\treturn w
\tdefault:
\t\treturn badtype("${op}")
\t}
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
\tswitch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        print $out <<EOS;
\tcase $tt:
\t\treturn $expr
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w", "x[i]");
        print $out <<EOS;
\tcase A$tt:
\t\tr := make(A$type, len(x))
\t\tfor i := range r {
\t\t\tr[i] = $iexpr
\t\t}
\t\treturn r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
\tcase AO:
\t\tr := make([]Object, len(x))
\t\tfor i := range r {
\t\t\tv := ${name}${t}O(w, x[i])
\t\t\te, ok := v.(E)
\t\t\tif ok {
\t\t\t\treturn e
\t\t\t}
\t\t\tr[i] = v
\t\t}
\t\treturn r
\tcase E:
\t\treturn w
\tdefault:
\t\treturn badtype("${op}")
\t}
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
\tswitch x := x.(type) {
EOS
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w[i]", "x");
        print $out <<EOS;
\tcase $tt:
\t\tr := make(A$type, len(w))
\t\tfor i := range r {
\t\t\tr[i] = $iexpr
\t\t}
\t\treturn r
EOS
    }
    for my $tt (sort keys %types) {
        my $expr = $cases->{"${t}_$tt"}->[0];
        my $type = $cases->{"${t}_$tt"}->[1];
        my $iexpr = subst($expr, "w[i]", "x[i]");
        print $out <<EOS;
\tcase A$tt:
\t\tr := make(A$type, len(x))
\t\tfor i := range r {
\t\t\tr[i] = $iexpr
\t\t}
\t\treturn r
EOS
    }
    print $out <<EOS if $t !~ /^A/;
\tcase AO:
\t\tr := make(AO, len(x))
\t\tfor i := range r {
\t\t\tv := ${name}A${t}O(w, x[i])
\t\t\te, ok := v.(E)
\t\t\tif ok {
\t\t\t\treturn e
\t\t\t}
\t\t\tr[i] = v
\t\t}
\t\treturn r
\tcase E:
\t\treturn w
\tdefault:
\t\treturn badtype("${op}")
\t}
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

#print <<EOS;
#package main
#EOS

#for my $k (sort keys %dyads) {
    #doFun($k, $dyads{$k})
#}

#sub doFun {
    #my ($name, $op) = @_;
    #print <<EOS;

#// ${name} returns w${op}x.
#func ${name}(w, x Object) Object {
	#switch w := w.(type) {
	#case I:
		#return ${name}I(w, x)
	#case F:
		#return ${name}F(w, x)
	#case AI:
		#return ${name}AI(w, x)
	#case AF:
		#return ${name}AF(w, x)
	#case AO:
		#return ${name}AO(w, x)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}I(i I, x Object) Object {
	#switch x := x.(type) {
	#case I:
		#return i ${op} x
	#case F:
		#return F(i) ${op} x
	#case AI:
		#return ${name}AII(x, i)
	#case AF:
		#return ${name}AFI(x, i)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}F(f F, x Object) Object {
	#switch x := x.(type) {
	#case I:
		#return f ${op} F(x)
	#case F:
		#return f ${op} x
	#case AI:
		#return ${name}AIF(x, f)
	#case AF:
		#return ${name}AFF(x, f)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}AI(a AI, x Object) Object {
	#switch x := x.(type) {
	#case I:
		#return ${name}AII(a, x)
	#case F:
		#return ${name}AIF(a, x)
	#case AI:
		#return ${name}AIAI(a, x)
	#case AF:
		#return ${name}AIAF(a, x)
	#case AO:
		#return ${name}AOAI(x, a)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}AF(a AF, x Object) Object {
	#switch x := x.(type) {
	#case I:
		#return ${name}AFI(a, x)
	#case F:
		#return ${name}AFF(a, x)
	#case AI:
		#return ${name}AIAF(x, a)
	#case AF:
		#return ${name}AFAF(a, x)
	#case AO:
		#return ${name}AOAF(x, a)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}AO(a AO, x Object) Object {
	#switch x := x.(type) {
	#case I:
		#return ${name}AOI(a, x)
	#case F:
		#return ${name}AOF(a, x)
	#case AI:
		#return ${name}AOAI(a, x)
	#case AF:
		#return ${name}AOAF(a, x)
	#case AO:
		#return ${name}AOAO(a, x)
	#}
	#// TODO
	#return badtype("${op}")
#}

#func ${name}AII(a AI, i I) AI {
	#r := make(AI, len(a))
	#for j := range r {
		#r[j] = a[j] ${op} i
	#}
	#return r
#}

#func ${name}AIF(a AI, f F) AF {
	#r := make(AF, len(a))
	#for j := range r {
		#r[j] = F(a[j]) ${op} f
	#}
	#return r
#}

#func ${name}AFI(a AF, i I) AF {
	#r := make(AF, len(a))
	#for j := range r {
		#r[j] = a[j] ${op} F(i)
	#}
	#return r
#}

#func ${name}AFF(a AF, f F) AF {
	#r := make(AF, len(a))
	#for j := range r {
		#r[j] = a[j] ${op} f
	#}
	#return r
#}

#func ${name}AIAI(a1 AI, a2 AI) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AI, len(a1))
	#for j := range r {
		#r[j] = a1[j] ${op} a2[j]
	#}
	#return r
#}

#func ${name}AIAF(a1 AI, a2 AF) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AF, len(a1))
	#for j := range r {
		#r[j] = F(a1[j]) ${op} a2[j]
	#}
	#return r
#}

#func ${name}AFAF(a1 AF, a2 AF) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AF, len(a1))
	#for j := range r {
		#r[j] = a1[j] ${op} a2[j]
	#}
	#return r
#}

#func ${name}AOI(a AO, i I) Object {
	#r := make(AO, len(a))
	#for j := range r {
		#r[j] = ${name}I(i, a[j])
		#err, ok := r[j].(E)
		#if ok {
			#return err
		#}
	#}
	#return r
#}

#func ${name}AOF(a AO, f F) Object {
	#r := make(AO, len(a))
	#for j := range r {
		#r[j] = ${name}F(f, a[j])
		#err, ok := r[j].(E)
		#if ok {
			#return err
		#}
	#}
	#return r
#}

#func ${name}AOAI(a1 AO, a2 AI) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AO, len(a1))
	#for j := range r {
		#r[j] = ${name}I(a2[j], a1[j])
	#}
	#return r
#}

#func ${name}AOAF(a1 AO, a2 AF) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AO, len(a1))
	#for j := range r {
		#r[j] = ${name}F(a2[j], a1[j])
	#}
	#return r
#}

#func ${name}AOAO(a1 AO, a2 AO) Object {
	#if len(a1) != len(a2) {
		#return badlen("${op}")
	#}
	#r := make(AO, len(a1))
	#for j := range r {
		#r[j] = ${name}(a2[j], a1[j])
	#}
	#return r
#}
#EOS
#}
