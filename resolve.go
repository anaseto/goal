package goal

import "fmt"

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
