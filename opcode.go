package goal

type opcode int32

const (
	opNop opcode = iota
	opConst
	opNil
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

	opArg = -1 // argument to be computed later
)

func (opc opcode) hasArg() bool {
	switch opc {
	case opNop, opNil, opApply, opApply2, opDrop:
		return false
	default:
		return true
	}
}
