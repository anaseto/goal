package goal

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type stringReplacer interface {
	Value
	replace(*Context, string) string
}

type nReplacer struct {
	olds S
	news S
	n    int
}

func (r *nReplacer) Matches(x Value) bool {
	xv, ok := x.(*nReplacer)
	return ok && r.olds == xv.olds && r.news == xv.news && r.n == xv.n
}

func (r *nReplacer) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "sub["...)
	dst = r.olds.Append(ctx, dst)
	dst = append(dst, ';')
	dst = r.news.Append(ctx, dst)
	dst = append(dst, ';')
	dst = strconv.AppendInt(dst, int64(r.n), 10)
	dst = append(dst, ']')
	return dst
}

func (r *nReplacer) Type() string {
	return "f"
}

func (r *nReplacer) stype() string {
	return "nReplacer"
}

func (r *nReplacer) rank(ctx *Context) int {
	return 1
}

func (r *nReplacer) replace(ctx *Context, s string) string {
	return strings.Replace(s, string(r.olds), string(r.news), r.n)
}

type replacer struct {
	r      *strings.Replacer
	oldnew *AS
}

func (r *replacer) Matches(x Value) bool {
	xv, ok := x.(*replacer)
	return ok && r.oldnew.Matches(xv.oldnew)
}

func (r *replacer) Append(ctx *Context, dst []byte) []byte {
	dst = append(dst, "sub["...)
	dst = r.oldnew.Append(ctx, dst)
	dst = append(dst, ']')
	return dst
}

func (r *replacer) Type() string {
	return "f"
}

func (r *replacer) stype() string {
	return "replacer"
}

func (r *replacer) rank(ctx *Context) int {
	return 1
}

func (r *replacer) replace(ctx *Context, s string) string {
	return r.r.Replace(s)
}

func applyS(s S, x V) V {
	if x.IsI() {
		r, err := applySI(string(s), x.I())
		if err != nil {
			return panicErr(err)
		}
		return NewS(r)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s@i : non-integer i (%g)", x.F())
		}
		return applyS(s, NewI(int64(x.F())))
	}
	switch xv := x.bv.(type) {
	case *AB:
		return applySIntegers(string(s), xv.elts)
	case *AI:
		return applySIntegers(string(s), xv.elts)
	case *AF:
		x := toAI(xv)
		if x.IsPanic() {
			return ppanic("s@i : ", x)
		}
		return applyS(s, x)
	case *AV:
		return Canonical(monadAV(xv, func(xi V) V { return applyS(s, xi) }))
	default:
		return panicType("s@i", "i", x)
	}
}

func applySIntegers[I integer](s string, x []I) V {
	r := make([]string, len(x))
	for i, xi := range x {
		ri, err := applySI(string(s), int64(xi))
		if err != nil {
			return panicErr(err)
		}
		r[i] = ri
	}
	return NewAS(r)
}

func applySI(s string, x int64) (string, error) {
	if x < 0 {
		x += int64(len(s))
	}
	if x < 0 || x > int64(len(s)) {
		return "", nil
	}
	return string(s[x:]), nil
}

func applyS2(s S, x V, y V) V {
	if x.IsI() {
		return applyS2I(s, x.I(), y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s[x;y] : non-integer x (%g)", x.F())
		}
		return applyS2I(s, int64(x.F()), y)
	}
	switch xv := x.bv.(type) {
	case *AB:
		return applyS2Is(s, xv.elts, y)
	case *AI:
		return applyS2Is(s, xv.elts, y)
	case *AF:
		x := toAI(xv)
		if x.IsPanic() {
			return ppanic("s[x;y] : x ", x)
		}
		return applyS2(s, x, y)
	case *AV:
		return Canonical(monadAV(xv, func(xi V) V { return applyS2(s, xi, y) }))
	default:
		return panicType("s[x;y]", "x", x)
	}
}

