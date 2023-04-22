package goal

import (
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
		xv := x.I()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return Panicf("s[i] : i out of bounds index (%d)", xv)
		}
		return NewV(s[xv:])
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s[x] : x non-integer (%g)", x.F())
		}
		return applyS(s, NewI(int64(x.F())))
	}
	switch xv := x.value.(type) {
	case *AB:
		return applyS(s, fromABtoAI(xv))
	case *AI:
		r := make([]string, xv.Len())
		for i, n := range xv.elts {
			if n < 0 {
				n += int64(len(s))
			}
			if n < 0 || n > int64(len(s)) {
				return Panicf("s[i] : i out of bounds index (%d)", n)
			}
			r[i] = string(s[n:])
		}
		return NewAS(r)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return applyS(s, z)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = applyS(s, xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalFast(NewAV(r))
	default:
		return panicType("s[x]", "x", x)
	}
}

func applyS2(s S, x V, y V) V {
	var l int64
	if y.IsI() {
		if y.I() < 0 {
			return Panicf("s[x;y] : y negative (%d)", y.I())
		}
		l = y.I()
	} else if y.IsF() {
		if !isI(y.F()) {
			return Panicf("s[x;y] : y non-integer (%g)", y.F())
		}
		l = int64(y.F())
	} else {
		switch yv := y.value.(type) {
		case *AI:
		case *AB:
			return applyS2(s, x, fromABtoAI(yv))
		case *AF:
			z := toAI(yv)
			if z.IsPanic() {
				return z
			}
			return applyS2(s, x, z)
		default:
			return panicType("s[x;y]", "y", y)
		}
	}
	if x.IsI() {
		xv := x.I()
		if xv < 0 {
			xv += int64(len(s))
		}
		if xv < 0 || xv > int64(len(s)) {
			return Panicf("s[i;y] : i out of bounds index (%d)", xv)
		}
		if _, ok := y.value.(*AI); ok {
			return Panicf("s[x;y] : x is an atom but y is an array")
		}
		if xv+l > int64(len(s)) {
			l = int64(len(s)) - xv
		}
		return NewV(s[xv : xv+l])

	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("s[x;y] : x non-integer (%g)", x.F())
		}
		return applyS2(s, NewI(int64(x.F())), y)
	}
	switch xv := x.value.(type) {
	case *AB:
		return applyS2(s, fromABtoAI(xv), y)
	case *AI:
		r := make([]string, xv.Len())
		if z, ok := y.value.(*AI); ok {
			if z.Len() != xv.Len() {
				return panicLength("s[x;y]", xv.Len(), z.Len())

			}
			for i, n := range xv.elts {
				if n < 0 {
					n += int64(len(s))
				}
				if n < 0 || n > int64(len(s)) {
					return Panicf("s[i;y] : i out of bounds index (%d)", n)
				}
				l := z.At(i)
				if n+l > int64(len(s)) {
					l = int64(len(s)) - n
				}
				r[i] = string(s[n : n+l])
			}
			return NewAS(r)
		}
		for i, n := range xv.elts {
			if n < 0 {
				n += int64(len(s))
			}
			if n < 0 || n > int64(len(s)) {
				return Panicf("s[i;y] : i out of bounds index (%d)", n)
			}
			l := l
			if n+l > int64(len(s)) {
				l = int64(len(s)) - n
			}
			r[i] = string(s[n : n+l])
		}
		return NewAS(r)
	case *AF:
		z := toAI(xv)
		if z.IsPanic() {
			return z
		}
		return applyS2(s, z, y)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = applyS2(s, xi, y)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalFast(NewAV(r))
	default:
		return panicType("s[x;y]", "x", x)
	}
}

