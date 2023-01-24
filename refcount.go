package goal

import (
	"fmt"
	"log"
)

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

func (x *AB) RC() *int32 { return x.rc }
func (x *AI) RC() *int32 { return x.rc }
func (x *AF) RC() *int32 { return x.rc }
func (x *AS) RC() *int32 { return x.rc }
func (x *AV) RC() *int32 { return x.rc }

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
		x.rc = nil // NOTE: not always necessary, maybe use two functions
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
	Value

	// IncrRC increments the reference count by one.
	IncrRC()

	// DecrRC decrements the reference count by one.
	DecrRC()

	// InitWithRC recursively sets the refcount pointer for reusable
	// values, and increments by 2 the refcount of non-reusable values (to
	// ensure immutability).
	InitWithRC(rc *int32)
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

// wellformedRC checks that RCs of the value are properly shared among
// subarrays. It is for testing purposes.
func wellformedRC(x V) bool {
	switch xv := x.value.(type) {
	case *AV:
		return sharesRC(x, xv.rc)
	default:
		return true
	}
}

func getRC(rc *int32) int32 {
	if rc != nil {
		return *rc
	}
	return 0
}

func sharesRC(x V, rc *int32) bool {
	switch xv := x.value.(type) {
	case *AV:
		if xv.RC() != rc {
			log.Printf("%p vs %p (%d vs %d)", xv.RC(), rc, getRC(xv.RC()), getRC(rc))
			// actually, %d > %d is ok
			return false
		}
		for _, xi := range xv.Slice {
			if !sharesRC(xi, rc) {
				return false
			}
		}
		return true
	case array:
		if xv.RC() != rc {
			log.Printf("%p vs %p (%d vs %d)", xv.RC(), rc, getRC(xv.RC()), getRC(rc))
			return false
		}
		return true
	default:
		return true
	}
}

func (ctx *Context) assertWellformedRC(x V) {
	if !wellformedRC(x) {
		panic(fmt.Sprintf("unshared rc: %s (%s)", x.Sprint(ctx), x.Type()))
	}
}

// InitRC initializes refcount if necessary.
func (x V) InitRC() {
	if x.kind != valBoxed {
		return
	}
	xa, ok := x.value.(array)
	if ok && xa.RC() == nil {
		var n int32
		xa.InitWithRC(&n)
	}
}

func (x V) InitWithRC(rc *int32) {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.value.(RefCounter)
	if ok {
		xrc.InitWithRC(rc)
	}
}

func (s S) InitWithRC(rc *int32) {}

func (e panicV) InitWithRC(rc *int32) {}

func (e *errV) InitWithRC(rc *int32) {
	e.V.InitWithRC(rc)
}

func (x *AB) InitWithRC(rc *int32) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AI) InitWithRC(rc *int32) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AF) InitWithRC(rc *int32) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AS) InitWithRC(rc *int32) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		return
	}
	*x.rc += 2
}

func (x *AV) InitWithRC(rc *int32) {
	if x.rc == nil || *x.rc <= 1 || x.rc == rc {
		x.rc = rc
		for _, xi := range x.Slice {
			xi.InitWithRC(rc)
		}
		return
	}
	*x.rc += 2
}

func (p *projection) InitWithRC(rc *int32) {
	p.Fun.InitWithRC(rc)
	for _, arg := range p.Args {
		arg.InitWithRC(rc)
	}
}

func (p *projectionFirst) InitWithRC(rc *int32) {
	p.Fun.InitWithRC(rc)
	p.Arg.InitWithRC(rc)
}

func (p *projectionMonad) InitWithRC(rc *int32) {
	p.Fun.InitWithRC(rc)
}

func (r *derivedVerb) InitWithRC(rc *int32) {
	r.Arg.InitWithRC(rc)
}

func (r *nReplacer) InitWithRC(rc *int32) {}

func (r *replacer) InitWithRC(rc *int32) {
	if r.oldnew.rc == nil || *r.oldnew.rc <= 1 || r.oldnew.rc == rc {
		r.oldnew.rc = rc
		return
	}
	*r.oldnew.rc += 2
}

func (r *rx) InitWithRC(rc *int32) {}

func (r *rxReplacer) InitWithRC(rc *int32) {
	r.repl.InitWithRC(rc)
}

func (x *AV) initRC() {
	if x.rc == nil {
		var n int32
		x.rc = &n
	}
	for _, xi := range x.Slice {
		xi.InitWithRC(x.rc)
	}
}
