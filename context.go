package main

// Context holds the state of the interpreter.
type Context struct {
	prog      *Program
	stack     []V
	sp        int
	globals   []V
	frame     []V
	callDepth int

	// from prog for direct access
	constants []V
}

func NewContext(prog *Program) *Context {
	ctx := &Context{prog: prog}
	max := 0
	for i := range prog.Globals {
		if i > max {
			max = i
		}
	}
	ctx.globals = make([]V, max+1)
	ctx.constants = prog.Constants
	return ctx
}
