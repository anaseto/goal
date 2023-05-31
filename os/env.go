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
	keys := goal.NewAS(ss[:len(env)])
	values := goal.NewAS(ss[len(env):])
	return goal.NewD(keys, values)
}

// VFEnv implements the os.env dyad.
//
// env s : retrieve environment variable s, or return an error if unset. As
// a special case, "" returns a dictionary representing the whole environment.
//
// x env s : sets the value of the environment variable x to s. It returns
// a true value of success, and an error otherwise. Also, the special form
// env[x;0] unsets a variable, and as a special case, env["";0] clears the
// whole environment.
func VFEnv(ctx *goal.Context, args []goal.V) goal.V {
	x := args[len(args)-1]
	name, ok := x.BV().(goal.S)
	switch len(args) {
	case 1:
		if !ok {
			return panicType("env s", "s", x)
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
			return panicType("x env s", "x", x)
		}
		y := args[0]
		s, ok := y.BV().(goal.S)
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
			return panicType("x env s", "s", y)
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
