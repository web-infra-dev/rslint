package require_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildMissingAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAwait",
		Description: "Function has no 'await' expression.",
	}
}

//nolint:unused
func buildRemoveAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAsync",
		Description: "Remove 'async'.",
	}
}

type scopeInfo struct {
	hasAwait      bool
	isAsyncYield  bool
	functionFlags ast.FunctionFlags
}

func hasCallSignature(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if len(utils.GetCallSignatures(typeChecker, typePart)) != 0 {
				return true
			}
		}
		return false
	}
	return len(utils.GetCallSignatures(typeChecker, t)) != 0
}

func isCallbackParameter(typeChecker *checker.Checker, param *ast.Symbol, node *ast.Node) bool {
	t := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(param, node))
	if param.ValueDeclaration != nil &&
		ast.IsParameterDeclaration(param.ValueDeclaration) &&
		param.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
		t = checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
		if t == nil {
			return false
		}
	}
	return hasCallSignature(typeChecker, t)
}

func hasThenCallbackSignature(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if hasThenCallbackSignature(typeChecker, node, typePart) {
				return true
			}
		}
		return false
	}
	for _, signature := range utils.GetCallSignatures(typeChecker, t) {
		parameters := checker.Signature_parameters(signature)
		if len(parameters) != 0 && isCallbackParameter(typeChecker, parameters[0], node) {
			return true
		}
	}
	return false
}

func isThenableTypePart(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	then := checker.Checker_getPropertyOfType(typeChecker, t, "then")
	if then == nil {
		return false
	}
	return hasThenCallbackSignature(typeChecker, node, typeChecker.GetTypeOfSymbolAtLocation(then, node))
}

// isThenableType mirrors utils.IsThenableType without materializing singleton
// slices for the overwhelmingly common non-union path.
func isThenableType(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if t == nil {
		t = typeChecker.GetTypeAtLocation(node)
	}
	t = checker.Checker_getApparentType(typeChecker, t)
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if isThenableTypePart(typeChecker, node, typePart) {
				return true
			}
		}
		return false
	}
	return isThenableTypePart(typeChecker, node, t)
}

func hasAsyncIterator(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, typePart := range t.Types() {
			if hasAsyncIterator(typeChecker, typePart) {
				return true
			}
		}
		return false
	}
	return utils.GetWellKnownSymbolPropertyOfType(t, "asyncIterator", typeChecker) != nil
}

