package goal

import (
	"fmt"
	"strings"
)

type opcode int32

const (
	opNop opcode = iota
	opNil
	opConst
	opInt
	opVariadic
	opLambda
	opLocal
	opLocalLast
	opGlobal
	opGlobalLast
	opAssignLocal
	opAssignGlobal
	opListAssignLocal
	opListAssignGlobal
	opApply
	opApplyV
	opApplyGlobal
	opDerive
	opApply2
	opApply2V
	opApplyN
	opApplyNGlobal
	opApplyNV
	opDrop
	opJump
	opJumpFalse
	opJumpTrue
	opReturn
	opTry

	opArg = -1 // argument to be computed later
)

func (opc opcode) argc() int {
	switch opc {
	case opNop, opNil, opApply, opApply2, opDrop, opReturn, opTry:
		return 0
	case opApplyNV, opApplyNGlobal:
		return 2
	default:
		return 1
	}
}

func (ctx *Context) opcodesString(ops []opcode, lc *lambdaCode) string {
	sb := strings.Builder{}
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		var pos int
		if lc != nil {
			pos = lc.Pos[i]
		} else {
			pos = ctx.gCode.Pos[i]
		}
		fmt.Fprintf(&sb, "%3d %3d %s\t", i, pos, op)
		switch op {
		case opConst, opInt, opLambda, opApplyN:
			fmt.Fprintf(&sb, "%d", ops[i+1])
		case opGlobal, opGlobalLast, opAssignGlobal, opApplyGlobal:
			fmt.Fprintf(&sb, "%d (%s)", ops[i+1], ctx.gNames[int(ops[i+1])])
		case opListAssignGlobal:
			ids := ctx.gAssignLists[ops[i+1]]
			names := make([]string, len(ids))
			for i, id := range ids {
				names[i] = ctx.gNames[id]
			}
			fmt.Fprintf(&sb, "%d (%s)", ops[i+1], strings.Join(names, ","))
		case opApplyNGlobal:
			fmt.Fprintf(&sb, "%d (%s)\t%d", ops[i+1], ctx.gNames[int(ops[i+1])], ops[i+2])
		case opLocal, opLocalLast, opAssignLocal:
			fmt.Fprintf(&sb, "%d (%s)", ops[i+1], lc.Names[int(ops[i+1])])
		case opListAssignLocal:
			ids := lc.AssignLists[ops[i+1]]
			names := make([]string, len(ids))
			for i, id := range ids {
				names[i] = lc.Names[id]
			}
			fmt.Fprintf(&sb, "%d (%s)", ops[i+1], strings.Join(names, ","))
		case opVariadic, opApplyV, opApply2V:
			fmt.Fprintf(&sb, "%s", ctx.variadicsNames[ops[i+1]])
		case opApplyNV:
			fmt.Fprintf(&sb, "%s\t%d", ctx.variadicsNames[ops[i+1]], ops[i+2])
		case opJump, opJumpFalse, opJumpTrue:
			fmt.Fprintf(&sb, "%d", int(ops[i+1])+1+i)
		}
		fmt.Fprint(&sb, "\n")
		i += op.argc()
	}
	return sb.String()
}
