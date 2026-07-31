package error_message

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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

			if node.Kind == ast.KindCallExpression {
				call := node.AsCallExpression()
				calleeNode = call.Expression
				if call.Arguments != nil {
					args = call.Arguments.Nodes
				}
			} else if node.Kind == ast.KindNewExpression {
				newExpr := node.AsNewExpression()
				calleeNode = newExpr.Expression
				if newExpr.Arguments != nil {
					args = newExpr.Arguments.Nodes
				}
			} else {
				return
			}

			if calleeNode == nil {
				return
			}

			callee := utils.SkipAssertionsAndParens(calleeNode)
			if callee.Kind != ast.KindIdentifier {
				return
			}

			constructorName := callee.AsIdentifier().Text
			if !builtinErrors[constructorName] {
				return
			}

			if utils.IsShadowed(callee, constructorName) {
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
			if argNode == nil {
				ctx.ReportNode(node, missingMessage(constructorName))
				return
			}

			unwrapped := utils.SkipAssertionsAndParens(argNode)

			// Fallback AST checks for literals that are definitely not strings
			switch unwrapped.Kind {
			case ast.KindNumericLiteral,
				ast.KindBigIntLiteral,
				ast.KindTrueKeyword,
				ast.KindFalseKeyword,
				ast.KindNullKeyword,
				ast.KindArrayLiteralExpression,
				ast.KindObjectLiteralExpression,
				ast.KindRegularExpressionLiteral,
				ast.KindFunctionExpression,
				ast.KindArrowFunction,
				ast.KindClassExpression:
				ctx.ReportNode(argNode, notStringMessage)
				return
			}

			// If type checking is available, check if the type is definitely not string-like
			if ctx.TypeChecker != nil {
				t := ctx.TypeChecker.GetTypeAtLocation(argNode)
				if t != nil {
					flags := t.Flags()
					if flags&(checker.TypeFlagsStringLike|checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsUnion|checker.TypeFlagsIntersection) == 0 {
						if flags&(checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid|checker.TypeFlagsObject) != 0 {
							ctx.ReportNode(argNode, notStringMessage)
							return
						}
					}
				}
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
