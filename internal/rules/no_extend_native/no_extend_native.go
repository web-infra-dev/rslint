package no_extend_native

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_extend_native.schema.json
var schemaJSON []byte

type options struct {
	exceptions map[string]bool
}

func parseOptions(rawOptions []any) options {
	opts := options{exceptions: make(map[string]bool)}
	if len(rawOptions) == 0 {
		return opts
	}
	m, _ := rawOptions[0].(map[string]any)
	exceptions, _ := m["exceptions"].([]any)
	for _, e := range exceptions {
		if s, ok := e.(string); ok {
			opts.exceptions[s] = true
		}
	}
	return opts
}

// isAssignmentOperator reports whether the given binary operator is a
// (compound) assignment, including logical assignments.
func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// memberObject returns the object expression of a property/element access node,
// or nil if the node is not a member access.
func memberObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// staticMemberName returns the static property name of a member access, or
// ("", false) if it cannot be determined statically.
func staticMemberName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text, true
		}
	case ast.KindElementAccessExpression:
		return utils.GetStaticExpressionValue(ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression))
	}
	return "", false
}

// skipParensUp walks up from `node` through ParenthesizedExpression parents.
func skipParensUp(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// https://eslint.org/docs/latest/rules/no-extend-native
var NoExtendNativeRule = rule.Rule{
	Name:   "no-extend-native",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		o := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				// ESLint builds this rule's candidate set from the latest
				// ECMAScript globals whose first character is uppercase. Membership
				// is edition-independent, while Access below decides whether the
				// candidate exists in this file's effective global scope.
				if name == "" || name[0] < 'A' || name[0] > 'Z' ||
					!ctx.Globals.IsECMAScriptGlobalName(name) || o.exceptions[name] {
					return
				}

				// In tsgo, parentheses are explicit nodes — `(Object).prototype.p`
				// wraps the identifier in a ParenthesizedExpression. Walk up
				// through any wrapping parens before checking the member access.
				identExpr := skipParensUp(node)
				parent := identExpr.Parent
				if parent == nil {
					return
				}

				// `identExpr` must be the object of a member access whose property is `prototype`.
				obj := memberObject(parent)
				if obj != identExpr {
					return
				}
				propName, ok := staticMemberName(parent)
				if !ok || propName != "prototype" {
					return
				}

				if utils.IsShadowed(node, name) {
					return
				}

				// ESLint looks the candidate up in the actual global scope. The
				// selected edition and authored overrides therefore both apply.
				if !ctx.Globals.Access(name).IsDeclared() {
					return
				}

				// Walk up through any wrapping parentheses to find the next significant parent.
				prototypeAccess := skipParensUp(parent)
				next := prototypeAccess.Parent
				if next == nil {
					return
				}

				// Case 1: assignment to a property of the prototype.
				// e.g. `Object.prototype.p = 0`, `(Object?.prototype).p = 0`,
				//      `Array.prototype.p &&= 0`.
				if memberObject(next) == prototypeAccess {
					memberAccess := skipParensUp(next)
					assign := memberAccess.Parent
					if assign != nil && assign.Kind == ast.KindBinaryExpression {
						bin := assign.AsBinaryExpression()
						if bin.OperatorToken != nil &&
							isAssignmentOperator(bin.OperatorToken.Kind) &&
							bin.Left == memberAccess {
							ctx.ReportNode(assign, rule.RuleMessage{
								Id:          "unexpected",
								Description: name + " prototype is read only, properties should not be added.",
							})
							return
						}
					}
				}

				// Case 2: first argument of `Object.defineProperty` /
				// `Object.defineProperties`.
				if next.Kind == ast.KindCallExpression {
					call := next.AsCallExpression()
					if call.Arguments == nil || len(call.Arguments.Nodes) == 0 ||
						call.Arguments.Nodes[0] != prototypeAccess {
						return
					}
					if utils.IsSpecificMemberAccess(call.Expression, "Object", "defineProperty") ||
						utils.IsSpecificMemberAccess(call.Expression, "Object", "defineProperties") {
						ctx.ReportNode(next, rule.RuleMessage{
							Id:          "unexpected",
							Description: name + " prototype is read only, properties should not be added.",
						})
					}
				}
			},
		}
	},
}
