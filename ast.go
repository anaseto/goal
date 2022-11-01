package main

import "fmt"

// Program represents a program writter in goal.
type Program struct {
	Body      []Expr // program body AST
	Constants []V    // constants indexed by ID
	Globals   []V    // globals indexed by ID

	constants map[string]int // for generating symbols
	globals   map[string]int // for generating symbols
}

// Expr is used to represent the AST of the program.
type Expr interface {
	node()
}

// Exprs represent a list of stack-based AST expressions, to be evaluated from
// right to left.
type Exprs []Expr

// AstConst represents a constant.
type AstConst struct {
	ID  int
	Pos int
}

// AstGlobal represents a global variable read.
type AstGlobal struct {
	Name string
	ID   int
	Pos  int
}

// AstLocal represents a local variable read.
type AstLocal struct {
	Name string
	ID   int
	Pos  int
}

// AstAssignGlobal represents a global variable assignment.
type AstAssignGlobal struct {
	Name string
	ID   int
	Pos  int
}

// AstAssignLocal represents a local variable assignment.
type AstAssignLocal struct {
	Name string
	ID   int
	Pos  int
}

// AstMonad represents a monadic verb.
type AstMonad struct {
	Monad Monad
	Pos   int
}

// AstDyad represents a dyadic verb.
type AstDyad struct {
	Dyad Dyad
	Pos  int
}

// AstAdverb represents an adverb.
type AstAdverb struct {
	Adverb Adverb
	Pos    int
}

// AstApply represents a value that should be applied.
type AstApply struct {
	Value Expr
	Arity int
	Pos   int
}

// AstCond represents $[cond; then; else].
type AstCond struct {
	If   Expr
	Then Expr
	Else Expr
	Pos  int
}

// AstLambda represents an user defined lambda like {x+1}.
type AstLambda struct {
	Body    []Expr   // body AST
	Locals  []Symbol // vars, args
	Globals []Symbol
	Vars    int // number of vars from enclosing lambda
	Pos     int

	args   map[string]int // for generating symbols
	locals map[string]int // for generating symbols
}

// Symbol represents an identifier name along its ID.
type Symbol struct {
	Name string
	ID   int
}

func (n Exprs) node()           {}
func (n AstConst) node()        {}
func (n AstGlobal) node()       {}
func (n AstLocal) node()        {}
func (n AstAssignGlobal) node() {}
func (n AstAssignLocal) node()  {}
func (n AstCond) node()         {}
func (n AstMonad) node()        {}
func (n AstDyad) node()         {}
func (n AstAdverb) node()       {}
func (n AstApply) node()        {}
func (n AstLambda) node()       {}

// ppExpr is built by the first left to right pass, allowing to build a
// tree of blocks, delimit expressions, and simplify token information,
// but leaving the representation of the stack-like semantics of the
// language to a second IR builtin on type Expr.
type ppExpr interface {
	ppexpr()
}

type ppToken struct {
	Type ppTokenType
	Pos  int
	Text string
}

func (ppt ppToken) String() string {
	return fmt.Sprintf("{%v %d %s}", ppt.Type, ppt.Pos, ppt.Text)
}

type ppTokenType int

const (
	ppSEP ppTokenType = iota
	ppEOF
	ppCLOSE
	ppADVERB
	ppIDENT
	ppNUMBER
	ppSTRING
	ppVERB
)

type ppBlock struct {
	Type    ppBlockType
	ppexprs []ppExpr
}

func (ppb ppBlock) String() (s string) {
	switch ppb.Type {
	case ppBRACE:
		s = fmt.Sprintf("{%v %v}", ppb.Type, ppb.ppexprs)
	case ppBRACKET:
		s = fmt.Sprintf("[%v %v]", ppb.Type, ppb.ppexprs)
	case ppPAREN:
		s = fmt.Sprintf("(%v %v)", ppb.Type, ppb.ppexprs)
	}
	return s
}

type ppBlockType int

const (
	ppBRACE ppBlockType = iota
	ppBRACKET
	ppPAREN
)

type ppStrand []ppToken // for stranding, like 1 23 456

func (ppb ppBlock) ppexpr()  {}
func (pps ppStrand) ppexpr() {}
func (t ppToken) ppexpr()    {}
