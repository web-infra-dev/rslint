package utils

import (
	"os"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/zeebo/xxh3"
)

// ParseCache owns two run/request-scoped cache layers shared across Programs:
// a source snapshot generation keyed by the compiler host's exact normalized
// file name, and an append-mostly parsed SourceFile cache keyed by every parse
// input. Multiple tsconfigs can therefore reuse both the source text/hash and
// the resulting *ast.SourceFile object.
//
// The AST key reuses the upstream project.ParseCacheKey — SourceFileParseOptions
// (FileName + Path + ExternalModuleIndicatorOptions) + ScriptKind + an xxh3
// hash of the file content. This is exactly the condition under which tsgo's
// own LSP shares SourceFile objects across projects (see
// typescript-go/internal/project/parsecache.go), so a cache hit is by
// construction equivalent to an independent parse of the same inputs.
//
// Sharing is an optimization, never a correctness requirement: two Programs
// holding two distinct objects for the same file is the no-cache baseline.
// That is what makes RetainOnly's deletion-only eviction safe (see I8 below).
//
// Invariants (enforced here, relied on by callers):
//
//	I1: within one source generation, the first successful read of an exact
//	    FileName is authoritative. Lexical/canonical/symlink/case aliases are
//	    never merged. Failed reads are not stored.
//	I2: a source snapshot publishes immutable text and its xxh3 hash as one
//	    value. The exact same text/hash pair is used to key and build the AST.
//	I3: file.Hash is never written. The --api EncodeAST header encodes
//	    sourceFile.Hash (bytes 4-19) and is all-zero today; writing it
//	    would change encoded output bytes.
//	I4: concurrent source and AST misses resolve via LoadOrStore. Every loser
//	    uses the published winner; duplicate cold reads/parses may occur, but
//	    callers within a generation observe one winning snapshot/AST per key.
//	I5: the source layer owned by one cache binds to one exact pointer-identity
//	    FS across its generations. Hosts backed by another or a non-pointer FS
//	    bypass only the source layer and still use the content-keyed AST cache.
//	I6: ScriptKind is derived exactly like the default compiler host.
//	I7: source invalidation atomically swaps in a fresh generation. It never
//	    clears a live map, so a pre-invalidation miss can only publish into its
//	    captured old generation. Invalidation is not a reader barrier.
//	I8: AST eviction is deletion-only (RetainOnly). A deleted entry is simply
//	    re-parsed on next request; selective rewriting of entries is
//	    forbidden — deletion can cost a parse, never a stale result.
//	I9: cache ownership is explicit and bounded to one CLI run or API request;
//	    no package-level singleton may extend source/AST lifetime implicitly.
type ParseCache struct {
	m sync.Map // project.ParseCacheKey -> *ast.SourceFile

	sourceGeneration atomic.Pointer[sourceSnapshotGeneration]
	sourceFSMu       sync.Mutex
	sourceFS         vfs.FS
}

type sourceSnapshotGeneration struct {
	entries sync.Map // exact opts.FileName -> sourceSnapshot
}

type sourceSnapshot struct {
	text string
	hash xxh3.Uint128
}

// NewParseCache returns a fresh cache, or nil (disabling caching entirely via
// the WithParseCache nil-passthrough) when RSLINT_DISABLE_PARSE_CACHE is set.
// The escape hatch is undocumented — it exists for bisecting user issues and
// A/B verification, not as a user-facing knob.
func NewParseCache() *ParseCache {
	if os.Getenv("RSLINT_DISABLE_PARSE_CACHE") != "" {
		return nil
	}
	cache := &ParseCache{}
	cache.sourceGeneration.Store(&sourceSnapshotGeneration{})
	return cache
}

// acquire returns the shared SourceFile for (opts, text), parsing on miss.
func (c *ParseCache) acquire(opts ast.SourceFileParseOptions, text string) *ast.SourceFile {
	return c.acquireSnapshot(opts, sourceSnapshot{
		text: text,
		hash: xxh3.HashString128(text),
	})
}

// acquireSnapshot returns the shared SourceFile for (opts, snapshot), parsing
// on miss. snapshot.hash must be the hash of snapshot.text; keeping them in one
// immutable value prevents callers from accidentally tearing the pair.
func (c *ParseCache) acquireSnapshot(opts ast.SourceFileParseOptions, snapshot sourceSnapshot) *ast.SourceFile {
	// I6: derive ScriptKind exactly like the default host
	// (typescript-go/internal/compiler/host.go) so the key always matches
	// what the parse below actually uses.
	scriptKind := core.GetScriptKindFromFileName(opts.FileName)
	key := project.NewParseCacheKey(opts, snapshot.hash, scriptKind)
	if v, ok := c.m.Load(key); ok {
		if cached, ok := v.(*ast.SourceFile); ok {
			return cached
		}
	}
	sf := parser.ParseSourceFile(opts, snapshot.text, scriptKind) // I2/I3: same text, Hash left zero
	actual, _ := c.m.LoadOrStore(key, sf)                         // I4: winner is the only object handed out
	if winner, ok := actual.(*ast.SourceFile); ok {
		return winner
	}
	return sf // unreachable: the map only ever holds *ast.SourceFile
}

