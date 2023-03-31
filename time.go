package goal

import (
	//"fmt"
	"strings"
	"time"
)

// vfTime implements the time variadic verb.
func vfTime(ctx *Context, args []V) V {
	x := args[len(args)-1]
	var cmd string
	switch xv := x.value.(type) {
	case S:
		cmd = string(xv)
	default:
		return panicType("time[cmd;t;format]", "cmd", x)
	}
	if len(args) == 1 {
		r := ftime(cmd, time.Now())
		if r.IsPanic() {
			return Panicf("time x : %v", r)
		}
		return r
	}
	y := args[len(args)-2]
	switch len(args) {
	case 2:
		return doTime(cmd, y, time.RFC3339, "")
	case 3:
		z := args[len(args)-3]
		format, ok := z.value.(S)
		if !ok {
			return panicType("time[cmd;t;format]", "format", z)
		}
		return doTime(cmd, y, getFormat(string(format)), "")
	case 4:
		z := args[len(args)-3]
		format, ok := z.value.(S)
		if !ok {
			return panicType("time[cmd;t;format]", "format", z)
		}
		loc, ok := args[len(args)-4].value.(S)
		if !ok {
			return panicType("time[cmd;t;format;loc]", "loc", args[len(args)-4])
		}
		return doTime(cmd, y, getFormat(string(format)), string(loc))
	default:
		return Panicf("time : too many arguments (%d)", len(args))
	}
}

func getFormat(name string) string {
	switch name {
	case "ANSIC":
		return time.ANSIC
	case "UnixDate":
		return time.UnixDate
	case "RubyDate":
		return time.RubyDate
	case "RFC822":
		return time.RFC822
	case "RFC822Z":
		return time.RFC822Z
	case "RFC850":
		return time.RFC850
	case "RFC1123":
		return time.RFC1123
	case "RFC1123Z":
		return time.RFC1123Z
	case "RFC3339", "":
		return time.RFC3339
	case "RFC3339Nano":
		return time.RFC3339Nano
	case "Kitchen":
		return time.Kitchen
	default:
		return name
	}
}

func doTime(cmd string, y V, layout, loc string) V {
	if y.IsI() {
		return doTimeI(cmd, y.I(), layout)
	}
	if y.IsF() {
		if !isI(y.F()) {
			return Panicf("time[cmd;t;...] : t non-integer number (%g)",
				y.F())
		}
		return doTimeI(cmd, int64(y.F()), layout)
	}
	switch yv := y.value.(type) {
	case *AB:
		return doTime(cmd, fromABtoAI(yv), layout, loc)
	case *AI:
		// doTime: allocations could be optimized depending on cmd.
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := doTimeI(cmd, yi, layout)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return canonicalFast(NewAV(r))
	case *AF:
		y = toAI(yv)
		if y.IsPanic() {
			return y
		}
		return doTime(cmd, y, layout, loc)
	case S:
		return doTimeS(cmd, string(yv), layout, loc)
	case *AS:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := doTimeS(cmd, yi, layout, loc)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return canonicalFast(NewAV(r))
	case *AV:
		r := make([]V, yv.Len())
		for i, yi := range yv.elts {
			ri := doTime(cmd, yi, layout, loc)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return canonicalFast(NewAV(r))
	default:
		return panicType("time[cmd;t;...]", "t", y)
	}
}

func doTimeI(cmd string, yv int64, layout string) V {
	switch layout {
	case "milli":
		return ftime(cmd, time.UnixMilli(yv))
	case "micro":
		return ftime(cmd, time.UnixMicro(yv))
	case "nano":
		return ftime(cmd, time.Unix(yv/1_000_000_000, yv%1_000_000_000))
	default:
		return ftime(cmd, time.Unix(yv, 0))
	}
}

func doTimeS(cmd string, yv string, layout, loc string) V {
	var t time.Time
	var err error
	if loc == "" {
		t, err := time.Parse(layout, string(yv))
		if err != nil {
			return Errorf("%v", err)
		}
		return ftime(cmd, t)
	}
	l, err := time.LoadLocation(loc)
	if err != nil {
		return Errorf("%v", err)
	}
	t, err = time.ParseInLocation(layout, string(yv), l)
	if err != nil {
		return Errorf("%v", err)
	}
	return ftime(cmd, t)
}

func ftime(cmd string, t time.Time) V {
	switch cmd {
	case "clock":
		h, m, s := t.Clock()
		return NewAI([]int64{int64(h), int64(m), int64(s)})
	case "date":
		y, m, d := t.Date()
		return NewAI([]int64{int64(y), int64(m), int64(d)})
	case "day":
		return NewI(int64(t.Day()))
	case "hour":
		return NewI(int64(t.Hour()))
	case "minute":
		return NewI(int64(t.Minute()))
	case "month":
		return NewI(int64(t.Month()))
	case "second":
		return NewI(int64(t.Second()))
	case "unix":
		return NewI(t.Unix())
	case "unixmicro":
		return NewI(t.UnixMicro())
	case "unixmilli":
		return NewI(t.UnixMilli())
	case "unixnano":
		return NewI(t.UnixNano())
	case "week":
		y, w := t.ISOWeek()
		return NewAI([]int64{int64(y), int64(w)})
	case "weekday":
		return NewI(int64(t.Weekday()))
	case "year":
		return NewI(int64(t.Year()))
	case "yearday":
		return NewI(int64(t.YearDay()))
	case "zone":
		zone, seconds := t.Zone()
		return NewAV([]V{NewS(zone), NewI(int64(seconds))})
	default:
		cmd = getFormat(cmd)
		if strings.ContainsAny(cmd, " 0123456789-") {
			// TODO: ftime: better error check
			return NewS(t.Format(cmd))
		}
		return Panicf("time[cmd;...] : unknown command %s", cmd)
	}
}
