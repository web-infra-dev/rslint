package linter

import (
	"context"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Two files in the same Program with the SAME rule name but DIFFERENT
// rule options must each get their own options through to the dispatcher.
// The pre-fix implementation built a single `unionRules` map keyed by
// rule name — `unionRules[ruleName] = cfg` overwrote per-file options,
// leaving the LAST file's options applied to every file. Silently
// produced wrong diagnostics in multi-config-shared-tsconfig setups
// (and in any single config that uses ESLint's `overrides` mechanism).
func TestRunLinter_CompatRulesPerFile_DivergentOptionsNotMerged(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"strict.ts": "const x = 1;",
		"loose.ts":  "const y = 2;",
	})

	// Build per-file ConfiguredRule sets: same plugin rule, different
	// options. This is the exact shape a multi-config monorepo with a
	// shared tsconfig would produce — file `strict.ts` from config A
	// (option {strict:true}) and `loose.ts` from config B
	// (option {strict:false}).
	strictOpts := []any{map[string]any{"strict": true}}
	looseOpts := []any{map[string]any{"strict": false}}

	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		var opts []any
		var configKey string
		if pathEndsWith(sf.FileName(), "strict.ts") {
			opts = strictOpts
			configKey = "/configA"
		} else {
			opts = looseOpts
			configKey = "/configB"
		}
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          configKey,
			Options:            opts,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	// Capture every CompatBatch the dispatcher receives. With the fix
	// each batch's `Rules` map should reflect ONE file's options
	// (because diverging-rule files now go into separate buckets);
	// without the fix, files with mismatched options share a single
	// batch where the last-written options apply to everyone.
	var dispatchedBatches []CompatBatch
	var batchMu sync.Mutex
	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		batchMu.Lock()
		// Deep-copy because the production code mutates `unionRules` per
		// file iteration; we want the snapshot at dispatch time.
		copyRules := make(map[string]CompatRuleConfig, len(batch.Rules))
		for k, v := range batch.Rules {
			copyRules[k] = v
		}
		copyFiles := make([]CompatLintFile, len(batch.Files))
		copy(copyFiles, batch.Files)
		dispatchedBatches = append(dispatchedBatches, CompatBatch{
			Files: copyFiles,
			Rules: copyRules,
		})
		batchMu.Unlock()
		return nil, nil
	})

	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})

	// For each dispatched batch, every file's expected options must match
	// the batch's Rules entry — i.e. the batch is HOMOGENEOUS.
	batchMu.Lock()
	defer batchMu.Unlock()
	if len(dispatchedBatches) == 0 {
		t.Fatal("dispatcher was never called")
	}
	for i, b := range dispatchedBatches {
		ruleCfg, ok := b.Rules["fake-plugin/no-x"]
		if !ok {
			t.Errorf("batch[%d]: missing rule fake-plugin/no-x", i)
			continue
		}
		for _, f := range b.Files {
			wantOpts := strictOpts
			if pathEndsWith(f.Path, "loose.ts") {
				wantOpts = looseOpts
			}
			gotJSON := normalizeOptsForCompare(ruleCfg.Options)
			wantJSON := normalizeOptsForCompare(wantOpts)
			if gotJSON != wantJSON {
				t.Errorf("batch[%d] file %s: rule options got %v, want %v",
					i, f.Path, ruleCfg.Options, wantOpts)
			}
		}
	}
}

// Single config (all files share rule options): bucketing collapses to
// exactly ONE dispatcher call. This is the typical CLI / LSP path and
// the fix must not regress its cost (extra IPC roundtrip per program).
func TestRunLinter_CompatRulesPerFile_HomogeneousFilesStillEmitOneBatch(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
		"b.ts": "const y = 2;",
		"c.ts": "const z = 3;",
	})

	sharedOpts := []any{map[string]any{"strict": true}}
	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          "/sharedConfig",
			Options:            sharedOpts,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	var calls int
	var mu sync.Mutex
	dispatcher := CompatBatchHandler(func(_ context.Context, _ CompatBatch) ([]CompatFileResult, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil, nil
	})

	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Errorf("expected 1 dispatcher call for homogeneous-rule program; got %d", calls)
	}
}

