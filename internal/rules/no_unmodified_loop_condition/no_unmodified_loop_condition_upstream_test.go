// TestNoUnmodifiedLoopConditionUpstream migrates the full valid/invalid suite
// from ESLint main at 9ef407a3b051e74f50dc7fb8914e2bd89b3e5e53 1:1.
// rslint-specific branch and edge-shape cases live in
// no_unmodified_loop_condition_extras_test.go.
package no_unmodified_loop_condition

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnmodifiedLoopConditionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnmodifiedLoopConditionRule,
		[]rule_tester.ValidTestCase{
			{Code: `var foo = 0; while (foo) { ++foo; }`},
			{Code: `let foo = 0; while (foo) { ++foo; }`},
			{Code: `var foo = 0; while (foo) { foo += 1; }`},
			{Code: `var foo = 0; while (foo++) { }`},
			{Code: `var foo = 0; while (foo = next()) { }`},
			{Code: `var foo = 0; while (ok(foo)) { }`},
			{Code: `var foo = 0, bar = 0; while (++foo < bar) { }`},
			{Code: `var foo = 0, obj = {}; while (foo === obj.bar) { }`},
			{Code: `var foo = 0, f = {}, bar = {}; while (foo === f(bar)) { }`},
			{Code: `var foo = 0, f = {}; while (foo === f()) { }`},
			{Code: "var foo = 0, tag = 0; while (foo === tag`abc`) { }"},
			{Code: `function* foo() { var foo = 0; while (yield foo) { } }`},
			{Code: `function* foo() { var foo = 0; while (foo === (yield)) { } }`},
			{Code: `var foo = 0; while (foo.ok) { }`},
			{Code: `var foo = 0; while (foo) { update(); } function update() { ++foo; }`},
			{Code: `var foo = 0, bar = 9; while (foo < bar) { foo += 1; }`},
			{Code: `var foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; }`},
			{Code: `var foo = 0, bar = 0; while (foo && bar) { ++foo; ++bar; }`},
			{Code: `var foo = 0, bar = 0; while (foo || bar) { ++foo; ++bar; }`},
			{Code: `var foo = 0; do { ++foo; } while (foo);`},
			{Code: `var foo = 0; do { } while (foo++);`},
			{Code: `for (var foo = 0; foo; ++foo) { }`},
			{Code: `for (var foo = 0; foo;) { ++foo }`},
			{Code: `var foo = 0, bar = 0; for (bar; foo;) { ++foo }`},
			{Code: `var foo; if (foo) { }`},
			{Code: `var a = [1, 2, 3]; var len = a.length; for (var i = 0; i < len - 1; i++) {}`},
			{
				Code:    `let foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; bar += 1; baz += 1; }`,
				Options: map[string]any{"checkConditionalExpressions": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var foo = 0; while (foo) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `var foo = 0; while (!foo) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    22,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			{
				Code: `var foo = 0; while (foo != null) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `var foo = 0, bar = 9; while (foo < bar) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "loopConditionNotModified",
						Message:   "'foo' is not modified in this loop.",
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 33,
					},
					{
						MessageId: "loopConditionNotModified",
						Message:   "'bar' is not modified in this loop.",
						Line:      1,
						Column:    36,
						EndLine:   1,
						EndColumn: 39,
					},
				},
			},
			{
				Code: `var foo = 0, bar = 0; while (foo && bar) { ++bar; } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code: `var foo = 0, bar = 0; while (foo && bar) { ++foo; } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'bar' is not modified in this loop.",
					Line:      1,
					Column:    37,
					EndLine:   1,
					EndColumn: 40,
				}},
			},
			{
				Code: `var a, b, c; while (a < c && b < c) { ++a; } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "loopConditionNotModified",
						Message:   "'b' is not modified in this loop.",
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 31,
					},
					{
						MessageId: "loopConditionNotModified",
						Message:   "'c' is not modified in this loop.",
						Line:      1,
						Column:    34,
						EndLine:   1,
						EndColumn: 35,
					},
				},
			},
			{
				Code: `var foo = 0; while (foo ? 1 : 0) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `var foo = 0; while (foo) { update(); } function update(foo) { ++foo; }`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    21,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			{
				Code: `var foo; do { } while (foo);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 27,
				}},
			},
			{
				Code: `for (var foo = 0; foo < 10; ) { } foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'foo' is not modified in this loop.",
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 22,
				}},
			},
			{
				Code:    `let foo = 0, bar = 1, baz = 2; while (foo ? bar : baz) { foo += 1; }`,
				Options: map[string]any{"checkConditionalExpressions": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "loopConditionNotModified",
						Message:   "'bar' is not modified in this loop.",
						Line:      1,
						Column:    45,
						EndLine:   1,
						EndColumn: 48,
					},
					{
						MessageId: "loopConditionNotModified",
						Message:   "'baz' is not modified in this loop.",
						Line:      1,
						Column:    51,
						EndLine:   1,
						EndColumn: 54,
					},
				},
			},
			{
				Code:    `let chunk = true, done = false; while (chunk ? !done : false) { chunk = nextOrNull(); }`,
				Options: map[string]any{"checkConditionalExpressions": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "loopConditionNotModified",
					Message:   "'done' is not modified in this loop.",
					Line:      1,
					Column:    49,
					EndLine:   1,
					EndColumn: 53,
				}},
			},
		},
	)
}
