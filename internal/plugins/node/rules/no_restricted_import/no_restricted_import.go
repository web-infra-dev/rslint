package no_restricted_import

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_restricted_import.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-restricted-import.js
var NoRestrictedImportRule = rule.Rule{
	Name:   "node/no-restricted-import",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		restrictions := nodeutil.ParseRestrictions(options)
		if len(restrictions) == 0 {
			return nil
		}
		var resolutionOptions [2]*nodeutil.ResolutionOptions
		return nodeutil.VisitImports(nodeutil.ImportVisitorOptions{IncludeCore: true}, func(source *ast.Node, name string, typeOnly bool) {
			message := restrictions.Match(name, func() string {
				if ctx.Program() == nil {
					return ""
				}
				index := 0
				if typeOnly {
					index = 1
				}
				if resolutionOptions[index] == nil {
					resolution := nodeutil.ImportResolutionOptions(ctx, typeOnly, nil)
					resolutionOptions[index] = &resolution
				}
				return nodeutil.ImportFilePath(ctx.Program(), name, ctx.SourceFile.FileName(), typeOnly, *resolutionOptions[index])
			})
			if message != nil {
				ctx.ReportNode(source, *message)
			}
		})
	},
}
