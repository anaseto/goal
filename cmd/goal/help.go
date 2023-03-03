// Code generated by scripts/help.goal. DO NOT EDIT.

package main

const helpTopics = "TOPICS HELP\nType help TOPIC or h TOPIC where TOPIC is one of:\n\n\"syn\"   syntax\n\"types\" value types\n\"+\"     verbs (like +*-%,)\n\"nv\"    named verbs (like in, sign)\n\"'\"     adverbs ('/\\)\n\"io\"    IO functions (like say, open, read)\n\"time\"  time handling\n\"goal\"  runtime system\n\nNotations:\n        s (string) f (function) F (2-args function)\n        n (number) i (integer) r (regexp) d (dict)\n        x,y (any other)\n"

const helpSyntax = "SYNTAX HELP\nnumbers         1     1.5     0b0110     1.7e-3     0xab\nstrings         \"text\\xff\\u00F\\n\"   \"\\\"\"   \"\\u65e5\"   \"interpolated $var\"\n                qq/$var\\n or ${var}/   qq#text#  (delimiters :+-*%!&|=~,^#_?@/')\nraw strings     `anything until first backquote`     `literal \\, no escaping`\n                rq/anything until single slash/      rq#doubling ## escapes #\narrays          1 2 -3 4      1 \"ab\" -2 \"cd\"      (1 2;\"a\";3 \"b\";(4 2;\"c\");*)\nregexps         rx/[a-z]/      (see https://pkg.go.dev/regexp/syntax for syntax)\nverbs           : + - * % ! & | < > = ~ , ^ # _ $ ? @ . ::   (right-associative)\n                abs bytes ceil error ...\nadverbs         / \\ '                                        (left-associative)\nexpressions     2*3+4 -> 14     1+|1 2 3 -> 4 3 2     +/'(1 2 3;4 5 6) -> 6 15\nseparator       ; or newline except when ignored after {[( and before )]}\nvariables       a  b.c  f  data  t1\nassign          a:2 (local within lambda, global otherwise)    a::2 (global)    \nop assign       a+:1 (sugar for a:a+1)       a-::2 (sugar for a::a-2)\nlist assign     (a;b;c):x (where 2<#x)       (a;b):1 2;b -> 2\nindex           x[y] or x y is sugar for x@y; x[] ~ x[*] ~ x[!#x] ~ x (arrays)\nindex deep      x[y;z;...] is sugar for x.(y;z;...) (except for x in (?;and;or))\nindex assign    x[y]:z is sugar for x:@[x;y;:;z]    (or . for x[y;...]:z)\nindex op assign x[y]op:z is sugar for x:@[x;y;op;z] (for symbol operator)\nlambdas         {x+y-z}[7;3;5] -> 5     {[a;b;c]a+b-c}[7;3;5] -> 5\nprojections     +[2;] 3 -> 5            (2+) 3 -> 5\ncond            ?[1;2;3] -> 2     ?[0;2;3] -> 3    ?[0;2;\"\";3;4] -> 4\nand/or          and[1;2] -> 2   and[1;0;3] -> 0   or[0;2] -> 2   or[0;0;0] -> 0\nsequence        [a:2;b:a+3;a+10] -> 12 (bracket block [] at start of expression)\nreturn          [1;:2;3] -> 2 (a : at start of expression)\ntry             'x is sugar for ?[\"e\"~@x;:x;x] (return if it's an error)\ncomments        from line with a single / until line with a single \\\n                or from / (after space or start of line) to end of line\n"

const helpTypes = "TYPES HELP\natom    array   name            examples\nn       N       number          0      1.5      !5      1.2 3 1.8\ns       S       string          \"abc\"    \"d\"    \"a\" \"b\" \"c\"\nr               regexp          rx/[a-z]/       rx/\\s+/\nd               dictionnary     \"a\" \"b\"!1 2\nf               function        +      {x*2}      (1-)      %[;2]\nh               handle          open \"/path/to/file\"    \"w\" open \"/path/to/file\"\ne               error           error \"msg\"\n        A       generic array   (\"a\" 1;\"b\" 2;\"c\" 3)     (+;-;*;\"any\")\n"

