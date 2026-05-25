package no_unsafe_call

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnsafeCallMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeCall",
		Description: fmt.Sprintf("Unsafe call of a(n) %v typed value.", t),
	}
}
func buildUnsafeCallThisMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "unsafeCallThis",
		Description: fmt.Sprintf("Unsafe call of a(n) %v typed value. `this` is typed as %v.\n", t, t) +
			"You can try to fix this by turning on the `noImplicitThis` compiler option, or adding a `this` parameter to the function.",
	}
}
func buildUnsafeNewMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeNew",
		Description: fmt.Sprintf("Unsafe construction of a(n) %v typed value.", t),
	}
}
func buildUnsafeTemplateTagMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTemplateTag",
		Description: fmt.Sprintf("Unsafe use of a(n) %v typed template tag.", t),
	}
}

var NoUnsafeCallRule = rule.CreateRule(rule.Rule{
	Name:             "no-unsafe-call",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		compilerOptions := ctx.Program.Options()
		// When noImplicitThis is not enabled (considering strict mode), object literal methods
		// can have implicit any this. We need to use IsStrictCompilerOptionEnabled to properly
		// handle the case where noImplicitThis is inherited from strict mode.
		shouldCheckImplicitAnyThis := !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.NoImplicitThis)

		checkCall := func(
			node *ast.Node,
			reportingNode *ast.Node,
			messageBuilder func(t string) rule.RuleMessage,
			newCall bool,
		) {
			// Note: Control flow differs from upstream typescript-eslint/no-unsafe-call
			//
			// Upstream logic:
			//   1. Check if callee type is any
			//   2. If any, check this (when !noImplicitThis) to choose message
			//   3. Report with appropriate message
			//
			// Our logic:
			//   1. Check this first (using IsInObjectLiteralMethod helper)
			//   2. If in object literal method without noImplicitThis, report and return
			//   3. Then check callee type
			//
			// Rationale: typescript-go's type checker may not automatically infer implicit any
			// for this in object literal methods when noImplicitThis=false (unlike TypeScript's
			// official compiler). We explicitly check IsInObjectLiteralMethod to compensate for
			// this limitation and ensure consistent behavior with upstream typescript-eslint.
			//
			// See: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-unsafe-call.ts

			// Check for unsafe this calls first, before checking the callee type
			thisExpression := utils.GetThisExpression(node)
			if thisExpression != nil {
				// Check the this type directly for actual any types
				thisType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, thisExpression)
				if utils.IsTypeAnyType(thisType) {
					msg := "`any`"
					if utils.IsIntrinsicErrorType(thisType) {
						msg = "`error` type"
					}

					// When noImplicitThis is not enabled and we're in an object literal method,
					// use a more specific error message suggesting to enable noImplicitThis
					if shouldCheckImplicitAnyThis {
						functionNode := utils.GetParentFunctionNode(thisExpression)
						if functionNode != nil && utils.IsInObjectLiteralMethod(functionNode) {
							ctx.ReportNode(reportingNode, buildUnsafeCallThisMessage(msg))
							return
						}
					}

					ctx.ReportNode(reportingNode, buildUnsafeCallThisMessage(msg))
					return
				}
			}

			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)

			if utils.IsTypeAnyType(t) {
				isErrorType := utils.IsIntrinsicErrorType(t)

				msg := "`any`"
				if isErrorType {
					msg = "`error` type"
				}
				ctx.ReportNode(reportingNode, messageBuilder(msg))
				return
			}

			if utils.IsBuiltinSymbolLike(ctx.Program, ctx.TypeChecker, t, "Function") {
				// this also matches subtypes of `Function`, like `interface Foo extends Function {}`.
				//
				// For weird TS reasons that I don't understand, these are
				//
				// safe to construct if:
				// - they have at least one call signature _that is not void-returning_,
				// - OR they have at least one construct signature.
				//
				// safe to call (including as template) if:
				// - they have at least one call signature
				// - OR they have at least one construct signature.

				constructSignatures := utils.GetConstructSignatures(ctx.TypeChecker, t)
				if len(constructSignatures) > 0 {
					return
				}

				callSignatures := utils.GetCallSignatures(ctx.TypeChecker, t)
				if newCall {
					if utils.Some(callSignatures, func(signature *checker.Signature) bool {
						return !utils.IsIntrinsicVoidType(checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature))
					}) {
						return
					}
				} else if len(callSignatures) > 0 {
					return
				}

				ctx.ReportNode(reportingNode, messageBuilder("`Function`"))
				return
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := node.Expression()
				if callee.Kind == ast.KindImportKeyword {
					return
				}
				checkCall(callee, callee, buildUnsafeCallMessage, false)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				callee := node.Expression()
				checkCall(callee, node, buildUnsafeNewMessage, true)
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tag := node.AsTaggedTemplateExpression().Tag
				checkCall(tag, tag, buildUnsafeTemplateTagMessage, false)
			},
		}
	},
})
