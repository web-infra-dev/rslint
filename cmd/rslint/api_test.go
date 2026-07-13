package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

type canonicalPathBaseFS struct {
	vfs.FS
	realpathCalls atomic.Int32
}

func (fs *canonicalPathBaseFS) Realpath(filePath string) string {
	fs.realpathCalls.Add(1)
	return fs.FS.Realpath(filePath)
}

func TestCanonicalPathVFS_UsesRequestHintBeforeBaseFilesystem(t *testing.T) {
	base := &canonicalPathBaseFS{FS: osvfs.FS()}
	fsys := &canonicalPathVFS{
		FS:             base,
		canonicalPaths: map[string]string{exactFilesystemPathID("/lexical/a.ts"): "/physical/a.ts"},
	}
	if got := fsys.Realpath("/lexical/a.ts"); got != "/physical/a.ts" {
		t.Fatalf("Realpath returned %q, want request hint", got)
	}
	if calls := base.realpathCalls.Load(); calls != 0 {
		t.Fatalf("request hint must avoid base realpath, got %d calls", calls)
	}
}

func TestHandleLint_RejectsMismatchedCanonicalFiles(t *testing.T) {
	_, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Files:          []string{"a.ts"},
		CanonicalFiles: []string{"a.ts", "b.ts"},
	})
	if err == nil || !strings.Contains(err.Error(), "canonicalFiles must be parallel to files") {
		t.Fatalf("expected canonicalFiles length error, got %v", err)
	}
}

func TestHandleLint_DefaultsToLintAllFiles(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	// Rules come solely from the config object (no separate ruleOptions
	// surface); a single-rule config keeps the diagnostic count deterministic.
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/no-unsafe-member-access": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	// The fixture tsconfig covers 9 src/*.ts files; .gitignore removes one, so 8
	// are linted and no-unsafe-member-access reports 5 diagnostics across them.
	// Exact counts catch a partial-lint regression (the "lint all" default
	// silently dropping files) that a >0 check would miss.
	if response.FileCount != 8 {
		t.Fatalf("expected all 8 fixture files linted, got FileCount=%d", response.FileCount)
	}
	if len(response.Diagnostics) != 5 {
		t.Fatalf("expected 5 diagnostics across the fixture, got %d", len(response.Diagnostics))
	}
}

