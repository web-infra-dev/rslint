package modules

import (
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// The key identifies syntax-only data stored by ts-go on each SourceFile. It
// carries no files or values itself, so an entry becomes unreachable with the
// exact AST it describes instead of requiring a Server or Program lifecycle
// event to evict it.
var sourceCacheKey = ast.NewSourceFileDataKey[*sourceFileModuleSpecifierCache]()

// sourceFileModuleSpecifierCache separates the syntax combinations requested
// for one immutable SourceFile. Its values may contain only scalar data and
// nodes owned by that same SourceFile: Program state, resolved targets, checker
// state, and other SourceFiles belong to shorter or independent lifetimes.
type sourceFileModuleSpecifierCache struct {
	mu      sync.Mutex
	byKinds map[ReferenceKinds][]Source
}

func newSourceFileModuleSpecifierCache(*ast.SourceFile) *sourceFileModuleSpecifierCache {
	return &sourceFileModuleSpecifierCache{}
}

// Collect returns module source expressions in source order. The returned
// slice and nodes are read-only. It performs no resolution. Programs
// that reuse this exact SourceFile share the collection; distinct SourceFile
// objects, including ones with the same path, own independent answers.
func Collect(file *ast.SourceFile, kinds ReferenceKinds) []Source {
	if file == nil || kinds == 0 {
		return nil
	}
	cache := ast.GetOrComputeSourceFileData(file, sourceCacheKey, newSourceFileModuleSpecifierCache)
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
		cache.byKinds = make(map[ReferenceKinds][]Source)
	}
	cache.byKinds[kinds] = specifiers
	return specifiers
}
