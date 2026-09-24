package prefer_spy_on

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Target is one assignment of a mock factory call to an object property.
type Target struct {
	// Assignment is the reported assignment expression.
	Assignment *ast.Node
	// Left is the assigned member access, with its parentheses skipped.
	Left *ast.Node
	// Object is the expression the property is read off.
	Object *ast.Node
	// Property is the property name of a dotted access, or the key
	// expression of a bracketed one.
	Property *ast.Node
	// FnCall is the mock factory call the right-hand side chain starts from.
	FnCall *ast.Node
	// Receiver is the object the mock factory is called on.
	Receiver *ast.Node
}

// Config describes one test framework's prefer-spy-on rule.
type Config struct {
	Name    string
	Message rule.RuleMessage
	// PlainAssignmentOnly limits the rule to `=`. Otherwise compound
	// assignments such as `??=` are reported as well.
	PlainAssignmentOnly bool
	// UnwrapTypeAssertions lets the right-hand side chain pass through
	// TypeScript assertions, which do not change the runtime value.
	UnwrapTypeAssertions bool
	// Prepare returns a matcher that reports whether a call is the
	// framework's mock factory call, and the object it is called on.
	Prepare func(ctx rule.RuleContext) func(call *ast.Node) (*ast.Node, bool)
	// Fix builds the edits for one target. It is only called when fixes are
	// requested, and may return nil to report without a fix.
	Fix func(ctx rule.RuleContext, target Target) []rule.RuleFix
}

// NewRule creates a prefer-spy-on rule for a test framework.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			isFnCall := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					if !ast.IsAssignmentExpression(node, config.PlainAssignmentOnly) {
						return
					}

					bin := node.AsBinaryExpression()
					if bin == nil {
						return
					}

					left := ast.SkipParentheses(bin.Left)
					if !IsMemberAccessNode(left) {
						return
					}

					prop := AccessExpressionPropertyNode(left)
					if prop != nil && prop.Kind == ast.KindPrivateIdentifier {
						return
					}

					fnCall, receiver := findFnCall(bin.Right, isFnCall, config.UnwrapTypeAssertions)
					if fnCall == nil {
						return
					}

					target := Target{
						Assignment: node,
						Left:       left,
						Object:     utils.AccessExpressionObject(left),
						Property:   prop,
						FnCall:     fnCall,
						Receiver:   receiver,
					}
					ctx.ReportNodeWithDeferredFixes(node, config.Message, func() []rule.RuleFix {
						return config.Fix(ctx, target)
					})
				},
			}
		},
	}
}

// findFnCall walks down the member and call chain that node heads, to the
// mock factory call it starts from. Arguments are not entered: a mock passed to
// another call is not what the property is assigned.
func findFnCall(
	node *ast.Node,
	isFnCall func(*ast.Node) (*ast.Node, bool),
	unwrapTypeAssertions bool,
) (*ast.Node, *ast.Node) {
	for node != nil {
		if unwrapTypeAssertions {
			node = utils.SkipAssertionsAndParens(node)
		} else {
			node = ast.SkipParentheses(node)
		}
		if node == nil {
			return nil, nil
		}

		switch {
		case node.Kind == ast.KindCallExpression:
			if receiver, ok := isFnCall(node); ok {
				return node, receiver
			}
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if !IsMemberAccessNode(callee) {
				return nil, nil
			}
			node = utils.AccessExpressionObject(callee)
		case IsMemberAccessNode(node):
			node = utils.AccessExpressionObject(node)
		default:
			return nil, nil
		}
	}
	return nil, nil
}

// IsMemberAccessNode reports whether node is a dotted or bracketed member
// access.
func IsMemberAccessNode(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression)
}

// AccessExpressionPropertyNode returns the property name of a dotted access or
// the key expression of a bracketed one.
func AccessExpressionPropertyNode(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().ArgumentExpression
	default:
		return nil
	}
}
