package linter

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/languages/js/estree"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func pluginRule(name string, opts []any, sev rule.DiagnosticSeverity) rule.ConfiguredRule {
	return rule.ConfiguredRule{Name: name, Options: opts, Severity: sev, IsEslintPluginRule: true}
}

func TestDispatchEslintPluginRulesAsyncPublishesStructuredOutcome(t *testing.T) {
	dispatchError := errors.New("transport closed")
	outcomeCh := testDispatchEslintPluginRulesAsync(
		context.Background(),
		EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return nil, dispatchError
		}),
		[]EslintPluginFileInput{{
			Path: "/repo/a.ts",
			Rules: []rule.ConfiguredRule{{
				Name:               "external/rule",
				Severity:           rule.SeverityError,
				IsEslintPluginRule: true,
			}},
		}},
		false,
		SuggestionsModeOff,
		nil,
	)

	var outcome EslintPluginDispatchOutcome
	select {
	case outcome = <-outcomeCh:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for async plugin outcome")
	}
	if !errors.Is(outcome.DispatchError, dispatchError) {
		t.Fatalf("published error = %v, want %v", outcome.DispatchError, dispatchError)
	}
	if len(outcome.Diagnostics) != 1 || outcome.Diagnostics[0].RuleName != "rslint/plugin-lint-error" {
		t.Fatalf("diagnostics = %v, want dispatch failure diagnostic", outcome.Diagnostics)
	}
}

func TestBuildEslintPluginFileInput_EffectiveLanguageOptions(t *testing.T) {
	raw := map[string]any{
		"parserOptions": map[string]any{"ecmaFeatures": map[string]any{"jsx": true}},
	}
	rules := []rule.ConfiguredRule{{
		Name:               "plugin/rule",
		IsEslintPluginRule: true,
		Environment:        &rule.RuleEnvironment{},
	}}

	input := buildEslintPluginFileInput("/repo/file.cjs", "/repo", prepareRuleSet(rules, true), raw, nil, nil, nil)
	if got := input.LanguageOptions["sourceType"]; got != "commonjs" {
		t.Errorf(".cjs sourceType = %v, want commonjs", got)
	}
	if got := input.LanguageOptions["ecmaVersion"]; got != rule.LatestECMAScriptVersion {
		t.Errorf("default ecmaVersion = %v, want %d", got, rule.LatestECMAScriptVersion)
	}
	if _, ok := input.LanguageOptions["parserOptions"]; !ok {
		t.Error("raw parserOptions were not preserved")
	}
	if _, ok := raw["sourceType"]; ok {
		t.Error("building plugin input mutated the shared raw languageOptions map")
	}
	if _, ok := raw["ecmaVersion"]; ok {
		t.Error("building plugin input mutated the shared raw languageOptions map")
	}

	rules[0].Environment = &rule.RuleEnvironment{LanguageOptions: rule.LanguageOptions{
		ECMAVersion: 2020,
		SourceType:  "script",
	}}
	input = buildEslintPluginFileInput("/repo/file.unknown", "/repo", prepareRuleSet(rules, true), raw, nil, nil, nil)
	if got := input.LanguageOptions["sourceType"]; got != "script" {
		t.Errorf("authored sourceType = %v, want script", got)
	}
	if got := input.LanguageOptions["ecmaVersion"]; got != 2020 {
		t.Errorf("authored ecmaVersion = %v, want 2020", got)
	}
}

func TestBuildEslintPluginFileInputsSharedRulesKeepFileDefaults(t *testing.T) {
	compilerProgram, paths := createTestProgramWithFilesAndCompilerOptions(t, map[string]string{
		"one.cjs": "exports.value = 1;",
		"two.mjs": "export const value = 1;",
	}, `{"allowJs":true}`)
	program := lintprogram.NewFromCompiler(compilerProgram)
	rules := []rule.ConfiguredRule{pluginRule("external/rule", nil, rule.SeverityWarning)}
	plan := mustPrepareLintPlan(t, PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{program},
		TargetsByProgram: [][]string{{paths["one.cjs"], paths["two.mjs"]}},
		GetRulesForFile:  func(*ast.SourceFile) []rule.ConfiguredRule { return rules },
	})
	raw := map[string]any{"parserOptions": map[string]any{"ecmaFeatures": map[string]any{"jsx": true}}}
	inputs := BuildEslintPluginFileInputs(plan, func(path string) EslintPluginFileConfig {
		return EslintPluginFileConfig{ConfigKey: path, LanguageOptions: raw}
	})
	if len(inputs) != 2 || &inputs[0].Rules[0] != &inputs[1].Rules[0] {
		t.Fatal("files did not share their prepared plugin projection")
	}
	if inputs[0].LanguageOptions["sourceType"] != "commonjs" || inputs[1].LanguageOptions["sourceType"] != "module" ||
		inputs[0].ConfigKey != paths["one.cjs"] || inputs[1].ConfigKey != paths["two.mjs"] {
		t.Fatal("shared rules conflated file defaults or plugin routing")
	}
	inputs[0].LanguageOptions["sourceType"] = "script"
	if inputs[1].LanguageOptions["sourceType"] != "module" || raw["sourceType"] != nil {
		t.Fatal("a file's effective language options mutated another file or the config input")
	}
}

func TestDispatchEslintPlugin_EmptyShortCircuit(t *testing.T) {
	called := 0
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		called++
		return &EslintPluginLintResult{}, nil
	})
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, nil, false, "off", nil, func(rule.RuleDiagnostic) {}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 0 {
		t.Errorf("expected dispatcher not called for empty files, got %d", called)
	}
}

func TestDispatchEslintPlugin_GroupsBySignature(t *testing.T) {
	// A,B share (configKey + rule + options) → one batch; C differs in
	// options → another; D differs in configKey → another. 3 batches total.
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
		{Path: "/b.ts", ConfigKey: "/cfg", Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
		{Path: "/c.ts", ConfigKey: "/cfg", Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o2"}, rule.SeverityError)}},
		{Path: "/d.ts", ConfigKey: "/other", Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
	}
	var (
		mu      sync.Mutex
		batches []EslintPluginLintRequest
	)
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		// dispatch is now invoked concurrently (one goroutine per batch), so the
		// shared slice needs a lock.
		mu.Lock()
		batches = append(batches, req)
		mu.Unlock()
		results := make([]EslintPluginFileResult, 0, len(req.Files))
		for _, f := range req.Files {
			results = append(results, EslintPluginFileResult{FilePath: f.Path})
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(rule.RuleDiagnostic) {}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	twoFileBatches := 0
	for _, b := range batches {
		if len(b.Files) == 2 {
			twoFileBatches++
			if b.Files[0].ConfigKey != "/cfg" {
				t.Errorf("expected the 2-file batch under /cfg, got %q", b.Files[0].ConfigKey)
			}
		}
	}
	if twoFileBatches != 1 {
		t.Errorf("expected exactly one 2-file batch (A+B), got %d", twoFileBatches)
	}
}

func TestDispatchEslintPlugin_SeverityDoesNotFragmentBatches(t *testing.T) {
	// Same configKey + rule + options, differing ONLY in severity. Severity is
	// not part of the batch key (it never goes on the wire, reattached per-file),
	// so these must share exactly ONE batch — yet each file keeps its severity.
	ta, tb := "a\n", "b\n"
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", Text: &ta, Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
		{Path: "/b.ts", ConfigKey: "/cfg", Text: &tb, Rules: []rule.ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityWarning)}},
	}
	var (
		mu         sync.Mutex
		batchCount int
	)
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		mu.Lock()
		batchCount++
		mu.Unlock()
		results := make([]EslintPluginFileResult, 0, len(req.Files))
		for _, f := range req.Files {
			results = append(results, EslintPluginFileResult{
				FilePath:    f.Path,
				Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
			})
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	sevByPath := map[string]rule.DiagnosticSeverity{}
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		sevByPath[d.FilePath] = d.Severity
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batchCount != 1 {
		t.Fatalf("severity must not fragment batches: expected 1 batch, got %d", batchCount)
	}
	if sevByPath["/a.ts"] != rule.SeverityError {
		t.Errorf("/a.ts severity = %v, want Error", sevByPath["/a.ts"])
	}
	if sevByPath["/b.ts"] != rule.SeverityWarning {
		t.Errorf("/b.ts severity = %v, want Warning", sevByPath["/b.ts"])
	}
}

func TestDispatchEslintPlugin_ConcurrentBatchesAllEmitted(t *testing.T) {
	// Many distinct batches (each a different option → its own batch) are
	// dispatched concurrently. Every batch's diagnostic must be emitted exactly
	// once. The `seen` map is written from onDiagnostic WITHOUT a lock on purpose:
	// onDiagnostic must stay single-threaded (emitted after Wait), so -race would
	// catch a regression that emits concurrently.
	const n = 20
	files := make([]EslintPluginFileInput, n)
	texts := make([]string, n)
	for i := range files {
		texts[i] = "x\n"
		files[i] = EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &texts[i],
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)}, // distinct options → distinct batch
		}
	}
	var (
		mu    sync.Mutex
		calls int
	)
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		mu.Lock()
		calls++
		mu.Unlock()
		results := make([]EslintPluginFileResult, 0, len(req.Files))
		for _, f := range req.Files {
			results = append(results, EslintPluginFileResult{
				FilePath:    f.Path,
				Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
			})
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	seen := map[string]int{}
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		seen[d.FilePath]++
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != n {
		t.Fatalf("expected %d batches dispatched, got %d", n, calls)
	}
	if len(seen) != n {
		t.Fatalf("expected diagnostics for all %d files, got %d distinct", n, len(seen))
	}
	for p, c := range seen {
		if c != 1 {
			t.Errorf("%s emitted %d times, want exactly 1", p, c)
		}
	}
}

