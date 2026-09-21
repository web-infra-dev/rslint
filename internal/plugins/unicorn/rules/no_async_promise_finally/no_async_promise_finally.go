package no_async_promise_finally

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-async-promise-finally"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not pass an async function to `Promise#finally()`.",
}

// NoAsyncPromiseFinallyRule disallows async Promise#finally callbacks.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/no-async-promise-finally.js
var NoAsyncPromiseFinallyRule = rule.Rule{
	Name:   "unicorn/no-async-promise-finally",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		propertyNames := utils.NewStaticStringEvaluatorWithReferenceResolver(
			ctx.TypeChecker,
			ctx.SourceFile,
			ctx.Refs,
		)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeRuntimeExpression(call.Expression)
				if callee == nil || !ast.IsAccessExpression(callee) {
					return
				}

				propertyName, ok := propertyNames.EvalAccessExpressionName(callee)
				if !ok || propertyName != "finally" {
					return
				}

				receiver := utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(callee))
				if receiver == nil || isDefinitelyNonPromiseReceiver(ctx, receiver) {
					return
				}

				arguments := node.Arguments()
				if len(arguments) == 0 || arguments[0] == nil ||
					arguments[0].Kind == ast.KindSpreadElement {
					return
				}

				callback := arguments[0]
				if !isAsyncFinallyCallback(ctx, callback) {
					return
				}

				ctx.ReportNode(callback, message)
			},
		}
	},
}

func isAsyncFinallyCallback(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)

	if isAsyncNonGeneratorFunction(node) {
		return true
	}

	if initializer := utils.GetConstVariableInitializer(node, ctx.TypeChecker); initializer != nil &&
		isAsyncNonGeneratorFunction(initializer) {
		return true
	}

	return isAsyncFunctionDeclarationReference(ctx, node)
}

func isAsyncFunctionDeclarationReference(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}

	symbol := ctx.Refs.Resolve(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}

	declaration := symbol.Declarations[0]
	return declaration != nil &&
		declaration.Kind == ast.KindFunctionDeclaration &&
		isAsyncNonGeneratorFunction(declaration)
}

func isAsyncNonGeneratorFunction(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)

	switch node.Kind {
	case ast.KindArrowFunction:
		return ast.IsAsyncFunction(node)
	case ast.KindFunctionExpression:
		return ast.IsAsyncFunction(node) &&
			node.AsFunctionExpression().AsteriskToken == nil
	case ast.KindFunctionDeclaration:
		return ast.IsAsyncFunction(node) &&
			node.AsFunctionDeclaration().AsteriskToken == nil
	default:
		return false
	}
}

func isDefinitelyNonPromiseReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}

	typ := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(node))

	sawPromise := false
	sawNonPromise := false
	for _, part := range utils.UnionTypeParts(typ) {
		if utils.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) ||
			utils.IsIntrinsicErrorType(part) {
			return false
		}

		if ctx.TypeChecker.GetPromisedTypeOfPromise(part) != nil {
			sawPromise = true
		} else {
			sawNonPromise = true
		}
		if sawPromise && sawNonPromise {
			return false
		}
	}

	return sawNonPromise
}
