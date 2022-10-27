package main

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

type ppTokenType int

const (
	ppSEP ppTokenType = iota
	ppCLOSE
	ppADVERB
	ppIDENT
	ppNUMBER
	ppSTRING
	ppVERB
)

type ppBlock struct {
	Type ppBlockType
	ppexprs []ppExpr
}

type ppBlockType int

const (
	ppBRACE ppBlockType = iota
	ppBRACKET
	ppPAREN
)

type ppStrand []ppToken // for stranding, like 1 23 456

func (ppb ppBlock) ppexpr()   {}
func (pps ppStrand) ppexpr()  {}
func (t ppToken) ppexpr()       {}
