package main

type Expr interface {
	expr()
}

type identExpr struct {
	index int
	name  string
}

type iExpr struct {
	value I
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

func (e astUnary) expr()  {}
func (e astBinary) expr() {}
