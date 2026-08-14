package rule

import (
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/tspath"
)

// ModuleSpecifierCache carries what a file writes about the modules it
// references from one lint run to the next. A run builds its own ModuleGraph
// over its own Program, and an editor produces a new Program on every
// keystroke, so without this every run re-reads the module specifiers of every
// file in the Program to answer for the one file it lints.
//
// Only the syntactic half is kept. What a specifier resolves to is the
// Program's answer, not the file's, and two Programs holding this same file
// can disagree about it — a file that appeared on disk in between changes what
// a relative specifier names without changing a byte of any file that writes
// it. Resolution therefore runs again for every run, and nothing here needs
// invalidating: an entry is the answer for exactly as long as the file object
// it was read from is the file, which pointer identity decides.
//
// The cache holds one file per path. An editor replaces a file object when its
// text changes, so an entry for a path is superseded rather than accumulated,
// and the cache stays the size of the project rather than the size of the
// session. Two Programs that hold distinct file objects for one path — which
// is what separate lint Programs built from separate hosts do — take turns
// instead of sharing, which costs the collection they would have shared and
// stays correct.
//
// Superseding is what bounds the cache, and only a later run over the same
// path supersedes anything. An owner that discards the Programs it was going
// to run against — the project reloaded, the config changed — leaves entries
// for files no later run will name, holding trees no Program holds any more,
// and says so with Reset.
//
// Rules never construct one: the linter takes it from whoever owns the
// sequence of runs, and a caller that owns no such sequence passes none.
type ModuleSpecifierCache struct {
	mu      sync.Mutex
	entries map[tspath.Path]*specifierCacheEntry
}

// specifierCacheEntry holds one path's answers. A file is asked about in as
// many syntaxes as the run's rules disagree on, which in practice is one or
// two, so the syntaxes are scanned rather than hashed.
type specifierCacheEntry struct {
	file      *ast.SourceFile
	bySyntax  []ModuleSyntax
	collected [][]moduleSpecifier
}

func NewModuleSpecifierCache() *ModuleSpecifierCache {
	return &ModuleSpecifierCache{entries: make(map[tspath.Path]*specifierCacheEntry)}
}

// Reset drops every entry. An entry names the file object it was read from,
// and a file object owns the tree it was parsed from, so the cache holds one
// tree per path it has answered for. That is the Program's tree for as long as
// runs keep coming over the same project, and nobody's once the owner discards
// the Programs: Reset is how the owner says it did, at the price of collecting
// again whatever the next run asks about.
func (cache *ModuleSpecifierCache) Reset() {
	if cache == nil {
		return
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	clear(cache.entries)
}

// Len is how many paths the cache currently answers for.
func (cache *ModuleSpecifierCache) Len() int {
	if cache == nil {
		return 0
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return len(cache.entries)
}

// get returns what file writes in these syntaxes, collecting it on the first
// request and on every request that names a file object this path has not been
// seen holding.
func (cache *ModuleSpecifierCache) get(file *ast.SourceFile, syntax ModuleSyntax) []moduleSpecifier {
	path := file.Path()

	cache.mu.Lock()
	if collected, ok := cache.entries[path].lookup(file, syntax); ok {
		cache.mu.Unlock()
		return collected
	}
	cache.mu.Unlock()

	// Collected outside the lock: the result depends on nothing but the file,
	// so two runs racing on one key cost a redundant collection at worst,
	// which is cheaper than serializing every file of a run behind one mutex.
	collected := collectSpecifiers(file, syntax)

	cache.mu.Lock()
	defer cache.mu.Unlock()
	entry := cache.entries[path]
	if entry == nil || entry.file != file {
		entry = &specifierCacheEntry{file: file}
		cache.entries[path] = entry
	}
	if stored, ok := entry.lookup(file, syntax); ok {
		return stored
	}
	entry.bySyntax = append(entry.bySyntax, syntax)
	entry.collected = append(entry.collected, collected)
	return collected
}

// lookup returns what this entry holds for file in these syntaxes. An entry
// read from a different file object answers for nothing: an editor replaces
// the object whenever the text it was parsed from changes.
func (entry *specifierCacheEntry) lookup(file *ast.SourceFile, syntax ModuleSyntax) ([]moduleSpecifier, bool) {
	if entry == nil || entry.file != file {
		return nil, false
	}
	for i, seen := range entry.bySyntax {
		if seen == syntax {
			return entry.collected[i], true
		}
	}
	return nil, false
}
