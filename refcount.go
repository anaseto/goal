package goal

// RefCounter is implemented by values that use a reference count. In goal the
// refcount is not used for memory management, but only for optimization of
// memory allocations.  Refcount is increased by each assignement, and each
// push operation on the stack, except for pushes corresponding to the last use
// of a variable (as approximated conservatively). It is reduced after each
// drop.  If refcount is equal or less than one, then the value is considered
// reusable.
//
// When defining a new type implementing the Value interface, it is only
// necessary to also implement RefCounter if the type definition makes use of a
// type implementing it (for example an array type or a generic V).
type RefCounter interface {
	Value

	// IncrRC increments the reference count by one. It can panic if the
	// value's refcount pointer has not been properly initialized.
	IncrRC()

	// DecrRC decrements the reference count by one, or zero if it is
	// already zero.
	DecrRC()

	// InitWithRC recursively sets the refcount pointer for reusable
	// values, and increments by 2 the refcount of non-reusable values, to
	// ensure immutability of non-reusable children without cloning them.
	InitWithRC(rc *int)

	// CloneWithRC returns a clone of the value, with rc as new refcount
	// pointer.  If the current value's current refcount pointer is nil or
	// equal to the passed one, the same value is returned after updating
	// the refcount pointer as needed, instead of doing a full clone.
	CloneWithRC(rc *int) Value
}

// RefCountHolder is a RefCounter that has a root refcount pointer. When such
// values are returned from a variadic function, if the refcount pointer is
// still nil, InitWithRC is automatically called with a newly allocated
// refcount pointer to a zero count.
type RefCountHolder interface {
	RefCounter

	// RC returns the value's root reference count pointer.
	RC() *int
}

// HasRC returns true if the value is boxed and implements RefCounter.
func (x V) HasRC() bool {
	if x.kind != valBoxed {
		return false
	}
	_, ok := x.value.(RefCounter)
	return ok
}

// IncrRC increments the value reference count (if it has any).
func (x V) IncrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.IncrRC()
	}
}

// incrRC2 increments by 2 the value reference count (if it has any).
func (x V) incrRC2() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.IncrRC()
		xrc.IncrRC()
	}
}

// DecrRC decrements the value reference count (if it has any).
func (x V) DecrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

// decrRC2 decrements by 2 the value reference count (if it has any).
func (x V) decrRC2() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
		xrc.DecrRC()
	}
}

