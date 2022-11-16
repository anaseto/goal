package goal

import (
	"fmt"
	"strings"
)

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
	opApplyV
	opApply2
	opApply2V
	opApplyN
	opApplyNV
	opDrop
	opJumpFalse
	opJump
	opReturn

	opArg = -1 // argument to be computed later
)

func (opc opcode) argc() int {
	switch opc {
	case opNop, opNil, opApply, opApply2, opDrop, opReturn:
		return 0
	case opApplyNV:
		return 2
	default:
		return 1
	}
}

func (ctx *Context) opcodesString(ops []opcode, lc *LambdaCode) string {
	sb := &strings.Builder{}
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		var pos int
		if lc != nil {
			pos = lc.Pos[i]
		} else {
			pos = ctx.prog.Pos[i]
		}
		fmt.Fprintf(sb, "%3d %3d %s\t", i, pos, op)
		switch op {
		case opConst:
			fmt.Fprintf(sb, "%d", ops[i+1])
		case opGlobal:
			fmt.Fprintf(sb, "%d (%s)", ops[i+1], ctx.gNames[int(ops[i+1])])
		case opLocal:
			fmt.Fprintf(sb, "%d (%s)", ops[i+1], lc.Names[int(ops[i+1])])
		case opAssignGlobal:
			fmt.Fprintf(sb, "%d (%s)", ops[i+1], ctx.gNames[int(ops[i+1])])
		case opAssignLocal:
			fmt.Fprintf(sb, "%d (%s)", ops[i+1], lc.Names[ops[i+1]])
		case opVariadic:
			fmt.Fprintf(sb, "%s", ctx.variadicsNames[ops[i+1]])
		case opLambda:
			fmt.Fprintf(sb, "%d", ops[i+1])
		case opApplyV:
			fmt.Fprintf(sb, "%s", ctx.variadicsNames[ops[i+1]])
		case opApply2V:
			fmt.Fprintf(sb, "%s", ctx.variadicsNames[ops[i+1]])
		case opApplyN:
			fmt.Fprintf(sb, "%d", ops[i+1])
		case opApplyNV:
			fmt.Fprintf(sb, "%s\t%d", ctx.variadicsNames[ops[i+1]], ops[i+2])
		case opJump:
			fmt.Fprintf(sb, "%d", ops[i+1])
		case opJumpFalse:
			fmt.Fprintf(sb, "%d", ops[i+1])
		}
		fmt.Fprint(sb, "\n")
		i += op.argc()
	}
	return sb.String()
}
