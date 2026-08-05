package error_message

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDMissing   = "missing-message"
	messageIDEmpty     = "message-is-empty-string"
	messageIDNotString = "message-is-not-a-string"
)

func missingMessage(constructorName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDMissing,
		Description: "Pass a message to the `" + constructorName + "` constructor.",
	}
}

var emptyMessage = rule.RuleMessage{
	Id:          messageIDEmpty,
	Description: "Error message should not be an empty string.",
}

var notStringMessage = rule.RuleMessage{
	Id:          messageIDNotString,
	Description: "Error message should be a string.",
}

var builtinErrors = map[string]bool{
	"Error":           true,
	"EvalError":       true,
	"RangeError":      true,
	"ReferenceError":  true,
	"SyntaxError":     true,
	"TypeError":       true,
	"URIError":        true,
	"AggregateError":  true,
	"SuppressedError": true,
}

// isLocalReference reports whether the error name resolves to a binding
// declared in this file rather than to the global constructor.
func isLocalReference(ctx rule.RuleContext, node *ast.Node, name string) bool {
	if ctx.Refs != nil {
		// Resolve can return a symbol declared elsewhere (globals, cross-file,
		// .d.ts), so a non-nil result alone doesn't mean local.
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			return utils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
		}
	}
	return utils.IsShadowed(node, name)
}

func getMessageArgumentIndex(constructorName string) int {
	switch constructorName {
	case "AggregateError":
		return 1
	case "SuppressedError":
		return 2
	default:
		return 0
	}
}

var ErrorMessageRule = rule.Rule{
	Name:   "unicorn/error-message",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		staticStrings := utils.NewStaticStringEvaluatorWithSourceFile(ctx.TypeChecker, ctx.SourceFile)

		checkExpression := func(node *ast.Node) {
			if ast.IsOptionalChain(node) {
				return
			}

			var calleeNode *ast.Node
			var args []*ast.Node

			switch node.Kind {
			case ast.KindCallExpression:
				call := node.AsCallExpression()
				calleeNode = call.Expression
				if call.Arguments != nil {
					args = call.Arguments.Nodes
				}
			case ast.KindNewExpression:
				newExpr := node.AsNewExpression()
				calleeNode = newExpr.Expression
				if newExpr.Arguments != nil {
					args = newExpr.Arguments.Nodes
				}
			default:
				return
			}

			if calleeNode == nil {
				return
			}

			// Only parentheses are skipped: a type assertion or non-null
			// assertion makes the callee a wrapper node, which upstream's
			// identifier check rejects.
			callee := ast.SkipOuterExpressions(calleeNode, ast.OEKParentheses)
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}

			constructorName := callee.AsIdentifier().Text
			if !builtinErrors[constructorName] {
				return
			}

			if isLocalReference(ctx, callee, constructorName) {
				return
			}

			messageArgumentIndex := getMessageArgumentIndex(constructorName)

			// If message is SpreadElement or there is SpreadElement before message
			for i, arg := range args {
				if i <= messageArgumentIndex && arg.Kind == ast.KindSpreadElement {
					return
				}
			}

			if len(args) <= messageArgumentIndex {
				ctx.ReportNode(node, missingMessage(constructorName))
				return
			}

			argNode := args[messageArgumentIndex]

			// Literals that can never be a string, for the shapes the static
			// evaluator leaves unresolved.
			switch utils.SkipAssertionsAndParens(argNode).Kind {
			case ast.KindNumericLiteral,
				ast.KindBigIntLiteral,
				ast.KindArrayLiteralExpression,
				ast.KindObjectLiteralExpression,
				ast.KindRegularExpressionLiteral:
				ctx.ReportNode(argNode, notStringMessage)
				return
			}

			val, ok := staticStrings.EvalValue(argNode)
			if !ok {
				return
			}

			strVal, isString := val.(string)
			if !isString {
				ctx.ReportNode(argNode, notStringMessage)
				return
			}

			if strVal == "" {
				ctx.ReportNode(argNode, emptyMessage)
				return
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: checkExpression,
			ast.KindNewExpression:  checkExpression,
		}
	},
}
