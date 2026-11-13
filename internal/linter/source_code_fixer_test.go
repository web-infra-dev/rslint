package linter

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/testutil/directivefixtures"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// mockDiagnostic implements LintMessage interface for testing
type mockDiagnostic struct {
	fixes []rule.RuleFix
}

func (m mockDiagnostic) Fixes() []rule.RuleFix {
	return m.fixes
}

func newMockDiagnostic(fixes ...rule.RuleFix) mockDiagnostic {
	return mockDiagnostic{fixes: fixes}
}

func newInsertFix(pos int, text string) rule.RuleFix {
	return rule.RuleFix{
		Text:  text,
		Range: core.NewTextRange(pos, pos),
	}
}

func newReplaceFix(start, end int, text string) rule.RuleFix {
	return rule.RuleFix{
		Text:  text,
		Range: core.NewTextRange(start, end),
	}
}

func TestApplyRuleFixes(t *testing.T) {
	tests := []struct {
		name                string
		code                string
		diagnostics         []mockDiagnostic
		expectedCode        string
		expectedUnapplied   int
		expectedFixedStatus bool
	}{
		{
			name: "single insertion fix",
			code: "function foo() {}",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async ")),
			},
			expectedCode:        "async function foo() {}",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "multiple non-overlapping insertion fixes",
			code: "function foo() { return bar(); }",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async ")),
				newMockDiagnostic(newInsertFix(24, "await ")),
			},
			expectedCode:        "async function foo() { return await bar(); }",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "duplicate insertion fixes at same position - should skip duplicates",
			code: "function foo() {}",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async ")),
				newMockDiagnostic(newInsertFix(0, "async ")),
				newMockDiagnostic(newInsertFix(0, "async ")),
			},
			expectedCode:        "async function foo() {}",
			expectedUnapplied:   2,
			expectedFixedStatus: true,
		},
		{
			name: "overlapping replacement fixes - should skip overlapping",
			code: "const foo = 'bar';",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(6, 9, "baz")),
				newMockDiagnostic(newReplaceFix(8, 11, "qux")),
			},
			expectedCode:        "const baz = 'bar';",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			name: "consecutive replacement fixes - should apply both",
			code: "const foo = bar;",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(6, 9, "baz")),
				newMockDiagnostic(newReplaceFix(12, 15, "qux")),
			},
			expectedCode:        "const baz = qux;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "adjacent replacement fixes (end == start) - should apply both",
			code: "abcdef",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(0, 3, "XXX")),
				newMockDiagnostic(newReplaceFix(3, 6, "YYY")),
			},
			expectedCode:        "XXXYYY",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name:                "no fixes",
			code:                "const foo = 'bar';",
			diagnostics:         []mockDiagnostic{},
			expectedCode:        "const foo = 'bar';",
			expectedUnapplied:   0,
			expectedFixedStatus: false,
		},
		{
			name: "diagnostic without fixes",
			code: "const foo = 'bar';",
			diagnostics: []mockDiagnostic{
				{fixes: []rule.RuleFix{}},
			},
			expectedCode:        "const foo = 'bar';",
			expectedUnapplied:   1,
			expectedFixedStatus: false,
		},
		{
			name: "mixed: insertion at position 0 followed by another at same position",
			code: "foo()",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "await ")),
				newMockDiagnostic(newInsertFix(0, "await ")),
			},
			expectedCode:        "await foo()",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			name: "insertion followed by replacement at different positions",
			code: "function foo() { return bar; }",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async ")),
				newMockDiagnostic(newReplaceFix(24, 27, "baz")),
			},
			expectedCode:        "async function foo() { return baz; }",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		// Tests for diagnostics with multiple fixes
		{
			name: "single diagnostic with multiple fixes at different positions",
			code: "function foo() { return bar(); }",
			diagnostics: []mockDiagnostic{
				// One diagnostic with two fixes: add async and add await
				newMockDiagnostic(newInsertFix(0, "async "), newInsertFix(24, "await ")),
			},
			expectedCode:        "async function foo() { return await bar(); }",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "two diagnostics with multiple fixes - second conflicts with first",
			code: "function foo() { return bar(); }",
			diagnostics: []mockDiagnostic{
				// First diagnostic: add async at 0, add await at 24
				newMockDiagnostic(newInsertFix(0, "async "), newInsertFix(24, "await ")),
				// Second diagnostic: tries to insert at position 10 (between first diagnostic's fixes)
				newMockDiagnostic(newInsertFix(10, "CONFLICT")),
			},
			expectedCode:        "async function foo() { return await bar(); }",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			name: "duplicate diagnostic with multiple fixes - should skip entire duplicate",
			code: "function foo() { return bar(); }",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async "), newInsertFix(24, "await ")),
				// Duplicate: same first fix position
				newMockDiagnostic(newInsertFix(0, "async "), newInsertFix(24, "await ")),
			},
			expectedCode:        "async function foo() { return await bar(); }",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		// Edge cases
		{
			name: "delete operation (empty replacement text)",
			code: "var foo = 1;",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(0, 4, "")), // delete "var "
			},
			expectedCode:        "foo = 1;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "insertion at end of code",
			code: "const x = 1",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(11, ";")),
			},
			expectedCode:        "const x = 1;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "replacement at end of code",
			code: "const x = 1",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(10, 11, "2;")),
			},
			expectedCode:        "const x = 2;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "empty code with insertion",
			code: "",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "hello")),
			},
			expectedCode:        "hello",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "different insertions at same position - should skip second",
			code: "function foo() {}",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "async ")),
				newMockDiagnostic(newInsertFix(0, "export ")), // different content, same position
			},
			expectedCode:        "async function foo() {}",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			name: "replacement followed by insertion at same end position - should skip insertion",
			code: "ABCDEFGHIJ",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(0, 5, "XXX")), // Replace ABCDE with XXX
				newMockDiagnostic(newInsertFix(5, "YYY")),     // Insert at position 5
			},
			expectedCode:        "XXXFGHIJ",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			name: "insertion followed by replacement starting at same position - should skip replacement",
			code: "ABCDEFGHIJ",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(0, "XXX")),     // Insert at 0
				newMockDiagnostic(newReplaceFix(0, 3, "YYY")), // Replace ABC with YYY
			},
			expectedCode:        "XXXABCDEFGHIJ",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
		},
		{
			// ESLint's unicode-bom fix: a range starting one position ahead of
			// the text, where the mark lives.
			name: "fix reaching before the text removes the byte order mark",
			code: "\uFEFFvar a = 123;",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(-1, 0, "")),
			},
			expectedCode:        "var a = 123;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "a fix elsewhere leaves the byte order mark in place",
			code: "\uFEFFconst x = 1",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newInsertFix(11, ";")),
			},
			expectedCode:        "\uFEFFconst x = 1;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "the mark is dropped once and the other fixes still land",
			code: "\uFEFFconst foo = bar",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(-1, 0, "")),
				newMockDiagnostic(newReplaceFix(6, 9, "baz")),
				newMockDiagnostic(newInsertFix(15, ";")),
			},
			expectedCode:        "const baz = bar;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "removing an absent mark changes nothing",
			code: "var a = 123;",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(-1, 0, "")),
			},
			expectedCode:        "var a = 123;",
			expectedUnapplied:   0,
			expectedFixedStatus: true,
		},
		{
			name: "inverted range is left for a later pass",
			code: "var a = 123;",
			diagnostics: []mockDiagnostic{
				newMockDiagnostic(newReplaceFix(9, 4, "x")),
			},
			expectedCode:        "var a = 123;",
			expectedUnapplied:   1,
			expectedFixedStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, unapplied, fixed := ApplyRuleFixes(tt.code, tt.diagnostics)

			if result != tt.expectedCode {
				t.Errorf("ApplyRuleFixes() code = %q, want %q", result, tt.expectedCode)
			}

			if len(unapplied) != tt.expectedUnapplied {
				t.Errorf("ApplyRuleFixes() unapplied count = %d, want %d", len(unapplied), tt.expectedUnapplied)
			}

			if fixed != tt.expectedFixedStatus {
				t.Errorf("ApplyRuleFixes() fixed = %v, want %v", fixed, tt.expectedFixedStatus)
			}
		})
	}
}

