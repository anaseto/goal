// Code generated by scripts/help.goal. DO NOT EDIT.

package main

const helpTopics = "TOPICS HELP\nType help TOPIC or h TOPIC where TOPIC is one of:\n\n\"s\"     syntax\n\"t\"     value types\n\"v\"     verbs (like +*-%,)\n\"nv\"    named verbs (like in, sign)\n\"a\"     adverbs ('/\\)\n\"io\"    IO verbs (like say, open, read)\n\"tm\"    time handling\n\"rt\"    runtime system\nop      where op is a builtin's name (like \"+\" or \"in\")\n\nNotations:\n        n (number) i (integer) s (string) r (regexp) d (dict)\n        f (function) F (dyadic function) e (error) h (handle)\n        x,y,z (any other) N,I,S,X,Y,A (arrays)\n"

const helpSyntax = "SYNTAX HELP\nnumbers         1     1.5     0b0110     1.7e-3     0xab     0n     0w     3h2m\nstrings         \"text\\xff\\u00F\\n\"   \"\\\"\"   \"\\u65e5\"   \"interpolated $var\"\n                qq/$var\\n or ${var}/   qq#text#  (delimiters :+-*%!&|=~,^#_?@/')\nraw strings     `anything until first backquote`     `literal \\, no escaping`\n                rq/anything until single slash/      rq#doubling ## escapes #\narrays          1 2 -3 4      1 \"ab\" -2 \"cd\"      (1 2;\"a\";3 \"b\";(4 2;\"c\");*)\nregexps         rx/[a-z]/      (see https://pkg.go.dev/regexp/syntax for syntax)\nverbs           : + - * % ! & | < > = ~ , ^ # _ $ ? @ . ::   (right-associative)\n                abs ceil error ...\nadverbs         / \\ ' (alone or after expr. with no space)    (left-associative)\nexpressions     2*3+4 -> 14      1+|2 3 4 -> 5 4 3      +/'(1 2 3;4 5 6) -> 6 15\nseparator       ; or newline except when ignored after {[( and before )]}\nvariables       a  b.c  f  data  t1\nassign          a:2 (local within lambda, global otherwise)        a::2 (global)\nop assign       a+:1 (sugar for a:a+1)         a-::2 (sugar for a::a-2)\nlist assign     (a;b;c):x (where 2<#x)         (a;b):1 2;b -> 2\neval. order     apply:f[e1;e2]   apply:e1 op e2                      (e2 before)\n                list:(e1;e2)     seq: [e1;e2]     lambda:{e1;e2}     (e1 before)\nsequence        [a:2;b:a+3;a+10] -> 12\nindex/apply     x[y] or x y is sugar for x@y; x[] ~ x[*] ~ x[!#x] ~ x (arrays)\nindex deep      x[y;z;...] is sugar for x.(y;z;...) (except for x in (?;and;or))\nindex assign    x[y]:z is sugar for x:@[x;y;:;z]           (or . for x[y;...]:z)\nindex op assign x[y]op:z is sugar for x:@[x;y;op;z]              (for symbol op)\nlambdas         {x+y-z}[3;5;7] -> 1       {[a;b;c]a+b-c}[3;5;7] -> 1\nprojections     +[2;] 3 -> 5              (2+) 3 -> 5      (partial application)\ncond            ?[1;2;3] -> 2     ?[0;2;3] -> 3    ?[0;2;\"\";3;4] -> 4\nand/or          and[1;2] -> 2   and[1;0;3] -> 0   or[0;2] -> 2   or[0;0;0] -> 0\nreturn          [1;:2;3] -> 2                       (a : at start of expression)\ntry             'x is sugar for ?[\"e\"~@x;:x;x]         (return if it's an error)\ncomments        from line with a single / until line with a single \\\n                or from / (after space or start of line) to end of line\n"

