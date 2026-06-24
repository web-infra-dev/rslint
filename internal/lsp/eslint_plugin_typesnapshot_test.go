package lsp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
	"github.com/microsoft/typescript-go/shim/project"
	"github.com/microsoft/typescript-go/shim/project/logging"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"

	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/typesnapshot"
)

// TestDispatchPluginLint_AttachesTypeSnapshot is the LSP type-aware linchpin test: it
// drives dispatchPluginLint against a REAL project-pool session (not a mock) and
// asserts the file it ships to the worker carries a NON-EMPTY, self-consistent
// TypeSnapshot.
//
// The chain under test (eslint_plugin.go dispatchPluginLint):
//
//	s.session.GetLanguageService(backgroundCtx, uri).GetProgram()
//	  → program.GetTypeChecker(backgroundCtx)   // LSP project pool returns its
//	                                              // DEFAULT query checker here —
//	                                              // NOT checkers[0]
//	  → linter.AttachTypeSnapshot(&input, queryChecker, program.GetSourceFile(path))
//	      → typesnapshot.Build + EncodeBinary → input.TypeSnapshot ([]byte)
//
// The open question this test answers: can that background QUERY checker (the one
// a real LSP project pool hands back by default) produce a snapshot whose
// Node2Type table is non-empty and internally consistent (every referenced
// type-id present, no node maps to two type-ids)? If it could not — e.g. because
// the query checker lazily type-checks and walking every node off it yielded
// nothing — type-aware plugin rules would silently get no type info under the
// LSP. This test fails loudly if that regresses.
func TestDispatchPluginLint_AttachesTypeSnapshot(t *testing.T) {
	if !bundled.Embedded {
		t.Skip("bundled lib files are not embedded; cannot build a type-checked program")
	}

	// ---- on-disk fixture: a tsconfig + a type-checked source file ----
	// `b = a` makes `b` infer `string | undefined`; the file is rich enough that
	// Build's node walk over the query checker must resolve several nodes to real
	// types (identifiers, the const declarations, the union annotation).
	dir := tspath.NormalizePath(t.TempDir())
	const aContent = "declare const a: string | undefined;\nconst b = a;\n"
	writeFile(t, filepath.Join(dir, "tsconfig.json"), `{
  "compilerOptions": { "strict": true, "target": "es2020", "module": "esnext", "moduleResolution": "bundler" },
  "files": ["a.ts"]
}`)
	writeFile(t, filepath.Join(dir, "a.ts"), aContent)

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	// Realpath-normalize so the URI we derive matches the path the program stores
	// its source file under (macOS /var → /private/var symlink, etc.). uriToPath
	// must round-trip back to this exact string for GetSourceFile to hit.
	dir = tspath.NormalizePath(fs.Realpath(dir))
	uri := lsproto.DocumentUri("file://" + dir + "/a.ts")
	configKey := "file://" + dir

	// Sanity: the URI must round-trip to the on-disk path the program will key on.
	if got := uriToPath(uri); got != dir+"/a.ts" {
		t.Fatalf("uri round-trip mismatch: uriToPath(%q) = %q, want %q", uri, got, dir+"/a.ts")
	}

	// ---- minimal Server wired with a REAL session (mirrors handleInitialized) ----
	s := newTestServer()
	s.backgroundCtx = context.Background()
	s.fs = fs
	s.cwd = dir
	s.session = project.NewSession(&project.SessionInit{
		BackgroundCtx: s.backgroundCtx,
		Options: &project.SessionOptions{
			CurrentDirectory:   dir,
			DefaultLibraryPath: bundled.LibPath(),
			PositionEncoding:   lsproto.PositionEncodingKindUTF8,
			WatchEnabled:       false, // no WatchFiles reverse request (no readLoop to answer it)
			LoggingEnabled:     false,
		},
		FS:     fs,
		Client: s,
		Logger: logging.NewLogger(io.Discard),
	})

	// Open the file so the session builds an overlay + configured project for it.
	s.documents[uri] = aContent
	s.session.DidOpenFile(s.backgroundCtx, uri, 1, aContent, lsproto.LanguageKindTypeScript)

	// Force the configured project's program to materialize now (synchronous flush)
	// and assert it actually type-checks our file, so a later empty snapshot can only
	// mean the snapshot builder failed — not that the program never loaded the file.
	ls, err := s.session.GetLanguageService(s.backgroundCtx, uri)
	if err != nil {
		t.Fatalf("GetLanguageService: %v", err)
	}
	program := ls.GetProgram()
	if program == nil {
		t.Fatal("GetLanguageService returned a nil program")
	}
	if program.GetSourceFile(dir+"/a.ts") == nil {
		t.Fatalf("program has no source file for %q; session did not load the opened file", dir+"/a.ts")
	}

	// ---- register a plain plugin rule (no requiresTypeChecking declaration) ----
	// Type-aware gating is project-based now: dispatchPluginLint's AttachTypeSnapshot
	// builds a snapshot whenever the file has a program, NOT gated on any meta. So a
	// plain plugin rule on a file with a program must still get a snapshot dispatched.
	config.RegisterEslintPluginRules([]config.EslintPluginEntry{
		{Prefix: "tlsp", RuleNames: []string{"need-type"}},
	})
	r, ok := config.GlobalRuleRegistry.GetRule("tlsp/need-type")
	if !ok || !r.IsEslintPluginRule {
		t.Fatalf("tlsp/need-type not registered as a plugin rule: ok=%v rule=%+v", ok, r)
	}
	s.jsConfigs[configKey] = config.RslintConfig{
		{Plugins: []string{"tlsp"}, Rules: config.Rules{"tlsp/need-type": "error"}},
	}

	// ---- capture the dispatched request; return an empty result ----
	// dispatchPluginLint builds input.TypeSnapshot on the MAIN goroutine
	// (AttachTypeSnapshot), then ships it via the dispatcher INSIDE its background
	// goroutine. The dispatcher's req carries the wire file we assert on.
	captured := make(chan linter.EslintPluginLintRequest, 1)
	s.eslintPluginDispatch = func(_ context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		select {
		case captured <- req:
		default:
		}
		// Empty per-file result keyed by path so applyEslintPluginResults matches
		// the input and produces zero diagnostics (we don't assert on diagnostics).
		return &linter.EslintPluginLintResult{
			Results: []linter.EslintPluginFileResult{{FilePath: uriToPath(uri)}},
		}, nil
	}

	// dispatchPluginLint's goroutine delivers its (empty) result to pluginResultCh;
	// newTestServer sizes that channel at 16 so the send never blocks even though we
	// never drain it — we assert on the CAPTURED request, not the merged result.
	s.docGeneration[uri] = 1
	s.dispatchPluginLint(uri, 1)

	var req linter.EslintPluginLintRequest
	select {
	case req = <-captured:
	case <-time.After(15 * time.Second):
		t.Fatal("dispatchPluginLint never invoked the dispatcher (no reverse request was built)")
	}

	if len(req.Files) != 1 {
		t.Fatalf("expected exactly 1 wire file, got %d", len(req.Files))
	}
	wf := req.Files[0]

	// ── command 1: the snapshot is non-empty on the wire ──
	if wf.TypeSnapshot == nil {
		t.Fatal("wire file TypeSnapshot is nil — the LSP did not attach any type snapshot for a type-aware file")
	}
	if len(wf.TypeSnapshot) == 0 {
		t.Fatal("wire file TypeSnapshot is empty ([]byte len 0)")
	}

	// ── command 2: it decodes and the type side actually hit ──
	snap, ok := typesnapshot.DecodeBinary(wf.TypeSnapshot)
	if !ok {
		t.Fatal("DecodeBinary returned ok=false — the attached snapshot is not a valid binary frame")
	}
	if len(snap.Node2Type) == 0 {
		t.Fatal("decoded snapshot has 0 node2type entries — the background query checker produced NO type bindings (type-aware linchpin broken)")
	}
	if len(snap.Types) == 0 {
		t.Fatal("decoded snapshot has 0 type blocks despite non-empty node2type")
	}

	// ── self-consistency: the query-checker snapshot is internally coherent ──
	// Every node2type entry references a type-id that is present in Types, and no
	// (tokenStart,end) span maps to two different type-ids (the invariant the
	// worker relies on to match an oxc node to exactly one type).
	seen := map[[2]int]typesnapshot.TypeID{}
	for _, e := range snap.Node2Type {
		if _, present := snap.Types[e.TypeID]; !present {
			t.Errorf("node2type entry (tokenStart=%d,end=%d) → type-id %d missing from Types", e.TokenStart, e.End, e.TypeID)
		}
		k := [2]int{e.TokenStart, e.End}
		if prev, dup := seen[k]; dup && prev != e.TypeID {
			t.Errorf("node span (tokenStart=%d,end=%d) maps to two type-ids %d and %d", e.TokenStart, e.End, prev, e.TypeID)
		}
		seen[k] = e.TypeID
	}

	// Transitive closure: every referenced member/arg/return type-id is also in Types.
	for _, b := range snap.Types {
		refs := append(append(append([]typesnapshot.TypeID{}, b.MemberTypes...), b.TypeArgs...), b.CallSigReturns...)
		for _, ref := range refs {
			if _, present := snap.Types[ref]; !present {
				t.Errorf("type %q(%d) references dangling type-id %d not in Types", b.Name, b.ID, ref)
			}
		}
	}

	t.Logf("query-checker snapshot OK: %d node2type entries, %d type blocks, undefinedID=%d",
		len(snap.Node2Type), len(snap.Types), snap.PrimTypes.Undefined)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
