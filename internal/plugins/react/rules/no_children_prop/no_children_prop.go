package no_children_prop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgNestChildren       = "Do not pass children as props. Instead, nest children between the opening and closing tags."
	msgPassChildrenAsArgs = "Do not pass children as props. Instead, pass them as additional arguments to React.createElement."
	msgNestFunction       = "Do not nest a function between the opening and closing tags. Instead, pass it as a prop."
	msgPassFunctionAsArgs = "Do not pass a function as an additional argument to React.createElement. Instead, pass it as a prop."
)

var NoChildrenPropRule = rule.Rule{
	Name: "react/no-children-prop",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		allowFunctions := false
		if optsMap := utils.GetOptionsMap(options); optsMap != nil {
			if v, ok := optsMap["allowFunctions"].(bool); ok {
				allowFunctions = v
			}
		}

		// Mirrors upstream `isFunction`: the value must be an arrow / non-arrow
		// function literal AND the allowFunctions option must be on. We unwrap
		// parentheses because tsgo preserves ParenthesizedExpression where
		// ESTree flattens it (`(() => {})` vs `() => {}`).
		isAllowedFunction := func(node *ast.Node) bool {
			if !allowFunctions || node == nil {
				return false
			}
			return ast.IsFunctionExpressionOrArrowFunction(ast.SkipParentheses(node))
		}

		pragma := reactutil.GetReactPragma(ctx.Settings)

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				nameNode := attr.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				if nameNode.AsIdentifier().Text != "children" {
					return
				}
				if initializer := attr.Initializer; initializer != nil && initializer.Kind == ast.KindJsxExpression {
					if expr := initializer.AsJsxExpression().Expression; isAllowedFunction(expr) {
						return
					}
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "nestChildren",
					Description: msgNestChildren,
				})
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCall(call.Expression, pragma) {
					return
				}
				if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
					return
				}
				secondArg := ast.SkipParentheses(call.Arguments.Nodes[1])
				if secondArg.Kind != ast.KindObjectLiteralExpression {
					return
				}
				obj := secondArg.AsObjectLiteralExpression()
				childrenValue, hasChildrenProp := findChildrenValue(obj)
				if hasChildrenProp {
					// Upstream guards on `childrenProp.value` being truthy; in
					// practice every realistic shape (regular, shorthand) carries
					// one, so `value == nil` is only reachable on malformed trees.
					if childrenValue == nil || isAllowedFunction(childrenValue) {
						return
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "passChildrenAsArgs",
						Description: msgPassChildrenAsArgs,
					})
					return
				}

				if len(call.Arguments.Nodes) == 3 && isAllowedFunction(call.Arguments.Nodes[2]) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "passFunctionAsArgs",
						Description: msgPassFunctionAsArgs,
					})
				}
			},
			ast.KindJsxElement: func(node *ast.Node) {
				jsx := node.AsJsxElement()
				if jsx.Children == nil || len(jsx.Children.Nodes) != 1 {
					return
				}
				child := jsx.Children.Nodes[0]
				if child.Kind != ast.KindJsxExpression {
					return
				}
				if !isAllowedFunction(child.AsJsxExpression().Expression) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "nestFunction",
					Description: msgNestFunction,
				})
			},
		}
	},
}

// findChildrenValue searches an object literal for a `children` property whose
// key is an Identifier, mirroring eslint-plugin-react's
// `'name' in prop.key && prop.key.name === 'children'` guard. String / numeric
// / computed keys and spread elements are deliberately skipped — using
// `utils.GetStaticPropertyName` here would over-match (e.g. it would treat
// `{"children": "x"}` as a hit, which upstream does not). The returned value
// has parentheses stripped so a later `isAllowedFunction` test sees the raw
// function literal.
func findChildrenValue(obj *ast.ObjectLiteralExpression) (value *ast.Node, found bool) {
	if obj.Properties == nil {
		return nil, false
	}
	for _, prop := range obj.Properties.Nodes {
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			pa := prop.AsPropertyAssignment()
			name := pa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "children" {
				if pa.Initializer != nil {
					return ast.SkipParentheses(pa.Initializer), true
				}
				return nil, true
			}
		case ast.KindShorthandPropertyAssignment:
			spa := prop.AsShorthandPropertyAssignment()
			name := spa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "children" {
				// Shorthand `{ children }` — the key and value are the same
				// Identifier node; the shorthand is a variable reference, not
				// a function literal, so `allowFunctions` never exempts it.
				return name, true
			}
		}
	}
	return nil, false
}
