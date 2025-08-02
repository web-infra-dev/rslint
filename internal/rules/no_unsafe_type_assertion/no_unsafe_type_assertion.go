package no_unsafe_type_assertion

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildUnsafeOfAnyTypeAssertionMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeOfAnyTypeAssertion",
		Description: fmt.Sprintf("Unsafe assertion from %v detected: consider using type guards or a safer assertion.", t),
	}
}
func buildUnsafeToAnyTypeAssertionMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeToAnyTypeAssertion",
		Description: fmt.Sprintf("Unsafe assertion to %v detected: consider using a more specific type to ensure safety.", t),
	}
}
func buildUnsafeToUnconstrainedTypeAssertionMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeToUnconstrainedTypeAssertion",
		Description: fmt.Sprintf("Unsafe type assertion: '%v' could be instantiated with an arbitrary type which could be unrelated to the original type.", t),
	}
}
func buildUnsafeTypeAssertionMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTypeAssertion",
		Description: fmt.Sprintf("Unsafe type assertion: type '%v' is more narrow than the original type.", t),
	}
}
func buildUnsafeTypeAssertionAssignableToConstraintMessage(t string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeTypeAssertionAssignableToConstraint",
		Description: fmt.Sprintf("Unsafe type assertion: the original type is assignable to the constraint of type '%v', but '%v' could be instantiated with a different subtype of its constraint.", t, t),
	}
}

func getAnyTypeName(t *checker.Type) string {
	if utils.IsIntrinsicErrorType(t) {
		return "error typed"
	}
	return "`any`"
}

func isObjectLiteralType(t *checker.Type) bool {
	return utils.IsObjectType(t) && checker.Type_objectFlags(t)&checker.ObjectFlagsObjectLiteral != 0
}

var NoUnsafeTypeAssertionRule = rule.Rule{
	Name: "no-unsafe-type-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		checkExpression := func(node *ast.Node) {
			expression := node.Expression()
			typeAnnotation := node.Type()
			expressionType := ctx.TypeChecker.GetTypeAtLocation(expression)
			assertedType := ctx.TypeChecker.GetTypeAtLocation(typeAnnotation)

			if expressionType == assertedType {
				return
			}

			// handle cases when asserting unknown ==> any.
			if utils.IsTypeAnyType(assertedType) && utils.IsTypeUnknownType(expressionType) {
				ctx.ReportNode(node, buildUnsafeToAnyTypeAssertionMessage("`any`"))
				return
			}

			_, sender, isUnsafeExpressionAny := utils.IsUnsafeAssignment(expressionType, assertedType, ctx.TypeChecker, expression)
			if isUnsafeExpressionAny {
				ctx.ReportNode(node, buildUnsafeOfAnyTypeAssertionMessage(getAnyTypeName(sender)))
				return
			}

			_, sender, isUnsafeAssertedAny := utils.IsUnsafeAssignment(assertedType, expressionType, ctx.TypeChecker, typeAnnotation)
			if isUnsafeAssertedAny {
				ctx.ReportNode(node, buildUnsafeToAnyTypeAssertionMessage(getAnyTypeName(sender)))
				return
			}

			// Use the widened type in case of an object literal so `isTypeAssignableTo()`
			// won't fail on excess property check.
			expressionWidenedType := expressionType
			if isObjectLiteralType(expressionType) {
				expressionWidenedType = checker.Checker_getWidenedType(ctx.TypeChecker, expressionType)
			}

			if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, expressionWidenedType, assertedType) {
				return
			}

			// Produce a more specific error message when targeting a type parameter
			if utils.IsTypeParameter(assertedType) {
				assertedTypeConstraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, assertedType)
				if assertedTypeConstraint == nil {
					// asserting to an unconstrained type parameter is unsafe
					ctx.ReportNode(node, buildUnsafeToUnconstrainedTypeAssertionMessage(ctx.TypeChecker.TypeToString(assertedType)))
					return
				}

				// special case message if the original type is assignable to the
				// constraint of the target type parameter
				if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, expressionWidenedType, assertedTypeConstraint) {
					ctx.ReportNode(node, buildUnsafeTypeAssertionAssignableToConstraintMessage(ctx.TypeChecker.TypeToString(assertedType)))
					return
				}
			}

			ctx.ReportNode(node, buildUnsafeTypeAssertionMessage(ctx.TypeChecker.TypeToString(assertedType)))
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkExpression,
			ast.KindTypeAssertionExpression: checkExpression,
		}
	},
}
