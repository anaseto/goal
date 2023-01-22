package goal

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

// IncrRC increments the value reference count (if it has any).
func (x V) DecrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

func (x V) rcdecrRefCounter() {
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

func (x *AB) RC() int32 {
	if x.rc == nil {
		return 0
	}
	return *x.rc
}
func (x *AI) RC() int32 {
	if x.rc == nil {
		return 0
	}
	return *x.rc
}
func (x *AF) RC() int32 {
	if x.rc == nil {
		return 0
	}
	return *x.rc
}
func (x *AS) RC() int32 {
	if x.rc == nil {
		return 0
	}
	return *x.rc
}
func (x *AV) RC() int32 {
	if x.rc == nil {
		return 0
	}
	return *x.rc
}

func reuseRCp(p *int32) bool {
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
	if reuseRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AB{Slice: make([]bool, x.Len())}
}

func (x *AI) reuse() *AI {
	if reuseRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AI{Slice: make([]int64, x.Len())}
}

func (x *AF) reuse() *AF {
	if reuseRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AF{Slice: make([]float64, x.Len())}
}

func (x *AS) reuse() *AS {
	if reuseRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AS{Slice: make([]string, x.Len())}
}

func (x *AV) reuse() *AV {
	if reuseRCp(x.rc) {
		x.flags = flagNone
		return x
	}
	return &AV{Slice: make([]V, x.Len())}
}

// RefCounter is implemented by values that use a reference count. In goal the
// refcount is not used for memory management, but only for optimization of
// memory allocations.  Refcount is increased by each assignement, and each use
// in an operation. It is reduced after each operation, and for each last use
// of a variable (as approximated conservatively). If refcount is equal or less
// than one, then the value is considered reusable.
//
// When defining a new type implementing the Value interface, it is only
// necessary to also implement RefCounter if the type definition contains makes
// use of a type implementing it (for example an array type or a generic V).
type RefCounter interface {
	// IncrRC increments the reference count by one.
	IncrRC()

	// DecrRC decrements the reference count by one.
	DecrRC()

	// InitRC recursively initializes the reference counter for the value
	// (if it needs any). It should not be called on values that already have
	// an RC.
	InitRC(rc *int32)
}

func zeroRCp(p *int32) {
	if p != nil {
		*p = 0
	}
}

func incrRCp(p **int32) {
	if *p == nil {
		var rc int32 = 1
		*p = &rc
		return
	}
	**p++
}

func decrRCp(p *int32) {
	if p != nil && *p > 0 {
		*p--
	}
}

func (x *AB) IncrRC() { incrRCp(&x.rc) }
func (x *AI) IncrRC() { incrRCp(&x.rc) }
func (x *AF) IncrRC() { incrRCp(&x.rc) }
func (x *AS) IncrRC() { incrRCp(&x.rc) }
func (x *AV) IncrRC() {
	if x.rc == nil {
		var rc int32 = 1
		x.InitRC(&rc)
		return
	}
	*x.rc++
}

func (x *AB) DecrRC() { decrRCp(x.rc) }
func (x *AI) DecrRC() { decrRCp(x.rc) }
func (x *AF) DecrRC() { decrRCp(x.rc) }
func (x *AS) DecrRC() { decrRCp(x.rc) }
func (x *AV) DecrRC() { decrRCp(x.rc) }

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

func (x V) InitRC(rc *int32) {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.InitRC()
	}
}

func (e *errV) InitRC(rc *int32) { e.V.InitRC(rc) }

func (x *AB) InitRC(rc *int32) {
	x.rc = rc
}

func (x *AI) InitRC(rc *int32) {
	x.rc = rc
}

func (x *AF) InitRC(rc *int32) {
	x.rc = rc
}

func (x *AS) InitRC(rc *int32) {
	x.rc = rc
}

func (x *AV) InitRC(rc *int32) {
	x.rc = rc
	for _, xi := range x.Slice {
		xi.InitRC(rc)
	}
}

func (p *projection) InitRC(rc *int32) {
	p.Fun.InitRC(rc)
	for _, arg := range p.Args {
		arg.InitRC(rc)
	}
}

func (p *projectionFirst) InitRC(rc *int32) {
	p.Fun.InitRC(rc)
	p.Arg.InitRC(rc)
}

func (p *projectionMonad) InitRC(rc *int32) {
	p.Fun.InitRC(rc)
}

func (r *derivedVerb) InitRC(rc *int32) {
	r.Arg.InitRC(rc)
}

func (r *replacer) InitRC(rc *int32) {
	r.oldnew.InitRC(rc)
}

func (r *rxReplacer) InitRC(rc *int32) {
	r.repl.InitRC(rc)
}