func applyS2I(s S, x int64, y V) V {
	if y.IsI() {
		r, err := applyS2II(string(s), x, y.I())
		if err != nil {
			return panicErr(err)
		}
		return NewS(r)
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("s[i;y] : non-integer y (%g)", y.F())
		}
		return applyS2I(s, x, NewI(int64(y.F())))
	}
	switch yv := y.bv.(type) {
	case *AB:
		return applyS2IIs(string(s), x, yv.elts)
	case *AI:
		return applyS2IIs(string(s), x, yv.elts)
	case *AF:
		y := toAI(yv)
		if y.IsPanic() {
			return ppanic("s[i;y] : y ", y)
		}
		return applyS2I(s, x, y)
	case *AV:
		return monadAV(yv, func(yi V) V { return applyS2I(s, int64(x), yi) })
	default:
		return panicType("s[i;y]", "y", y)
	}
}

func applyS2IIs[I integer](s string, x int64, y []I) V {
	r := make([]string, len(y))
	for i, yi := range y {
		ri, err := applyS2II(s, int64(x), int64(yi))
		if err != nil {
			return panicErr(err)
		}
		r[i] = ri
	}
	return NewAS(r)
}

func applyS2Is[I integer](s S, x []I, y V) V {
	if y.IsI() {
		l := y.I()
		r := make([]string, len(x))
		for i, xi := range x {
			ri, err := applyS2II(string(s), int64(xi), l)
			if err != nil {
				return panicErr(err)
			}
			r[i] = ri
		}
		return NewAS(r)
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("s[i;y] : non-integer y (%g)", y.F())
		}
		return applyS2Is(s, x, NewI(int64(y.F())))
	}
	switch yv := y.bv.(type) {
	case *AB:
		return applyS2IsIs(string(s), x, yv.elts)
	case *AI:
		return applyS2IsIs(string(s), x, yv.elts)
	case *AF:
		y := toAI(yv)
		if y.IsPanic() {
			return ppanic("s[i;y] : y ", y)
		}
		return applyS2Is(s, x, y)
	case *AV:
		if len(x) != yv.Len() {
			return panicLength("s[x;y]", len(x), yv.Len())
		}
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := applyS2I(s, int64(x[i]), yi)
			if ri.IsPanic() {
				return ri
			}
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	default:
		return panicType("s[i;y]", "y", y)
	}
}

func applyS2IsIs[I integer, J integer](s string, x []I, y []J) V {
	if len(x) != len(y) {
		return panicLength("s[x;y]", len(x), len(y))
	}
	r := make([]string, len(x))
	for i, xi := range x {
		ri, err := applyS2II(string(s), int64(xi), int64(y[i]))
		if err != nil {
			return panicErr(err)
		}
		r[i] = ri
	}
	return NewAS(r)
}

func applyS2II(s string, i, l int64) (string, error) {
	if i < 0 {
		i += int64(len(s))
	}
	if i < 0 || i > int64(len(s)) {
		return "", nil
	}
	if l < 0 {
		to := int64(len(s)) + l
		if to < i {
			to = i
		}
		return s[i:to], nil
	}
	to := i + l
	if to > int64(len(s)) {
		to = int64(len(s))
	}
	return s[i:to], nil
}

// cast implements s$y.
func cast(ctx *Context, s S, y V) V {
	switch s {
	case "i":
		return casti(y)
	case "n":
		return castn(y)
	case "s":
		return casts(ctx, y)
	case "b":
		return castb(y)
	case "c":
		return castc(y)
	default:
		return castFormat(ctx, string(s), y)
	}
}