var RequireAwaitRule = rule.CreateRule(rule.Rule{
	Name:             "require-await",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var scopes []scopeInfo

		enterFunction := func(node *ast.Node) {
			scope := scopeInfo{
				hasAwait:      false,
				isAsyncYield:  false,
				functionFlags: ast.FunctionFlagsNormal,
			}
			if utils.HasNonEmptyFunctionBody(node) {
				scope.functionFlags = ast.GetFunctionFlags(node)
			}
			scopes = append(scopes, scope)
		}

		exitFunction := func(node *ast.Node) {
			scope := &scopes[len(scopes)-1]
			if scope.functionFlags&ast.FunctionFlagsAsync != 0 && !scope.hasAwait && (scope.functionFlags&ast.FunctionFlagsGenerator == 0 || !scope.isAsyncYield) {
				// TODO(port): implement suggestions
				// // If the function belongs to a method definition or
				// // property, then the function's range may not include the
				// // `async` keyword and we should look at the parent instead.
				// const nodeWithAsyncKeyword =
				//   (node.parent.type === AST_NODE_TYPES.MethodDefinition &&
				//     node.parent.value === node) ||
				//   (node.parent.type === AST_NODE_TYPES.Property &&
				//     node.parent.method &&
				//     node.parent.value === node)
				//     ? node.parent
				//     : node;
				//
				// const asyncToken = nullThrows(
				//   context.sourceCode.getFirstToken(
				//     nodeWithAsyncKeyword,
				//     token => token.value === 'async',
				//   ),
				//   'The node is an async function, so it must have an "async" token.',
				// );
				//
				// const asyncRange: Readonly<AST.Range> = [
				//   asyncToken.range[0],
				//   nullThrows(
				//     context.sourceCode.getTokenAfter(asyncToken, {
				//       includeComments: true,
				//     }),
				//     'There will always be a token after the "async" keyword.',
				//   ).range[0],
				// ] as const;
				//
				// // Removing the `async` keyword can cause parsing errors if the
				// // current statement is relying on automatic semicolon insertion.
				// // If ASI is currently being used, then we should replace the
				// // `async` keyword with a semicolon.
				// const nextToken = nullThrows(
				//   context.sourceCode.getTokenAfter(asyncToken),
				//   'There will always be a token after the "async" keyword.',
				// );
				// const addSemiColon =
				//   nextToken.type === AST_TOKEN_TYPES.Punctuator &&
				//   (nextToken.value === '[' || nextToken.value === '(') &&
				//   (nodeWithAsyncKeyword.type === AST_NODE_TYPES.MethodDefinition ||
				//     isStartOfExpressionStatement(nodeWithAsyncKeyword)) &&
				//   needsPrecedingSemicolon(context.sourceCode, nodeWithAsyncKeyword);
				//
				// const changes = [
				//   { range: asyncRange, replacement: addSemiColon ? ';' : undefined },
				// ];
				//
				// // If there's a return type annotation and it's a
				// // `Promise<T>`, we can also change the return type
				// // annotation to just `T` as part of the suggestion.
				// // Alternatively, if the function is a generator and
				// // the return type annotation is `AsyncGenerator<T>`,
				// // then we can change it to `Generator<T>`.
				// if (
				//   node.returnType?.typeAnnotation.type ===
				//   AST_NODE_TYPES.TSTypeReference
				// ) {
				//   if (scopeInfo.isGen) {
				//     if (hasTypeName(node.returnType.typeAnnotation, 'AsyncGenerator')) {
				//       changes.push({
				//         range: node.returnType.typeAnnotation.typeName.range,
				//         replacement: 'Generator',
				//       });
				//     }
				//   } else if (
				//     hasTypeName(node.returnType.typeAnnotation, 'Promise') &&
				//     node.returnType.typeAnnotation.typeArguments != null
				//   ) {
				//     const openAngle = nullThrows(
				//       context.sourceCode.getFirstToken(
				//         node.returnType.typeAnnotation,
				//         token =>
				//           token.type === AST_TOKEN_TYPES.Punctuator &&
				//           token.value === '<',
				//       ),
				//       'There are type arguments, so the angle bracket will exist.',
				//     );
				//     const closeAngle = nullThrows(
				//       context.sourceCode.getLastToken(
				//         node.returnType.typeAnnotation,
				//         token =>
				//           token.type === AST_TOKEN_TYPES.Punctuator &&
				//           token.value === '>',
				//       ),
				//       'There are type arguments, so the angle bracket will exist.',
				//     );
				//     changes.push(
				//       // Remove the closing angled bracket.
				//       { range: closeAngle.range, replacement: undefined },
				//       // Remove the "Promise" identifier
				//       // and the opening angled bracket.
				//       {
				//         range: [
				//           node.returnType.typeAnnotation.typeName.range[0],
				//           openAngle.range[1],
				//         ],
				//         replacement: undefined,
				//       },
				//     );
				//   }
				// }
				//
				// context.report({
				//   loc: getFunctionHeadLoc(node, context.sourceCode),
				//   node,
				//   messageId: 'missingAwait',
				//   data: {
				//     name: upperCaseFirst(getFunctionNameWithKind(node)),
				//   },
				//   suggest: [
				//     {
				//       messageId: 'removeAsync',
				//       fix: (fixer): RuleFix[] =>
				//         changes.map(change =>
				//           change.replacement != null
				//             ? fixer.replaceTextRange(change.range, change.replacement)
				//             : fixer.removeRange(change.range),
				//         ),
				//     },
				//   ],
				// });
				// TODO(port): getFunctionHeadLoc
				ctx.ReportNode(node, buildMissingAwaitMessage())
			}

			scopes = scopes[:len(scopes)-1]
		}

		markAsHasAwait := func() {
			if len(scopes) != 0 {
				scopes[len(scopes)-1].hasAwait = true
			}
		}

		exitArrowFunction := func(node *ast.Node) {
			scope := &scopes[len(scopes)-1]
			if scope.functionFlags&ast.FunctionFlagsAsync != 0 && !scope.hasAwait {
				body := ast.SkipParentheses(node.Body())
				if !ast.IsBlock(body) &&
					!ast.IsAwaitExpression(body) &&
					isThenableType(ctx.TypeChecker, body, ctx.TypeChecker.GetTypeAtLocation(body)) {
					scope.hasAwait = true
				}
			}
			exitFunction(node)
		}

		return rule.RuleListeners{
			// from isFunctionLikeDeclarationKind
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindConstructor:                              enterFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitArrowFunction,

			ast.KindAwaitExpression: func(node *ast.Node) { markAsHasAwait() },
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier != nil {
					markAsHasAwait()
				}
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarAwaitUsing(node) {
					markAsHasAwait()
				}
			},
			/**
			 * Mark `scopeInfo.isAsyncYield` to `true` if it
			 *  1) delegates async generator function
			 *    or
			 *  2) yields thenable type
			 */
			rule.ListenerOnExit(ast.KindYieldExpression): func(node *ast.Node) {
				if len(scopes) == 0 {
					return
				}
				scope := &scopes[len(scopes)-1]
				argument := node.Expression()
				if scope.hasAwait ||
					scope.isAsyncYield ||
					scope.functionFlags&ast.FunctionFlagsAsyncGenerator != ast.FunctionFlagsAsyncGenerator ||
					argument == nil {
					return
				}

				if ast.IsLiteralExpression(argument) {
					// ignoring this as for literals we don't need to check the definition
					// eg : async function* run() { yield* 1 }
					return
				}

				if node.AsYieldExpression().AsteriskToken == nil {
					if isThenableType(ctx.TypeChecker, argument, ctx.TypeChecker.GetTypeAtLocation(argument)) {
						scope.isAsyncYield = true
					}
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(argument)
				if hasAsyncIterator(ctx.TypeChecker, t) {
					scope.isAsyncYield = true
				}
			},
			rule.ListenerOnExit(ast.KindReturnStatement): func(node *ast.Node) {
				if len(scopes) == 0 {
					return
				}
				scope := &scopes[len(scopes)-1]
				if scope.hasAwait || scope.functionFlags&ast.FunctionFlagsAsync == 0 {
					return
				}

				expr := node.Expression()
				if expr != nil && isThenableType(ctx.TypeChecker, expr, ctx.TypeChecker.GetTypeAtLocation(expr)) {
					scope.hasAwait = true
				}
			},
		}
	},
})
