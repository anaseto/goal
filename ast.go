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

func (prog *AstProgram) String() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "---- AST -----")
	fmt.Fprintln(sb, "Instructions:")
	for _, expr := range prog.Body {
		fmt.Fprintf(sb, "\t%#v\n", expr)
	}
	fmt.Fprintln(sb, "Globals:")
	for name, id := range prog.Globals {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(sb, "Constants:")
	for id, v := range prog.Constants {
		fmt.Fprintf(sb, "\t%d\t%v\n", id, v)
	}
	for id, lc := range prog.Lambdas {
		fmt.Fprintf(sb, "---- Lambda %d -----\n", id)
		fmt.Fprintf(sb, "%s", lc)
	}
	return sb.String()
}

func (prog *AstProgram) storeConst(v V) int {
	prog.Constants = append(prog.Constants, v)
	return len(prog.Constants) - 1
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

func (lc *AstLambdaCode) String() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "Instructions:")
	for _, expr := range lc.Body {
		fmt.Fprintf(sb, "\t%#v\n", expr)
	}
	fmt.Fprintln(sb, "Locals:")
	for name, local := range lc.Locals {
		fmt.Fprintf(sb, "\t%s\tid:%d\ttype:%d\n", name, local.ID, local.Type)
	}
	return sb.String()
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
	ID  int // identifier
	Pos int // position information
}

// AstNil represents a constant.
type AstNil struct {
	Pos int // position information
}

// AstGlobal represents a global variable read.
type AstGlobal struct {
	Name string
	ID   int
	Pos  int
}

// AstLocal represents a local variable read.
type AstLocal struct {
	Name  string
	Local Local
	Pos   int
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

// AstVariadic represents built-in verbs with variable arity.
type AstVariadic struct {
	Variadic Variadic
	Pos      int
}

// AstLambda represents an user Lambda.
type AstLambda struct {
	Lambda Lambda
}

// AstApply applies the top stack value at the previous, dropping those values
// and pushing the result.
type AstApply struct{}

// AstApply2 applies the top stack value at the 2 previous ones, dropping those
// values and pushing the result.
type AstApply2 struct{}

// AstApplyN applies the top stack value at the N previous ones, dropping
// those values and pushing the result.
type AstApplyN struct {
	N int
}

// AstDrop represents a separator, in practice discarding the final
// value of the previous expression.
type AstDrop struct{}

func (n AstConst) node()        {}
func (n AstNil) node()          {}
func (n AstGlobal) node()       {}
func (n AstLocal) node()        {}
func (n AstAssignGlobal) node() {}
func (n AstAssignLocal) node()  {}
func (n AstVariadic) node()     {}
func (n AstLambda) node()       {}
func (n AstApply) node()        {}
func (n AstApply2) node()       {}
func (n AstApplyN) node()       {}
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

func (it *ppIter) Index() int {
	return it.i
}

func (it *ppIter) Set(i int) {
	it.i = i
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

type ppAdverbs []ppToken // for an adverb sequence

type ppParenExpr ppExprs // for parenthesized sub-expressions

type ppBlock struct {
	Type ppBlockType
	Body []ppExprs
	Args []string
}

func (ppb ppBlock) String() (s string) {
	switch ppb.Type {
	case ppLAMBDA:
		args := "[" + strings.Join([]string(ppb.Args), ";") + "]"
		s = fmt.Sprintf("{%s %v %v}", args, ppb.Type, ppb.Body)
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

func (tok ppToken) ppNode()     {}
func (pps ppStrand) ppNode()    {}
func (ppa ppAdverbs) ppNode()   {}
func (ppp ppParenExpr) ppNode() {}
func (ppb ppBlock) ppNode()     {}
