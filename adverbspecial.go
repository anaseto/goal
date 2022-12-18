package goal

import (
	"math"
	"strings"
)

func fold2vAdd(x V) V {
	switch xv := x.value.(type) {
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
		var b strings.Builder
		b.Grow(n)
		for _, s := range xv.Slice {
			b.WriteString(s)
		}
		return NewS(b.String())
	case *AV:
		if xv.Len() == 0 {
			return NewI(0)
		}
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = add(r, xi)
		}
		return r
	default:
		return x
	}
}

func fold2vMax(x V) V {
	if Length(x) == 0 {
		return NewI(math.MinInt64)
	}
	switch xv := x.value.(type) {
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
		}
		return r
	default:
		return x
	}
}

func fold2vMin(x V) V {
	if Length(x) == 0 {
		return NewI(math.MaxInt64)
	}
	switch xv := x.value.(type) {
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
		}
		return r
	default:
		return x
	}
}
