package goal

// Clone creates an identical deep copy of a value, or the value itself if it
// is reusable. It initializes refcount if necessary.
func (x V) Clone() V {
	if x.kind != valBoxed {
		return x
	}
	var p *int
	switch xv := x.bv.(type) {
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

// CloneWithRC clones the given value using its CloneWithRC method, if it is a
// RefCounter, or returns it as-is otherwise for immutable values that do not
// need cloning.
func (x V) CloneWithRC(rc *int) V {
	if x.kind != valBoxed {
		return x
	}
	xc, ok := x.bv.(RefCounter)
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

// CloneWithRC satisfies the specification of the RefCounter interface.
func (x *AB) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AB{elts: make([]byte, x.Len()), rc: rc}
	copy(r.elts, x.elts)
	return r
}

// CloneWithRC satisfies the specification of the RefCounter interface.
func (x *AI) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AI{elts: make([]int64, x.Len()), rc: rc}
	copy(r.elts, x.elts)
	return r
}

// CloneWithRC satisfies the specification of the RefCounter interface.
func (x *AF) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AF{elts: make([]float64, x.Len()), rc: rc}
	copy(r.elts, x.elts)
	return r
}

// CloneWithRC satisfies the specification of the RefCounter interface.
func (x *AS) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		return x
	}
	r := &AS{elts: make([]string, x.Len()), rc: rc}
	copy(r.elts, x.elts)
	return r
}

// CloneWithRC satisfies the specification of the RefCounter interface.
func (x *AV) CloneWithRC(rc *int) Value {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.setFlags(flagNone)
		x.rc = rc
		for i, xi := range x.elts {
			x.elts[i] = xi.CloneWithRC(rc)
		}
		return x
	}
	r := &AV{elts: make([]V, x.Len()), rc: rc}
	for i, xi := range x.elts {
		r.elts[i] = xi.CloneWithRC(rc)
	}
	return r
}

// CloneWithRC satisfies the specification of the RefCounter interface.
func (d *Dict) CloneWithRC(rc *int) Value {
	return &Dict{keys: d.keys.CloneWithRC(rc).(array), values: d.values.CloneWithRC(rc).(array)}
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
	olnew := &AS{elts: make([]string, r.oldnew.Len()), rc: rc}
	copy(olnew.elts, r.oldnew.elts)
	return &replacer{r: r.r, oldnew: olnew}
}

func (r *rxReplacer) CloneWithRC(rc *int) Value {
	if r.repl.HasRC() {
		return &rxReplacer{r: r.r, repl: r.repl.CloneWithRC(rc)}
	}
	return r
}

func (x V) immutable() {
	if x.kind != valBoxed {
		return
	}
	xh, ok := x.bv.(RefCountHolder)
	if !ok {
		xc, ok := x.bv.(RefCounter)
		if ok {
			var n int = 2
			xc.InitWithRC(&n)
		}
		return
	}
	rc := xh.RC()
	if rc != nil {
		*rc += 2
	} else {
		var n int = 2
		xh.InitWithRC(&n)
	}
}
