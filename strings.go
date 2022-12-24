package goal

import (
	"fmt"
	"strings"
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

func (r *nReplacer) Sprint(ctx *Context, sb *strings.Builder) {
	sb.WriteString("sub[")
	r.olds.Sprint(ctx, sb)
	sb.WriteByte(';')
	r.news.Sprint(ctx, sb)
	sb.WriteByte(';')
	fmt.Fprintf(sb, "%d", r.n)
	sb.WriteByte(']')
}

func (r *nReplacer) Type() string {
	return "f"
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

func (r *replacer) Sprint(ctx *Context, sb *strings.Builder) {
	sb.WriteString("sub[")
	r.oldnew.Sprint(ctx, sb)
	sb.WriteByte(']')
}

func (r *replacer) Type() string {
	return "f"
}

func (r *replacer) replace(ctx *Context, s string) string {
	return r.r.Replace(s)
}

func (r *replacer) rcincr() {
	r.oldnew.rcincr()
}

func (r *replacer) rcdecr() {
	r.oldnew.rcdecr()
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
		for i, n := range xv.Slice {
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
		for i, xi := range xv.Slice {
			r[i] = applyS(s, xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("s[x] : x non-integer (%s)", x.Type())
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
				return Panicf("s[x;y] : length mismatch: %d (#x) %d (#y)",
					xv.Len(), z.Len())

			}
			for i, n := range xv.Slice {
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
		for i, n := range xv.Slice {
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
		for i, xi := range xv.Slice {
			r[i] = applyS2(s, xi, y)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return Panicf("s[x;y] : x non-integer (%s)", x.Type())
	}
}

func bytes(x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewI(int64(len(xv)))
	case *AS:
		r := make([]int64, xv.Len())
		for i, s := range xv.Slice {
			r[i] = int64(len(s))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = bytes(xi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panicType("bytes x", "x", x)
	}
}

// cast implements s$y.
func cast(s S, y V) V {
	switch s {
	case "i":
		return casti(y)
	case "n":
		return castn(y)
	case "s":
		return casts(y)
	default:
		return Panicf("s$y : unsupported \"%s\" value for s", s)
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
		runes := []rune(yv)
		r := make([]int64, len(runes))
		for i, rc := range runes {
			r[i] = int64(rc)
		}
		return NewAI(r)
	case *AB:
		return y
	case *AI:
		return y
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.Slice {
			r[i] = casti(NewS(s))
		}
		return NewAV(r)
	case *AF:
		return toAI(floor(y).value.(*AF))
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casti(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
	default:
		return panicType("\"i\"$y", "y", y)
	}
}

func castn(y V) V {
	if y.IsI() || y.IsF() {
		return y
	}
	switch yv := y.value.(type) {
	case S:
		xi, err := parseNumber(string(yv))
		if err != nil {
			return Errorf("%v", err)
		}
		return xi
	case *AB:
		return y
	case *AI:
		return y
	case *AS:
		r := make([]V, yv.Len())
		for i, s := range yv.Slice {
			n, err := parseNumber(s)
			if err != nil {
				return Errorf("%v", err)
			}
			r[i] = n
		}
		return Canonical(NewAV(r))
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
	default:
		return panicType("\"n\"$y", "y", y)
	}
}

func casts(y V) V {
	if y.IsI() {
		return NewS(string(rune(y.I())))
	}
	if y.IsF() {
		return casts(NewI(int64(y.F())))
	}
	switch yv := y.value.(type) {
	case *AB:
		return casts(fromABtoAI(yv))
	case *AI:
		sb := strings.Builder{}
		for _, i := range yv.Slice {
			sb.WriteRune(rune(i))
		}
		return NewS(sb.String())
	case *AF:
		return casts(toAI(yv))
	case *AV:
		r := make([]V, yv.Len())
		for i := range r {
			r[i] = casts(yv.At(i))
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return Canonical(NewAV(r))
	default:
		return panicType("\"s\"$y", "y", y)
	}
}

func drops(s S, y V) V {
	switch yv := y.value.(type) {
	case S:
		return NewS(strings.TrimPrefix(string(yv), string(s)))
	case *AS:
		r := make([]string, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = strings.TrimPrefix(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = drops(s, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
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
		for i, yi := range yv.Slice {
			r[i] = strings.Trim(string(yi), string(s))
		}
		return NewAS(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			r[i] = trim(s, yi)
			if r[i].IsPanic() {
				return r[i]
			}
		}
		return NewAV(r)
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
		return NewV(&replacer{r: strings.NewReplacer(xv.Slice...), oldnew: xv})
	default:
		return panicType("sub[x]", "x", x)
	}
}

func sub2(x, y V) V {
	switch xv := x.value.(type) {
	case S:
		yv, ok := y.value.(S)
		if !ok {
			return panicType("sub[s;y]", "y", y)
		}
		return NewV(&nReplacer{olds: xv, news: yv, n: -1})
	case *AS:
		yv, ok := y.value.(*AS)
		if !ok {
			return panicType("sub[S;y]", "y", y)
		}
		if xv.Len() != yv.Len() {
			return Panicf("sub[S;S] : length mismatch (%d vs %d)", xv.Len(), yv.Len())
		}
		oldnew := make([]string, 2*xv.Len())
		for i, xi := range xv.Slice {
			oldnew[2*i] = xi
			oldnew[2*i+1] = yv.Slice[i]
		}
		return NewV(&replacer{r: strings.NewReplacer(oldnew...), oldnew: &AS{Slice: oldnew}})
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
			return panicType("sub[s;y;z]", "y", y)
		}
		var n int64
		if z.IsI() {
			n = z.I()
		} else if z.IsF() {
			if !isI(z.F()) {
				return panicType("sub[s;y;z]", "z", z)
			}
			n = int64(z.F())
		}
		return NewV(&nReplacer{olds: xv, news: yv, n: int(n)})
	default:
		return panicType("sub[x;y;z]", "x", x)
	}
}

func (ctx *Context) replace(f stringReplacer, x V) V {
	switch xv := x.value.(type) {
	case S:
		return NewS(f.replace(ctx, string(xv)))
	case *AS:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			r.Slice[i] = f.replace(ctx, xi)
		}
		return NewV(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := ctx.replace(f, xi)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
	default:
		return panicType("sub[...] x", "x", x)
	}
}

func containedInS(x V, s string) V {
	switch xv := x.value.(type) {
	case S:
		return NewI(b2i(strings.Contains(s, string(xv))))
	case *AS:
		r := make([]bool, xv.Len())
		for i, xi := range xv.Slice {
			r[i] = strings.Contains(s, xi)
		}
		return NewAB(r)
	case *AV:
		r := xv.reuse()
		for i, xi := range xv.Slice {
			ri := containedInS(xi, s)
			if ri.IsPanic() {
				return ri
			}
			r.Slice[i] = ri
		}
		return NewV(r)
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
		for i, yi := range yv.Slice {
			r[i] = int64(strings.Count(string(yi), string(s)))
		}
		return NewAI(r)
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.Slice {
			ri := scount(s, yi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	default:
		return panicType("s#y", "y", y)
	}
}
