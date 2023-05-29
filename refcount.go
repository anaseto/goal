package goal

import "math"

// RefCounter is implemented by values that use a reference count. In goal the
// refcount is not used for memory management, but only for optimization
// purposes.  Refcount is increased by each assignement, and each push
// operation on the stack, except for pushes corresponding to the last use of a
// variable (as approximated conservatively). It is reduced after each drop.
// If refcount is equal or less than one, then the value is considered
// reusable, unless it was marked as immutable.
//
// When defining a new type implementing the Value interface, it is only
// necessary to also implement RefCounter if the type definition makes use of a
// type implementing it (for example an array type or a generic V).
type RefCounter interface {
	Value

	// IncrRC increments the reference count by one.
	IncrRC()

	// DecrRC decrements the reference count by one, or zero if it is
	// already non-positive.
	DecrRC()

	// MarkImmutable marks the value as definitively non-reusable, even if
	// the reference counter is less than one. Extensions might use this
	// function to keep a value around without having to track its
	// reference count anymore.
	MarkImmutable()

	// Clone returns a clone of the value. Note that the cloned value might
	// still share some structures with its parent if they're deemed
	// reusable.
	Clone() Value
}

// HasRC returns true if the value is boxed and implements RefCounter.
func (x V) HasRC() bool {
	if x.kind != valBoxed {
		return false
	}
	_, ok := x.bv.(RefCounter)
	return ok
}

// IncrRC increments the value reference count (if it has any).
func (x V) IncrRC() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.bv.(RefCounter)
	if ok {
		xrc.IncrRC()
	}
}

// incrRC2 increments by 2 the value reference count (if it has any).
func (x V) incrRC2() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.bv.(RefCounter)
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
	xrc, ok := x.bv.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

// decrRC2 decrements by 2 the value reference count (if it has any).
func (x V) decrRC2() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.bv.(RefCounter)
	if ok {
		xrc.DecrRC()
		xrc.DecrRC()
	}
}

func (x V) rcdecrRefCounter() {
	xrc, ok := x.bv.(RefCounter)
	if ok {
		xrc.DecrRC()
	}
}

// Reusable returns true if the array value is reusable.
func (x *A[T]) reusable() bool {
	if x.rc <= 1 && x.flags&flagImmutable == 0 {
		x.rc = 0
		return true
	}
	return false
}

// Reusable returns true if the array value is reusable.
func (x *AB) reusable() bool {
	return (*A[byte])(x).reusable()
}

// Reusable returns true if the array value is reusable.
func (x *AI) reusable() bool {
	return (*A[int64])(x).reusable()
}

// Reusable returns true if the array value is reusable.
func (x *AF) reusable() bool {
	return (*A[float64])(x).reusable()
}

// Reusable returns true if the array value is reusable.
func (x *AS) reusable() bool {
	return (*A[string])(x).reusable()
}

// Reusable returns true if the array value is reusable.
func (x *AV) reusable() bool {
	return (*A[V])(x).reusable()
}

// reuse returns an array of same size that may share memory with the parent
// one if it was reusable. Any flags of the parent are reset.
func (x *A[T]) reuse() *A[T] {
	if x.reusable() {
		x.flags = flagNone
		return x
	}
	return &A[T]{elts: make([]T, len(x.elts))}
}

func (x *AB) reuse() *AB {
	return (*AB)((*A[byte])(x).reuse())
}

func (x *AI) reuse() *AI {
	return (*AI)((*A[int64])(x).reuse())
}

func (x *AF) reuse() *AF {
	return (*AF)((*A[float64])(x).reuse())
}

func (x *AS) reuse() *AS {
	return (*AS)((*A[string])(x).reuse())
}

func (x *AV) reuse() *AV {
	return (*AV)((*A[V])(x).reuse())
}

// IncrRC increments the reference count by one.
func (x *AB) IncrRC() { x.rc++ }

