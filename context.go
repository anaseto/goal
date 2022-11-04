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
