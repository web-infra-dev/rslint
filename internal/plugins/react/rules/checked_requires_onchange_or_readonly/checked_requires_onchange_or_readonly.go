package checked_requires_onchange_or_readonly

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgMissingProperty           = "`checked` should be used with either `onChange` or `readOnly`."
	msgExclusiveCheckedAttribute = "Use either `checked` or `defaultChecked`, but not both."
)

// targetProps mirrors upstream's `targetPropSet`. Only these four names are
// tracked; every other attribute / property is ignored. Note the rule does
// NOT inspect the `type` attribute — any `<input>` (or `createElement('input')`)
// carrying `checked` is checked, regardless of `type="checkbox"` vs `"radio"`
// vs anything else.
var targetProps = map[string]struct{}{
	"checked":        {},
	"onChange":       {},
	"readOnly":       {},
	"defaultChecked": {},
}

type options struct {
	ignoreMissingProperties         bool
	ignoreExclusiveCheckedAttribute bool
}

func parseOptions(opts any) options {
	o := options{}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap == nil {
		return o
	}
	if v, ok := optsMap["ignoreMissingProperties"].(bool); ok {
		o.ignoreMissingProperties = v
	}
	if v, ok := optsMap["ignoreExclusiveCheckedAttribute"].(bool); ok {
		o.ignoreExclusiveCheckedAttribute = v
	}
	return o
}

var CheckedRequiresOnchangeOrReadonlyRule = rule.Rule{
	Name: "react/checked-requires-onchange-or-readonly",
	Run: func(ctx rule.RuleContext, _opts []any) rule.RuleListeners {
		opts := rule.UnwrapOptions(_opts)
		o := parseOptions(opts)
		pragma := reactutil.GetReactPragma(ctx.Settings)

		// checkAndReport mirrors upstream's `checkAttributesAndReport`: report
		// only when `checked` is present, then emit the exclusive-attribute and
		// missing-property diagnostics in that order (so a node carrying both
		// `checked` and `defaultChecked` reports exclusiveCheckedAttribute first).
		checkAndReport := func(node *ast.Node, propSet map[string]struct{}) {
			if _, hasChecked := propSet["checked"]; !hasChecked {
				return
			}
			if _, hasDefaultChecked := propSet["defaultChecked"]; !o.ignoreExclusiveCheckedAttribute && hasDefaultChecked {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "exclusiveCheckedAttribute",
					Description: msgExclusiveCheckedAttribute,
				})
			}
			_, hasOnChange := propSet["onChange"]
			_, hasReadOnly := propSet["readOnly"]
			if !o.ignoreMissingProperties && !hasOnChange && !hasReadOnly {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingProperty",
					Description: msgMissingProperty,
				})
			}
		}

		// checkJsxElement handles the upstream `JSXOpeningElement` listener. In
		// tsgo a self-closing `<input/>` is a JsxSelfClosingElement and a paired
		// `<input></input>` produces a JsxOpeningElement, so both kinds route
		// here. The element node itself is the report target, matching upstream's
		// `report(..., { node })` on the JSXOpeningElement (whose loc is the
		// opening tag, not the whole element).
		checkJsxElement := func(node *ast.Node) {
			if reactutil.GetJsxElementTypeString(node) != "input" {
				return
			}
			checkAndReport(node, jsxPropSet(reactutil.GetJsxElementAttributes(node)))
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     checkJsxElement,
			ast.KindJsxSelfClosingElement: checkJsxElement,
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCall(call.Expression, pragma) {
					return
				}
				// Upstream requires `arguments[0]` to be a string Literal "input"
				// and `arguments[1]` to be an ObjectExpression; anything else
				// (missing args, non-literal tag, non-object props) bails out.
				if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
					return
				}
				firstArg := ast.SkipParentheses(call.Arguments.Nodes[0])
				if firstArg.Kind != ast.KindStringLiteral || firstArg.AsStringLiteral().Text != "input" {
					return
				}
				secondArg := ast.SkipParentheses(call.Arguments.Nodes[1])
				if secondArg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				checkAndReport(node, objectPropSet(secondArg.AsObjectLiteralExpression()))
			},
		}
	},
}

// jsxPropSet collects the subset of `targetProps` present as JSX attributes,
// mirroring upstream `extractTargetProps(node.attributes, 'name')`.
// reactutil.GetJsxPropName returns the plain attribute name for an identifier
// attribute, "ns:name" for a namespaced one, and "spread" for a spread
// attribute — none of which (besides a plain identifier) can land in
// targetProps, exactly like upstream's `attr.name.name` access where a
// spread has no `.name` and a namespaced `.name` is a node, not a string.
func jsxPropSet(attrs []*ast.Node) map[string]struct{} {
	set := make(map[string]struct{})
	for _, attr := range attrs {
		name := reactutil.GetJsxPropName(attr)
		if _, ok := targetProps[name]; ok {
			set[name] = struct{}{}
		}
	}
	return set
}

// objectPropSet collects the subset of `targetProps` present as object keys,
// mirroring upstream `extractTargetProps(secondArg.properties, 'key')`.
func objectPropSet(obj *ast.ObjectLiteralExpression) map[string]struct{} {
	set := make(map[string]struct{})
	if obj.Properties == nil {
		return set
	}
	for _, prop := range obj.Properties.Nodes {
		name, ok := objectKeyName(prop)
		if !ok {
			continue
		}
		if _, isTarget := targetProps[name]; isTarget {
			set[name] = struct{}{}
		}
	}
	return set
}

// objectKeyName reproduces upstream's `prop.key.name` access on an object
// property. Upstream reads `.name` without gating on `computed`, so:
//   - `{ checked: … }`      → "checked"  (Identifier key)
//   - `{ checked }`         → "checked"  (shorthand)
//   - `{ [checked]: … }`    → "checked"  (computed Identifier — `.name` exists)
//   - `{ "checked": … }`    → not matched (string Literal key has `.value`, not `.name`)
//   - `{ ["checked"]: … }`  → not matched (computed string Literal key)
//   - `{ ...rest }`         → not matched (SpreadAssignment has no `.key`)
func objectKeyName(prop *ast.Node) (string, bool) {
	var nameNode *ast.Node
	switch prop.Kind {
	case ast.KindPropertyAssignment:
		nameNode = prop.AsPropertyAssignment().Name()
	case ast.KindShorthandPropertyAssignment:
		nameNode = prop.AsShorthandPropertyAssignment().Name()
	default:
		return "", false
	}
	if nameNode == nil {
		return "", false
	}
	switch nameNode.Kind {
	case ast.KindIdentifier:
		return nameNode.AsIdentifier().Text, true
	case ast.KindComputedPropertyName:
		inner := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
		if inner != nil && inner.Kind == ast.KindIdentifier {
			return inner.AsIdentifier().Text, true
		}
	}
	return "", false
}
