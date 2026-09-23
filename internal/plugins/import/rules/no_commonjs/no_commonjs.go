package no_commonjs

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed no_commonjs.schema.json
var schemaJSON []byte

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-commonjs.js
var NoCommonjsRule = rule.Rule{
	Name:   "import/no-commonjs",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowPrimitiveModules, allowRequire, allowConditionalRequire := false, false, true
		if len(options) > 0 {
			switch option := options[0].(type) {
			case string:
				allowPrimitiveModules = option == "allow-primitive-modules"
			case map[string]any:
				allowPrimitiveModules, _ = option["allowPrimitiveModules"].(bool)
				allowRequire, _ = option["allowRequire"].(bool)
				if value, ok := option["allowConditionalRequire"].(bool); ok {
					allowConditionalRequire = value
				}
			}
		}
		sourceType := ctx.LanguageOptions.EffectiveSourceType()

		checkMember := func(node *ast.Node) {
			if utils.IsInJsxTagName(node) {
				return
			}
			object, property := utils.MemberExpressionParts(node)
			object = utils.ESTreeRuntimeExpression(object)
			property = utils.ESTreeRuntimeExpression(property)
			if object == nil || object.Kind != ast.KindIdentifier {
				return
			}
			switch object.Text() {
			case "module":
				// Upstream checks property.name, including computed identifiers
				// and private names, but not string literals such as ["exports"].
				isExports := property != nil && ((property.Kind == ast.KindIdentifier && property.Text() == "exports") ||
					(property.Kind == ast.KindPrivateIdentifier && property.Text() == "#exports"))
				if !isExports {
					return
				}
				// An outer optional chain has an ESTree ChainExpression parent.
				if allowPrimitiveModules && !ast.IsOptionalChain(node) {
					parent := utils.ESTreeParent(node)
					if parent != nil && ast.IsAssignmentExpression(parent, false) &&
						utils.ESTreeRuntimeExpression(parent.AsBinaryExpression().Right).Kind != ast.KindObjectLiteralExpression {
						return
					}
				}
			case "exports":
				current := scopeanalysis.Declarations(ctx).Acquire(node)
				// Only declarations in the current scope count upstream; an
				// outer binding must not suppress the report.
				if len(current.Declarations("exports")) > 0 {
					return
				}
				// Configured globals belong to the global scope, outside an
				// ES module or a JavaScript CommonJS wrapper.
				if current.Kind == scope.KindGlobal && !ctx.Refs.HasNonGlobalProgramScope() && ctx.Globals.Access("exports").IsDeclared() {
					return
				}
			default:
				return
			}
			// Upstream reports literal messages without message IDs or edits.
			ctx.ReportNode(node, rule.RuleMessage{Description: `Expected "export" or "export default"`})
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindElementAccessExpression:  checkMember,
			ast.KindQualifiedName:            checkMember,
			ast.KindCallExpression: func(node *ast.Node) {
				if allowRequire || sourceType != "module" {
					return
				}
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.Text() != "require" ||
					call.Arguments == nil || len(call.Arguments.Nodes) != 1 ||
					!ast.IsStringLiteralLike(utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])) {
					return
				}
				if allowConditionalRequire && ast.FindAncestor(node.Parent, isConditional) != nil {
					return
				}
				// The scope utility collapses ESLint's module scope into Global.
				if scopeanalysis.Declarations(ctx).Acquire(node).VariableScope().Kind != scope.KindGlobal {
					return
				}
				ctx.ReportNode(callee, rule.RuleMessage{Description: `Expected "import" instead of "require()"`})
			},
		}
	},
}

func isConditional(node *ast.Node) bool {
	return node.Kind == ast.KindIfStatement || node.Kind == ast.KindTryStatement ||
		node.Kind == ast.KindConditionalExpression || ast.IsLogicalOrCoalescingBinaryExpression(node)
}