// Worker-emitted ruleErrors must be surfaced to stderr. Before the fix
// the Go side dropped the field on the floor — a plugin's create() bug
// showed up as "rule silently doesn't fire" with zero diagnostics, no
// indication to the user.
func TestRunLinter_CompatRuleErrors_ReportedToStderr(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/broken-rule",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          "/c",
			Options:            nil,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	// Capture stderr by redirecting os.Stderr to a pipe. We do this
	// process-wide for the test scope and restore in cleanup.
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr; _ = r.Close() })

	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		results := make([]CompatFileResult, len(batch.Files))
		for i, f := range batch.Files {
			results[i] = CompatFileResult{
				FilePath: f.Path,
				RuleErrors: []CompatRuleError{{
					Rule:    "fake-plugin/broken-rule",
					Message: "create: TypeError: cannot read property 'x' of undefined",
				}},
			}
		}
		return results, nil
	})

	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})

	// Close the writer so the pipe reader sees EOF.
	_ = w.Close()
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	stderr := string(buf[:n])

	if !strings.Contains(stderr, "rule fake-plugin/broken-rule failed") {
		t.Errorf("expected stderr to mention failed rule, got: %q", stderr)
	}
	if !strings.Contains(stderr, "TypeError") {
		t.Errorf("expected stderr to include the worker's error message, got: %q", stderr)
	}
}

// Cancelled compat files MUST NOT stream their partial diagnostics. A
// superseded LSP keystroke would otherwise paint stale diagnostics
// before the fresh lint overwrites them — visible flicker. The fix:
// when `fileResult.Cancelled` is true, skip the file entirely.
func TestRunLinter_CompatCancelledFile_DropsPartialDiagnostics(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/rule",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          "/c",
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		// Worker reports cancelled=true AND a partial diagnostic from
		// before the cancel flag fired. Go side must drop the partial.
		results := make([]CompatFileResult, len(batch.Files))
		for i, f := range batch.Files {
			results[i] = CompatFileResult{
				FilePath:  f.Path,
				Cancelled: true,
				Diagnostics: []CompatDiagnostic{{
					RuleName:  "fake-plugin/rule",
					Message:   "partial — should be dropped",
					MessageId: "x",
					StartPos:  0,
					EndPos:    1,
				}},
			}
		}
		return results, nil
	})

	var got []rule.RuleDiagnostic
	var mu sync.Mutex
	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			got = append(got, d)
			mu.Unlock()
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 0 {
		t.Errorf("expected 0 diagnostics from cancelled file; got %d", len(got))
	}
}

func pathEndsWith(p, suffix string) bool {
	if len(p) < len(suffix) {
		return false
	}
	return p[len(p)-len(suffix):] == suffix
}

// normalizeOptsForCompare turns an opaque options slice into a stable
// string for comparison. We don't want the test to depend on map key
// iteration order in the underlying JSON encoder.
func normalizeOptsForCompare(opts any) string {
	keys := []string{}
	// Drill down: opts is typically []any{map[string]any{...}}
	asArr, ok := opts.([]any)
	if !ok {
		return ""
	}
	if len(asArr) == 0 {
		return "[]"
	}
	first, ok := asArr[0].(map[string]any)
	if !ok {
		return ""
	}
	for k := range first {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(k + "=" + asBool(first[k]))
	}
	sb.WriteString("}")
	return sb.String()
}

func asBool(v any) string {
	if b, ok := v.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return "?"
}

// A2 regression — CLI compat dispatcher cardinality guard.
//
// The previous implementation looped over `for _, fileResult := range results`
// directly. If a Node-side worker crashed mid-batch and the response
// truncated to fewer results than input files, the loop processed the
// available results and silently swallowed the trailing files'
// diagnostics — visually identical to "those files had no issues" but
// actually a runner failure that should escalate to the runner-failure
// exit code.
//
// The fix surfaces a cardinality mismatch as a dispatcher failure
// (returns true from dispatchCompatBucket → counted into the
// program's CompatDispatchFailures), which the CLI surfaces as the
// runner-failure exit code (2).
func TestRunLinter_CompatDispatcher_CardinalityMismatch_MarksFailure(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	// Buggy dispatcher: returns one fewer result than requested,
	// mimicking a worker that crashed after processing 2 of 3 files.
	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		out := make([]CompatFileResult, 0, len(batch.Files)-1)
		for i, f := range batch.Files {
			if i == len(batch.Files)-1 {
				break // drop the last
			}
			out = append(out, CompatFileResult{FilePath: f.Path})
		}
		return out, nil
	})

	prevStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = prevStderr }()

	result, runErr := RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})
	if runErr != nil {
		t.Fatalf("RunLinter: %v", runErr)
	}

	_ = w.Close()
	stderrBuf := make([]byte, 1024)
	n, _ := r.Read(stderrBuf)
	stderrOut := string(stderrBuf[:n])

	if result.CompatDispatchErrors == 0 {
		t.Fatalf("expected CompatDispatchErrors > 0 from cardinality mismatch, got 0; stderr=%q", stderrOut)
	}
	if !strings.Contains(stderrOut, "compat dispatcher returned") {
		t.Errorf("expected stderr to mention the cardinality mismatch, got %q", stderrOut)
	}
}

