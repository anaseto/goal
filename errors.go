package goal

import "fmt"

func errNYI(s string) E {
	return E("NYI: " + s)
}

func errType(x V) E {
	return E("bad type: `" + x.Type())
}

func errs(s string) E {
	return E(s)
}

func errsw(s string) E {
	return E("left argument:" + s)
}

func errf(format string, a ...interface{}) E {
	return E(fmt.Sprintf(format, a...))
}