const helpVerbs = "VERBS HELP\n:x  identity    :[42] -> 42 (recall that : is also syntax for return)\nx:y right       2:3 -> 3        \"a\":\"b\" -> \"b\"\n+x  flip        +(1 2;3 4) -> (1 3;2 4)         +42 -> ,,42\nn+n add         2+3 -> 5            2+3 4 -> 5 6\ns+s concat      \"a\"+\"b\" -> \"ab\"     \"a\" \"b\"+\"c\" -> \"ac\" \"bc\"\n-x  negate      - 2 3 -> -2 -3      -(1 2.5;3 4) -> (-1 -2.5;-3 -4)\nn-n subtract    5-3 -> 2            5 4-3 -> 2 1\ns-s trim suffix \"file.txt\"-\".txt\" -> \"file\"\n*x  first       *3 2 4 -> 3     *\"ab\" -> \"ab\"    *(+;*) -> +\nn*n multiply    2*3 -> 6            1 2 3*3 -> 3 6 9\ns*i repeat      \"a\"*3 2 1 0 -> \"aaa\" \"aa\" \"a\" \"\"\n%x  classify    %1 2 3 1 2 3 -> 0 1 2 0 1 2     %\"a\" \"b\" \"a\" -> 0 1 0\nx%y divide      3%2 -> 1.5          3 4%2 -> 2 1.5\n!i  enum        !5 -> 0 1 2 3 4\n!d  keys        !\"a\" \"b\"!1 2 -> \"a\" \"b\"\n!x  odometer    !2 3 -> (0 0 0 1 1 1;0 1 2 0 1 2)\ni!y colsplit    2!!6 -> (0 1;2 3;4 5)   2!\"a\" \"b\" \"c\" -> (\"a\" \"b\";,\"c\")\nx!y dict        d:\"a\" \"b\"!1 2;d \"a\" -> 1\n&I  where       &0 0 1 0 0 0 1 -> 2 6           &2 3 -> 0 0 1 1 1\n&d  keys where  &\"a\" \"b\" \"e\" \"c\"!0 1 1 0 -> \"b\" \"e\"\nx&y min         2&3 -> 2        4&3 -> 3        \"b\"&\"a\" -> \"a\"\n|x  reverse     |!5 -> 4 3 2 1 0\nx|y max         2|3 -> 3        4|3 -> 4        \"b\"|\"a\" -> \"b\"\n<x  ascend      <2 4 3 -> 0 2 1 (index permutation for ascending order)\nx<y less        2<3 -> 1        \"c\" < \"a\" -> 0\n>x  descend     >2 4 3 -> 1 2 0 (index permutation for descending order)\nx>y greater     2>3 -> 0        \"c\" > \"a\" -> 1\n=I  group       =1 0 2 1 2 -> (,1;0 3;2 4)      =-1 2 -1 2 -> (!0;!0;1 3)\n=d  group keys  =\"a\"\"b\"\"c\"!0 1 0 -> (\"a\" \"c\";,\"b\")\nf=y group by    {2 mod x}=!10 -> (0 2 4 6 8;1 3 5 7 9)\nx=y equal       2 3 4=3 -> 0 1 0        \"ab\" = \"ba\" -> 0\n~x  not         ~0 1 2 -> 1 0 0         ~\"a\" \"\" \"0\" -> 0 1 0\nx~y match       3~3 -> 1        2 3~3 2 -> 0       (\"a\";%)~(\"b\";%) -> 0 1\n,x  enlist      ,1 -> ,1 (list with one element)\nd,d merge       (\"a\"\"b\"!1 2),\"b\"\"c\"!3 4 -> \"a\"\"b\"\"c\"!1 3 4\nx,y join        1,2 -> 1 2      \"ab\" \"c\",\"d\" -> \"ab\" \"c\" \"d\"\n^x  sort        ^3 5 0 -> 0 3 5       ^\"ca\" \"ab\" \"bc\" -> \"ab\" \"bc\" \"ca\"\ni^s windows     2^\"abcd\" -> \"ab\" \"bc\" \"cd\"      (2-bytes strings)\ni^y windows     2^!4 -> (0 1;1 2;2 3)\ns^y trim        \" []\"^\"  [text]  \" -> \"text\"    \"\\n\"^\"\\nline\\n\" -> \"line\"\nx^y without     2 3^1 1 2 3 3 4 -> 1 1 4\n#x  length      #2 4 5 -> 3      #\"ab\" \"cd\" -> 2      #42 -> 1     #\"ab\" -> 1\ni#y take        2#4 1 5 -> 4 1    4#3 1 5 -> 3 1 5 3 (cyclic)    3#1 -> 1 1 1\ns#y count       \"ab\"#\"cabdab\" \"cd\" \"deab\" -> 2 0 1\nf#y replicate   {0 1 1 0}#4 1 5 3 -> 1 5    {x>0}#2 -3 1 -> 2 1\nx#y keep only   2 3^1 1 2 3 3 4 -> 2 3 3\n_n  floor       _2.3 -> 2           _1.5 3.7 -> 1 3\n_s  to lower    _\"ABC\" -> \"abc\"     _\"AB\" \"CD\" -> \"ab\" \"cd\"\ni_s drop bytes  2_\"abcde\" -> \"cde\"  -2_\"abcde\" -> \"abc\"\ni_y drop        2_3 4 5 6 -> 5 6    -2_3 4 5 6 -> 3 4\ns_i delete      \"abc\"_1 -> \"ac\"\nx_i delete      4 3 2 1_1 -> 4 2 1      4 3 2 1_-3 -> 4 2 1\ns_s trim prefix \"pref-\"_\"pref-name\" -> \"name\"\nI_s cut string  1 3_\"abcdef\" -> \"bc\" \"def\"      (I ascending)\nI_y cut         2 5_!10 -> (2 3 4;5 6 7 8 9)    (I ascending)\nf_y weed out    {0 1 1 0}_4 1 5 3 -> 4 3    {x>0}_2 -3 1 -> ,-3\n$x  string      $2 3 -> \"2 3\"     $\"text\" -> \"\\\"text\\\"\"\ni$s pad         3$\"a\" -> \"a  \"    -3$\"1\" \"23\" \"456\" -> \"  1\" \" 23\" \"456\"\ns$y cast        \"i\"$2.3 -> 2    \"i\"$\"ab\" -> 97 98   \"s\"$97 98 -> \"ab\"\ns$s parse num   \"n\"$\"1.5\" -> 1.5        \"n\"$\"2\" \"1e+7\" \"0b100\" -> 2 1e+07 4\nx$y binsearch   2 3 5 7$8 2 7 5 5.5 3 0 -> 4 1 4 3 3 2 0   (x ascending)\n?i  uniform     ?2 -> 0.6046602879796196 0.9405090880450124\n?x  uniq        ?2 2 3 4 3 3 -> 2 3 4\ni?y roll        5?100 -> 10 51 21 51 37\ni?y deal        -5?100 -> 19 26 0 73 94 (always distinct)\ns?r rindex      \"abcde\"?rx/b../ -> 1 3      (offset;length)\ns?s index       \"a = a + 1\"?\"=\" \"+\" -> 2 6\nd?y find key    (\"a\" \"b\"!3 4)?4 -> \"b\"      (\"a\" \"b\"!3 4)?5 -> \"\"\nx?y find        3 2 1?2 -> 1                 3 2 1?0 -> 3\n@x  type        @2 -> \"n\"    @\"ab\" -> \"s\"    @2 3 -> \"N\"       @+ -> \"f\"\ns@y substr      \"abcdef\"@2  -> \"cdef\" (s[offset])\nr@y match       rx/[a-z]/\"abc\" -> 1     rx/\\s/\"abc\" -> 0\nr@y find group  m:rx/[a-z](.)/\"abc\" -> \"ab\" \"b\" (m[0] whole match, m[1] group)\nf@y apply       (|)@1 2 -> 2 1 (like |[1 2] -> 2 1 or |1 2)\nd@y at key      (\"a\" \"b\"!1 2)@\"a\" -> 1\nx@y at          1 2 3@2 -> 3      1 2 3[2 0] -> 3 1      7 8 9@-2 -> 8\n.s  reval       .\"2+3\" -> 5    (restricted eval with new context: see also eval)\n.e  get error   .error \"msg\" -> \"msg\"\n.d  values      .\"a\" \"b\"!1 2 -> 1 2\ns.y substr      \"abcdef\"[2;3] -> \"cde\" (s[offset;length])\nr.y findN       rx/[a-z]/[\"abc\";2] -> \"a\"\"b\"    rx/[a-z]/[\"abc\";-1] -> \"a\"\"b\"\"c\"\nr.y findN group rx/[a-z](.)/[\"abcdef\";2] -> (\"ab\" \"b\";\"cd\" \"d\")\nx.y applyN      {x+y}.2 3 -> 5    {x+y}[2;3] -> 5    (1 2;3 4)[0;1] -> 2\n«x  shift       «1 2 -> 2 0    «\"a\" \"b\" -> \"b\" \"\"  (ASCII alternative: shift x)\nx«y shift       \"a\" \"b\"«1 2 3 -> 3 \"a\" \"b\"\n»x  rshift      »1 2 -> 0 1    »\"a\" \"b\" -> \"\" \"a\"  (ASCII alternative: rshift x)\nx»y rshift      \"a\" \"b\"»1 2 3 -> \"a\" \"b\" 1\n\n::x         get global  a:3;::\"a\" -> 3\nx::y        set global  \"a\"::3;a -> 3\n@[x;y;f]    amend       @[1 2 3;0 1;10+] -> 11 12 3\n@[x;y;F;z]  amend       @[8 4 5;(1 2;0);+;(10 5;-2)] -> 6 14 10\n.[x;y;f]    deep amend  .[(1 2;3 4);0 1;-] -> (1 -2;3 4)\n.[x;y;F;z]  deep amend  .[(1 2;3 4);(0 1 0;1);+;1] -> (1 4;3 5)\n                        .[(1 2;3 4);(*;1);:;42] -> (1 42;3 42)\n.[f;x;f]    try         .[+;2 3;{\"msg\"}] -> 5   .[+;2 \"a\";{\"msg\"}] -> \"msg\"\n"

