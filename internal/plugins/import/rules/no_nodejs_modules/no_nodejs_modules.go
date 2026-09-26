package no_nodejs_modules

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	import_utils "github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_nodejs_modules.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-nodejs-modules.js
var NoNodejsModulesRule = rule.Rule{
	Name:   "import/no-nodejs-modules",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var allowed []any
		if len(options) > 0 {
			if opts, ok := options[0].(map[string]any); ok {
				allowed, _ = opts["allow"].([]any)
			}
		}
		var settings *import_utils.ModuleSettings
		var resolver *import_utils.ImportResolver
		reportedResolverError := false
		return import_utils.VisitModules(func(source *ast.StringLiteralLike, node *ast.Node) {
			name := source.Text()
			if slices.Contains(allowed, any(name)) {
				return
			}
			if resolver == nil {
				settings = import_utils.SettingsFor(ctx)
				resolver = import_utils.NewImportResolver(ctx)
			}
			path, _, resolveError := resolver.Resolve(source)
			if resolveError != "" && !reportedResolverError {
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{Description: "Resolve error: " + resolveError})
				reportedResolverError = true
			}
			if settings.IsBuiltinSpecifier(name, path) {
				// Upstream reports the entire declaration/call, with no message ID or edits.
				ctx.ReportNode(node, rule.RuleMessage{Description: `Do not import Node.js builtin module "` + name + `"`})
			}
		}, import_utils.VisitModulesOptions{ESModule: true, Commonjs: true})
	},
}
