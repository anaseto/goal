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
	opApplyVariadic
	opApply2
	opApply2Variadic
	opApplyN
	opApplyNVariadic
	opDrop

	opArg = -1 // argument to be computed later
)

func (opc opcode) argc() int {
	switch opc {
	case opNop, opNil, opApply, opApply2, opDrop:
		return 0
	case opApplyNVariadic:
		return 2
	default:
		return 1
	}
}

func (ctx *Context) opcodesString(ops []opcode, lc *LambdaCode) string {
	sb := &strings.Builder{}
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		switch op {
		case opNop:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opConst:
			fmt.Fprintf(sb, "%d\t%s\t\t%d\n", i, op, ops[i+1])
		case opNil:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], ctx.gNames[int(ops[i+1])])
		case opLocal:
			fmt.Fprintf(sb, "%d\t%s\t\t%d (%s)\n", i, op, ops[i+1], lc.Names[int(ops[i+1])])
		case opAssignGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], ctx.gNames[int(ops[i+1])])
		case opAssignLocal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], lc.Names[ops[i+1]])
		case opVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, ctx.variadicsNames[ops[i+1]])
		case opLambda:
			fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
		case opApply:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opApplyVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, ctx.variadicsNames[ops[i+1]])
		case opApply2:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opApply2Variadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, ctx.variadicsNames[ops[i+1]])
		case opApplyN:
			fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
		case opApplyNVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\t%d\n", i, op, ctx.variadicsNames[ops[i+1]], ops[i+2])
		case opDrop:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		}
		i += op.argc()
	}
	return sb.String()
}
