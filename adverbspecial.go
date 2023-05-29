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
			ri := first(xi)
			ri.MarkImmutable()
			r[i] = ri
		}
		return canonicalVs(r)
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
	switch xv := x.bv.(type) {
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
	switch yv := y.bv.(type) {
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
		if s, ok := x.bv.(S); ok {
			return NewS(concatStrings(string(s), yv.elts))
		}
		return fold3Generic(x, yv, add)
	case *AV:
		return fold3Generic(x, yv, add)
	default:
		return add(x, y)
	}
}

func sumNumbers[T number, U number](x U, y []T) U {
	for _, yi := range y {
		x += U(yi)
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
	switch xv := x.bv.(type) {
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
		return NewS(trimSuffixes(xv.elts[0], xv.elts[1:]))
	case *AV:
		return fold2Generic(xv, subtract)
	default:
		return x
	}
}

func fold3vSubtract(x, y V) V {
	switch yv := y.bv.(type) {
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
		if s, ok := x.bv.(S); ok {
			return NewS(trimSuffixes(string(s), yv.elts))
		}
		return fold3Generic(x, yv, subtract)
	case *AV:
		return fold3Generic(x, yv, subtract)
	default:
		return subtract(x, y)
	}
}

func subtractNumbers[T number, U number](x U, y []T) U {
	for _, yi := range y {
		x -= U(yi)
	}
	return x
}

func trimSuffixes(x string, y []string) string {
	for _, yi := range y {
		x = strings.TrimSuffix(x, yi)
	}
	return x
}

func fold2vMultiply(x V) V {
	switch xv := x.bv.(type) {
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
	switch yv := y.bv.(type) {
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

func multiplyNumbers[T number, U number](x U, y []T) U {
	for _, yi := range y {
		x *= U(yi)
	}
	return x
}

func fold2vMax(x V) V {
	switch xv := x.bv.(type) {
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
	switch yv := y.bv.(type) {
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
		if s, ok := x.bv.(S); ok {
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

func maxNumbers[T number, U number](x U, y []T) U {
	for _, yi := range y {
		if U(yi) > x {
			x = U(yi)
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
	switch xv := x.bv.(type) {
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
	switch yv := y.bv.(type) {
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
		if s, ok := x.bv.(S); ok {
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

func minNumbers[T number, U number](x U, y []T) U {
	for _, yi := range y {
		if U(yi) < x {
			x = U(yi)
		}
	}
	return x
}

func fold2vJoin(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return fold2vJoin(NewV(xv.values))
	case *AV:
		if isFlat(xv.elts) {
			return x
		}
		r := xv.elts[0]
		for _, xi := range xv.elts[1:] {
			r = join(r, xi) // does not panic
		}
		return r
	default:
		return x
	}
}

func fold3vJoin(x, y V) V {
	switch yv := y.bv.(type) {
	case *Dict:
		return fold3vJoin(x, NewV(yv.values))
	case *AV:
		for _, yi := range yv.elts {
			x = join(x, yi) // does not panic
		}
		return x
	default:
		return join(x, y)
	}
}

func convergeJoin(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return convergeJoin(NewV(xv.values))
	case *AV:
		if isFlat(xv.elts) {
			return x
		}
		r := xv.elts[0]
		for _, xi := range xv.elts[1:] {
			r = join(r, xi) // does not panic
		}
		return convergeJoin(r)
	default:
		return x
	}
}

func scan2Generic(x *AV, f func(V, V) V) V {
	if x.Len() == 0 {
		return NewV(x)
	}
	r := x.reuse()
	r.elts[0] = x.elts[0]
	for i, xi := range x.elts[1:] {
		next := f(r.elts[i], xi)
		if next.IsPanic() {
			return next
		}
		next.MarkImmutable()
		r.elts[i+1] = next
	}
	// Will never be canonical (for currently used fs), so normalizing is
	// not needed.
	return NewV(r)
}

func scan3Generic(x V, y array, f func(V, V) V) V {
	if y.Len() == 0 {
		return NewV(y)
	}
	r := make([]V, y.Len())
	for i := 0; i < y.Len(); i++ {
		x = f(x, y.at(i))
		if x.IsPanic() {
			return x
		}
		x.MarkImmutable()
		r[i] = x
	}
	// Will never be canonical (for currently used fs), so normalizing is
	// not needed.
	return newAVu(r)
}

func scan2vAdd(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vAdd(NewV(xv.values)))
	case *AB:
		if xv.IsBoolean() && xv.Len() < 256 {
			r := xv.reuse()
			scanSumNumbers(r.elts, 0, xv.elts)
			r.flags |= flagAscending
			return NewV(r)
		}
		r := make([]int64, xv.Len())
		scanSumNumbers(r, 0, xv.elts)
		return NewV(&AI{elts: r, flags: flagAscending})
	case *AI:
		r := xv.reuse()
		scanSumNumbers(r.elts, 0, xv.elts)
		return NewV(r)
	case *AF:
		r := xv.reuse()
		scanSumNumbers(r.elts, 0.0, xv.elts)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return NewAS(nil)
		}
		return scanConcatStrings("", xv)
	case *AV:
		return scan2Generic(xv, add)
	default:
		return x
	}
}

func scan3vAdd(x, y V) V {
	switch yv := y.bv.(type) {
	case *Dict:
		return newDictValues(yv.keys, scan3vAdd(x, NewV(yv.values)))
	case *AB:
		if x.IsI() {
			r := make([]int64, yv.Len())
			scanSumNumbers(r, x.I(), yv.elts)
			return NewV(&AI{elts: r, flags: flagAscending})
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSumNumbers(r, x.F(), yv.elts)
			return NewV(&AF{elts: r, flags: flagAscending})
		}
		return scan3Generic(x, yv, add)
	case *AI:
		if x.IsI() {
			r := yv.reuse()
			scanSumNumbers(r.elts, x.I(), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSumNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, add)
	case *AF:
		if x.IsI() {
			r := yv.reuse()
			scanSumNumbers(r.elts, float64(x.I()), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSumNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, add)
	case *AS:
		if s, ok := x.bv.(S); ok {
			return scanConcatStrings(string(s), yv)
		}
		return scan3Generic(x, yv, add)
	case *AV:
		return scan3Generic(x, yv, add)
	default:
		return add(x, y)
	}
}

func scanSumNumbers[T number, U number](r []U, x U, y []T) {
	for i, yi := range y {
		x += U(yi)
		r[i] = x
	}
}

func scanConcatStrings(x string, y *AS) V {
	n := len(x)
	for _, s := range y.elts {
		n += len(s)
	}
	var sb strings.Builder
	sb.Grow(n)
	sb.WriteString(x)
	for _, s := range y.elts {
		sb.WriteString(s)
	}
	rs := sb.String()
	r := y.reuse()
	n = len(x)
	for i, s := range y.elts {
		n += len(s)
		r.elts[i] = rs[:n]
	}
	return NewV(r)
}

func scan2vSubtract(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vSubtract(NewV(xv.values)))
	case *AB:
		if xv.Len() == 0 {
			return x
		}
		r := make([]int64, xv.Len())
		r[0] = int64(xv.elts[0])
		scanSubtractNumbers(r[1:], int64(xv.elts[0]), xv.elts[1:])
		return NewAI(r)
	case *AI:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		scanSubtractNumbers(r.elts[1:], xv.elts[0], xv.elts[1:])
		return NewV(r)
	case *AF:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		scanSubtractNumbers(r.elts[1:], xv.elts[0], xv.elts[1:])
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		scanTrimSuffixes(r.elts[1:], xv.elts[0], xv.elts[1:])
		return NewV(r)
	case *AV:
		return scan2Generic(xv, subtract)
	default:
		return x
	}
}

func scan3vSubtract(x, y V) V {
	switch yv := y.bv.(type) {
	case *Dict:
		return newDictValues(yv.keys, scan3vSubtract(x, NewV(yv.values)))
	case *AB:
		if x.IsI() {
			r := make([]int64, yv.Len())
			scanSubtractNumbers(r, x.I(), yv.elts)
			return NewV(&AI{elts: r, flags: flagAscending})
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSubtractNumbers(r, x.F(), yv.elts)
			return NewV(&AF{elts: r, flags: flagAscending})
		}
		return scan3Generic(x, yv, subtract)
	case *AI:
		if x.IsI() {
			r := yv.reuse()
			scanSubtractNumbers(r.elts, x.I(), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSubtractNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, subtract)
	case *AF:
		if x.IsI() {
			r := yv.reuse()
			scanSubtractNumbers(r.elts, float64(x.I()), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanSubtractNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, subtract)
	case *AS:
		if s, ok := x.bv.(S); ok {
			r := yv.reuse()
			scanTrimSuffixes(r.elts, string(s), yv.elts)
			return NewV(r)
		}
		return scan3Generic(x, yv, subtract)
	case *AV:
		return scan3Generic(x, yv, subtract)
	default:
		return subtract(x, y)
	}
}

func scanSubtractNumbers[T number, U number](r []U, x U, y []T) {
	for i, yi := range y {
		x -= U(yi)
		r[i] = x
	}
}

func scanTrimSuffixes(r []string, x string, y []string) {
	for i, yi := range y {
		x = strings.TrimSuffix(x, yi)
		r[i] = x
	}
}

func scan2vMax(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vMax(NewV(xv.values)))
	case *AB:
		fl := xv.flags & flagBool
		r := xv.reuse()
		scanMaxNumbers(r.elts, 0, xv.elts)
		r.flags = fl | flagAscending
		return NewV(r)
	case *AI:
		r := xv.reuse()
		scanMaxNumbers(r.elts, 0, xv.elts)
		r.flags = flagAscending
		return NewV(r)
	case *AF:
		r := xv.reuse()
		scanMaxNumbers(r.elts, 0.0, xv.elts)
		r.flags = flagAscending
		return NewV(r)
	case *AS:
		r := xv.reuse()
		scanMaxStrings(r.elts, "", xv.elts)
		r.flags = flagAscending
		return NewV(r)
	case *AV:
		return scan2Generic(xv, maximum)
	default:
		return x
	}
}

func scan3vMax(x, y V) V {
	switch yv := y.bv.(type) {
	case *Dict:
		return newDictValues(yv.keys, scan3vMax(x, NewV(yv.values)))
	case *AB:
		if x.IsI() {
			if x.I() >= 0 && x.I() < 256 {
				var fl flags
				if x.I() <= 1 {
					fl = yv.flags & flagBool
				}
				r := yv.reuse()
				scanMaxNumbers(r.elts, byte(x.I()), yv.elts)
				r.flags = fl | flagAscending
				return NewV(r)
			}
			r := make([]int64, yv.Len())
			scanMaxNumbers(r, x.I(), yv.elts)
			return NewV(&AI{elts: r, flags: flagAscending})
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMaxNumbers(r, x.F(), yv.elts)
			return NewV(&AF{elts: r, flags: flagAscending})
		}
		return scan3Generic(x, yv, maximum)
	case *AI:
		if x.IsI() {
			r := yv.reuse()
			scanMaxNumbers(r.elts, x.I(), yv.elts)
			r.flags = flagAscending
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMaxNumbers(r, x.F(), yv.elts)
			return NewV(&AF{elts: r, flags: flagAscending})
		}
		return scan3Generic(x, yv, maximum)
	case *AF:
		if x.IsI() {
			r := yv.reuse()
			scanMaxNumbers(r.elts, float64(x.I()), yv.elts)
			r.flags = flagAscending
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMaxNumbers(r, x.F(), yv.elts)
			return NewV(&AF{elts: r, flags: flagAscending})
		}
		return scan3Generic(x, yv, maximum)
	case *AS:
		if s, ok := x.bv.(S); ok {
			r := yv.reuse()
			scanMaxStrings(r.elts, string(s), yv.elts)
			r.flags = flagAscending
			return NewV(r)
		}
		return scan3Generic(x, yv, maximum)
	case *AV:
		return scan3Generic(x, yv, maximum)
	default:
		return maximum(x, y)
	}
}

func scanMaxNumbers[T number, U number](dst []U, x U, y []T) {
	for i, yi := range y {
		if U(yi) > x {
			x = U(yi)
		}
		dst[i] = x
	}
}

func scanMaxStrings(dst []string, x string, y []string) {
	for i, yi := range y {
		if yi > x {
			x = yi
		}
		dst[i] = x
	}
}

func scan2vMin(x V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return newDictValues(xv.keys, scan2vMin(NewV(xv.values)))
	case *AB:
		r := xv.reuse()
		scanMinNumbers(r.elts, math.MaxUint8, xv.elts)
		return NewV(r)
	case *AI:
		r := xv.reuse()
		scanMinNumbers(r.elts, math.MaxInt64, xv.elts)
		return NewV(r)
	case *AF:
		r := xv.reuse()
		scanMinNumbers(r.elts, math.MaxFloat64, xv.elts)
		return NewV(r)
	case *AS:
		if xv.Len() == 0 {
			return x
		}
		r := xv.reuse()
		r.elts[0] = xv.elts[0]
		scanMinStrings(r.elts[1:], xv.elts[0], xv.elts[1:])
		return NewV(r)
	case *AV:
		return scan2Generic(xv, minimum)
	default:
		return x
	}
}

func scan3vMin(x, y V) V {
	switch yv := y.bv.(type) {
	case *Dict:
		return newDictValues(yv.keys, scan3vMin(x, NewV(yv.values)))
	case *AB:
		if x.IsI() {
			if x.I() >= 0 && x.I() < 256 {
				var fl flags
				if x.I() <= 1 {
					fl = flagBool
				}
				r := yv.reuse()
				scanMinNumbers(r.elts, byte(x.I()), yv.elts)
				r.flags = fl
				return NewV(r)
			}
			r := make([]int64, yv.Len())
			scanMinNumbers(r, x.I(), yv.elts)
			return NewAI(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMinNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, minimum)
	case *AI:
		if x.IsI() {
			r := yv.reuse()
			scanMinNumbers(r.elts, x.I(), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMinNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, minimum)
	case *AF:
		if x.IsI() {
			r := yv.reuse()
			scanMinNumbers(r.elts, float64(x.I()), yv.elts)
			return NewV(r)
		}
		if x.IsF() {
			r := make([]float64, yv.Len())
			scanMinNumbers(r, x.F(), yv.elts)
			return NewAF(r)
		}
		return scan3Generic(x, yv, minimum)
	case *AS:
		if s, ok := x.bv.(S); ok {
			r := yv.reuse()
			scanMinStrings(r.elts, string(s), yv.elts)
			return NewV(r)
		}
		return scan3Generic(x, yv, minimum)
	case *AV:
		return scan3Generic(x, yv, minimum)
	default:
		return minimum(x, y)
	}
}

func scanMinNumbers[T number, U number](dst []U, x U, y []T) {
	for i, yi := range y {
		if U(yi) < x {
			x = U(yi)
		}
		dst[i] = x
	}
}

func scanMinStrings(dst []string, x string, y []string) {
	for i, yi := range y {
		if yi < x {
			x = yi
		}
		dst[i] = x
	}
}

func each3Match(x, y V) V {
	switch xv := x.bv.(type) {
	case *Dict:
		return newDictValues(xv.keys, each3Match(NewV(xv.values), y))
	default:
		xa, ok := x.bv.(array)
		if !ok {
			break
		}
		ya, ok := y.bv.(array)
		if !ok {
			yd, ok := y.bv.(*Dict)
			if !ok {
				break
			}
			ya = yd.values
		}
		xlen := xa.Len()
		r := make([]byte, xlen)
		for i := 0; i < xlen; i++ {
			xi := xa.at(i)
			yi := ya.at(i)
			r[i] = b2B(xi.Matches(yi))
		}
		return newABb(r)
	}
	l := maxInt(x.Len(), y.Len())
	r := make([]byte, l)
	for i := 0; i < l; i++ {
		xi := x.at(i)
		yi := y.at(i)
		r[i] = b2B(xi.Matches(yi))
	}
	return newABb(r)
}