const helpTypes = "TYPES HELP\natom    array   name            examples\nn       N       number          0       1.5       !5       1.2 3 1.8\ns       S       string          \"abc\"   \"d\"       \"a\" \"b\" \"c\"\nr               regexp          rx/[a-z]/         rx/\\s+/\nd               dictionary      \"a\" \"b\"!1 2       keys!values\nf               function        +      {x*2}      (1-)      %[;2]\nh               handle          open \"/path/to/file\"    \"w\" open \"/path/to/file\"\ne               error           error \"msg\"\n        A       generic array   (\"a\" 1;\"b\" 2;\"c\" 3)     (+;-;*;\"any\")\n"

const helpVerbs = "VERBS HELP\n:x  identity    :[42] -> 42 (recall that : is also syntax for return and assign)\nx:y right       2:3 -> 3            \"a\":\"b\" -> \"b\"\n+d  swap k/v    +\"a\"\"b\"!0 1 -> 0 1!\"a\" \"b\"\n+x  flip        +(1 2;3 4) -> (1 3;2 4)                   +42 -> ,,42\nn+n add         2+3 -> 5            2+3 4 -> 5 6\ns+s concat      \"a\"+\"b\" -> \"ab\"     \"a\" \"b\"+\"c\" -> \"ac\" \"bc\"\n-n  negate      - 2 3 -> -2 -3      -(1 2.5;3 4) -> (-1 -2.5;-3 -4)\n-s  rtrim space -\"a\\tb \\r\\n\" \" c d \\n\" -> \"a\\tb\" \" c d\"  (Unicode's White Space)\nn-n subtract    5-3 -> 2            5 4-3 -> 2 1\ns-s trim suffix \"file.txt\"-\".txt\" -> \"file\"\n*x  first       *3 2 4 -> 3         *\"ab\" -> \"ab\"         *(+;*) -> +\nn*n multiply    2*3 -> 6            1 2 3*3 -> 3 6 9\ns*i repeat      \"a\"*3 2 1 0 -> \"aaa\" \"aa\" \"a\" \"\"\n%X  classify    %7 8 9 7 8 9 -> 0 1 2 0 1 2      %\"a\" \"b\" \"a\" -> 0 1 0\nx%y divide      3%2 -> 1.5          3 4%2 -> 1.5 2\n!i  enum        !5 -> 0 1 2 3 4\n!s  fields      !\"a b\\tc\\nd \u00a0e\" -> \"a\" \"b\" \"c\" \"d\" \"e\"   (Unicode's White Space)\n!I  odometer    !2 3 -> (0 0 0 1 1 1;0 1 2 0 1 2)\n!d  keys        !\"a\" \"b\"!1 2 -> \"a\" \"b\"\ni!i range       2!5 -> 2 3 4        5!2 -> !0\ni!s cut shape   3!\"abcdefghijk\" -> \"abc\" \"defg\" \"hijk\"\ni!Y cut shape   3!!6 -> (0 1;2 3;4 5)            -3!!6 -> (0 1 2;3 4 5)\nX!Y dict        d:\"a\" \"b\"!1 2;d \"a\" -> 1\n&s  byte-count  &\"abc\" -> 3     &\"π\" -> 2        &\"αβγ\" -> 6\n&I  where       &0 0 1 0 0 0 1 -> 2 6            &2 3 -> 0 0 1 1 1\n&d  keys where  &\"a\"\"b\"\"c\"\"d\"!0 1 1 0 -> \"b\" \"c\"\nx&y min         2&3 -> 2        4&3 -> 3         \"b\"&\"a\" -> \"a\"\n|X  reverse     |!5 -> 4 3 2 1 0\nx|y max         2|3 -> 3        4|3 -> 4         \"b\"|\"a\" -> \"b\"\n<d  sort up     <\"a\"\"b\"\"c\"!2 3 1 -> \"c\"\"a\"\"b\"!1 2 3\n<X  ascend      <3 5 4 -> 0 2 1          (index permutation for ascending order)\nx<y less        2<3 -> 1        \"c\" < \"a\" -> 0\n>d  sort down   >\"a\"\"b\"\"c\"!2 3 1 -> \"b\"\"a\"\"c\"!3 2 1\n>X  descend     >3 5 4 -> 1 2 0         (index permutation for descending order)\nx>y greater     2>3 -> 0        \"c\" > \"a\" -> 1\n=s  lines       =\"ab\\ncd\\r\\nef gh\" -> \"ab\" \"cd\" \"ef gh\"\n=I  index-count =1 0 0 2 2 3 -1 2 1 1 1 -> 2 4 3 1\n=d  group keys  =\"a\"\"b\"\"c\"!0 1 0 -> (\"a\" \"c\";,\"b\")         =\"a\"\"b\"!0 -1 -> ,,\"a\"\nf=Y group by    (2 mod)=!10 -> (0 2 4 6 8;1 3 5 7 9)\nx=y equal       2 3 4=3 -> 0 1 0        \"ab\" = \"ba\" -> 0\n~x  not         ~0 1 2 -> 1 0 0         ~\"a\" \"\" \"0\" -> 0 1 0\nx~y match       3~3 -> 1        2 3~3 2 -> 0             (\"a\";%)~'(\"b\";%) -> 0 1\n,x  enlist      ,1 -> ,1        #,2 3 -> 1               (list with one element)\nd,d merge       (\"a\"\"b\"!1 2),\"b\"\"c\"!3 4 -> \"a\"\"b\"\"c\"!1 3 4\nx,y join        1,2 -> 1 2              \"ab\" \"c\",\"d\" -> \"ab\" \"c\" \"d\"\n^d  sort keys   ^\"c\"\"a\"\"b\"!1 2 3 -> \"a\"\"b\"\"c\"!2 3 1\n^X  sort        ^3 5 0 -> 0 3 5         ^\"ca\" \"ab\" \"bc\" -> \"ab\" \"bc\" \"ca\"\ni^s windows     2^\"abcde\" -> \"abcd\" \"bcde\"\ni^Y windows     2^!4 -> (0 1 2;1 2 3)   -2^!4 -> (0 1;1 2;2 3)\ns^s trim        \" []\"^\"  [text]  \" -> \"text\"      \"\"^\" \\nstuff\\t\" -> \"stuff\"\nX^d w/o keys    (,\"b\")^\"a\"\"b\"\"c\"!0 1 2 -> \"a\"\"c\"!0 2\nX^Y without     2 3^1 1 2 3 3 4 -> 1 1 4\n#x  length      #2 4 5 -> 3       #\"ab\" \"cd\" -> 2       #42 -> 1      #\"ab\" -> 1\ni#y take/pad    2#6 7 8 -> 6 7    4#6 7 8 -> 6 7 8 0    -4#6 7 8 -> 0 6 7 8\ns#s count       \"ab\"#\"cabdab\" \"cd\" \"deab\" -> 2 0 1      \"\"#\"αβγ\" -> 4\nf#y replicate   {0 1 2 0}#4 1 5 3 -> 1 5 5        {x>0}#2 -3 1 -> 2 1\nX#d with keys   \"a\"\"c\"\"e\"#\"a\"\"b\"\"c\"!2 3 4 -> \"a\"\"c\"\"e\"!2 4 0\nX#Y with only   2 3#1 1 2 3 3 4 -> 2 3 3\n_n  floor       _2.3 -> 2               _1.5 3.7 -> 1 3\n_s  to lower    _\"ABC\" -> \"abc\"         _\"AB\" \"CD\" -> \"ab\" \"cd\"\ni_s drop bytes  2_\"abcde\" -> \"cde\"      -2_\"abcde\" -> \"abc\"\ni_Y drop        2_3 4 5 6 -> 5 6        -2_3 4 5 6 -> 3 4\ns_s trim prefix \"pref-\"_\"pref-name\" -> \"name\"\nf_y weed out    {0 1 1 0}_4 1 5 3 -> 4 3          {x>0}_2 -3 1 -> ,-3\nI_s cut string  1 3_\"abcdef\" -> \"bc\" \"def\"                         (I ascending)\nI_Y cut         2 5_!10 -> (2 3 4;5 6 7 8 9)                       (I ascending)\n$x  string      $2 3 -> \"2 3\"       $\"text\" -> \"\\\"text\\\"\"\ni$s pad         3$\"a\" -> \"a  \"      -3$\"1\" \"23\" \"456\" -> \"  1\" \" 23\" \"456\"\ns$y to integer  \"i\"$2.3 -> 2      \"i\"$\"aπ\" -> 97 960      \"b\"$\"aπ\" -> 97 207 128\ns$I to string   \"s\"$97 960 -> \"aπ\"\ns$s parse num   \"n\"$\"1.5\" -> 1.5    \"n\"$\"2\" \"1e+7\" \"0b100\" -> 2 1e+07 4\nX$y binsearch   2 3 5 7$8 2 7 5 5.5 3 0 -> 4 1 4 3 3 2 0           (x ascending)\n?i  uniform     ?2 -> 0.6046602879796196 0.9405090880450124    (between 0 and 1)\n?i  normal      ?-2 -> -1.233758177597947 -0.12634751070237293   (mean 0, dev 1)\n?X  uniq        ?2 2 3 4 3 3 -> 2 3 4\ni?i roll        5?100 -> 10 51 21 51 37\ni?Y roll array  5?\"a\" \"b\" \"c\" -> \"c\" \"a\" \"c\" \"c\" \"b\"\ni?i deal        -5?100 -> 19 26 0 73 94                        (always distinct)\ni?Y deal array  -3?\"a\"\"b\"\"c\" -> \"a\"\"c\"\"b\"                      (always distinct)\ns?r rindex      \"abcde\"?rx/b../ -> 1 3                           (offset;length)\ns?s index       \"a = a + 1\"?\"=\" \"+\" -> 2 6\nd?y find key    (\"a\" \"b\"!3 4)?4 -> \"b\"       (\"a\" \"b\"!3 4)?5 -> \"\"\nX?y find        9 8 7?8 -> 1                 9 8 7?6 -> 3\n@x  type        @2 -> \"n\"    @\"ab\" -> \"s\"    @2 3 -> \"N\"     @+ -> \"f\"\ni@y take/repeat 2@6 7 8 -> 6 7    4@6 7 8 -> 6 7 8 6 (cyclic)       3@1 -> 1 1 1\ns@i substr      \"abcdef\"@2  -> \"cdef\"                                (s[offset])\nr@s match       rx/^[a-z]+$/\"abc\" -> 1       rx/\\s/\"abc\" -> 0\nr@s find group  rx/([a-z])(.)/\"&a+c\" -> \"a+\" \"a\" \"+\"     (whole match, group(s))\nf@y apply       (|)@1 2 -> 2 1                      (like |[1 2] -> 2 1 or |1 2)\nd@y at key      (\"a\" \"b\"!1 2)@\"a\" -> 1\nX@i at          7 8 9@2 -> 9         7 8 9[2 0] -> 9 7       7 8 9@-2 -> 8\n.s  reval       .\"2+3\" -> 5    (restricted eval with new context: see also eval)\n.e  get error   .error \"msg\" -> \"msg\"\n.d  values      .\"a\"\"b\"!1 2 -> 1 2             (\"a\" \"b\"!1 2)[] -> 1 2 (special)\n.X  self-dict   .\"a\"\"b\" -> \"a\"\"b\"!\"a\"\"b\"       .!3 -> 0 1 2!0 1 2\ns.I substr      \"abcdef\"[2;3] -> \"cde\"                        (s[offset;length])\nr.y findN       rx/[a-z]/[\"abc\";2] -> \"a\"\"b\"    rx/[a-z]/[\"abc\";-1] -> \"a\"\"b\"\"c\"\nr.y findN group rx/[a-z](.)/[\"abcdef\";2] -> (\"ab\" \"b\";\"cd\" \"d\")\nf.y applyN      {x+y}.2 3 -> 5       {x+y}[2;3] -> 5\nX.y at deep     (6 7;8 9)[0;1] -> 7  (6 7;8 9)[;1] -> 7 9\n«X  shift       «8 9 -> 9 0    «\"a\" \"b\" -> \"b\" \"\"   (ASCII alternative: shift x)\nx«Y shift       \"a\" \"b\"«1 2 3 -> 3 \"a\" \"b\"\n»X  rshift      »8 9 -> 0 8    »\"a\" \"b\" -> \"\" \"a\"  (ASCII alternative: rshift x)\nx»Y rshift      \"a\" \"b\"»1 2 3 -> \"a\" \"b\" 1\n\n::x         get global  a:3;::\"a\" -> 3\nx::y        set global  \"a\"::3;a -> 3\n@[d;y;f]    amend       @[\"a\"\"b\"\"c\"!7 8 9;\"a\"\"b\"\"b\";10+] -> \"a\"\"b\"\"c\"!17 28 9\n@[X;i;f]    amend       @[7 8 9;0 1 1;10+] -> 17 28 9\n@[d;y;F;z]  amend       @[\"a\"\"b\"\"c\"!7 8 9;\"a\";:;42] -> \"a\"\"b\"\"c\"!42 8 9\n@[X;i;F;z]  amend       @[7 8 9;1 2 0;+;10 20 -10] -> -3 18 29\n@[f;x;f]    try         @[2+;3;{\"msg\"}] -> 5         @[2+;\"a\";{\"msg\"}] -> \"msg\"\n.[X;y;f]    deep amend  .[(6 7;8 9);0 1;-] -> (6 -7;8 9)\n.[X;y;F;z]  deep amend  .[(6 7;8 9);(0 1 0;1);+;10] -> (6 27;8 19)\n                        .[(6 7;8 9);(*;1);:;42] -> (6 42;8 42)\n.[f;x;f]    tryN        .[+;2 3;{\"msg\"}] -> 5        .[+;2 \"a\";{\"msg\"}] -> \"msg\"\n"

