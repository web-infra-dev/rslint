package linter

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Cancellation propagation: RunLinterOptions.Ctx must abort the per-Program
// file loop at the next file boundary and surface ctx.Err() from RunLinter.
// We count distinct files the rule's listeners fired on; cancel must bound
// that count strictly below the total file count.
func TestRunLinter_CtxCancel_PartialResultReturned(t *testing.T) {
	const totalFiles = 50
	files := map[string]string{}
	for i := range totalFiles {
		files[twoDigit(i)+".ts"] = "const x = 1;"
	}
	program, _ := createTestProgramWithFiles(t, files)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Track distinct files visited via the rule listener. We can't use
	// LintedFileCount because that's set to len(filesToLint) BEFORE the
	// rule loop runs — that's the "candidate" count, by design. The
	// cancellation contract is about the RULE LOOP not visiting all
	// candidates; observing identifier-fire-per-file is the right
	// probe.
	visitedFiles := make(map[string]struct{})
	var visitedMu sync.Mutex
	var fileVisitCount atomic.Int32

	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:     "test-rule",
			Severity: rule.SeverityWarning,
			Run: func(rc rule.RuleContext) rule.RuleListeners {
				return rule.RuleListeners{
					ast.KindIdentifier: func(node *ast.Node) {
						visitedMu.Lock()
						if _, seen := visitedFiles[rc.SourceFile.FileName()]; !seen {
							visitedFiles[rc.SourceFile.FileName()] = struct{}{}
							n := fileVisitCount.Add(1)
							// On the 5th file's first identifier, cancel.
							// The CURRENT file finishes; the NEXT iteration's
							// file-boundary check observes the cancel and breaks.
							if n == 5 {
								cancel()
							}
						}
						visitedMu.Unlock()
					},
				}
			},
		}}
	}

	res, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: getRules,
		OnDiagnostic:    func(rule.RuleDiagnostic) {},
		Ctx:             ctx,
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil partial LintResult after cancel")
	}

	visitedMu.Lock()
	final := len(visitedFiles)
	visitedMu.Unlock()
	// The cancel was triggered on the 5th file. The 5th file's loop
	// body completes (we don't interrupt rule traversal mid-file), so
	// we expect exactly 5 — the loop bails before file 6. Allow a
	// small upper slack (here +3) for race-driven over-shoot in
	// parallel build configurations, but reject anything close to
	// totalFiles. The previous [1, 49] bound was so wide it would
	// pass even if cancel only fired on file 49 of 50.
	if final < 5 {
		t.Errorf("visited %d files; cancel fired on file 5 so at least 5 expected", final)
	}
	if final > 8 {
		t.Errorf("visited %d files (cap=8); cancel did not bound the loop tightly enough", final)
	}
	t.Logf("cancel observed; visited %d/%d files", final, totalFiles)
}

// CompatBatchHandler receives the lint ctx. A cancelled ctx must be
// observable inside the dispatcher so a long-running compat IPC can be
// interrupted promptly instead of running to its own per-batch timeout.
//
// We can't easily run a "real" compat dispatch in a unit test (it
// requires a live Node WorkerPool), so we install a stub dispatcher
// that records its received ctx and asserts cancellation propagates.
func TestRunLinter_Ctx_PropagatedToCompatDispatcher(t *testing.T) {
	files := map[string]string{}
	for i := range 3 {
		files[twoDigit(i)+".ts"] = "const x = 1;"
	}
	program, _ := createTestProgramWithFiles(t, files)

	ctx, cancel := context.WithCancel(context.Background())

	// Rules that look like eslint-plugin rules so the dispatcher fires.
	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		return []ConfiguredRule{{
			Name:               "fake-plugin/no-x",
			Severity:           rule.SeverityWarning,
			IsEslintPluginRule: true,
			Run:                func(rc rule.RuleContext) rule.RuleListeners { return nil },
		}}
	}

	var dispatcherCalled atomic.Bool
	var dispatcherCtxErr error
	dispatcher := CompatBatchHandler(func(dctx context.Context, batch CompatBatch) ([]CompatFileResult, error) {
		dispatcherCalled.Store(true)
		// Cancel mid-dispatch from the test side; the dispatcher must
		// observe the cancellation on the passed ctx.
		cancel()
		// Yield once so the cancel callback fires.
		select {
		case <-dctx.Done():
			dispatcherCtxErr = dctx.Err()
		case <-time.After(2 * time.Second):
			dispatcherCtxErr = errors.New("ctx did not fire within 2s")
		}
		return nil, dctx.Err()
	})

	_, _ = RunLinter(RunLinterOptions{
		Programs:             []*compiler.Program{program},
		SingleThreaded:       true,
		GetRulesForFile:      getRules,
		OnDiagnostic:         func(rule.RuleDiagnostic) {},
		CompatRuleDispatcher: dispatcher,
		Ctx:                  ctx,
	})

	if !dispatcherCalled.Load() {
		t.Fatal("dispatcher was never invoked — rules with IsEslintPluginRule=true should have triggered it")
	}
	if !errors.Is(dispatcherCtxErr, context.Canceled) {
		t.Errorf("dispatcher's ctx did not observe cancellation; got %v, want context.Canceled", dispatcherCtxErr)
	}
}

