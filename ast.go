package main

import "fmt"

// AstProgram represents a program written in goal.
type AstProgram struct {
	Body      []Expr          // program body AST
	Constants []V             // constants indexed by ID
	Globals   []V             // globals indexed by ID
	Lambdas   []AstLambdaCode // list of user defined lambdas

	constants map[string]int // for generating symbols
	globals   map[string]int // for generating symbols
}

func (prog *AstProgram) storeConst(v V) int {
	prog.Constants = append(prog.Constants, v)
	return len(prog.Constants) - 1
}

func (prog *AstProgram) pushExpr(e Expr) {
	prog.Body = append(prog.Body, e)
}

// AstLambdaCode represents an user defined lambda like {x+1}.
type AstLambdaCode struct {
	Body   []Expr // body AST
	Args   []Symbol
	Locals []Symbol
	Pos    int

	args   map[string]int // for generating symbols
	locals map[string]int // for generating symbols
}

// Expr is used to represent the AST of the program with stack-like
// semantics.
type Expr interface {
	node()
}

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

// AstLambda represents an user Lambda.
type AstLambda struct {
	Lambda Lambda
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

// AstDrop represents a separator, in practice discarding the final
// value of the previous expression.
type AstDrop struct{}

// Symbol represents an identifier name along its ID.
type Symbol struct {
	Name string
	ID   int
}

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
func (n AstDrop) node()         {}

// ppExpr are built by the first left to right pass, resulting in a tree
// of blocks producing a whole expression, with simplified token
// information, and stack-like order. It is a pre-processing IR. The
// representation of specific semantics of the language is left to a
// second IR built around type Expr.
type ppExpr interface {
	ppNode()
}

// ppExprs is used to represent a whole expression or a parenthesized
// sub-expression.
type ppExprs []ppExpr

// ppIter is an iterator for ppExprs slices, with peek functionality.
type ppIter struct {
	pps ppExprs
	i   int
}

func newppIter(pps ppExprs) ppIter {
	return ppIter{pps: pps, i: -1}
}

func (it *ppIter) Next() bool {
	it.i++
	return it.i < len(it.pps)
}

func (it *ppIter) Expr() ppExpr {
	return it.pps[it.i]
}

func (it *ppIter) Peek() ppExpr {
	if it.i+1 >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+1]
}

func (it *ppIter) PeekN(n int) ppExpr {
	if it.i+n >= len(it.pps) {
		return nil
	}
	return it.pps[it.i+n]
}

// ppToken represents a simplified token after processing into ppExpr.
type ppToken struct {
	Type ppTokenType
	Pos  int
	Text string
}

func (ppt ppToken) String() string {
	return fmt.Sprintf("{%v %d %s}", ppt.Type, ppt.Pos, ppt.Text)
}

// ppTokenType represents tokens in a ppExpr.
type ppTokenType int

// These constants represent the possible tokens in a ppExpr. The SEP,
// EOF and CLOSE types are not emitted in the final result.
const (
	ppSEP ppTokenType = iota
	ppEOF
	ppCLOSE

	ppNUMBER
	ppSTRING
	ppIDENT
	ppVERB
	ppADVERB
)

type ppStrand []ppToken // for stranding, like 1 23 456

type ppBlock struct {
	Type    ppBlockType
	ppexprs []ppExprs
}

func (ppb ppBlock) String() (s string) {
	switch ppb.Type {
	case ppLAMBDA:
		s = fmt.Sprintf("{%v %v}", ppb.Type, ppb.ppexprs)
	case ppARGS:
		s = fmt.Sprintf("[%v %v]", ppb.Type, ppb.ppexprs)
	case ppLIST:
		s = fmt.Sprintf("(%v %v)", ppb.Type, ppb.ppexprs)
	}
	return s
}

func (ppb ppBlock) push(ppe ppExpr) {
	ppb.ppexprs[len(ppb.ppexprs)-1] = append(ppb.ppexprs[len(ppb.ppexprs)-1], ppe)
}

type ppBlockType int

const (
	ppLAMBDA ppBlockType = iota
	ppARGS
	ppLIST
)

func (tok ppToken) ppNode()  {}
func (pps ppStrand) ppNode() {}
func (pps ppExprs) ppNode()  {}
func (ppb ppBlock) ppNode()  {}