func TestDispatchEslintPlugin_BatchFailureDoesNotAbortOthers(t *testing.T) {
	// One batch's dispatch fails; the others must still run and emit their
	// diagnostics (concurrent dispatch no longer aborts on the first failure),
	// and the error must still surface to the caller.
	const n = 10
	const failOn = "/f3.ts"
	files := make([]EslintPluginFileInput, n)
	texts := make([]string, n)
	for i := range files {
		texts[i] = "x\n"
		files[i] = EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &texts[i],
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)},
		}
	}
	wantErr := errors.New("boom")
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if req.Files[0].Path == failOn {
			return nil, wantErr
		}
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    req.Files[0].Path,
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
		}}}, nil
	})
	seen := map[string]int{}
	err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		seen[d.FilePath]++
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected the failing batch's error to surface, got %v", err)
	}
	if len(seen) != n-1 {
		t.Errorf("expected all %d non-failing files emitted, got %d", n-1, len(seen))
	}
	if _, ok := seen[failOn]; ok {
		t.Errorf("failing file %s must have no diagnostic", failOn)
	}
}

func TestDispatchEslintPlugin_FirstBatchErrorWins(t *testing.T) {
	// Multiple batches fail; the returned error must be the first in batch order
	// (groupEslintPluginFiles keeps first-seen order), matching the previous
	// serial "stop at the first failure" error semantics regardless of which
	// batch's goroutine finished first.
	errEarly := errors.New("early")
	errLate := errors.New("late")
	texts := make([]string, 6)
	files := make([]EslintPluginFileInput, 6)
	for i := range files {
		texts[i] = "x\n"
		files[i] = EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &texts[i],
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)}, // distinct options → own batch, order f0..f5
		}
	}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		switch req.Files[0].Path {
		case "/f2.ts":
			return nil, errEarly
		case "/f4.ts":
			return nil, errLate
		default:
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: req.Files[0].Path}}}, nil
		}
	})
	err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(rule.RuleDiagnostic) {})
	if !errors.Is(err, errEarly) {
		t.Fatalf("expected the first-in-batch-order error (f2's), got %v", err)
	}
}

func TestDispatchEslintPlugin_SeverityReattachAndClamp(t *testing.T) {
	text := "abc\n" // len 4 bytes
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityWarning)}},
	}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{
				{RuleName: "uc/x", Message: "in-bounds", StartPos: 1, EndPos: 3},
				{RuleName: "uc/x", Message: "out-of-bounds", StartPos: 10, EndPos: 20},
				{RuleName: "uc/x", Message: "inverted", StartPos: 3, EndPos: 1},
			},
		}}}, nil
	})
	var diags []rule.RuleDiagnostic
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		diags = append(diags, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 3 {
		t.Fatalf("expected 3 diagnostics, got %d", len(diags))
	}
	for _, d := range diags {
		if d.Severity != rule.SeverityWarning {
			t.Errorf("severity should be reattached as Warning, got %v", d.Severity)
		}
		if d.FilePath != "/a.ts" {
			t.Errorf("FilePath should be /a.ts, got %q", d.FilePath)
		}
		if d.SourceFile == nil || d.SourceFile.Text() != text {
			t.Errorf("SourceFile.Text() should equal the file text")
		}
	}
	if diags[0].Range.Pos() != 1 || diags[0].Range.End() != 3 {
		t.Errorf("in-bounds: want [1,3], got [%d,%d]", diags[0].Range.Pos(), diags[0].Range.End())
	}
	if diags[1].Range.Pos() != 4 || diags[1].Range.End() != 4 {
		t.Errorf("out-of-bounds: want clamped [4,4], got [%d,%d]", diags[1].Range.Pos(), diags[1].Range.End())
	}
	if diags[2].Range.Pos() != 3 || diags[2].Range.End() != 3 {
		t.Errorf("inverted: want normalized [3,3], got [%d,%d]", diags[2].Range.Pos(), diags[2].Range.End())
	}
}

func TestDispatchEslintPlugin_ParseErrorClasses(t *testing.T) {
	text := "x"
	mkFiles := func() []EslintPluginFileInput {
		return []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	}
	run := func(fr EslintPluginFileResult) []rule.RuleDiagnostic {
		dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{fr}}, nil
		})
		var diags []rule.RuleDiagnostic
		if err := testDispatchEslintPluginRules(context.Background(), dispatch, mkFiles(), false, "off", nil, func(d rule.RuleDiagnostic) {
			diags = append(diags, d)
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return diags
	}

	// Every abnormal parseError must surface as one visible error diagnostic.
	// oxc recovers from syntax errors (no parseError ever), so each of these is
	// a genuine worker failure that would otherwise be a silent false-green.
	abnormal := []string{
		"parse: source too large (12345 bytes)",
		"normalize: cannot normalize AST",
		"worker fs.readFile failed for /a.ts: ENOENT",
		`worker: configKey "/x" not declared in workerData.configs[]; known: []`,
		"worker exception: boom",
		"worker not initialized",
		"worker_crashed: boom",
		"task_timeout",
		"pool_degraded",
		"postMessage_failed: detached",
	}
	for _, pe := range abnormal {
		d := run(EslintPluginFileResult{FilePath: "/a.ts", ParseError: pe})
		if len(d) != 1 {
			t.Errorf("parseError %q should surface 1 diagnostic, got %d", pe, len(d))
			continue
		}
		if d[0].RuleName != "rslint/plugin-lint-error" {
			t.Errorf("parseError %q: ruleName = %q, want rslint/plugin-lint-error", pe, d[0].RuleName)
		}
		if d[0].Severity != rule.SeverityError {
			t.Errorf("parseError %q: severity = %v, want SeverityError", pe, d[0].Severity)
		}
		if !strings.Contains(d[0].Message.Description, pe) {
			t.Errorf("parseError %q: message %q should contain the parseError text", pe, d[0].Message.Description)
		}
	}

	// "shutdown" is the sole benign sentinel (graceful pool teardown racing a
	// task) → 0 diagnostics.
	if d := run(EslintPluginFileResult{FilePath: "/a.ts", ParseError: "shutdown"}); len(d) != 0 {
		t.Errorf("shutdown sentinel should yield 0 diagnostics, got %d", len(d))
	}

	// ruleErrors → one visible diagnostic per error, rebuilt precisely.
	d := run(EslintPluginFileResult{FilePath: "/a.ts", RuleErrors: []EslintPluginRuleError{{Rule: "uc/x", Message: "threw"}}})
	if len(d) != 1 {
		t.Fatalf("ruleErrors should yield 1 diagnostic, got %d", len(d))
	}
	if d[0].RuleName != "uc/x" {
		t.Errorf("ruleError ruleName = %q, want uc/x", d[0].RuleName)
	}
	if d[0].Message.Description != "rule error: threw" {
		t.Errorf("ruleError message = %q, want \"rule error: threw\"", d[0].Message.Description)
	}
	if d[0].Severity != rule.SeverityError {
		t.Errorf("ruleError severity = %v, want SeverityError", d[0].Severity)
	}
	if d[0].Range.Pos() != 0 || d[0].Range.End() != 0 {
		t.Errorf("ruleError range = [%d,%d], want [0,0]", d[0].Range.Pos(), d[0].Range.End())
	}
}

