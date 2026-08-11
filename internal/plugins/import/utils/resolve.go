package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
)

// Resolve names the file a specifier in the linted file points at. Resolution
// itself is a property of the effective source runtime rather than of these
// rules, and lives in utils; this is the spelling the import rules read most
// naturally.
func Resolve(moduleSpecifier *ast.StringLiteralLike, ctx rule.RuleContext) (string, bool) {
	if !ctx.HasSourceRuntime() {
		return "", false
	}
	return rslint_utils.ResolveModulePath(ctx.ModuleResolutionRuntime(), ctx.SourceFile, moduleSpecifier)
}