const helpNamedVerbs = "NAMED VERBS HELP\nabs n      abs value    abs -3 -1.5 2 -> 3 1.5 2\nbytes s    byte-count   bytes \"abc\" -> 3\nceil x     ceil/upper   ceil 1.5 -> 2   ceil \"ab\" -> \"AB\"\nerror x    error        r:{?[~x=0;1%x;error \"zero\"]}0;?[\"e\"~@r;.r;r] -> \"zero\"\neval s     comp/run     a:5;eval \"a+2\" -> 7         (unrestricted variant of .s)\nfirsts x   mark firsts  firsts 0 0 2 3 0 2 3 4 -> 1 0 1 1 0 0 0 1\nocount x   occur-count  ocount 3 2 5 3 2 2 7 -> 0 0 0 1 1 2 0\npanic s    panic        panic \"msg\" (for fatal programming-errors)\nrx s       comp. regex  rx \"[a-z]\"  (like rx/[a-z]/ but compiled at runtime)\nsign n     sign         sign -3 -1 0 1.5 5 -> -1 -1 0 1 1\n\nx csv y    csv read     csv \"1,2,3\" -> ,\"1\" \"2\" \"3\"\n                        \" \" csv \"1 2 3\" -> ,\"1\" \"2\" \"3\" (\" \" as separator)\n           csv write    csv ,\"1\" \"2\" \"3\" -> \"1,2,3\\n\"\n                        \" \" csv ,\"1\" \"2\" \"3\" -> \"1 2 3\\n\"\nx in s     contained    \"bc\" \"ac\" in \"abcd\" -> 1 0\nx in y     member of    2 3 in 0 2 4 -> 1 0\nn mod n    modulus      3 mod 5 4 3 -> 2 1 0\nn nan n    fill NaNs    42 nan (1.5;sqrt -1) -> 1.5 42\ni rotate y rotate       2 rotate 1 2 3 -> 3 1 2       -2 rotate 1 2 3 -> 2 3 1\n\nsub[r;s]   regsub       sub[rx/[a-z]/;\"Z\"] \"aBc\" -> \"ZBZ\"\nsub[r;f]   regsub       sub[rx/[A-Z]/;_] \"aBc\" -> \"abc\"\nsub[s;s]   replace      sub[\"b\";\"B\"] \"abc\" -> \"aBc\"\nsub[s;s;i] replaceN     sub[\"a\";\"b\";2] \"aaa\" -> \"bba\" (stop after 2 times)\nsub[S]     replaceS     sub[\"b\" \"d\" \"c\" \"e\"] \"abc\" -> \"ade\"\nsub[S;S]   replaceS     sub[\"b\" \"c\";\"d\" \"e\"] \"abc\" -> \"ade\"\n\neval[s;loc;pfx]         like eval s, but provide name loc as location (usually\n                        a filename), and prefix pfx+\".\" for globals\n\nMATH: acos, asin, atan, cos, exp, log, round, sin, sqrt, tan, nan\nUTF-8: utf8.rcount (number of code points), utf8.valid\n"

