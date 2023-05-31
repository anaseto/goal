package goal

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x V) Clone() V {
	if x.kind != valBoxed {
		return x
	}
	switch xv := x.bv.(type) {
	case RefCounter:
		x.bv = xv.Clone()
		return x
	default:
		return x
	}
}

func (e *errV) Clone() BV {
	if e.V.HasRC() {
		return &errV{V: e.V.Clone()}
	}
	return e
}

// sclone returns a shallow clone of an array. It may share memory with the parent if it
// was reusable. Any parent flags are reset.
func (x *A[T]) sclone() *A[T] {
	if x.reusable() {
		x.flags = flagNone
		return x
	}
	r := &A[T]{elts: make([]T, len(x.elts))}
	copy(r.elts, x.elts)
	return r
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x *AB) Clone() BV {
	return (*AB)((*A[byte])(x).sclone())
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x *AI) Clone() BV {
	return (*AI)((*A[int64])(x).sclone())
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x *AF) Clone() BV {
	return (*AF)((*A[float64])(x).sclone())
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x *AS) Clone() BV {
	return (*AS)((*A[string])(x).sclone())
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (x *AV) Clone() BV {
	x = (*AV)((*A[V])(x).sclone())
	for i, xi := range x.elts {
		x.elts[i] = xi.Clone()
	}
	return x
}

// Clone returns a clone of the value. Note that the cloned value might still
// share some structures with its parent if they're deemed reusable.
func (d *D) Clone() BV {
	return &D{keys: d.keys.Clone().(array), values: d.values.Clone().(array)}
}

func (p *projection) Clone() BV {
	np := &projection{Fun: p.Fun.Clone(), Args: make([]V, len(p.Args))}
	for i, arg := range p.Args {
		np.Args[i] = arg.Clone()
	}
	return np
}

func (p *projectionFirst) Clone() BV {
	if p.Fun.HasRC() || p.Arg.HasRC() {
		return &projectionFirst{Fun: p.Fun.Clone(), Arg: p.Arg.Clone()}
	}
	return p
}

func (p *projectionMonad) Clone() BV {
	if p.Fun.HasRC() {
		return &projectionMonad{Fun: p.Fun.Clone()}
	}
	return p
}

func (r *derivedVerb) Clone() BV {
	if r.Arg.HasRC() {
		return &derivedVerb{Fun: r.Fun, Arg: r.Arg.Clone()}
	}
	return r
}

func (r *replacer) Clone() BV {
	if r.oldnew.reusable() {
		return r
	}
	return &replacer{r: r.r, oldnew: (*AS)((*A[string])(r.oldnew).sclone())}
}

func (r *rxReplacer) Clone() BV {
	if r.repl.HasRC() {
		return &rxReplacer{r: r.r, repl: r.repl.Clone()}
	}
	return r
}
