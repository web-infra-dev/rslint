package no_self_import

import (
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// See: https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-self-import.js
var NoSelfImportRule = rule.Rule{
	Name:   "import/no-self-import",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return utils.VisitModules(func(source modules.Source) {
			isImportingSelf(ctx, source)
		}, utils.VisitModulesOptions{
			Commonjs: true,
			ESModule: true,
		})
	},
}

// https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-self-import.js#L12-L22
func isImportingSelf(ctx rule.RuleContext, source modules.Source) {
	if utils.HasDefaultOnlyModuleView(source) {
		return
	}
	filePath := ctx.SourceFile.FileName()

	if resolvedPath, _, ok := ctx.Program().ResolveModule(ctx.SourceFile, source.Specifier); ok {
		if /** filePath != "<text>" && */ filePath == resolvedPath {
			ctx.ReportNode(source.Declaration, rule.RuleMessage{
				Id:          "import/no-self-import",
				Description: "Module imports itself.",
			})
		}
	}
}
