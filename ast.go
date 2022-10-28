package main

import "fmt"

type Expr interface {
	expr()
}

type astExprs []Expr

type astIdent struct {
	index int
	name  string
}

type astInt struct {
	value I
}

type astString struct {
	value S
}

type astUnary struct {
	token Token
	right Expr
}

type astBinary struct {
	token Token
	left  Expr
	right Expr
}

func (e astExprs) expr()  {}
func (e astIdent) expr()  {}
func (e astInt) expr()    {}
func (e astString) expr() {}
func (e astUnary) expr()  {}
func (e astBinary) expr() {}

type ppExpr interface {
	ppexpr()
}

type ppToken struct {
	Type ppTokenType
	Line int
	Text string
}

func (ppt ppToken) String() string {
	return fmt.Sprintf("{%v %d %s}", ppt.Type, ppt.Line, ppt.Text)
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
