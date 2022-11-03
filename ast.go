package main

import (
	"fmt"
	"strings"
)

// AstProgram represents a program written in goal.
type AstProgram struct {
	Body      []Expr           // program body AST
	Constants []V              // constants indexed by ID
	Globals   map[string]int   // name: ID
	Lambdas   []*AstLambdaCode // list of user defined lambdas

	nglobals int // number of globals
}

func (prog *AstProgram) storeConst(v V) int {
	prog.Constants = append(prog.Constants, v)
	return len(prog.Constants) - 1
}

func (prog *AstProgram) pushExpr(e Expr) {
	prog.Body = append(prog.Body, e)
}

func (prog *AstProgram) global(s string) int {
	id, ok := prog.Globals[s]
	if ok {
		return id
	}
	prog.Globals[s] = prog.nglobals
	prog.nglobals++
	return prog.nglobals - 1
}

// AstLambdaCode represents an user defined lambda like {x+1}.
type AstLambdaCode struct {
	Body      []Expr           // body AST
	Locals    map[string]Local // arguments and variables
	NamedArgs bool             // named arguments instead of x, y, z
	Pos       int

	nVars int // number of non-argument local variables
}

// Local represents either an argument or a local variable. IDs are
// unique for a given type only.
type Local struct {
	Type LocalType
	ID   int
}

// LocalType represents different kinds of locals.
type LocalType int

// These constants describe the supported kinds of locals.
const (
	LocalArg LocalType = iota
	LocalVar
)

func (l *AstLambdaCode) local(s string) (Local, bool) {
	param, ok := l.Locals[s]
	if ok {
		return param, true
	}
	if !l.NamedArgs && len(s) == 1 {
		switch r := rune(s[0]); r {
		case 'x', 'y', 'z':
			id := r - 'x'
			arg := Local{Type: LocalArg, ID: int(id)}
			l.Locals[s] = arg
			return arg, true
		}
	}
	return Local{}, false
}

// Expr is used to represent the AST of the program with stack-like
// semantics.
type Expr interface {
	node() // AST node
}

// AstConst represents a constant.
type AstConst struct {
	ID   int // identifier
	Pos  int // position information
	Argc int // argument count: 0 (push), >0 (apply)
}

// AstGlobal represents a global variable read.
type AstGlobal struct {
	Name string
	ID   int
	Pos  int
	Argc int
}

// AstLocal represents a local variable read.
type AstLocal struct {
	Name  string
	Local Local
	Pos   int
	Argc  int
}

// AstAssignGlobal represents a global variable assignment.
type AstAssignGlobal struct {
	Name string
	ID   int
	Pos  int
}

// AstAssignLocal represents a local variable assignment.
type AstAssignLocal struct {
	Name  string
	Local Local
	Pos   int
}

// AstMonad represents a monadic verb.
type AstMonad struct {
	Monad Monad
	Pos   int
	Argc  int
}

// AstDyad represents a dyadic verb.
type AstDyad struct {
	Dyad Dyad
	Pos  int
	Argc int
}

// AstAdverb represents an adverb.
type AstAdverb struct {
	Adverb Adverb
	Pos    int
	Argc   int
}

// AstLambda represents an user Lambda.
type AstLambda struct {
	Lambda Lambda
	Pos    int
	Argc   int
}

// AstCond represents $[cond; then; else].
type AstCond struct {
	If   []Expr
	Then []Expr
	Else []Expr
	Pos  int
	Argc int
}

// AstDrop represents a separator, in practice discarding the final
// value of the previous expression.
type AstDrop struct{}

func (n AstConst) node()        {}
func (n AstGlobal) node()       {}
func (n AstLocal) node()        {}
func (n AstAssignGlobal) node() {}
func (n AstAssignLocal) node()  {}
func (n AstCond) node()         {}
func (n AstMonad) node()        {}
func (n AstDyad) node()         {}
func (n AstAdverb) node()       {}
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

// ppExprs is used to represent a whole expression. It is not a ppExpr.
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

type ppParenExpr ppExprs // for parenthesized sub-expressions

type ppBlock struct {
	Type ppBlockType
	Body []ppExprs
}

func (ppb ppBlock) String() (s string) {
	switch ppb.Type {
	case ppLAMBDA:
		s = fmt.Sprintf("{%v %v}", ppb.Type, ppb.Body)
	case ppARGS:
		s = fmt.Sprintf("[%v %v]", ppb.Type, ppb.Body)
	case ppLIST:
		s = fmt.Sprintf("(%v %v)", ppb.Type, ppb.Body)
	}
	return s
}

func (ppb ppBlock) push(ppe ppExpr) {
	ppb.Body[len(ppb.Body)-1] = append(ppb.Body[len(ppb.Body)-1], ppe)
}

type ppBlockType int

const (
	ppLAMBDA ppBlockType = iota
	ppARGS
	ppSEQ
	ppLIST
)

type ppArgs []string

func (ppa ppArgs) String() (s string) {
	return "[ARGS: " + strings.Join([]string(ppa), ";") + "]"
}

func (tok ppToken) ppNode()     {}
func (pps ppStrand) ppNode()    {}
func (ppp ppParenExpr) ppNode() {}
func (ppb ppBlock) ppNode()     {}
func (ppa ppArgs) ppNode()      {}