const helpNamedVerbs = "NAMED VERBS HELP\nabs n      abs value    abs -3 -1.5 2 -> 3 1.5 2\nceil x     ceil/upper   ceil 1.5 -> 2       ceil \"ab\" -> \"AB\"\ncsv s      csv read     csv \"1,2,3\" -> ,\"1\" \"2\" \"3\"\ncsv A      csv write    csv ,\"1\" \"2\" \"3\" -> \"1,2,3\\n\"\nerror x    error        r:error \"msg\"; (@r;.r) -> \"e\" \"msg\"\neval s     comp/run     a:5;eval \"a+2\" -> 7         (unrestricted variant of .s)\nfirsts X   mark firsts  firsts 0 0 2 3 0 2 3 4 -> 1 0 1 1 0 0 0 1\njson s     parse json   ^json `{\"a\":true,\"b\":\"text\"}` -> \"a\" \"b\"!(1;\"text\")\nnan n      isNaN        nan (0n;2;sqrt -1) -> 1 0 1\nocount X   occur-count  ocount 3 4 5 3 4 4 7 -> 0 0 0 1 1 2 0\npanic s    panic        panic \"msg\"               (for fatal programming-errors)\nrx s       comp. regex  rx \"[a-z]\"      (like rx/[a-z]/ but compiled at runtime)\nsign n     sign         sign -3 -1 0 1.5 5 -> -1 -1 0 1 1\n\ns csv s    csv read     \" \" csv \"1 2 3\" -> ,\"1\" \"2\" \"3\"       (\" \" as separator)\ns csv A    csv write    \" \" csv ,\"1\" \"2\" \"3\" -> \"1 2 3\\n\"     (\" \" as separator)\nx in s     contained    \"bc\" \"ac\" in \"abcd\" -> 1 0\nx in Y     member of    2 3 in 0 2 4 -> 1 0\nn mod n    modulus      3 mod 5 4 3 -> 2 1 0\nn nan n    fill NaNs    42 nan (1.5;sqrt -1) -> 1.5 42\ni rotate Y rotate       2 rotate 7 8 9 -> 9 7 8         -2 rotate 7 8 9 -> 8 9 7\n\nsub[r;s]   regsub       sub[rx/[a-z]/;\"Z\"] \"aBc\" -> \"ZBZ\"\nsub[r;f]   regsub       sub[rx/[A-Z]/;_] \"aBc\" -> \"abc\"\nsub[s;s]   replace      sub[\"b\";\"B\"] \"abc\" -> \"aBc\"\nsub[s;s;i] replaceN     sub[\"a\";\"b\";2] \"aaa\" -> \"bba\"       (stop after 2 times)\nsub[S]     replaceS     sub[\"b\" \"d\" \"c\" \"e\"] \"abc\" -> \"ade\"\nsub[S;S]   replaceS     sub[\"b\" \"c\";\"d\" \"e\"] \"abc\" -> \"ade\"\n\neval[s;loc;pfx]         like eval s, but provide loc as location (usually a\n                        path), and prefix pfx+\".\" for globals\n\nutf8 s     is UTF-8     utf8 \"aπc\" -> 1                        utf8 \"a\\xff\" -> 0\ns utf8 s   to UTF-8     \"b\" utf8 \"a\\xff\" -> \"ab\"      (replace invalid with \"b\")\n\nMATH: atan2[n;n]; cos n; exp n; log n; round n; sin n; sqrt n\n"