func casti(y V) V {
	if y.IsI() {
		return y
	}
	if y.IsF() {
		return NewI(int64(y.F()))
	}
	switch yv := y.bv.(type) {
	case S:
		xi, err := parseInt(string(yv))
		if err != nil {
			return NewI(math.MinInt64)
		}
		return NewI(xi)
	case *AB:
		return y
	case *AI:
		return y
	case *AS:
		r := make([]int64, yv.Len())
		bo, by := true, true
		for i, s := range yv.elts {
			n, err := parseInt(string(s))
			if err != nil {
				n = math.MinInt64
			}
			if n < 0 || n >= 256 {
				by = false
			} else if n > 1 {
				bo = false
			}
			r[i] = n
		}
		if by {
			rb := &AB{elts: make([]byte, yv.Len())}
			for i, ri := range r {
				rb.elts[i] = byte(ri)
			}
			if bo {
				rb.flags = flagBool
			}
			return NewV(rb)
		}
		return NewAI(r)
	case *AF:
		return castToAI(yv)
	case *AV:
		return Canonical(monadAV(yv, casti))
	case *Dict:
		return newDictValues(yv.keys, casti(NewV(yv.values)))
	default:
		return panicType("\"i\"$y", "y", y)
	}
}

func parseInt(s string) (int64, error) {
	if s == "0i" {
		return math.MinInt64, nil
	}
	i, errI := strconv.ParseInt(s, 0, 0)
	if errI == nil {
		return i, nil
	}
	d, errT := time.ParseDuration(s)
	if errT == nil {
		return int64(d), nil
	}
	err := errI.(*strconv.NumError)
	return 0, err.Err
}

func castn(y V) V {
	if y.IsI() {
		return NewF(float64(y.I()))
	}
	if y.IsF() {
		return y
	}
	switch yv := y.bv.(type) {
	case S:
		xi, err := parseFloat(string(yv))
		if err != nil {
			return NewF(math.NaN())
		}
		return NewF(xi)
	case *AB:
		return fromABtoAF(yv)
	case *AI:
		return toAF(yv)
	case *AS:
		r := make([]float64, yv.Len())
		for i, s := range yv.elts {
			n, err := parseFloat(s)
			if err != nil {
				n = math.NaN()
			}
			r[i] = n
		}
		return NewAF(r)
	case *AF:
		return y
	case *AV:
		return Canonical(monadAV(yv, castn))
	case *Dict:
		return newDictValues(yv.keys, castn(NewV(yv.values)))
	default:
		return panicType("\"n\"$y", "y", y)
	}
}

func parseFloat(s string) (float64, error) {
	switch s {
	case "0n":
		s = "NaN"
	case "0w":
		s = "Inf"
	case "-0w":
		s = "-Inf"
	}
	f, errF := strconv.ParseFloat(s, 64)
	if errF == nil {
		return f, nil
	}
	err := errF.(*strconv.NumError)
	return 0, err.Err
}

func casts(ctx *Context, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return y
	case *AS:
		return y
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.elts {
			ri := casts(ctx, yi)
			ri.MarkImmutable()
			r.elts[i] = ri
		}
		return canonicalAV(r)
	case array:
		return each2String(ctx, yv)
	case *Dict:
		return newDictValues(yv.keys, casts(ctx, NewV(yv.values)))
	default:
		return NewS(y.Sprint(ctx))
	}
}

func castb(y V) V {
	if y.IsI() {
		return NewS(string([]byte{byte(y.I())}))
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf(`"b"$i : non-integer i (%g)`, y.F())
		}
		return NewS(string([]byte{byte(y.F())}))
	}
	switch yv := y.bv.(type) {
	case S:
		return NewAB([]byte(string(yv)))
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.elts {
			r[i] = NewV(&AB{elts: []byte(s), flags: flagImmutable})
		}
		return newAVu(r)
	case *AB:
		return NewS(string(yv.elts))
	case *AI:
		var sb strings.Builder
		sb.Grow(yv.Len())
		for _, xi := range yv.elts {
			sb.WriteByte(byte(xi))
		}
		return NewS(sb.String())
	case *AF:
		y := toAI(yv)
		if y.IsPanic() {
			return ppanic(`"b"$i : `, y)
		}
		return castb(y)
	case *AV:
		return Canonical(monadAV(yv, castb))
	default:
		return panicType("\"b\"$y", "y", y)
	}
}

