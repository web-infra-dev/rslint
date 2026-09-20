package no_missing_import

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
)

//go:embed no_missing_import.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-missing-import.js
var NoMissingImportRule = rule.Rule{
	Name:   "node/no-missing-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		fileName := ctx.SourceFile.FileName()
		if fileName == "<input>" {
			return nil
		}
		p := ctx.Program()
		if p == nil {
			return nil
		}
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		ignoreTypeImport, _ := opts["ignoreTypeImport"].(bool)
		allowed := nodeutil.StringListSetting("allowModules", opts, ctx.Settings)
		var resolutionOptions [2]*moduleresolver.Options
		var resolveErrors [2]map[string]string
		return nodeutil.VisitImports(ctx, nodeutil.ImportVisitorOptions{IgnoreTypeImport: ignoreTypeImport}, func(source *ast.Node, specifier string, typeOnly bool) {
			index := 0
			if typeOnly {
				index = 1
			}
			resolveError, checked := resolveErrors[index][specifier]
			if !checked {
				name, resource := nodeutil.ImportModuleName(specifier)
				if name != "" && slices.Contains(allowed, name) {
					return
				}
				if resolutionOptions[index] == nil {
					resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, opts)
					resolutionOptions[index] = &resolution
					resolveErrors[index] = map[string]string{}
				}
				resolveError = nodeutil.ImportResolveError(p, resource, fileName, typeOnly, *resolutionOptions[index])
				resolveErrors[index][specifier] = resolveError
			}
			if resolveError != "" {
				ctx.ReportNode(source, rule.RuleMessage{
					Id: "notFound", Description: resolveError,
					Data: map[string]string{"resolveError": resolveError},
				})
			}
		})
	},
}
