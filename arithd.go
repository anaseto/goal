package main

// Add returns w+x.
func Add(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return AddI(w, x)
	case F:
		return AddF(w, x)
	case AI:
		return AddAI(w, x)
	case AF:
		return AddAF(w, x)
	case AO:
		return AddAO(w, x)
	}
	// TODO
	return badtype("+")
}

func AddI(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i + x
	case F:
		return F(i) + x
	case AI:
		return AddAII(x, i)
	case AF:
		return AddAFI(x, i)
	}
	// TODO
	return badtype("+")
}

func AddF(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f + F(x)
	case F:
		return f + x
	case AI:
		return AddAIF(x, f)
	case AF:
		return AddAFF(x, f)
	}
	// TODO
	return badtype("+")
}

func AddAI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return AddAII(a, x)
	case F:
		return AddAIF(a, x)
	case AI:
		return AddAIAI(a, x)
	case AF:
		return AddAIAF(a, x)
	case AO:
		return AddAOAI(x, a)
	}
	// TODO
	return badtype("+")
}

func AddAF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return AddAFI(a, x)
	case F:
		return AddAFF(a, x)
	case AI:
		return AddAIAF(x, a)
	case AF:
		return AddAFAF(a, x)
	case AO:
		return AddAOAF(x, a)
	}
	// TODO
	return badtype("+")
}

func AddAO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return AddAOI(a, x)
	case F:
		return AddAOF(a, x)
	case AI:
		return AddAOAI(a, x)
	case AF:
		return AddAOAF(a, x)
	case AO:
		return AddAOAO(a, x)
	}
	// TODO
	return badtype("+")
}

func AddAII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] + i
	}
	return r
}

func AddAIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) + f
	}
	return r
}

func AddAFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] + F(i)
	}
	return r
}

func AddAFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] + f
	}
	return r
}

func AddAIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] + a2[j]
	}
	return r
}

func AddAIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) + a2[j]
	}
	return r
}

func AddAFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] + a2[j]
	}
	return r
}

func AddAOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = AddI(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func AddAOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = AddF(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func AddAOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = AddI(a2[j], a1[j])
	}
	return r
}

func AddAOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = AddF(a2[j], a1[j])
	}
	return r
}

func AddAOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("+")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = Add(a2[j], a1[j])
	}
	return r
}

// Divide returns w/x.
func Divide(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return DivideI(w, x)
	case F:
		return DivideF(w, x)
	case AI:
		return DivideAI(w, x)
	case AF:
		return DivideAF(w, x)
	case AO:
		return DivideAO(w, x)
	}
	// TODO
	return badtype("/")
}

func DivideI(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i / x
	case F:
		return F(i) / x
	case AI:
		return DivideAII(x, i)
	case AF:
		return DivideAFI(x, i)
	}
	// TODO
	return badtype("/")
}

func DivideF(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f / F(x)
	case F:
		return f / x
	case AI:
		return DivideAIF(x, f)
	case AF:
		return DivideAFF(x, f)
	}
	// TODO
	return badtype("/")
}

func DivideAI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return DivideAII(a, x)
	case F:
		return DivideAIF(a, x)
	case AI:
		return DivideAIAI(a, x)
	case AF:
		return DivideAIAF(a, x)
	case AO:
		return DivideAOAI(x, a)
	}
	// TODO
	return badtype("/")
}

func DivideAF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return DivideAFI(a, x)
	case F:
		return DivideAFF(a, x)
	case AI:
		return DivideAIAF(x, a)
	case AF:
		return DivideAFAF(a, x)
	case AO:
		return DivideAOAF(x, a)
	}
	// TODO
	return badtype("/")
}

func DivideAO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return DivideAOI(a, x)
	case F:
		return DivideAOF(a, x)
	case AI:
		return DivideAOAI(a, x)
	case AF:
		return DivideAOAF(a, x)
	case AO:
		return DivideAOAO(a, x)
	}
	// TODO
	return badtype("/")
}

func DivideAII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] / i
	}
	return r
}

func DivideAIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) / f
	}
	return r
}

func DivideAFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] / F(i)
	}
	return r
}

func DivideAFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] / f
	}
	return r
}

func DivideAIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] / a2[j]
	}
	return r
}

func DivideAIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) / a2[j]
	}
	return r
}

func DivideAFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] / a2[j]
	}
	return r
}

func DivideAOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = DivideI(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func DivideAOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = DivideF(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func DivideAOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = DivideI(a2[j], a1[j])
	}
	return r
}

func DivideAOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = DivideF(a2[j], a1[j])
	}
	return r
}

func DivideAOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("/")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = Divide(a2[j], a1[j])
	}
	return r
}

// Multiply returns w*x.
func Multiply(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return MultiplyI(w, x)
	case F:
		return MultiplyF(w, x)
	case AI:
		return MultiplyAI(w, x)
	case AF:
		return MultiplyAF(w, x)
	case AO:
		return MultiplyAO(w, x)
	}
	// TODO
	return badtype("*")
}

func MultiplyI(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i * x
	case F:
		return F(i) * x
	case AI:
		return MultiplyAII(x, i)
	case AF:
		return MultiplyAFI(x, i)
	}
	// TODO
	return badtype("*")
}

func MultiplyF(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f * F(x)
	case F:
		return f * x
	case AI:
		return MultiplyAIF(x, f)
	case AF:
		return MultiplyAFF(x, f)
	}
	// TODO
	return badtype("*")
}

