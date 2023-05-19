package goal

import (
	"math"
	"strconv"
	"strings"
)

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
		r := make([]byte, xv.Len())
		for i := range r {
			r[i] = 1
		}
		return newABb(r)
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

func fold2Generic(x *AV, f func(V, V) V) V {
	if x.Len() == 0 {
		return NewV(x)
	}
	r := x.At(0)
	for _, xi := range x.elts[1:] {
		r = f(r, xi)
		if r.IsPanic() {
			return r
		}
	}
	return r
}

func fold3Generic(x V, y array, f func(V, V) V) V {
	for i := 0; i < y.Len(); i++ {
		x = f(x, y.at(i))
		if x.IsPanic() {
			return x
		}
	}
	return x
}

func fold2vAdd(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vAdd(NewV(xv.values))
	case *AB:
		return NewI(sumIntegers(xv.elts))
	case *AI:
		return NewI(sumIntegers(xv.elts))
	case *AF:
		return NewF(sumNumbers(0.0, xv.elts))
	case *AS:
		return NewS(concatStrings("", xv.elts))
	case *AV:
		return fold2Generic(xv, add)
	default:
		return x
	}
}

func fold3vAdd(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vAdd(x, NewV(yv.values))
	case *AB:
		if x.IsI() {
			return NewI(sumNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(sumNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, add)
	case *AI:
		if x.IsI() {
			return NewI(sumNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(sumNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, add)
	case *AF:
		if x.IsI() {
			return NewF(sumNumbers(float64(x.I()), yv.elts))
		}
		if x.IsF() {
			return NewF(sumNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, add)
	case *AS:
		if s, ok := x.value.(S); ok {
			return NewS(concatStrings(string(s), yv.elts))
		}
		return fold3Generic(x, yv, add)
	case *AV:
		return fold3Generic(x, yv, add)
	default:
		return add(x, y)
	}
}

func sumNumbers[S number, T number](x S, y []T) S {
	for _, yi := range y {
		x += S(yi)
	}
	return x
}

func concatStrings(x string, y []string) string {
	if len(y) == 0 {
		return x
	}
	n := len(x)
	for _, s := range y {
		n += len(s)
	}
	var sb strings.Builder
	sb.Grow(n)
	sb.WriteString(x)
	for _, s := range y {
		sb.WriteString(s)
	}
	return sb.String()
}

func fold2vSubtract(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vSubtract(NewV(xv.values))
	case *AB:
		if xv.Len() == 0 {
			return NewI(0)
		}
		return NewI(subtractNumbers(int64(xv.elts[0]), xv.elts[1:]))
	case *AI:
		if xv.Len() == 0 {
			return NewI(0)
		}
		return NewI(subtractNumbers(xv.elts[0], xv.elts[1:]))
	case *AF:
		if xv.Len() == 0 {
			return NewF(0)
		}
		return NewF(subtractNumbers(xv.elts[0], xv.elts[1:]))
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		return NewS(trimSuffixs(xv.elts[0], xv.elts[1:]))
	case *AV:
		return fold2Generic(xv, subtract)
	default:
		return x
	}
}

func fold3vSubtract(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vSubtract(x, NewV(yv.values))
	case *AB:
		if x.IsI() {
			return NewI(subtractNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(subtractNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, subtract)
	case *AI:
		if x.IsI() {
			return NewI(subtractNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(subtractNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, subtract)
	case *AF:
		if x.IsI() {
			return NewF(subtractNumbers(float64(x.I()), yv.elts))
		}
		if x.IsF() {
			return NewF(subtractNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, subtract)
	case *AS:
		if s, ok := x.value.(S); ok {
			return NewS(trimSuffixs(string(s), yv.elts))
		}
		return fold3Generic(x, yv, subtract)
	case *AV:
		return fold3Generic(x, yv, subtract)
	default:
		return subtract(x, y)
	}
}

func subtractNumbers[S number, T number](x S, y []T) S {
	for _, yi := range y {
		x -= S(yi)
	}
	return x
}

func trimSuffixs(x string, y []string) string {
	for _, yi := range y {
		x = strings.TrimSuffix(x, yi)
	}
	return x
}

func fold2vMultiply(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMultiply(NewV(xv.values))
	case *AB:
		return NewI(multiplyNumbers(int64(1), xv.elts))
	case *AI:
		return NewI(multiplyNumbers(int64(1), xv.elts))
	case *AF:
		return NewF(multiplyNumbers(1.0, xv.elts))
	case *AS:
		return panicType("*/x", "x", x)
	case *AV:
		return fold2Generic(xv, multiply)
	default:
		return x
	}
}

func fold3vMultiply(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vMultiply(x, NewV(yv.values))
	case *AB:
		if x.IsI() {
			return NewI(multiplyNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(multiplyNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, multiply)
	case *AI:
		if x.IsI() {
			return NewI(multiplyNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(multiplyNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, multiply)
	case *AF:
		if x.IsI() {
			return NewF(multiplyNumbers(float64(x.I()), yv.elts))
		}
		if x.IsF() {
			return NewF(multiplyNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, multiply)
	case *AS:
		return fold3Generic(x, yv, multiply)
	case *AV:
		return fold3Generic(x, yv, multiply)
	default:
		return multiply(x, y)
	}
}

func multiplyNumbers[T number, S number](x S, y []T) S {
	for _, yi := range y {
		x *= S(yi)
	}
	return x
}

func fold2vMax(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMax(NewV(xv.values))
	case *AB:
		if xv.Len() == 0 {
			return NewI(math.MinInt64)
		}
		if xv.IsBoolean() {
			return NewI(maxBools(xv.elts))
		}
		return NewI(maxIntegers(xv.elts))
	case *AI:
		return NewI(maxIntegers(xv.elts))
	case *AF:
		return NewF(maxNumbers(math.Inf(-1), xv.elts))
	case *AS:
		return NewS(maxStrings("", xv.elts))
	case *AV:
		return fold2Generic(xv, maximum)
	default:
		return x
	}
}

func fold3vMax(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vMax(x, NewV(yv.values))
	case *AB:
		if x.IsI() {
			return NewI(maxNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(maxNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, maximum)
	case *AI:
		if x.IsI() {
			return NewI(maxNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(maxNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, maximum)
	case *AF:
		if x.IsI() {
			return NewF(maxNumbers(float64(x.I()), yv.elts))
		}
		if x.IsF() {
			return NewF(maxNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, maximum)
	case *AS:
		if s, ok := x.value.(S); ok {
			return NewS(maxStrings(string(s), yv.elts))
		}
		return fold3Generic(x, yv, maximum)
	case *AV:
		return fold3Generic(x, yv, maximum)
	default:
		return maximum(x, y)
	}
}

func maxBools(x []byte) int64 {
	var max byte
	for _, xi := range x {
		max |= xi
	}
	return int64(max)
}

func maxNumbers[S number, T number](x S, y []T) S {
	for _, yi := range y {
		if S(yi) > x {
			x = S(yi)
		}
	}
	return x
}

func maxStrings(x string, y []string) string {
	for _, yi := range y {
		if yi > x {
			x = yi
		}
	}
	return x
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

func fold2vMin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vMin(NewV(xv.values))
	case *AB:
		if xv.Len() == 0 {
			return NewI(math.MaxInt64)
		}
		if xv.IsBoolean() {
			return NewI(minBools(xv.elts))
		}
		return NewI(minIntegers(xv.elts))
	case *AI:
		return NewI(minIntegers(xv.elts))
	case *AF:
		return NewF(minNumbers(math.Inf(1), xv.elts))
	case *AS:
		if xv.Len() == 0 {
			return NewS("")
		}
		return NewS(minStrings(xv.elts[0], xv.elts[1:]))
	case *AV:
		return fold2Generic(xv, minimum)
	default:
		return x
	}
}

func fold3vMin(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vMin(x, NewV(yv.values))
	case *AB:
		if x.IsI() {
			return NewI(minNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(minNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, minimum)
	case *AI:
		if x.IsI() {
			return NewI(minNumbers(x.I(), yv.elts))
		}
		if x.IsF() {
			return NewF(minNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, minimum)
	case *AF:
		if x.IsI() {
			return NewF(minNumbers(float64(x.I()), yv.elts))
		}
		if x.IsF() {
			return NewF(minNumbers(x.F(), yv.elts))
		}
		return fold3Generic(x, yv, minimum)
	case *AS:
		if s, ok := x.value.(S); ok {
			return NewS(minStrings(string(s), yv.elts))
		}
		return fold3Generic(x, yv, minimum)
	case *AV:
		return fold3Generic(x, yv, minimum)
	default:
		return minimum(x, y)
	}
}

func minBools(x []byte) int64 {
	var min byte = 1
	for _, xi := range x {
		min &= xi
	}
	return int64(min)
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

func minStrings(x string, y []string) string {
	for _, yi := range y {
		if yi < x {
			x = yi
		}
	}
	return x
}

func minNumbers[S number, T number](x S, y []T) S {
	for _, yi := range y {
		if S(yi) < x {
			x = S(yi)
		}
	}
	return x
}

func fold2vJoin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return fold2vJoin(NewV(xv.values))
	case *AV:
		if isFlat(xv.elts) {
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

func fold3vJoin(x, y V) V {
	switch yv := y.value.(type) {
	case *Dict:
		return fold3vJoin(x, NewV(yv.values))
	case *AV:
		for _, yi := range yv.elts {
			x = joinTo(x, yi) // does not panic
		}
		return x
	default:
		return joinTo(x, y)
	}
}

func convergeJoin(x V) V {
	switch xv := x.value.(type) {
	case *Dict:
		return convergeJoin(NewV(xv.values))
	case *AV:
		if isFlat(xv.elts) {
			return x
		}
		r := xv.elts[0]
		for _, xi := range xv.elts[1:] {
			r = joinTo(r, xi) // does not panic
		}
		return convergeJoin(r)
	default:
		return x
	}
}

func scanGeneric(x *AV, f func(V, V) V) V {
	if x.Len() == 0 {
		return NewV(x)
	}
	r := x.reuse()
	r.elts[0] = x.elts[0]
	for i, xi := range x.elts[1:] {
		last := r.elts[i]
		last.incrRC2()
		next := f(last, xi)
		next.InitRC()
		last.decrRC2()
		if next.IsPanic() {
			return next
		}
		r.elts[i+1] = next
	}
	// Will never be canonical, so normalizing is not needed.
	return NewV(r)
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
			r.flags |= flagAscending
			return NewV(r)
		}
		r := make([]int64, xv.Len())
		var n int64
		for i, xi := range xv.elts {
			n += int64(xi)
			r[i] = n
		}
		return NewV(&AI{elts: r, flags: flagAscending})
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
		return scanGeneric(xv, add)
	default:
		return x
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
		r.flags |= flagAscending
		return NewV(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[int64](r.elts, xv.elts)
		r.flags |= flagAscending
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[float64](r.elts, xv.elts)
		r.flags |= flagAscending
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		scan2vMaxSlice[string](r.elts, xv.elts)
		r.flags |= flagAscending
		return NewV(r)
	case *AV:
		return scanGeneric(xv, maximum)
	default:
		return x
	}
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
		return scanGeneric(xv, minimum)
	default:
		return x
	}
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
