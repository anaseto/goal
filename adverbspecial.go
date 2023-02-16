package goal

import (
	"math"
	"strings"
)

func fold2vAdd(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vAdd(NewV(xv.values))
	case *AB:
		n := int64(0)
		for _, b := range xv.Slice {
			if b {
				n++
			}
		}
		return NewI(n)
	case *AI:
		n := int64(0)
		for _, xi := range xv.Slice {
			n += xi
		}
		return NewI(n)
	case *AF:
		n := 0.0
		for _, xi := range xv.Slice {
			n += xi
		}
		return NewF(n)
	case *AS:
		if xv.Len() == 0 {
			// NOTE: Maybe not a great idea to have this exception,
			// as empty arrays do not really have a specific type
			// element, but in the event that this code is run,
			// this would probably be the desired result.
			return NewS("")
		}
		n := 0
		for _, s := range xv.Slice {
			n += len(s)
		}
		var sb strings.Builder
		sb.Grow(n)
		for _, s := range xv.Slice {
			sb.WriteString(s)
		}
		return NewS(sb.String())
	case *AV:
		if xv.Len() == 0 {
			return NewI(0)
		}
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = add(r, xi)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		return x
	}
}

func scan2vAdd(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vAdd(NewV(xv.values)))
	case *AB:
		r := make([]int64, xv.Len())
		n := int64(0)
		for i, b := range xv.Slice {
			if b {
				n++
			}
			r[i] = n
		}
		return NewAI(r)
	case *AI:
		r := xv.reuse()
		n := int64(0)
		for i, xi := range xv.Slice {
			n += xi
			r.Slice[i] = n
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		n := 0.0
		for i, xi := range xv.Slice {
			n += xi
			r.Slice[i] = n
		}
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return NewAS(nil)
		}
		n := 0
		for _, s := range xv.Slice {
			n += len(s)
		}
		var sb strings.Builder
		sb.Grow(n)
		for _, s := range xv.Slice {
			sb.WriteString(s)
		}
		rs := sb.String()
		r := xv.reuse()
		n = 0
		for i, s := range xv.Slice {
			n += len(s)
			r.Slice[i] = rs[:n]
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		if xv.Len() == 0 {
			return x
		}
		r.Slice[0] = xv.Slice[0]
		for i, xi := range xv.Slice[1:] {
			last := r.Slice[i]
			last.incrRC2()
			next := add(last, xi)
			last.decrRC2()
			if next.IsPanic() {
				return next
			}
			r.Slice[i+1] = next
		}
		// Will never be canonical, so normalizing is not needed.
		return NewV(r)
	default:
		return x
	}
}

func fold2vMax(x V) V {
	if Length(x) == 0 {
		return NewF(math.Inf(-1))
	}
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMax(NewV(xv.values))
	case *AB:
		for _, b := range xv.Slice {
			if b {
				return NewI(1)
			}
		}
		return NewI(0)
	case *AI:
		return NewI(maxAI(xv))
	case *AF:
		return NewF(maxAF(xv))
	case *AS:
		max := xv.Slice[0]
		for _, s := range xv.Slice[1:] {
			if s > max {
				max = s
			}
		}
		return NewS(max)
	case *AV:
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = maximum(r, xi)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		return x
	}
}

func fold2vMin(x V) V {
	if Length(x) == 0 {
		return NewF(math.Inf(1))
	}
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMin(NewV(xv.values))
	case *AB:
		for _, b := range xv.Slice {
			if !b {
				return NewI(0)
			}
		}
		return NewI(1)
	case *AI:
		return NewI(minAI(xv))
	case *AF:
		return NewF(minAF(xv))
	case *AS:
		min := xv.Slice[0]
		for _, s := range xv.Slice[1:] {
			if s < min {
				min = s
			}
		}
		return NewS(min)
	case *AV:
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = minimum(r, xi)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		return x
	}
}