func TestApplyRuleFixes_UnsortedDiagnostics(t *testing.T) {
	// Diagnostics provided in reverse position order — should be sorted internally
	code := "aaa bbb ccc"
	diagnostics := []mockDiagnostic{
		newMockDiagnostic(newReplaceFix(8, 11, "CCC")), // last
		newMockDiagnostic(newReplaceFix(0, 3, "AAA")),  // first
		newMockDiagnostic(newReplaceFix(4, 7, "BBB")),  // middle
	}
	result, unapplied, fixed := ApplyRuleFixes(code, diagnostics)
	if result != "AAA BBB CCC" {
		t.Errorf("got %q, want %q", result, "AAA BBB CCC")
	}
	if len(unapplied) != 0 {
		t.Errorf("unapplied = %d, want 0", len(unapplied))
	}
	if !fixed {
		t.Error("expected fixed = true")
	}
}

func TestApplyRuleFixes_UnsortedFixesWithinDiagnostic(t *testing.T) {
	// Single diagnostic with multiple fixes in reverse order
	code := "function foo() { return bar(); }"
	diagnostics := []mockDiagnostic{
		// Fixes given in reverse order: await at 24, async at 0
		newMockDiagnostic(newInsertFix(24, "await "), newInsertFix(0, "async ")),
	}
	result, _, fixed := ApplyRuleFixes(code, diagnostics)
	expected := "async function foo() { return await bar(); }"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
	if !fixed {
		t.Error("expected fixed = true")
	}
}

