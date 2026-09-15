package program

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The key identifies syntax-only data stored by ts-go on each SourceFile. It
// carries no files or values itself, so an entry becomes unreachable with the
// exact AST it describes instead of requiring a Server or Program lifecycle
// event to evict it.
var moduleSpecifierCacheKey = ast.NewSourceFileDataKey[*sourceFileModuleSpecifierCache]()

// sourceFileModuleSpecifierCache separates the syntax combinations requested
// for one immutable SourceFile. Its values may contain only scalar data and
// nodes owned by that same SourceFile: Program state, resolved targets, checker
// state, and other SourceFiles belong to shorter or independent lifetimes.
type sourceFileModuleSpecifierCache struct {
	mu      sync.Mutex
	byKinds map[ModuleReferenceKinds][]moduleSpecifier
}

func newSourceFileModuleSpecifierCache(*ast.SourceFile) *sourceFileModuleSpecifierCache {
	return &sourceFileModuleSpecifierCache{}
}

// cachedModuleSpecifiers returns what file writes in these syntaxes. Programs
// that reuse this exact SourceFile share the collection; distinct SourceFile
// objects, including ones with the same path, own independent answers.
func cachedModuleSpecifiers(file *ast.SourceFile, kinds ModuleReferenceKinds) []moduleSpecifier {
	cache := ast.GetOrComputeSourceFileData(file, moduleSpecifierCacheKey, newSourceFileModuleSpecifierCache)
	cache.mu.Lock()
	specifiers, ok := cache.byKinds[kinds]
	cache.mu.Unlock()
	if ok {
		return specifiers
	}

	// Collection is pure and intentionally happens outside the lock: unrelated
	// files and syntax sets never serialize, and a panic publishes no partial
	// value. A same-key race may duplicate one collection; the first completed
	// immutable result wins.
	specifiers = collectSpecifiers(file, kinds)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if existing, ok := cache.byKinds[kinds]; ok {
		return existing
	}
	if cache.byKinds == nil {
		cache.byKinds = make(map[ModuleReferenceKinds][]moduleSpecifier)
	}
	cache.byKinds[kinds] = specifiers
	return specifiers
}
