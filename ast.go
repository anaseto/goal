package main

type Expr interface {
	Eval(Context) O
}

type astUnary struct {
	v     O
	right Expr
}

type astBinary struct {
	v     O
	left  Expr
	right Expr
}
