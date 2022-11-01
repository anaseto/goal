package main

// Context holds the state of the interpreter.
type Context struct {
	prog *Program
}

type Program struct {
}

// LambdaCode represents a compiled user defined function.
type LambdaCode struct {
	Arity  int
	Body   []opcode
	Locals []int
}
