package no_exports_assign

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-exports-assign.js
var NoExportsAssignRule = rule.Rule{
	Name:   "node/no-exports-assign",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		exportsAccess := ctx.Globals.Access("exports")
		// Module-local declarations cannot supply the global variable this
		// rule requires. Globals already includes inline directives.
		if exportsAccess != utils.GlobalAccessReadonly && exportsAccess != utils.GlobalAccessWritable &&
			ctx.Refs.HasNonGlobalProgramScope() {
			return nil
		}

		var scopes *scope.Manager
		isGlobalIdentifier := func(node *ast.Node, name string) bool {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil || node.Kind != ast.KindIdentifier || node.Text() != name {
				return false
			}
			// A direct program expression with no file-local declaration cannot
			// be shadowed. RefStore distinguishes authored locals from tsgo's
			// implicit CommonJS exports binding and JSDoc declarations.
			// Once scopes are built, reuse them without another fast-path query.
			if scopes != nil || !isDirectProgramExpression(node) ||
				(ctx.SourceFile.Locals[name] != nil && !ctx.Refs.IsGlobalNameReference(node, name, scope.ReferenceDual.DeclarationMeaning())) {
				// Upstream's syntactic lookup includes type-only declarations and
				// sees function-body bindings from parameter defaults. Ordinary
				// reference resolution alone cannot model these nested scopes.
				if scopes == nil {
					scopes = scope.Build(ctx.SourceFile, scope.Options{})
				}
				for current := scopes.Acquire(node); current != nil; current = current.Parent {
					if len(current.Declarations(name)) != 0 {
						return current == scopes.Global && !ctx.Refs.HasNonGlobalProgramScope()
					}
				}
			}
			access := ctx.Globals.Access(name)
			return access == utils.GlobalAccessReadonly || access == utils.GlobalAccessWritable
		}
		isModuleExports := func(node *ast.Node) bool {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil || node.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(node) {
				return false
			}
			member := node.AsPropertyAccessExpression()
			return member.Name().Kind == ast.KindIdentifier && member.Name().Text() == "exports" &&
				isGlobalIdentifier(member.Expression, "module")
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				assignment := node.AsBinaryExpression()
				if !ast.IsAssignmentOperator(assignment.OperatorToken.Kind) {
					return
				}
				left := utils.ESTreeRuntimeExpression(assignment.Left)
				if left == nil || left.Kind != ast.KindIdentifier || left.Text() != "exports" ||
					utils.IsDefaultValueInDestructuringAssignment(node) || !isGlobalIdentifier(left, "exports") {
					return
				}
				// module.exports = exports = {}
				if parent := utils.ESTreeParent(node); parent != nil && ast.IsAssignmentExpression(parent, false) &&
					!utils.IsDefaultValueInDestructuringAssignment(parent) {
					outer := parent.AsBinaryExpression()
					if utils.ESTreeRuntimeExpression(outer.Right) == node && isModuleExports(outer.Left) {
						return
					}
				}
				// exports = module.exports = {}
				if right := utils.ESTreeRuntimeExpression(assignment.Right); ast.IsAssignmentExpression(right, false) &&
					isModuleExports(right.AsBinaryExpression().Left) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "forbidden",
					Description: "Unexpected assignment to 'exports' variable. Use 'module.exports' instead.",
				})
			},
		}
	},
}

// Limit the fast path to expressions whose ancestors cannot introduce a
// lexical scope. Blocks, functions, classes, namespaces and patterns use the
// shared ESLint scope model instead of TypeScript's broader top-level test.
func isDirectProgramExpression(node *ast.Node) bool {
	for parent := utils.ESTreeParent(node); parent != nil; parent = utils.ESTreeParent(parent) {
		switch parent.Kind {
		case ast.KindBinaryExpression, ast.KindPropertyAccessExpression:
			continue
		case ast.KindExpressionStatement:
			return parent.Parent != nil && parent.Parent.Kind == ast.KindSourceFile
		default:
			return false
		}
	}
	return false
}
