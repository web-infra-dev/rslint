package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func Resolve(moduleSpecifier *ast.StringLiteralLike, ctx rule.RuleContext) string {
	module := ctx.Program.GetResolvedModuleFromModuleSpecifier(ctx.SourceFile, moduleSpecifier)
	return module.ResolvedFileName
}
