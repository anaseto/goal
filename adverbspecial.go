package goal

import (
	"math"
	"strconv"
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
			next.InitRC()
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

func fold2vSubtract(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vSubtract(NewV(xv.values))
	case *AB:
		if xv.Len() == 0 {
			return NewI(0)
		}
		var n int64
		if xv.Slice[0] {
			n++
		}
		for _, b := range xv.Slice[1:] {
			if b {
				n--
			}
		}
		return NewI(n)
	case *AI:
		if xv.Len() == 0 {
			return NewI(0)
		}
		var n int64 = xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			n -= xi
		}
		return NewI(n)
	case *AF:
		if xv.Len() == 0 {
			return NewI(0)
		}
		var n float64 = xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			n -= xi
		}
		return NewF(n)
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		var r string = xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			r = strings.TrimSuffix(r, xi)
		}
		return NewS(r)
	case *AV:
		if xv.Len() == 0 {
			return NewI(0)
		}
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = subtract(r, xi)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		return x
	}
}

func fold2vMultiply(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMultiply(NewV(xv.values))
	case *AB:
		for _, b := range xv.Slice {
			if !b {
				return NewI(0)
			}
		}
		return NewI(1)
	case *AI:
		if xv.Len() == 0 {
			return NewI(1)
		}
		var n int64 = xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			n *= xi
		}
		return NewI(n)
	case *AF:
		if xv.Len() == 0 {
			return NewI(1)
		}
		var n float64 = xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			n *= xi
		}
		return NewF(n)
	case *AS:
		return panics("*/x : bad type in x (S)")
	case *AV:
		if xv.Len() == 0 {
			return NewI(1)
		}
		r := xv.At(0)
		for _, xi := range xv.Slice[1:] {
			r = multiply(r, xi)
			if r.IsPanic() {
				return r
			}
		}
		return r
	default:
		return x
	}
}

func fold2vMax(x V) V {
	if x.Len() == 0 {
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

type ordered interface {
	float64 | int64 | string
}

func scan2vMaxSlice[T ordered](dst, xs []T) {
	max := xs[0]
	dst[0] = max
	for i, xi := range xs[1:] {
		if xi > max {
			max = xi
		}
		dst[i+1] = max
	}
}

func scan2vMax(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vMax(NewV(xv.values)))
	case *AB:
		r := xv.reuse()
		for i, b := range xv.Slice {
			if b {
				for j := i; j < xv.Len(); j++ {
					r.Slice[j] = true
				}
				break
			}
		}
		return NewV(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[int64](r.Slice, xv.Slice)
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[float64](r.Slice, xv.Slice)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[string](r.Slice, xv.Slice)
		return NewV(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.Slice[0] = xv.Slice[0]
		for i, xi := range xv.Slice[1:] {
			last := r.Slice[i]
			last.incrRC2()
			next := maximum(last, xi)
			next.InitRC()
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

func fold2vMin(x V) V {
	if x.Len() == 0 {
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

func scan2vMinSlice[T ordered](dst, xs []T) {
	min := xs[0]
	for i, xi := range xs[1:] {
		if xi < min {
			min = xi
		}
		dst[i+1] = min
	}
}

func scan2vMin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vMin(NewV(xv.values)))
	case *AB:
		r := xv.reuse()
		for i, b := range xv.Slice {
			if !b {
				for j := i; j < len(r.Slice); j++ {
					r.Slice[j] = false
				}
				break
			}
			r.Slice[i] = true
		}
		return NewV(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[int64](r.Slice, xv.Slice)
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[float64](r.Slice, xv.Slice)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[string](r.Slice, xv.Slice)
		return NewV(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.Slice[0] = xv.Slice[0]
		for i, xi := range xv.Slice[1:] {
			last := r.Slice[i]
			last.incrRC2()
			next := minimum(last, xi)
			next.InitRC()
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

func fold2vJoin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vJoin(NewV(xv.values))
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.Slice[0]
		for _, xi := range xv.Slice[1:] {
			r = joinTo(r, xi) // does not panic
		}
		return r
	default:
		return x
	}
}

func each2String(ctx *Context, x array) V {
	switch xv := x.(type) {
	case *AB:
		r := make([]string, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = strconv.FormatInt(B2I(xi), 10)
		}
		return NewAS(r)
	case *AI:
		r := make([]string, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = strconv.FormatInt(xi, 10)
		}
		return NewAS(r)
	case *AF:
		r := make([]string, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = strconv.FormatFloat(xi, 'g', ctx.prec, 64)
		}
		return NewAS(r)
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = strconv.Quote(xi)
		}
		return NewV(r)
	case *AV:
		r := make([]string, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = xi.Sprint(ctx)
		}
		return NewAS(r)
	default:
		panic("each2String")
	}
}

func each2First(ctx *Context, x array) V {
	switch xv := x.(type) {
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = first(xi)
		}
		return Canonical(NewAV(r))
	default:
		return NewV(x)
	}
}

func each2Length(ctx *Context, x array) V {
	switch xv := x.(type) {
	case *AV:
		r := make([]int64, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = int64(xi.Len())
		}
		return NewAI(r)
	default:
		r := make([]int64, xv.Len())
		for i := range r {
			r[i] = 1
		}
		return NewAI(r)
	}
}

func each2Type(ctx *Context, x array) V {
	switch xv := x.(type) {
	case *AS:
		r := xv.reuse()
		for i := range r.Slice {
			r.Slice[i] = "s"
		}
		return NewV(r)
	case *AV:
		r := make([]string, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = xi.Type()
		}
		return NewAS(r)
	default:
		r := make([]string, x.Len())
		for i := range r {
			r[i] = "n"
		}
		return NewAS(r)
	}
}
