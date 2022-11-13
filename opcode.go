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

func (ctx *Context) opcodesString(ops []opcode, lc *LambdaCode) string {
	sb := &strings.Builder{}
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		switch op {
		case opNop:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opConst:
			fmt.Fprintf(sb, "%d\t%s\t\t%d\n", i, op, ops[i+1])
			i++
		case opNil:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], ctx.gNames[int(ops[i+1])])
			i++
		case opLocal:
			fmt.Fprintf(sb, "%d\t%s\t\t%d (%s)\n", i, op, ops[i+1], lc.Names[int(ops[i+1])])
			//fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opAssignGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], ctx.gNames[int(ops[i+1])])
			i++
		case opAssignLocal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], lc.Names[ops[i+1]])
			i++
		case opVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, ctx.variadicsNames[ops[i+1]])
			i++
		case opLambda:
			fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opApply:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opApply2:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opApplyN:
			fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opDrop:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		}
	}
	return sb.String()
}
