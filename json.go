package goal

import (
	"encoding/json"
	"strings"
)

func fJSON(x V) V {
	switch xv := x.bv.(type) {
	case S:
		return jsonStringToGoal(string(xv))
	case *AS:
		r := make([]V, xv.Len())
		for i, xi := range xv.elts {
			ri := jsonStringToGoal(xi)
			ri.MarkImmutable()
			r[i] = ri
		}
		return canonicalVs(r)
	case *AV:
		return monadAV(xv, fJSON)
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
		return NewI(b2I(vv))
	case float64:
		return NewF(vv)
	case string:
		return NewS(vv)
	case []any:
		r := make([]V, len(vv))
		for i, vi := range vv {
			ri := jsonToGoal(vi)
			ri.MarkImmutable()
			r[i] = ri
		}
		return canonicalVs(r)
	case map[string]any:
		keys := make([]string, 0, len(vv))
		values := make([]V, 0, len(vv))
		for k, vk := range vv {
			v := jsonToGoal(vk)
			v.MarkImmutable()
			values = append(values, v)
			keys = append(keys, k)
		}
		return NewDict(NewAS(keys), canonicalVs(values))
	default:
		return NewError(NewS("null"))
	}
}