const helpAdverbs = "ADVERBS HELP\nf'x    each      #'(4 5;6 7 8) -> 2 3\nx F'y  each      2 3#'4 5 -> (4 4;5 5 5)    {(x;y;z)}'[1;2 3;4] -> (1 2 4;1 3 4)\nF/x    fold      +/!10 -> 45\nF\\x    scan      +\\!10 -> 0 1 3 6 10 15 21 28 36 45\nx F/y  fold      5 6+/!4 -> 11 12                    {x+y-z}/[5;4 3;2 1] -> 9\nx F\\y  scan      5 6+\\!4 -> (5 6;6 7;8 9;11 12)      {x+y-z}\\[5;4 3;2 1] -> 7 9\ni f/y  do        3{x*2}/4 -> 32\ni f\\y  dos       3{x*2}\\4 -> 4 8 16 32\nf f/y  while     {x<100}{x*2}/4 -> 128\nf f\\y  whiles    {x<100}{x*2}\\4 -> 4 8 16 32 64 128\nf/x    converge  {1+1.0%x}/1 -> 1.618033988749895    {-x}/1 -> -1\nf\\x    converges {_x%2}\\10 -> 10 5 2 1 0             {-x}\\1 -> 1 -1\ns/S    join      \",\"/\"a\" \"b\" \"c\" -> \"a,b,c\"\ns\\s    split     \",\"\\\"a,b,c\" -> \"a\" \"b\" \"c\"          \"\"\\\"aπc\" -> \"a\" \"π\" \"c\"\nr\\s    split     rx/[,;]/\\\"a,b;c\" -> \"a\" \"b\" \"c\"\ni s\\s  splitN    (2) \",\"\\\"a,b,c\" -> \"a\" \"b,c\"\nI/x    encode    24 60 60/1 2 3 -> 3723              2/1 1 0 -> 6\nI\\x    decode    24 60 60\\3723 -> 1 2 3              2\\6 -> 1 1 0\n"

