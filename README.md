# Goal

[![pkg.go.dev](https://pkg.go.dev/badge/codeberg.org/anaseto/goal.svg)](https://pkg.go.dev/codeberg.org/anaseto/goal)
[![godocs.io](https://godocs.io/codeberg.org/anaseto/goal?status.svg)](https://godocs.io/codeberg.org/anaseto/goal)

Goal is an embeddable array programming language with a bytecode interpreter,
written in Go. It provides both a command line intepreter (that can be used in
the REPL), and a library interface. Though the language is still young, its
core features are mostly there and tested, so it is usable both for writing
useful short scripts and playing with the REPL. The library interface might
still experiment some changes.

Like in most array programming languages, Goal's builtins vectorize operations
on immutable arrays, and encourage a functional style for control and data
transformations, supported by a simple dynamic type system with little
abstraction, and mutable variables (but no mutable values).

It's main distinctive features are as follows:

* Syntax inspired mainly from the K language, but with quite a few deviations.
  For example, backquotes produce Go-like raw strings instead of symbols,
  rx/[a-z]/ is a regular expression literal (checked and processed at
  compile-time), and there are no tacit compositions, and digraph operator
  verbs and adverbs are gone or done differently (except for global assignment
  with ::).
* Primitive semantics are both inspired from the
  [ngn/k](https://codeberg.org/ngn/k) variant of the K language and
  [BQN](https://mlochbaum.github.io/BQN/index.html). For example, group,
  classify, shifts, windows, find (index of) and occurrence count take after
  BQN's semantics. Multi-dimensional versions, when present in BQN, have been
  left out, though, as Goal has only free-form immutable arrays, like K.  Some
  primitives use words instead of symbols (like ocount for occurrence count).
  Also, K-like dictionnaries are supported, but not tables.
* Unlike in typical array languages, strings are atoms, and common string
  handling functions (like index, substr or trim) have been integrated into the
  primitives, including regular expression functions.
* Error handling makes a distinction between fatal errors (panics) and
  recoverable errors which are handled as values.
* Easily embeddable and extensible in Go, meaning easy access to the standard
  library.
* Integrated support for csv and time handling.
* Array performance is unsurprising and good enough most of the time, with
  basic (but good in code with limited branching) variable liveness analysis to
  reduce cloning by reusing dead immutable arrays, though it is not a goal to
  reach state-of-the-art (no SIMD, and there is still room for more special
  code and specialized algorithms). Scalar performance is typical for a
  bytecode-compiled interpreter (without JIT), somewhat slower than a C
  bytecode interpreter (value representation in Go is somewhat less compact
  than how it can be done in C).

If this list is not enough to satisfy your curiosity, there's also a
[Why.md](docs/Why.md) text for you. You can also read the [Credits.md](Credits.md)
to know about main inspiration sources for the language. Last, but not least,
there are some [implementation notes](docs/Implementation.md) too.

# Install

To install the command line interpreter, first do the following:

+ Install the [go compiler](https://golang.org/).
+ Add `$(go env GOPATH)/bin` to your `$PATH` (for example `export PATH="$PATH:$(go env GOPATH)/bin"`).

Then you can build the intepreter with:

	go install ./cmd/goal

Alternatively, you may type `go build -o /path/to/bin/goal ./cmd/goal` to put
the resulting binary in a custom location in your $PATH.

The `goal` command should now be available. Type `goal --help` for command-line
usage.

Typing just `goal` opens the REPL. For a better experience using the REPL (to
get typical keyboard shortcuts), you can install the readline wrapper `rlwrap`
program (available as a package in most systems), and then use instead `rlwrap
goal`.

# Editor support

- [vim-goal](https://codeberg.org/anaseto/vim-goal) : vim files for Goal.

# Examples

A few short examples can be found in the `testdata/scripts` directory. Because
they're used for testing, they come along an expected result after a `/RESULT:`
comment.

Also, various code generation scripts in the toplevel `scripts` directory are
written in Goal.

# Documentation

Currently, documentation consists of the REPL help system with usage summary and
examples for all implemented features, so some prior knowledge of another array
language, in particular K, can be useful. The full contents are replicated
below.

```
TOPICS HELP
Type help TOPIC or h TOPIC where TOPIC is one of:

"syn"   syntax
"types" types
"+"     verbs (like +*-%,)
"nv"    named verbs (like in, sign)
"'"     adverbs ('/\)
"io"    io functions (slurp, say)
"time"  time functions
"goal"  runtime functions

Notations:
        s (string) f (function) F (2-args function)
        n (number) i (integer) r (regexp) d (dict)
        x,y (any other)

SYNTAX HELP
numbers         1     1.5     0b0110     1.7e-3
strings         "text\xff\u00F\n"  "\""  "\u65e5"  (backquotes for raw strings)
arrays          1 2 -3 4      1 "ab" -2 "cd"      (1 2;"a";3 "b";(4 2;"c");*)
regexps         rx/[a-z]/     (see https://pkg.go.dev/regexp/syntax for syntax)
operators       :  +  -  *  %  !  &  |  ^  #  _  $  ?  @  .  ::
expressions     2*3+4 -> 14 (no priority)    1+|1 2 3 -> 4 3 2     +/1 2 3 -> 6
separator       ; or newline (except ignored around parens, brackets and braces)
variables       a   b   f   data    (any word matching rx/[a-zA-Z][a-zA-Z0-9]*/)
assign          a:2 (local within lambda, global otherwise)    a::2 (global)    
op assign       a+:1 (sugar for a:a+1)       a+::2 (sugar for a::a+2)
list assign     (a;b;c):x   (where 2<#x)     (a;b):1 2;b -> 2
index           x[y] or x y is sugar for x@y; x[] ~ x[*] ~ x[!#x] ~ x (arrays)
index deep      x[y;z;...] is sugar for x.(y;z;...) (except for x in (?;and;or))
index assign    x[y]:z is sugar for x:@[x;y;:;z]    (or . for x[y;...]:z)
index op assign x[y]op:z is sugar for x:@[x;y;op;z] (for symbol operator)
lambdas         {x+y+z}[2;3;0] -> 5     {[a;b;c]a+b+c}[1;2;3] -> 6
projections     {x+y}[2;] 3 -> 5        (2+) 3 -> 5
cond            ?[1;2;3] -> 2     ?[0;2;3] -> 3    ?[0;2;"";3;4] -> 4
and/or          and[1;2] -> 2   and[1;0;3] -> 0   or[0;2] -> 2   or[0;0;0] -> 0
sequence        [a:2;b:a+3;a+10] -> 12 (bracket block [] at start of expression)
return          [1;:2;3] -> 2 (a : at start of expression)
try             'x is sugar for ?["e"~@x;:x;x] (return if it's an error)
comments        from line with a single / until line with a single \
                or from / (after space or start of line) to end of line

TYPES HELP
atom    array   name            examples
n       N       number          0      1.5      !5      1.2 3 1.8
s       S       string          "abc"   "d"     "a" "b" "c"
r               regexp          rx/[a-z]/       rx/\s+/
d               dictionnary     "a" "b"!1 2
f               function        +      {x*2}   (1-)    %[;2]
h               handle          open "/path/to/file"    "w" open "/path/to/file"
e               error           error "msg"
        A       generic array   ("a" 1;"b" 2;"c" 3)     (+;-;*;"any")

VERBS HELP
:x  identity    :[42] -> 42 (recall that : is also syntax for return)
x:y right       2:3 -> 3        "a":"b" -> "b"
+x  flip        +(1 2;3 4) -> (1 3;2 4)         +42 -> ,,42
n+n add         2+3 -> 5            2+3 4 -> 5 6
s+s concat      "a"+"b" -> "ab"     "a" "b"+"c" -> "ac" "bc"
-x  negate      - 2 3 -> -2 -3      -(1 2.5;3 4) -> (-1 -2.5;-3 -4)
n-n subtract    5-3 -> 2            5 4-3 -> 2 1
s-s trim suffix "file.txt"-".txt" -> "file"
*x  first       *3 2 4 -> 3     *"ab" -> "ab"    *(+;*) -> +
n*n multiply    2*3 -> 6            1 2 3*3 -> 3 6 9
s*i repeat      "a"*3 2 1 0 -> "aaa" "aa" "a" ""
%x  classify    %1 2 3 1 2 3 -> 0 1 2 0 1 2     %"a" "b" "a" -> 0 1 0
x%y divide      3%2 -> 1.5          3 4%2 -> 2 1.5
!i  enum        !5 -> 0 1 2 3 4
!d  keys        !"a" "b"!1 2 -> "a" "b"
!x  odometer    !2 3 -> (0 0 0 1 1 1;0 1 2 0 1 2)
i!y colsplit    2!!6 -> (0 1;2 3;4 5)   2!"a" "b" "c" -> ("a" "b";,"c")
x!y dict        d:"a" "b"!1 2;d "a" -> 1
&I  where       &0 0 1 0 0 0 1 -> 2 6           &2 3 -> 0 0 1 1 1
&d  keys where  &"a" "b" "e" "c"!0 1 1 0 -> "b" "e"
x&y min         2&3 -> 2        4&3 -> 3        "b"&"a" -> "a"
|x  reverse     |!5 -> 4 3 2 1 0
x|y max         2|3 -> 3        4|3 -> 4        "b"|"a" -> "b"
<x  ascend      <2 4 3 -> 0 2 1 (index permutation for ascending order)
x<y less        2<3 -> 1        "c" < "a" -> 0
>x  descend     >2 4 3 -> 1 2 0 (index permutation for descending order)
x>y greater     2>3 -> 0        "c" > "a" -> 1
=I  group       =1 0 2 1 2 -> (,1;0 3;2 4)      =-1 2 -1 2 -> (!0;!0;1 3)
=d  group keys  ="a""b""c"!0 1 0 -> ("a" "c";,"b")
f=y group by    {1=2!x}=!10 -> (0 2 4 6 8;1 3 5 7 9)
x=y equal       2 3 4=3 -> 0 1 0        "ab" = "ba" -> 0
~x  not         ~0 1 2 -> 1 0 0         ~"a" "" "0" -> 0 1 0
x~y match       3~3 -> 1        2 3~3 2 -> 0       ("a";%)~("b";%) -> 0 1
,x  enlist      ,1 -> ,1 (list with one element)
d,d merge       ("a""b"!1 2),"b""c"!3 4 -> "a""b""c"!1 3 4
x,y join        1,2 -> 1 2      "ab" "c","d" -> "ab" "c" "d"
^x  sort        ^3 5 0 -> 0 3 5       ^"ca" "ab" "bc" -> "ab" "bc" "ca"
i^s windows     2^"abcd" -> "ab" "bc" "cd"      (2-bytes strings)
i^y windows     2^!4 -> (0 1;1 2;2 3)
s^y trim        " []"^"  [text]  " -> "text"    "\n"^"\nline\n" -> "line"
x^y without     2 3^1 1 2 3 3 4 -> 1 1 4
#x  length      #2 4 5 -> 3      #"ab" "cd" -> 2      #42 -> 1     #"ab" -> 1
i#y take        2#4 1 5 -> 4 1    4#3 1 5 -> 3 1 5 3 (cyclic)    3#1 -> 1 1 1
s#y count       "ab"#"cabdab" "cd" "deab" -> 2 0 1
f#y replicate   {0 1 1 0}#4 1 5 3 -> 1 5    {x>0}#2 -3 1 -> 2 1
x#y keep only   2 3^1 1 2 3 3 4 -> 2 3 3
_n  floor       _2.3 -> 2           _1.5 3.7 -> 1 3
_s  to lower    _"ABC" -> "abc"     _"AB" "CD" -> "ab" "cd"
i_s drop bytes  2_"abcde" -> "cde"  -2_"abcde" -> "abc"
i_y drop        2_3 4 5 6 -> 5 6    -2_3 4 5 6 -> 3 4
s_i delete      "abc"_1 -> "ac"
x_i delete      4 3 2 1_1 -> 4 2 1      4 3 2 1_-3 -> 4 2 1
s_s trim prefix "pref-"_"pref-name" -> "name"
I_s cut string  1 3_"abcdef" -> "bc" "def"      (I ascending)
I_y cut         2 5_!10 -> (2 3 4;5 6 7 8 9)    (I ascending)
f_y weed out    {0 1 1 0}_4 1 5 3 -> 4 3    {x>0}_2 -3 1 -> ,-3
$x  string      $2 3 -> "2 3"     $"text" -> "\"text\""
i$s pad         3$"a" -> "a  "    -3$"1" "23" "456" -> "  1" " 23" "456"
s$y cast        "i"$2.3 -> 2    "i"$"ab" -> 97 98   "s"$97 98 -> "ab"
s$s parse num   "n"$"1.5" -> 1.5        "n"$"2" "1e+7" "0b100" -> 2 1e+07 4
x$y binsearch   2 3 5 7$8 2 7 5 5.5 3 0 -> 4 1 4 3 3 2 0
?i  uniform     ?2 -> 0.6046602879796196 0.9405090880450124
?x  uniq        ?2 2 3 4 3 3 -> 2 3 4
i?y roll        5?100 -> 10 51 21 51 37
i?y deal        -5?100 -> 19 26 0 73 94 (always distinct)
s?r rindex      "abcde"?rx/b../ -> 1 4
s?s index       "a = a + 1"?"=" "+" -> 2 6
d?y find key    ("a" "b"!3 4)?4 -> "b"      ("a" "b"!3 4)?5 -> ""
x?y find        3 2 1?2 -> 1    3 2 1?0 -> 3
@x  type        @2 -> "n"    @"ab" -> "s"    @2 3 -> "N"       @+ -> "f"
s@y substr      "abcdef"@2  -> "cdef" (s[offset])
r@y match       rx/[a-z]/"abc" -> 1     rx/\s/"abc" -> 0
r@y find group  m:rx/[a-z](.)/"abc" -> "ab" "b" (m[0] whole match, m[1] group)
r@y findN       rx/[a-z]/["abc";2] -> "a""b"    rx/[a-z]/["abc";-1] -> "a""b""c"
f@y apply       (|)@1 2 -> 2 1 (like |[1 2] -> 2 1 or |1 2)
d@y at key      ("a" "b"!1 2)@"a" -> 1
x@y at          1 2 3@2 -> 3     1 2 3[2 0] -> 3 1     7 8 9@-2 -> 8
.s  reval       ."2+3" -> 5     a:1;."a" -> panic ".s : undefined global: a"
.e  get error   .error "msg" -> "msg"
.d  values      ."a" "b"!1 2 -> 1 2
s.y substr      "abcdef"[2;3] -> "cde" (s[offset;length])
r.y findN       rx/[a-z]/["abc";2] -> "a" "b" (stop at 2 matches; -1 for all)
x.y applyN      {x+y}.2 3 -> 5    {x+y}[2;3] -> 5    (1 2;3 4)[0;1] -> 2

::x         get global  a:3;::"a" -> 3
::[x;y]     set global  ::["a";3];a -> 3
@[x;y;f]    amend       @[1 2 3;0 1;10+] -> 11 12 3
@[x;y;F;z]  amend       @[8 4 5;(1 2;0);+;(10 5;-2)] -> 6 14 10
.[x;y;f]    deep amend  .[(1 2;3 4);0 1;-] -> (1 -2;3 4)
.[x;y;F;z]  deep amend  .[(1 2;3 4);(0 1 0;1);+;1] -> (1 4;3 5)
                        .[(1 2;3 4);(*;1);:;42] -> (1 42;3 42)
.[f;x;f]    try         .[+;2 3;{"msg"}] -> 5   .[+;2 "a";{"msg"}] -> "msg"

NAMED VERBS HELP
abs n     abs value     abs -3 -1.5 2 -> 3 1.5 2
bytes s   byte-count    bytes "abc" -> 3
ceil n    ceil/upper    ceil 1.5 -> 2   ceil "ab" -> "AB"
error x   error         r:{?[~x=0;1%x;error "zero"]}0;?["e"~@r;.r;r] -> "zero"
eval s    eval          a:5;eval "a+2" -> 7 (unrestricted eval)
firsts x  mark firsts   firsts 0 0 2 3 0 2 3 4 -> 1 0 1 1 0 0 0 1
icount x  index-count   icount 0 0 1 -1 0 1 2 3 2 -> 3 2 2 1 (same as #'=x)
ocount x  occur-count   ocount 3 2 5 3 2 2 7 -> 0 0 0 1 1 2 0
panic s   panic         panic "msg" (for fatal programming-errors)
rshift x  right shift   rshift 1 2 -> 0 1       rshift "a" "b" -> "" "a"
rx s      comp. regex   rx "[a-z]"  (like rx/[a-z]/ but compiled at runtime)
seed i    rand seed     seed 42 (for non-secure pseudo-rand with ?)
shift x   left shift    shift 1 2 -> 2 0        shift "a" "b" -> "b" ""
sign n    sign          sign -3 -1 0 1.5 5 -> -1 -1 0 1 1

x csv y     csv read    csv "1,2,3" -> ,"1" "2" "3"
                        " " csv "1 2 3" -> ,"1" "2" "3" (" " as separator)
            csv write   csv ,"1" "2" "3" -> "1,2,3\n"
                        " " csv ,"1" "2" "3" -> "1 2 3\n"
x in s      contained   "bc" "ac" in "abcd" -> 1 0
x in y      member of   2 3 in 0 2 4 -> 1 0
x nan y     fill NaNs   42 nan (1.5;sqrt -1) -> 1.5 42
n mod n     modulus     3 mod 5 4 3 -> 2 1 0
x rotate y  rotate      2 rotate 1 2 3 4 -> 3 4 1 2
x rshift y  right shift "a" "b" rshift 1 2 3 -> "a" "b" 1
x shift y   left shift  "a" "b" shift 1 2 3 -> 3 "a" "b"

sub[r;s]    regsub      sub[rx/[a-z]/;"Z"] "aBc" -> "ZBZ"
sub[r;f]    regsub      sub[rx/[A-Z]/;_] "aBc" -> "abc"
sub[s;s]    replace     sub["b";"B"] "abc" -> "aBc"
sub[s;s;i]  replaceN    sub["a";"b";2] "aaa" -> "bba" (stop after 2 times)
sub[S]      replaceS    sub["b" "d" "c" "e"] "abc" -> "ade"
sub[S;S]    replaceS    sub["b" "c";"d" "e"] "abc" -> "ade"

eval[s;n;p] eval        like eval s, but provide name n as location and prefix
                        p for globals

MATH: acos, asin, atan, cos, exp, log, round, sin, sqrt, tan, nan
UTF-8: utf8.rcount (number of code points), utf8.valid

ADVERBS HELP
f'x    each      #'(4 5;6 7 8) -> 2 3
x F'y  each      2 3#'1 2 -> (1 1;2 2 2)    {(x;y;z)}'[1;2 4;3] -> (1 2 3;1 4 3)
F/x    fold      +/!10 -> 45
F\x    scan      +\!10 -> 0 1 3 6 10 15 21 28 36 45
x F/y  fold      1 2+/!10 -> 46 47                 {x+y-z}/[9;3 4;2 7] -> 7
x F\y  scan      5 6+\1 2 3 -> (6 7;8 9;11 12)     {x+y-z}\[9;3 4;2 7] -> 10 7
i f/y  do        3{x*2}/4 -> 32
i f\y  dos       3{x*2}\4 -> 4 8 16 32
f f/y  while     {x<100}{x*2}/4 -> 128
f f\y  whiles    {x<100}{x*2}\4 -> 4 8 16 32 64 128
f/x    converge  {1+1.0%x}/1 -> 1.618033988749895     {-x}/1 -> -1
f\x    converges {_x%2}\10 -> 10 5 2 1 0              {-x}\1 -> 1 -1
s/x    join      ","/"a" "b" "c" -> "a,b,c"
s\x    split     ","\"a,b,c" -> "a" "b" "c"
r\x    split     rx/[,;]/\"a,b;c" -> "a" "b" "c"
i s\y  splitN    (2) ","\"a,b,c" -> "a" "b,c"
I/x    encode    24 60 60/1 2 3 -> 3723  2/1 1 0 -> 6
I\x    decode    24 60 60\3723 -> 1 2 3  2\6 -> 1 1 0

IO/OS HELP
close h      flush any buffered data, then close filehandle h
import s     eval file s+".goal" and import globals with prefix s+"."
open s       open path s for reading, returning a filehandle (h)
print s      print "Hello, world!\n" (works for non-string values too)
read h       read from filehandle h until EOF or an error occurs.
run s        run command            run "pwd"        run "ls" "-l"
             inherits stdin, stdout, and stderr, returns true on success
say s        same as print, but appends a newline    say !5
shell s      run command as-is through the shell     shell "ls -l"
             inherits stderr, returns its output
slurp s      read file named s       lines:"\n"\slurp["/path/to/file"]

p import s   like import s but with prefix p+"." for globals
m open s     open path s with mode m in "r" "r+" "w" "w+" "a" "a+"
             or pipe from (mode "-|") or to (mode "|-") command (s or S)
h print s    print s to writer or filename h     "filename" print "content"
n read h     read n bytes from reader h or until EOF, or an error occurs
s read h     read from reader h until 1-byte s, EOF, or an error occurs
h say s      same as print, but appends a newline

os.STDIN     standard input filehandle
os.STDOUT    standard output filehandle
os.STDERR    standard error filehandle
os.ARGS      command-line arguments, starting with script name
os.ENV       keys!values strings dictionnary representing environment

TIME HELP
time cmd              time command with current time
cmd time t            time command with time t
time[cmd;t;fmt]       time command with time t in given format
time[cmd;t;fmt;loc]   time command with time t in given format and location

Time t should be either an integer representing unix epochtime, or a string
in the given format (RFC3339 format layout "2006-01-02T15:04:05Z07:00" is the
default). See https://pkg.go.dev/time for information on layouts and locations,
as goal uses the same conventions as Go's time package.

Currently available commands:
        "day"           day number (i)
        "date"          year, month, day (I)
        "clock"         hour, minute, second (I)
        "hour"          0-23 hour (i)
        "minute"        0-59 minute (i)
        "second"        0-59 second (i)
        "unix"          unix epoch time (i)
        "unixmilli"     unix (millisecond version, only for current time) (i)
        "unixmicro"     unix (microsecond version, only for current time) (i)
        "unixnano"      unix (nanosecond version, only for current time) (i)
        "year"          year (i)
        "yearday"       1-365/6 year day (i)
        "week"          year, week (I)
        "weekday"       0-7 weekday (starts from Sunday) (i)
        format (s)      format time using given layout (s)

RUNTIME HELP
goal[cmd;...]   runtime command cmd for goal

The supported command forms are as follows:

"prec" goal n   set floating point formatting precision to n (default -1)
```