const helpAdverbs = "ADVERBS HELP\nf'x    each      #'(4 5;6 7 8) -> 2 3\nx F'y  each      2 3#'1 2 -> (1 1;2 2 2)    {(x;y;z)}'[1;2 4;3] -> (1 2 3;1 4 3)\nF/x    fold      +/!10 -> 45\nF\\x    scan      +\\!10 -> 0 1 3 6 10 15 21 28 36 45\nx F/y  fold      1 2+/!10 -> 46 47                 {x+y-z}/[9;3 4;2 7] -> 7\nx F\\y  scan      5 6+\\1 2 3 -> (6 7;8 9;11 12)     {x+y-z}\\[9;3 4;2 7] -> 10 7\ni f/y  do        3{x*2}/4 -> 32\ni f\\y  dos       3{x*2}\\4 -> 4 8 16 32\nf f/y  while     {x<100}{x*2}/4 -> 128\nf f\\y  whiles    {x<100}{x*2}\\4 -> 4 8 16 32 64 128\nf/x    converge  {1+1.0%x}/1 -> 1.618033988749895     {-x}/1 -> -1\nf\\x    converges {_x%2}\\10 -> 10 5 2 1 0              {-x}\\1 -> 1 -1\ns/x    join      \",\"/\"a\" \"b\" \"c\" -> \"a,b,c\"\ns\\x    split     \",\"\\\"a,b,c\" -> \"a\" \"b\" \"c\"\nr\\x    split     rx/[,;]/\\\"a,b;c\" -> \"a\" \"b\" \"c\"\ni s\\y  splitN    (2) \",\"\\\"a,b,c\" -> \"a\" \"b,c\"\nI/x    encode    24 60 60/1 2 3 -> 3723  2/1 1 0 -> 6\nI\\x    decode    24 60 60\\3723 -> 1 2 3  2\\6 -> 1 1 0\n"

