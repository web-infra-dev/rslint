package callback_return

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed callback_return.schema.json
var schemaJSON []byte

var missingReturn = rule.RuleMessage{
	Id:          "missingReturn",
	Description: "Expected return with your callback function.",
}

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/callback-return.js
var CallbackReturnRule = rule.Rule{
	Name:   "node/callback-return",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		callbacks := []any{"callback", "cb", "next"}
		if len(options) > 0 {
			if names, ok := options[0].([]any); ok {
				callbacks = names
			}
		}
		if len(callbacks) == 0 {
			return nil
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
				if !hasIdentifierReceiver(callee) || !slices.Contains(callbacks, any(scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, callee, false))) {
					return
				}

				closest := ast.FindAncestor(node.Parent, func(parent *ast.Node) bool {
					switch parent.Kind {
					case ast.KindReturnStatement, ast.KindArrowFunction:
						return true
					case ast.KindBlock:
						return parent.Parent.Kind != ast.KindClassStaticBlockDeclaration
					}
					return false
				})
				if closest != nil {
					if closest.Kind != ast.KindBlock {
						return
					}
					statements := closest.AsBlock().Statements.Nodes
					if len(statements) > 0 {
						last := statements[len(statements)-1]
						if isCallbackExpression(node, last) && utils.IsFunctionLikeContainer(closest.Parent) {
							return
						}
						if last.Kind == ast.KindReturnStatement && len(statements) > 1 &&
							isCallbackExpression(node, statements[len(statements)-2]) {
							return
						}
					}
				}

				for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
					// ESTree places method keys and decorators outside the function value.
					if utils.IsFunctionLikeContainer(parent) && parent.Body() != nil &&
						parent.Name() != child && child.Kind != ast.KindDecorator {
						ctx.ReportNode(node, missingReturn)
						return
					}
				}
			},
		}
	},
}

// Upstream restricts the receiver chain, but accepts any property syntax.
// Callback names are matched against source text, including spaces and comments.
func hasIdentifierReceiver(node *ast.Node) bool {
	for node != nil {
		if node.Kind == ast.KindIdentifier {
			return true
		}
		object, _ := utils.MemberExpressionParts(node)
		node = utils.ESTreeCallCallee(object)
	}
	return false
}

func isCallbackExpression(call, statement *ast.Node) bool {
	// Optional calls have a ChainExpression wrapper in ESTree.
	if statement.Kind != ast.KindExpressionStatement || ast.IsOptionalChain(call) {
		return false
	}
	expression := utils.ESTreeRuntimeExpression(statement.AsExpressionStatement().Expression)
	if expression == call {
		return true
	}
	if expression.Kind == ast.KindBinaryExpression {
		binary := expression.AsBinaryExpression()
		// Assignments and sequences are separate expression kinds in ESTree.
		return !ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.OperatorToken.Kind != ast.KindCommaToken &&
			utils.ESTreeRuntimeExpression(binary.Right) == call
	}
	return false
}
