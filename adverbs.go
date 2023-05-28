package goal

import (
	"fmt"
	"strings"
)

func fold2(ctx *Context, args []V) V {
	f := args[1]
	if !f.IsFunction() {
		if f.IsI() {
			return decode(f, args[0])
		}
		if f.IsF() {
			return decode(f, args[0])
		}
		switch fv := f.value.(type) {
		case S:
			return joinStrings(fv, args[0])
		case *AB, *AI, *AF:
			return decode(f, args[0])
		default:
			return panicType("F/x", "F", f)
		}
	}
	if f.Rank(ctx) == 1 {
		return converge(ctx, f, args[0])
	}
	x := args[0]
	return foldfx(ctx, f, x)
}

func foldfx(ctx *Context, f, x V) V {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vAdd:
			return fold2vAdd(x)
		case vSubtract:
			return fold2vSubtract(x)
		case vMultiply:
			return fold2vMultiply(x)
		case vMax:
			return fold2vMax(x)
		case vMin:
			return fold2vMin(x)
		case vJoin:
			return fold2vJoin(x)
		}
	}
	switch xv := x.value.(type) {
	case *Dict:
		return foldfx(ctx, f, NewV(xv.values))
	case array:
		if xv.Len() == 0 {
			switch xv.(type) {
			case *AS:
				return NewS("")
			case *AV:
				return x
			case *AF:
				return NewF(0)
			default:
				return NewI(0)
			}
		}
		r := xv.at(0)
		ctx.pushNoRC(V{})
		f.IncrRC()
		for i := 1; i < xv.Len(); i++ {
			ctx.replaceTop(xv.at(i))
			ctx.push(r)
			r = f.applyN(ctx, 2)
		}
		f.DecrRC()
		ctx.drop()
		return r
	default:
		return x
	}
}

func joinStrings(sep S, x V) V {
	switch xv := x.value.(type) {
	case S:
		return x
	case *AS:
		return NewS(strings.Join([]string(xv.elts), string(sep)))
	default:
		if x.Len() == 0 {
			return NewS("")
		}
		return panicType("s/S", "S", x)
	}
}

const maxConvergeIters = 1_000_000

func converge(ctx *Context, f, x V) V {
	if dv, ok := f.value.(*derivedVerb); ok && dv.Fun == vFold && dv.Arg.kind == valVariadic {
		switch dv.Arg.variadic() {
		case vJoin:
			return convergeJoin(x)
		}
	}
	n := 0
	f.IncrRC()
	first := x
	first.IncrRC()
	ctx.push(x)
	for {
		x.IncrRC()
		r := f.applyN(ctx, 1)
		x.DecrRC()
		if r.IsPanic() {
			f.DecrRC()
			first.DecrRC()
			ctx.drop()
			return r
		}
		if r.Matches(x) || r.Matches(first) {
			f.DecrRC()
			first.DecrRC()
			ctx.drop()
			return x
		}
		ctx.replaceTop(r)
		x = r
		n++
		if n > maxConvergeIters {
			f.DecrRC()
			first.DecrRC()
			ctx.drop()
			return panics("f/x : too many iterations")
		}
	}
}

func fold3(ctx *Context, args []V) V {
	f := args[2]
	if !f.IsFunction() {
		return panicType("x F/y", "F", f)
	}
	rank := f.Rank(ctx)
	if rank == 1 {
		return doWhile(ctx, args)
	}
	if rank == 2 {
		return foldxfy(ctx, args[1], f, args[0])
	}
	return panicRankN("x F/y", "F", rank, len(args)-1)
}

func foldxfy(ctx *Context, x, f, y V) V {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vAdd:
			return fold3vAdd(x, y)
		case vSubtract:
			return fold3vSubtract(x, y)
		case vMultiply:
			return fold3vMultiply(x, y)
		case vMax:
			return fold3vMax(x, y)
		case vMin:
			return fold3vMin(x, y)
		case vJoin:
			return fold3vJoin(x, y)
		}
	}
	switch yv := y.value.(type) {
	case *Dict:
		return foldxfy(ctx, x, f, NewV(yv.values))
	case array:
		r := x
		if yv.Len() == 0 {
			return r
		}
		f.IncrRC()
		ctx.pushNoRC(V{})
		for i := 0; i < yv.Len(); i++ {
			ctx.replaceTop(yv.at(i))
			ctx.push(r)
			r = f.applyN(ctx, 2)
			if r.IsPanic() {
				f.DecrRC()
				ctx.drop()
				return r
			}
		}
		f.DecrRC()
		ctx.drop()
		return r
	default:
		ctx.push(y)
		ctx.push(x)
		r := f.applyN(ctx, 2)
		ctx.drop()
		return r
	}
}