const helpIO = "IO/OS HELP\nclose h     flush any buffered data, then close filehandle h\nenv s       get environment variable s, or an error if unset\n            returns a dictionnary representing the whole environment for s~\"\"\nflush h     flush any buffered data for filehandle h\nimport s    read/eval wrapper roughly equivalent to eval[read s;s;s+\".\"]\nopen s      open path s for reading, returning a filehandle (h)\nprint s     print \"Hello, world!\\n\" (uses implicit $x for non-string values)\nread h      read from filehandle h until EOF or an error occurs\nread s      read file named s       lines:\"\\n\"\\read\"/path/to/file\"\nrun s       run command s or S      run \"pwd\"        run \"ls\" \"-l\"\n            inherits stdin, stdout, and stderr, returns true on success\nsay s       same as print, but appends a newline        say !5\nshell s     run command as-is through the shell         shell \"ls -l\"\n            inherits stderr, returns its own standard output or an error\n\nx env s     sets environment variable x to s, or returns an error.\nx env 0     unset environment variable x, or clear environment if x~\"\"\nx import s  read/eval wrapper roughly equivalent to eval[read s;s;x+\".\"]\nx open s    open path s with mode x in \"r\" \"r+\" \"w\" \"w+\" \"a\" \"a+\"\n            or pipe from (mode \"-|\") or to (mode \"|-\") command (s or S)\nx print s   print s to filehandle/name x        \"/path/to/file\" print \"content\"\nn read h    read n bytes from reader h or until EOF, or an error occurs\ns read h    read from reader h until 1-byte s, EOF, or an error occurs\nx run s     run command s or S with input string x as stdin\n            inherits stderr, returns its own standard output or an error\nx say s     same as print, but appends a newline\n\nARGS        command-line arguments, starting with script name\nSTDIN       standard input filehandle (buffered)\nSTDOUT      standard output filehandle (buffered)\nSTDERR      standard error filehandle (buffered)\n"

const helpTime = "TIME HELP\ntime cmd              time command with current time\ncmd time t            time command with time t\ntime[cmd;t;fmt]       time command with time t in given format\ntime[cmd;t;fmt;loc]   time command with time t in given format and location\n\nTime t should be either an integer representing unix epochtime, or a string\nin the given format (RFC3339 format layout \"2006-01-02T15:04:05Z07:00\" is the\ndefault). See https://pkg.go.dev/time for information on layouts and locations,\nas goal uses the same conventions as Go's time package. Supported values for\ncmd are as follows:\n\n    cmd (s)       result (type)\n    ------        -------------\n    \"day\"         day number (i)\n    \"date\"        year, month, day (I)\n    \"clock\"       hour, minute, second (I)\n    \"hour\"        0-23 hour (i)\n    \"minute\"      0-59 minute (i)\n    \"second\"      0-59 second (i)\n    \"unix\"        unix epoch time (i)\n    \"unixmilli\"   unix (millisecond version, only for current time) (i)\n    \"unixmicro\"   unix (microsecond version, only for current time) (i)\n    \"unixnano\"    unix (nanosecond version, only for current time) (i)\n    \"year\"        year (i)\n    \"yearday\"     1-365/6 year day (i)\n    \"week\"        year, week (I)\n    \"weekday\"     0-7 weekday (starts from Sunday) (i)\n    format (s)    format time using given layout (s)\n"

const helpGoal = "RUNTIME HELP\ngoal \"globals\"   return dictionnary with a copy of global variables\n\"prec\" goal i    set floating point formatting precision to i (default -1)\n\"seed\" goal i    set non-secure pseudo-rand seed to i (used by the ? verb)\n"

