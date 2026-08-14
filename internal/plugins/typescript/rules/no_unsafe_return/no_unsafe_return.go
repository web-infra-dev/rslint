package no_unsafe_return

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const unsafeReturnThisHelp = "\nYou can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function."

func buildUnsafeReturnMessage(t string) rule.RuleMessage {
	var description string
	switch t {
	case "error":
		description = "Unsafe return of a value of type error."
	case "`any`":
		description = "Unsafe return of a value of type `any`."
	case "`Promise<any>`":
		description = "Unsafe return of a value of type `Promise<any>`."
	case "`any[]`":
		description = "Unsafe return of a value of type `any[]`."
	default:
		description = "Unsafe return of a value of type " + t + "."
	}
	return rule.RuleMessage{
		Id:          "unsafeReturn",
		Description: description,
	}
}
func buildUnsafeReturnAssignmentMessage(sender, receiver string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeReturnAssignment",
		Description: "Unsafe return of type `" + sender + "` from function with return type `" + receiver + "`.",
	}
}
func buildUnsafeReturnThisMessage(t string) rule.RuleMessage {
	var description string
	switch t {
	case "error":
		description = "Unsafe return of a value of type `error`. `this` is typed as `any`." + unsafeReturnThisHelp
	case "`any`":
		description = "Unsafe return of a value of type ``any``. `this` is typed as `any`." + unsafeReturnThisHelp
	case "`Promise<any>`":
		description = "Unsafe return of a value of type ``Promise<any>``. `this` is typed as `any`." + unsafeReturnThisHelp
	case "`any[]`":
		description = "Unsafe return of a value of type ``any[]``. `this` is typed as `any`." + unsafeReturnThisHelp
	default:
		description = "Unsafe return of a value of type `" + t + "`. `this` is typed as `any`." + unsafeReturnThisHelp
	}
	return rule.RuleMessage{
		Id:          "unsafeReturnThis",
		Description: description,
	}
}

func discriminateReturnType(
	t *checker.Type,
	typeChecker *checker.Checker,
	sourceProgram *program.Program,
	node *ast.Node,
) utils.DiscriminatedAnyType {
	if utils.IsTypeAnyType(t) {
		return utils.DiscriminatedAnyTypeAny
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsPrimitive|checker.TypeFlagsUnknown|checker.TypeFlagsNever) {
		return utils.DiscriminatedAnyTypeSafe
	}
	if utils.IsTypeAnyArrayType(t, typeChecker) {
		return utils.DiscriminatedAnyTypeAnyArray
	}
	return utils.DiscriminateAnyType(t, typeChecker, sourceProgram, node)
}

var NoUnsafeReturnRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-return",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		compilerOptions := ctx.Program().Options()
		// When noImplicitThis is not enabled (considering strict mode), object literal methods
		// can have implicit any this. We need to use IsStrictCompilerOptionEnabled to properly
		// handle the case where noImplicitThis is inherited from strict mode.
		shouldCheckImplicitAnyThis := !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.NoImplicitThis)

		checkReturn := func(
			returnNode *ast.Node,
			reportingNode *ast.Node,
		) {
			functionNode := utils.GetParentFunctionNode(returnNode)
			if functionNode == nil {
				return
			}

			returnNodeType := ctx.TypeChecker.GetTypeAtLocation(returnNode)
			anyType := discriminateReturnType(
				returnNodeType,
				ctx.TypeChecker,
				ctx.Program(),
				returnNode,
			)
			if anyType == utils.DiscriminatedAnyTypeSafe &&
				utils.IsTypeFlagSet(returnNodeType, checker.TypeFlagsPrimitive|checker.TypeFlagsUnknown|checker.TypeFlagsNever) {
				return
			}

			// function expressions will not have their return type modified based on receiver typing
			// so we have to use the contextual typing in these cases, i.e.
			// const foo1: () => Set<string> = () => new Set<any>();
			// the return type of the arrow function is Set<any> even though the variable is typed as Set<string>
			var functionType *checker.Type
			if ast.IsFunctionExpression(functionNode) || ast.IsArrowFunction(functionNode) {
				functionType = utils.GetContextualType(ctx.TypeChecker, functionNode)
			}
			if functionType == nil {
				functionType = ctx.TypeChecker.GetTypeAtLocation(functionNode)
			}
			var callSignatures []*checker.Signature
			callSignaturesLoaded := false
			// If there is an explicit type annotation *and* that type matches the actual
			// function return type, we shouldn't complain (it's intentional, even if unsafe)
			if functionNode.Type() != nil {
				callSignatures = utils.CollectAllCallSignatures(ctx.TypeChecker, functionType)
				callSignaturesLoaded = true
				for _, signature := range callSignatures {
					signatureReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

					if returnNodeType == signatureReturnType ||
						utils.IsTypeFlagSet(
							signatureReturnType,
							checker.TypeFlagsAny|checker.TypeFlagsUnknown,
						) {
						return
					}
					if ast.HasSyntacticModifier(functionNode, ast.ModifierFlagsAsync) {
						awaitedSignatureReturnType := checker.Checker_getAwaitedType(ctx.TypeChecker, signatureReturnType)
						awaitedReturnNodeType := checker.Checker_getAwaitedType(ctx.TypeChecker, returnNodeType)

						if awaitedSignatureReturnType == awaitedReturnNodeType || (awaitedSignatureReturnType != nil && utils.IsTypeFlagSet(awaitedSignatureReturnType, checker.TypeFlagsAny|checker.TypeFlagsUnknown)) {
							return
						}
					}
				}
			}

			if anyType != utils.DiscriminatedAnyTypeSafe {
				if !callSignaturesLoaded {
					callSignatures = utils.CollectAllCallSignatures(ctx.TypeChecker, functionType)
				}
				// Allow cases when the declared return type of the function is either unknown or unknown[]
				// and the function is returning any or any[].
				for _, signature := range callSignatures {
					functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
					if anyType == utils.DiscriminatedAnyTypeAny && utils.IsTypeUnknownType(functionReturnType) {
						return
					}

					if anyType == utils.DiscriminatedAnyTypeAnyArray && utils.IsTypeUnknownArrayType(functionReturnType, ctx.TypeChecker) {
						return
					}
					if anyType == utils.DiscriminatedAnyTypePromiseAny {
						awaitedType := checker.Checker_getAwaitedType(ctx.TypeChecker, functionReturnType)
						if awaitedType != nil && utils.IsTypeUnknownType(awaitedType) {
							return
						}
					}
				}

				if anyType == utils.DiscriminatedAnyTypePromiseAny && !ast.HasSyntacticModifier(functionNode, ast.ModifierFlagsAsync) {
					return
				}

				var typeString string
				if anyType == utils.DiscriminatedAnyTypeAny &&
					utils.IsIntrinsicErrorType(utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, returnNode)) {
					typeString = "error"
				} else if anyType == utils.DiscriminatedAnyTypeAny {
					typeString = "`any`"
				} else if anyType == utils.DiscriminatedAnyTypePromiseAny {
					typeString = "`Promise<any>`"
				} else {
					typeString = "`any[]`"
				}

				// Check this type only when noImplicitThis is not enabled
				// This matches upstream typescript-eslint/no-unsafe-return behavior
				// See: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-unsafe-return.ts
				if shouldCheckImplicitAnyThis {
					// `return this`
					thisExpression := utils.GetThisExpression(returnNode)
					if thisExpression != nil && utils.IsTypeAnyType(utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)) {
						ctx.ReportNode(reportingNode, buildUnsafeReturnThisMessage(typeString))
						return
					}
				}
				// If the function return type was not unknown/unknown[], mark usage as unsafeReturn.
				ctx.ReportNode(reportingNode, buildUnsafeReturnMessage(typeString))
				return
			}

			if !callSignaturesLoaded || utils.IsUnionType(functionType) || utils.IsIntersectionType(functionType) {
				callSignatures = utils.GetCallSignatures(ctx.TypeChecker, functionType)
			}
			if len(callSignatures) < 1 {
				return
			}

			signature := callSignatures[0]
			functionReturnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)

			receiver, sender, unsafe := utils.IsUnsafeAssignment(
				returnNodeType,
				functionReturnType,
				ctx.TypeChecker,
				returnNode,
			)

			if !unsafe {
				return
			}

			ctx.ReportNode(reportingNode, buildUnsafeReturnAssignmentMessage(ctx.TypeChecker.TypeToString(sender), ctx.TypeChecker.TypeToString(receiver)))
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if !ast.IsBlock(body) {
					body = ast.SkipParentheses(body)
					checkReturn(body, body)
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				argument := node.Expression()
				if argument == nil {
					return
				}

				checkReturn(argument, node)
			},
		}
	},
})
