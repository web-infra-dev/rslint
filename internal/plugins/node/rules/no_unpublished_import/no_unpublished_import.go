package no_unpublished_import

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed no_unpublished_import.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unpublished-import.js
var NoUnpublishedImportRule = rule.Rule{
	Name:   "node/no-unpublished-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		p, fileName := ctx.Program(), ctx.SourceFile.FileName()
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		if len(modules.Collect(ctx.SourceFile, modules.ESModuleReferences)) == 0 {
			return nil
		}
		publication := nodeutil.NewPublicationCheck(ctx, opts)
		if publication == nil {
			return nil
		}
		ignoreTypeImport, _ := opts["ignoreTypeImport"].(bool)
		var resolutionOptions [2]*moduleresolver.Options
		var unpublished [2]map[string]bool
		return nodeutil.VisitImports(ctx, nodeutil.ImportVisitorOptions{IgnoreTypeImport: ignoreTypeImport}, func(source *ast.Node, specifier string, typeOnly bool) {
			name, resource := nodeutil.ImportModuleName(specifier)
			if name != "" {
				if !publication.IsUnpublishedDependency(name) {
					return
				}
			} else {
				index := 0
				if typeOnly {
					index = 1
				}
				ignored, checked := unpublished[index][resource]
				if !checked {
					if resolutionOptions[index] == nil {
						resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, opts)
						resolutionOptions[index] = &resolution
						unpublished[index] = map[string]bool{}
					}
					target := nodeutil.ImportFilePath(p, resource, fileName, typeOnly, *resolutionOptions[index])
					ignored = publication.IsUnpublishedFile(target)
					// Options and package metadata stay fixed within this file.
					// Type imports can resolve differently; each occurrence still reports.
					unpublished[index][resource] = ignored
				}
				if !ignored {
					return
				}
				name = specifier
			}
			ctx.ReportNode(source, rule.RuleMessage{
				Id: "notPublished", Description: `"` + name + `" is not published.`,
				Data: map[string]string{"name": name},
			})
		})
	},
}
