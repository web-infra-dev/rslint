package no_self_import

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/import/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// See: https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-self-import.js
var NoSelfImportRule = rule.Rule{
	Name: "import/no-self-import",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return utils.VisitModules(func(source, node *ast.Node) {
			isImportingSelf(ctx, source, node)
		}, utils.VisitModulesOptions{
			Commonjs: true,
			ESModule: true,
		})
	},
}

// https://github.com/import-js/eslint-plugin-import/blob/01c9eb04331d2efa8d63f2d7f4bfec3bc44c94f3/src/rules/no-self-import.js#L12-L22
func isImportingSelf(ctx rule.RuleContext, source *ast.StringLiteralLike, node *ast.ImportSpecifierNode) {
	filePath := utils.GetPhysicalFilename(ctx)

	if resolvedPath, ok := utils.Resolve(source, ctx); ok {
		if /** filePath != "<text>" && */ filePath == resolvedPath {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "import/no-self-import",
				Description: "Module imports itself.",
			})
		}
	}
}
