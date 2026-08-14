package rule

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The key identifies syntax-only data stored by ts-go on each SourceFile. It
// carries no files or values itself, so an entry becomes unreachable with the
// exact AST it describes instead of requiring a Server or Program lifecycle
// event to evict it.
var moduleSpecifierCacheKey = ast.NewSourceFileDataKey[*sourceFileModuleSpecifierCache]()

// sourceFileModuleSpecifierCache separates the syntax combinations requested
// for one immutable SourceFile. Collection stays in LazyMap rather than the
// SourceFile data cell's sync.Once: a panicking collection is not published and
// can be retried, matching the run-local graph's existing behavior.
type sourceFileModuleSpecifierCache struct {
	bySyntax utils.LazyMap[ModuleSyntax, []moduleSpecifier]
}

func newSourceFileModuleSpecifierCache(*ast.SourceFile) *sourceFileModuleSpecifierCache {
	return &sourceFileModuleSpecifierCache{}
}

// cachedModuleSpecifiers returns what file writes in these syntaxes. Programs
// that reuse this exact SourceFile share the collection; distinct SourceFile
// objects, including ones with the same path, own independent answers.
func cachedModuleSpecifiers(file *ast.SourceFile, syntax ModuleSyntax) []moduleSpecifier {
	cache := ast.GetOrComputeSourceFileData(file, moduleSpecifierCacheKey, newSourceFileModuleSpecifierCache)
	return cache.bySyntax.Get(syntax, func() []moduleSpecifier {
		return collectSpecifiers(file, syntax)
	})
}
