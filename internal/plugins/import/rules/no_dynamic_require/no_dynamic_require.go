package no_dynamic_require

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_dynamic_require.schema.json
var schemaJSON []byte

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-dynamic-require.js
var NoDynamicRequireRule = rule.Rule{
	Name:   "import/no-dynamic-require",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		esmodule := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				esmodule, _ = option["esmodule"].(bool)
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil {
					return
				}
				var message string
				switch {
				case callee.Kind == ast.KindIdentifier && callee.Text() == "require":
					message = "Calls to require() should use string literals"
				case esmodule && callee.Kind == ast.KindImportKeyword:
					message = "Calls to import() should use string literals"
				default:
					return
				}

				argument := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
				// Upstream accepts every ESTree Literal, not just strings.
				if utils.IsESTreeLiteralKind(argument.Kind) || argument.Kind == ast.KindNoSubstitutionTemplateLiteral {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{Description: message})
			},
		}
	},
}
