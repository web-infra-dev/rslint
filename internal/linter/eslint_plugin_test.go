package linter

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func pluginRule(name string, opts any, sev rule.DiagnosticSeverity) ConfiguredRule {
	return ConfiguredRule{Name: name, Options: opts, Severity: sev, IsEslintPluginRule: true}
}

func TestDispatchEslintPlugin_EmptyShortCircuit(t *testing.T) {
	called := 0
	dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		called++
		return &EslintPluginLintResult{}, nil
	}
	if err := DispatchEslintPluginRules(context.Background(), dispatch, nil, false, "off", func(rule.RuleDiagnostic) {}); err != nil {
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
		{Path: "/a.ts", ConfigKey: "/cfg", Rules: []ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
		{Path: "/b.ts", ConfigKey: "/cfg", Rules: []ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
		{Path: "/c.ts", ConfigKey: "/cfg", Rules: []ConfiguredRule{pluginRule("uc/x", []any{"o2"}, rule.SeverityError)}},
		{Path: "/d.ts", ConfigKey: "/other", Rules: []ConfiguredRule{pluginRule("uc/x", []any{"o1"}, rule.SeverityError)}},
	}
	var batches []EslintPluginLintRequest
	dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		batches = append(batches, req)
		results := make([]EslintPluginFileResult, 0, len(req.Files))
		for _, f := range req.Files {
			results = append(results, EslintPluginFileResult{FilePath: f.Path})
		}
		return &EslintPluginLintResult{Results: results}, nil
	}
	if err := DispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", func(rule.RuleDiagnostic) {}); err != nil {
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

func TestDispatchEslintPlugin_SeverityReattachAndClamp(t *testing.T) {
	text := "abc\n" // len 4 bytes
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityWarning)}},
	}
	dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{
				{RuleName: "uc/x", Message: "in-bounds", StartPos: 1, EndPos: 3},
				{RuleName: "uc/x", Message: "out-of-bounds", StartPos: 10, EndPos: 20},
				{RuleName: "uc/x", Message: "inverted", StartPos: 3, EndPos: 1},
			},
		}}}, nil
	}
	var diags []rule.RuleDiagnostic
	if err := DispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", func(d rule.RuleDiagnostic) {
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
		return []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	}
	run := func(fr EslintPluginFileResult) []rule.RuleDiagnostic {
		dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
			return &EslintPluginLintResult{Results: []EslintPluginFileResult{fr}}, nil
		}
		var diags []rule.RuleDiagnostic
		if err := DispatchEslintPluginRules(context.Background(), dispatch, mkFiles(), false, "off", func(d rule.RuleDiagnostic) {
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
	files := []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{FilePath: "/a.ts", Cancelled: true}}}, nil
	}
	err := DispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", func(rule.RuleDiagnostic) {})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDispatchEslintPlugin_MissingResultNoPanic(t *testing.T) {
	text := "x"
	files := []EslintPluginFileInput{{Path: "/a.ts", ConfigKey: "/cfg", Text: &text, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}}}
	dispatch := func(ctx context.Context, req EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: nil}, nil // no result for /a.ts
	}
	var diags []rule.RuleDiagnostic
	if err := DispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", func(d rule.RuleDiagnostic) {
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
	sf := newTextSourceFile(code)
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{SourceFile: sf}); !ok || got != sf {
		t.Errorf("SourceFile branch must reuse the provided frame verbatim")
	}
	// (b) LSP overlay: the frame keeps the BOM VERBATIM (it must match the
	// BOM-inclusive editor document + native frame; the worker's BOM-stripped
	// offsets are shifted back past the BOM during rebuild).
	withBOM := "\ufeff" + code
	if got, ok := eslintPluginSourceFile(EslintPluginFileInput{Text: &withBOM}); !ok || got.Text() != withBOM {
		t.Errorf("overlay frame must keep the BOM verbatim: text = %q, want %q", got.Text(), withBOM)
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
	sf := newTextSourceFile("const 变量 = 1;")
	files := []EslintPluginFileInput{
		{Path: "/a.ts", ConfigKey: "/cfg", SourceFile: sf, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch := func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    "/a.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 6, EndPos: 12}},
		}}}, nil
	}
	var diags []rule.RuleDiagnostic
	if err := DispatchEslintPluginRules(context.Background(), dispatch, files, false, "off", func(d rule.RuleDiagnostic) {
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

	// (b) LSP overlay with a leading BOM. The frame KEEPS the BOM (matching the
	// BOM-inclusive editor document + native diagnostics + the with-BOM content
	// source.fixAll splices into). The worker reports BOM-stripped offsets [6,7]
	// (the 'x'); Go shifts them +3 past the BOM into this frame \u2192 [9,10], so the
	// fix splices the 'x' (NOT the BOM). Pre-fix this rebuilt against a BOM-
	// stripped frame and corrupted source.fixAll on well-formed BOM files.
	const code = "const x = 1;"
	bomText := "\ufeff" + code
	files2 := []EslintPluginFileInput{
		{Path: "/b.ts", ConfigKey: "/cfg", Text: &bomText, Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch2 := func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath: "/b.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 6, EndPos: 7,
				Fixes: []EslintPluginFix{{Range: [2]int{6, 7}, Text: "y"}}}},
		}}}, nil
	}
	var diags2 []rule.RuleDiagnostic
	if err := DispatchEslintPluginRules(context.Background(), dispatch2, files2, true, "eager", func(d rule.RuleDiagnostic) {
		diags2 = append(diags2, d)
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags2) != 1 {
		t.Fatalf("want 1 diagnostic, got %d", len(diags2))
	}
	if diags2[0].SourceFile.Text() != bomText {
		t.Errorf("overlay frame must keep the BOM: got %q, want %q", diags2[0].SourceFile.Text(), bomText)
	}
	if diags2[0].Range.Pos() != 9 || diags2[0].Range.End() != 10 {
		t.Errorf("overlay range: worker [6,7] + BOM shift \u2192 want [9,10], got [%d,%d]", diags2[0].Range.Pos(), diags2[0].Range.End())
	}
	fx := diags2[0].Fixes()
	if len(fx) != 1 || fx[0].Range.Pos() != 9 || fx[0].Range.End() != 10 || fx[0].Text != "y" {
		t.Errorf("overlay fix: worker [6,7] + BOM shift \u2192 want [9,10]='y', got %+v", fx)
	}
	// End-to-end: applying the fix to the with-BOM overlay must replace the 'x'
	// and PRESERVE the BOM \u2014 no off-by-3 corruption.
	if fixed, _, ok := ApplyRuleFixes(bomText, diags2); !ok || fixed != "\ufeff"+"const y = 1;" {
		t.Errorf("fixAll on a BOM file must preserve the BOM + replace x\u2192y, got %q", fixed)
	}

	// (c) Neither SourceFile nor Text -> surface a 'no source frame' error.
	files3 := []EslintPluginFileInput{
		{Path: "/c.ts", ConfigKey: "/cfg", Rules: []ConfiguredRule{pluginRule("uc/x", nil, rule.SeverityError)}},
	}
	dispatch3 := func(_ context.Context, _ EslintPluginLintRequest) (*EslintPluginLintResult, error) {
		return &EslintPluginLintResult{Results: []EslintPluginFileResult{{
			FilePath:    "/c.ts",
			Diagnostics: []EslintPluginDiagnostic{{RuleName: "uc/x", Message: "m", StartPos: 0, EndPos: 1}},
		}}}, nil
	}
	var diags3 []rule.RuleDiagnostic
	if err := DispatchEslintPluginRules(context.Background(), dispatch3, files3, false, "off", func(d rule.RuleDiagnostic) {
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

func TestOptionsToArray(t *testing.T) {
	if got := optionsToArray(nil); len(got) != 0 {
		t.Errorf("nil → empty array, got %v", got)
	}
	got := optionsToArray([]any{"a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("array → passthrough preserving elements, got %v", got)
	}
	single := optionsToArray(map[string]any{"k": 1})
	if len(single) != 1 || !reflect.DeepEqual(single[0], map[string]any{"k": 1}) {
		t.Errorf("single object → wrapped 1-element array holding the object, got %v", single)
	}
	// A lone array-valued option (["error", ["a","b"]] → Options [["a","b"]])
	// must surface as context.options == [["a","b"]]: one element that is the
	// array itself, not the two strings flattened into the options list.
	nested := optionsToArray([]any{[]any{"a", "b"}})
	if len(nested) != 1 || !reflect.DeepEqual(nested[0], []any{"a", "b"}) {
		t.Errorf("single array option → context.options [[a,b]], got %v", nested)
	}
}

// TestEslintPluginWire_RoundTrip pins the wire JSON keys against the Node
// worker's plugin-lint-protocol.ts. A silent key drift (e.g. startPos →
// start) would drop data on the wire without a compile error, so assert every
// field explicitly: the request keys via marshal, the result fields via decode
// of a Node-shaped payload.
func TestEslintPluginWire_RoundTrip(t *testing.T) {
	text := "const x = 1;"
	req := EslintPluginLintRequest{
		Files: []EslintPluginLintFile{{
			Path:            "/proj/a.ts",
			Text:            &text,
			ConfigKey:       "/proj",
			LanguageOptions: map[string]any{"sourceType": "module"},
			Settings:        map[string]any{"foo": "bar"},
		}},
		Rules:           map[string]EslintPluginRuleConfig{"p/r": {Options: []any{"opt"}}},
		Fix:             true,
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
	for _, k := range []string{"files", "rules", "fix", "suggestionsMode"} {
		if _, ok := got[k]; !ok {
			t.Errorf("request JSON missing top-level key %q", k)
		}
	}
	f0 := got["files"].([]any)[0].(map[string]any)
	for _, k := range []string{"path", "text", "configKey", "languageOptions", "settings"} {
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
