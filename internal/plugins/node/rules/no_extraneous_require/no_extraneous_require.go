package no_extraneous_require

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_extraneous_require.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-extraneous-require.js
var NoExtraneousRequireRule = rule.Rule{
	Name:   "node/no-extraneous-require",
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
		pkg := nodeutil.FindPackage(p, fileName)
		if pkg == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
				var resolution *nodeutil.ResolutionOptions
				targets := map[string]string{}
				for _, target := range nodeutil.CollectRequireTargets(ctx) {
					specifier := target.Name
					name, found := targets[specifier]
					if !found {
						var resource string
						name, resource = nodeutil.ImportModuleName(specifier)
						if name != "" && (pkg.AllowsDependency(p, name) || slices.Contains(allowed, name) || nodeutil.HasTypeScriptAlias(p, fileName, resource)) {
							name = ""
						}
						if name != "" {
							if resolution == nil {
								value := nodeutil.RequireResolutionOptions(ctx, opts)
								resolution = &value
							}
							if nodeutil.ResolveModule(p, resource, fileName, *resolution) == "" {
								name = ""
							}
						}
						targets[specifier] = name
					}
					if name != "" {
						ctx.ReportNode(target.Node, rule.RuleMessage{
							Id: "extraneous", Description: `"` + name + `" is extraneous.`,
							Data: map[string]string{"moduleName": name},
						})
					}
				}
			},
		}
	},
}