func getIterLen(args []V) (int, error) {
	mlen := -1
	for _, x := range args {
		switch xv := x.value.(type) {
		case countable:
			switch {
			case mlen < 0:
				mlen = xv.Len()
			case mlen != xv.Len():
				return mlen, fmt.Errorf("length mismatch (%d vs %d)", mlen, xv.Len())
			}
		}
	}
	return mlen, nil
}

func foldN(ctx *Context, args []V) V {
	f := args[len(args)-1]
	if !f.IsFunction() {
		return panicType("f/[x;y;...]", "f", f)
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return panicRankN("f/[x;y;...]", "f", f.Rank(ctx), n)
	}
	mlen, err := getIterLen(args[:len(args)-2])
	if err != nil {
		return Panicf("f/[x;y;...] : %v", err)
	}
	if mlen == -1 {
		return ctx.ApplyN(f, args[:len(args)-1])
	}
	x := args[len(args)-2]
	if mlen == 0 {
		return x
	}
	f.IncrRC()
	ctx.pushNoRC(V{})
	r := x
	for i := 0; i < mlen; i++ {
		ctx.replaceTop(args[0].at(i))
		for j := 1; j < len(args)-2; j++ {
			ctx.push(args[j].at(i))
		}
		ctx.push(r)
		r = f.applyN(ctx, n)
		if r.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return r
		}
	}
	f.DecrRC()
	ctx.drop()
	return r
}

func doWhile(ctx *Context, args []V) V {
	f := args[2]
	x := args[1]
	y := args[0]
	if x.IsI() {
		return doTimes(ctx, x.I(), f, y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("n f/y : non-integer n (%g)", x.F())
		}
		return doTimes(ctx, int64(x.F()), f, y)
	}
	if x.IsFunction() {
		f.IncrRC()
		x.IncrRC()
		ctx.push(y)
		for {
			y.IncrRC()
			cond := x.applyN(ctx, 1)
			y.DecrRC()
			if cond.IsPanic() {
				x.DecrRC()
				f.DecrRC()
				ctx.drop()
				return cond
			}
			if !cond.IsTrue() {
				x.DecrRC()
				f.DecrRC()
				ctx.drop()
				return y
			}
			y = f.applyN(ctx, 1)
			if y.IsPanic() {
				x.DecrRC()
				f.DecrRC()
				ctx.drop()
				return y
			}
			ctx.replaceTop(y)
		}
	}
	return panicType("x f/y", "x", x)
}

func doTimes(ctx *Context, n int64, f, y V) V {
	f.IncrRC()
	ctx.push(y)
	for i := int64(0); i < n; i++ {
		y = f.applyN(ctx, 1)
		if y.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return y
		}
		ctx.replaceTop(y)
	}
	f.DecrRC()
	ctx.drop()
	return y
}

func scan2(ctx *Context, f, x V) V {
	if !f.IsFunction() {
		if f.IsI() {
			return encode(f, x)
		}
		if f.IsF() {
			return encode(f, x)
		}
		switch fv := f.value.(type) {
		case S:
			return splitS(fv, x)
		case *rx:
			return splitRx(fv, x)
		case *AB, *AI, *AF:
			return encode(f, x)
		default:
			return panicType("f\\x", "f", f)
		}
	}
	if f.Rank(ctx) != 2 {
		return converges(ctx, f, x)
	}
	return scanfx(ctx, f, x)
}

func scanfx(ctx *Context, f, x V) V {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vAdd:
			return scan2vAdd(x)
		case vSubtract:
			return scan2vSubtract(x)
		case vMax:
			return scan2vMax(x)
		case vMin:
			return scan2vMin(x)
		}
	}
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, scanfx(ctx, f, NewV(xv.values)))
	case array:
		if xv.Len() == 0 {
			return x
		}
		r := make([]V, xv.Len())
		next := xv.at(0)
		r[0] = next
		next.IncrRC()
		f.IncrRC()
		ctx.pushNoRC(V{})
		for i := 1; i < xv.Len(); i++ {
			ctx.replaceTop(xv.at(i))
			ctx.push(next)
			next = f.applyN(ctx, 2)
			if next.IsPanic() {
				f.DecrRC()
				ctx.drop()
				rcdecrArgs(r)
				return next
			}
			r[i] = next
			next.IncrRC()
		}
		f.DecrRC()
		ctx.drop()
		return normalizedArgs(r)
	default:
		return x
	}
}

