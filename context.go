package goal

import (
	"errors"
)

// Context holds the state of the interpreter.
type Context struct {
	// program representations (AST and compiled)
	gCode   *globalCode
	lambdas []*lambdaCode

	// execution and stack handling
	stack     []V
	frameIdx  int32
	callDepth int32
	lambda    int // currently executed lambda (if any)

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
	fname    string              // filename
	sources  map[string]string   // filename: source
	assigned bool                // last instruction was opAssignGlobal
	keywords map[string]NameType // special keyword names
	vNames   map[string]variadic // variadic keywords

	// error positions stack
	errPos []position
}

// NewContext returns a new context for compiling and interpreting code.
func NewContext() *Context {
	ctx := &Context{}
	ctx.gCode = &globalCode{}
	ctx.gIDs = map[string]int{}
	ctx.stack = make([]V, 0, 32)
	ctx.sources = map[string]string{}
	ctx.initVariadics()
	return ctx
}

// RegisterVariadic adds a variadic function to the context, and returns it, so
// that it can be assigned to a global variable.
func (ctx *Context) RegisterVariadic(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	return newVariadic(variadic(id))
}

// RegisterMonad adds a variadic function to the context, and generates a new
// monadic keyword for that variadic (parsing will not search for a left
// argument). The variadic is also returned.
// Note that while that a keyword defined in such a way will not take a left
// argument, it is still possible to pass several arguments to it with bracket
// indexing, like for any value.
func (ctx *Context) RegisterMonad(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	ctx.keywords[name] = NameMonad
	ctx.vNames[name] = variadic(id)
	return newVariadic(variadic(id))
}

// RegisterDyad adds a variadic function to the context, and generates a new
// dyadic keyword for that variadic (parsing will search for a left argument).
// The variadic is also returned.
func (ctx *Context) RegisterDyad(name string, vf VariadicFun) V {
	id := len(ctx.variadics)
	ctx.variadics = append(ctx.variadics, vf)
	ctx.variadicsNames = append(ctx.variadicsNames, name)
	ctx.keywords[name] = NameDyad
	ctx.vNames[name] = variadic(id)
	return newVariadic(variadic(id))
}

// AssignGlobal assigns a value to a global variable name.
func (ctx *Context) AssignGlobal(name string, x V) {
	id := ctx.global(name)
	ctx.globals[id] = x
}

// GetGlobal returns the value attached to a global variable with the given
// name.
func (ctx *Context) GetGlobal(name string) (V, bool) {
	id, ok := ctx.gIDs[name]
	if !ok {
		return V{}, false
	}
	return ctx.globals[id], true
}

// Compile parses and compiles code from the given source string. The name
// argument is used for error reporting and represents, usually, the filename.
func (ctx *Context) Compile(name string, s string) error {
	if len(ctx.gCode.Body) > 0 {
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
	}
	ctx.fname = name
	ctx.sources[name] = s
	ctx.scanner = NewScanner(ctx.keywords, s)
	ctx.compiler = newCompiler(ctx)
	llen := len(ctx.lambdas)
	err := ctx.compiler.ParseCompile()
	if err != nil {
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
		ctx.lambdas = ctx.lambdas[:llen]
		ctx.assigned = false
		return ctx.getError(err, true)
	}
	ctx.checkAssign()
	return nil
}

// Run runs compiled code, if not already done, and returns the result value.
func (ctx *Context) Run() (V, error) {
	if len(ctx.gCode.Body) == 0 {
		return V{}, nil
	}
	err := ctx.exec()
	if err != nil {
		return V{}, err
	}
	return ctx.pop(), nil
}

// Eval calls Compile with the given string as unnamed source, and then Run.
func (ctx *Context) Eval(s string) (V, error) {
	err := ctx.Compile("", s)
	if err != nil {
		return V{}, err
	}
	return ctx.Run()
}

// AssignedLast returns true if the last compiled expression was an assignment.
func (ctx *Context) AssignedLast() bool {
	return ctx.assigned
}

