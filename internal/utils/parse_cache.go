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
//	I4: concurrent source and AST misses are single-flight. Every caller within
//	    a generation observes one winning snapshot/AST per key, without
//	    duplicating successful cold reads, content hashes, or parses. Failed
//	    reads are retried, and a panic never leaves waiters blocked.
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
	m        sync.Map // project.ParseCacheKey -> *ast.SourceFile
	inflight sync.Map // project.ParseCacheKey -> *parseCacheFlight

	sourceGeneration atomic.Pointer[sourceSnapshotGeneration]
	sourceFSMu       sync.Mutex
	sourceFS         vfs.FS
}

type sourceSnapshotGeneration struct {
	entries  sync.Map // exact opts.FileName -> sourceSnapshot
	inflight sync.Map // exact opts.FileName -> *sourceSnapshotFlight
}

type sourceSnapshot struct {
	text string
	hash xxh3.Uint128
}

type sourceSnapshotFlight struct {
	ready     chan struct{}
	snapshot  sourceSnapshot
	ok        bool
	completed bool
}

type parseCacheFlight struct {
	ready      chan struct{}
	sourceFile *ast.SourceFile
	completed  bool
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
	if value, ok := c.m.Load(key); ok {
		if sourceFile, ok := value.(*ast.SourceFile); ok {
			return sourceFile
		}
	}
	return c.acquireParseMiss(key, func() *ast.SourceFile {
		return parser.ParseSourceFile(opts, snapshot.text, scriptKind) // I2/I3: same text, Hash left zero
	})
}

func (c *ParseCache) acquireParseMiss(
	key project.ParseCacheKey,
	parse func() *ast.SourceFile,
) *ast.SourceFile {
	for {
		candidate := &parseCacheFlight{ready: make(chan struct{})}
		actual, loaded := c.inflight.LoadOrStore(key, candidate)
		if loaded {
			flight := mustParseCacheFlight(actual)
			<-flight.ready
			if flight.completed {
				return flight.sourceFile
			}
			continue
		}

		// A previous owner can publish the final entry and remove its flight
		// between acquireSnapshot's fast-path miss and our LoadOrStore above.
		if value, ok := c.m.Load(key); ok {
			if sourceFile, ok := value.(*ast.SourceFile); ok {
				candidate.sourceFile = sourceFile
				candidate.completed = true
				c.completeParseFlight(key, candidate)
				return sourceFile
			}
		}
		return c.parseAndPublish(key, candidate, parse)
	}
}

func (c *ParseCache) parseAndPublish(
	key project.ParseCacheKey,
	flight *parseCacheFlight,
	parse func() *ast.SourceFile,
) *ast.SourceFile {
	completed := false
	defer func() {
		if !completed {
			// Preserve the panic while allowing a waiting caller to retry.
			c.completeParseFlight(key, flight)
		}
	}()

	sourceFile := parse()
	if sourceFile != nil {
		actual, _ := c.m.LoadOrStore(key, sourceFile)
		if winner, ok := actual.(*ast.SourceFile); ok {
			sourceFile = winner
		}
	}
	flight.sourceFile = sourceFile
	flight.completed = true
	completed = true
	c.completeParseFlight(key, flight)
	return sourceFile
}

func (c *ParseCache) completeParseFlight(key project.ParseCacheKey, flight *parseCacheFlight) {
	c.inflight.CompareAndDelete(key, flight)
	close(flight.ready)
}

func mustParseCacheFlight(value any) *parseCacheFlight {
	flight, ok := value.(*parseCacheFlight)
	if !ok {
		panic("parse cache in-flight map contains an unexpected value")
	}
	return flight
}

// RetainOnly evicts every entry whose SourceFile is not referenced by any of
// the given programs (live set = union of all GetSourceFiles() pointers,
// including gap-fallback programs). One sweep reclaims both --fix
// intermediate versions (old-hash objects absent from rebuilt programs).
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
		sourceFile, ok := v.(*ast.SourceFile)
		if !ok {
			return true
		}
		if _, ok := live[sourceFile]; !ok {
			c.m.CompareAndDelete(k, v)
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
		snapshot, ok := h.acquireSourceSnapshot(generation, opts.FileName)
		if !ok {
			return nil // I1: failed reads are never published
		}
		return h.cache.acquireSnapshot(opts, snapshot)
	}

	text, ok := h.FS().ReadFile(opts.FileName)
	if !ok {
		return nil // same as the default host, and never cached
	}
	return h.cache.acquire(opts, text)
}

func (h *cachingCompilerHost) acquireSourceSnapshot(
	generation *sourceSnapshotGeneration,
	fileName string,
) (sourceSnapshot, bool) {
	for {
		if value, ok := generation.entries.Load(fileName); ok {
			if snapshot, ok := value.(sourceSnapshot); ok {
				return snapshot, true
			}
		}

		candidate := &sourceSnapshotFlight{ready: make(chan struct{})}
		actual, loaded := generation.inflight.LoadOrStore(fileName, candidate)
		if loaded {
			flight := mustSourceSnapshotFlight(actual)
			<-flight.ready
			if flight.completed {
				return flight.snapshot, flight.ok
			}
			// The owner panicked. Retry rather than leaving this caller blocked.
			continue
		}

		// As with the AST layer, close the small race between the initial
		// cache lookup and becoming the single-flight owner.
		if value, ok := generation.entries.Load(fileName); ok {
			if snapshot, ok := value.(sourceSnapshot); ok {
				candidate.snapshot = snapshot
				candidate.ok = true
				candidate.completed = true
				h.completeSourceSnapshotFlight(generation, fileName, candidate)
				return snapshot, true
			}
		}
		return h.readAndPublishSourceSnapshot(generation, fileName, candidate)
	}
}

func (h *cachingCompilerHost) readAndPublishSourceSnapshot(
	generation *sourceSnapshotGeneration,
	fileName string,
	flight *sourceSnapshotFlight,
) (sourceSnapshot, bool) {
	completed := false
	defer func() {
		if !completed {
			// A panicking VFS must not strand Programs waiting on the same file.
			h.completeSourceSnapshotFlight(generation, fileName, flight)
		}
	}()

	text, ok := h.FS().ReadFile(fileName)
	if !ok {
		flight.completed = true
		completed = true
		h.completeSourceSnapshotFlight(generation, fileName, flight)
		return sourceSnapshot{}, false
	}

	snapshot := sourceSnapshot{text: text, hash: xxh3.HashString128(text)}
	actual, _ := generation.entries.LoadOrStore(fileName, snapshot)
	flight.snapshot = snapshot
	if winner, ok := actual.(sourceSnapshot); ok {
		flight.snapshot = winner
	}
	flight.ok = true
	flight.completed = true
	completed = true
	h.completeSourceSnapshotFlight(generation, fileName, flight)
	return flight.snapshot, true
}

func (h *cachingCompilerHost) completeSourceSnapshotFlight(
	generation *sourceSnapshotGeneration,
	fileName string,
	flight *sourceSnapshotFlight,
) {
	generation.inflight.CompareAndDelete(fileName, flight)
	close(flight.ready)
}

func mustSourceSnapshotFlight(value any) *sourceSnapshotFlight {
	flight, ok := value.(*sourceSnapshotFlight)
	if !ok {
		panic("source snapshot in-flight map contains an unexpected value")
	}
	return flight
}