func converges(ctx *Context, f, x V) V {
	n := 0
	r := []V{}
	f.IncrRC()
	first := x
	ctx.push(x)
	for {
		r = append(r, x)
		x.IncrRC()
		y := f.applyN(ctx, 1)
		if y.IsPanic() {
			f.DecrRC()
			rcdecrArgs(r)
			ctx.drop()
			return y
		}
		if y.Matches(x) || y.Matches(first) {
			f.DecrRC()
			ctx.drop()
			return normalizedArgs(r)
		}
		ctx.replaceTop(y)
		x = y
		n++
		if n > maxConvergeIters {
			f.DecrRC()
			rcdecrArgs(r)
			ctx.drop()
			return panics(`f\x : too many iterations`)
		}
	}
}

func splitS(sep S, x V) V {
	r := splitN(-1, sep, x)
	if r.IsPanic() {
		return ppanic("s/x : x ", r)
	}
	return r
}

func scan3(ctx *Context, args []V) V {
	f := args[2]
	if !f.IsFunction() {
		switch fv := f.value.(type) {
		case S:
			return splitNS(args[1], fv, args[0])
		default:
			return panicType("x f'y", "f", f)
		}
	}
	if f.Rank(ctx) == 1 {
		return doWhiles(ctx, args)
	}
	y := args[0]
	x := args[1]
	return scanxfy(ctx, x, f, y)
}

func scanxfy(ctx *Context, x, f, y V) V {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vAdd:
			return scan3vAdd(x, y)
		case vSubtract:
			return scan3vSubtract(x, y)
		case vMax:
			return scan3vMax(x, y)
		case vMin:
			return scan3vMin(x, y)
		}
	}
	switch yv := y.value.(type) {
	case *Dict:
		return newDictValues(yv.keys, scanxfy(ctx, x, f, NewV(yv.values)))
	case array:
		if yv.Len() == 0 {
			return y
		}
		f.IncrRC()
		ctx.pushNoRC(V{})
		r := make([]V, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			ctx.replaceTop(yv.at(i))
			ctx.push(x)
			x = f.applyN(ctx, 2)
			if x.IsPanic() {
				f.DecrRC()
				ctx.drop()
				rcdecrArgs(r)
				return x
			}
			r[i] = x
			x.IncrRC()
		}
		f.DecrRC()
		ctx.drop()
		return normalizedArgs(r)
	default:
		ctx.push(y)
		ctx.push(x)
		r := f.applyN(ctx, 2)
		ctx.drop()
		return r
	}
}

func scanN(ctx *Context, args []V) V {
	f := args[len(args)-1]
	if !f.IsFunction() {
		return panicType(`f[x;y;...]`, "f", f)
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return panicRankN(`f\[x;y;...]`, "f", f.Rank(ctx), n)
	}
	mlen, err := getIterLen(args[:len(args)-2])
	if err != nil {
		return Panicf("f\\[x;y;...] : %v", err)
	}
	if mlen == -1 {
		return toArray(ctx.ApplyN(f, args[:len(args)-1]))
	}
	x := args[len(args)-2]
	if mlen == 0 {
		return protoV(args[len(args)-3])
	}
	f.IncrRC()
	ctx.pushNoRC(V{})
	r := make([]V, mlen)
	for i := 0; i < mlen; i++ {
		ctx.replaceTop(args[0].at(i))
		for j := 1; j < len(args)-2; j++ {
			ctx.push(args[j].at(i))
		}
		ctx.push(x)
		next := f.applyN(ctx, n)
		if next.IsPanic() {
			f.DecrRC()
			ctx.drop()
			rcdecrArgs(r)
			return next
		}
		x = next
		r[i] = x
		x.IncrRC()
	}
	f.DecrRC()
	ctx.drop()
	if d, ok := args[len(args)-3].value.(*Dict); ok {
		return newDictValues(d.keys, normalizedArgs(r))
	}
	return normalizedArgs(r)
}

func splitNS(x V, sep S, y V) V {
	var n int
	if x.IsI() {
		n = int(x.I())
	} else if x.IsF() {
		if !isI(x.F()) {
			return Panicf("i s/y : i non-integer (%g)", x.F())
		}
		n = int(x.F())
	} else {
		return panicType("i s/y", "i", x)
	}
	r := splitN(n, sep, y)
	if r.IsPanic() {
		return ppanic("i s/y : y ", r)
	}
	return r
}

