package main

type opcode int32

const (
	opNop opcode = iota
	opConst
	opGlobal
	opLocal
	opAssignGlobal
	opAssignLocal
	opValue
	opJumpTrue
	opJumpFalse
	opApply
	opApplyN
	opReturn
)
