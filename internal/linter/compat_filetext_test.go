package linter

import (
	"context"
	"testing"
)

// #3 regression: the LSP/--api path may lint an UNSAVED editor buffer whose
// content differs from disk. CompatLintFile gained an optional Text field
// the worker prefers over re-reading disk; DispatchCompat must populate it
// only when IncludeFileText is set (LSP), and omit it otherwise (CLI keeps
// the IPC payload bounded by letting the worker read disk).
func TestDispatchCompat_IncludeFileText_ShipsOverlayElseOmits(t *testing.T) {
	var captured CompatBatch
	dispatcher := CompatBatchHandler(
		func(_ context.Context, batch CompatBatch) ([]CompatFileResult, error) {
			captured = batch
			out := make([]CompatFileResult, 0, len(batch.Files))
			for _, f := range batch.Files {
				out = append(out, CompatFileResult{FilePath: f.Path})
			}
			return out, nil
		},
	)
	entry := CompatFileEntry{
		Path:  "/a.ts",
		Text:  "const overlay = 1;",
		Rules: map[string]CompatRuleConfig{"r": {}},
	}

	// IncludeFileText=true: ship the (possibly unsaved) overlay so the
	// worker lints it instead of re-reading the stale on-disk file.
	if _, err := DispatchCompat(DispatchCompatOptions{
		Files:           []CompatFileEntry{entry},
		Dispatcher:      dispatcher,
		IncludeFileText: true,
	}); err != nil {
		t.Fatalf("DispatchCompat: %v", err)
	}
	if len(captured.Files) != 1 {
		t.Fatalf("expected 1 batched file, got %d", len(captured.Files))
	}
	if captured.Files[0].Text == nil {
		t.Fatal("IncludeFileText=true must ship Text, got nil")
	}
	if *captured.Files[0].Text != "const overlay = 1;" {
		t.Errorf("Text = %q, want the overlay", *captured.Files[0].Text)
	}

	// IncludeFileText=false (CLI default): no Text — worker reads disk.
	captured = CompatBatch{}
	if _, err := DispatchCompat(DispatchCompatOptions{
		Files:      []CompatFileEntry{entry},
		Dispatcher: dispatcher,
	}); err != nil {
		t.Fatalf("DispatchCompat: %v", err)
	}
	if len(captured.Files) != 1 {
		t.Fatalf("expected 1 batched file, got %d", len(captured.Files))
	}
	if captured.Files[0].Text != nil {
		t.Errorf("IncludeFileText=false must omit Text, got %q", *captured.Files[0].Text)
	}
}
