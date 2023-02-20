package os

import (
	"os"
	"strings"

	"codeberg.org/anaseto/goal"
)

// Environ returns a dictionary with the environment. NOTE: may be changed into
// a function in the future.
func Environ() goal.V {
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
