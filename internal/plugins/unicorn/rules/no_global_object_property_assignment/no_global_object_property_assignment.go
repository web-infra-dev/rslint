package no_global_object_property_assignment

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "no-global-object-property-assignment"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not assign properties on the global object.",
}

var globalObjectNames = map[string]struct{}{
	"global":     {},
	"globalThis": {},
	"self":       {},
	"window":     {},
}

var NoGlobalObjectPropertyAssignmentRule = rule.Rule{
	Name:   "unicorn/no-global-object-property-assignment",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		propertyNames := utils.NewStaticStringEvaluatorWithReferenceResolver(
			ctx.TypeChecker,
			ctx.SourceFile,
			ctx.Refs,
		)
		propertyNames.GlobalAccess = ctx.Globals.Access

		check := func(node *ast.Node) {
			object := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
			if !ast.IsIdentifier(object) {
				return
			}
			if _, ok := globalObjectNames[object.Text()]; !ok || !unicornutil.IsGlobalReference(ctx, object) {
				return
			}
			assignmentTarget := node
			for assignmentTarget.Parent != nil && utils.SkipAssertionsAndParens(assignmentTarget.Parent) == node {
				assignmentTarget = assignmentTarget.Parent
			}
			if !ast.IsAssignmentTarget(assignmentTarget) {
				return
			}
			knownStringKey := node.Kind == ast.KindElementAccessExpression &&
				ast.IsStringLiteralLike(node.AsElementAccessExpression().ArgumentExpression)
			if !knownStringKey && !propertyNames.HasStaticAccessExpressionKey(node) {
				return
			}

			ctx.ReportNode(node, message)
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: check,
			ast.KindElementAccessExpression:  check,
		}
	},
}
