package rule

import "github.com/web-infra-dev/rslint/internal/utils/scope"

type scopeAnalysisKey struct{}

// ScopeAnalysis returns this file pass's shared, lazy scope analysis cache.
// Native rules normally access it through the scopeanalysis facade. Keeping
// ownership here lets framework capabilities share the same graphs without
// making the scope package depend on RuleContext or creating an import cycle.
// A manually assembled context without FileCache receives a fresh cache.
func (ctx *RuleContext) ScopeAnalysis() *scope.Cache {
	sourceFile := ctx.SourceFile
	return CachedByFile(*ctx, scopeAnalysisKey{}, func() *scope.Cache {
		return scope.NewCache(sourceFile)
	})
}
