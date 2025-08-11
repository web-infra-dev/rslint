package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func Resolve(moduleSpecifier *ast.StringLiteralLike, ctx rule.RuleContext) (string, bool) {
	module := ctx.Program.GetResolvedModuleFromModuleSpecifier(ctx.SourceFile, moduleSpecifier)

	if module != nil {
		return module.ResolvedFileName, true
	}

	return "", false
}
