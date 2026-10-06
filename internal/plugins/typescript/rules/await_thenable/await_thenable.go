package await_thenable

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "await",
		Description: "Unexpected `await` of a non-Promise (non-\"Thenable\") value.",
	}
}

func buildRemoveAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAwait",
		Description: "Remove unnecessary `await`.",
	}
}

func buildForAwaitOfNonAsyncIterableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "forAwaitOfNonAsyncIterable",
		Description: "Unexpected `for await...of` of a value that is not async iterable.",
	}
}

func buildConvertToOrdinaryForMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "convertToOrdinaryFor",
		Description: "Convert to an ordinary `for...of` loop.",
	}
}

func buildAwaitUsingOfNonAsyncDisposableMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "awaitUsingOfNonAsyncDisposable",
		Description: "Unexpected `await using` of a value that is not async disposable.",
	}
}

func buildInvalidPromiseAggregatorInputMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidPromiseAggregatorInput",
		Description: "Unexpected iterable of non-Promise (non-\"Thenable\") values passed to promise aggregator.",
	}
}

var AwaitThenableRule = rule.CreateRule(rule.Rule{
	Name:             "await-thenable",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := ast.SkipParentheses(call.Expression)
				if !ast.IsAccessExpression(callee) || len(call.Arguments.Nodes) == 0 {
					return
				}
				method, known := utils.AccessExpressionStaticName(callee)
				if !known {
					if evaluator == nil {
						evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
						evaluator.GlobalAccess = ctx.Globals.Access
					}
					method, _ = evaluator.EvalAccessExpressionName(callee)
				}
				switch method {
				case "all", "allSettled", "any", "race":
				default:
					return
				}
				calleeType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, callee.Expression())
				// Upstream recognizes PromiseConstructor and its derived interfaces,
				// not the broader Promise class constructors accepted by our helper.
				if !utils.IsBuiltinSymbolLike(ctx.Program(), ctx.TypeChecker, calleeType, "PromiseConstructor") {
					return
				}

				argument := ast.SkipParentheses(call.Arguments.Nodes[0])
				if ast.IsArrayLiteralExpression(argument) {
					for _, element := range argument.AsArrayLiteralExpression().Elements.Nodes {
						if element.Kind == ast.KindOmittedExpression {
							continue
						}
						element = ast.SkipParentheses(element)
						t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, element)
						if isAlwaysNonAwaitableType(ctx.TypeChecker, element, t) {
							ctx.ReportNode(element, buildInvalidPromiseAggregatorInputMessage())
						}
					}
					return
				}
				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, argument)
				if isInvalidPromiseAggregatorInput(ctx.TypeChecker, argument, t) {
					ctx.ReportNode(argument, buildInvalidPromiseAggregatorInputMessage())
				}
			},
			ast.KindAwaitExpression: func(node *ast.Node) {
				awaitArgument := node.AsAwaitExpression().Expression
				awaitArgumentType := ctx.TypeChecker.GetTypeAtLocation(awaitArgument)
				certainty := utils.NeedsToBeAwaited(ctx.TypeChecker, awaitArgument, awaitArgumentType)

				if certainty == utils.TypeAwaitableNever {
					ctx.ReportNodeWithDeferredSuggestions(node, buildAwaitMessage(), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message: buildRemoveAwaitMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())),
							},
						}}
					})
				}
			},
			ast.KindForOfStatement: func(node *ast.Node) {
				stmt := node.AsForInOrOfStatement()
				if stmt.AwaitModifier == nil {
					return
				}

				exprType := ctx.TypeChecker.GetTypeAtLocation(stmt.Expression)
				if utils.IsTypeAnyType(exprType) {
					return
				}

				for _, typePart := range utils.UnionTypeParts(exprType) {
					if utils.GetWellKnownSymbolPropertyOfType(typePart, "asyncIterator", ctx.TypeChecker) != nil {
						return
					}
				}

				ctx.ReportRangeWithDeferredSuggestions(
					utils.GetForStatementHeadLoc(ctx.SourceFile, node),
					buildForAwaitOfNonAsyncIterableMessage(),
					// Note that this suggestion causes broken code for sync iterables
					// of promises, since the loop variable is not awaited.
					func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message: buildConvertToOrdinaryForMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemove(ctx.SourceFile, stmt.AwaitModifier),
							},
						}}
					},
				)
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if !ast.IsVarAwaitUsing(node) {
					return
				}

				declaration := node.AsVariableDeclarationList()
			DeclaratorLoop:
				for _, declarator := range declaration.Declarations.Nodes {
					init := declarator.Initializer()
					if init == nil {
						continue
					}
					initType := ctx.TypeChecker.GetTypeAtLocation(init)
					if utils.IsTypeAnyType(initType) {
						continue
					}

					for _, typePart := range utils.UnionTypeParts(initType) {
						if utils.GetWellKnownSymbolPropertyOfType(typePart, "asyncDispose", ctx.TypeChecker) != nil {
							continue DeclaratorLoop
						}
					}

					// let the user figure out what to do if there's
					// await using a = b, c = d, e = f;
					// it's rare and not worth the complexity to handle.
					if len(declaration.Declarations.Nodes) != 1 {
						ctx.ReportNode(init, buildAwaitUsingOfNonAsyncDisposableMessage())
						continue
					}
					ctx.ReportNodeWithDeferredSuggestions(init, buildAwaitUsingOfNonAsyncDisposableMessage(), func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{
							Message: buildRemoveAwaitMessage(),
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())),
							},
						}}
					})
				}
			},
		}
	},
})

func isAlwaysNonAwaitableType(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	for _, part := range utils.UnionTypeParts(t) {
		if utils.NeedsToBeAwaited(typeChecker, node, part) != utils.TypeAwaitableNever {
			return false
		}
	}
	return true
}

func isInvalidPromiseAggregatorInput(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	parts := utils.UnionTypeParts(t)
	// Non-iterable inputs already produce a TypeScript error. Match upstream by
	// requiring every union constituent to be iterable before checking values.
	for _, part := range parts {
		if utils.GetWellKnownSymbolPropertyOfType(part, "iterator", typeChecker) == nil {
			return false
		}
	}
	for _, part := range parts {
		for _, valueType := range getValueTypesOfArrayLike(typeChecker, part) {
			// A literal element is reported only when it cannot be awaited, but
			// an iterable variable is reported if it can contain a non-Thenable.
			for _, valuePart := range utils.UnionTypeParts(valueType) {
				if utils.NeedsToBeAwaited(typeChecker, node, valuePart) == utils.TypeAwaitableNever {
					return true
				}
			}
		}
	}
	return false
}

func getValueTypesOfArrayLike(typeChecker *checker.Checker, t *checker.Type) []*checker.Type {
	if checker.IsTupleType(t) {
		return checker.Checker_getTypeArguments(typeChecker, t)
	}
	if typeChecker.IsArrayLikeType(t) {
		return []*checker.Type{typeChecker.GetNumberIndexType(t)}
	}
	// For other iterable references, upstream checks only the first type
	// argument (the yielded value, not a generator's return or next type).
	if utils.IsTypeReference(t) {
		arguments := checker.Checker_getTypeArguments(typeChecker, t)
		return arguments[:min(1, len(arguments))]
	}
	return nil
}