// A global-ignore entry must drop its files from LintedFiles, FileCount, and the
// diagnostics — and LintedFiles must share Diagnostic.FilePath's path space so
// the JS side seeds one result per actually-linted file.
func TestHandleLint_LintedFilesExcludesIgnored(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[
		{ "ignores": ["src/index.ts"] },
		{
			"files": ["**/*.ts"],
			"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
			"rules": { "@typescript-eslint/no-unsafe-member-access": "error" },
			"plugins": ["@typescript-eslint"]
		}
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	// 9 fixture files minus the gitignored fixture and the config-ignored entry.
	if len(response.LintedFiles) != 7 {
		t.Fatalf("expected 7 linted files (9 minus two ignored files), got %d: %v", len(response.LintedFiles), response.LintedFiles)
	}
	if response.FileCount != 7 {
		t.Fatalf("expected FileCount=7 (== len(LintedFiles)), got %d", response.FileCount)
	}
	linted := make(map[string]bool, len(response.LintedFiles))
	for _, f := range response.LintedFiles {
		if f == "src/index.ts" {
			t.Fatalf("ignored file src/index.ts must be absent from LintedFiles, got %v", response.LintedFiles)
		}
		if f == "src/gitignored.ts" {
			t.Fatalf("gitignored file src/gitignored.ts must be absent from LintedFiles, got %v", response.LintedFiles)
		}
		linted[f] = true
	}
	// LintedFiles is the same path space as Diagnostic.FilePath: every diagnostic
	// lands on a linted file, and none on the ignored one.
	for _, d := range response.Diagnostics {
		if d.FilePath == "src/index.ts" {
			t.Fatalf("no diagnostic should be reported on the ignored file, got %+v", d)
		}
		if !linted[d.FilePath] {
			t.Fatalf("diagnostic on %q is not in LintedFiles %v", d.FilePath, response.LintedFiles)
		}
	}
}

// When the only requested file is config-ignored, LintedFiles must be an empty
// (non-nil) slice — it serializes as `[]` so the JS side yields zero results
// rather than falling back to the glob matches (the phantom-empty-result bug).
func TestHandleLint_AllIgnored_EmptyLintedFiles(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[
		{ "ignores": ["src/index.ts"] },
		{
			"files": ["**/*.ts"],
			"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
			"rules": { "@typescript-eslint/no-unsafe-member-access": "error" },
			"plugins": ["@typescript-eslint"]
		}
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{filepath.Join(fixturesDir, "src", "index.ts")},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if len(response.LintedFiles) != 0 {
		t.Fatalf("expected empty LintedFiles for an all-ignored lint, got %v", response.LintedFiles)
	}
	// Pin the WIRE shape, not just len(): an empty (non-nil) slice must serialize
	// as `[]`, never `null` (a nil slice) or an absent field (omitempty). The JS
	// side keys its glob-fallback on the field being ABSENT, so a nil/omitempty
	// collapse would re-seed phantom empty results — and len()==0 is blind to
	// all three cases.
	b, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if !strings.Contains(string(b), `"lintedFiles":[]`) {
		t.Fatalf(`LintedFiles must serialize as "lintedFiles":[] (not null/absent), got: %s`, string(b))
	}
	if response.FileCount != 0 {
		t.Fatalf("expected FileCount=0, got %d", response.FileCount)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics, got %d", len(response.Diagnostics))
	}
}

func TestHandleLint_AllIgnoredDoesNotResolveInactiveProject(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored.ts")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	config := json.RawMessage(`[
		{ "ignores": ["ignored.ts"] },
		{
			"languageOptions": { "parserOptions": { "project": ["./missing.json"] } },
			"rules": { "no-debugger": "error" }
		}
	]`)

	response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("an inactive project must not fail plain API lint: %v", err)
	}
	if response.FileCount != 0 || len(response.LintedFiles) != 0 {
		t.Fatalf("ignored request should lint no files: %+v", response)
	}
}

func TestHandleLint_SelectedTargetResolvesGoverningProject(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "selected.ts")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	config := json.RawMessage(`[{
		"languageOptions": { "parserOptions": { "project": ["./missing.json"] } },
		"rules": { "no-debugger": "error" }
	}]`)

	_, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err == nil || !strings.Contains(err.Error(), "missing.json") {
		t.Fatalf("selected target must resolve its governing project, got %v", err)
	}
}

// When multiple declared projects contain a target, the first project owns the
// lint pass and the API reports one caller-visible result.
func TestHandleLint_FirstContainingProgramReportsFileOnce(t *testing.T) {
	dir := t.TempDir()

	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("shared.ts", "let a: Array<string> = [];\n")
	write("tsconfig.a.json", `{"include": ["shared.ts"]}`)
	write("tsconfig.b.json", `{"include": ["shared.ts"]}`)

	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.a.json", "./tsconfig.b.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{filepath.Join(dir, "shared.ts")},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	// shared.ts belongs to both programs → linted (and diagnosed) twice, but it
	// is ONE physical file: LintedFiles dedupes it and FileCount mirrors that.
	// Reverting FileCount to lintResult.LintedFileCount makes this report 2.
	if len(response.LintedFiles) != 1 {
		t.Fatalf("expected 1 unique linted file, got %d: %v", len(response.LintedFiles), response.LintedFiles)
	}
	if response.FileCount != 1 {
		t.Fatalf("expected FileCount=1 (deduped across programs), got %d — the per-program visit count must not leak into FileCount", response.FileCount)
	}
}

func TestHandleLint_ExplicitFilesRestrictsScope(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/no-unsafe-member-access": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	targetFile := filepath.Join(fixturesDir, "src", "index.ts")

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		Files:            []string{targetFile},
		WorkingDirectory: fixturesDir,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	if response.FileCount != 1 {
		t.Fatalf("expected lint to process exactly one file, got %d", response.FileCount)
	}
	if len(response.Diagnostics) == 0 {
		t.Fatal("expected lint to report diagnostics for the explicitly requested file")
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.FilePath != "src/index.ts" {
			t.Fatalf("expected diagnostics to be limited to src/index.ts, got %q", diagnostic.FilePath)
		}
	}
}

func TestHandleLint_ExplicitFileOutsideFilesIsCountedWithNoRules(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "explicit.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"rules": { "no-debugger": "error" }
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("explicit files-scope miss should still count as one lint result, got %d", response.FileCount)
	}
	if len(response.LintedFiles) != 1 || response.LintedFiles[0] != "explicit.js" {
		t.Fatalf("expected explicit.js in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics because no rules match the file, got %+v", response.Diagnostics)
	}
	if response.RuleCount != 0 {
		t.Fatalf("expected no executed rules, got %d", response.RuleCount)
	}
}

func TestHandleLint_ExplicitMalformedFileOutsideFilesReportsSyntaxDiagnostic(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "explicit.js")
	if err := os.WriteFile(target, []byte("const = ;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"rules": { "no-debugger": "error" }
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("explicit files-scope miss should still count as one lint result, got %d", response.FileCount)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].RuleName != "TypeScript(TS1134)" {
		t.Fatalf("expected the selected zero-rule target to report TS1134, got %+v", response.Diagnostics)
	}
	if response.RuleCount != 0 {
		t.Fatalf("expected no executed rules, got %d", response.RuleCount)
	}
}

func TestHandleLint_MalformedFileDoesNotRunRules(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "explicit.js")
	if err := os.WriteFile(target, []byte("debugger;\nconst = ;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config:           json.RawMessage(`[{"rules":{"no-debugger":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("malformed target should still count as one linted file, got %d", response.FileCount)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].RuleName != "TypeScript(TS1134)" {
		t.Fatalf("expected only TS1134 and no rule diagnostics, got %+v", response.Diagnostics)
	}
	if response.RuleCount != 0 {
		t.Fatalf("malformed target must not execute rules, got %d", response.RuleCount)
	}
}

func TestHandleLint_FileContentsDependenciesDoNotWidenExplicitTargets(t *testing.T) {
	dir := t.TempDir()
	writeProgramTestFiles(t, dir, map[string]string{
		"target.ts":     "import './dependency';\ndebugger;\n",
		"dependency.ts": "export const clean = true;\n",
		"tsconfig.json": `{"include":["*.ts"]}`,
	})
	target := filepath.Join(dir, "target.ts")
	dependency := filepath.Join(dir, "dependency.ts")
	config := json.RawMessage(`[{
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "no-debugger": "error" }
	}]`)

	response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents: map[string]string{
			dependency: "debugger;\n",
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 || len(response.LintedFiles) != 1 || response.LintedFiles[0] != "target.ts" {
		t.Fatalf("overlay dependency must not widen explicit targets: count=%d files=%v", response.FileCount, response.LintedFiles)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].FilePath != "target.ts" || response.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("expected only target.ts no-debugger diagnostic, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_FilesPresenceControlsWhetherFileContentsAreTargets(t *testing.T) {
	dir := t.TempDir()
	virtualFile := filepath.Join(dir, "virtual.ts")
	config := json.RawMessage(`[{"rules":{"no-debugger":"error"}}]`)

	tests := []struct {
		name            string
		files           []string
		wantFileCount   int
		wantDiagnostics int
	}{
		{
			name:            "files omitted",
			files:           nil,
			wantFileCount:   1,
			wantDiagnostics: 1,
		},
		{
			name:            "files explicitly empty",
			files:           []string{},
			wantFileCount:   0,
			wantDiagnostics: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
				Config:           config,
				ConfigDirectory:  dir,
				WorkingDirectory: dir,
				Files:            tt.files,
				FileContents:     map[string]string{virtualFile: "debugger;\n"},
			})
			if err != nil {
				t.Fatalf("HandleLint returned error: %v", err)
			}
			if response.FileCount != tt.wantFileCount || len(response.Diagnostics) != tt.wantDiagnostics {
				t.Fatalf("got fileCount=%d diagnostics=%d, want fileCount=%d diagnostics=%d", response.FileCount, len(response.Diagnostics), tt.wantFileCount, tt.wantDiagnostics)
			}
		})
	}
}

func TestHandleLint_SourceSnapshotsAreRequestScoped(t *testing.T) {
	dir := t.TempDir()
	virtualFile := filepath.Join(dir, "virtual.ts")
	config := json.RawMessage(`[{"rules":{"no-debugger":"error"}}]`)
	handler := &IPCHandler{}

	request := func(content string) *api.LintResponse {
		t.Helper()
		response, err := handler.HandleLint(api.LintRequest{
			Config:           config,
			ConfigDirectory:  dir,
			WorkingDirectory: dir,
			Files:            []string{virtualFile},
			FileContents:     map[string]string{virtualFile: content},
		})
		if err != nil {
			t.Fatalf("HandleLint returned error: %v", err)
		}
		return response
	}

	withDebugger := request("debugger;\n")
	if len(withDebugger.Diagnostics) != 1 || withDebugger.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("first request did not lint its overlay content: %+v", withDebugger.Diagnostics)
	}
	clean := request("export const clean = true;\n")
	if len(clean.Diagnostics) != 0 {
		t.Fatalf("second request reused the first request's source snapshot: %+v", clean.Diagnostics)
	}
}

func TestHandleLint_ExplicitFileEntryIgnoredIsCountedWithNoRules(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	config := json.RawMessage(`[{
		"files": ["**/*.js"],
		"ignores": ["ignored.js"],
		"rules": { "no-debugger": "error" }
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("entry-level ignores should not remove an explicit file from the result set, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 1 || response.LintedFiles[0] != "ignored.js" {
		t.Fatalf("expected ignored.js in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics because entry-level ignores leave no matching rules, got %+v", response.Diagnostics)
	}
	if response.RuleCount != 0 {
		t.Fatalf("expected no executed rules, got %d", response.RuleCount)
	}
}

func TestHandleLint_ExplicitFileGloballyIgnoredIsSkipped(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	config := json.RawMessage(`[
		{ "ignores": ["ignored.js"] },
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 0 {
		t.Fatalf("globally ignored explicit files should be skipped, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 0 {
		t.Fatalf("globally ignored file must not be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for globally ignored file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_ExplicitFileGitignoredIsSkipped(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored.js\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	config := json.RawMessage(`[
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 0 {
		t.Fatalf("gitignored explicit files should be skipped, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 0 {
		t.Fatalf("gitignored file must not be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for gitignored file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_OverlayGitignoreIsApplied(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "source.js")
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	response, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config: json.RawMessage(`[
			{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
		]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents: map[string]string{
			filepath.Join(dir, ".gitignore"): "source.js\n",
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 0 || len(response.LintedFiles) != 0 || len(response.Diagnostics) != 0 {
		t.Fatalf("overlay .gitignore was not applied: %+v", response)
	}
}

func TestHandleLint_GitignoredParentBlocksNestedNegation(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "ignored", "src", "file.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("ignored/\n"), 0o644); err != nil {
		t.Fatalf("write root .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored", ".gitignore"), []byte("!src/file.js\n"), 0o644); err != nil {
		t.Fatalf("write nested .gitignore: %v", err)
	}
	config := json.RawMessage(`[
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 0 {
		t.Fatalf("parent-gitignored explicit file should stay skipped despite nested negation, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 0 {
		t.Fatalf("parent-gitignored file must not be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics for parent-gitignored file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_ParentGitignoreDoesNotSuppressConfigRoot(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "packages", "app")
	target := filepath.Join(appDir, "src", "file.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("packages/\n"), 0o644); err != nil {
		t.Fatalf("write root .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, ".gitignore"), []byte("!src/file.js\n"), 0o644); err != nil {
		t.Fatalf("write app .gitignore: %v", err)
	}
	config := json.RawMessage(`[
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  appDir,
		WorkingDirectory: appDir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("parent .gitignore must not suppress a config-owned file, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 1 {
		t.Fatalf("config-owned file must be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("expected no-debugger for config-owned file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_IntermediateParentGitignoreIsNotRead(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "packages", "app")
	target := filepath.Join(appDir, "src", "file.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("packages/\n"), 0o644); err != nil {
		t.Fatalf("write root .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "packages", ".gitignore"), []byte("!app/src/file.js\n"), 0o644); err != nil {
		t.Fatalf("write packages .gitignore: %v", err)
	}
	config := json.RawMessage(`[
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  appDir,
		WorkingDirectory: appDir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("parent .gitignore files must not suppress a config-owned file, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 1 {
		t.Fatalf("config-owned file must be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("expected no-debugger for config-owned file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_ParentWildcardDoesNotReadIntermediateGitignore(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "packages", "app")
	target := filepath.Join(appDir, "src", "generated", "file.js")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("packages/*\n!packages/app/\n"), 0o644); err != nil {
		t.Fatalf("write root .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "packages", ".gitignore"), []byte("app/src/generated/\n"), 0o644); err != nil {
		t.Fatalf("write packages .gitignore: %v", err)
	}
	config := json.RawMessage(`[
		{ "files": ["**/*.js"], "rules": { "no-debugger": "error" } }
	]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  appDir,
		WorkingDirectory: appDir,
		Files:            []string{target},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("parent intermediate .gitignore must not skip generated explicit file, got FileCount=%d", response.FileCount)
	}
	if len(response.LintedFiles) != 1 {
		t.Fatalf("config-owned generated file must be in LintedFiles, got %v", response.LintedFiles)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].RuleName != "no-debugger" {
		t.Fatalf("expected no-debugger for config-owned generated file, got %+v", response.Diagnostics)
	}
}

func TestHandleLint_NoFilesEntryScansDefaultExtensions(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a.jsx":      "debugger;\n",
		"b.cjs":      "debugger;\n",
		"c.cts":      "debugger;\n",
		"styles.css": "debugger;\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	config := json.RawMessage(`[{
		"rules": { "no-debugger": "error" }
	}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	expectedFiles := []string{"a.jsx", "b.cjs", "c.cts"}
	if strings.Join(response.LintedFiles, ",") != strings.Join(expectedFiles, ",") {
		t.Fatalf("expected default lint files %v, got %v", expectedFiles, response.LintedFiles)
	}
	if response.FileCount != len(expectedFiles) {
		t.Fatalf("expected FileCount=%d, got %d", len(expectedFiles), response.FileCount)
	}
	if len(response.Diagnostics) != len(expectedFiles) {
		t.Fatalf("expected one no-debugger diagnostic per default file, got %d: %+v", len(response.Diagnostics), response.Diagnostics)
	}
	for _, diagnostic := range response.Diagnostics {
		if diagnostic.FilePath == "styles.css" {
			t.Fatalf("unsupported css file must not be linted, diagnostics=%+v", response.Diagnostics)
		}
	}
}

func TestHandleLint_NoConfigEnablesNoRules(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}

	handler := &IPCHandler{}
	// No Config: the JS side resolves config; an absent config means "no rules"
	// (clean 0-diagnostic return, NOT a crash and NOT auto-discovery). Mirrors
	// ESLint flat-config (a file matched by no config runs no rules).
	response, err := handler.HandleLint(api.LintRequest{
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{filepath.Join(fixturesDir, "src", "index.ts")},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if len(response.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics with no config, got %d", len(response.Diagnostics))
	}
}

func TestHandleLint_FileContentsKeepTypeInfoWhenProgramUsesSymlinkSource(t *testing.T) {
	realDir := t.TempDir()
	linkDir := filepath.Join(filepath.Dir(realDir), filepath.Base(realDir)+"-link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	defer os.Remove(linkDir)

	srcDir := filepath.Join(realDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.ts"), []byte("const clean = 1;\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "tsconfig.json"), []byte(`{"include":["src/a.ts"]}`), 0o644); err != nil {
		t.Fatalf("write tsconfig: %v", err)
	}

	config := json.RawMessage(`[{
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"plugins": ["@typescript-eslint"],
		"rules": { "@typescript-eslint/no-unsafe-member-access": "error" }
	}]`)
	realTarget := filepath.Join(realDir, "src", "a.ts")

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:                    config,
		ConfigDirectory:           linkDir,
		WorkingDirectory:          linkDir,
		Files:                     []string{realTarget},
		FileContents:              map[string]string{realTarget: "let b: any = 10;\nb.c = 20;\n"},
		IncludeEncodedSourceFiles: true,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected one type-aware overlay diagnostic, got %d: %+v", len(response.Diagnostics), response.Diagnostics)
	}
	if got := response.Diagnostics[0].RuleName; got != "@typescript-eslint/no-unsafe-member-access" {
		t.Fatalf("expected no-unsafe-member-access diagnostic from typed overlay content, got %q", got)
	}
	expectedPath, err := filepath.Rel(linkDir, realTarget)
	if err != nil {
		t.Fatalf("relative target path: %v", err)
	}
	expectedPath = filepath.ToSlash(expectedPath)
	if len(response.LintedFiles) != 1 || response.LintedFiles[0] != expectedPath {
		t.Fatalf("expected requested target path %q in LintedFiles, got %v", expectedPath, response.LintedFiles)
	}
	if response.Diagnostics[0].FilePath != expectedPath {
		t.Fatalf("expected requested target path %q in diagnostic, got %q", expectedPath, response.Diagnostics[0].FilePath)
	}
	if _, ok := response.EncodedSourceFiles[expectedPath]; !ok {
		t.Fatalf("expected requested target path %q in encoded source files, got keys %v", expectedPath, response.EncodedSourceFiles)
	}
}

// An in-memory target outside every tsconfig Program must still run
// non-type-aware rules through the fallback Program.
func TestHandleLint_GapFile_NonTypeAwareRuleRuns(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	// The tsconfig only covers src/, so a selected file at the fixtures root is
	// bound to the non-project-backed fallback. array-type is a non-type-aware
	// (syntactic) rule, so it must still run there.
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-scenario-b.ts")

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "let a: Array<string> = [];\n"},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if response.FileCount != 1 {
		t.Fatalf("expected the gap file to be linted (fileCount=1), got %d", response.FileCount)
	}
	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic on the gap file, got %d: %+v", len(response.Diagnostics), response.Diagnostics)
	}
	if got := response.Diagnostics[0].RuleName; got != "@typescript-eslint/array-type" {
		t.Fatalf("expected array-type diagnostic, got rule %q", got)
	}
	if got := response.Diagnostics[0].MessageId; got != "errorStringArray" {
		t.Fatalf("expected messageId errorStringArray, got %q", got)
	}
}

// A type-aware rule on a gap file must be filtered before execution. Running
// it with a nil TypeChecker would crash the process.
func TestHandleLint_GapFile_TypeAwareRuleGatedOff(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	// array-type (non-type-aware) + no-unsafe-member-access (type-aware) on the
	// SAME gap file. The gate must drop the type-aware rule but keep array-type,
	// so array-type firing proves the file was actually linted — without it, the
	// "no type-aware diagnostic" assertion would also hold if the file were
	// silently not linted at all (a false green this distinguishes against).
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": {
			"@typescript-eslint/no-unsafe-member-access": "error",
			"@typescript-eslint/array-type": "error"
		},
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-scenario-a.ts")

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "let a: Array<string> = [];\nlet b: any = 10;\nb.c = 20;\n"},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	sawArrayType := false
	for _, d := range response.Diagnostics {
		if d.RuleName == "@typescript-eslint/no-unsafe-member-access" {
			t.Fatalf("type-aware rule must be gated off the gap file, but it reported: %+v", d)
		}
		if d.RuleName == "@typescript-eslint/array-type" {
			sawArrayType = true
		}
	}
	if !sawArrayType {
		t.Fatal("expected the non-type-aware array-type rule to fire on the gap file: " +
			"its absence would mean the file was never linted, making the " +
			"\"no type-aware diagnostic\" assertion vacuous")
	}
}

func TestHandleLint_MalformedConfigReturnsError(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}

	handler := &IPCHandler{}
	// A config that is not a RslintConfigEntry[] must surface as an IPC error,
	// not panic / kill the long-lived --api process.
	_, err = handler.HandleLint(api.LintRequest{
		Config:           json.RawMessage(`{"not":"an array"}`),
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
	})
	if err == nil {
		t.Fatal("expected an error for a malformed (non-array) config")
	}
}

func TestHandleLint_RejectsEmptyFilesArrayConfig(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}

	handler := &IPCHandler{}
	_, err = handler.HandleLint(api.LintRequest{
		Config:           json.RawMessage(`[{"files":[],"rules":{"no-console":"error"}}]`),
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
	})
	if err == nil {
		t.Fatal("expected an error for empty files array")
	}
	if !strings.Contains(err.Error(), `key "files": expected value to be a non-empty array`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ④: severity is emitted as a string ("error"/"warning") and counts are split
// by level — errorCount counts only errors, not the total (ESLint semantics).
func TestHandleLint_SeverityStringAndCountBuckets(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	gapFile := filepath.Join(fixturesDir, "gap-severity.ts")
	const content = "let a: Array<string> = [];\n" // array-type fires exactly once

	run := func(level string) *api.LintResponse {
		config := json.RawMessage(`[{
			"files": ["**/*.ts"],
			"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
			"rules": { "@typescript-eslint/array-type": "` + level + `" },
			"plugins": ["@typescript-eslint"]
		}]`)
		handler := &IPCHandler{}
		resp, runErr := handler.HandleLint(api.LintRequest{
			Config:           config,
			ConfigDirectory:  fixturesDir,
			WorkingDirectory: fixturesDir,
			Files:            []string{gapFile},
			FileContents:     map[string]string{gapFile: content},
		})
		if runErr != nil {
			t.Fatalf("HandleLint(%s): %v", level, runErr)
		}
		return resp
	}

	errResp := run("error")
	if len(errResp.Diagnostics) != 1 {
		t.Fatalf("error level: expected 1 diagnostic, got %d", len(errResp.Diagnostics))
	}
	if errResp.Diagnostics[0].Severity != "error" {
		t.Fatalf("error level: expected severity \"error\", got %q", errResp.Diagnostics[0].Severity)
	}
	if errResp.ErrorCount != 1 || errResp.WarningCount != 0 {
		t.Fatalf("error level: expected errorCount=1 warningCount=0, got %d/%d", errResp.ErrorCount, errResp.WarningCount)
	}

	warnResp := run("warn")
	if len(warnResp.Diagnostics) != 1 {
		t.Fatalf("warn level: expected 1 diagnostic, got %d", len(warnResp.Diagnostics))
	}
	// rslint's DiagnosticSeverity.String() uses config-level spelling ("warn",
	// not "warning"); the ESLint numeric mapping (1|2) is the new Rslint
	// class's job, not Go's.
	if warnResp.Diagnostics[0].Severity != "warn" {
		t.Fatalf("warn level: expected severity \"warn\", got %q", warnResp.Diagnostics[0].Severity)
	}
	if warnResp.ErrorCount != 0 || warnResp.WarningCount != 1 {
		t.Fatalf("warn level: expected errorCount=0 warningCount=1, got %d/%d", warnResp.ErrorCount, warnResp.WarningCount)
	}
}

// ④: suggestions (optional, user-selected fixes) are converted from the rule
// diagnostic with messageId / message / fixes preserved.
func TestHandleLint_SuggestionsConverted(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	// no-explicit-any is non-type-aware and emits exactly two suggestions
	// (unknown / never) for a plain `any`.
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/no-explicit-any": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-suggestions.ts")

	handler := &IPCHandler{}
	resp, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "let a: any;\n"},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	var anyDiag *api.Diagnostic
	for i := range resp.Diagnostics {
		if resp.Diagnostics[i].RuleName == "@typescript-eslint/no-explicit-any" {
			anyDiag = &resp.Diagnostics[i]
			break
		}
	}
	if anyDiag == nil {
		t.Fatalf("expected a no-explicit-any diagnostic, got %+v", resp.Diagnostics)
	}
	if anyDiag.MessageId != "unexpectedAny" {
		t.Fatalf("expected messageId unexpectedAny, got %q", anyDiag.MessageId)
	}
	if len(anyDiag.Suggestions) != 2 {
		t.Fatalf("expected 2 suggestions (unknown/never), got %d: %+v", len(anyDiag.Suggestions), anyDiag.Suggestions)
	}
	if anyDiag.Suggestions[0].MessageId != "suggestUnknown" {
		t.Fatalf("expected suggestion[0] messageId suggestUnknown, got %q", anyDiag.Suggestions[0].MessageId)
	}
	if len(anyDiag.Suggestions[0].Fixes) != 1 || anyDiag.Suggestions[0].Fixes[0].Text != "unknown" {
		t.Fatalf("expected suggestion[0] single fix text \"unknown\", got %+v", anyDiag.Suggestions[0].Fixes)
	}
	if anyDiag.Suggestions[1].MessageId != "suggestNever" {
		t.Fatalf("expected suggestion[1] messageId suggestNever, got %q", anyDiag.Suggestions[1].MessageId)
	}
	if len(anyDiag.Suggestions[1].Fixes) != 1 || anyDiag.Suggestions[1].Fixes[0].Text != "never" {
		t.Fatalf("expected suggestion[1] single fix text \"never\", got %+v", anyDiag.Suggestions[1].Fixes)
	}
}

// F2: pins the suggestion `data` conversion (api.go: Data: sug.Message.Data).
// no-explicit-any's suggestions carry no data, so use no-restricted-types whose
// suggestion carries {name, replacement}.
func TestHandleLint_SuggestionDataConverted(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/no-restricted-types": ["error", { "types": { "Foo": { "suggest": ["Bar"] } } }] },
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-restricted.ts")

	handler := &IPCHandler{}
	resp, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "let x: Foo;\n"},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	var diag *api.Diagnostic
	for i := range resp.Diagnostics {
		if resp.Diagnostics[i].RuleName == "@typescript-eslint/no-restricted-types" {
			diag = &resp.Diagnostics[i]
			break
		}
	}
	if diag == nil {
		t.Fatalf("expected a no-restricted-types diagnostic, got %+v", resp.Diagnostics)
	}
	if len(diag.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d: %+v", len(diag.Suggestions), diag.Suggestions)
	}
	s := diag.Suggestions[0]
	if s.MessageId != "bannedTypeReplacement" {
		t.Fatalf("expected suggestion messageId bannedTypeReplacement, got %q", s.MessageId)
	}
	if s.Data["name"] != "Foo" || s.Data["replacement"] != "Bar" {
		t.Fatalf("expected suggestion data {name:Foo, replacement:Bar}, got %+v", s.Data)
	}
	if len(s.Fixes) != 1 || s.Fixes[0].Text != "Bar" {
		t.Fatalf("expected suggestion fix text \"Bar\", got %+v", s.Fixes)
	}
}

// ⑥: fix:true applies fixes in-band and returns the fixed source per file in
// Output (not written to disk), plus fixable*Count split by severity.
func TestHandleLint_FixProducesInBandOutput(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-fix.ts")

	handler := &IPCHandler{}
	resp, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "let a: Array<string> = [];\n"},
		Fix:              true,
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if resp.FixableErrorCount != 1 || resp.FixableWarningCount != 0 {
		t.Fatalf("expected fixableErrorCount=1 fixableWarningCount=0, got %d/%d", resp.FixableErrorCount, resp.FixableWarningCount)
	}
	if resp.Output == nil {
		t.Fatal("expected non-nil Output for fix:true")
	}
	got, ok := resp.Output["gap-fix.ts"]
	if !ok {
		t.Fatalf("expected in-band output keyed \"gap-fix.ts\", got %d entries", len(resp.Output))
	}
	if want := "let a: string[] = [];\n"; got != want {
		t.Fatalf("expected fixed output %q, got %q", want, got)
	}
}

// ⑥: fix/suggestion ranges are flat UTF-16 offsets, not byte offsets. A CJK
// char before the fix (1 UTF-16 unit but 3 UTF-8 bytes) makes the two diverge.
func TestHandleLint_FixRangeIsUTF16(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	config := json.RawMessage(`[{
		"files": ["**/*.ts"],
		"languageOptions": { "parserOptions": { "project": ["./tsconfig.json"] } },
		"rules": { "@typescript-eslint/array-type": "error" },
		"plugins": ["@typescript-eslint"]
	}]`)
	gapFile := filepath.Join(fixturesDir, "gap-utf16.ts")

	// "type 名 = Array<string>;" — `Array<string>` begins at UTF-16 offset 9
	// ("type 名 = " is 9 UTF-16 units) but byte offset 11 (名 is 3 bytes).
	handler := &IPCHandler{}
	resp, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{gapFile},
		FileContents:     map[string]string{gapFile: "type 名 = Array<string>;\n"},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}
	if len(resp.Diagnostics) != 1 || len(resp.Diagnostics[0].Fixes) != 1 {
		t.Fatalf("expected 1 diagnostic with 1 fix, got %+v", resp.Diagnostics)
	}
	if got := resp.Diagnostics[0].Fixes[0].StartPos; got != 9 {
		t.Fatalf("expected fix StartPos 9 (UTF-16 offset, not byte 11), got %d", got)
	}
}

type apiRequesterFunc func(context.Context, ipc.MessageKind, any) (*ipc.Message, error)

func (f apiRequesterFunc) SendRequest(ctx context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
	return f(ctx, kind, payload)
}

func TestHandleLint_EslintPluginDiagnosticAndFix(t *testing.T) {
	dir := t.TempDir()
	pluginConfigDirectory := filepath.Join(dir, "authored-config")
	target := filepath.Join(dir, "input.js")
	const source = "const bad = 1;\n"
	config := json.RawMessage(`[{
		"plugins": ["community"],
		"rules": { "community/rename": ["error", { "replacement": "good" }] }
	}]`)

	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		calls.Add(1)
		if kind != api.KindPluginLint {
			t.Fatalf("expected pluginLint reverse request, got %q", kind)
		}
		req, ok := payload.(linter.EslintPluginLintRequest)
		if !ok {
			t.Fatalf("unexpected pluginLint payload type %T", payload)
		}
		if !req.Fix || req.SuggestionsMode != linter.SuggestionsModeEager {
			t.Fatalf("unexpected fix settings: fix=%v suggestions=%q", req.Fix, req.SuggestionsMode)
		}
		if len(req.Files) != 1 || req.Files[0].Text == nil || *req.Files[0].Text != source {
			t.Fatalf("plugin request must carry the exact overlay text, got %+v", req.Files)
		}
		if req.Files[0].ConfigKey != pluginConfigDirectory {
			t.Fatalf("expected configKey %q, got %q", pluginConfigDirectory, req.Files[0].ConfigKey)
		}
		ruleConfig, ok := req.Rules["community/rename"]
		if !ok || len(ruleConfig.Options) != 1 {
			t.Fatalf("missing normalized plugin rule options: %+v", req.Rules)
		}

		result := linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath: req.Files[0].Path,
			Diagnostics: []linter.EslintPluginDiagnostic{{
				RuleName:  "community/rename",
				MessageId: "rename",
				Message:   "rename bad",
				StartPos:  6,
				EndPos:    9,
				Fixes: []linter.EslintPluginFix{{
					Range: [2]int{6, 9},
					Text:  "good",
				}},
			}},
		}}}
		return ipc.NewMessage(ipc.KindResponse, 1, result)
	})

	response, err := (&IPCHandler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Config:                config,
		ConfigDirectory:       dir,
		PluginConfigDirectory: pluginConfigDirectory,
		WorkingDirectory:      dir,
		Files:                 []string{target},
		FileContents:          map[string]string{target: source},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "community",
			RuleNames: []string{"rename"},
		}},
		Fix: true,
	}, requester)
	if err != nil {
		t.Fatalf("HandleLintWithContext returned error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one pluginLint request, got %d", calls.Load())
	}
	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected one plugin diagnostic, got %+v", response.Diagnostics)
	}
	diagnostic := response.Diagnostics[0]
	if diagnostic.RuleName != "community/rename" || diagnostic.MessageId != "rename" || diagnostic.FilePath != "input.js" {
		t.Fatalf("unexpected plugin diagnostic: %+v", diagnostic)
	}
	if diagnostic.Range.Start.Line != 1 || diagnostic.Range.Start.Column != 7 ||
		diagnostic.Range.End.Line != 1 || diagnostic.Range.End.Column != 10 {
		t.Fatalf("unexpected plugin diagnostic range: %+v", diagnostic.Range)
	}
	if len(diagnostic.Fixes) != 1 || diagnostic.Fixes[0].StartPos != 6 || diagnostic.Fixes[0].EndPos != 9 {
		t.Fatalf("unexpected plugin fix: %+v", diagnostic.Fixes)
	}
	if response.ErrorCount != 1 || response.FixableErrorCount != 1 || response.RuleCount != 1 {
		t.Fatalf("unexpected plugin counts: errors=%d fixable=%d rules=%d", response.ErrorCount, response.FixableErrorCount, response.RuleCount)
	}
	if got := response.Output["input.js"]; got != "const good = 1;\n" {
		t.Fatalf("expected plugin fix in Output, got %q", got)
	}
}

func TestHandleLint_EslintPluginSyntaxErrorSkipsDispatch(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "broken.js")
	var calls atomic.Int32
	requester := apiRequesterFunc(func(context.Context, ipc.MessageKind, any) (*ipc.Message, error) {
		calls.Add(1)
		return nil, errors.New("plugin dispatcher must not be called for a syntax error")
	})

	response, err := (&IPCHandler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Config:           json.RawMessage(`[{"plugins":["syntax-plugin"],"rules":{"syntax-plugin/rule":"error"}}]`),
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: "const = ;\n"},
		EslintPlugins: []api.EslintPluginEntry{{
			Prefix:    "syntax-plugin",
			RuleNames: []string{"rule"},
		}},
	}, requester)
	if err != nil {
		t.Fatalf("HandleLintWithContext returned error: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("syntax-error file issued %d pluginLint requests", calls.Load())
	}
	if len(response.Diagnostics) != 1 || !strings.HasPrefix(response.Diagnostics[0].RuleName, "TypeScript(TS") {
		t.Fatalf("expected only the syntax diagnostic, got %+v", response.Diagnostics)
	}
	if response.RuleCount != 0 {
		t.Fatalf("syntax-error file must execute no rules, got %d", response.RuleCount)
	}
}

func TestHandleLint_NoEslintPluginMetadataDoesNotDispatchStalePlaceholder(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "input.js")
	const source = "const value = 1;\n"
	config := json.RawMessage(`[{"plugins":["request-plugin"],"rules":{"request-plugin/rule":"error"}}]`)
	var calls atomic.Int32
	requester := apiRequesterFunc(func(_ context.Context, _ ipc.MessageKind, payload any) (*ipc.Message, error) {
		calls.Add(1)
		req := payload.(linter.EslintPluginLintRequest)
		return ipc.NewMessage(ipc.KindResponse, 1, linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{
			FilePath: req.Files[0].Path,
		}}})
	})
	handler := &IPCHandler{}
	baseRequest := api.LintRequest{
		Config:           config,
		ConfigDirectory:  dir,
		WorkingDirectory: dir,
		Files:            []string{target},
		FileContents:     map[string]string{target: source},
	}

	withMetadata := baseRequest
	withMetadata.EslintPlugins = []api.EslintPluginEntry{{Prefix: "request-plugin", RuleNames: []string{"rule"}}}
	if _, err := handler.HandleLintWithContext(context.Background(), withMetadata, requester); err != nil {
		t.Fatalf("register plugin placeholder: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected initial plugin dispatch, got %d", calls.Load())
	}

	response, err := handler.HandleLintWithContext(context.Background(), baseRequest, requester)
	if err != nil {
		t.Fatalf("lint without metadata: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("request without metadata unexpectedly dispatched pluginLint; calls=%d", calls.Load())
	}
	if response.RuleCount != 0 || len(response.Diagnostics) != 0 {
		t.Fatalf("stale placeholder leaked into metadata-free request: rules=%d diagnostics=%+v", response.RuleCount, response.Diagnostics)
	}
}
