package goal

import (
	"encoding/json"
	"strings"
)

func fJSON(x V) V {
	switch xv := x.value.(type) {
	case S:
		return jsonStringToGoal(string(xv))
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			r[i] = jsonStringToGoal(xi)
		}
		return Canonical(NewAV(r))
	case *AV:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := fJSON(xi)
			if ri.IsPanic() {
				return ri
			}
			r[i] = ri
		}
		return NewAV(r)
	default:
		return panicType("json x", "x", x)
	}
}

func jsonStringToGoal(s string) V {
	sr := strings.NewReader(s)
	dec := json.NewDecoder(sr)
	var v any
	err := dec.Decode(&v)
	if err != nil {
		return Errorf("%v", err)
	}
	return jsonToGoal(v)
}

func jsonToGoal(v any) V {
	switch vv := v.(type) {
	case bool:
		return NewI(B2I(vv))
	case float64:
		return NewF(vv)
	case string:
		return NewS(vv)
	case []any:
		r := make([]V, len(vv))
		for i, vi := range vv {
			r[i] = jsonToGoal(vi)
		}
		return Canonical(NewAV(r))
	case map[string]any:
		keys := make([]string, 0, len(vv))
		values := make([]V, 0, len(vv))
		for k, vk := range vv {
			values = append(values, jsonToGoal(vk))
			keys = append(keys, k)
		}
		return NewDict(NewAS(keys), Canonical(NewAV(values)))
	default:
		return NewError(NewS("null"))
	}
}
