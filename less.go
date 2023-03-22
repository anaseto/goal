package goal

// LessV returns true if x is ordered before y. It represents a strict
// total order. Values are ordered as follows: unboxed atoms first
// (nums, variadics, then lambdas), then boxed values. Otherwise, values
// are compared with < and > when comparable, and otherwise using their
// Type string value. As a special case, comparable arrays are compared
// first by length, or lexicographically if they are of equal length.
func (x V) LessV(y V) bool {
	switch x.kind {
	case valInt:
		if y.IsI() {
			return x.I() < y.I()
		}
		if y.IsF() {
			return float64(x.I()) < y.F()
		}
	case valFloat:
		if y.IsI() {
			return x.F() < float64(y.I())
		}
		if y.IsF() {
			return x.F() < y.F()
		}
	case valVariadic:
		if y.kind == valVariadic {
			return x.n < y.n
		}
	case valLambda:
		if y.kind == valLambda {
			return x.n < y.n
		}
	case valBoxed:
		if y.kind == valBoxed {
			return x.value.LessV(y.value)
		}
	case valPanic:
		if y.kind == valPanic {
			return x.value.LessV(y.value)
		}
	}
	return x.kind < y.kind
}

func (xv S) LessV(y Value) bool {
	switch yv := y.(type) {
	case S:
		return xv < yv
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *AB) LessV(y Value) bool {
	switch yv := y.(type) {
	case *AB:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len(); i++ {
			if !xv.At(i) && yv.At(i) {
				return true
			}
			if xv.At(i) && !yv.At(i) {
				return false
			}
		}
		return false
	case *AF:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len(); i++ {
			if b2f(xv.At(i)) < yv.At(i) {
				return true
			}
			if b2f(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	case *AI:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len(); i++ {
			if b2i(xv.At(i)) < yv.At(i) {
				return true
			}
			if b2i(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *AI) LessV(y Value) bool {
	switch yv := y.(type) {
	case *AB:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < b2i(yv.At(i)) {
				return true
			}
			if xv.At(i) > b2i(yv.At(i)) {
				return false
			}
		}
		return false
	case *AF:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if float64(xv.At(i)) < yv.At(i) {
				return true
			}
			if float64(xv.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	case *AI:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *AF) LessV(y Value) bool {
	switch yv := y.(type) {
	case *AB:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < b2f(yv.At(i)) {
				return true
			}
			if xv.At(i) > b2f(yv.At(i)) {
				return false
			}
		}
		return false
	case *AF:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	case *AI:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < float64(yv.At(i)) {
				return true
			}
			if xv.At(i) > float64(yv.At(i)) {
				return false
			}
		}
		return false
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *AS) LessV(y Value) bool {
	switch yv := y.(type) {
	case *AS:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i) < yv.At(i) {
				return true
			}
			if xv.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *AV) LessV(y Value) bool {
	switch yv := y.(type) {
	case *AV:
		if xv.Len() != yv.Len() {
			return xv.Len() < yv.Len()
		}
		for i := 0; i < xv.Len() && i < yv.Len(); i++ {
			if xv.At(i).LessV(yv.At(i)) {
				return true
			}
			if yv.At(i).LessV(xv.At(i)) {
				return false
			}
		}
		return false
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *derivedVerb) LessV(y Value) bool {
	switch yv := y.(type) {
	case *derivedVerb:
		return xv.Fun < yv.Fun ||
			xv.Fun == yv.Fun && xv.Arg.LessV(yv.Arg)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projection) LessV(y Value) bool {
	switch yv := y.(type) {
	case *projection:
		return xv.Fun.LessV(yv.Fun) ||
			Match(xv.Fun, yv.Fun) && NewAV(xv.Args).LessV(NewAV(yv.Args))
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projectionFirst) LessV(y Value) bool {
	switch yv := y.(type) {
	case *projectionFirst:
		return xv.Fun.LessV(yv.Fun) ||
			Match(xv.Fun, yv.Fun) && xv.Arg.LessV(yv.Arg)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projectionMonad) LessV(y Value) bool {
	switch yv := y.(type) {
	case *projectionMonad:
		return xv.Fun.LessV(yv.Fun)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *nReplacer) LessV(y Value) bool {
	switch yv := y.(type) {
	case *nReplacer:
		return xv.olds < yv.olds || xv.olds == yv.olds &&
			(xv.news < yv.news || xv.news == yv.news && xv.n < yv.n)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *replacer) LessV(y Value) bool {
	switch yv := y.(type) {
	case *replacer:
		return xv.oldnew.LessV(yv.oldnew)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *rxReplacer) LessV(y Value) bool {
	switch yv := y.(type) {
	case *rxReplacer:
		return xv.r.LessV(yv.r) || xv.r.Matches(yv.r) && xv.repl.LessV(yv.repl)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *rx) LessV(y Value) bool {
	switch yv := y.(type) {
	case *rx:
		return xv.Regexp.String() < yv.Regexp.String()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *errV) LessV(y Value) bool {
	switch yv := y.(type) {
	case *errV:
		return xv.V.LessV(yv.V)
	default:
		return xv.Type() < y.Type()
	}
}

func (xv panicV) LessV(y Value) bool {
	switch yv := y.(type) {
	case panicV:
		return xv < yv
	default:
		// Should not happen in regular goal code, but panics are
		// always valPanic, greater than other kinds of boxed values.
		return false
	}
}
