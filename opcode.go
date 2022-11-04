package main

type opcode int32

const (
	opNop opcode = iota
	opConst
	opGlobal
	opLocal
	opAssignGlobal
	opAssignLocal
	opMonad
	opDyad
	opAdverb
	opVariadic
	opLambda
	opApply
	opApplyN
	opDrop
)
