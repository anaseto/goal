package goal

// LessT returns true if x is ordered before y. It represents a strict
// total order (except non-strict for NaNs). Values are ordered as follows:
// unboxed atoms first (numbers, variadics, then lambdas), then boxed values.
// Otherwise, values are compared with < and > when comparable, and otherwise
// using their Type string value. As a special case, comparable arrays are
// compared first by length, or lexicographically if they are of equal length.
func (x V) LessT(y V) bool {
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
			return x.uv < y.uv
		}
	case valLambda:
		if y.kind == valLambda {
			return x.uv < y.uv
		}
	case valBoxed:
		if y.kind == valBoxed {
			return x.bv.LessT(y.bv)
		}
	case valPanic:
		if y.kind == valPanic {
			return x.bv.LessT(y.bv)
		}
	}
	return x.kind < y.kind
}

// LessT satisfies the specification of the BV interface.
func (s S) LessT(y BV) bool {
	switch yv := y.(type) {
	case S:
		return s < yv
	default:
		return s.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (x *AB) LessT(y BV) bool {
	switch yv := y.(type) {
	case *AB:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len(); i++ {
			if x.At(i) < yv.At(i) {
				return true
			}
			if x.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	case *AI:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len(); i++ {
			if int64(x.At(i)) < yv.At(i) {
				return true
			}
			if int64(x.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	case *AF:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len(); i++ {
			if float64(x.At(i)) < yv.At(i) {
				return true
			}
			if float64(x.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return x.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (x *AI) LessT(y BV) bool {
	switch yv := y.(type) {
	case *AB:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < int64(yv.At(i)) {
				return true
			}
			if x.At(i) > int64(yv.At(i)) {
				return false
			}
		}
		return false
	case *AI:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < yv.At(i) {
				return true
			}
			if x.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	case *AF:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if float64(x.At(i)) < yv.At(i) {
				return true
			}
			if float64(x.At(i)) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return x.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (x *AF) LessT(y BV) bool {
	switch yv := y.(type) {
	case *AB:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < float64(yv.At(i)) {
				return true
			}
			if x.At(i) > float64(yv.At(i)) {
				return false
			}
		}
		return false
	case *AI:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < float64(yv.At(i)) {
				return true
			}
			if x.At(i) > float64(yv.At(i)) {
				return false
			}
		}
		return false
	case *AF:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < yv.At(i) {
				return true
			}
			if x.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return x.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (x *AS) LessT(y BV) bool {
	switch yv := y.(type) {
	case *AS:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i) < yv.At(i) {
				return true
			}
			if x.At(i) > yv.At(i) {
				return false
			}
		}
		return false
	default:
		return x.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (x *AV) LessT(y BV) bool {
	switch yv := y.(type) {
	case *AV:
		if x.Len() != yv.Len() {
			return x.Len() < yv.Len()
		}
		for i := 0; i < x.Len() && i < yv.Len(); i++ {
			if x.At(i).LessT(yv.At(i)) {
				return true
			}
			if yv.At(i).LessT(x.At(i)) {
				return false
			}
		}
		return false
	default:
		return x.Type() < y.Type()
	}
}

// LessT satisfies the specification of the BV interface.
func (d *D) LessT(y BV) bool {
	switch yv := y.(type) {
	case *D:
		return d.keys.LessT(yv.keys) || d.keys.Matches(yv.keys) && d.values.LessT(yv.values)
	default:
		return d.Type() < y.Type()
	}
}

func (xv *derivedVerb) LessT(y BV) bool {
	switch yv := y.(type) {
	case *derivedVerb:
		return xv.Fun < yv.Fun ||
			xv.Fun == yv.Fun && xv.Arg.LessT(yv.Arg)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projection) LessT(y BV) bool {
	switch yv := y.(type) {
	case *projection:
		return xv.Fun.LessT(yv.Fun) ||
			xv.Fun.Matches(yv.Fun) && newAVu(xv.Args).LessT(newAVu(yv.Args))
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projectionFirst) LessT(y BV) bool {
	switch yv := y.(type) {
	case *projectionFirst:
		return xv.Fun.LessT(yv.Fun) ||
			xv.Fun.Matches(yv.Fun) && xv.Arg.LessT(yv.Arg)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *projectionMonad) LessT(y BV) bool {
	switch yv := y.(type) {
	case *projectionMonad:
		return xv.Fun.LessT(yv.Fun)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *nReplacer) LessT(y BV) bool {
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

func (xv *replacer) LessT(y BV) bool {
	switch yv := y.(type) {
	case *replacer:
		return xv.oldnew.LessT(yv.oldnew)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *rxReplacer) LessT(y BV) bool {
	switch yv := y.(type) {
	case *rxReplacer:
		return xv.r.LessT(yv.r) || xv.r.Matches(yv.r) && xv.repl.LessT(yv.repl)
	case function:
		return xv.stype() < yv.stype()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *rx) LessT(y BV) bool {
	switch yv := y.(type) {
	case *rx:
		return xv.Regexp.String() < yv.Regexp.String()
	default:
		return xv.Type() < y.Type()
	}
}

func (xv *errV) LessT(y BV) bool {
	switch yv := y.(type) {
	case *errV:
		return xv.V.LessT(yv.V)
	default:
		return xv.Type() < y.Type()
	}
}

func (xv panicV) LessT(y BV) bool {
	switch yv := y.(type) {
	case panicV:
		return xv < yv
	default:
		// Should not happen in regular goal code, but panics are
		// always valPanic, greater than other kinds of boxed values.
		return false
	}
}