// cast implements s$y.
func cast(s S, y V) V {
	switch s {
	case "b":
		return castb(y)
	case "i":
		return casti(y)
	case "n":
		return castn(y)
	case "c":
		return castc(y)
	default:
		return Panicf("s$y : unsupported \"%s\" value for s", s)
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
	switch yv := y.value.(type) {
	case S:
		return NewAI(toAIBytes(string(yv)))
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.elts {
			r[i] = NewAI(toAIBytes(s))
		}
		return NewAV(r)
	case *AB:
		return castb(fromABtoAI(yv))
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
			return y
		}
		return castb(y)
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = castb(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	case *Dict:
		return newDictValues(yv.keys, castb(NewV(yv.values)))
	default:
		return panicType("\"b\"$y", "y", y)
	}
}

func casti(y V) V {
	if y.IsI() {
		return y
	}
	if y.IsF() {
		return NewI(int64(y.F()))
	}
	switch yv := y.value.(type) {
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
		for i, s := range yv.elts {
			n, err := parseInt(string(s))
			if err != nil {
				n = math.MinInt64
			}
			r[i] = n
		}
		return NewAI(r)
	case *AF:
		return castToAI(yv)
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casti(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	case *Dict:
		return newDictValues(yv.keys, casti(NewV(yv.values)))
	default:
		return panicType("\"i\"$y", "y", y)
	}
}

func parseInt(s string) (int64, error) {
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

func toAIrunes(s string) []int64 {
	n := utf8.RuneCountInString(s)
	r := make([]int64, n)
	i := 0
	for _, c := range s {
		r[i] = int64(c)
		i++
	}
	return r
}

func toAIBytes(s string) []int64 {
	r := make([]int64, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = int64(s[i])
	}
	return r
}

func castn(y V) V {
	if y.IsI() {
		return NewF(float64(y.I()))
	}
	if y.IsF() {
		return y
	}
	switch yv := y.value.(type) {
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
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = castn(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	case *Dict:
		return newDictValues(yv.keys, castn(NewV(yv.values)))
	default:
		return panicType("\"n\"$y", "y", y)
	}
}

func castc(y V) V {
	if y.IsI() {
		return NewS(string(rune(y.I())))
	}
	if y.IsF() {
		return castc(NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
	case S:
		return NewAI(toAIrunes(string(yv)))
	case *AB:
		return castc(fromABtoAI(yv))
	case *AI:
		sb := strings.Builder{}
		for _, i := range yv.elts {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case *AF:
		return castc(toAI(yv))
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.elts {
			r[i] = NewAI(toAIrunes(s))
		}
		return NewAV(r)
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = castc(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return canonicalFast(NewAV(r))
	default:
		return panicType("\"c\"$y", "y", y)
	}
}

func dropS(s S, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewS(strings.TrimPrefix(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.TrimPrefix(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			r[i] = dropS(s, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	case *Dict:
		return newDictValues(yv.keys, dropS(s, NewV(yv.values)))
	default:
		return panicType("s_y", "y", y)
	}
}

// trim returns s^y.
func trim(s S, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewS(strings.Trim(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.Trim(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			r[i] = trim(s, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	case *Dict:
		return newDictValues(yv.keys, trim(s, NewV(yv.values)))
	default:
		return panicType("s^y", "y", y)
	}
}

// trimSpaces returns ""^y.
func trimSpaces(y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewS(strings.TrimSpace(string(yv)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.elts {
			r[i] = strings.TrimSpace(string(yi))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			r[i] = trimSpaces(yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	case *Dict:
		return newDictValues(yv.keys, trimSpaces(NewV(yv.values)))
	default:
		return panicType("s^y", "y", y)
	}
}

func sub1(x V) V {
	switch xv := x.value.(type) {
	case *AS:
		if xv.Len()%2 != 0 {
			return panics("sub[S] : non-even length array")
		}
		return NewV(&replacer{r: strings.NewReplacer(xv.elts...), oldnew: xv})
	default:
		return panicType("sub[x]", "x", x)
	}
}

func sub2(x, y V) V {
	switch xv := x.value.(type) {
	case S:
		yv, ok := y.value.(S)
		if !ok {
			return panicType("sub[s;s]", "s", y)
		}
		return NewV(&nReplacer{olds: xv, news: yv, n: -1})
	case *AS:
		yv, ok := y.value.(*AS)
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
		return NewV(&replacer{r: strings.NewReplacer(oldnew...), oldnew: &AS{elts: oldnew, rc: reuseRCp(yv.rc)}})
	case *rx:
		switch y.value.(type) {
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
	switch xv := x.value.(type) {
	case S:
		yv, ok := y.value.(S)
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
	switch xv := x.value.(type) {
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
	switch xv := x.value.(type) {
	case S:
		return NewI(B2I(strings.Contains(s, string(xv))))
	case *AS:
		r := make([]bool, xv.Len())
		for i, xi := range xv.elts {
			r[i] = strings.Contains(s, xi)
		}
		return NewAB(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.elts {
			ri := containedInS(xi, s)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
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
	switch yv := y.value.(type) {
	case S:
		return NewI(int64(strings.Count(string(yv), string(s))))
	case *AS:
		r := make([]int64, yv.Len())
		for i, yi := range yv.elts {
			r[i] = int64(strings.Count(string(yi), string(s)))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := scount(s, yi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	case *Dict:
		return newDictValues(yv.keys, scount(s, NewV(yv.values)))
	default:
		return panicType("s#y", "y", y)
	}
}

func splitN(n int, sep S, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewAS(strings.SplitN(string(yv), string(sep), n))
	case *AS:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = NewAS(strings.SplitN(yv.At(i), string(sep), n))
		}
		return NewAV(r)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.elts {
			ri := splitN(n, sep, yi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
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

func padStrings(x int, y V) V {
	switch yv := y.value.(type) {
	case S:
		if len(yv) < x || len(yv) < -x {
			return NewS(padString(x, string(yv)))
		}
		return y
	case *AS:
		r := yv.reuse()
		for i, yi := range yv.elts {
			r.elts[i] = padString(x, yi)
		}
		return NewV(r)
	case *AV:
		r := yv.reuse()
		for i, yi := range yv.elts {
			ri := padStrings(x, yi)
			if ri.IsPanic() {
				return ri
			}
			r.elts[i] = ri
		}
		return NewV(r)
	default:
		return panicType("i$y", "y", y)
	}
}

func padString(x int, s string) string {
	switch {
	case len(s) < x:
		var sb strings.Builder
		sb.Grow(x)
		sb.WriteString(s)
		for i := 0; i < x-len(s); i++ {
			sb.WriteByte(' ')
		}
		return sb.String()
	case len(s) < -x:
		var sb strings.Builder
		sb.Grow(-x)
		for i := 0; i < -x-len(s); i++ {
			sb.WriteByte(' ')
		}
		sb.WriteString(s)
		return sb.String()
	default:
		return s
	}
}