func TestDispatchEslintPlugin_Cancelled(t *testing.T) {
	text := "x"
	files := []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: "/a.ts", Cancelled: true}}}, nil
	})
	err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(rule.RuleDiagnostic) {})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDispatchEslintPlugin_MissingResultNoPanic(t *testing.T) {
	text := "x"
	files := []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: nil}, nil // no result for /a.ts
	})
	var diags []rule.RuleDiagnostic
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		diags = append(diags, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Errorf("missing result should contribute 0 diagnostics, got %d", len(diags))
	}
}

func TestTextSourceFile_LineMap(t *testing.T) {
	text := "a\nbb\nccc"
	tsf := newTextSourceFile(text)
	if tsf.Text() != text {
		t.Errorf("Text() mismatch")
	}
	got := tsf.ECMALineMap()
	want := core.ComputeECMALineStarts(text)
	if len(got) != len(want) {
		t.Fatalf("line map length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line map[%d]: got %d want %d", i, got[i], want[i])
		}
	}
}

// TestEslintPluginSourceFile pins the rebuild-frame selection: CLI reuses the
// caller-provided SourceFile verbatim (the ts-go frame the native pass loaded —
// no re-read); LSP wraps the overlay Text, BOM-stripped to match the worker's
// req.text frame; neither -> ok=false.
func TestEslintPluginSourceFile(t *testing.T) {
	const code = "const x = 1;"

	// (a) CLI: the provided SourceFile is returned verbatim (reused, not copied).
	sf := pluginSyntax("/a.ts", code)
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{SourceFile: sf}); !ok || got != sf {
		t.Errorf("SourceFile branch must reuse the provided frame verbatim")
	}
	// (b) LSP overlay: the frame drops a leading BOM, so its offsets are the
	// worker's and the native frame's.
	withBOM := "\ufeff" + code
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{Text: &withBOM}); !ok || got.Text() != code {
		t.Errorf("overlay frame must drop the BOM: text = %q, want %q", got.Text(), code)
	}
	plain := code
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{Text: &plain}); !ok || got.Text() != code {
		t.Errorf("overlay no-BOM: text = %q, want %q", got.Text(), code)
	}
	// (c) neither SourceFile nor Text -> ok=false.
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{Path: "/a.ts"}); ok || got != nil {
		t.Errorf("no frame: got (%v,%v), want (nil,false)", got, ok)
	}
}