func doWhiles(ctx *Context, args []V) V {
	f := args[2]
	x := args[1]
	y := args[0]
	if x.IsI() {
		return dosTimes(ctx, x.I(), f, y)
	}
	if x.IsF() {
		if !isI(x.F()) {
			return Panicf("n f\\y : non-integer n (%g)", x.F())
		}
		return dosTimes(ctx, int64(x.F()), f, y)
	}
	if x.IsFunction() {
		r := []V{y}
		f.IncrRC()
		x.IncrRC()
		ctx.push(y)
		for {
			cond := x.applyN(ctx, 1)
			if cond.IsPanic() {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				rcdecrArgs(r)
				return cond
			}
			if !cond.IsTrue() {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				return normalizedArgs(r)
			}
			y = f.applyN(ctx, 1)
			if y.IsPanic() {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				rcdecrArgs(r)
				return y
			}
			r = append(r, y)
			y.IncrRC()
			ctx.replaceTop(y)
		}
	}
	return panicType("x f\\y", "x", x)
}

func dosTimes(ctx *Context, n int64, f, y V) V {
	r := make([]V, n+1)
	r[0] = y
	y.IncrRC()
	f.IncrRC()
	ctx.push(y)
	for i := int64(1); i <= n; i++ {
		y = f.applyN(ctx, 1)
		if y.IsPanic() {
			f.DecrRC()
			ctx.drop()
			rcdecrArgs(r)
			return y
		}
		r[i] = y
		y.IncrRC()
		ctx.replaceTop(y)
	}
	f.DecrRC()
	ctx.drop()
	return normalizedArgs(r)
}

func each2(ctx *Context, f, x V) V {
	if !f.IsFunction() {
		return panicType(`f'x`, "f", f)
	}
	switch xv := x.value.(type) {
	case *Dict:
		return newDictValues(xv.keys, eachfx(ctx, f, xv.values))
	case array:
		return eachfx(ctx, f, xv)
	default:
		return ctx.Apply(f, x)
	}
}

func eachfx(ctx *Context, f V, x array) V {
	if f.kind == valVariadic {
		switch f.variadic() {
		case vShape:
			return each2String(ctx, x)
		case vTake:
			return each2Length(x)
		case vMultiply:
			return each2First(x)
		case vApply:
			return each2Type(x)
		}
	}
	r := make([]V, x.Len())
	f.IncrRC()
	ctx.pushNoRC(V{})
	for i := range r {
		ctx.replaceTop(x.at(i))
		next := f.applyN(ctx, 1)
		if next.IsPanic() {
			f.DecrRC()
			ctx.drop()
			rcdecrArgs(r)
			return next
		}
		r[i] = next
		next.IncrRC()
	}
	f.DecrRC()
	ctx.drop()
	return normalizedArgs(r)
}

func eachN(ctx *Context, args []V) V {
	f := args[len(args)-1]
	if !f.IsFunction() {
		return panicType(`f'[x;y;...]`, "f", f)
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return panicRankN(`f'[x;y;...]`, "f", f.Rank(ctx), n)
	}
	mlen, err := getIterLen(args[:len(args)-1])
	if err != nil {
		return Panicf("f'[x;y;...] : %v", err)
	}
	if mlen == -1 {
		return ctx.ApplyN(f, args[:len(args)-1])
	}
	y := args[0]
	if n == 2 && f.kind == valVariadic && f.variadic() == vMatch {
		return each3Match(args[1], y)
	}
	r := make([]V, mlen)
	f.IncrRC()
	ctx.pushNoRC(V{})
	for i := range r {
		ctx.replaceTop(y.at(i))
		for j := 1; j < len(args)-1; j++ {
			ctx.push(args[j].at(i))
		}
		next := f.applyN(ctx, n)
		if next.IsPanic() {
			f.DecrRC()
			ctx.drop()
			rcdecrArgs(r)
			return next
		}
		r[i] = next
		next.IncrRC()
	}
	f.DecrRC()
	ctx.drop()
	if d, ok := args[len(args)-2].value.(*Dict); ok {
		return newDictValues(d.keys, normalizedArgs(r))
	}
	return normalizedArgs(r)
}

func (x V) at(i int) V {
	switch xv := x.value.(type) {
	case *Dict:
		return xv.values.at(i)
	case array:
		return xv.at(i)
	default:
		return x
	}
}

func normalizedArgs(r []V) V {
	ra, ok := normalize(&AV{elts: r})
	if !ok {
		rcdecrArgs(r)
	}
	return NewV(ra)
}
