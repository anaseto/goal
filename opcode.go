package main

type opcode int32

const (
	opNop opcode = iota
	opConst
	opGlobal
	opLocal
	opAssignGlobal
	opAssignLocal
	opAdverb
	opVariadic
	opLambda
	opApply
	opApply2
	opApplyN
	opDrop
)