// TestDispatchEslintPlugin_FrameReuseOverlayAndNoFrame verifies the end-to-end
// rebuild over the three frame sources: (a) CLI reuses the native SourceFile,
// honoring byte offsets across multi-byte chars; (b) LSP overlay strips a BOM;
// (c) no frame surfaces an error instead of collapsing offsets to 0.
func TestDispatchEslintPlugin_FrameReuseOverlayAndNoFrame(t *testing.T) {
	// (a) CLI SourceFile-reuse with a multi-byte frame: the two CJK chars occupy
	// bytes 6..12 (two 3-byte runes), so a [6,12] worker offset must land on them.
	sf := pluginSyntax("/a.ts", "const 变量 = 1;")
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", SourceFile: sf, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 6, EndPos: 12}},
		}}}, nil
	})
	var diags []rule.RuleDiagnostic
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		diags = append(diags, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(diags))
	}
	if diags[0].SourceFile != sf {
		t.Errorf("CLI must REUSE the provided SourceFile frame, not a copy")
	}
	if diags[0].Range.Pos() != 6 || diags[0].Range.End() != 12 {
		t.Errorf("range over reused frame: want [6,12], got [%d,%d]", diags[0].Range.Pos(), diags[0].Range.End())
	}

	// (b) LSP overlay with a leading BOM. The frame drops the mark, matching the
	// worker, the native diagnostics and source.fixAll: an offset never counts
	// a byte order mark. The worker's [6,7] (the 'x') therefore lands as it
	// stands, and the fix splices the 'x' rather than the mark.
	const code = "const x = 1;"
	bomText := "\ufeff" + code
	files2 := []EslintPluginFileInput{
		{Path: "/b.ts", ConfigKey: "/cfg", Text: &bomText, Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch2 := EslintPluginDispatchFunc(func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/b.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 6, EndPos: 7,
				Fixes: []EslintPluginFix{{Range: [2]int{6, 7}, Text: "y"}}}},
		}}}, nil
	})
	var diags2 []rule.RuleDiagnostic
	if err := testDispatchEslintPluginRules(context.Background(), dispatch2, files2, true, "eager", nil, func(d rule.RuleDiagnostic) {
		diags2 = append(diags2, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags2) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(diags2))
	}
	if diags2[0].SourceFile.Text() != code {
		t.Errorf("overlay frame must drop the BOM: got %q, want %q", diags2[0].SourceFile.Text(), code)
	}
	if diags2[0].Range.Pos() != 6 || diags2[0].Range.End() != 7 {
		t.Errorf("overlay range: worker [6,7] \u2192 want [6,7], got [%d,%d]", diags2[0].Range.Pos(), diags2[0].Range.End())
	}
	fx := diags2[0].Fixes()
	if len(fx) != 1 || fx[0].Range.Pos() != 6 || fx[0].Range.End() != 7 || fx[0].Text != "y" {
		t.Errorf("overlay fix: worker [6,7] \u2192 want [6,7]='y', got %+v", fx)
	}
	// End-to-end: applying the fix to the with-BOM overlay must replace the 'x'
	// and PRESERVE the BOM \u2014 no off-by-3 corruption.
	if fixed, _, ok := ApplyRuleFixes(bomText, diags2); !ok || fixed != "\ufeff"+"const y = 1;" {
		t.Errorf("fixAll on a BOM file must preserve the BOM + replace x\u2192y, got %q", fixed)
	}

	// (c) Neither SourceFile nor Text -> surface a 'no source frame' error.
	files3 := []EslintPluginFileInput{
		{Path: "/c.ts", ConfigKey: "/cfg", Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch3 := EslintPluginDispatchFunc(func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    "/c.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
		}}}, nil
	})
	var diags3 []rule.RuleDiagnostic
	if err := testDispatchEslintPluginRules(context.Background(), dispatch3, files3, false, "off", nil, func(d rule.RuleDiagnostic) {
		diags3 = append(diags3, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags3) != 1 || diags3[0].RuleName != "rslint/plugin-lint-error" {
		t.Fatalf("no-frame should surface 1 plugin-lint-error, got %+v", diags3)
	}
	if !strings.Contains(diags3[0].Message.Description, "no source frame") {
		t.Errorf("message should mention no source frame, got %q", diags3[0].Message.Description)
	}
}

// TestEslintPluginWire_RoundTrip pins the wire JSON keys against the Node
// worker's plugin-lint-protocol.ts. A silent key drift (e.g. startPos →
// start) would drop data on the wire without a compile error, so assert every
// field explicitly: the request keys via marshal, the result fields via decode
// of a Node-shaped payload.
func TestEslintPluginWire_FrameAttachmentsPreserveBytesAndMetadata(t *testing.T) {
	frames := [][]byte{{0, 255, 1}, {}, []byte("\ufeffconst café = '😀';")}
	files := make([]EslintPluginLintFile, len(frames))
	for i, frame := range frames {
		files[i] = EslintPluginLintFile{Path: strconv.Itoa(i), Frame: frame, ConfigKey: "config", Settings: map[string]any{"x": true}}
	}
	req := EslintPluginLintRequest{Files: files, Generation: "generation", CollectFixes: true}
	wire, attachments := req.WithFrameAttachments()
	if !reflect.DeepEqual(attachments, frames) {
		t.Fatal("frame bytes changed")
	}
	for i, file := range wire.Files {
		if file.Frame != nil || file.FrameAttachment == nil || int(*file.FrameAttachment) != i || !reflect.DeepEqual(file.Settings, files[i].Settings) {
			t.Fatal("frame attachment projection")
		}
		if req.Files[i].FrameAttachment != nil {
			t.Fatal("request mutated")
		}
	}
	wire.Files = req.Files
	if !reflect.DeepEqual(wire, req) {
		t.Fatal("request metadata changed")
	}
}

func TestEslintPluginWire_RoundTrip(t *testing.T) {
	text := "const x = 1;"
	req := EslintPluginLintRequest{
		Files: []EslintPluginLintFile{{
			Path:            "/proj/a.ts",
			Frame:           []byte(text),
			ConfigKey:       "/proj",
			LanguageOptions: map[string]any{"sourceType": "module"},
			Settings:        map[string]any{"foo": "bar"},
		}},
		Rules:           map[string]EslintPluginRuleConfig{"p/r": {Options: []any{"opt"}}},
		CollectFixes:    true,
		SuggestionsMode: "eager",
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}
	for _, k := range []string{"files", "rules", "collectFixes", "suggestionsMode"} {
		if _, ok := got[k]; !ok {
			t.Errorf("request JSON missing top-level key %q", k)
		}
	}
	if _, ok := got["fix"]; ok {
		t.Error(`request JSON unexpectedly contains obsolete key "fix"`)
	}
	f0 := got["files"].([]any)[0].(map[string]any)
	for _, k := range []string{"path", "frame", "configKey", "languageOptions", "settings"} {
		if _, ok := f0[k]; !ok {
			t.Errorf("request file JSON missing key %q", k)
		}
	}
	if _, ok := got["rules"].(map[string]any)["p/r"].(map[string]any)["options"]; !ok {
		t.Error(`request rule JSON missing key "options"`)
	}

	// Decode a Node-shaped result payload and assert each wire field maps to
	// its Go field (filePath/startPos/endPos/fixes.range/suggestions/ruleErrors).
	const resJSON = `{"results":[{"filePath":"/proj/a.ts","cancelled":false,` +
		`"diagnostics":[{"ruleName":"p/r","messageId":"mid","message":"m",` +
		`"startPos":6,"endPos":11,"fixes":[{"range":[6,11],"text":"y"}],` +
		`"suggestions":[{"messageId":"s","desc":"d","fixes":[{"range":[0,1],"text":"z"}]}]}],` +
		`"ruleErrors":[{"rule":"p/r","message":"boom"}]}]}`
	var res EslintPluginLintResult
	if err := json.Unmarshal([]byte(resJSON), &res); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("want 1 result, got %d", len(res.Results))
	}
	fr := res.Results[0]
	if fr.FilePath != "/proj/a.ts" || fr.Cancelled {
		t.Errorf("file result decode wrong: %+v", fr)
	}
	if len(fr.Diagnostics) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(fr.Diagnostics))
	}
	d := fr.Diagnostics[0]
	if d.RuleName != "p/r" || d.MessageId != "mid" || d.Message != "m" || d.StartPos != 6 || d.EndPos != 11 {
		t.Errorf("diagnostic decode wrong: %+v", d)
	}
	if len(d.Fixes) != 1 || d.Fixes[0].Range != [2]int{6, 11} || d.Fixes[0].Text != "y" {
		t.Errorf("fix decode wrong: %+v", d.Fixes)
	}
	if len(d.Suggestions) != 1 || d.Suggestions[0].MessageId != "s" || d.Suggestions[0].Desc != "d" ||
		len(d.Suggestions[0].Fixes) != 1 || d.Suggestions[0].Fixes[0].Range != [2]int{0, 1} {
		t.Errorf("suggestion decode wrong: %+v", d.Suggestions)
	}
	if len(fr.RuleErrors) != 1 || fr.RuleErrors[0].Rule != "p/r" || fr.RuleErrors[0].Message != "boom" {
		t.Errorf("ruleErrors decode wrong: %+v", fr.RuleErrors)
	}
}

// TestDispatchEslintPlugin_MultipleRulesPerFile pins the normal object-form
// usage: one file enables SEVERAL rules of one plugin. buildEslintPluginRequest
// must carry ALL of the file's rules in the request rules map, and severity must
// reattach per rule (each diagnostic keeps its own rule's level).
func TestDispatchEslintPlugin_MultipleRulesPerFile(t *testing.T) {
	text := "x\n"
	files := []EslintPluginFileInput{{
		Path: "/a.ts", ConfigKey: "/cfg", Text: &text,
		Rules: []rule.ConfiguredRule{
			pluginRule("uc/a", []any{"o"}, rule.SeverityError),
			pluginRule("uc/b", nil, rule.SeverityWarning),
		},
	}}
	var captured EslintPluginLintRequest
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		captured = req
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{
				{RuleName: "uc/a", Message: "ma", StartPos: 0, EndPos: 1},
				{RuleName: "uc/b", Message: "mb", StartPos: 0, EndPos: 1},
			},
		}}}, nil
	})
	sevByRule := map[string]rule.DiagnosticSeverity{}
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		sevByRule[d.RuleName] = d.Severity
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.Rules) != 2 {
		t.Fatalf("request rules map must carry both rules, got %d: %v", len(captured.Rules), captured.Rules)
	}
	if _, ok := captured.Rules["uc/a"]; !ok {
		t.Error("uc/a missing from the request rules map")
	}
	if _, ok := captured.Rules["uc/b"]; !ok {
		t.Error("uc/b missing from the request rules map")
	}
	if sevByRule["uc/a"] != rule.SeverityError {
		t.Errorf("uc/a severity = %v, want Error", sevByRule["uc/a"])
	}
	if sevByRule["uc/b"] != rule.SeverityWarning {
		t.Errorf("uc/b severity = %v, want Warning", sevByRule["uc/b"])
	}
}

