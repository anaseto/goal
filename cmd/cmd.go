package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"goal"
	"log"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"strings"
)

type Config struct {
	Help        map[string]string
	ProgramName string
	Man         string
}

// Cmd runs a goal interpreter with starting context ctx and the given help
// strings when using the repl. Command line usage is then as follows:
//
//	program-name [-e command] [-d] [path]
func Cmd(ctx *goal.Context, cfg Config) {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	optE := flag.String("e", "", "execute command")
	optD := flag.Bool("d", false, "debug info")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-e command] [path]\n", os.Args[0])
		flag.PrintDefaults()
		if cfg.Man != "" {
			fmt.Fprintf(os.Stderr, "See man page %s(1) for details (TODO).\n", cfg.Man)
		}
	}
	flag.Parse()
	if *cpuprofile != "" {
		// profiling
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v", cfg.ProgramName, err)
			os.Exit(1)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	args := flag.Args()
	if *optD {
		defer runDebug(ctx, cfg)
	}
	if *optE != "" {
		runCommand(ctx, *optE, cfg.ProgramName)
	}
	if *optE == "" && len(args) == 0 || len(args) == 1 && args[0] == "-" {
		runStdin(ctx, cfg)
		return
	}
	if len(args) == 0 {
		return
	}
	fname := args[0]
	ctx.AssignGlobal("ARGS", goal.NewAS(args[1:]))
	bs, err := os.ReadFile(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", cfg.ProgramName, err)
		os.Exit(1)
	}
	source := string(bs)
	if len(source) > 2 && source[:2] == "#!" {
		// skip shellbang #! line
		i := strings.IndexByte(source, '\n')
		if i > 0 {
			source = source[i+1:]
		} else {
			source = ""
		}
	}
	err = ctx.Compile(fname, source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", cfg.ProgramName, err)
		if *optD {
			printProgram(ctx, cfg)
		}
		os.Exit(1)
	}
	if *optD {
		printProgram(ctx, cfg)
		os.Exit(0)
	}
	r, err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", cfg.ProgramName, err)
		os.Exit(1)
	}
	if r.IsError() {
		fmt.Fprint(os.Stderr, r.Error())
		os.Exit(1)
	}
}

func runStdin(ctx *goal.Context, cfg Config) {
	help := cfg.Help
	helpLast := false
	helpv := ctx.RegisterMonad("help", func(ctx *goal.Context, args []goal.V) goal.V {
		if len(args) >= 1 {
			arg, ok := args[0].Value().(goal.S)
			if !ok {
				return goal.Panicf("help x : x not a string (%s)", args[0].Type())
			}
			fmt.Println(strings.TrimSpace(help[string(arg)]))
		}
		helpLast = true
		return goal.NewI(1)
	})
	// We define an alias for help as a global to allow redefinition.
	ctx.AssignGlobal("h", helpv)
	lr := lineReader{r: bufio.NewReader(os.Stdin)}
	fmt.Printf("%s repl, type help\"\" for basic info.\n", cfg.ProgramName)
	for {
		fmt.Print("  ")
		s, err := lr.readLine()
		if err != nil && s == "" {
			return
		}
		r, err := ctx.Eval(s)
		if err != nil {
			fmt.Println("'ERROR " + strings.TrimSuffix(err.Error(), "\n"))
			continue
		}
		assigned := ctx.AssignedLast()
		if !assigned && !helpLast {
			echo(ctx, r)
		}
		helpLast = false
	}
}

type lineReader struct {
	r *bufio.Reader
}

type scanner struct {
	depth  []rune // (){}[] depth stack
	state  scanState
	done   bool
	escape bool
}

type scanState int

const (
	scanNormal scanState = iota
	scanRawString
	scanString
	scanRegexp
)

func (lr lineReader) readLine() (string, error) {
	s := scanner{}
	sb := strings.Builder{}
	for {
		r, _, err := lr.r.ReadRune()
		if err != nil {
			return sb.String(), err
		}
		switch r {
		case '\r':
			continue
		default:
			sb.WriteRune(r)
		}
		switch s.state {
		case scanNormal:
			switch r {
			case '\n':
				if len(s.depth) == 0 || s.done {
					return sb.String(), nil
				}
			case '"':
				s.state = scanString
			case '`':
				s.state = scanRawString
			case '/':
				if strings.HasSuffix(sb.String(), "rx/") {
					s.state = scanRegexp
				}
			case '{', '(', '[':
				s.depth = append(s.depth, r)
			case '}', ')', ']':
				if len(s.depth) > 0 && s.depth[len(s.depth)-1] == opening(r) {
					s.depth = s.depth[:len(s.depth)-1]
				} else {
					// error, so return on next \n
					s.done = true
				}
			}
		case scanRegexp:
			switch r {
			case '\n':
				// non terminated string
				return sb.String(), nil
			case '\\':
				s.escape = !s.escape
			case '/':
				if !s.escape {
					s.state = scanNormal
				}
				s.escape = false
			default:
				s.escape = false
			}
		case scanString:
			switch r {
			case '\n':
				// non terminated regexp
				return sb.String(), nil
			case '\\':
				s.escape = !s.escape
			case '"':
				if !s.escape {
					s.state = scanNormal
				}
				s.escape = false
			default:
				s.escape = false
			}
		case scanRawString:
			if r == '`' {
				s.state = scanNormal
			}
		}
	}
}

func opening(r rune) rune {
	switch r {
	case ')':
		return '('
	case ']':
		return '['
	case '}':
		return '{'
	default:
		return r
	}
}

func printProgram(ctx *goal.Context, cfg Config) {
	fmt.Fprintf(os.Stderr, "%s: debug info below:\n%v", cfg.ProgramName, ctx.Show())
}

func runDebug(ctx *goal.Context, cfg Config) {
	if r := recover(); r != nil {
		printProgram(ctx, cfg)
		log.Printf("Caught panic: %v\nStack Trace:\n", r)
		debug.PrintStack()
	}
}

func runCommand(ctx *goal.Context, cmd string, name string) {
	r, err := ctx.Eval(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", name, err)
		os.Exit(1)
	}
	if r.IsError() {
		fmt.Fprint(os.Stderr, r.Error())
		os.Exit(1)
	}
}

func echo(ctx *goal.Context, x goal.V) {
	if x != (goal.V{}) {
		fmt.Printf("%s\n", x.Sprint(ctx))
	}
}
