package no_unsafe_member_access

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnsafeComputedMemberAccessMessage(property, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeComputedMemberAccess",
		Description: fmt.Sprintf("Computed name %v resolves to an %v value.", property, t),
	}
}
func buildUnsafeMemberExpressionMessage(property, t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeMemberExpression",
		Description: fmt.Sprintf("Unsafe member access %v on an %v value.", property, t),
	}
}
func buildUnsafeThisMemberExpressionMessage(property string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unsafeThisMemberExpression",
		Description: fmt.Sprintf("Unsafe member access %v on an `any` value. `this` is typed as `any`.", property) +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}

type state uint8

const (
	stateUnsafe state = iota
	stateSafe   state = iota
)

func createDataType(t *checker.Type) string {
	if utils.IsIntrinsicErrorType(t) {
		return "`error` typed"
	}
	return "`any`"
}

var NoUnsafeMemberAccessRule = rule.CreateRule(rule.Rule{
	Name: "no-unsafe-member-access",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		// When noImplicitThis is not enabled (considering strict mode), object literal methods
		// can have implicit any this. We need to use IsStrictCompilerOptionEnabled to properly
		// handle the case where noImplicitThis is inherited from strict mode.
		shouldCheckImplicitAnyThis := !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.NoImplicitThis)

		stateCache := map[*ast.Node]state{}

		var checkMemberExpression func(node *ast.Node) state
		checkMemberExpression = func(node *ast.Node) state {
			cachedState, ok := stateCache[node]
			if ok {
				return cachedState
			}

			// TODO(port-perf): couldn't it be done at visitchildren phase?
			parent := node.Parent
			for !ast.IsSourceFile(parent) {
				// ignore MemberExpressions with ancestors of type `TSClassImplements` or `TSInterfaceHeritage`
				if ast.IsHeritageClause(parent) {
					return stateSafe
				}
				parent = parent.Parent
			}

			expression := node.Expression()
			if ast.IsAccessExpression(expression) {
				objectState := checkMemberExpression(expression)
				if objectState == stateUnsafe {
					// if the object is unsafe, we know this will be unsafe as well
					// we don't need to report, as we have already reported on the inner member expr
					stateCache[node] = objectState
					return objectState
				}
			}

			// Note: Control flow differs from upstream typescript-eslint/no-unsafe-member-access
			//
			// Upstream logic:
			//   1. Check if member expression type is any
			//   2. If any, check this (when !noImplicitThis) to choose message
			//   3. Report with appropriate message
			//
			// Our logic:
			//   1. Check this first (using IsInObjectLiteralMethod helper)
			//   2. If in object literal method without noImplicitThis, report and return
			//   3. Then check member expression type
			//
			// Rationale: typescript-go's type checker may not automatically infer implicit any
			// for this in object literal methods when noImplicitThis=false (unlike TypeScript's
			// official compiler). We explicitly check IsInObjectLiteralMethod to compensate for
			// this limitation and ensure consistent behavior with upstream typescript-eslint.
			//
			// See: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-unsafe-member-access.ts

			// Check for unsafe this member access first
			thisExpression := utils.GetThisExpression(node)
			if thisExpression != nil {
				// Check the this type directly for actual any types
				thisType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
				if utils.IsTypeAnyType(thisType) {
					state := stateUnsafe
					stateCache[node] = state

					property, propertyName := utils.GetPropertyInfo(ctx.SourceFile, node)

					// When noImplicitThis is not enabled and we're in an object literal method,
					// use a more specific error message suggesting to enable noImplicitThis
					if shouldCheckImplicitAnyThis {
						functionNode := utils.GetParentFunctionNode(thisExpression)
						if functionNode != nil && utils.IsInObjectLiteralMethod(functionNode) {
							ctx.ReportNode(property, buildUnsafeThisMemberExpressionMessage(propertyName))
							return state
						}
					}

					ctx.ReportNode(property, buildUnsafeThisMemberExpressionMessage(propertyName))
					return state
				}
			}

			t := ctx.TypeChecker.GetTypeAtLocation(expression)
			state := stateSafe
			if utils.IsTypeAnyType(t) {
				state = stateUnsafe
			}
			stateCache[node] = state

			if state == stateUnsafe {
				property, propertyName := utils.GetPropertyInfo(ctx.SourceFile, node)
				ctx.ReportNode(property, buildUnsafeMemberExpressionMessage(propertyName, createDataType(t)))
			}

			return state
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberExpression(node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkMemberExpression(node)

				arg := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
				// x[1]
				if ast.IsLiteralExpression(arg) {
					// perf optimizations - literals can obviously never be `any`
					return
				}

				// x[1++] x[++x] etc
				// FUN FACT - **all** update expressions return type number, regardless of the argument's type,
				// because JS engines return NaN if there the argument is not a number.
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
					ctx.ReportNode(arg, buildUnsafeComputedMemberAccessMessage(propertyName, createDataType(t)))
				}
			},
		}
	},
})
