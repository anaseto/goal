package main

// Conjugate returns +x.
func Conjugate(x Object) Object {
	switch x := x.(type) {
	case I, F:
		return x
	}
	// TODO: complex values (conjugate)
	return badtype("+")
}

// Add returns w+x.
func Add(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return addI(w, x)
	case F:
		return addF(w, x)
	case AI:
		return addAI(w, x)
	case AF:
		return addAF(w, x)
	case AO:
		return addAO(w, x)
	}
	// TODO
	return badtype("+")
}

func addI(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i + x
	case F:
		return F(i) + x
	case AI:
		return addAII(x, i)
	case AF:
		return addAFI(x, i)
	}
	// TODO
	return badtype("+")
}

func addF(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f + F(x)
	case F:
		return f + x
	case AI:
		return addAIF(x, f)
	case AF:
		return addAFF(x, f)
	}
	// TODO
	return badtype("+")
}

func addAI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return addAII(a, x)
	case F:
		return addAIF(a, x)
	case AI:
		return addAIAI(a, x)
	case AF:
		return addAIAF(a, x)
	case AO:
		return addAOAI(x, a)
	}
	// TODO
	return badtype("+")
}

func addAF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return addAFI(a, x)
	case F:
		return addAFF(a, x)
	case AI:
		return addAIAF(x, a)
	case AF:
		return addAFAF(a, x)
	case AO:
		return addAOAF(x, a)
	}
	// TODO
	return badtype("+")
}

func addAO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return addAOI(a, x)
	case F:
		return addAOF(a, x)
	case AI:
		return addAOAI(a, x)
	case AF:
		return addAOAF(a, x)
	case AO:
		return addAOAO(a, x)
	}
	// TODO
	return badtype("+")
}

func addAII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] + i
	}
	return r
}

func addAIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) + f
	}
	return r
}

func addAFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] + F(i)
	}
	return r
}

func addAFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] + f
	}
	return r
}

func addAIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] + a2[j]
	}
	return r
}

func addAIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) + a2[j]
	}
	return r
}

func addAFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] + a2[j]
	}
	return r
}

func addAOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = addI(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func addAOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = addF(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func addAOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = addI(a2[j], a1[j])
	}
	return r
}

func addAOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = addF(a2[j], a1[j])
	}
	return r
}

func addAOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = Add(a2[j], a1[j])
	}
	return r
}