func (x V) rcdecrRefCounter() {
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

// RC returns the array's reference count pointer.
func (x *AB) RC() *int { return x.rc }

// RC returns the array's reference count pointer.
func (x *AI) RC() *int { return x.rc }

// RC returns the array's reference count pointer.
func (x *AF) RC() *int { return x.rc }

// RC returns the array's reference count pointer.
func (x *AS) RC() *int { return x.rc }

// RC returns the array's reference count pointer.
func (x *AV) RC() *int { return x.rc }

func reuseRCp(p *int) *int {
	if !reusableRCp(p) {
		var n int
		p = &n
	}
	return p
}

func reusableRCp(p *int) bool {
	if p == nil {
		return true
	}
	if *p <= 1 {
		*p = 0
		return true
	}
	return false
}

func (x *AB) reuse() *AB {
	if reusableRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AB{Slice: make([]bool, x.Len())}
}

func (x *AI) reuse() *AI {
	if reusableRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AI{Slice: make([]int64, x.Len())}
}

func (x *AF) reuse() *AF {
	if reusableRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AF{Slice: make([]float64, x.Len())}
}

func (x *AS) reuse() *AS {
	if reusableRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AS{Slice: make([]string, x.Len())}
}

func (x *AV) reuse() *AV {
	if reusableRCp(x.rc) {
		x.flags = flagNone
		x.rc = nil // NOTE: not always necessary, maybe use two functions
		return x
	}
	return &AV{Slice: make([]V, x.Len())}
}

func decrRCp(p *int) {
	if p != nil && *p > 0 {
		*p--
	}
}

func (x *AB) IncrRC() { *x.rc++ }
func (x *AI) IncrRC() { *x.rc++ }
func (x *AF) IncrRC() { *x.rc++ }
func (x *AS) IncrRC() { *x.rc++ }
func (x *AV) IncrRC() { *x.rc++ }

func (x *AB) DecrRC() { decrRCp(x.rc) }
func (x *AI) DecrRC() { decrRCp(x.rc) }
func (x *AF) DecrRC() { decrRCp(x.rc) }
func (x *AS) DecrRC() { decrRCp(x.rc) }
func (x *AV) DecrRC() { decrRCp(x.rc) }

func (d *Dict) IncrRC() {
	d.keys.IncrRC()
	d.values.IncrRC()
}

func (d *Dict) DecrRC() {
	d.keys.DecrRC()
	d.values.DecrRC()
}

func (r *derivedVerb) IncrRC() { r.Arg.IncrRC() }
func (r *derivedVerb) DecrRC() { r.Arg.DecrRC() }

func (p *projection) IncrRC() {
	p.Fun.IncrRC()
	for _, arg := range p.Args {
		arg.IncrRC()
	}
}

func (p *projection) DecrRC() {
	p.Fun.DecrRC()
	for _, arg := range p.Args {
		arg.DecrRC()
	}
}

func (p *projectionFirst) IncrRC() {
	p.Fun.IncrRC()
	p.Arg.IncrRC()
}

func (p *projectionFirst) DecrRC() {
	p.Fun.DecrRC()
	p.Arg.DecrRC()
}

func (p *projectionMonad) IncrRC() {
	p.Fun.IncrRC()
}

func (p *projectionMonad) DecrRC() {
	p.Fun.DecrRC()
}

func (e *errV) IncrRC()       { e.V.IncrRC() }
func (e *errV) DecrRC()       { e.V.DecrRC() }
func (r *replacer) IncrRC()   { r.oldnew.IncrRC() }
func (r *replacer) DecrRC()   { r.oldnew.DecrRC() }
func (r *rxReplacer) IncrRC() { r.repl.IncrRC() }
func (r *rxReplacer) DecrRC() { r.repl.DecrRC() }

// InitRC initializes refcount if the value is a RefCountHolder with nil
// refcount.
func (x V) InitRC() {
	if x.kind != valBoxed {
		return
	}
	xrch, ok := x.value.(RefCountHolder)
	if ok {
		initRC(xrch)
	}
}

func initRC(x RefCountHolder) {
	if x.RC() == nil {
		var n int
		x.InitWithRC(&n)
	}
}

// InitWithRC calls the method of the same name on boxed values.
func (x V) InitWithRC(rc *int) {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.InitWithRC(rc)
	}
}

func (e *errV) InitWithRC(rc *int) {
	e.V.InitWithRC(rc)
}

func (x *AB) InitWithRC(rc *int) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AI) InitWithRC(rc *int) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AF) InitWithRC(rc *int) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AS) InitWithRC(rc *int) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AV) InitWithRC(rc *int) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		for _, xi := range x.Slice {
			xi.InitWithRC(rc)
		}
		return
	}
	*x.rc += 2
}

func (d *Dict) InitWithRC(rc *int) {
	d.keys.InitWithRC(rc)
	d.values.InitWithRC(rc)
}

func (p *projection) InitWithRC(rc *int) {
	p.Fun.InitWithRC(rc)
	for _, arg := range p.Args {
		arg.InitWithRC(rc)
	}
}

func (p *projectionFirst) InitWithRC(rc *int) {
	p.Fun.InitWithRC(rc)
	p.Arg.InitWithRC(rc)
}

func (p *projectionMonad) InitWithRC(rc *int) {
	p.Fun.InitWithRC(rc)
}

func (r *derivedVerb) InitWithRC(rc *int) {
	r.Arg.InitWithRC(rc)
}

func (r *replacer) InitWithRC(rc *int) {
	if r.oldnew.rc == nil || *r.oldnew.rc <= 1 || r.oldnew.rc == rc {
		r.oldnew.rc = rc
		return
	}
	*r.oldnew.rc += 2
}

func (r *rxReplacer) InitWithRC(rc *int) {
	r.repl.InitWithRC(rc)
}