func TestApplyRuleFixes_ReplaceEntireContent(t *testing.T) {
	code := "old content here"
	diagnostics := []mockDiagnostic{
		newMockDiagnostic(newReplaceFix(0, len(code), "new content")),
	}
	result, unapplied, fixed := ApplyRuleFixes(code, diagnostics)
	if result != "new content" {
		t.Errorf("got %q, want %q", result, "new content")
	}
	if len(unapplied) != 0 {
		t.Errorf("unapplied = %d, want 0", len(unapplied))
	}
	if !fixed {
		t.Error("expected fixed = true")
	}
}

func TestApplyRuleFixes_MixedWithAndWithoutFixes(t *testing.T) {
	code := "var x = 1;"
	diagnostics := []mockDiagnostic{
		{fixes: []rule.RuleFix{}},                       // no fix → unapplied
		newMockDiagnostic(newReplaceFix(0, 3, "const")), // has fix → applied
		{fixes: []rule.RuleFix{}},                       // no fix → unapplied
	}
	result, unapplied, fixed := ApplyRuleFixes(code, diagnostics)
	if result != "const x = 1;" {
		t.Errorf("got %q, want %q", result, "const x = 1;")
	}
	if len(unapplied) != 2 {
		t.Errorf("unapplied = %d, want 2", len(unapplied))
	}
	if !fixed {
		t.Error("expected fixed = true")
	}
}

// TestApplyRuleFixes_DuplicateInsertionRegression is a regression test for issue #451
// where multiple diagnostics at the same position caused "async async async" to be inserted
func TestApplyRuleFixes_DuplicateInsertionRegression(t *testing.T) {
	// Simulate what happens when the same file is linted by multiple tsconfig projects
	// Each project reports the same diagnostic, resulting in duplicate fixes
	code := "function fetchData(): Promise<string> { return Promise.resolve('data'); }"

	// Three diagnostics from three different tsconfig projects, all pointing to the same fix
	diagnostics := []mockDiagnostic{
		newMockDiagnostic(newInsertFix(0, " async ")),
		newMockDiagnostic(newInsertFix(0, " async ")),
		newMockDiagnostic(newInsertFix(0, " async ")),
	}

	result, unapplied, fixed := ApplyRuleFixes(code, diagnostics)

	expectedCode := " async function fetchData(): Promise<string> { return Promise.resolve('data'); }"

	if result != expectedCode {
		t.Errorf("Regression test failed: got %q, want %q", result, expectedCode)
	}

	if len(unapplied) != 2 {
		t.Errorf("Expected 2 unapplied diagnostics (duplicates), got %d", len(unapplied))
	}

	if !fixed {
		t.Error("Expected fixed to be true")
	}

	// Verify we don't have "async async async"
	if result == " async  async  async function fetchData(): Promise<string> { return Promise.resolve('data'); }" {
		t.Error("Regression: duplicate insertions were not prevented!")
	}
}