// RetainOnly evicts every entry whose SourceFile is not referenced by any of
// the given programs (live set = union of all GetSourceFiles() pointers,
// including gap-fallback programs). One sweep reclaims both --fix
// intermediate versions (old-hash objects absent from rebuilt programs) and
// build-time dedup losers (parsed but never included in a program).
// Deletion-only per I8: an evicted entry that is requested again is re-parsed
// into a fresh object, which is the no-cache baseline behavior. Source
// snapshots have a separate generation lifetime and are not pruned here.
func (c *ParseCache) RetainOnly(programs []*compiler.Program) {
	if c == nil {
		return
	}
	live := make(map[*ast.SourceFile]struct{})
	for _, p := range programs {
		if p == nil {
			continue
		}
		for _, sf := range p.GetSourceFiles() {
			live[sf] = struct{}{}
		}
	}
	c.m.Range(func(k, v any) bool {
		if _, ok := live[v.(*ast.SourceFile)]; !ok {
			c.m.Delete(k)
		}
		return true
	})
}

// InvalidateSourceSnapshots atomically installs an empty source generation.
// Lookups that already captured the previous generation may finish against it;
// subsequent lookups use the new generation. The AST cache is intentionally
// retained because unchanged content can still reuse its parsed SourceFile.
func (c *ParseCache) InvalidateSourceSnapshots() {
	if c == nil {
		return
	}
	c.sourceGeneration.Store(&sourceSnapshotGeneration{})
}

func (c *ParseCache) currentSourceGeneration() *sourceSnapshotGeneration {
	for {
		if generation := c.sourceGeneration.Load(); generation != nil {
			return generation
		}
		generation := &sourceSnapshotGeneration{}
		if c.sourceGeneration.CompareAndSwap(nil, generation) {
			return generation
		}
	}
}

// bindSourceSnapshotFS binds the source layer owned by this cache to one exact
// FS instance across all generations. Only pointer-backed implementations have
// unambiguous instance identity; all other implementations conservatively
// bypass the source layer. This check runs once per compiler-host wrapper,
// never on the GetSourceFile hot path.
func (c *ParseCache) bindSourceSnapshotFS(fs vfs.FS) bool {
	c.sourceFSMu.Lock()
	defer c.sourceFSMu.Unlock()

	if fs == nil || reflect.TypeOf(fs).Kind() != reflect.Pointer {
		return false
	}
	if c.sourceFS == nil {
		c.sourceFS = fs
		return true
	}
	return c.sourceFS == fs
}

// cachingCompilerHost overrides only GetSourceFile; the remaining
// CompilerHost methods (FS, DefaultLibraryPath, GetCurrentDirectory, Trace,
// GetResolvedProjectReference) are delegated through interface embedding.
type cachingCompilerHost struct {
	compiler.CompilerHost
	cache              *ParseCache
	useSourceSnapshots bool
}

// WithParseCache wraps host with the shared parse cache. A nil cache returns
// the host unchanged, so callers need no branches. The cache must be created
// at the pipeline entry and passed down explicitly (I9) — never stored in a
// package-level singleton, which would silently leak sharing into unrelated
// paths (rule_tester, --api getAstInfo) and grow without bound in
// long-running processes.
func WithParseCache(host compiler.CompilerHost, cache *ParseCache) compiler.CompilerHost {
	if cache == nil {
		return host
	}
	return &cachingCompilerHost{
		CompilerHost:       host,
		cache:              cache,
		useSourceSnapshots: cache.bindSourceSnapshotFS(host.FS()),
	}
}

func (h *cachingCompilerHost) GetSourceFile(opts ast.SourceFileParseOptions) *ast.SourceFile {
	if h.useSourceSnapshots {
		generation := h.cache.currentSourceGeneration() // I7: capture exactly one generation
		if value, ok := generation.entries.Load(opts.FileName); ok {
			if snapshot, ok := value.(sourceSnapshot); ok {
				return h.cache.acquireSnapshot(opts, snapshot)
			}
		}

		text, ok := h.FS().ReadFile(opts.FileName)
		if !ok {
			return nil // I1: failed reads are never published
		}
		candidate := sourceSnapshot{
			text: text,
			hash: xxh3.HashString128(text),
		}
		actual, _ := generation.entries.LoadOrStore(opts.FileName, candidate)
		snapshot, ok := actual.(sourceSnapshot)
		if !ok {
			snapshot = candidate // unreachable: entries only ever stores sourceSnapshot values
		}
		return h.cache.acquireSnapshot(opts, snapshot) // I4: always use the winner
	}

	text, ok := h.FS().ReadFile(opts.FileName)
	if !ok {
		return nil // same as the default host, and never cached
	}
	return h.cache.acquire(opts, text)
}
