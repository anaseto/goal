package goal

import (
	"fmt"
	"strings"
)

// Program represents a compiled program.
type Program struct {
	Body    []opcode
	Lambdas []*LambdaCode
}

// LambdaCode represents a compiled user defined function.
type LambdaCode struct {
	Body   []opcode
	Locals []string
	Rank   int
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
	for i, name := range lc.Locals {
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
			fmt.Fprintf(sb, "%d\t%s\t\t%d (%s)\n", i, op, ops[i+1], lc.Locals[int(ops[i+1])])
			//fmt.Fprintf(sb, "%d\t%s\t%d\n", i, op, ops[i+1])
			i++
		case opAssignGlobal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], ctx.gNames[int(ops[i+1])])
			i++
		case opAssignLocal:
			fmt.Fprintf(sb, "%d\t%s\t%d (%s)\n", i, op, ops[i+1], lc.Locals[ops[i+1]])
			i++
		case opVariadic:
			fmt.Fprintf(sb, "%d\t%s\t%s\n", i, op, builtins[ops[i+1]].Name)
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

// Compile transforms an AstProgram into a Program.
func (ctx *Context) compile() {
	ctx.compileBody()
	ctx.compileLambdas()
}

func (ctx *Context) compileBody() {
	for _, expr := range ctx.ast.Body[ctx.ast.cBody:] {
		ctx.prog.Body, _ = compileExpr(ctx.prog.Body, expr)
	}
	ctx.ast.cBody = len(ctx.ast.Body)
}

func compileExpr(body []opcode, expr Expr) ([]opcode, bool) {
	switch expr := expr.(type) {
	case AstConst:
		body = append(body, opConst, opcode(expr.ID))
	case AstNil:
		body = append(body, opNil)
	case AstGlobal:
		body = append(body, opGlobal, opcode(expr.ID))
	case AstAssignGlobal:
		body = append(body, opAssignGlobal, opcode(expr.ID))
	case AstVariadic:
		body = append(body, opVariadic, opcode(expr.Variadic))
	case AstLambda:
		body = append(body, opLambda, opcode(expr.Lambda))
	case AstApply:
		body = append(body, opApply)
	case AstApply2:
		body = append(body, opApply2)
	case AstApplyN:
		body = append(body, opApplyN, opcode(expr.N))
	case AstDrop:
		body = append(body, opDrop)
	default:
		return body, false
	}
	return body, true
}

func (ctx *Context) compileLambdas() {
	for _, lc := range ctx.ast.Lambdas[ctx.ast.cLambdas:] {
		ctx.prog.Lambdas = append(ctx.prog.Lambdas, ctx.compileLambda(lc))
	}
	ctx.ast.cLambdas = len(ctx.ast.Lambdas)
}

func (ctx *Context) compileLambda(lc *AstLambdaCode) *LambdaCode {
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
		nargs = 1
	}
	clc := &LambdaCode{}
	clc.Rank = nargs
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
