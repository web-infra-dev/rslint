package no_amd

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-amd.js
var NoAmdRule = rule.Rule{
	Name:   "import/no-amd",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.LanguageOptions.EffectiveSourceType() != "module" {
			return nil
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier ||
					(callee.Text() != "require" && callee.Text() != "define") ||
					call.Arguments == nil || len(call.Arguments.Nodes) != 2 ||
					utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0]).Kind != ast.KindArrayLiteralExpression {
					return
				}
				// The scope model folds the ES module scope into Global. Check
				// the immediate scope: even a top-level block is excluded upstream.
				if scopeanalysis.Declarations(ctx).Acquire(node).Kind != scope.KindGlobal {
					return
				}
				// Upstream uses literal messages without message IDs or edits.
				message := "Expected imports instead of AMD define()."
				if callee.Text() == "require" {
					message = "Expected imports instead of AMD require()."
				}
				ctx.ReportNode(node, rule.RuleMessage{Description: message})
			},
		}
	},
}
