package goal

import (
	"fmt"
	"strings"
)

func fold2(ctx *Context, args []V) V {
	f := args[1]
	switch f.kind {
	case valVariadic:
		switch f.variadic() {
		case vAdd:
			return fold2vAdd(args[0])
		case vMax:
			return fold2vMax(args[0])
		case vMin:
			return fold2vMin(args[0])
		}
	}
	if !f.IsFunction() {
		if f.IsI() {
			return decode(f, args[0])
		}
		if f.IsF() {
			return decode(f, args[0])
		}
		switch fv := f.value.(type) {
		case S:
			return joinS(fv, args[0])
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
	switch xv := x.value.(type) {
	case array:
		if xv.Len() == 0 {
			if f.kind == valVariadic {
				return f.variadic().zero()
			}
			return NewI(0)
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

func joinS(sep S, x V) V {
	switch xv := x.value.(type) {
	case S:
		return x
	case *AS:
		return NewS(strings.Join([]string(xv.Slice), string(sep)))
	case *AV:
		return Panicf("s/x : x not a string array (%s)", x.Type())
	default:
		return Panicf("s/x : x not a string array (%s)", x.Type())
	}
}

const maxConvergeIters = 1_000_000

func converge(ctx *Context, f, x V) V {
	n := 0
	f.IncrRC()
	ctx.push(x)
	for {
		x.IncrRC()
		r := f.applyN(ctx, 1)
		x.DecrRC()
		if r.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return r
		}
		if Match(r, x) {
			f.DecrRC()
			ctx.drop()
			return r
		}
		ctx.replaceTop(r)
		x = r
		n++
		if n > maxConvergeIters {
			ctx.drop()
			return panics("f/x : too many iterations")
		}
	}
}

func fold3(ctx *Context, args []V) V {
	f := args[2]
	if !f.IsFunction() {
		return Panicf("x F/y : F not a function (%s)", f.Type())
	}
	rank := f.Rank(ctx)
	if rank == 1 {
		return doWhile(ctx, args)
	}
	if rank == 2 {
		return foldSeedfy(ctx, f, args[1], args[0])
	}
	return Panicf("x F/y : F expects %d arguments, but got %d", rank, len(args)-1)
}

func foldSeedfy(ctx *Context, f, x, y V) V {
	switch yv := y.value.(type) {
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
		case array:
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
		return Panicf("f/[x;y;...] : f not a function (%s)", f.Type())
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return Panicf("f/[x;y;...] : f expects %d arguments, but got %d", f.Rank(ctx), n)
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
			if !isTrue(cond) {
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
	switch xv := x.value.(type) {
	case array:
		if xv.Len() == 0 {
			return NewAV(nil)
		}
		r := make([]V, xv.Len())
		r[0] = xv.at(0)
		f.IncrRC()
		ctx.pushNoRC(V{})
		for i := 1; i < xv.Len(); i++ {
			ctx.replaceTop(xv.at(i))
			last := r[i-1]
			ctx.push(last)
			last.IncrRC()
			next := f.applyN(ctx, 2)
			last.DecrRC()
			if next.IsPanic() {
				f.DecrRC()
				ctx.drop()
				return next
			}
			r[i] = next
		}
		f.DecrRC()
		ctx.drop()
		return Canonical(NewAV(r))
	default:
		return x
	}
}

func converges(ctx *Context, f, x V) V {
	n := 0
	r := []V{}
	defer func() {
		for _, ri := range r {
			ri.DecrRC()
		}
	}()
	f.IncrRC()
	ctx.push(x)
	for {
		x.IncrRC()
		r = append(r, x)
		y := f.applyN(ctx, 1)
		if y.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return y
		}
		if Match(y, x) {
			f.DecrRC()
			ctx.drop()
			return Canonical(NewAV(r))
		}
		ctx.replaceTop(y)
		x = y
		n++
		if n > maxConvergeIters {
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
	switch yv := y.value.(type) {
	case array:
		if yv.Len() == 0 {
			return NewAV(nil)
		}
		f.IncrRC()
		ctx.pushNoRC(V{})
		r := make([]V, yv.Len())
		for i := 0; i < yv.Len(); i++ {
			ctx.replaceTop(yv.at(i))
			ctx.push(x)
			x.IncrRC()
			next := f.applyN(ctx, 2)
			x.DecrRC()
			if next.IsPanic() {
				f.DecrRC()
				ctx.drop()
				return next
			}
			x = next
			r[i] = x
		}
		f.DecrRC()
		ctx.drop()
		return Canonical(NewAV(r))
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
		return Panicf("f\\[x;y;...] : f not a function (%s)", f.Type())
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return Panicf("f\\[x;y;...] : f expects %d arguments, but got %d", f.Rank(ctx), n)
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
		return NewAV(nil)
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
		x.IncrRC()
		next := f.applyN(ctx, n)
		x.DecrRC()
		if next.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return next
		}
		x = next
		r[i] = x
	}
	f.DecrRC()
	ctx.drop()
	return Canonical(NewAV(r))
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
		return Panicf("i s/y : i bad type (%s)", x.Type())
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
			y.IncrRC()
			cond := x.applyN(ctx, 1)
			y.DecrRC()
			if cond.IsPanic() {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				return cond
			}
			if !isTrue(cond) {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				return Canonical(NewAV(r))
			}
			y = f.applyN(ctx, 1)
			if y.IsPanic() {
				f.DecrRC()
				x.DecrRC()
				ctx.drop()
				return y
			}
			ctx.replaceTop(y)
			r = append(r, y)
		}
	}
	return panicType("x f\\y", "x", x)
}

func dosTimes(ctx *Context, n int64, f, y V) V {
	r := make([]V, n+1)
	r[0] = y
	f.IncrRC()
	ctx.push(y)
	for i := int64(1); i <= n; i++ {
		y = f.applyN(ctx, 1)
		if y.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return y
		}
		ctx.replaceTop(y)
		r[i] = y
	}
	f.DecrRC()
	ctx.drop()
	return Canonical(NewAV(r))
}

func each2(ctx *Context, f, x V) V {
	if !f.IsFunction() {
		return Panicf("f'x : f not a function (%s)", f.Type())
	}
	xv, ok := x.value.(array)
	if !ok {
		return ctx.Apply(f, x)
	}
	r := make([]V, xv.Len())
	f.IncrRC()
	ctx.pushNoRC(V{})
	for i := range r {
		ctx.replaceTop(xv.at(i))
		next := f.applyN(ctx, 1)
		if next.IsPanic() {
			f.DecrRC()
			ctx.drop()
			return next
		}
		r[i] = next
	}
	f.DecrRC()
	ctx.drop()
	return Canonical(NewAV(r))
}

func eachN(ctx *Context, args []V) V {
	f := args[len(args)-1]
	if !f.IsFunction() {
		return Panicf("f'[x;y;...] : f not a function (%s)", f.Type())
	}
	n := len(args) - 1
	if f.Rank(ctx) != n {
		return Panicf("f'[x;y;...] : f expects %d arguments, but got %d", f.Rank(ctx), n)
	}
	mlen, err := getIterLen(args[:len(args)-1])
	if err != nil {
		return Panicf("f'[x;y;...] : %v", err)
	}
	if mlen == -1 {
		return ctx.ApplyN(f, args[:len(args)-1])
	}
	y := args[0]
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
			return next
		}
		r[i] = next
	}
	f.DecrRC()
	ctx.drop()
	return Canonical(NewAV(r))
}

func (x V) at(i int) V {
	switch xv := x.value.(type) {
	case array:
		return xv.at(i)
	default:
		return x
	}
}