// TestDispatchEslintPlugin_DeadlineExceededSurfaces pins that when a batch's
// dispatch fails with context.DeadlineExceeded (the LSP fixAll timeout firing),
// DispatchEslintPluginRules surfaces an error for which errors.Is(err,
// context.DeadlineExceeded) holds — so an integration's timeout policy can
// recognize it — while the other (completed) batches still emit their
// diagnostics. (No mutex on the emitted map on purpose: onDiagnostic stays
// single-threaded, so -race catches a regression that emits concurrently.)
func TestDispatchEslintPlugin_DeadlineExceededSurfaces(t *testing.T) {
	const n = 6
	const failOn = "/f3.ts"
	files := make([]EslintPluginFileInput, n)
	texts := make([]string, n)
	for i := range files {
		texts[i] = "x\n"
		files[i] = EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &texts[i],
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)}, // distinct option → distinct batch
		}
	}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if req.Files[0].Path == failOn {
			return nil, context.DeadlineExceeded
		}
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    req.Files[0].Path,
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
		}}}, nil
	})
	emitted := map[string]bool{}
	err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		emitted[d.FilePath] = true
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected the surfaced error to wrap context.DeadlineExceeded, got %v", err)
	}
	if len(emitted) != n-1 {
		t.Errorf("expected the %d non-failing batches to still emit, got %d", n-1, len(emitted))
	}
	if emitted[failOn] {
		t.Errorf("the timed-out batch (%s) must not emit a diagnostic", failOn)
	}
}

// TestDispatchEslintPlugin_FixAndSuggestionsRebuilt pins the full
// rebuild+clamp+emit chain for a plugin rule that returns BOTH a fix and a
// suggestion. Elsewhere only the raw json.Unmarshal touches these fields; here
// the emitted RuleDiagnostic must carry the rebuilt fix and the rebuilt
// suggestion (message + range-clamped fix).
func TestDispatchEslintPlugin_FixAndSuggestionsRebuilt(t *testing.T) {
	text := "const x = 1;\n"
	files := []EslintPluginFileInput{{
		Path: "/a.ts", ConfigKey: "/cfg", Text: &text,
		Rules: []rule.ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)},
	}}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{{
				RuleName: "uc/x", Message: "m", StartPos: 6, EndPos: 7,
				Fixes:       []EslintPluginFix{{Range: [2]int{6, 7}, Text: "y"}},
				Suggestions: []EslintPluginSuggestion{{MessageId: "sid", Desc: "use z", Fixes: []EslintPluginFix{{Range: [2]int{6, 7}, Text: "z"}}}},
			}},
		}}}, nil
	})
	var got rule.RuleDiagnostic
	gotN := 0
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, true, "eager", nil, func(d rule.RuleDiagnostic) {
		got = d
		gotN++
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotN != 1 {
		t.Fatalf("expected exactly 1 diagnostic, got %d", gotN)
	}
	fixes := got.Fixes()
	if len(fixes) != 1 || fixes[0].Text != "y" || fixes[0].Range.Pos() != 6 || fixes[0].Range.End() != 7 {
		t.Errorf("fix not rebuilt correctly: %+v", fixes)
	}
	if got.Suggestions == nil || len(*got.Suggestions) != 1 {
		t.Fatalf("expected exactly 1 rebuilt suggestion, got %v", got.Suggestions)
	}
	sg := (*got.Suggestions)[0]
	if sg.Message.Id != "sid" || sg.Message.Description != "use z" {
		t.Errorf("suggestion message wrong: %+v", sg.Message)
	}
	sgFixes := sg.Fixes()
	if len(sgFixes) != 1 || sgFixes[0].Text != "z" || sgFixes[0].Range.Pos() != 6 || sgFixes[0].Range.End() != 7 {
		t.Errorf("suggestion fix not rebuilt correctly: %+v", sgFixes)
	}
}

// TestDispatchEslintPlugin_EmitsInBatchOrderDespiteOutOfOrderCompletion pins the
// #1050 determinism invariant: batches dispatch concurrently but diagnostics are
// EMITTED in batch (input) order, keeping CLI output / --fix stable run to run.
// The goroutines here finish in REVERSE order (batch 0 sleeps longest); emission
// must still be f0,f1,f2,f3.
func TestDispatchEslintPlugin_EmitsInBatchOrderDespiteOutOfOrderCompletion(t *testing.T) {
	const n = 4
	files := make([]EslintPluginFileInput, n)
	texts := make([]string, n)
	for i := range files {
		texts[i] = "x\n"
		files[i] = EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &texts[i],
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)}, // distinct option → own batch
		}
	}
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		var idx int
		_, _ = fmt.Sscanf(req.Files[0].Path, "/f%d.ts", &idx)
		// Earlier batches sleep longer, so they complete LAST.
		time.Sleep(time.Duration(n-idx) * 15 * time.Millisecond)
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    req.Files[0].Path,
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
		}}}, nil
	})
	var order []string
	if err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {
		order = append(order, d.FilePath)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"/f0.ts", "/f1.ts", "/f2.ts", "/f3.ts"}
	if !reflect.DeepEqual(order, want) {
		t.Errorf("emission order = %v, want batch order %v", order, want)
	}
}

// TestDispatchEslintPlugin_RealErrorOutranksCanceled pins that with concurrent
// batches, an earlier batch's context.Canceled (a superseded file) does not mask
// a later batch's genuine transport error — the real error surfaces.
func TestDispatchEslintPlugin_RealErrorOutranksCanceled(t *testing.T) {
	mkFile := func(i int) EslintPluginFileInput {
		txt := "x\n"
		return EslintPluginFileInput{
			Path:      fmt.Sprintf("/f%d.ts", i),
			ConfigKey: "/cfg",
			Text:      &txt,
			Rules:     []rule.ConfiguredRule{pluginRule("uc/x", []any{i}, rule.SeverityError)},
		}
	}
	files := []EslintPluginFileInput{mkFile(0), mkFile(1)}
	boom := errors.New("transport boom")
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if req.Files[0].Path == "/f0.ts" {
			// earlier batch: cooperative cancel (worker dropped a superseded file)
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: "/f0.ts", Cancelled: true}}}, nil
		}
		return nil, boom // later batch: a real transport failure
	})
	err := testDispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", nil, func(d rule.RuleDiagnostic) {})
	if !errors.Is(err, boom) {
		t.Fatalf("expected the real transport error to surface, got %v", err)
	}
}

func TestBuildEslintPluginFileInputsUsesPreparedPlan(t *testing.T) {
	compilerProgram, paths := createTestProgramWithFiles(t, map[string]string{
		"native.ts": "export const native = 1;\n",
		"plugin.ts": "export const plugin = 1;\n",
	})
	nativeFile := compilerProgram.GetSourceFile(paths["native.ts"])
	pluginFile := compilerProgram.GetSourceFile(paths["plugin.ts"])
	if nativeFile == nil || pluginFile == nil {
		t.Fatalf("prepared plan fixtures are missing: native=%v plugin=%v", nativeFile != nil, pluginFile != nil)
	}
	nativeRule := rule.ConfiguredRule{Name: "native/rule", Severity: rule.SeverityError}
	pluginConfiguredRule := pluginRule("external/rule", []any{"option"}, rule.SeverityWarning)
	plan := &LintPlan{programs: []programLintPlan{{files: []lintFilePlan{
		{file: nativeFile, rules: prepareRuleSet([]rule.ConfiguredRule{nativeRule}, true)},
		{file: pluginFile, rules: prepareRuleSet([]rule.ConfiguredRule{nativeRule, pluginConfiguredRule}, true)},
	}}}}

	resolveCalls := 0
	inputs := BuildEslintPluginFileInputs(plan, func(filePath string) EslintPluginFileConfig {
		resolveCalls++
		if filePath != paths["plugin.ts"] {
			t.Errorf("resolved config for %q, want only plugin.ts", filePath)
		}
		return EslintPluginFileConfig{
			ConfigKey:       "owner-key",
			LanguageOptions: map[string]any{"sourceType": "module"},
			Settings:        map[string]any{"workspace": true},
		}
	})

	if resolveCalls != 1 {
		t.Fatalf("config resolver calls = %d, want one for plugin targets only", resolveCalls)
	}
	if len(inputs) != 1 {
		t.Fatalf("plugin inputs = %d, want 1", len(inputs))
	}
	input := inputs[0]
	if input.Path != paths["plugin.ts"] || input.SourceFile != pluginFile || input.Text != nil {
		t.Errorf("input source frame = %+v, want prepared plugin.ts source", input)
	}
	if input.ConfigKey != "owner-key" || input.LanguageOptions["sourceType"] != "module" || input.Settings["workspace"] != true {
		t.Errorf("input config projection = %+v", input)
	}
	if len(input.Rules) != 1 || input.Rules[0].Name != "external/rule" {
		t.Errorf("input rules = %v, want only the plugin rule", input.Rules)
	}

	withoutConfig := BuildEslintPluginFileInputs(plan, nil)
	if len(withoutConfig) != 1 || withoutConfig[0].ConfigKey != "" {
		t.Errorf("nil config resolver inputs = %+v, want one input without a routing key", withoutConfig)
	}
}

