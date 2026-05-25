// Package no_access_key ports eslint-plugin-jsx-a11y's `no-access-key` rule.
// The rule discourages the `accessKey` (or case-insensitive variants like
// `accesskey`, `acCesSKeY`) prop on JSX elements: keyboard shortcuts assigned
// via `accessKey` collide with the keyboard commands used by screen readers
// and keyboard-only users, causing accessibility regressions.
//
// Upstream is intentionally minimal — there are no options, the rule reports
// once per JsxOpeningElement whose `accesskey` prop has a truthy
// jsx-ast-utils `getPropValue`. The case-insensitive name lookup, spread
// expansion, and TS-wrapper unwrapping all live in `jsxa11yutil`; this rule
// just wires the listener and the truthy gate.
package no_access_key

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// errorMessage mirrors upstream's `errorMessage` string verbatim.
const errorMessage = "No access key attribute allowed. Inconsistencies between keyboard shortcuts and keyboard commands used by screen readers and keyboard-only users create a11y complications."

var NoAccessKeyRule = rule.Rule{
	Name: "jsx-a11y/no-access-key",
	Run: func(ctx rule.RuleContext, _ any) rule.RuleListeners {
		check := func(node *ast.Node) {
			attrs := reactutil.GetJsxElementAttributes(node)
			accessKeyAttr := jsxa11yutil.FindAttributeByName(attrs, "accesskey")
			if accessKeyAttr == nil {
				return
			}
			// Mirrors upstream's `if (accessKey && accessKeyValue)` —
			// `accessKey` is the prop node (truthy when present),
			// `accessKeyValue` is `getPropValue(accessKey)` and we gate on
			// its JS-truthiness. PropValueIsTruthy handles the full TYPES
			// path: literal coercion of "true"/"false", identifier-name
			// stringification (non-undefined identifiers are truthy),
			// template-literal substitution (always truthy), conditional
			// / logical / binary short-circuits, the boolean-attribute
			// form (extractValue null-attr-value → true), and so on.
			if !jsxa11yutil.PropValueIsTruthy(accessKeyAttr) {
				return
			}
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noAccessKey",
				Description: errorMessage,
			})
		}
		// tsgo splits ESTree's `JSXOpeningElement` into KindJsxOpeningElement
		// (paired tags `<div>...</div>`) and KindJsxSelfClosingElement
		// (`<div />`). Upstream's single `JSXOpeningElement` listener covers
		// both; we mirror by registering on both kinds.
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
