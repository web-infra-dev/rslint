package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// newPragmaImportMatcher mirrors isDestructuredFromPragmaImport's name lookup.
// Upstream searches the current scope, its first child, and that child's first
// child before moving to the parent. It reads the last definition regardless
// of whether that definition declares a value. Reference resolution alone
// therefore cannot answer this particular React question.
func newPragmaImportMatcher(ctx rule.RuleContext, pragma, name string) func(*ast.Node) bool {
	sourceFile := ctx.SourceFile
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	var pragmaLower string
	var definitions *VariableDefinitionLookup
	return func(ident *ast.Node) bool {
		if sourceFile == nil || ident == nil || ident.Kind != ast.KindIdentifier || ident.Text() != name {
			return false
		}
		if definitions == nil {
			pragmaLower = ecmascript.StringToLowerCase(pragma)
			definitions = NewVariableDefinitionLookup(ctx)
		}
		definition := definitions.Last(ident, name)
		return isDestructuredFromPragmaDeclaration(definition, pragma, pragmaLower)
	}
}