const helpIO = "IO/OS HELP\nchdir s     change current working directory to s, or return an error\nclose h     flush any buffered data, then close filehandle h\nenv s       get environment variable s, or an error if unset\n            return a dictionary representing the whole environment if s~\"\"\nflush h     flush any buffered data for filehandle h\nimport s    read/eval wrapper roughly equivalent to eval[read path;path;pfx]\n            where 1) path~s or is derived from s by appending \".goal\" and/or\n                     prefixing with env \"GOALLIB\"\n                  2) pfx is path's basename without extension\nopen s      open path s for reading, returning a filehandle (h)\nprint s     print \"Hello, world!\\n\"     (uses implicit $x for non-string values)\nread h      read from filehandle h until EOF or an error occurs\nread s      read file named s                     lines:\"\\n\"\\read\"/path/to/file\"\nrun s       run command s or S (with arguments)   run \"pwd\"        run \"ls\" \"-l\"\n            inherits stdin and stderr, returns its standard output or an error\n            dict with keys \"code\" \"msg\" \"out\"\nsay s       same as print, but appends a newline                   say !5\nshell s     same as s run \"/bin/sh\"                                shell \"ls -l\"\n\nx env s     set environment variable x to s, or return an error\nx env 0     unset environment variable x, or clear environment if x~\"\"\nx import s  same as import s, but using prefix x for globals\nx open s    open path s with mode x in \"r\" \"r+\" \"w\" \"w+\" \"a\" \"a+\"\n            or pipe from (x~\"-|\") or to (x~\"|-\") command s or S\nx print s   print s to filehandle/name x        \"/path/to/file\" print \"content\"\ni read h    read i bytes from reader h or until EOF, or an error occurs\ns read h    read from reader h until 1-byte s, EOF, or an error occurs\nx run s     same as run s but with input string x as stdin\nx say s     same as print, but appends a newline\n\nARGS        command-line arguments, starting with script name\nSTDIN       standard input filehandle (buffered)\nSTDOUT      standard output filehandle (buffered)\nSTDERR      standard error filehandle (buffered)\n"

