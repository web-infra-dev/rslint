// Package no_autofocus ports eslint-plugin-jsx-a11y's `no-autofocus` rule.
// The rule discourages the `autoFocus` prop on JSX elements: programmatically
// shifting focus on render disrupts both sighted users (unexpected viewport
// jumps) and assistive-technology users (screen-reader / keyboard navigation
// gets reset out from under them).
//
// Upstream signature:
//
//	options: { ignoreNonDOM?: boolean (default false) }
//
// Intentional difference from eslint-plugin-jsx-a11y v6.10.2: upstream reports
// every autoFocus attribute without inspecting its value. Rslint permits values
// statically resolved to false or "false" so explicitly disabled props can stay.
// See Differences from upstream in no_autofocus.md.
//
// Trigger: a JsxAttribute whose name is exactly `autoFocus` (case-sensitive
// per upstream `propName`) and whose extracted value is anything other than
// the JS boolean `false` or the literal string `"false"`. The boolean-attribute
// form (`<div autoFocus />`) extracts to JS `true` and trips the rule.
//
// `ignoreNonDOM`: when true, the rule resolves the parent element's name via
// `getElementType` (honoring `polymorphicPropName` and the `components` map
// from `settings['jsx-a11y']`) and only fires when the resolved name is in
// aria-query's `dom` map — i.e. an HTML element. Custom components are
// silently skipped.
//
// All AST plumbing (case-sensitive prop name lookup, value coercion,
// boolean-form handling, element-type resolution) lives in `jsxa11yutil`;
// this rule just wires the gates.
package no_autofocus

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_autofocus.schema.json
var schemaJSON []byte

// errorMessage says enabled rather than upstream's used to reflect the
// intentional allowance for statically disabled values.
const errorMessage = "The autoFocus prop should not be enabled, as it can reduce usability and accessibility for users."

// options holds the rule's parsed configuration. Mirrors the upstream
// `generateObjSchema({ ignoreNonDOM: { type: 'boolean', default: false } })`
// shape.
type options struct {
	IgnoreNonDOM bool
}

func parseOptions(raw []any) options {
	opts := options{}
	if len(raw) == 0 {
		return opts
	}
	optsMap, _ := raw[0].(map[string]interface{})
	if v, ok := optsMap["ignoreNonDOM"].(bool); ok {
		opts.IgnoreNonDOM = v
	}
	return opts
}

var NoAutofocusRule = rule.Rule{
	Name:   "jsx-a11y/no-autofocus",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(attr *ast.Node) {
				// Upstream's `propName(attribute) === 'autoFocus'` is a strict
				// case-sensitive equality. Lowercase variants like `autofocus`
				// are NOT matched (they're the HTML-DOM attribute, not React's
				// camelCased prop) — locked by the upstream valid case
				// `<div autofocus />`.
				if reactutil.GetJsxPropName(attr) != "autoFocus" {
					return
				}
				if opts.IgnoreNonDOM {
					// Resolve via the parent JsxOpeningElement /
					// JsxSelfClosingElement so polymorphicPropName and the
					// `components` map are honored. Skip when the parent isn't
					// an element-like node (defensive — JsxAttribute always has
					// a JsxAttributes parent, which always has an opening /
					// self-closing element parent in legal source).
					parent := reactutil.GetJsxParentElement(attr)
					if parent == nil {
						return
					}
					elementType := jsxa11yutil.GetElementType(parent, ctx.Settings)
					if !jsxa11yutil.IsDOMElement(elementType) {
						return
					}
				}
				// Intentional extension to upstream's presence-only check: allow
				// values resolved to boolean false or the string "false" by our
				// jsx-ast-utils-compatible evaluator. Other falsy values such as
				// null, undefined, 0 and "" still report; this is not a truthiness
				// check. Upstream v6.10.2 does not call getPropValue here.
				if boolVal, ok := jsxa11yutil.PropStaticBoolValue(attr); ok && !boolVal {
					return
				}
				if strVal, ok := jsxa11yutil.PropStaticStringValue(attr); ok && strVal == "false" {
					return
				}
				ctx.ReportNode(attr, rule.RuleMessage{
					Id:          "noAutoFocus",
					Description: errorMessage,
				})
			},
		}
	},
}
