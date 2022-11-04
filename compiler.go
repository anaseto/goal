package main

import (
	"fmt"
	"strings"
)

// Program represents a compiled program.
type Program struct {
	Body      []opcode
	Constants []V
	Globals   map[int]string
	Lambdas   []*LambdaCode
}

// LambdaCode represents a compiled user defined function.
type LambdaCode struct {
	Body   []opcode
	Locals []string
	Arity  int
}

func (prog *Program) String() string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "---- Program -----\n")
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, prog.opcodesString(prog.Body, nil))
	fmt.Fprintln(sb, "Globals:")
	for id, name := range prog.Globals {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, id)
	}
	fmt.Fprintln(sb, "Constants:")
	for id, v := range prog.Constants {
		fmt.Fprintf(sb, "\t%d\t%v\n", id, v)
	}
	for id, lc := range prog.Lambdas {
		fmt.Fprintf(sb, "---- Lambda %d (Arity: %d) -----\n", id, lc.Arity)
		fmt.Fprintf(sb, "%s", prog.lambdaString(lc))
	}
	return sb.String()
}

func (prog *Program) lambdaString(lc *LambdaCode) string {
	sb := &strings.Builder{}
	fmt.Fprintln(sb, "Instructions:")
	fmt.Fprint(sb, prog.opcodesString(lc.Body, lc))
	fmt.Fprintln(sb, "Locals:")
	for i, name := range lc.Locals {
		fmt.Fprintf(sb, "\t%s\t%d\n", name, i)
	}
	return sb.String()
}

func (prog *Program) opcodesString(ops []opcode, lc *LambdaCode) string {
	sb := &strings.Builder{}
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		switch op {
		case opNop:
			fmt.Fprintf(sb, "%d\t%s\n", i, op)
		case opConst:
			fmt.Fprintf(sb, "%d\t%s\t\t%d\n", i, op, ops[i+1])
			i++
		case opGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], prog.Globals[int(ops[i+1])])
			i++
		case opLocal:
			fmt.Fprintf(sb, "%d\t%s\t\t%d (%s)\n", i, op, ops[i+1], lc.Locals[int(ops[i+1])])
			//fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opAssignGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], prog.Globals[int(ops[i+1])])
			i++
		case opAssignLocal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], lc.Locals[ops[i+1]])
			i++
		case opMonad:
			fmt.Fprintf(sb, "%d\t%s\t\t%s\n", i, op, Monad(ops[i+1]))
			i++
		case opDyad:
			fmt.Fprintf(sb, "%d\t%s\t\t%s\n", i, op, Dyad(ops[i+1]))
			i++
		case opAdverb:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, Adverb(ops[i+1]))
			i++
		case opVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, Variadic(ops[i+1]))
			i++
		case opLambda:
			fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opApply:
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

// Compile transforms an AstProgram into a Program.
func Compile(aprog *AstProgram) *Program {
	prog := &Program{}
	prog.Constants = aprog.Constants
	prog.Globals = map[int]string{}
	for k, id := range aprog.Globals {
		prog.Globals[id] = k
	}
	prog.compileBody(aprog.Body)
	prog.compileLambdas(aprog.Lambdas)
	return prog
}

func (prog *Program) compileBody(body []Expr) {
	for _, expr := range body {
		prog.Body, _ = compileExpr(prog.Body, expr)
	}
}

func compileExpr(body []opcode, expr Expr) ([]opcode, bool) {
	switch expr := expr.(type) {
	case AstConst:
		body = append(body, opConst, opcode(expr.ID))
	case AstGlobal:
		body = append(body, opGlobal, opcode(expr.ID))
	case AstAssignGlobal:
		body = append(body, opAssignGlobal, opcode(expr.ID))
	case AstMonad:
		body = append(body, opMonad, opcode(expr.Monad))
	case AstDyad:
		body = append(body, opDyad, opcode(expr.Dyad))
	case AstVariadic:
		body = append(body, opVariadic, opcode(expr.Variadic))
	case AstAdverb:
		body = append(body, opAdverb, opcode(expr.Adverb))
	case AstLambda:
		body = append(body, opLambda, opcode(expr.Lambda))
	case AstApply:
		body = append(body, opApply)
	case AstDrop:
		body = append(body, opDrop)
	case AstApplyN:
		body = append(body, opApplyN, opcode(expr.N))
	default:
		return body, false
	}
	return body, true
}

func (prog *Program) compileLambdas(lcs []*AstLambdaCode) {
	for _, lc := range lcs {
		prog.Lambdas = append(prog.Lambdas, prog.compileLambda(lc))
	}
}

func (prog *Program) compileLambda(lc *AstLambdaCode) *LambdaCode {
	nargs := 0
	nlocals := 0
	for _, local := range lc.Locals {
		nlocals++
		if local.Type == LocalArg {
			nargs++
		}
	}
	clc := &LambdaCode{}
	clc.Arity = nargs
	locals := make([]string, nlocals)
	getID := func(local Local) int {
		switch local.Type {
		case LocalArg:
			return local.ID
		case LocalVar:
			return local.ID + nargs
		default:
			panic(fmt.Sprintf("unknown local type: %d", local.Type))
		}
	}
	for k, local := range lc.Locals {
		locals[getID(local)] = k
	}
	clc.Locals = locals
	for _, expr := range lc.Body {
		var done bool
		clc.Body, done = compileExpr(clc.Body, expr)
		if !done {
			switch expr := expr.(type) {
			case AstLocal:
				clc.Body = append(clc.Body, opLocal, opcode(getID(expr.Local)))
			case AstAssignLocal:
				clc.Body = append(clc.Body, opAssignLocal, opcode(getID(expr.Local)))
			}
		}
	}
	return clc
}