func TestDispatchEslintPluginRulesWithOutcomeSurfacesFailure(t *testing.T) {
	goodText := "good\n"
	badText := "bad\n"
	files := []EslintPluginFileInput{
		{
			Path:      "/repo/good.ts",
			ConfigKey: "/repo",
			Text:      &goodText,
			Rules:     []rule.ConfiguredRule{pluginRule("external/rule", []any{"good"}, rule.SeverityWarning)},
		},
		{
			Path:      "/repo/bad.ts",
			ConfigKey: "/repo",
			Text:      &badText,
			Rules:     []rule.ConfiguredRule{pluginRule("external/rule", []any{"bad"}, rule.SeverityError)},
		},
	}
	dispatchError := errors.New("WorkerPool: closed")
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		file := request.Files[0]
		if file.Path == "/repo/bad.ts" {
			return nil, dispatchError
		}
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: file.Path,
			Diagnostics: []EslintPluginDiagnostic{{
				RuleName: "external/rule",
				Message:  "reported before another batch failed",
				StartPos: 0,
				EndPos:   4,
			}},
		}}}, nil
	})

	outcome := testDispatchEslintPluginRulesWithOutcome(
		context.Background(),
		dispatch,
		files,
		false,
		SuggestionsModeOff,
		nil,
	)
	if !errors.Is(outcome.DispatchError, dispatchError) {
		t.Fatalf("dispatch error = %v, want %v", outcome.DispatchError, dispatchError)
	}
	if len(outcome.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want partial result plus failure: %+v", len(outcome.Diagnostics), outcome.Diagnostics)
	}
	if outcome.Diagnostics[0].RuleName != "external/rule" {
		t.Errorf("first diagnostic = %q, want successful batch result", outcome.Diagnostics[0].RuleName)
	}
	failure := outcome.Diagnostics[1]
	if failure.RuleName != "rslint/plugin-lint-error" || failure.Severity != rule.SeverityError {
		t.Errorf("failure diagnostic = %q/%v", failure.RuleName, failure.Severity)
	}
	if failure.FilePath != files[0].Path {
		t.Errorf("failure path = %q, want first input %q", failure.FilePath, files[0].Path)
	}
	if !strings.Contains(failure.Message.Description, dispatchError.Error()) {
		t.Errorf("failure message = %q, want dispatch error", failure.Message.Description)
	}
}

func TestDispatchEslintPluginRulesWithOutcomePreservesCancellation(t *testing.T) {
	dispatch := EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return nil, context.Canceled
	})
	outcome := testDispatchEslintPluginRulesWithOutcome(
		context.Background(),
		dispatch,
		[]EslintPluginFileInput{{
			Path:      "/repo/a.ts",
			ConfigKey: "/repo",
			Rules:     []rule.ConfiguredRule{pluginRule("external/rule", nil, rule.SeverityError)},
		}},
		false,
		SuggestionsModeOff,
		nil,
	)
	if !errors.Is(outcome.DispatchError, context.Canceled) {
		t.Fatalf("dispatch error = %v, want context.Canceled", outcome.DispatchError)
	}
	if len(outcome.Diagnostics) != 0 {
		t.Errorf("cancellation diagnostics = %v, want none", outcome.Diagnostics)
	}
}

func TestDispatchEslintPluginRulesWithOutcomeReturnsProtocolNotices(t *testing.T) {
	text := "const value = 1;"
	files := []EslintPluginFileInput{
		{
			Path:  "/repo/present.ts",
			Text:  &text,
			Rules: []rule.ConfiguredRule{pluginRule("external/configured", nil, rule.SeverityWarning)},
		},
		{
			Path:  "/repo/missing.ts",
			Text:  &text,
			Rules: []rule.ConfiguredRule{pluginRule("external/configured", nil, rule.SeverityWarning)},
		},
	}
	outcome := testDispatchEslintPluginRulesWithOutcome(
		context.Background(),
		EslintPluginDispatchFunc(func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
				FilePath: "/repo/present.ts",
				Diagnostics: []EslintPluginDiagnostic{{
					RuleName: "external/unconfigured",
					Message:  "unexpected diagnostic",
				}},
			}}}, nil
		}),
		files,
		false,
		SuggestionsModeOff,
		nil,
	)
	if outcome.DispatchError != nil || len(outcome.Diagnostics) != 1 {
		t.Fatalf("outcome = %+v, want one usable diagnostic and no dispatch error", outcome)
	}
	want := []EslintPluginProtocolNotice{
		{Kind: EslintPluginUnconfiguredDiagnostic, FilePath: "/repo/present.ts", RuleName: "external/unconfigured"},
		{Kind: EslintPluginMissingFileResult, FilePath: "/repo/missing.ts"},
	}
	if !reflect.DeepEqual(outcome.Notices, want) {
		t.Fatalf("notices = %+v, want %+v", outcome.Notices, want)
	}
}

func TestDispatchEslintPluginRulesWithOutcomeSkipsEmptyInput(t *testing.T) {
	called := false
	dispatch := EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		called = true
		return &EslintPluginLintResult{}, nil
	})
	outcome := testDispatchEslintPluginRulesWithOutcome(
		context.Background(),
		dispatch,
		nil,
		false,
		SuggestionsModeOff,
		nil,
	)
	if outcome.DispatchError != nil || len(outcome.Diagnostics) != 0 {
		t.Errorf("empty dispatch outcome = %+v, want zero value", outcome)
	}
	if called {
		t.Error("dispatch called for empty input")
	}
}

func TestDispatchEslintPluginReleasesProjectedFrameAfterTransfer(t *testing.T) {
	text := "const value = '" + strings.Repeat("x", 5<<20) + "';"
	input := EslintPluginFileInput{
		Path:       "/repo/file.ts",
		Text:       &text,
		SourceFile: pluginSyntax("/repo/file.ts", text),
		Rules:      []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)},
	}
	outcome := DispatchEslintPluginRulesWithOutcome(context.Background(),
		EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			if len(request.Files) != 1 || len(request.Files[0].Frame) <= 4<<20 {
				return nil, errors.New("expected a large projected frame")
			}
			wire, frames := request.WithFrameAttachments()
			frame := weak.Make(&frames[0][0])
			// Model completed publication without mutating the request. Only the
			// transport's copy and frame-free wire metadata are needed from here.
			published := bytes.Clone(frames[0])
			runtime.GC()
			if frame.Value() != nil {
				return nil, errors.New("projected frame remains reachable after transfer")
			}
			runtime.KeepAlive(published)
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: wire.Files[0].Path}}}, nil
		}), []EslintPluginFileInput{input}, false, SuggestionsModeOff, nil)
	if outcome.DispatchError != nil {
		t.Fatal(outcome.DispatchError)
	}
}

