package main

import "fmt"

type Program struct {
	Body      []Expr
	Constants []V
	Globals   []V

	constants map[string]int // for generating symbols
	globals   map[string]int // for generating symbols
}

type Expr interface {
	node()
}

type Exprs []Expr

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
type AstAdverb Dyad

type AstApply struct {
	Value Expr
	Arity int
	Pos   int
}

type AstLambda struct {
	Body   []Expr
	Locals []Symbol
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

// ppExpr represents a preprocessing builds blocks and forms nouns
// without giving meaning yet.
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
