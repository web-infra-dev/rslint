package global_require

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/global-require.js
var GlobalRequireRule = rule.Rule{
	Name:   "node/global-require",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "require" ||
					hasAcceptableParents(node) ||
					ctx.Refs.ResolveInFileWithMeaning(callee, scope.ReferenceValue.DeclarationMeaning()) != nil {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Unexpected require().",
				})
			},
		}
	},
}

func hasAcceptableParents(node *ast.Node) bool {
	for current := node; current != nil; current = utils.ESTreeParent(current) {
		// ESTree wraps an optional chain in a disallowed ChainExpression,
		// including when the require call itself is optional.
		if ast.IsOptionalChain(current) {
			return false
		}
		switch current.Kind {
		case ast.KindSourceFile, ast.KindVariableDeclaration, ast.KindVariableDeclarationList,
			ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindExpressionStatement, ast.KindConditionalExpression:
			continue
		case ast.KindCallExpression:
			// Dynamic import is an ImportExpression in ESTree.
			if ast.IsImportCall(current) {
				return false
			}
		case ast.KindVariableStatement:
			// ExportNamedDeclaration is a separate ancestor in ESTree.
			if ast.HasSyntacticModifier(current, ast.ModifierFlagsExport) {
				return false
			}
		case ast.KindBinaryExpression:
			if !ast.IsAssignmentExpression(current, false) || utils.IsDefaultValueInDestructuringAssignment(current) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
