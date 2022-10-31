package main

import "fmt"

// Program represents a program writter in goal.
type Program struct {
	Body      []Expr // program body ast
	Constants []V    // constants indexed by ID
	Globals   []V    // globals indexed by ID

	constants map[string]int // for generating symbols
	globals   map[string]int // for generating symbols
}

// Expr is used to represent the ast of the program.
type Expr interface {
	node()
}

// Exprs represent a list of stack-based ast expressions, to be evaluated from
// right to left.
type Exprs []Expr

// AstConst represents a constant.
type AstConst struct {
	ID  int
	Pos int
}

type AstGlobal struct {
	Name string
	ID   int
	Pos  int
}

type AstLocal struct {
	Name string
	ID   int
	Pos  int
}

// AstAssignGlobal represents a global variable assignment. A global variable
// can only be assigned once, that is, they are immutable.
type AstAssignGlobal struct {
	Name string
	ID   int
	Pos  int
}

type AstAssignLocal struct {
	Name string
	ID   int
	Pos  int
}

type AstCond struct {
	If   Expr
	Then Expr
	Else Expr
	Pos  int
}

type AstMonad Monad
type AstDyad Dyad
type AstAdverb Adverb

type AstApply struct {
	Value Expr
	Arity int
	Pos   int
}

type AstLambda struct {
	Body    []Expr   // body ast
	Locals  []Symbol // vars, args
	Globals []Symbol
	Vars    int // number of vars from enclosing lambda

	args   map[string]int // for generating symbols
	locals map[string]int // for generating symbols
}

type Symbol struct {
	ID   int
	Name string
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

// ppExpr represents a pre-ast that builds blocks and forms nouns without
// giving meaning yet.
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

var ppTokenStrings = [...]string{
	ppSEP:    "ppSEP",
	ppEOF:    "ppEOF",
	ppCLOSE:  "ppCLOSE",
	ppADVERB: "ppADVERB",
	ppIDENT:  "ppIDENT",
	ppNUMBER: "ppNUMBER",
	ppSTRING: "ppSTRING",
	ppVERB:   "ppVERB",
}

func (pptt ppTokenType) String() string {
	return ppTokenStrings[pptt]
}

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

var ppBlockStrings = [...]string{
	ppBRACE:   "ppBRACE",
	ppBRACKET: "ppBRACKET",
	ppPAREN:   "ppPAREN",
}

func (pptt ppBlockType) String() string {
	return ppBlockStrings[pptt]
}

type ppStrand []ppToken // for stranding, like 1 23 456

func (ppb ppBlock) ppexpr()  {}
func (pps ppStrand) ppexpr() {}
func (t ppToken) ppexpr()    {}
