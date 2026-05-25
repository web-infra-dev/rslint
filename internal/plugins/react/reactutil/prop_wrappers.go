package reactutil

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
)

// PropWrapperEntry encodes one entry of `settings.propWrapperFunctions`. The
// raw entries can be either a bare string (`"forbidExtraProps"`) or an
// `{object, property}` pair (`{ object: "Object", property: "assign" }` →
// matches `Object.assign(...)`). Both shapes are normalized to this struct.
type PropWrapperEntry struct {
	// Object is the receiver portion of a member-call wrapper (e.g.
	// `"Object"` for `Object.assign`). Empty for bare-identifier wrappers.
	Object string
	// Property is the function name (e.g. `"assign"` for `Object.assign`,
	// or `"forbidExtraProps"` for a bare-identifier wrapper).
	Property string
}

// GetPropWrapperFunctions reads `settings.propWrapperFunctions` from the
// rslint config and returns the parsed entries. Unknown shapes (a non-array
// value, an entry that's neither a string nor a `{object, property}` map,
// an entry with empty `property`) are silently dropped — this matches
// eslint-plugin-react's `propWrapperUtil` permissive parsing.
func GetPropWrapperFunctions(settings map[string]interface{}) []PropWrapperEntry {
	v, ok := settings["propWrapperFunctions"]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var out []PropWrapperEntry
	for _, entry := range arr {
		switch t := entry.(type) {
		case string:
			if t != "" {
				if dot := strings.IndexByte(t, '.'); dot > 0 && dot < len(t)-1 {
					// Allow `"Object.assign"` style strings (legacy upstream
					// shape) by splitting on the first dot.
					out = append(out, PropWrapperEntry{Object: t[:dot], Property: t[dot+1:]})
				} else {
					out = append(out, PropWrapperEntry{Property: t})
				}
			}
		case map[string]interface{}:
			obj, _ := t["object"].(string)
			prop, _ := t["property"].(string)
			if prop == "" {
				continue
			}
			out = append(out, PropWrapperEntry{Object: obj, Property: prop})
		}
	}
	return out
}

// IsPropWrapperCall reports whether `call` is a CallExpression whose callee
// matches one of the user-configured `propWrapperFunctions` entries.
//
// Supports:
//   - bare identifier callees: `forbidExtraProps(...)`, `merge(...)` —
//     match an entry with empty `Object`.
//   - dotted-property callees: `Object.assign(...)`, `_.assign(...)` —
//     match an entry whose `Object` and `Property` both equal the receiver
//     and method names respectively.
//
// `call` may be wrapped in parens / TS expression wrappers; the callee is
// unwrapped via `SkipExpressionWrappers`. Anything more complex (computed
// access, optional-chain wrappers around the callee head) is treated as
// not matching.
func IsPropWrapperCall(call *ast.Node, wrappers []PropWrapperEntry) bool {
	if len(wrappers) == 0 || call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callee := SkipExpressionWrappers(call.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		name := callee.AsIdentifier().Text
		for _, w := range wrappers {
			if w.Object == "" && w.Property == name {
				return true
			}
		}
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := SkipExpressionWrappers(pa.Expression)
		nameNode := pa.Name()
		if obj.Kind != ast.KindIdentifier || nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		objText := obj.AsIdentifier().Text
		propText := nameNode.AsIdentifier().Text
		for _, w := range wrappers {
			if w.Object == objText && w.Property == propText {
				return true
			}
		}
	}
	return false
}