// Pre-cancelled ctx: when RunLinter is invoked with an already-cancelled
// ctx, no file is visited.
func TestRunLinter_CtxCancel_PreCancelledSkipsAllFiles(t *testing.T) {
	files := map[string]string{}
	for i := range 10 {
		files[twoDigit(i)+".ts"] = "const x = 1;"
	}
	program, _ := createTestProgramWithFiles(t, files)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	var visited atomic.Int32
	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		visited.Add(1)
		return nil
	}

	res, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: getRules,
		OnDiagnostic:    func(rule.RuleDiagnostic) {},
		Ctx:             ctx,
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil partial LintResult")
	}
	if visited.Load() != 0 {
		t.Errorf("ctx pre-cancelled but getRulesForFile invoked %d times", visited.Load())
	}
}

// Phase-2 type-check must NOT run after Phase-1 was cancelled. Before
// the fix, RunLinter unconditionally proceeded to type-check even when
// the caller had asked to stop — user Ctrl-C waited for type-check to
// finish (seconds-to-minutes on real projects). The fix: gate Phase 2
// on ctx.Err() == nil.
func TestRunLinter_CancelledCtxSkipsTypeCheck(t *testing.T) {
	files := map[string]string{}
	for i := range 5 {
		// File content that WOULD produce TS errors if type-check ran
		// (assigning string to number).
		files[twoDigit(i)+".ts"] = "const x: number = 'hello';"
	}
	program, _ := createTestProgramWithFiles(t, files)

	// Pre-cancel the ctx so Phase 1 bails immediately and we observe
	// Phase 2 behavior directly.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var tsDiagSeen atomic.Bool
	onDiag := func(d rule.RuleDiagnostic) {
		if strings.HasPrefix(d.RuleName, "TypeScript(") {
			tsDiagSeen.Store(true)
		}
	}

	_, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: func(*ast.SourceFile) []ConfiguredRule { return nil },
		TypeCheck:       true,
		OnDiagnostic:    onDiag,
		Ctx:             ctx,
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if tsDiagSeen.Load() {
		t.Error("type-check ran despite ctx being cancelled before RunLinter started")
	}
}

// Nil ctx: the existing behavior (uncancellable) must keep working — no
// regression for callers that don't opt in.
func TestRunLinter_NilCtx_StillCompletes(t *testing.T) {
	program, _ := createTestProgramWithFiles(t, map[string]string{
		"a.ts": "const a = 1;",
		"b.ts": "const b = 2;",
	})

	var visited atomic.Int32
	getRules := func(sf *ast.SourceFile) []ConfiguredRule {
		visited.Add(1)
		return nil
	}

	res, err := RunLinter(RunLinterOptions{
		Programs:        []*compiler.Program{program},
		SingleThreaded:  true,
		GetRulesForFile: getRules,
		OnDiagnostic:    func(rule.RuleDiagnostic) {},
		// Ctx left nil — covers the "uncancellable" path.
	})

	if err != nil {
		t.Fatalf("unexpected error with nil ctx: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if visited.Load() != 2 {
		t.Errorf("expected both files visited, got %d", visited.Load())
	}
}

// twoDigit pads small integers so map iteration / file-name ordering is
// readable in test failure output.
func twoDigit(i int) string {
	if i < 10 {
		return "0" + string(rune('0'+i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}
