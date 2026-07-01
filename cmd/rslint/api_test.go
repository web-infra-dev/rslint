package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	api "github.com/web-infra-dev/rslint/internal/api"
)

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

	// The fixture tsconfig covers exactly the 8 src/*.ts files and
	// no-unsafe-member-access reports 5 diagnostics across them. Exact counts
	// catch a partial-lint regression (the "lint all" default silently dropping
	// files) that a >0 check would miss.
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

	// 8 fixture files minus the one ignored entry.
	if len(response.LintedFiles) != 7 {
		t.Fatalf("expected 7 linted files (8 minus the ignored one), got %d: %v", len(response.LintedFiles), response.LintedFiles)
	}
	if response.FileCount != 7 {
		t.Fatalf("expected FileCount=7 (== len(LintedFiles)), got %d", response.FileCount)
	}
	linted := make(map[string]bool, len(response.LintedFiles))
	for _, f := range response.LintedFiles {
		if f == "src/index.ts" {
			t.Fatalf("ignored file src/index.ts must be absent from LintedFiles, got %v", response.LintedFiles)
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

// A file rooted by two tsconfig programs is linted (and diagnosed) by both, so
// RunLinter's per-program LintedFileCount counts it twice; FileCount must mirror
// len(LintedFiles) (deduped by canonical path) and report it once.
func TestHandleLint_FileCountDeduplicatesAcrossPrograms(t *testing.T) {
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

// TestHandleLint_SchemaDrivenRuleViaConfig guards against a regression where
// schema-driven rules (Schema + RunWithOptions) panicked with a nil-pointer
// dereference: GetEnabledRules was calling the now-nil legacy Run callback
// directly instead of dispatching to RunWithOptions with validated,
// default-hydrated options.
func TestHandleLint_SchemaDrivenRuleViaConfig(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	targetFile := filepath.Join(fixturesDir, "src", "index.ts")

	// no-console is a schema-driven (RunWithOptions) rule. Pass options via the
	// config object (the only rule-options surface in --api mode).
	config := json.RawMessage(`[{"rules": {"no-console": ["error", {"allow": ["log"]}]}}]`)

	handler := &IPCHandler{}
	response, err := handler.HandleLint(api.LintRequest{
		Config:           config,
		ConfigDirectory:  fixturesDir,
		WorkingDirectory: fixturesDir,
		Files:            []string{targetFile},
		FileContents: map[string]string{
			targetFile: "console.log('a'); console.error('b');\n",
		},
	})
	if err != nil {
		t.Fatalf("HandleLint returned error: %v", err)
	}

	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic (console.error, with console.log allowed), got %d: %+v", len(response.Diagnostics), response.Diagnostics)
	}
	if response.Diagnostics[0].RuleName != "no-console" {
		t.Fatalf("expected diagnostic from no-console, got %q", response.Diagnostics[0].RuleName)
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

// scenario b (in-memory file命门): a requested file outside every tsconfig
// Program (a "gap" file) must still be linted by non-type-aware rules via the
// fallback Program — it must NOT be silently skipped with 0 diagnostics.
func TestHandleLint_GapFile_NonTypeAwareRuleRuns(t *testing.T) {
	fixturesDir, err := filepath.Abs(filepath.Join("..", "..", "packages", "rslint", "fixtures"))
	if err != nil {
		t.Fatalf("resolve fixtures dir: %v", err)
	}
	// Non-empty `files` activates gap discovery; the tsconfig only covers
	// src/, so a file at the fixtures root is a gap file. array-type is a
	// non-type-aware (syntactic) rule, so it must still run on the fallback.
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

// scenario a (the唯一防线): a type-aware rule on a gap file (no type info) must
// be filtered out by the gate, NOT run against a nil TypeChecker (which would
// SIGSEGV the whole process). Reaching the assertions means no crash; the rule
// must have produced no diagnostic.
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
