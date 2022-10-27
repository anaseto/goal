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

type ppBracket []ppExpr
type ppBrace []ppExpr
type ppParen []ppExpr
type ppStrand []Token // for stranding, like 1 23 456

func (pps ppBracket) ppexpr() {}
func (pps ppBrace) ppexpr()   {}
func (pps ppParen) ppexpr()   {}
func (pps ppStrand) ppexpr()  {}
func (t Token) ppexpr()       {}
