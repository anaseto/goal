package goal

import (
	"fmt"
	"strings"
)

// Program represents a compiled program.
type Program struct {
	Body    []opcode
	Pos     []int
	Lambdas []*LambdaCode

	cLambdas int // index next of last compiled lambda
	cBody    int // number of already processed body ops
	last     int // index of last non-argument opcode
}

// LambdaCode represents a compiled user defined function.
type LambdaCode struct {
	Body      []opcode
	Pos       []int
	Names     []string
	Rank      int
	NamedArgs bool
	Locals    map[string]Local // arguments and variables

	locals map[int]Local // opcode index -> local variable
	nVars  int
}

// Local represents either an argument or a local variable. IDs are
// unique for a given type only.
type Local struct {
	Type LocalType
	ID   int
}

// LocalType represents different kinds of locals.
type LocalType int

// These constants describe the supported kinds of locals.
const (
	LocalArg LocalType = iota
	LocalVar
)

func (l *LambdaCode) local(s string) (Local, bool) {
	param, ok := l.Locals[s]
	if ok {
		return param, true
	}
	if !l.NamedArgs && len(s) == 1 {
		switch r := rune(s[0]); r {
		case 'x', 'y', 'z':
			id := r - 'x'
			arg := Local{Type: LocalArg, ID: int(id)}
			l.Locals[s] = arg
			return arg, true
		}
	}
	return Local{}, false
}

// ProgramString returns a string representation of the compiled program and
// relevant data.
func (ctx *Context) ProgramString() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "---- Compiled program -----")
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, ctx.opcodesString(ctx.prog.Body, nil))
	fmt.Fprintln(sb, "Globals:")
	for id, name := range ctx.gNames {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(sb, "Constants:")
	for id, v := range ctx.constants {
		fmt.Fprintf(sb, "\t%d\t%v\n", id, v)
	}
	for id, lc := range ctx.prog.Lambdas {
		fmt.Fprintf(sb, "---- Lambda %d (Rank: %d) -----\n", id, lc.Rank)
		fmt.Fprintf(sb, "%s", ctx.lambdaString(lc))
	}
	return sb.String()
}

func (ctx *Context) lambdaString(lc *LambdaCode) string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, ctx.opcodesString(lc.Body, lc))
	fmt.Fprintln(sb, "Locals:")
	for i, name := range lc.Names {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, i)
	}
	return sb.String()
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

// resolve resolves names in lambdas and updates the object code. It returns
// false if all code had already been processed.
func (ctx *Context) resolve() bool {
	if ctx.prog.cLambdas == len(ctx.prog.Lambdas) &&
		ctx.prog.cBody == len(ctx.prog.Body) {
		return false
	}
	ctx.resolveLambdas()
	ctx.prog.cBody = len(ctx.prog.Body)
	return true

}

func (ctx *Context) resolveLambdas() {
	for _, lc := range ctx.prog.Lambdas[ctx.prog.cLambdas:] {
		ctx.resolveLambda(lc)
	}
	ctx.prog.cLambdas = len(ctx.prog.Lambdas)
}

func (ctx *Context) resolveLambda(lc *LambdaCode) {
	nargs := 0
	nlocals := 0
	for _, local := range lc.Locals {
		nlocals++
		if local.Type == LocalArg {
			nargs++
		}
	}
	if nargs == 0 {
		// All lambdas have at least one argument, even if not used.
		nlocals++
		nargs = 1
	}
	nvars := nlocals - nargs
	lc.Rank = nargs
	names := make([]string, nlocals)
	getID := func(local Local) int {
		switch local.Type {
		case LocalArg:
			return local.ID + nvars
		case LocalVar:
			return local.ID
		default:
			panic(fmt.Sprintf("unknown local type: %d", local.Type))
		}
	}
	for k, local := range lc.Locals {
		names[getID(local)] = k
	}
	lc.Names = names
	for ip := 0; ip < len(lc.Body); {
		op := lc.Body[ip]
		ip++
		switch op {
		case opLocal:
			lc.Body[ip] = opcode(getID(lc.locals[ip]))
		case opAssignLocal:
			lc.Body[ip] = opcode(getID(lc.locals[ip]))
		}
		if op.hasArg() {
			ip++
		}
	}
}
