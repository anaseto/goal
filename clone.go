package goal

func (v V) Clone(rc *int32) V {
	if v.kind != valBoxed {
		return v
	}
	return NewV(v.value.Clone(rc))
}

func (s S) Clone(rc *int32) Value {
	return s
}

func (e panicV) Clone(rc *int32) Value {
	return e
}

func (e *errV) Clone(rc *int32) Value {
	return &errV{V: e.V.Clone(rc)}
}

func (x *AB) Clone(rc *int32) Value {
	if x.rc == rc {
		return x
	}
	r := &AB{Slice: make([]bool, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AI) Clone(rc *int32) Value {
	if x.rc == rc {
		return x
	}
	r := &AI{Slice: make([]int64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AF) Clone(rc *int32) Value {
	if x.rc == rc {
		return x
	}
	r := &AF{Slice: make([]float64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AS) Clone(rc *int32) Value {
	if x.rc == rc {
		return x
	}
	r := &AS{Slice: make([]string, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AV) Clone(rc *int32) Value {
	if x.rc == rc {
		// We assume all childs have same rc.
		return x
	}
	r := &AV{Slice: make([]V, x.Len()), rc: rc}
	for i, xi := range x.Slice {
		r.Slice[i] = xi.Clone(rc)
	}
	return r
}

func (p *projection) Clone(rc *int32) Value {
	np := &projection{Fun: p.Fun.Clone(rc), Args: make([]V, len(p.Args))}
	for i, arg := range p.Args {
		np.Args[i] = arg.Clone(rc)
	}
	return np
}

func (p *projectionFirst) Clone(rc *int32) Value {
	return &projectionFirst{Fun: p.Fun.Clone(rc), Arg: p.Arg.Clone(rc)}
}

func (p *projectionMonad) Clone(rc *int32) Value {
	return &projectionMonad{Fun: p.Fun.Clone(rc)}
}

func (r *derivedVerb) Clone(rc *int32) Value {
	return &derivedVerb{Fun: r.Fun, Arg: r.Arg.Clone(rc)}
}

func (r *nReplacer) Clone(rc *int32) Value {
	return r
}

func (r *replacer) Clone(rc *int32) Value {
	if r.oldnew.rc == rc {
		return r
	}
	olnew := &AS{Slice: make([]string, r.oldnew.Len()), rc: rc}
	copy(olnew.Slice, r.oldnew.Slice)
	return &replacer{r: r.r, oldnew: olnew}
}

func (r *rx) Clone(rc *int32) Value {
	return r
}

func (r *rxReplacer) Clone(rc *int32) Value {
	return &rxReplacer{r: r.r, repl: r.repl.Clone(rc)}
}
