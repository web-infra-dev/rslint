package no_unused_labels

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoUnusedLabelsExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row it covers, so future refactors can't
// silently regress them without breaking a named lock-in.
func TestNoUnusedLabelsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedLabelsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `A: { function f() { B: { break B; } } break A; }`},
			// ---- Dimension 4: declaration/container forms ----
			{Code: `A: { class C { method() { B: { break B; } } } break A; }`},
			// ---- Dimension 4: declaration/container forms ----
			{Code: `A: { const f = () => { B: { break B; } }; break A; }`},
			// ---- Dimension 4: declaration/container forms ----
			{Code: `async function* f() { A: { break A; } }`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `class C { static { A: { break A; } } }`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `A: { B: { C: { if (ok) break C; } break B; } break A; }`},
			// ---- Dimension 4: nesting/traversal boundary ----
			{Code: `A: switch (kind) { case 1: while (ok) break A; default: break; }`},
			// ---- Dimension 4: declaration/container forms ----
			{Code: `A: for (const x of xs) { switch (x.kind) { case "skip": continue A; case "stop": break A; } }`},
			// ---- Dimension 4: access/key forms are non-label object keys ----
			{Code: `const obj = { A: foo, "B": bar, 0: baz, [key]: value, ...rest };`},
			// ---- Real-user: eslint/eslint#5052 nested loop label used from inner loop ----
			{Code: `B: for (let i = 0; i < 10; ++i) { for (let j = 0; j < 10; ++j) { if (foo()) { break B; } } bar(); }`},
			// ---- Real-user: eslint/eslint#14191 object expression in arrow return is not labels ----
			{Code: `const get = () => Promise.resolve({ json: () => ({ billing_contact: { id: "54321" } }) });`},
			// N/A: receiver/access/key expression wrappers do not apply; labels are identifiers, not member/key expressions.
			// N/A: overload/abstract/declare members cannot be labeled statements.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream markAsUsed() arm 1: unlabeled break/continue are ignored.
			{
				Code:   `A: while (a) { break; continue; }`,
				Output: []string{`while (a) { break; continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// Locks in upstream markAsUsed() arm 3: unmatched labels do not mark another label as used.
			{
				Code:   `A: while (a) { while (b) { break; } }`,
				Output: []string{`while (a) { while (b) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// Locks in upstream exitLabeledScope() arm 1: unused labels report with exact message text.
			{
				Code:   `A: foo();`,
				Output: []string{`foo();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unused",
						Message:   "'A:' is defined but never used.",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 2,
					},
				},
			},
			// Locks in upstream markAsUsed() arm 2: continue marks a matching outer loop label as used.
			{
				Code:   `A: for (const x of xs) { B: { if (x) continue A; } }`,
				Output: []string{`A: for (const x of xs) { { if (x) continue A; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
			// Locks in upstream isFixable() arm 1: comments between label and body suppress autofix.
			{
				Code: `A
/* comment */: foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// Locks in upstream isFixable() arm 1: comments after the colon also suppress autofix.
			{
				Code: `A: // comment
foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// Locks in upstream isFixable() arm 2: ordinary expression labels are fixable.
			{
				Code:   "A:\nfoo();",
				Output: []string{"foo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: type-expression wrappers in the labeled body ----
			{
				Code:   `A: (foo as any);`,
				Output: []string{`(foo as any);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: type-expression wrappers in the labeled body ----
			{
				Code:   `A: (foo satisfies T);`,
				Output: []string{`(foo satisfies T);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: expression wrappers in the labeled body ----
			{
				Code:   `A: foo!;`,
				Output: []string{`foo!;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: optional-chain expression in the labeled body ----
			{
				Code:   `A: foo?.();`,
				Output: []string{`foo?.();`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   `A: switch (kind) { case 1: break; default: break; }`,
				Output: []string{`switch (kind) { case 1: break; default: break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: nesting/traversal boundary ----
			{
				Code:   `A: { B: { break A; } }`,
				Output: []string{`A: { { break A; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
				},
			},
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   `A: { function f() {} }`,
				Output: []string{`{ function f() {} }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: declaration/container forms ----
			{
				Code:   `class C { static { A: foo(); } }`,
				Output: []string{`class C { static { foo(); } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},
			// ---- Dimension 4: Unicode label and UTF-16 positions ----
			{
				Code:   "const face = \"😀\";\n中文: foo();",
				Output: []string{"const face = \"😀\";\nfoo();"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Message: "'中文:' is defined but never used.", Line: 2, Column: 1, EndLine: 2, EndColumn: 3},
				},
			},
			// ---- Dimension 4: object spread in expression body must not be confused with labeled object keys ----
			{
				Code:   `A: ({ ...obj, value: 1 });`,
				Output: []string{`({ ...obj, value: 1 });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: graceful degradation ----
			{
				Code:   `A: {}`,
				Output: []string{`{}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Dimension 4: graceful degradation ----
			{
				Code:   `A: ;`,
				Output: []string{`;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 1, EndLine: 1, EndColumn: 2},
				},
			},
			// ---- Real-user: eslint/eslint#16988 function-body directive creation ----
			{
				Code: `function test3() {
    NoDirective: "use strict";
    return 7;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 2, Column: 5, EndLine: 2, EndColumn: 16},
				},
			},
			// ---- Real-user: eslint/eslint#16988 parenthesized directive candidate ----
			{
				Code: `function test3() {
    NoDirective: ("use strict");
    return 7;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 2, Column: 5, EndLine: 2, EndColumn: 16},
				},
			},
			// Locks in upstream isFixable() directive branch: conservative even after a previous statement.
			{
				Code: `function test3() {
    foo();
    NoDirective: "use strict";
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 3, Column: 5, EndLine: 3, EndColumn: 16},
				},
			},
			// Locks in upstream isFixable() directive branch: template expressions with substitutions are fixable.
			{
				Code:   "function test3() { NoDirective: `use ${mode}`; }",
				Output: []string{"function test3() { `use ${mode}`; }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 20, EndLine: 1, EndColumn: 31},
				},
			},
			// Locks in upstream isFixable() ancestor branch: a string label in a non-function block is fixable.
			{
				Code:   `if (foo) { bar: "baz" }`,
				Output: []string{`if (foo) { "baz" }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 12, EndLine: 1, EndColumn: 15},
				},
			},
			// ---- Real-user: eslint/eslint#14191 arrow block body with labels, not an object expression ----
			{
				Code:   `const get = () => Promise.resolve({ json: () => { billing_contact: { id: "54321" } } });`,
				Output: []string{`const get = () => Promise.resolve({ json: () => { { "54321" } } });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unused", Line: 1, Column: 70, EndLine: 1, EndColumn: 72},
					{MessageId: "unused", Line: 1, Column: 51, EndLine: 1, EndColumn: 66},
				},
			},
		},
	)
}
