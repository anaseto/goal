package goal

import (
	"testing"
	 "strings"
	 "fmt"
)

func FuzzGeneric(f *testing.F) {
	f.Add("0")
	f.Fuzz(func (t *testing.T, s string) {
		defer func() {
			if r := recover(); r != nil {
				s := fmt.Sprintf("%v", r)
				switch {
				case strings.Contains(s, "makeslice"):
					// For cases when we ask for a
					// too big slice.
				case strings.Contains(s, "out of memory"):
					// actually, typically not possible to catch
				default:
					panic(r)
				}
			}
		}()
		ctx := NewContext()
		if len(s) > 8 {
			// We limit size, to avoid slow or out of memory
			// cases with big numbers.
			s = s[:8]
		}
		if strings.ContainsAny(s, "eEpP!\\//") {
			// We skip inputs susceptible (even with the
			// size limitation) of being too slow or
			// panicking with out of memory.
			return
		}
		ctx.Eval(s)
	})
}
