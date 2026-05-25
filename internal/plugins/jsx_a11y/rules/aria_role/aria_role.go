// Package aria_role ports eslint-plugin-jsx-a11y's `aria-role` rule. The
// rule flags any literal `role` JSX attribute whose value (after space-
// splitting) contains a token that is not a recognized non-abstract ARIA
// role and is not in the user's `allowedInvalidRoles` allow-list.
//
// Upstream signature: one optional object option:
//
//   - `allowedInvalidRoles`: string[] — additional role names accepted
//     alongside the aria-query non-abstract set.
//   - `ignoreNonDOM`: boolean — when true, skip the rule for custom-component
//     JSX elements (anything `aria-query`'s `dom` map doesn't recognize).
//     Default false.
//
// Trigger order (mirrors upstream's JSXAttribute listener exactly):
//
//  1. If `ignoreNonDOM` is set, resolve the parent element's effective type
//     via `getElementType(context)(attribute.parent)` and skip when it is
//     not in the DOM map. Done BEFORE the `name === 'ROLE'` check upstream
//     (mirrored here for byte-for-byte parity — the observable behavior is
//     identical either way).
//  2. Compare the upper-cased attribute name to 'ROLE'. Skip otherwise.
//  3. Detect `ArrayLiteralExpression` values first — upstream's
//     `LITERAL_TYPES.ArrayExpression` evaluates each element and joins by
//     `,`. `LiteralPropArrayAsString` mirrors that path so we don't lose
//     diagnostics on `<div role={[]} />` / `<div role={['foobar']} />`.
//  4. Otherwise fetch the literal-typed value via `getLiteralPropValue`.
//     Skip when the value is `undefined` (explicit `{undefined}`) or the
//     LITERAL_TYPES noop path (`null` — Identifier non-undefined / Call /
//     Member / Conditional / Logical / Binary / TS-wrapper kinds). Note:
//     explicit `{null}` does NOT skip — upstream's `LITERAL_TYPES.Literal`
//     override maps the JS `null` to the literal string "null", which then
//     fails validation.
//  5. Stringify the value (JS `String(value)`) and space-split it. The
//     attribute is valid iff every space-delimited token is in the
//     non-abstract role set OR the user's `allowedInvalidRoles` list.
//     Otherwise emit a single diagnostic on the JsxAttribute node.
package aria_role

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// errorMessage mirrors upstream's exact string. Reproduced verbatim so a
// future audit can diff against `src/rules/aria-role.js` byte-for-byte.
const errorMessage = "Elements with ARIA roles must use a valid, non-abstract ARIA role."

type options struct {
	ignoreNonDOM        bool
	allowedInvalidRoles map[string]struct{}
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["ignoreNonDOM"].(bool); ok {
		opts.ignoreNonDOM = v
	}
	if arr, ok := m["allowedInvalidRoles"].([]interface{}); ok && len(arr) > 0 {
		opts.allowedInvalidRoles = make(map[string]struct{}, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				opts.allowedInvalidRoles[s] = struct{}{}
			}
		}
	}
	return opts
}

// reportIfInvalid mirrors upstream's `values.every(...)` short-circuit: a
// single diagnostic is emitted the first time a space-delimited token in
// `str` is neither in the allow-list nor in the canonical non-abstract role
// set. JS `String.prototype.split(' ')` is a single-space split (not
// `\s+`); embedded tabs / newlines / leading / trailing / double spaces all
// produce tokens that fail the canonical lookup.
func reportIfInvalid(ctx rule.RuleContext, attr *ast.Node, str string, allowed map[string]struct{}) {
	for _, tok := range strings.Split(str, " ") {
		if _, ok := allowed[tok]; ok {
			continue
		}
		if jsxa11yutil.IsValidAriaRole(tok) {
			continue
		}
		ctx.ReportNode(attr, rule.RuleMessage{
			Id:          "invalidAriaRole",
			Description: errorMessage,
		})
		return
	}
}

var AriaRoleRule = rule.Rule{
	Name: "jsx-a11y/aria-role",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Step 1: ignoreNonDOM. Upstream gates this BEFORE the
				// name === 'ROLE' check; mirror order even though the
				// observable result is identical, to keep the diff against
				// `src/rules/aria-role.js` mechanical.
				if opts.ignoreNonDOM {
					parent := reactutil.GetJsxParentElement(attr)
					if parent == nil {
						return
					}
					elType := jsxa11yutil.GetElementType(parent, ctx.Settings)
					if !jsxa11yutil.IsDOMElement(elType) {
						return
					}
				}

				// Step 2: `propName(attribute).toUpperCase() !== 'ROLE'`.
				// Upstream uses a case-INSENSITIVE compare here even though
				// JSX is case-sensitive at runtime — `<div ROLE="datepicker">`
				// still triggers the rule.
				if !strings.EqualFold(reactutil.GetJsxPropName(attr), "role") {
					return
				}

				// Step 3: ArrayLiteralExpression — upstream's LITERAL_TYPES
				// keeps arrays evaluable (each element via TYPES / full
				// evaluator, null elements filtered, `,`-joined). The shared
				// `LiteralPropAriaValue` collapses arrays to NoLit for
				// callers that don't need element-level coercion, so handle
				// the array shape here to keep observable behavior aligned
				// with `<div role={[]} />` / `<div role={['x']} />`.
				if arrStr, isArr := jsxa11yutil.LiteralPropArrayAsString(attr); isArr {
					reportIfInvalid(ctx, attr, arrStr, opts.allowedInvalidRoles)
					return
				}

				// Step 4: `getLiteralPropValue(attr) === null || === undefined`.
				// AriaLiteralNoLit covers the LITERAL_TYPES noop path (Identifier
				// non-undefined / Call / Member / Conditional / Logical / Binary
				// / TS-wrapper kinds — all noop → null upstream).
				// AriaLiteralUndef covers explicit `{undefined}`.
				// Critically, `{null}` does NOT land here — LITERAL_TYPES.Literal
				// overrides null to the string "null", which falls into
				// AriaLiteralString and is validated like any other string.
				value := jsxa11yutil.LiteralPropAriaValue(attr)
				if value.Kind == jsxa11yutil.AriaLiteralNoLit || value.Kind == jsxa11yutil.AriaLiteralUndef {
					return
				}

				// Step 5: `String(value).split(' ').every(...)`. JS String() on
				// the boolean form yields "true" / "false"; on a number yields
				// the standard Number-to-String form; on a string passes
				// through verbatim. The split is by a single ASCII space (not
				// `\s+`), so any tab / newline embedded in the role value
				// becomes part of a token and fails the lookup.
				str := jsxa11yutil.AriaLiteralValueAsJSString(value)
				reportIfInvalid(ctx, attr, str, opts.allowedInvalidRoles)
			},
		}
	},
}