func pluginSyntax(path, text string) *ast.SourceFile {
	return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(tspath.NormalizePath(path))}, strings.TrimPrefix(text, "\ufeff"), core.GetScriptKindFromFileName(path))
}

// Dispatcher tests exercise batching/result policies with real frames, while
// retaining their authored diagnostic source setup independently.
func pluginTestFrames(inputs []EslintPluginFileInput) []EslintPluginFileInput {
	inputs = append([]EslintPluginFileInput(nil), inputs...)
	var builder estree.Builder
	for index, input := range inputs {
		if input.Frame != nil || input.FrameError != "" {
			continue
		}
		file := input.SourceFile
		if file == nil {
			text := ""
			if input.Text != nil {
				text = *input.Text
			}
			file = pluginSyntax(input.Path, text)
		}
		sourceType := "module"
		if value, ok := input.LanguageOptions["sourceType"].(string); ok {
			sourceType = value
		}
		frame, err := builder.Build(context.Background(), file, estree.Options{SourceType: sourceType})
		if err != nil {
			panic(err)
		}
		inputs[index].Frame = frame
	}
	return inputs
}
func testDispatchEslintPluginRules(ctx context.Context, dispatch EslintPluginDispatchFunc, inputs []EslintPluginFileInput, fixes bool, suggestions string, timing *TimingCollector, emit DiagnosticHandler) error {
	return DispatchEslintPluginRules(ctx, dispatch, pluginTestFrames(inputs), fixes, suggestions, timing, emit)
}
func testDispatchEslintPluginRulesWithOutcome(ctx context.Context, dispatch EslintPluginDispatchFunc, inputs []EslintPluginFileInput, fixes bool, suggestions string, timing *TimingCollector) EslintPluginDispatchOutcome {
	return DispatchEslintPluginRulesWithOutcome(ctx, dispatch, pluginTestFrames(inputs), fixes, suggestions, timing)
}
func testDispatchEslintPluginRulesAsync(ctx context.Context, dispatch EslintPluginDispatchFunc, inputs []EslintPluginFileInput, fixes bool, suggestions string, timing *TimingCollector) <-chan EslintPluginDispatchOutcome {
	return DispatchEslintPluginRulesAsync(ctx, dispatch, pluginTestFrames(inputs), fixes, suggestions, timing)
}
func pluginFrameText(data []byte) string { return string(data[binary.LittleEndian.Uint32(data[60:]):]) }

