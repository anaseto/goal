package goal

// Clone creates an identical deep copy of a value, or the value itself if it
// is reusable. It initializes refcount if necessary.
func (x V) Clone() V {
	if x.kind != valBoxed {
		return x
	}
	var p *int
	switch xv := x.value.(type) {
	case RefCountHolder:
		p = xv.RC()
		if !reusableRCp(p) {
			var n int
			p = &n
		}
	default:
		var n int
		p = &n
	}
	return x.CloneWithRC(p)
}

func (x V) CloneWithRC(rc *int) V {
	if x.kind != valBoxed {
		return x
	}
	xc, ok := x.value.(RefCounter)
	if ok {
		return NewV(xc.CloneWithRC(rc))
	}
	return x
}

func (e *errV) CloneWithRC(rc *int) Value {
	if e.V.HasRC() {
		return &errV{V: e.V.CloneWithRC(rc)}
	}
	return e
}

func (x *AB) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AB{Slice: make([]bool, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AI) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AI{Slice: make([]int64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AF) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AF{Slice: make([]float64, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AS) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AS{Slice: make([]string, x.Len()), rc: rc}
	copy(r.Slice, x.Slice)
	return r
}

func (x *AV) CloneWithRC(rc *int) Value {
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

func (d *Dict) CloneWithRC(rc *int) Value {
	return &Dict{keys: d.keys.CloneWithRC(rc).(array), values: d.values.CloneWithRC(rc).(array)}
}

func (d *Dict) clone() *Dict {
	krc := d.keys.RC()
	if !reusableRCp(krc) || krc == nil {
		var n int
		krc = &n
	}
	vrc := d.values.RC()
	if !reusableRCp(vrc) || vrc == nil {
		var n int
		vrc = &n
	}
	return &Dict{keys: d.keys.CloneWithRC(krc).(array), values: d.values.CloneWithRC(vrc).(array)}
}

func (p *projection) CloneWithRC(rc *int) Value {
	np := &projection{Fun: p.Fun.CloneWithRC(rc), Args: make([]V, len(p.Args))}
	for i, arg := range p.Args {
		np.Args[i] = arg.CloneWithRC(rc)
	}
	return np
}

func (p *projectionFirst) CloneWithRC(rc *int) Value {
	if p.Fun.HasRC() || p.Arg.HasRC() {
		return &projectionFirst{Fun: p.Fun.CloneWithRC(rc), Arg: p.Arg.CloneWithRC(rc)}
	}
	return p
}

func (p *projectionMonad) CloneWithRC(rc *int) Value {
	if p.Fun.HasRC() {
		return &projectionMonad{Fun: p.Fun.CloneWithRC(rc)}
	}
	return p
}

func (r *derivedVerb) CloneWithRC(rc *int) Value {
	if r.Arg.HasRC() {
		return &derivedVerb{Fun: r.Fun, Arg: r.Arg.CloneWithRC(rc)}
	}
	return r
}

func (r *replacer) CloneWithRC(rc *int) Value {
	if r.oldnew.rc == nil || *r.oldnew.rc <= 1 || r.oldnew.rc == rc {
		r.oldnew.rc = rc
		return r
	}
	olnew := &AS{Slice: make([]string, r.oldnew.Len()), rc: rc}
	copy(olnew.Slice, r.oldnew.Slice)
	return &replacer{r: r.r, oldnew: olnew}
}

func (r *rxReplacer) CloneWithRC(rc *int) Value {
	if r.repl.HasRC() {
		return &rxReplacer{r: r.r, repl: r.repl.CloneWithRC(rc)}
	}
	return r
}
