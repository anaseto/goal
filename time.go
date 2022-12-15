package goal

import (
	"fmt"
	"time"
)

// VTime implements the time variadic verb.
func VTime(ctx *Context, args []V) V {
	var x V
	if len(args) == 1 {
		x = args[0]
	} else {
		x = args[len(args)-2]
	}
	var cmd string
	switch xv := x.value.(type) {
	case S:
		cmd = string(xv)
	default:
		return Panicf("time[[t;]cmd[;...]] : non-string cmd (%s)", x.Type())
	}
	if len(args) == 1 {
		r := timem(time.Now(), cmd)
		if r.IsPanic() {
			return Panicf("time x : %v", r)
		}
		return r
	}
	t, err := parseTime(args[len(args)-1])
	if err != nil {
		return Panicf("time[t;cmd[;...]] : %v", err)
	}
	if len(args) > 2 {
		return panics("time[t;cmd;...] : more than two arguments (NYI)")
	}
	r := timem(t, cmd)
	if r.IsPanic() {
		return Panicf("time[t;cmd[;...]] : %v", r)
	}
	return r
}

func parseTime(x V) (time.Time, error) {
	if x.IsI() {
		return time.Unix(x.I(), 0), nil
	}
	if x.IsF() {
		if !isI(x.F()) {
			return time.Time{}, fmt.Errorf("time x non-integer (%g)", x.F())
		}
		return parseTime(NewI(int64(x.F())))
	}
	switch xv := x.value.(type) {
	case S:
		return time.Parse(time.RFC3339, string(xv))
	default:
		return time.Time{}, fmt.Errorf("bad type for time x (%s)", x.Type())
	}
}

func timem(t time.Time, cmd string) V {
	switch cmd {
	case "":
		return NewS(t.Format(time.RFC3339))
	case "day":
		return NewI(int64(t.Day()))
	case "date":
		y, m, d := t.Date()
		return NewAI([]int64{int64(y), int64(m), int64(d)})
	case "clock":
		h, m, s := t.Clock()
		return NewAI([]int64{int64(h), int64(m), int64(s)})
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
	case "unixmilli":
		return NewI(t.UnixMilli())
	case "unixmicro":
		return NewI(t.UnixMicro())
	case "unixnano":
		return NewI(t.UnixNano())
	case "year":
		return NewI(int64(t.Year()))
	case "yearday":
		return NewI(int64(t.YearDay()))
	case "weekday":
		return NewI(int64(t.Weekday()))
	default:
		return panics("unknown command")
	}
}
