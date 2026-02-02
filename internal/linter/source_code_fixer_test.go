package linter

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/core"
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
				newMockDiagnostic(newInsertFix(0, "XXX")),      // Insert at 0
				newMockDiagnostic(newReplaceFix(0, 3, "YYY")),  // Replace ABC with YYY
			},
			expectedCode:        "XXXABCDEFGHIJ",
			expectedUnapplied:   1,
			expectedFixedStatus: true,
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
