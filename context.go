package goal

import (
	"errors"
	"fmt"
	"strings"
)

// Context holds the state of the interpreter.
type Context struct {
	// program representations (AST and compiled)
	ast  *AstProgram
	prog *Program

	// stack handling
	stack     []V
	frameIdx  int32
	callDepth int32

	// values
	globals   []V
	constants []V

	// symbol handling
	gNames []string
	gIDs   map[string]int

	// parsing, scanning
	scanner *Scanner
}

// NewContext returns a new context for compiling and interpreting code.
func NewContext() *Context {
	ctx := &Context{}
	ctx.ast = &AstProgram{}
	ctx.prog = &Program{}
	ctx.gIDs = map[string]int{}
	ctx.stack = make([]V, 1, 64)
	ctx.scanner = &Scanner{}
	return ctx
}

// RunString compiles the code in string s, then executes it.
func (ctx *Context) RunString(s string) (V, error) {
	sr := strings.NewReader(s)
	ctx.scanner.Init(sr)
	p := &parser{ctx: ctx}
	p.Init()
	err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser:%v", err)
	}
	ctx.compile()
	err = ctx.execute(ctx.prog.Body)
	if err != nil {
		return nil, fmt.Errorf("parser:%v", err)
	}
	if len(ctx.stack) == 0 {
		// should not happen
		return nil, errors.New("no result: empty stack")
	}
	return ctx.top(), nil
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