// N17 regression: cardinality alone isn't enough — a buggy dispatcher
// returning the RIGHT number of results in the WRONG order would
// M2: a buggy dispatcher returning the right COUNT but with REORDERED
// results used to be flagged as a dispatcher fault. After the fix,
// reordering is allowed (downstream lookups are path-keyed, not index-
// keyed); the guard now enforces (presence + uniqueness) of every
// input path. So this test asserts the new contract: reorder DOES NOT
// cause a compat-failed mark.
func TestRunLinter_CompatDispatcher_ReorderedResults_StillAccepted(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
		"c.ts": "const c = 3;",
	})

	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		// Return reversed order — same set of paths, same count.
		out := make([]CompatFileResult, len(batch.Files))
		for i, f := range batch.Files {
			out[len(batch.Files)-1-i] = CompatFileResult{FilePath: f.Path}
		}
		return out, nil
	})

	result, runErr := RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})
	if runErr != nil {
		t.Fatalf("RunLinter: %v", runErr)
	}
	if result.CompatDispatchErrors != 0 {
		t.Errorf("reorder must not fail: got CompatDispatchErrors=%d", result.CompatDispatchErrors)
	}
}

// M2: a dispatcher returning DUPLICATE paths (same file twice) IS
// still a fault — silently dropping another file's diagnostics is
// data loss. Pin this so the relaxed-ordering fix doesn't accidentally
// also relax uniqueness.
func TestRunLinter_CompatDispatcher_DuplicatePathsMarkFailure(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		// Right count, but BOTH results reference the same file —
		// b.ts's slot got overwritten with a.ts. The unique-path
		// guard must trip this.
		return []CompatFileResult{
			{FilePath: batch.Files[0].Path},
			{FilePath: batch.Files[0].Path},
		}, nil
	})

	prevStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = prevStderr }()

	result, runErr := RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})
	if runErr != nil {
		t.Fatalf("RunLinter: %v", runErr)
	}

	_ = w.Close()
	stderrBuf := make([]byte, 1024)
	n, _ := r.Read(stderrBuf)
	stderrOut := string(stderrBuf[:n])

	if result.CompatDispatchErrors == 0 {
		t.Fatalf("expected CompatDispatchErrors > 0 from duplicate paths, got 0; stderr=%q", stderrOut)
	}
	if !strings.Contains(stderrOut, "duplicate") {
		t.Errorf("expected stderr to mention duplicate, got %q", stderrOut)
	}
}

// M2: a dispatcher returning an UNKNOWN path that doesn't even
// normalize-match an input file is a real fault — most likely a
// client-side bug. Pin the failure path so the M2 fallback doesn't
// accidentally accept arbitrary garbage.
func TestRunLinter_CompatDispatcher_UnknownPathMarksFailure(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
	})

	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	dispatcher := CompatBatchHandler(func(_ context.Context, _ CompatBatch) ([]CompatFileResult, error) {
		return []CompatFileResult{
			{FilePath: "/totally/unrelated.ts"},
		}, nil
	})

	prevStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = prevStderr }()

	result, runErr := RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
	})
	if runErr != nil {
		t.Fatalf("RunLinter: %v", runErr)
	}

	_ = w.Close()
	stderrBuf := make([]byte, 1024)
	n, _ := r.Read(stderrBuf)
	stderrOut := string(stderrBuf[:n])

	if result.CompatDispatchErrors == 0 {
		t.Fatalf("expected CompatDispatchErrors > 0 from unknown path, got 0; stderr=%q", stderrOut)
	}
	if !strings.Contains(stderrOut, "unknown path") {
		t.Errorf("expected stderr to mention unknown path, got %q", stderrOut)
	}
}

