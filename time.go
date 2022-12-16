package goal

import (
	"fmt"
	"strings"
	"time"
)

// VTime implements the time variadic verb.
func VTime(ctx *Context, args []V) V {
	var x V
	x = args[len(args)-1]
	var cmd string
	switch xv := x.value.(type) {
	case S:
		cmd = string(xv)
	default:
		return Panicf("time[cmd;t;format] : non-string cmd (%s)", x.Type())
	}
	if len(args) == 1 {
		r := timem(time.Now(), cmd)
		if r.IsPanic() {
			return Panicf("time x : %v", r)
		}
		return r
	}
	y := args[len(args)-2]
	var t time.Time
	var err error
	switch len(args) {
	case 2:
		t, err = parseTime(y, time.RFC3339, "")
	case 3:
		format, ok := args[len(args)-3].value.(S)
		if !ok {
			return Panicf("time[t;cmd;format] : non-string format (%s)",
				args[len(args)-3].Type())
		}
		t, err = parseTime(y, getFormat(string(format)), "")
	case 4:
		format, ok := args[len(args)-3].value.(S)
		if !ok {
			return Panicf("time[t;cmd;format;loc] : non-string format (%s)",
				args[len(args)-3].Type())
		}
		loc, ok := args[len(args)-4].value.(S)
		if !ok {
			return Panicf("time[t;cmd;format;loc] : non-string location (%s)",
				args[len(args)-4].Type())
		}
		t, err = parseTime(y, getFormat(string(format)), string(loc))
	default:
		return Panicf("time : too many arguments (%d)", len(args))
	}
	if err != nil {
		return Panicf("time[t;cmd;format] : %v", err)
	}
	r := timem(t, cmd)
	if r.IsPanic() {
		return Panicf("time[t;cmd;format] : %v", r)
	}
	return r
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

func parseTime(x V, layout, loc string) (time.Time, error) {
	if x.IsI() {
		return time.Unix(x.I(), 0), nil
	}
	if x.IsF() {
		if !isI(x.F()) {
			return time.Time{}, fmt.Errorf("time x non-integer (%g)", x.F())
		}
		return parseTime(NewI(int64(x.F())), layout, loc)
	}
	switch xv := x.value.(type) {
	case S:
		if loc == "" {
			return time.Parse(layout, string(xv))
		}
		l, err := time.LoadLocation(loc)
		if err != nil {
			return time.Time{}, err
		}
		return time.ParseInLocation(layout, string(xv), l)
	default:
		return time.Time{}, fmt.Errorf("bad type for time x (%s)", x.Type())
	}
}

func timem(t time.Time, cmd string) V {
	switch cmd {
	case "week":
		y, w := t.ISOWeek()
		return NewAI([]int64{int64(y), int64(w)})
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
		cmd = getFormat(cmd)
		if strings.ContainsAny(cmd, " 0123456789-") {
			// TODO: better condition
			return NewS(t.Format(cmd))
		}
		return panics("unknown command")
	}
}