const helpTime = "TIME HELP\ntime cmd              time command with current time\ncmd time t            time command with time t\ntime[cmd;t;fmt]       time command with time t in given format\ntime[cmd;t;fmt;loc]   time command with time t in given format and location\n\nTime t should be either an integer representing unix epochtime, or a string in\nthe given format (RFC3339 format layout \"2006-01-02T15:04:05Z07:00\" is the\ndefault). See https://pkg.go.dev/time for information on layouts and locations,\nas goal uses the same conventions as Go's time package. Supported values for\ncmd are as follows:\n\n    cmd (s)       result (type)\n    ------        -------------\n    \"clock\"       hour, minute, second (I)\n    \"date\"        year, month, day (I)\n    \"day\"         day number (i)\n    \"hour\"        0-23 hour (i)\n    \"minute\"      0-59 minute (i)\n    \"second\"      0-59 second (i)\n    \"unix\"        unix epoch time (i)\n    \"unixmicro\"   unix (microsecond version) (i)\n    \"unixmilli\"   unix (millisecond version) (i)\n    \"unixnano\"    unix (nanosecond version) (i)\n    \"week\"        year, week (I)\n    \"weekday\"     0-7 weekday starting from Sunday (i)\n    \"year\"        year (i)\n    \"yearday\"     1-365/6 year day (i)\n    \"zone\"        name, offset in seconds east of UTC (s;i)\n    format (s)    format time using given layout (s)\n"

const helpRuntime = "RUNTIME HELP\nrt.ofs s        set output field separator for print S and \"$S\"    (default \" \")\n                returns previous value\nrt.prec i       set floating point formatting precision to i        (default -1)\n                returns previous value\nrt.seed i       set non-secure pseudo-rand seed to i        (used by the ? verb)\nrt.time[s;i]    eval s for i times (default 1), return average time (ns)\nrt.time[f;x;i]  call f.x for i times (default 1), return average time (ns)\nrt.vars s       return dictionary with a copy of global variables\n                s~\"\" for all variables, \"f\" functions, \"v\" non-functions\n"

