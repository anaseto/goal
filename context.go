package goal

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Context holds the state of the interpreter.
type Context struct {
	// program representations (AST and compiled)
	ast  *AstProgram
	prog *Program

	// execution and stack handling
	stack     []V
	frameIdx  int32
	callDepth int32
	ipNext    int

	// values
	globals   []V
	constants []V

	// symbol handling
	gNames []string
	gIDs   map[string]int

	// parsing, scanning
	scanner *Scanner
	parser  *parser
	fname   string
}

// NewContext returns a new context for compiling and interpreting code.
// SetSource should be called to set a source, and
func NewContext() *Context {
	ctx := &Context{}
	ctx.ast = &AstProgram{}
	ctx.prog = &Program{}
	ctx.gIDs = map[string]int{}
	ctx.stack = make([]V, 1, 64)
	ctx.scanner = &Scanner{}
	ctx.parser = newParser(ctx)
	return ctx
}

// SetSource sets the reader source for running code. The name is used for
// error reporting.
func (ctx *Context) SetSource(name string, r io.Reader) {
	ctx.fname = name
	ctx.scanner.Init(r)
}

// Run compiles the code from current source, then executes it.
func (ctx *Context) Run() (V, error) {
	if ctx.scanner.bReader == nil {
		return nil, errors.New("no source specified")
	}
	err := ctx.parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser:%v", err)
	}
	ctx.compile()
	ip, err := ctx.execute(ctx.prog.Body[ctx.ipNext:])
	ctx.ipNext += ip
	if err != nil {
		return nil, fmt.Errorf("parser:%v", err)
	}
	if len(ctx.stack) == 0 {
		// should not happen
		return nil, errors.New("no result: empty stack")
	}
	return ctx.top(), nil
}

// RunString calls Run with the given string as source.
func (ctx *Context) RunString(s string) (V, error) {
	ctx.SetSource("", strings.NewReader(s))
	return ctx.Run()
}

// Show prints internal information about the context.
func (ctx *Context) Show() {
	fmt.Printf("%s\n", ctx.ast)
	fmt.Printf("%s\n", ctx.ProgramString())
}

func (ctx *Context) storeConst(v V) int {
	ctx.constants = append(ctx.constants, v)
	return len(ctx.constants) - 1
}

func (ctx *Context) global(s string) int {
	id, ok := ctx.gIDs[s]
	if ok {
		return id
	}
	ctx.globals = append(ctx.globals, nil)
	ctx.gIDs[s] = len(ctx.gNames)
	ctx.gNames = append(ctx.gNames, s)
	return len(ctx.gNames) - 1
}