func (ctx *Context) checkAssign() {
	if len(ctx.gCode.Body) == 0 {
		ctx.assigned = false
	}
	switch ctx.gCode.Body[ctx.gCode.last] {
	case opAssignGlobal:
		ctx.assigned = true
	default:
		ctx.assigned = false
	}
}

func (ctx *Context) exec() error {
	ip, err := ctx.execute(ctx.gCode.Body)
	if err != nil {
		ctx.stack = ctx.stack[0:]
		ctx.push(V{})
		ctx.updateErrPos(ip, nil)
		ctx.gCode.Body = ctx.gCode.Body[:0]
		ctx.gCode.Pos = ctx.gCode.Pos[:0]
		ctx.gCode.last = 0
		return ctx.getError(err, false)
	}
	ctx.gCode.Body = ctx.gCode.Body[:0]
	ctx.gCode.Pos = ctx.gCode.Pos[:0]
	ctx.gCode.last = 0
	if len(ctx.stack) == 0 {
		// should not happen
		return ctx.getError(errors.New("no result: empty stack"), false)
	}
	return nil
}

func (ctx *Context) getError(err error, compile bool) error {
	e := &PanicError{
		Msg:       err.Error(),
		positions: ctx.errPos,
		sources:   ctx.sources,
		compile:   compile,
	}
	ctx.errPos = nil
	return e
}

func (ctx *Context) updateErrPos(ip int, lc *lambdaCode) {
	fname := ctx.fname
	if lc != nil {
		fname = lc.Filename
	}
	if len(ctx.gCode.Body) == 0 {
		// should not happen during execution
		ctx.errPos = append(ctx.errPos, position{Filename: fname})
		return
	}
	if lc != nil {
		if ip >= len(lc.Body) || ip < 0 {
			ip = len(lc.Body) - 1
		}
		pos := lc.Pos[ip]
		ctx.errPos = append(ctx.errPos, position{Filename: fname, Pos: pos, lambda: lc})
	} else {
		if ip >= len(ctx.gCode.Body) || ip < 0 {
			ip = len(ctx.gCode.Body) - 1
		}
		pos := ctx.gCode.Pos[ip]
		ctx.errPos = append(ctx.errPos, position{Filename: fname, Pos: pos})
	}
}

// Show returns a string representation with debug information about the
// context.
func (ctx *Context) Show() string {
	return ctx.programString()
}

func (ctx *Context) storeConst(x V) int {
	x.rcincr()
	ctx.constants = append(ctx.constants, x)
	return len(ctx.constants) - 1
}

func (ctx *Context) global(s string) int {
	id, ok := ctx.gIDs[s]
	if ok {
		return id
	}
	ctx.globals = append(ctx.globals, V{})
	ctx.gIDs[s] = len(ctx.gNames)
	ctx.gNames = append(ctx.gNames, s)
	return len(ctx.gNames) - 1
}

// derive returns a context derived from ctx, suitable for eval.
func (ctx *Context) derive() *Context {
	nctx := &Context{}
	nctx.gCode = &globalCode{}
	nctx.stack = make([]V, 0, 32)

	nctx.variadics = ctx.variadics
	nctx.variadicsNames = ctx.variadicsNames
	nctx.keywords = ctx.keywords
	nctx.vNames = ctx.vNames
	nctx.lambdas = ctx.lambdas
	nctx.globals = ctx.globals
	nctx.gNames = ctx.gNames
	nctx.gIDs = ctx.gIDs
	nctx.sources = ctx.sources
	nctx.errPos = ctx.errPos
	return nctx
}

// merge integrates changes from a context created with derive.
func (ctx *Context) merge(nctx *Context) {
	ctx.lambdas = nctx.lambdas
	ctx.globals = nctx.globals
	ctx.gNames = nctx.gNames
	ctx.gIDs = nctx.gIDs
	ctx.sources = nctx.sources
	ctx.errPos = nctx.errPos
}
