package utils

import (
	"os"
	"sync"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/zeebo/xxh3"
)

// ParseCache is a run-scoped, append-mostly cache of parsed SourceFiles shared
// across Programs. Multiple tsconfigs in one run re-parse the same lib /
// node_modules / shared-source files once per Program today; entries here let
// every Program whose parse inputs match reuse one *ast.SourceFile object.
//
// The key reuses the upstream project.ParseCacheKey — SourceFileParseOptions
// (FileName + Path + ExternalModuleIndicatorOptions) + ScriptKind + an xxh3
// hash of the file content. This is exactly the condition under which tsgo's
// own LSP shares SourceFile objects across projects (see
// typescript-go/internal/project/parsecache.go), so a cache hit is by
// construction equivalent to an independent parse of the same inputs.
//
// Sharing is an optimization, never a correctness requirement: two Programs
// holding two distinct objects for the same file is the no-cache baseline.
// That is what makes RetainOnly's deletion-only eviction safe (see I7 below).
//
// Invariants (enforced here, relied on by callers):
//
//	I1: the content hash is computed from the exact text handed to the
//	    parser, read once. This keeps the cache consistent with whatever
//	    the FS layer returns — even if the file changes between calls or a
//	    future vfs layer adds read caching, an entry's hash always matches
//	    its AST, so staleness can never exceed the no-cache baseline.
//	I2: file.Hash is never written. The --api EncodeAST header encodes
//	    sourceFile.Hash (bytes 4-19) and is all-zero today; writing it
//	    would change encoded output bytes.
//	I3: concurrent misses on one key resolve via LoadOrStore — the winner's
//	    object is returned to every caller, losers are discarded. Today
//	    program construction is serial and per-Program parsing dedups by
//	    path, so same-key concurrency cannot happen; this is defense for a
//	    future parallelization.
//	I4: failed reads return nil and are not cached (no negative entries).
//	I7: eviction is deletion-only (RetainOnly). A deleted entry is simply
//	    re-parsed on next request; selective rewriting of entries is
//	    forbidden — deletion can cost a parse, never a stale result.
type ParseCache struct {
	m sync.Map // project.ParseCacheKey -> *ast.SourceFile
}

// NewParseCache returns a fresh cache, or nil (disabling caching entirely via
// the WithParseCache nil-passthrough) when RSLINT_DISABLE_PARSE_CACHE is set.
// The escape hatch is undocumented — it exists for bisecting user issues and
// A/B verification, not as a user-facing knob.
func NewParseCache() *ParseCache {
	if os.Getenv("RSLINT_DISABLE_PARSE_CACHE") != "" {
		return nil
	}
	return &ParseCache{}
}

// acquire returns the shared SourceFile for (opts, text), parsing on miss.
func (c *ParseCache) acquire(opts ast.SourceFileParseOptions, text string) *ast.SourceFile {
	// I6: derive ScriptKind exactly like the default host
	// (typescript-go/internal/compiler/host.go) so the key always matches
	// what the parse below actually uses.
	scriptKind := core.GetScriptKindFromFileName(opts.FileName)
	key := project.NewParseCacheKey(opts, xxh3.HashString128(text), scriptKind)
	if v, ok := c.m.Load(key); ok {
		if cached, ok := v.(*ast.SourceFile); ok {
			return cached
		}
	}
	sf := parser.ParseSourceFile(opts, text, scriptKind) // I1/I2: same text, Hash left zero
	actual, _ := c.m.LoadOrStore(key, sf)                // I3: winner is the only object handed out
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
// Deletion-only per I7: an evicted entry that is requested again is re-parsed
// into a fresh object, which is the no-cache baseline behavior.
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

// cachingCompilerHost overrides only GetSourceFile; the remaining
// CompilerHost methods (FS, DefaultLibraryPath, GetCurrentDirectory, Trace,
// GetResolvedProjectReference) are delegated through interface embedding.
type cachingCompilerHost struct {
	compiler.CompilerHost
	cache *ParseCache
}

// WithParseCache wraps host with the shared parse cache. A nil cache returns
// the host unchanged, so callers need no branches. The cache must be created
// at the pipeline entry and passed down explicitly (I5) — never stored in a
// package-level singleton, which would silently leak sharing into unrelated
// paths (rule_tester, --api getAstInfo) and grow without bound in
// long-running processes.
func WithParseCache(host compiler.CompilerHost, cache *ParseCache) compiler.CompilerHost {
	if cache == nil {
		return host
	}
	return &cachingCompilerHost{CompilerHost: host, cache: cache}
}

func (h *cachingCompilerHost) GetSourceFile(opts ast.SourceFileParseOptions) *ast.SourceFile {
	text, ok := h.FS().ReadFile(opts.FileName) // I1: single read; hash and parse share this text
	if !ok {
		return nil // I4: same as the default host, and never cached
	}
	return h.cache.acquire(opts, text)
}