// #2 regression: worker-emitted positions past the Go-side text length
// (unsaved-buffer overlay / BOM / encoding skew) must be clamped to
// [0, len(text)] before the diagnostic (and fix) Range is built.
// Without the clamp, the formatter's
// scanner.GetECMALineAndUTF16CharacterOfPosition slices text[:pos] on a
// shorter text and panics, crashing the whole process.
func TestRunLinter_CompatOutOfRangePosition_Clamped(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})

	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/rule",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          "/c",
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	// Worker reports positions far past the file length, plus an
	// out-of-range fix range.
	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		results := make([]CompatFileResult, len(batch.Files))
		for i, f := range batch.Files {
			results[i] = CompatFileResult{
				FilePath: f.Path,
				Diagnostics: []CompatDiagnostic{{
					RuleName:  "fake-plugin/rule",
					MessageId: "x",
					Message:   "out of range",
					StartPos:  100,
					EndPos:    200,
					Fixes:     []CompatFix{{Range: [2]int{150, 300}, Text: "z"}},
				}},
			}
		}
		return results, nil
	})

	var got []rule.RuleDiagnostic
	var mu sync.Mutex
	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		CollectFixes:         true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			got = append(got, d)
			mu.Unlock()
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(got))
	}
	d := got[0]
	// Clamp target is len(sourceFile.Text()) — read it back from the
	// diagnostic's own SourceFile so a trailing newline / normalization
	// can't make the bound wrong.
	n := len(d.SourceFile.Text())
	if p := d.Range.Pos(); p < 0 || p > n {
		t.Errorf("diag Range.Pos()=%d not clamped to [0,%d]", p, n)
	}
	if e := d.Range.End(); e < 0 || e > n {
		t.Errorf("diag Range.End()=%d not clamped to [0,%d]", e, n)
	}
	if d.FixesPtr == nil || len(*d.FixesPtr) != 1 {
		t.Fatalf("expected 1 materialized fix, got %v", d.FixesPtr)
	}
	fr := (*d.FixesPtr)[0].Range
	if fr.Pos() < 0 || fr.Pos() > n || fr.End() < 0 || fr.End() > n {
		t.Errorf("fix Range [%d,%d] not clamped to [0,%d]", fr.Pos(), fr.End(), n)
	}
}

// #2 (defensive): a worker range with start > end — never produced by
// valid ESLint output, but a malformed worker could — must NOT yield an
// inverted core.TextRange, which would panic the formatter's
// text[start:end] slice. clampRange re-orders so End() >= Pos(). Reverting
// the `if end < start { end = start }` guard fails this test.
func TestRunLinter_CompatInvertedRange_Clamped(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const x = 1;",
	})
	getRules := func(_ *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/rule",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			ConfigKey:          "/c",
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}
	// In-bounds but INVERTED (start > end) on both the diagnostic and its fix.
	dispatcher := CompatBatchHandler(func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		results := make([]CompatFileResult, len(batch.Files))
		for i, f := range batch.Files {
			results[i] = CompatFileResult{
				FilePath: f.Path,
				Diagnostics: []CompatDiagnostic{{
					RuleName:  "fake-plugin/rule",
					MessageId: "x",
					Message:   "inverted",
					StartPos:  8,
					EndPos:    3,
					Fixes:     []CompatFix{{Range: [2]int{9, 4}, Text: "z"}},
				}},
			}
		}
		return results, nil
	})

	var got []rule.RuleDiagnostic
	var mu sync.Mutex
	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		CompatRuleDispatcher: dispatcher,
		CollectFixes:         true,
		OnDiagnostic: func(d rule.RuleDiagnostic) {
			mu.Lock()
			got = append(got, d)
			mu.Unlock()
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(got))
	}
	d := got[0]
	if d.Range.End() < d.Range.Pos() {
		t.Errorf("diag Range inverted: Pos()=%d > End()=%d", d.Range.Pos(), d.Range.End())
	}
	if d.FixesPtr == nil || len(*d.FixesPtr) != 1 {
		t.Fatalf("expected 1 fix, got %v", d.FixesPtr)
	}
	fr := (*d.FixesPtr)[0].Range
	if fr.End() < fr.Pos() {
		t.Errorf("fix Range inverted: Pos()=%d > End()=%d", fr.Pos(), fr.End())
	}
}

// #17: the path-presence guard accepts a result whose path is byte-divergent
// but tspath.NormalizePath-equal to an input (a client that lightly
// normalized paths), rewriting it back to the input bytes so downstream
// lookups (keyed by input bytes) still hit. Every other compat-validation
// test uses byte-equal paths, so this exercises the normalize-fallback
// branch specifically.
func TestValidateAndNormalizeCompatResults_NormalizeFallback(t *testing.T) {
	files := []CompatLintFile{{Path: "x/a.ts"}, {Path: "x/b.ts"}}
	// First result's path only matches AFTER normalization ("x/./a.ts" →
	// "x/a.ts"); the second is byte-equal.
	results := []CompatFileResult{
		{FilePath: "x/./a.ts"},
		{FilePath: "x/b.ts"},
	}
	if err := ValidateAndNormalizeCompatResults(results, files); err != nil {
		t.Fatalf("normalize-equal path must be accepted, got error: %v", err)
	}
	// Rewritten back to the exact input bytes for downstream path lookups.
	if results[0].FilePath != "x/a.ts" {
		t.Errorf("FilePath should be rewritten to input %q, got %q", "x/a.ts", results[0].FilePath)
	}
	if results[1].FilePath != "x/b.ts" {
		t.Errorf("byte-equal FilePath should be unchanged, got %q", results[1].FilePath)
	}
}
