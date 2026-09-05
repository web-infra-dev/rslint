package no_unsafe_member_access

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_unsafe_member_access.schema.json
var schemaJSON []byte

func buildUnsafeComputedMemberAccessMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeComputedMemberAccess",
		Description: "Computed name " + property + " resolves to an `any` value.",
	}
}

func buildErrorComputedMemberAccessMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorComputedMemberAccess",
		Description: "The type of computed name " + property + " cannot be resolved.",
	}
}

func buildUnsafeMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeMemberExpression",
		Description: "Unsafe member access " + property + " on an `any` value.",
	}
}

func buildErrorMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorMemberExpression",
		Description: "Unsafe member access " + property + " on a type that cannot be resolved.",
	}
}

func buildUnsafeThisMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeThisMemberExpression",
		Description: "Unsafe member access " + property + " on an `any` value. `this` is typed as `any`.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

func buildErrorThisMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "errorThisMemberExpression",
		Description: "Unsafe member access " + property + ". The type of `this` cannot be resolved.\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

type state uint8

const (
	stateUnsafe state = iota + 1
	stateSafe
	stateChained
)

type noUnsafeMemberAccessOptions struct {
	allowOptionalChaining bool
}

func parseOptions(options []any) noUnsafeMemberAccessOptions {
	if len(options) == 0 {
		return noUnsafeMemberAccessOptions{}
	}

	optionsMap, _ := options[0].(map[string]any)
	allowOptionalChaining, _ := optionsMap["allowOptionalChaining"].(bool)
	return noUnsafeMemberAccessOptions{allowOptionalChaining: allowOptionalChaining}
}

func isIgnoredHeritageClause(node *ast.Node) bool {
	clause := node.AsHeritageClause()
	return clause.Token == ast.KindImplementsKeyword || ast.IsInterfaceDeclaration(node.Parent)
}

var NoUnsafeMemberAccessRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-member-access",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		compilerOptions := ctx.Program().Options()
		isNoImplicitThis := compilerOptions.GetStrictOptionValue(compilerOptions.NoImplicitThis)

		stateCache := map[*ast.Node]state{}
		// Traversal is source ordered, so the latest ignored heritage range is
		// enough to classify all of its descendants without an exit listener.
		ignoredHeritageStart := 0
		ignoredHeritageEnd := 0

		var checkMemberExpression func(node *ast.Node) state
		checkMemberExpression = func(node *ast.Node) state {
			if ignoredHeritageEnd != 0 && node.Pos() >= ignoredHeritageStart && node.End() <= ignoredHeritageEnd {
				return stateSafe
			}

			// ESTree's `node.optional` maps to the link that owns the literal `?.`,
			// not every tsgo node carrying the optional-chain flag.
			if opts.allowOptionalChaining && ast.IsOptionalChainRoot(node) {
				object := ast.SkipParentheses(node.Expression())
				if ast.IsAccessExpression(object) {
					// ESLint sorts diagnostics by source position. Checking the optional
					// link's receiver first preserves that order in rslint's
					// streaming reporter without changing the cached state.
					checkMemberExpression(object)
				}
				stateCache[node] = stateChained
				return stateChained
			}

			if cachedState, ok := stateCache[node]; ok {
				return cachedState
			}

			expression := node.Expression()
			unwrappedExpression := ast.SkipParentheses(expression)
			// Parentheses are erased in ESTree for ordinary accesses, but an optional
			// chain is represented by a ChainExpression boundary and must not recurse.
			if ast.IsAccessExpression(unwrappedExpression) {
				objectState := checkMemberExpression(unwrappedExpression)
				if expression == unwrappedExpression || !ast.IsOptionalChain(unwrappedExpression) {
					if objectState == stateUnsafe {
						// The inner access was already reported, so suppress the rest of
						// the same ordinary member chain.
						stateCache[node] = objectState
						return objectState
					}
				}
			}
			if ast.IsCallExpression(unwrappedExpression) {
				callee := ast.SkipParentheses(unwrappedExpression.Expression())
				if ast.IsAccessExpression(callee) {
					// Calls end member-chain suppression. Check the unsafe callee first
					// so rslint's streaming output retains ESLint's source order.
					checkMemberExpression(callee)
				}
			}

			t := ctx.TypeChecker.GetTypeAtLocation(expression)
			currentState := stateSafe
			if utils.IsTypeAnyType(t) {
				currentState = stateUnsafe
			}
			stateCache[node] = currentState

			if currentState == stateUnsafe {
				property, propertyName := utils.GetPropertyInfo(ctx.SourceFile, node)
				message := buildUnsafeMemberExpressionMessage(propertyName)
				if utils.IsIntrinsicErrorType(t) {
					message = buildErrorMemberExpressionMessage(propertyName)
				}

				if !isNoImplicitThis {
					thisExpression := utils.GetThisExpression(node)
					if thisExpression != nil {
						thisType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
						if utils.IsTypeAnyType(thisType) {
							message = buildUnsafeThisMemberExpressionMessage(propertyName)
							if utils.IsIntrinsicErrorType(thisType) {
								message = buildErrorThisMemberExpressionMessage(propertyName)
							}
						}
					}
				}

				ctx.ReportNode(property, message)
			}

			return currentState
		}

		return rule.RuleListeners{
			ast.KindHeritageClause: func(node *ast.Node) {
				if isIgnoredHeritageClause(node) {
					ignoredHeritageStart = node.Pos()
					ignoredHeritageEnd = node.End()
				}
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberExpression(node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if opts.allowOptionalChaining && ast.IsOptionalChainRoot(node) {
					checkMemberExpression(node)
					return
				}

				checkMemberExpression(node)

				arg := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
				// ESTree represents all of these as Literal nodes.
				if ast.IsLiteralExpression(arg) || ast.IsBooleanLiteral(arg) || arg.Kind == ast.KindNullKeyword {
					return
				}

				// All update expressions have type number, even when their operand is
				// any, because JavaScript returns NaN for non-numeric values.
				unaryOperatorKind := ast.KindUnknown
				if ast.IsPrefixUnaryExpression(arg) {
					unaryOperatorKind = arg.AsPrefixUnaryExpression().Operator
				} else if arg.Kind == ast.KindPostfixUnaryExpression {
					unaryOperatorKind = arg.AsPostfixUnaryExpression().Operator
				}
				if unaryOperatorKind == ast.KindPlusPlusToken || unaryOperatorKind == ast.KindMinusMinusToken {
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(arg)
				if utils.IsTypeAnyType(t) {
					loc := utils.TrimNodeTextRange(ctx.SourceFile, arg)
					propertyName := "[" + ctx.SourceFile.Text()[loc.Pos():loc.End()] + "]"
					message := buildUnsafeComputedMemberAccessMessage(propertyName)
					if utils.IsIntrinsicErrorType(t) {
						message = buildErrorComputedMemberAccessMessage(propertyName)
					}
					ctx.ReportNode(arg, message)
				}
			},
		}
	},
})
