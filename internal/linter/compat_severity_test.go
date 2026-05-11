package linter

import (
	"context"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// #10 regression: dispatchCompatBucket must not index sevByRule with a raw
// map access. A missing key returns the zero value, and SeverityError is
// iota 0 — so a worker-reported diagnostic for a rule that isn't in the
// batch's config would be silently surfaced as a build-breaking error
// (and a configured 'warn' likewise promoted). The fix drops + warns on the
// unconfigured rule instead, and configured rules keep their real severity.
//
// Pre-fix this test FAILS: the unconfigured diagnostic surfaces as a second
// (Error-severity) diagnostic. It is not a happens-to-pass assertion.
func TestDispatchCompat_DropsDiagnosticForUnconfiguredRule(t *testing.T) {
	dispatcher := CompatBatchHandler(
		func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
			return []CompatFileResult{{
				FilePath: batch.Files[0].Path,
				Diagnostics: []CompatDiagnostic{
					{RuleName: "pkg/configured", Message: "configured", StartPos: 0, EndPos: 5},
					{RuleName: "pkg/unconfigured", Message: "stray", StartPos: 0, EndPos: 5},
				},
			}}, nil
		},
	)
	entry := CompatFileEntry{
		Path:     "/a.ts",
		Text:     "const x = 1;",
		Rules:    map[string]CompatRuleConfig{"pkg/configured": {}},
		Severity: map[string]rule.DiagnosticSeverity{"pkg/configured": rule.SeverityWarning},
	}

	var got []rule.RuleDiagnostic
	if _, err := DispatchCompat(DispatchCompatOptions{
		Files:        []CompatFileEntry{entry},
		Dispatcher:   dispatcher,
		OnDiagnostic: func(d rule.RuleDiagnostic) { got = append(got, d) },
	}); err != nil {
		t.Fatalf("DispatchCompat: %v", err)
	}

	// Only the configured rule's diagnostic surfaces; the unconfigured one is
	// dropped rather than indexed into the zero-value severity.
	if len(got) != 1 {
		t.Fatalf("expected 1 diagnostic (unconfigured dropped), got %d: %+v", len(got), got)
	}
	if got[0].RuleName != "pkg/configured" {
		t.Errorf("surfaced diagnostic RuleName = %q, want pkg/configured", got[0].RuleName)
	}
	// The configured rule keeps its real severity (Warning), NOT the
	// zero-value Error a missing-key map index would have produced.
	if got[0].Severity != rule.SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning (zero-value Error trap)", got[0].Severity)
	}
}

// #2 (review): Ctrl-C / LSP supersession cancels ctx mid-dispatch; the
// dispatcher then returns context.Canceled. That is a deliberate stop, not a
// dispatcher failure — DispatchCompat must NOT count it toward
// CompatDispatchErrors, and must short-circuit the remaining buckets rather
// than fire each one into a dead ctx (worker/child cleanup is owned by the
// signal path, not this loop).
//
// Pre-fix this FAILS: no ctx.Err() short-circuit (both buckets dispatched →
// dispatchCalls=2) and context.Canceled counted as a failure
// (CompatDispatchErrors=2).
func TestDispatchCompat_CancellationShortCircuitsAndIsNotFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	dispatchCalls := 0
	dispatcher := CompatBatchHandler(
		func(_ context.Context, _ CompatBatch) ([]CompatFileResult, error) {
			dispatchCalls++
			cancel() // the signal lands during the first dispatch
			return nil, context.Canceled
		},
	)
	// Two distinct rule signatures → two buckets; without the short-circuit
	// the second bucket would also be dispatched.
	files := []CompatFileEntry{
		{
			Path:     "/a.ts",
			Text:     "const x = 1;",
			Rules:    map[string]CompatRuleConfig{"p/r1": {}},
			Severity: map[string]rule.DiagnosticSeverity{"p/r1": rule.SeverityError},
		},
		{
			Path:     "/b.ts",
			Text:     "const y = 1;",
			Rules:    map[string]CompatRuleConfig{"p/r2": {}},
			Severity: map[string]rule.DiagnosticSeverity{"p/r2": rule.SeverityError},
		},
	}

	res, err := DispatchCompat(DispatchCompatOptions{
		Files:        files,
		Dispatcher:   dispatcher,
		Ctx:          ctx,
		OnDiagnostic: func(rule.RuleDiagnostic) {},
	})
	if err != nil {
		t.Fatalf("DispatchCompat returned error: %v", err)
	}
	if res.CompatDispatchErrors != 0 {
		t.Errorf("CompatDispatchErrors = %d, want 0 (cancellation is not a dispatch failure)", res.CompatDispatchErrors)
	}
	if dispatchCalls != 1 {
		t.Errorf("dispatcher called %d times, want 1 (loop must short-circuit after cancellation)", dispatchCalls)
	}
}
