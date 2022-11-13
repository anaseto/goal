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
	prog *Program

	// execution and stack handling
	stack     []V
	frameIdx  int32
	callDepth int32
	ipNext    int
	advanced  bool
	lambda    int

	// values
	globals        []V
	constants      []V
	variadics      []VariadicFun
	variadicsNames []string

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
	ctx.prog = &Program{}
	ctx.gIDs = map[string]int{}
	ctx.stack = make([]V, 0, 64)
	ctx.scanner = &Scanner{}
	ctx.parser = newParser(ctx)
	ctx.initVariadics()
	return ctx
}

// RegisterVariadic adds a variadic function to the context.
func (ctx *Context) RegisterVariadic(name string, vf VariadicFun) Variadic {
	id := len(ctx.variadics)
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	return Variadic(id)
}

// AssignGlobal assigns a value to a global variable name.
func (ctx *Context) AssignGlobal(name string, v V) {
	id := ctx.global(name)
	ctx.globals[id] = v
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
	blen, llen, last := len(ctx.prog.Body), len(ctx.prog.Lambdas), ctx.prog.last
	err := ctx.parser.Parse()
	if err != nil {
		ctx.prog.Body = ctx.prog.Body[:blen]
		ctx.prog.Lambdas = ctx.prog.Lambdas[:llen]
		ctx.prog.last = last
		return nil, fmt.Errorf("%v", err)
	}
	done, err := ctx.compileExec()
	if !done || err != nil {
		return nil, err
	}
	return ctx.top(), nil
}

// RunExpr compiles a whole expression from current source, then executes it.
// It returns ErrEOF if the end of input was reached without issues.
func (ctx *Context) RunExpr() (V, error) {
	if ctx.scanner.bReader == nil {
		return nil, errors.New("no source specified")
	}
	var eof bool
	blen, llen, last := len(ctx.prog.Body), len(ctx.prog.Lambdas), ctx.prog.last
	err := ctx.parser.ParseNext()
	if err != nil {
		_, eof = err.(ErrEOF)
		if !eof {
			ctx.prog.Body = ctx.prog.Body[:blen]
			ctx.prog.Lambdas = ctx.prog.Lambdas[:llen]
			ctx.prog.last = last
			return nil, fmt.Errorf("%v", err)
		}
	}
	done, err := ctx.compileExec()
	if !done || err != nil {
		return nil, err
	}
	if eof {
		err = ErrEOF{}
	}
	if ctx.advanced {
		return ctx.top(), err
	}
	return nil, err
}

// RunString calls Run with the given string as source.
func (ctx *Context) RunString(s string) (V, error) {
	ctx.SetSource("", strings.NewReader(s))
	return ctx.Run()
}

// LastIsAssign returns true if the last parsed expression was an assignment.
// This can be used by a repl to avoid printing results when assigning.
func (ctx *Context) LastIsAssign() bool {
	if len(ctx.prog.Body) == 0 {
		return false
	}
	switch ctx.prog.Body[ctx.prog.last] {
	case opAssignLocal, opAssignGlobal:
		return true
	default:
		return false
	}
}

func (ctx *Context) compileExec() (bool, error) {
	done := ctx.resolve()
	if !done {
		return false, nil
	}
	//fmt.Print(ctx.ProgramString())
	ip, err := ctx.execute(ctx.prog.Body[ctx.ipNext:])
	if err != nil {
		ctx.ipNext = len(ctx.prog.Body)
		ctx.stack = ctx.stack[0:]
		ctx.push(nil)
		return false, fmt.Errorf("%v", err)
	}
	ctx.ipNext += ip
	ctx.advanced = ip > 0
	if len(ctx.stack) == 0 {
		// should not happen
		return false, errors.New("no result: empty stack")
	}
	return true, nil
}

// Show prints internal information about the context.
func (ctx *Context) Show() {
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
