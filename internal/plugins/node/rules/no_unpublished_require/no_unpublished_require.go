package no_unpublished_require

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_unpublished_require.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-require.js
var NoUnpublishedRequireRule = rule.Rule{
	Name:   "node/no-unpublished-require",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !nodeutil.MayHaveRequire(ctx.SourceFile) {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		publication := nodeutil.NewPublicationCheck(ctx, opts)
		if publication == nil {
			return nil
		}
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				var resolution *moduleresolver.Options
				messages := map[string]*rule.RuleMessage{}
				for _, target := range nodeutil.CollectRequireTargets(ctx) {
					if modules.IsNodeBuiltin(target.Name) {
						continue
					}
					message, checked := messages[target.Name]
					if !checked {
						name, resource := nodeutil.ImportModuleName(target.Name)
						var ignored bool
						if name != "" {
							ignored = publication.IsUnpublishedDependency(name)
						} else {
							if resolution == nil {
								value := nodeutil.RequireResolutionOptions(ctx, opts)
								resolution = &value
							}
							filePath := nodeutil.RequireFilePath(ctx.Program(), resource, ctx.SourceFile.FileName(), *resolution)
							ignored = publication.IsUnpublishedFile(filePath)
						}
						if ignored {
							if name == "" {
								name = target.Name
							}
							message = &rule.RuleMessage{
								Id: "notPublished", Description: `"` + name + `" is not published.`,
								Data: map[string]string{"name": name},
							}
						}
						messages[target.Name] = message
					}
					if message != nil {
						ctx.ReportNode(target.Node, *message)
					}
				}
			},
		}
	},
}
