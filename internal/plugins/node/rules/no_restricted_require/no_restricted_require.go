package no_restricted_require

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
)

//go:embed no_restricted_require.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-require.js
var NoRestrictedRequireRule = rule.Rule{
	Name:   "node/no-restricted-require",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if len(options) == 0 || !nodeutil.MayHaveRequire(ctx.SourceFile) {
			return nil
		}
		restrictions := nodeutil.ParseRestrictions(options)
		if len(restrictions) == 0 {
			return nil
		}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				var resolution *moduleresolver.Options
				for _, target := range nodeutil.CollectRequireTargets(ctx) {
					message := restrictions.Match(target.Name, func() string {
						if ctx.Program() == nil {
							return ""
						}
						if resolution == nil {
							options := nodeutil.RequireResolutionOptions(ctx, nil)
							resolution = &options
						}
						return nodeutil.RequireFilePath(ctx.Program(), target.Name, ctx.SourceFile.FileName(), *resolution)
					})
					if message != nil {
						ctx.ReportNode(target.Node, *message)
					}
				}
			},
		}
	},
}
