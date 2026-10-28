package scope

import (
	"slices"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

type managerKey struct {
	collectReferences bool
	filtered          bool
	names             string
}

// Cache shares read-only scope graphs for one source file and lint pass.
// Owners must serialize access, as native rules do for each file.
// Reference options never change the complete declaration tree.
type Cache struct {
	sourceFile         *ast.SourceFile
	managers           map[managerKey]*Manager
	declarations       *Manager
	completeReferences *Manager
}

// NewCache creates a lazy scope cache. It does not build a scope tree.
func NewCache(sourceFile *ast.SourceFile) *Cache {
	return &Cache{sourceFile: sourceFile}
}

// Get returns the graph for the exact options, without widening filtered
// reference requests. The caller must treat graphs as read-only.
func (cache *Cache) Get(options Options) *Manager {
	key := keyFor(options)
	if manager := cache.managers[key]; manager != nil {
		return manager
	}
	manager := Build(cache.sourceFile, options)
	if cache.managers == nil {
		cache.managers = make(map[managerKey]*Manager)
	}
	cache.managers[key] = manager
	if options.CollectReferences && options.ReferenceNames == nil {
		cache.completeReferences = manager
		cache.declarations = manager
	} else if cache.declarations == nil {
		cache.declarations = manager
	}
	return manager
}

// References returns a graph containing at least the requested names.
func (cache *Cache) References(names map[string]struct{}) *Manager {
	if cache.completeReferences != nil {
		return cache.completeReferences
	}
	return cache.Get(Options{CollectReferences: true, ReferenceNames: names})
}

// Declarations returns a graph whose declaration tree is complete. Its
// reference slices may be absent or filtered and must not be inspected.
func (cache *Cache) Declarations() *Manager {
	if cache.declarations != nil {
		return cache.declarations
	}
	return cache.Get(Options{})
}

// PeekDeclarations returns an existing declaration-capable graph, or nil.
// Unlike Declarations it never constructs a graph.
func (cache *Cache) PeekDeclarations() *Manager { return cache.declarations }

func keyFor(options Options) managerKey {
	key := managerKey{collectReferences: options.CollectReferences}
	if !options.CollectReferences || options.ReferenceNames == nil {
		return key
	}
	key.filtered = true
	names := make([]string, 0, len(options.ReferenceNames))
	for name := range options.ReferenceNames {
		names = append(names, name)
	}
	slices.Sort(names)
	var encoded strings.Builder
	for _, name := range names {
		encoded.WriteString(strconv.Quote(name))
	}
	key.names = encoded.String()
	return key
}