func TestUnusedDirectiveESLintFixOutputs(t *testing.T) {
	for _, file := range []string{"upstream.json", "source.json"} {
		corpus := directivefixtures.Load(t, "../rule/testdata/eslint-directives/"+file)
		for index, tc := range corpus.Cases {
			t.Run(fmt.Sprintf("%s/%03d", file, index), func(t *testing.T) {
				for _, prefix := range []string{"eslint", "rslint"} {
					source := tc.Source
					if prefix == "rslint" {
						source = tc.RslintSource
					}
					sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/conformance.js", Path: "/conformance.js"}, source, core.ScriptKindJS)
					for _, v := range tc.Variants {
						t.Run(fmt.Sprintf("%s/%s/noFix=%v", prefix, v.Severity, v.DisableFixes), func(t *testing.T) {
							manager := rule.NewDisableManager(sf, rule.NewCommentStore(sf))
							invalid := map[string]bool{}
							for _, name := range tc.InvalidRules {
								invalid[name] = true
							}
							manager.ValidateRules(func(name string) bool { return !invalid[name] })
							var diagnostics []rule.RuleDiagnostic
							for _, report := range tc.Reports {
								suppressed := manager.IsRuleDisabled(report.RuleID, directivefixtures.ByteOffset(t, source, report.Pos))
								if suppressed != report.Suppressed {
									t.Fatalf("%s: unexpected suppression for %s", tc.Name, report.RuleID)
								}
								if !suppressed && report.Fix != nil {
									offset := func(pos int) int {
										if pos == -1 {
											return -1
										}
										return directivefixtures.ByteOffset(t, source, pos)
									}
									fixes := []rule.RuleFix{{Range: core.NewTextRange(offset(report.Fix.Range[0]), offset(report.Fix.Range[1])), Text: report.Fix.Text}}
									diagnostics = append(diagnostics, rule.RuleDiagnostic{FixesPtr: &fixes})
								}
							}
							ignored := map[string]bool{}
							for _, name := range tc.Ignored {
								ignored[name] = true
							}
							demand := rule.EditDemandAutofix
							if v.DisableFixes {
								demand = 0
							}
							// Retain ordinary-rule oracle fixes, then adapt only the
							// original directive replacements that delete line breaks.
							expectedDiagnostics := append([]rule.RuleDiagnostic(nil), diagnostics...)
							linePreserving := false
							for _, diagnostic := range v.Diagnostics {
								if diagnostic.Fix == nil {
									continue
								}
								fix := directivefixtures.LinePreservingFix(t, source, *diagnostic.Fix)
								linePreserving = linePreserving || fix.Text != diagnostic.Fix.Text
								fixes := []rule.RuleFix{{Range: core.NewTextRange(directivefixtures.ByteOffset(t, source, fix.Range[0]), directivefixtures.ByteOffset(t, source, fix.Range[1])), Text: fix.Text}}
								expectedDiagnostics = append(expectedDiagnostics, rule.RuleDiagnostic{FixesPtr: &fixes})
							}
							diagnostics = append(diagnostics, manager.UnusedDiagnostics("/conformance.js", rule.ParseSeverity(v.Severity), demand, ignored)...)
							code := source
							if tc.BOM {
								code = "\ufeff" + code
							}
							fixed, _, _ := ApplyRuleFixes(code, diagnostics)
							want := v.Fixed
							if prefix == "rslint" {
								want = v.RslintFixed
							}
							// ESLint's later-round output is recorded only for adjacent deletions.
							batched := v.BatchedFixed
							if prefix == "rslint" {
								batched = v.RslintBatchedFixed
							}
							if linePreserving {
								want, _, _ = ApplyRuleFixes(code, expectedDiagnostics)
								batched = nil
							}
							if fixed != want && (batched == nil || fixed != *batched) {
								t.Fatalf("%s\nfixed: %q\nwant:  %q", tc.Name, fixed, want)
							}
						})
					}
				}
			})
		}
	}
}

func TestUnusedDirectiveAdjacentLineBreaks(t *testing.T) {
	for _, prefix := range []string{"eslint", "rslint"} {
		for _, tc := range []struct {
			name, source, fixed string
			usedRules           []string
		}{
			{"separate CR LF", "/* %[1]s-disable a\r*//* %[1]s-disable b\n*/x();", "\r  \nx();", nil},
			{"original CRLF", "/* %[1]s-disable a\r\n*//* %[1]s-disable b\r\n*/x();", "\r\n\r\nx();", nil},
			{"LF CR", "/* %[1]s-disable a\n*//* %[1]s-disable b\r*/x();", " \n\r x();", nil},
			{"partial separate CR LF", "/* %[1]s-disable a, b\r, c\n, d */x();", "/* %[1]s-disable a\r \n, d */x();", []string{"a", "d"}},
		} {
			t.Run(prefix+"/"+tc.name, func(t *testing.T) {
				source := fmt.Sprintf(tc.source, prefix)
				expected := tc.fixed
				if len(tc.usedRules) != 0 {
					expected = fmt.Sprintf(expected, prefix)
				}
				sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/adjacent.js", Path: "/adjacent.js"}, source, core.ScriptKindJS)
				manager := rule.NewDisableManager(sf, rule.NewCommentStore(sf))
				for _, name := range tc.usedRules {
					if !manager.IsRuleDisabled(name, strings.Index(source, "x();")) {
						t.Fatalf("expected %s to be suppressed", name)
					}
				}
				diagnostics := manager.UnusedDiagnostics("/adjacent.js", rule.SeverityWarning, rule.EditDemandAutofix, nil)
				if len(diagnostics) != 2 {
					t.Fatalf("got %d diagnostics, want 2", len(diagnostics))
				}
				fixed, unapplied, applied := ApplyRuleFixes(source, diagnostics)
				if fixed != expected || !applied || len(unapplied) != 0 {
					t.Fatalf("fixed = %q, want %q; applied = %v, unapplied = %d", fixed, expected, applied, len(unapplied))
				}
			})
		}
	}
}
