package no_missing_require

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
)

//go:embed no_missing_require.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-missing-require.js
var NoMissingRequireRule = rule.Rule{
	Name:   "node/no-missing-require",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !nodeutil.MayHaveRequire(ctx.SourceFile) {
			return nil
		}
		p := ctx.Program()
		fileName := ctx.SourceFile.FileName()
		if p == nil || fileName == "<input>" {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
				var resolution *moduleresolver.Options
				resolveErrors := map[string]string{}
				for _, target := range nodeutil.CollectRequireTargets(ctx) {
					resolveError, checked := resolveErrors[target.Name]
					if !checked {
						if len(allowed) > 0 {
							name, _ := nodeutil.ImportModuleName(target.Name)
							if name != "" && slices.Contains(allowed, name) {
								resolveErrors[target.Name] = ""
								continue
							}
						}
						if resolution == nil {
							value := nodeutil.RequireResolutionOptions(ctx, opts)
							resolution = &value
						}
						resolveError = nodeutil.RequireResolveError(p, target.Name, fileName, *resolution)
						resolveErrors[target.Name] = resolveError
					}
					if resolveError != "" {
						ctx.ReportNode(target.Node, rule.RuleMessage{
							Id: "notFound", Description: resolveError,
							Data: map[string]string{"resolveError": resolveError},
						})
					}
				}
			},
		}
	},
}
