package os

import (
	"os"
	"strings"

	"codeberg.org/anaseto/goal"
)

// environ returns a dictionary with the environment.
func environ() goal.V {
	env := os.Environ()
	ss := make([]string, len(env)*2)
	for i, s := range env {
		b, a, _ := strings.Cut(s, "=")
		ss[i] = b
		ss[i+len(env)] = a
	}
	keys := &goal.AS{Slice: ss[:len(env)]}
	values := &goal.AS{Slice: ss[len(env):]}
	var n int = 2
	keys.InitWithRC(&n)
	values.InitWithRC(&n)
	return goal.NewDict(goal.NewV(keys), goal.NewV(values))
}

// VEnv implements the os.env dyad.
//
// env s : retrieve environment variable s, or return an error if unset. As
// a special case, "" returns a dictionary representing the whole environment.
//
// x env s : sets the value of the environment variable x to s. It returns
// a true value of success, and an error otherwise. Also, the special form
// env[x;0] unsets a variable, and as a special case, env["";0] clears the
// whole environment.
func VEnv(ctx *goal.Context, args []goal.V) goal.V {
	x := args[len(args)-1]
	name, ok := x.Value().(goal.S)
	switch len(args) {
	case 1:
		if !ok {
			return goal.Panicf("env s : s not a string (%s)", x.Type())
		}
		if name == "" {
			return environ()
		}
		s, ok := os.LookupEnv(string(name))
		if !ok {
			return goal.Errorf("variable %s is unset", name)
		}
		return goal.NewS(s)
	case 2:
		if !ok {
			return goal.Panicf("x env s : x not a string (%s)", x.Type())
		}
		y := args[0]
		s, ok := y.Value().(goal.S)
		if !ok {
			if y.IsI() && y.I() == 0 || y.IsF() && y.F() == 0 {
				if name == "" {
					os.Clearenv()
					return goal.NewI(1)
				}
				err := os.Unsetenv(string(name))
				if err != nil {
					return goal.Errorf("%v", err)
				}
				return goal.NewI(1)
			}
			return goal.Panicf("n env s : s not a string (%s)", y.Type())
		}
		err := os.Setenv(string(name), string(s))
		if err != nil {
			return goal.Errorf("%v", err)
		}
		return goal.NewI(1)
	default:
		return goal.Panicf("env : too many arguments (%d)", len(args))
	}
}
