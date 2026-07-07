// Package aria_proptypes ports eslint-plugin-jsx-a11y's `aria-proptypes`
// rule. The rule flags any recognized ARIA state / property whose literal
// value disagrees with the type declared in `aria-query`'s `ariaPropsMap`
// (e.g. `aria-hidden="yes"` — `aria-hidden` is a boolean, "yes" is not).
//
// Upstream signature: no options — schema is `generateObjSchema()` (an
// empty object).
//
// Trigger order (mirrors upstream's JSXAttribute listener exactly):
//
//  1. Attribute name normalized to lowercase. If it doesn't start with
//     `aria-` or isn't in aria-query's map, skip.
//  2. `getPropValue(attribute) == null` — skip. Catches absent values,
//     explicit `{null}` / `{undefined}` (and TS-wrapped variants), and
//     any expression whose static value resolves to null/undefined.
//  3. `getLiteralPropValue(attribute)` — fetch the literal-typed value.
//  4. If the literal value is the upstream-null sentinel (Identifier
//     non-undefined / Call / Member / Conditional / Logical / Binary /
//     TS-wrapper kinds — all noop in LITERAL_TYPES → null), skip. The
//     rule only inspects values it can resolve statically.
//  5. Run the type-specific validity check. For `boolean`, the value
//     must be a JS boolean; for `string` / `id`, a JS string; for
//     `integer` / `number`, a non-boolean coercible to a non-NaN
//     number; for `tristate`, a boolean OR the literal string "mixed";
//     for `token`, the value (lowercased if string) must be in the
//     permitted-values list; for `tokenlist`, the value must be a
//     space-separated list where every space-delimited token (lowercased)
//     is in the permitted-values list; for `idlist`, the value must be
//     a JS string (split-and-recurse always succeeds since each
//     space-delimited slice is still a string).
//  6. If the check fails AND the property doesn't permit `undefined` via
//     `allowUndefined`, emit a type-specific diagnostic.
//
// The diagnostic is reported on the JsxAttribute node (matches upstream's
// `context.report({ node: attribute, ... })`).
package aria_proptypes

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage(name, type, permittedValues)`
// byte-for-byte. The `${permittedValues}` interpolation in the JS template
// literal hits Array.prototype.toString, which joins with a bare comma (no
// space). Booleans stringify as "true" / "false". We preserve order.
//
// "a integer" / "a number" / "a boolean" / "a string" are upstream's literal
// strings — note the (grammatically incorrect) "a integer" / "a undefined".
// Reproduced verbatim so a future audit can diff against upstream exactly.
func errorMessage(name, propType string, permittedValues []any) string {
	switch propType {
	case "tristate":
		return "The value for " + name + ` must be a boolean or the string "mixed".`
	case "token":
		return "The value for " + name + " must be a single token from the following: " + joinPermittedValues(permittedValues) + "."
	case "tokenlist":
		return "The value for " + name + " must be a list of one or more tokens from the following: " + joinPermittedValues(permittedValues) + "."
	case "idlist":
		return "The value for " + name + " must be a list of strings that represent DOM element IDs (idlist)"
	case "id":
		return "The value for " + name + " must be a string that represents a DOM element ID"
	}
	// boolean / string / integer / number / default — upstream's switch
	// fall-through collapses into the same template.
	return "The value for " + name + " must be a " + propType + "."
}

// joinPermittedValues mirrors `Array.prototype.toString` on the
// permitted-values array: bare-comma join, booleans stringified as
// "true" / "false" — matching upstream's `${permittedValues}` template
// interpolation.
func joinPermittedValues(values []any) string {
	parts := make([]string, len(values))
	for i, v := range values {
		switch x := v.(type) {
		case string:
			parts[i] = x
		case bool:
			if x {
				parts[i] = "true"
			} else {
				parts[i] = "false"
			}
		}
	}
	return strings.Join(parts, ",")
}

// validityCheck mirrors upstream's `validityCheck(value, expectedType,
// permittedValues)`. Returns true when the literal value satisfies the
// ARIA type contract.
//
// Implementation notes:
//
//   - `boolean`: typeof === 'boolean'. Maps to AriaLiteralBool.
//   - `string` / `id`: typeof === 'string'. Maps to AriaLiteralString.
//     Note `id` adds no validation beyond "is string" — the actual DOM
//     ID check is a runtime concern.
//   - `tristate`: typeof === 'boolean' || value === 'mixed' (case-sensitive
//     against the bare string "mixed"). aria-checked / aria-pressed.
//   - `integer` / `number`: typeof !== 'boolean' && !isNaN(Number(value)).
//     The boolean exclusion is intentional — JS coerces `true`/`false` to
//     `1`/`0` which would otherwise satisfy `!isNaN(Number(...))` and
//     trivially pass for any boolean-typed value. Number coercion itself
//     is delegated to `jsxa11yutil.AriaLiteralValueAsJSNumber`, the same
//     engine that powers tabindex / tabindex-no-positive — single source
//     of truth for ECMA-262 StringToNumber semantics (trim, hex / oct /
//     bin prefixes, empty → 0, BigInt → float64 cast).
//   - `token`: permittedValues.indexOf(typeof === 'string' ? value.toLowerCase()
//     : value) > -1. Strings are lowercased; booleans / others match by
//     strict equality. aria-current / aria-haspopup / aria-invalid have
//     heterogeneous lists with booleans intermixed.
//   - `tokenlist`: typeof === 'string' && value.split(' ').every(tok →
//     permittedValues.indexOf(tok.toLowerCase()) > -1). Empty strings or
//     unknown tokens fail. aria-dropeffect / aria-relevant.
//   - `idlist`: typeof === 'string' && value.split(' ').every(tok → typeof
//     tok === 'string'). The inner check is trivially true for a string
//     split — collapsed to "value is a string". aria-controls / aria-describedby
//     / aria-flowto / aria-labelledby / aria-owns.
func validityCheck(value jsxa11yutil.AriaLiteralValue, expectedType string, permittedValues []any) bool {
	switch expectedType {
	case "boolean":
		return value.Kind == jsxa11yutil.AriaLiteralBool
	case "string", "id":
		return value.Kind == jsxa11yutil.AriaLiteralString
	case "tristate":
		if value.Kind == jsxa11yutil.AriaLiteralBool {
			return true
		}
		return value.Kind == jsxa11yutil.AriaLiteralString && value.Str == "mixed"
	case "integer", "number":
		if value.Kind == jsxa11yutil.AriaLiteralBool {
			return false
		}
		// Reuse jsxa11yutil's JS Number() implementation — same engine
		// used by tabindex / tabindex-no-positive for string/number/bigint
		// coercion. The boolean returned encodes `!isNaN(Number(value))`
		// so we don't need a separate NaN check.
		_, ok := jsxa11yutil.AriaLiteralValueAsJSNumber(value)
		return ok
	case "token":
		return tokenMatch(value, permittedValues)
	case "tokenlist":
		if value.Kind != jsxa11yutil.AriaLiteralString {
			return false
		}
		for _, tok := range strings.Split(value.Str, " ") {
			lower := strings.ToLower(tok)
			if !permittedValuesContainsString(permittedValues, lower) {
				return false
			}
		}
		return true
	case "idlist":
		// `value.split(' ').every(token => typeof token === 'string')` — since
		// String.split always yields strings, this collapses to "value is a
		// string". Mirrors upstream's recursive `validityCheck(token, 'id', [])`
		// which itself only checks typeof === 'string'.
		return value.Kind == jsxa11yutil.AriaLiteralString
	}
	// Unrecognized type tag — upstream returns false. Defensive parity;
	// AriaPropertyDefinitions has no other tags.
	return false
}

// tokenMatch mirrors upstream's `permittedValues.indexOf(typeof value ===
// 'string' ? value.toLowerCase() : value) > -1`. Strings are lowercased
// before lookup; non-strings (booleans) match the permitted list by
// strict equality.
//
// The lookup uses JS `===` semantics: string-vs-bool returns false even
// when textually similar (e.g. value=`true` (string) does NOT match the
// boolean `true` in the permitted list — but jsx-ast-utils' Literal
// extractor coerces case-insensitive "true"/"false" strings to actual
// booleans first, so the typical `aria-haspopup="true"` flow still
// matches the boolean entry).
func tokenMatch(value jsxa11yutil.AriaLiteralValue, permittedValues []any) bool {
	switch value.Kind {
	case jsxa11yutil.AriaLiteralString:
		lower := strings.ToLower(value.Str)
		return permittedValuesContainsString(permittedValues, lower)
	case jsxa11yutil.AriaLiteralBool:
		for _, v := range permittedValues {
			if b, ok := v.(bool); ok && b == value.Bool {
				return true
			}
		}
		return false
	}
	// Numbers / BigInts / undef / no-lit — never appear in any ARIA
	// permitted-values list. Upstream's indexOf returns -1 for these.
	return false
}

// permittedValuesContainsString reports whether the (already-lowercased)
// candidate string is present in the heterogeneous permitted-values list.
// Booleans in the list are silently skipped — they only match jvBool
// candidates via tokenMatch's separate arm.
func permittedValuesContainsString(values []any, candidate string) bool {
	for _, v := range values {
		if s, ok := v.(string); ok && s == candidate {
			return true
		}
	}
	return false
}

var AriaProptypesRule = rule.Rule{
	Name: "jsx-a11y/aria-proptypes",
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				rawName := reactutil.GetJsxPropName(attr)
				// Upstream normalizes to lowercase BEFORE the prefix check
				// and the aria-query lookup. `ARIA-HIDDEN`, `Aria-Hidden`
				// all resolve to `aria-hidden` for purposes of this rule.
				// This differs from jsx-a11y/aria-props which uses a
				// CASE-SENSITIVE prefix check — same plugin, different gate.
				normalized := strings.ToLower(rawName)
				if !strings.HasPrefix(normalized, "aria-") {
					return
				}
				def, ok := jsxa11yutil.AriaPropertyDefinitions[normalized]
				if !ok {
					return
				}

				// step-1: `getPropValue(attribute) == null` — loose equality
				// with null catches both null and undefined static values.
				// PropValueIsNullish mirrors the full semantic, including
				// the boolean attribute form (returns false → don't gate),
				// explicit `{null}` / `{undefined}` (returns true → gate),
				// and unrecognized expressions (returns true → gate).
				if jsxa11yutil.PropValueIsNullish(attr) {
					return
				}

				// step-3 setup: fetch the literal-typed value.
				value := jsxa11yutil.LiteralPropAriaValue(attr)

				// step-3: `if (value === null) return;` — gates the
				// LITERAL_TYPES noop path (Identifier non-undefined / Call /
				// Member / Conditional / Logical / Binary / TS-wrappers / JSX).
				// Upstream skips because it cannot statically verify the
				// runtime value.
				if value.Kind == jsxa11yutil.AriaLiteralNoLit {
					return
				}

				// step-4 & 5: validity check OR allowUndefined exemption.
				if validityCheck(value, def.Type, def.Values) {
					return
				}
				if def.AllowUndefined && value.Kind == jsxa11yutil.AriaLiteralUndef {
					// Upstream preserves this branch even though step-1
					// usually short-circuits on explicit `{undefined}` via
					// the `getPropValue == null` check. Mirrored for
					// byte-for-byte parity in case future jsx-ast-utils
					// changes break that assumption.
					return
				}

				ctx.ReportNode(attr, rule.RuleMessage{
					Id:          "invalidAriaPropType",
					Description: errorMessage(rawName, def.Type, def.Values),
				})
			},
		}
	},
}