func MultiplyAI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return MultiplyAII(a, x)
	case F:
		return MultiplyAIF(a, x)
	case AI:
		return MultiplyAIAI(a, x)
	case AF:
		return MultiplyAIAF(a, x)
	case AO:
		return MultiplyAOAI(x, a)
	}
	// TODO
	return badtype("*")
}

func MultiplyAF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return MultiplyAFI(a, x)
	case F:
		return MultiplyAFF(a, x)
	case AI:
		return MultiplyAIAF(x, a)
	case AF:
		return MultiplyAFAF(a, x)
	case AO:
		return MultiplyAOAF(x, a)
	}
	// TODO
	return badtype("*")
}

func MultiplyAO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return MultiplyAOI(a, x)
	case F:
		return MultiplyAOF(a, x)
	case AI:
		return MultiplyAOAI(a, x)
	case AF:
		return MultiplyAOAF(a, x)
	case AO:
		return MultiplyAOAO(a, x)
	}
	// TODO
	return badtype("*")
}

func MultiplyAII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] * i
	}
	return r
}

func MultiplyAIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) * f
	}
	return r
}

func MultiplyAFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] * F(i)
	}
	return r
}

func MultiplyAFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] * f
	}
	return r
}

func MultiplyAIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] * a2[j]
	}
	return r
}

func MultiplyAIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) * a2[j]
	}
	return r
}

func MultiplyAFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] * a2[j]
	}
	return r
}

func MultiplyAOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = MultiplyI(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func MultiplyAOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = MultiplyF(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func MultiplyAOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = MultiplyI(a2[j], a1[j])
	}
	return r
}

func MultiplyAOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = MultiplyF(a2[j], a1[j])
	}
	return r
}

func MultiplyAOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("*")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = Multiply(a2[j], a1[j])
	}
	return r
}

// Subtract returns w-x.
func Subtract(w, x Object) Object {
	switch w := w.(type) {
	case I:
		return SubtractI(w, x)
	case F:
		return SubtractF(w, x)
	case AI:
		return SubtractAI(w, x)
	case AF:
		return SubtractAF(w, x)
	case AO:
		return SubtractAO(w, x)
	}
	// TODO
	return badtype("-")
}

func SubtractI(i I, x Object) Object {
	switch x := x.(type) {
	case I:
		return i - x
	case F:
		return F(i) - x
	case AI:
		return SubtractAII(x, i)
	case AF:
		return SubtractAFI(x, i)
	}
	// TODO
	return badtype("-")
}

func SubtractF(f F, x Object) Object {
	switch x := x.(type) {
	case I:
		return f - F(x)
	case F:
		return f - x
	case AI:
		return SubtractAIF(x, f)
	case AF:
		return SubtractAFF(x, f)
	}
	// TODO
	return badtype("-")
}

func SubtractAI(a AI, x Object) Object {
	switch x := x.(type) {
	case I:
		return SubtractAII(a, x)
	case F:
		return SubtractAIF(a, x)
	case AI:
		return SubtractAIAI(a, x)
	case AF:
		return SubtractAIAF(a, x)
	case AO:
		return SubtractAOAI(x, a)
	}
	// TODO
	return badtype("-")
}

func SubtractAF(a AF, x Object) Object {
	switch x := x.(type) {
	case I:
		return SubtractAFI(a, x)
	case F:
		return SubtractAFF(a, x)
	case AI:
		return SubtractAIAF(x, a)
	case AF:
		return SubtractAFAF(a, x)
	case AO:
		return SubtractAOAF(x, a)
	}
	// TODO
	return badtype("-")
}

func SubtractAO(a AO, x Object) Object {
	switch x := x.(type) {
	case I:
		return SubtractAOI(a, x)
	case F:
		return SubtractAOF(a, x)
	case AI:
		return SubtractAOAI(a, x)
	case AF:
		return SubtractAOAF(a, x)
	case AO:
		return SubtractAOAO(a, x)
	}
	// TODO
	return badtype("-")
}

func SubtractAII(a AI, i I) AI {
	r := make(AI, len(a))
	for j := range r {
		r[j] = a[j] - i
	}
	return r
}

func SubtractAIF(a AI, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = F(a[j]) - f
	}
	return r
}

func SubtractAFI(a AF, i I) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] - F(i)
	}
	return r
}

func SubtractAFF(a AF, f F) AF {
	r := make(AF, len(a))
	for j := range r {
		r[j] = a[j] - f
	}
	return r
}

func SubtractAIAI(a1 AI, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AI, len(a1))
	for j := range r {
		r[j] = a1[j] - a2[j]
	}
	return r
}

func SubtractAIAF(a1 AI, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = F(a1[j]) - a2[j]
	}
	return r
}

func SubtractAFAF(a1 AF, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AF, len(a1))
	for j := range r {
		r[j] = a1[j] - a2[j]
	}
	return r
}

func SubtractAOI(a AO, i I) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = SubtractI(i, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func SubtractAOF(a AO, f F) Object {
	r := make(AO, len(a))
	for j := range r {
		r[j] = SubtractF(f, a[j])
		err, ok := r[j].(E)
		if ok {
			return err
		}
	}
	return r
}

func SubtractAOAI(a1 AO, a2 AI) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = SubtractI(a2[j], a1[j])
	}
	return r
}

func SubtractAOAF(a1 AO, a2 AF) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = SubtractF(a2[j], a1[j])
	}
	return r
}

func SubtractAOAO(a1 AO, a2 AO) Object {
	if len(a1) != len(a2) {
		return badlen("-")
	}
	r := make(AO, len(a1))
	for j := range r {
		r[j] = Subtract(a2[j], a1[j])
	}
	return r
}
