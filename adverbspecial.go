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
		return NewI(sumIntegers(xv.elts))
	case *AI:
		return NewI(sumIntegers(xv.elts))
	case *AF:
		n := 0.0
		for _, xi := range xv.elts {
			n += xi
		}
		return NewF(n)
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		n := 0
		for _, s := range xv.elts {
			n += len(s)
		}
		var sb strings.Builder
		sb.Grow(n)
		for _, s := range xv.elts {
			sb.WriteString(s)
		}
		return NewS(sb.String())
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.At(0)
		for _, xi := range xv.elts[1:] {
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

func sumIntegers[T integer](x []T) int64 {
	var n int64
	for _, xi := range x {
		n += int64(xi)
	}
	return n
}

func scan2vAdd(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vAdd(NewV(xv.values)))
	case *AB:
		if xv.IsBoolean() && xv.Len() < 256 {
			r := xv.reuse()
			var n byte
			for i, xi := range xv.elts {
				n += xi
				r.elts[i] = n
			}
			return NewV(r)
		} else {
			r := make([]int64, xv.Len())
			var n int64
			for i, xi := range xv.elts {
				n += int64(xi)
				r[i] = n
			}
			return NewAI(r)
		}
	case *AI:
		r := xv.reuse()
		var n int64
		for i, xi := range xv.elts {
			n += xi
			r.elts[i] = n
		}
		return NewV(r)
	case *AF:
		r := xv.reuse()
		n := 0.0
		for i, xi := range xv.elts {
			n += xi
			r.elts[i] = n
		}
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return NewAS(nil)
		}
		n := 0
		for _, s := range xv.elts {
			n += len(s)
		}
		var sb strings.Builder
		sb.Grow(n)
		for _, s := range xv.elts {
			sb.WriteString(s)
		}
		rs := sb.String()
		r := xv.reuse()
		n = 0
		for i, s := range xv.elts {
			n += len(s)
			r.elts[i] = rs[:n]
		}
		return NewV(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		for i, xi := range xv.elts[1:] {
			last := r.elts[i]
			last.incrRC2()
			next := add(last, xi)
			next.InitRC()
			last.decrRC2()
			if next.IsPanic() {
				return next
			}
			r.elts[i+1] = next
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
		n := int64(xv.elts[0])
		for _, xi := range xv.elts[1:] {
			n -= int64(xi)
		}
		return NewI(n)
	case *AI:
		if xv.Len() == 0 {
			return NewI(0)
		}
		var n int64 = xv.elts[0]
		for _, xi := range xv.elts[1:] {
			n -= xi
		}
		return NewI(n)
	case *AF:
		if xv.Len() == 0 {
			return NewI(0)
		}
		var n float64 = xv.elts[0]
		for _, xi := range xv.elts[1:] {
			n -= xi
		}
		return NewF(n)
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		var r string = xv.elts[0]
		for _, xi := range xv.elts[1:] {
			r = strings.TrimSuffix(r, xi)
		}
		return NewS(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.At(0)
		for _, xi := range xv.elts[1:] {
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
		return NewI(multiplyIntegers(xv.elts))
	case *AI:
		return NewI(multiplyIntegers(xv.elts))
	case *AF:
		var n float64 = 1.0
		for _, xi := range xv.elts {
			n *= xi
		}
		return NewF(n)
	case *AS:
		return panics("*/x : bad type in x (S)")
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.At(0)
		for _, xi := range xv.elts[1:] {
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

func multiplyIntegers[T integer](x []T) int64 {
	var n int64 = 1
	for _, xi := range x {
		n *= int64(xi)
	}
	return n
}

func fold2vMax(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMax(NewV(xv.values))
	case *AB:
		return NewI(maxIntegers(xv.elts))
	case *AI:
		return NewI(maxIntegers(xv.elts))
	case *AF:
		if xv.Len() == 0 {
			return NewF(math.Inf(-1))
		}
		return NewF(maxFloat64s(xv.elts))
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		max := xv.elts[0]
		for _, s := range xv.elts[1:] {
			if s > max {
				max = s
			}
		}
		return NewS(max)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.At(0)
		for _, xi := range xv.elts[1:] {
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

func maxIntegers[I integer](x []I) int64 {
	var max int64 = math.MinInt64
	for _, xi := range x {
		if int64(xi) > max {
			max = int64(xi)
		}
	}
	return max
}

func maxFloat64s(x []float64) float64 {
	max := math.Inf(-1)
	for _, xi := range x {
		// NOTE: not equivalent to math.Max(xi, max) if there are NaNs,
		// but faster, so keep it this way.
		if xi > max {
			max = xi
		}
	}
	return max
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
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[byte](r.elts, xv.elts)
		return NewV(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[int64](r.elts, xv.elts)
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[float64](r.elts, xv.elts)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[string](r.elts, xv.elts)
		return NewV(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		for i, xi := range xv.elts[1:] {
			last := r.elts[i]
			last.incrRC2()
			next := maximum(last, xi)
			next.InitRC()
			last.decrRC2()
			if next.IsPanic() {
				return next
			}
			r.elts[i+1] = next
		}
		// Will never be canonical, so normalizing is not needed.
		return NewV(r)
	default:
		return x
	}
}

func fold2vMin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMin(NewV(xv.values))
	case *AB:
		return NewI(minIntegers(xv.elts))
	case *AI:
		return NewI(minIntegers(xv.elts))
	case *AF:
		if x.Len() == 0 {
			return NewF(math.Inf(1))
		}
		return NewF(minFloat64s(xv.elts))
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		min := xv.elts[0]
		for _, s := range xv.elts[1:] {
			if s < min {
				min = s
			}
		}
		return NewS(min)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.At(0)
		for _, xi := range xv.elts[1:] {
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

func minIntegers[T integer](x []T) int64 {
	var min int64 = math.MaxInt64
	for _, xi := range x {
		if int64(xi) < min {
			min = int64(xi)
		}
	}
	return min
}

func minFloat64s(x []float64) float64 {
	min := math.Inf(1)
	for _, xi := range x {
		// NOTE: not equivalent to math.Min(xi, min) if there are NaNs,
		// but faster, so keep it this way.
		if xi < min {
			min = xi
		}
	}
	return min
}

func scan2vMinSlice[T ordered](dst, xs []T) {
	min := xs[0]
	dst[0] = min
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
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[byte](r.elts, xv.elts)
		return NewV(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[int64](r.elts, xv.elts)
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[float64](r.elts, xv.elts)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMinSlice[string](r.elts, xv.elts)
		return NewV(r)
	case *AV:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		for i, xi := range xv.elts[1:] {
			last := r.elts[i]
			last.incrRC2()
			next := minimum(last, xi)
			next.InitRC()
			last.decrRC2()
			if next.IsPanic() {
				return next
			}
			r.elts[i+1] = next
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
		r := xv.elts[0]
		for _, xi := range xv.elts[1:] {
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
		return NewAS(stringIntegers(xv.elts))
	case *AI:
		return NewAS(stringIntegers(xv.elts))
	case *AF:
		return NewAS(stringFloat64s(xv.elts, ctx.Prec))
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = strconv.Quote(xi)
		}
		return NewV(r)
	case *AV:
		return NewAS(stringVs(xv.elts, ctx))
	default:
		panic("each2String")
	}
}

func stringIntegers[T integer](x []T) []string {
	r := make([]string, len(x))
	for i, xi := range x {
		r[i] = strconv.FormatInt(int64(xi), 10)
	}
	return r
}

func stringFloat64s(x []float64, prec int) []string {
	r := make([]string, len(x))
	for i, xi := range x {
		r[i] = strconv.FormatFloat(xi, 'g', prec, 64)
	}
	return r
}

func stringVs(x []V, ctx *Context) []string {
	r := make([]string, len(x))
	for i, xi := range x {
		r[i] = xi.Sprint(ctx)
	}
	return r
}

func each2First(x array) V {
	switch xv := x.(type) {
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = first(xi)
		}
		return Canonical(NewAV(r))
	default:
		return NewV(x)
	}
}

func each2Length(x array) V {
	switch xv := x.(type) {
	case *AV:
		r := make([]int64, xv.Len())
		for i, xi := range xv.elts {
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

func each2Type(x array) V {
	switch xv := x.(type) {
	case *AF:
		r := make([]string, x.Len())
		for i := range r {
			r[i] = "n"
		}
		return NewAS(r)
	case *AS:
		r := xv.reuse()
		for i := range r.elts {
			r.elts[i] = "s"
		}
		return NewV(r)
	case *AV:
		r := make([]string, xv.Len())
		for i, xi := range xv.elts {
			r[i] = xi.Type()
		}
		return NewAS(r)
	default:
		r := make([]string, x.Len())
		for i := range r {
			r[i] = "i"
		}
		return NewAS(r)
	}
}