func TestDispatchEslintPluginProjectionWindowsOwnFrames(t *testing.T) {
	t.Run("bounded output blocks preserve every borrowed view", func(t *testing.T) {
		var builder pluginProjectionBuffers
		for round := range 2 {
			var frames [][]byte
			for index, size := range []int{1, 64 << 10, 1<<20 + 3, 2<<20 + 7, 32 << 10, 4<<20 + 1} {
				frame := builder.allocate(size)
				if len(frame) != size || cap(frame) != size {
					t.Fatal("output view can extend into another frame")
				}
				for offset := range frame {
					frame[offset] = byte(index + round)
				}
				frames = append(frames, frame)
				for previous, borrowed := range frames {
					for _, value := range borrowed {
						if value != byte(previous+round) {
							t.Fatal("allocation changed an earlier borrowed frame")
						}
					}
				}
			}
			capacity := 0
			for _, block := range builder.output {
				capacity += cap(block)
			}
			if capacity != builder.outputCapacity || capacity > 4<<20 {
				t.Fatal("retained output exceeded the cache bound")
			}
			builder.reset()
			for _, block := range builder.output {
				if len(block) != 0 {
					t.Fatal("reset retained a reserved output span")
				}
			}
		}
	})
	inputs := make([]EslintPluginFileInput, 65)
	for index := range inputs {
		path := fmt.Sprintf("/repo/file%d.ts", index)
		// Vary frame sizes enough to allocate several stable output blocks
		// and exceed its cache bound within each complete window.
		text := fmt.Sprintf("const value%d = %d; /*%s*/", index, index, strings.Repeat("pad ", (index%pluginWindowFiles+1)*2048))
		inputs[index] = EslintPluginFileInput{Path: path, Text: &text, SourceFile: pluginSyntax(path, text), Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
	}
	var windows []int
	var frames [][]byte
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		windows = append(windows, len(request.Files))
		results := make([]EslintPluginFileResult, len(request.Files))
		for index, file := range request.Files {
			if !strings.HasPrefix(pluginFrameText(file.Frame), "const value") {
				t.Fatal("request lost its source")
			}
			// Dispatch borrows newly projected bytes until it returns. A
			// retained snapshot, like shared/inline transport, owns its copy.
			frames = append(frames, bytes.Clone(file.Frame))
			results[index].FilePath = file.Path
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	ownedInputs := slices.Clone(inputs)
	if err := dispatchOneBatch(context.Background(), dispatch, ownedInputs, make(chan *pluginProjectionBuffers, 1), false, SuggestionsModeOff, nil, func(rule.RuleDiagnostic) {}, func(EslintPluginProtocolNotice) {}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(windows, []int{32, 32, 1}) {
		t.Fatalf("unbounded windows: %v", windows)
	}
	for index, frame := range frames {
		if ownedInputs[index].SourceFile != nil || ownedInputs[index].Text != nil {
			t.Fatal("completed projection retained an owned input slot")
		}
		if pluginFrameText(frame) != *inputs[index].Text {
			t.Fatal("later projection overwrote an earlier frame")
		}
		if inputs[index].Frame != nil {
			t.Fatal("dispatch retained a whole-group frame cache")
		}
	}
	t.Run("byte limit leaves the next output unallocated", func(t *testing.T) {
		inputs := make([]EslintPluginFileInput, 3)
		for index := range inputs {
			path := fmt.Sprintf("/repo/lookahead%d.ts", index)
			text := fmt.Sprintf("const lookahead%d = %d;", index, index)
			inputs[index] = EslintPluginFileInput{Path: path, Text: &text, SourceFile: pluginSyntax(path, text), Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
		}
		// Frozen inputs are opaque to the scheduler. Fill the byte window so
		// the next projection must wait for the next request before allocating.
		inputs[0].Frame = make([]byte, pluginWindowBytes)
		calls := 0
		dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			calls++
			for _, file := range request.Files {
				if file.Path != inputs[0].Path {
					for _, input := range inputs[1:] {
						if file.Path == input.Path && pluginFrameText(file.Frame) != *input.Text {
							t.Fatal("recycled storage overwrote the pending frame")
						}
					}
				}
			}
			return &EslintPluginLintResult{}, nil
		})
		allocations, releases := 0, 0
		provider := testPluginBatchProvider(func(context.Context) (*EslintPluginBatch, error) {
			return &EslintPluginBatch{Allocate: func(size int) []byte { allocations++; return make([]byte, size) }, Dispatch: dispatch, Release: func() { releases++ }}, nil
		})
		if err := dispatchOneBatch(context.Background(), provider, slices.Clone(inputs), make(chan *pluginProjectionBuffers, 1), false, SuggestionsModeOff, nil, func(rule.RuleDiagnostic) {}, func(EslintPluginProtocolNotice) {}); err != nil {
			t.Fatal(err)
		}
		if calls != 2 || releases != 2 || allocations != 3 {
			t.Fatalf("windows/callbacks/allocations = %d/%d/%d, want 2/2/3", calls, releases, allocations)
		}
	})
}

func TestDispatchEslintPluginProjectionFailureCannotFallBack(t *testing.T) {
	current := "const current = 1;"
	stale := "const stale = 2;"
	input := EslintPluginFileInput{Path: "/repo/file.ts", Text: &stale, SourceFile: pluginSyntax("/repo/file.ts", current), Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
	var diagnostics []rule.RuleDiagnostic
	called := false
	outcome := DispatchEslintPluginRulesWithOutcome(context.Background(), EslintPluginDispatchFunc(func(context.Context, EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		called = true
		return &EslintPluginLintResult{}, nil
	}), []EslintPluginFileInput{input}, false, SuggestionsModeOff, nil)
	diagnostics = outcome.Diagnostics
	if called || len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message.Description, "does not match Program syntax") {
		t.Fatalf("projection failure became a successful lint: %v/%+v", called, outcome)
	}
	input.Text = nil
	var builder estree.Builder
	if _, err := projectPluginFrame(context.Background(), &builder, input, nil); err == nil {
		t.Fatal("missing snapshot accepted")
	}
	input.Text = &current
	input.SourceFile.Statements.Nodes[0].Kind = ast.KindUnknown
	if frame, err := projectPluginFrame(context.Background(), &builder, input, nil); err == nil || frame != nil {
		t.Fatal("unknown syntax accepted")
	}
}

func TestDispatchEslintPluginHomogeneousWindowsOverlapAndKeepOrder(t *testing.T) {
	limit := dispatchConcurrency()
	inputs := make([]EslintPluginFileInput, pluginWindowFiles*(limit+3)+1)
	for index := range inputs {
		text := fmt.Sprintf("const value%d = %d;", index, index)
		path := fmt.Sprintf("/repo/%d.ts", index)
		inputs[index] = EslintPluginFileInput{Path: path, Text: &text, SourceFile: pluginSyntax(path, text), Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started := make(chan struct{}, limit)
	release := make(chan struct{})
	dispatch := EslintPluginDispatchFunc(func(ctx context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		before := make([][]byte, len(request.Files))
		for index, file := range request.Files {
			before[index] = bytes.Clone(file.Frame)
		}
		select {
		case started <- struct{}{}:
		default:
		}
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		results := make([]EslintPluginFileResult, len(request.Files))
		for index, file := range request.Files {
			runtime.Gosched()
			if !bytes.Equal(file.Frame, before[index]) {
				return nil, errors.New("another request reused an in-flight output buffer")
			}
			results[index] = EslintPluginFileResult{FilePath: file.Path, Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "diagnostic", StartPos: 0, EndPos: 1}}}
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	outcomes := make(chan EslintPluginDispatchOutcome, 1)
	go func() {
		outcomes <- DispatchEslintPluginRulesWithOutcome(ctx, dispatch, inputs, false, SuggestionsModeOff, nil)
	}()
	for range limit {
		select {
		case <-started:
		case <-ctx.Done():
			<-outcomes
			t.Fatal("homogeneous windows did not use existing dispatch concurrency")
		}
	}
	close(release)
	outcome := <-outcomes
	if outcome.DispatchError != nil || len(outcome.Diagnostics) != len(inputs) {
		t.Fatalf("outcome = %+v", outcome)
	}
	for index, diagnostic := range outcome.Diagnostics {
		if diagnostic.FilePath != inputs[index].Path {
			t.Fatal("concurrent projection windows reordered diagnostics")
		}
	}
}

func TestDispatchEslintPluginProjectionErrorsKeepFileOrder(t *testing.T) {
	inputs := make([]EslintPluginFileInput, 3)
	for index := range inputs {
		path := fmt.Sprintf("/repo/%d.ts", index)
		text := "const x = 1;"
		inputs[index] = EslintPluginFileInput{Path: path, Text: &text, SourceFile: pluginSyntax(path, text), Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
	}
	stale := "const stale = 1;"
	inputs[1].Text = &stale
	dispatch := EslintPluginDispatchFunc(func(_ context.Context, request EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		if len(request.Files) != 2 {
			t.Fatalf("failed file sent to worker: %+v", request.Files)
		}
		results := make([]EslintPluginFileResult, len(request.Files))
		for index, file := range request.Files {
			results[index] = EslintPluginFileResult{FilePath: file.Path, Diagnostics: []EslintPluginDiagnostic{{RuleName: "plugin/check", Message: "diagnostic", StartPos: 0, EndPos: 1}}}
		}
		return &EslintPluginLintResult{Results: results}, nil
	})
	outcome := DispatchEslintPluginRulesWithOutcome(context.Background(), dispatch, inputs, false, SuggestionsModeOff, nil)
	if outcome.DispatchError != nil || len(outcome.Diagnostics) != 3 {
		t.Fatalf("outcome = %+v", outcome)
	}
	for index, diagnostic := range outcome.Diagnostics {
		if diagnostic.FilePath != inputs[index].Path {
			t.Fatalf("projection errors reordered results: %+v", outcome.Diagnostics)
		}
	}
}

// Test storage invalidates every borrowed byte at Release, independently of the
// owned-frame adapter, so result rebuilding cannot accidentally depend on it.
type testPluginBatchProvider func(context.Context) (*EslintPluginBatch, error)

func (f testPluginBatchProvider) OpenBatch(ctx context.Context) (*EslintPluginBatch, error) {
	return f(ctx)
}

func TestDispatchEslintPluginUsesBatchOwnedOutput(t *testing.T) {
	for _, failure := range []string{"", "dispatch panic"} {
		t.Run(failure, func(t *testing.T) {
			opened, released, allocated := 0, 0, 0
			inputs := make([]EslintPluginFileInput, 33)
			expected := make(map[string][]byte)
			for i := range inputs {
				path := fmt.Sprintf("/repo/owned%d.ts", i)
				text := fmt.Sprintf("const value%d = '世界😀';", i)
				file := pluginSyntax(path, text)
				inputs[i] = EslintPluginFileInput{Path: path, Text: &text, SourceFile: file, Rules: []rule.ConfiguredRule{pluginRule("plugin/check", nil, rule.SeverityError)}}
				var b estree.Builder
				data, err := b.Build(context.Background(), file, estree.Options{SourceType: "module"})
				if err != nil {
					t.Fatal(err)
				}
				expected[path] = data
			}
			provider := testPluginBatchProvider(func(context.Context) (*EslintPluginBatch, error) {
				opened++
				storage := make([]byte, 64<<10)
				used := 0
				closed := false
				return &EslintPluginBatch{
					Allocate: func(size int) []byte {
						if closed {
							panic("allocation after release")
						}
						allocated++
						data := storage[used : used+size : used+size]
						used += size
						return data
					},
					Dispatch: func(_ context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
						if closed {
							panic("dispatch after release")
						}
						if failure != "" {
							panic("fixture dispatch panic")
						}
						for _, f := range req.Files {
							if !bytes.Equal(f.Frame, expected[f.Path]) {
								return nil, errors.New("provided output changed the encoded frame")
							}
						}
						return &EslintPluginLintResult{}, nil
					},
					Release: func() {
						if closed {
							panic("double output release")
						}
						closed = true
						released++
						for i := range storage {
							storage[i] = 0xee
						}
					},
				}, nil
			})
			err := dispatchOneBatch(context.Background(), provider, slices.Clone(inputs), make(chan *pluginProjectionBuffers, 1), false, SuggestionsModeOff, nil, func(rule.RuleDiagnostic) {}, func(EslintPluginProtocolNotice) {})
			if failure == "" && err != nil {
				t.Fatal(err)
			}
			if failure != "" && err == nil {
				t.Fatal("dispatch panic was lost")
			}
			if opened != released || (failure == "" && (opened != 2 || allocated != 33)) {
				t.Fatalf("output lifetime opened/released/allocated=%d/%d/%d", opened, released, allocated)
			}
		})
	}
}
