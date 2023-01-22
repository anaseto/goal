package goal

// Clone creates an identical deep copy of a value, or the value itself if it
// is reusable. It initializes refcount if necessary.
func (x V) Clone() V {
	if x.kind != valBoxed {
		return x
	}
	var p *int32
	switch xv := x.value.(type) {
	case array:
		p = xv.RC()
		if !reuseRCp(p) {
			var n int32
			p = &n
		}
	default:
		var n int32
		p = &n
	}
	return x.CloneWithRC(p)
}

func (v V) CloneWithRC(rc *int32) V {
	if v.kind != valBoxed {
		return v
	}
	return NewV(v.value.CloneWithRC(rc))
}

func (s S) CloneWithRC(rc *int32) Value {
	return s
}

func (e panicV) CloneWithRC(rc *int32) Value {
	return e
}

func (e *errV) CloneWithRC(rc *int32) Value {
	return &errV{V: e.V.CloneWithRC(rc)}
}

func (x *AB) CloneWithRC(rc *int32) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AB{Slice: make([]bool, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AI) CloneWithRC(rc *int32) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AI{Slice: make([]int64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AF) CloneWithRC(rc *int32) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AF{Slice: make([]float64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AS) CloneWithRC(rc *int32) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AS{Slice: make([]string, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AV) CloneWithRC(rc *int32) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		for i, xi := range x.Slice {
			x.Slice[i] = xi.CloneWithRC(rc)
		}
		return x
	}
	r := &AV{Slice: make([]V, x.Len()), rc: rc}
	for i, xi := range x.Slice {
		r.Slice[i] = xi.CloneWithRC(rc)
	}
	return r
}

func (p *projection) CloneWithRC(rc *int32) Value {
	np := &projection{Fun: p.Fun.CloneWithRC(rc), Args: make([]V, len(p.Args))}
	for i, arg := range p.Args {
		np.Args[i] = arg.CloneWithRC(rc)
	}
	return np
}

func (p *projectionFirst) CloneWithRC(rc *int32) Value {
	if p.Fun.HasRC() || p.Arg.HasRC() {
		return &projectionFirst{Fun: p.Fun.CloneWithRC(rc), Arg: p.Arg.CloneWithRC(rc)}
	}
	return p
}

func (p *projectionMonad) CloneWithRC(rc *int32) Value {
	if p.Fun.HasRC() {
		return &projectionMonad{Fun: p.Fun.CloneWithRC(rc)}
	}
	return p
}

func (r *derivedVerb) CloneWithRC(rc *int32) Value {
	if r.Arg.HasRC() {
		return &derivedVerb{Fun: r.Fun, Arg: r.Arg.CloneWithRC(rc)}
	}
	return r
}

func (r *nReplacer) CloneWithRC(rc *int32) Value {
	return r
}

func (r *replacer) CloneWithRC(rc *int32) Value {
	if r.oldnew.rc == nil || *r.oldnew.rc <= 1 || r.oldnew.rc == rc {
		r.oldnew.rc = rc
		return r
	}
	olnew := &AS{Slice: make([]string, r.oldnew.Len()), rc: rc}
	copy(olnew.Slice, r.oldnew.Slice)
	return &replacer{r: r.r, oldnew: olnew}
}

func (r *rx) CloneWithRC(rc *int32) Value {
	return r
}

func (r *rxReplacer) CloneWithRC(rc *int32) Value {
	if r.repl.HasRC() {
		return &rxReplacer{r: r.r, repl: r.repl.CloneWithRC(rc)}
	}
	return r
}