// IncrRC increments the reference count by one.
func (x *AI) IncrRC() { x.rc++ }

// IncrRC increments the reference count by one.
func (x *AF) IncrRC() { x.rc++ }

// IncrRC increments the reference count by one.
func (x *AS) IncrRC() { x.rc++ }

// IncrRC increments the reference count by one.
func (x *AV) IncrRC() { x.rc++ }

// DecrRC decrements the reference count by one, or zero if it is already non
// positive.
func (x *AB) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

// DecrRC decrements the reference count by one, or zero if it is already non
// positive.
func (x *AI) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

// DecrRC decrements the reference count by one, or zero if it is already non
// positive.
func (x *AF) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

// DecrRC decrements the reference count by one, or zero if it is already non
// positive.
func (x *AS) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

// DecrRC decrements the reference count by one, or zero if it is already non
// positive.
func (x *AV) DecrRC() {
	if x.rc > 0 {
		x.rc--
	}
}

// IncrRC increments the reference count of both the key and value arrays by
// one.
func (d *Dict) IncrRC() {
	d.keys.IncrRC()
	d.values.IncrRC()
}

// DecrRC decrements the reference count of both the key and value arrays by
// one, or zero if they are already non positive.
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

// MarkImmutable marks the value as definitively non-reusable.
func (x V) MarkImmutable() {
	if x.kind != valBoxed {
		return
	}
	xrc, ok := x.bv.(RefCounter)
	if ok {
		xrc.MarkImmutable()
	}
}

// MarkImmutable marks the value as definitively non-reusable.
func (e *errV) MarkImmutable() {
	e.V.MarkImmutable()
}

// MarkImmutable marks the value as definitively non-reusable.
func (x *AB) MarkImmutable() {
	x.flags |= flagImmutable
}

// MarkImmutable marks the value as definitively non-reusable.
func (x *AI) MarkImmutable() {
	x.flags |= flagImmutable
}

// MarkImmutable marks the value as definitively non-reusable.
func (x *AF) MarkImmutable() {
	x.flags |= flagImmutable
}

// MarkImmutable marks the value as definitively non-reusable.
func (x *AS) MarkImmutable() {
	x.flags |= flagImmutable
}

// MarkImmutable marks the value as definitively non-reusable.
func (x *AV) MarkImmutable() {
	x.flags |= flagImmutable
}

// MarkImmutable marks the value as definitively non-reusable.
func (d *Dict) MarkImmutable() {
	d.keys.MarkImmutable()
	d.values.MarkImmutable()
}

func (p *projection) MarkImmutable() {
	p.Fun.MarkImmutable()
	for _, arg := range p.Args {
		arg.MarkImmutable()
	}
}

func (p *projectionFirst) MarkImmutable() {
	p.Fun.MarkImmutable()
	p.Arg.MarkImmutable()
}

func (p *projectionMonad) MarkImmutable() {
	p.Fun.MarkImmutable()
}

func (r *derivedVerb) MarkImmutable() {
	r.Arg.MarkImmutable()
}

func (r *replacer) MarkImmutable() {
	r.oldnew.MarkImmutable()
}

func (r *rxReplacer) MarkImmutable() {
	r.repl.MarkImmutable()
}

func refcounts(x V) V {
	if x.kind != valBoxed {
		return NewI(-1)
	}
	switch xv := x.bv.(type) {
	case S:
		return NewI(-1)
	case *errV:
		return refcounts(xv.V)
	case *AB:
		return NewI(int64(xv.rc))
	case *AI:
		return NewI(int64(xv.rc))
	case *AF:
		return NewI(int64(xv.rc))
	case *AS:
		return NewI(int64(xv.rc))
	case *AV:
		return NewI(int64(xv.rc))
	case *Dict:
		return Canonical(NewAV([]V{refcounts(NewV(xv.keys)), refcounts(NewV(xv.values))}))
	default:
		return NewI(math.MinInt64)
	}
}