func castc(y V) V {
	if y.IsI() {
		return NewS(string(rune(y.I())))
	}
	if y.IsF() {
		return castc(NewI(int64(y.F())))
	}
	switch yv := y.bv.(type) {
	case S:
		return castcString(string(yv))
	case *AB:
		sb := strings.Builder{}
		for _, i := range yv.elts {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case *AI:
		sb := strings.Builder{}
		for _, i := range yv.elts {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case *AF:
		y = toAI(yv)
		if y.IsPanic() {
			return ppanic(`"c"$i : `, y)
		}
		return castc(y)
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.elts {
			ri := castcString(s)
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *AV:
		return Canonical(monadAV(yv, castc))
	default:
		return panicType("\"c\"$y", "y", y)
	}
}

func castcString(s string) V {
	var ascii bool = true
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			ascii = false
			break
		}
	}
	if ascii {
		return NewAB([]byte(s))
	}
	n := utf8.RuneCountInString(s)
	r := make([]int64, n)
	i := 0
	for _, c := range s {
		r[i] = int64(c)
		i++
	}
	return NewAI(r)
}

func castFormat(ctx *Context, s string, y V) V {
	n := strings.Count(s, "%")
	ne := strings.Count(s, "%%")
	nv := n - 2*ne
	if nv == 0 {
		return Panicf(`s$y : unsupported "%s" for s`, s)
	}
	if nv > 1 {
		return castFormatN(ctx, s, y, nv)
	}
	return castFormat1(ctx, s, y)
}

func castFormat1(ctx *Context, s string, y V) V {
	if y.IsI() {
		return NewS(fmt.Sprintf(s, y.I()))
	}
	if y.IsF() {
		return NewS(fmt.Sprintf(s, y.F()))
	}
	switch yv := y.bv.(type) {
	case S:
		return NewS(fmt.Sprintf(s, string(yv)))
	case *AB:
		return NewAS(format1Array(s, yv.elts))
	case *AI:
		return NewAS(format1Array(s, yv.elts))
	case *AF:
		return NewAS(format1Array(s, yv.elts))
	case *AS:
		r := yv.reuse()
		for i, yi := range yv.elts {
			r.elts[i] = fmt.Sprintf(s, yi)
		}
		return NewV(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := castFormat1(ctx, s, yi)
			ri.MarkImmutable()
			r[i] = ri
		}
		return canonicalVs(r)
	case *Dict:
		return newDictValues(yv.keys, castFormat1(ctx, s, NewV(yv.values)))
	default:
		return NewS(fmt.Sprintf(s, y.Sprint(ctx)))
	}
}

func format1Array[T any](s string, y []T) []string {
	r := make([]string, len(y))
	for i, yi := range y {
		r[i] = fmt.Sprintf(s, yi)
	}
	return r
}

func castFormatN(ctx *Context, s string, y V, n int) V {
	if y.IsI() {
		return NewS(fmt.Sprintf(s, y.I()))
	}
	if y.IsF() {
		return NewS(fmt.Sprintf(s, y.F()))
	}
	switch yv := y.bv.(type) {
	case S:
		return NewS(fmt.Sprintf(s, string(yv)))
	case *AB:
		return NewS(formatNArray(s, yv.elts, n))
	case *AI:
		return NewS(formatNArray(s, yv.elts, n))
	case *AF:
		return NewS(formatNArray(s, yv.elts, n))
	case *AS:
		return NewS(formatNArray(s, yv.elts, n))
	case *AV:
		r := make([]any, yv.Len())
		for i, yi := range yv.elts {
			r[i] = valueToAny(ctx, yi)
		}
		return NewS(fmt.Sprintf(s, r...))
	default:
		return NewS(fmt.Sprintf(s, y.Sprint(ctx)))
	}
}

func valueToAny(ctx *Context, x V) any {
	if x.IsI() {
		return x.I()
	}
	if x.IsF() {
		return x.F()
	}
	switch xv := x.bv.(type) {
	case S:
		return string(xv)
	default:
		return x.Sprint(ctx)
	}
}

func formatNArray[T any](s string, y []T, n int) string {
	buf := make([]any, 0, n)
	to := minInt(n, len(y))
	for i := 0; i < to; i++ {
		buf = append(buf, y[i])
	}
	return fmt.Sprintf(s, buf...)
}

func dropS(s S, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewS(strings.TrimPrefix(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.TrimPrefix(yi, string(s))
		}
		return NewAS(r)
	case *AV:
		return monadAV(yv, func(yi V) V { return dropS(s, yi) })
	case *Dict:
		return newDictValues(yv.keys, dropS(s, NewV(yv.values)))
	default:
		return panicType("s_y", "y", y)
	}
}

func dropAS(x *AS, y V) V {
	switch yv := y.bv.(type) {
	case S:
		r := make([]string, x.Len())
		for i, xi := range x.elts {
			r[i] = strings.TrimPrefix(string(yv), xi)
		}
		return NewAS(r)
	case *AS:
		if x.Len() != yv.Len() {
			return panicLength("S_S", x.Len(), yv.Len())
		}
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.TrimPrefix(string(yi), x.At(i))
		}
		return NewAS(r)
	case *AV:
		if x.Len() != yv.Len() {
			return panicLength("S_S", x.Len(), yv.Len())
		}
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := dropS(S(x.At(i)), yi)
			if ri.IsPanic() {
				return ri
			}
			ri.MarkImmutable()
			r[i] = ri
		}
		return newAVu(r)
	case *Dict:
		if x.Len() != yv.Len() {
			return panicLength("S_S", x.Len(), yv.Len())
		}
		return newDictValues(yv.keys, dropAS(x, NewV(yv.values)))
	default:
		return panicType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewS(strings.Trim(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.Trim(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		return monadAV(yv, func(yi V) V { return trim(s, yi) })
	case *Dict:
		return newDictValues(yv.keys, trim(s, NewV(yv.values)))
	default:
		return panicType("s^y", "y", y)
	}
}

// trimSpaces returns ""^y.
func trimSpaces(y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewS(strings.TrimSpace(string(yv)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.TrimSpace(string(yi))
		}
		return NewAS(r)
	case *AV:
		return monadAV(yv, trimSpaces)
	case *Dict:
		return newDictValues(yv.keys, trimSpaces(NewV(yv.values)))
	default:
		return panicType("s^y", "y", y)
	}
}

func sub1(x V) V {
	switch xv := x.bv.(type) {
	case *AS:
		if xv.Len()%2 != 0 {
			return panics("sub[S] : non-even length array")
		}
		return NewV(&replacer{r: strings.NewReplacer(xv.elts...), oldnew: xv})
	case *Dict:
		return sub2(NewV(xv.keys), NewV(xv.values))
	default:
		return panicType("sub[x]", "x", x)
	}
}

func sub2(x, y V) V {
	switch xv := x.bv.(type) {
	case S:
		yv, ok := y.bv.(S)
		if !ok {
			return panicType("sub[s;s]", "s", y)
		}
		return NewV(&nReplacer{olds: xv, news: yv, n: -1})
	case *AS:
		yv, ok := y.bv.(*AS)
		if !ok {
			return panicType("sub[S;S]", "S", y)
		}
		if xv.Len() != yv.Len() {
			return panicLength("sub[S;S]", xv.Len(), yv.Len())
		}
		oldnew := make([]string, 2*xv.Len())
		for i, xi := range xv.elts {
			oldnew[2*i] = xi
			oldnew[2*i+1] = yv.elts[i]
		}
		return NewV(&replacer{r: strings.NewReplacer(oldnew...), oldnew: &AS{elts: oldnew}})
	case *rx:
		switch y.bv.(type) {
		case S:
			return NewV(&rxReplacer{r: xv, repl: y})
		default:
			if y.IsFunction() {
				return NewV(&rxReplacer{r: xv, repl: y})
			}
			return panicType("sub[r;y]", "y", y)
		}
	default:
		return panicType("sub[x;y]", "x", x)
	}
}

func sub3(x, y, z V) V {
	switch xv := x.bv.(type) {
	case S:
		yv, ok := y.bv.(S)
		if !ok {
			return panicType("sub[s;y;i]", "y", y)
		}
		var n int64
		if z.IsI() {
			n = z.I()
		} else if z.IsF() {
			if !isI(z.F()) {
				return Panicf("sub[s;s;i] : non-integer i (%g)", z.F())
			}
			n = int64(z.F())
		} else {
			return panicType("sub[s;s;i]", "i", z)
		}
		return NewV(&nReplacer{olds: xv, news: yv, n: int(n)})
	default:
		return panicType("sub[x;s;i]", "x", x)
	}
}

func (ctx *Context) replace(f stringReplacer, x V) V {
	switch xv := x.bv.(type) {
	case S:
		return NewS(f.replace(ctx, string(xv)))
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.elts {
			r.elts[i] = f.replace(ctx, xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := ctx.replace(f, xi)
			if ri.IsPanic() {
				return ri
			}
			ri.MarkImmutable()
			r.elts[i] = ri
		}
		return NewV(r)
	case *Dict:
		return newDictValues(xv.keys, ctx.replace(f, NewV(xv.values)))
	default:
		return panicType("sub[...] x", "x", x)
	}
}

func containedInS(x V, s string) V {
	switch xv := x.bv.(type) {
	case S:
		return NewI(b2I(strings.Contains(s, string(xv))))
	case *AS:
		r := make([]byte, xv.Len())
		for i, xi := range xv.elts {
			r[i] = b2B(strings.Contains(s, xi))
		}
		return newABb(r)
	case *AV:
		return monadAV(xv, func(xi V) V { return containedInS(xi, s) })
	case *Dict:
		return newDictValues(xv.keys, containedInS(NewV(xv.values), s))
	default:
		return panicType("x in s", "x", x)
	}
}

func srepeat(s S, n int64) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(string(s), int(n))
}

func scount(s S, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewI(int64(strings.Count(string(yv), string(s))))
	case *AS:
		r := make([]int64, yv.Len())
		for i, yi := range yv.elts {
			r[i] = int64(strings.Count(string(yi), string(s)))
		}
		return NewAI(r)
	case *AV:
		return monadAV(yv, func(yi V) V { return scount(s, yi) })
	case *Dict:
		return newDictValues(yv.keys, scount(s, NewV(yv.values)))
	default:
		return panicType("s#y", "y", y)
	}
}

func splitN(n int, sep S, y V) V {
	switch yv := y.bv.(type) {
	case S:
		return NewAS(strings.SplitN(string(yv), string(sep), n))
	case *AS:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = NewV(&AS{elts: strings.SplitN(yv.At(i), string(sep), n), flags: flagImmutable})
		}
		return newAVu(r)
	case *AV:
		return monadAV(yv, func(yi V) V { return splitN(n, sep, yi) })
	case *Dict:
		return newDictValues(yv.keys, splitN(n, sep, NewV(yv.values)))
	default:
		return Panicf("bad type \"%s\" in y", y.Type())
	}
}

func lineSplit(s string) []string {
	n := strings.Count(s, "\n") + 1
	r := make([]string, n)
	n--
	i := 0
	for i < n {
		j := strings.IndexByte(s, '\n')
		if j < 0 {
			// should not happen
			break
		}
		r[i] = dropCR(s[:j])
		s = s[j+1:]
		i++
	}
	r[i] = s
	return r[:i+1]
}

func dropCR(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\r' {
		return s[0 : len(s)-1]
	}
	return s
}

func padStringRight(x int, s string) string {
	switch {
	case len(s) < x:
		var sb strings.Builder
		sb.Grow(x)
		sb.WriteString(s)
		for i := 0; i < x-len(s); i++ {
			sb.WriteByte(' ')
		}
		return sb.String()
	default:
		return s
	}
}

func padStringLeft(x int, s string) string {
	switch {
	case len(s) < x:
		var sb strings.Builder
		sb.Grow(x)
		for i := 0; i < x-len(s); i++ {
			sb.WriteByte(' ')
		}
		sb.WriteString(s)
		return sb.String()
	default:
		return s
	}
}
