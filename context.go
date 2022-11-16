package goal

import (
	"errors"
	"fmt"
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
	scanner  *Scanner
	compiler *compiler
	fname    string
	sources  map[string]string

	// error positions stack
	errPos []Position
}

// NewContext returns a new context for compiling and interpreting code.
// SetSource should be called to set a source, and
func NewContext() *Context {
	ctx := &Context{}
	ctx.prog = &Program{}
	ctx.gIDs = map[string]int{}
	ctx.stack = make([]V, 0, 64)
	ctx.scanner = &Scanner{}
	ctx.compiler = newCompiler(ctx)
	ctx.sources = map[string]string{}
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

// SetSource sets the reader string source for running code. The name is used
// for error reporting.
func (ctx *Context) SetSource(name string, s string) {
	ctx.fname = name
	ctx.sources[name] = s
	ctx.scanner.Init(s)
}

// Run compiles the code from current source, then executes it.
func (ctx *Context) Run() (V, error) {
	if ctx.scanner.reader == nil {
		panic("Run: no source specified with SetSource")
	}
	blen, llen, last := len(ctx.prog.Body), len(ctx.prog.Lambdas), ctx.prog.last
	err := ctx.compiler.ParseCompile()
	if err != nil {
		ctx.prog.Body = ctx.prog.Body[:blen]
		ctx.prog.Lambdas = ctx.prog.Lambdas[:llen]
		ctx.prog.last = last
		if ctx.errPos == nil {
			ctx.errPos = append(ctx.errPos,
				Position{Filename: ctx.fname, Pos: ctx.compiler.p.token.Pos})
		}
		return nil, ctx.getError(err)
	}
	if !ctx.changed(blen, llen, last) {
		return nil, nil
	}
	_, err = ctx.compileExec()
	if err != nil {
		return nil, err
	}
	return ctx.top(), nil
}

func (ctx *Context) changed(blen, llen, last int) bool {
	return blen != len(ctx.prog.Body) ||
		llen != len(ctx.prog.Lambdas) ||
		last != ctx.prog.last
}

// RunExpr compiles a whole expression from current source, then executes it.
// It returns ErrEOF if the end of input was reached without issues.
func (ctx *Context) RunExpr() (V, error) {
	if ctx.scanner.reader == nil {
		panic("RunExpr: no source specified with SetSource")
	}
	var eof bool
	blen, llen, last := len(ctx.prog.Body), len(ctx.prog.Lambdas), ctx.prog.last
	err := ctx.compiler.ParseCompileNext()
	if err != nil {
		_, eof = err.(ErrEOF)
		if !eof {
			ctx.prog.Body = ctx.prog.Body[:blen]
			ctx.prog.Lambdas = ctx.prog.Lambdas[:llen]
			ctx.prog.last = last
			if ctx.errPos == nil {
				ctx.errPos = append(ctx.errPos,
					Position{Filename: ctx.fname, Pos: ctx.compiler.p.token.Pos})
			}
			return nil, ctx.getError(err)
		}
	}
	if !ctx.changed(blen, llen, last) {
		return nil, nil
	}
	advanced, err := ctx.compileExec()
	if err != nil {
		return nil, err
	}
	if eof {
		err = ErrEOF{}
	}
	if advanced {
		return ctx.top(), err
	}
	return nil, err
}

// RunString calls Run with the given string as source.
func (ctx *Context) RunString(s string) (V, error) {
	ctx.fname = ""
	ctx.SetSource("", s)
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
	//fmt.Print(ctx.ProgramString())
	ip, err := ctx.execute(ctx.prog.Body[ctx.ipNext:])
	if err != nil {
		ctx.stack = ctx.stack[0:]
		ctx.push(nil)
		fmt.Printf("ip: %d\n", ip+ctx.ipNext)
		ctx.updateErrPos(ip+ctx.ipNext, nil)
		ctx.ipNext = len(ctx.prog.Body)
		return false, ctx.getError(err)
	}
	ctx.ipNext += ip
	if len(ctx.stack) == 0 {
		// should not happen
		return false, ctx.getError(errors.New("no result: empty stack"))
	}
	return ip > 0, nil
}

func (ctx *Context) getError(err error) error {
	e := &Error{
		Msg:       err.Error(),
		Positions: ctx.errPos,
		ctx:       ctx,
	}
	ctx.errPos = nil
	return e
}

func (ctx *Context) updateErrPos(ip int, lc *LambdaCode) {
	fname := ctx.fname
	if lc != nil {
		fname = lc.Filename
	}
	if len(ctx.prog.Body) == 0 {
		// should not happen during execution
		ctx.errPos = append(ctx.errPos, Position{Filename: fname})
		return
	}
	if lc != nil {
		if ip >= len(lc.Body) || ip < 0 {
			ip = len(lc.Body) - 1
		}
		pos := lc.Pos[ip]
		ctx.errPos = append(ctx.errPos, Position{Filename: fname, Pos: pos, Lambda: lc})
	} else {
		if ip >= len(ctx.prog.Body) || ip < 0 {
			ip = len(ctx.prog.Body) - 1
		}
		pos := ctx.prog.Pos[ip]
		ctx.errPos = append(ctx.errPos, Position{Filename: fname, Pos: pos})
	}
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
