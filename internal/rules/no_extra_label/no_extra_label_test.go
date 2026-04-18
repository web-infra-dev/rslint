package no_extra_label

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExtraLabelRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExtraLabelRule,
		[]rule_tester.ValidTestCase{
			// =================================================================
			// Upstream ESLint valid suite — mirror 1:1
			// =================================================================
			{Code: `A: break A;`},
			{Code: `A: { if (a) break A; }`},
			{Code: `A: { while (b) { break A; } }`},
			{Code: `A: { switch (b) { case 0: break A; } }`},
			{Code: `A: while (a) { while (b) { break; } break; }`},
			{Code: `A: while (a) { while (b) { break A; } }`},
			{Code: `A: while (a) { while (b) { continue A; } }`},
			{Code: `A: while (a) { switch (b) { case 0: break A; } }`},
			{Code: `A: while (a) { switch (b) { case 0: continue A; } }`},
			{Code: `A: switch (a) { case 0: while (b) { break A; } }`},
			{Code: `A: switch (a) { case 0: switch (b) { case 0: break A; } }`},
			{Code: `A: for (;;) { while (b) { break A; } }`},
			{Code: `A: do { switch (b) { case 0: break A; break; } } while (a);`},
			{Code: `A: for (a in obj) { while (b) { break A; } }`},
			{Code: `A: for (a of ary) { switch (b) { case 0: break A; } }`},

			// =================================================================
			// Naked break / continue (no label at all) — rule must ignore
			// =================================================================
			{Code: `while (a) { break; }`},
			{Code: `while (a) { continue; }`},
			{Code: `do { break; } while (a);`},
			{Code: `for (;;) { break; continue; }`},
			{Code: `for (const x in obj) { continue; }`},
			{Code: `for (const x of arr) { continue; }`},
			{Code: `switch (a) { case 0: break; default: break; }`},

			// =================================================================
			// Labels on non-breakable bodies — break/continue targeting them
			// carries information (naked break would hit a different target)
			// =================================================================
			{Code: `A: if (x) { if (y) break A; }`},
			{Code: `A: { if (y) break A; }`},
			{Code: `A: { for (;;) { break A; } }`},
			{Code: `A: { do { break A; } while (x); }`},
			{Code: `A: { switch (y) { case 0: break A; } }`},

			// =================================================================
			// Chained labels on a breakable: only the innermost (directly
			// labeling) label is redundant. `break A` in `A: B: while` still
			// jumps past A (outer of B), which naked break wouldn't do.
			// =================================================================
			{Code: `A: B: while (true) { break A; }`},
			{Code: `A: B: while (true) { continue A; }`},
			{Code: `A: B: C: while (true) { break A; continue B; }`},

			// =================================================================
			// Deep nesting — outer label justified by depth
			// =================================================================
			{Code: `A: while (a) { while (b) { while (c) { break A; } } }`},
			{Code: `A: for (;;) { for (;;) { for (;;) { continue A; } } }`},
			{Code: `A: while (a) { while (b) { while (c) { while (d) { break A; } } } }`},

			// =================================================================
			// Real-world: matrix search, tokenizer state machine
			// =================================================================
			{Code: `outer: for (let i = 0; i < 10; i++) { for (let j = 0; j < 10; j++) { if (i === j) { break outer; } } }`},
			{Code: `scan: for (const t of ary) { switch (t) { case 1: break; case 2: break scan; } }`},

			// =================================================================
			// Mix of labeled loops, switches, blocks — every break/continue
			// lands on an ancestor that is NOT its direct label target
			// =================================================================
			{Code: `A: while (a) { B: { C: for (;;) { break A; } } }`},
			{Code: `A: while (a) { B: { while (x) { if (y) break B; else continue A; } } }`},

			// =================================================================
			// Multi-label cross-targeting where only non-direct labels are
			// used (direct labels would be redundant and move to invalid)
			// =================================================================
			{Code: `A: while (a) { B: while (b) { break A; } }`},
			{Code: `A: for (;;) { B: for (;;) { C: for (;;) { break A; continue B; } } }`},

			// =================================================================
			// for-in / for-of with continue to outer loop
			// =================================================================
			{Code: `A: for (const k in obj) { for (const v of ary) { if (k === v) continue A; } }`},
			{Code: `A: for (const x of ary) { for (;;) { break A; } }`},

			// =================================================================
			// Labels inside a function / arrow / method — the function body
			// creates a breakable-unfriendly boundary. The rule tracks scope
			// without resetting on function entry, but since the innermost
			// breakable always intercepts naked break first, outer labels
			// don't produce false positives.
			// =================================================================
			{Code: `A: while (a) { function f() { while (b) { break; } } }`},
			{Code: `A: while (a) { const f = () => { while (b) { break; } }; }`},
			{Code: `A: while (a) { class C { m() { while (b) { break; } } } }`},

			// =================================================================
			// Async / generator / class wrappers — all preserve scope
			// correctness; labels still necessary
			// =================================================================
			{Code: `A: while (a) { async function f() { while (b) { break; } } }`},
			{Code: `A: while (a) { function* g() { while (b) { yield 1; break; } } }`},

			// =================================================================
			// TypeScript-specific: type annotations / generics / assertions
			// =================================================================
			{Code: `A: for (const x of (ary as number[])) { for (const y of ary) { if (x > y) break A; } }`},
			{Code: `function f<T>(xs: T[]) { A: for (const x of xs) { for (const y of xs) { if (x === y) break A; } } }`},

			// =================================================================
			// Break/continue whose label doesn't appear in the chain — rule
			// walks up, finds no match on the breakable's label, returns
			// without reporting (validity gated by parser, not this rule)
			// =================================================================
			{Code: `A: while (a) { while (b) { break; } }`},

			// =================================================================
			// Empty bodies / no break at all — silent
			// =================================================================
			{Code: `A: while (a) {}`},
			{Code: `A: {}`},
			{Code: `A: ;`},
		},
		[]rule_tester.InvalidTestCase{
			// =================================================================
			// Upstream ESLint invalid suite — mirror 1:1
			// =================================================================
			{
				Code:   `A: while (a) break A;`,
				Output: []string{`A: while (a) break;`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "This label 'A' is unnecessary.",
						Line:      1,
						Column:    20,
						EndLine:   1,
						EndColumn: 21,
					},
				},
			},
			{
				Code:   `A: while (a) { B: { continue A; } }`,
				Output: []string{`A: while (a) { B: { continue; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    30,
						EndLine:   1,
						EndColumn: 31,
					},
				},
			},
			{
				Code:   `X: while (x) { A: while (a) { B: { break A; break B; continue X; } } }`,
				Output: []string{`X: while (x) { A: while (a) { B: { break; break B; continue X; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 42},
				},
			},
			{
				Code:   `A: do { break A; } while (a);`,
				Output: []string{`A: do { break; } while (a);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			{
				Code:   `A: for (;;) { break A; }`,
				Output: []string{`A: for (;;) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code:   `A: for (a in obj) { break A; }`,
				Output: []string{`A: for (a in obj) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},
			{
				Code:   `A: for (a of ary) { break A; }`,
				Output: []string{`A: for (a of ary) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},
			{
				Code:   `A: switch (a) { case 0: break A; }`,
				Output: []string{`A: switch (a) { case 0: break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 31},
				},
			},
			{
				Code:   `X: while (x) { A: switch (a) { case 0: break A; } }`,
				Output: []string{`X: while (x) { A: switch (a) { case 0: break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 46},
				},
			},
			{
				Code:   `X: switch (a) { case 0: A: while (b) break A; }`,
				Output: []string{`X: switch (a) { case 0: A: while (b) break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 44},
				},
			},
			// Multi-line: only the outer `break A` is unnecessary; the inner
			// one targets the outer loop past the unlabeled inner loop.
			{
				Code: "                A: while (true) {\n" +
					"                    break A;\n" +
					"                    while (true) {\n" +
					"                        break A;\n" +
					"                    }\n" +
					"                }\n" +
					"            ",
				Output: []string{
					"                A: while (true) {\n" +
						"                    break;\n" +
						"                    while (true) {\n" +
						"                        break A;\n" +
						"                    }\n" +
						"                }\n" +
						"            ",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      2,
						Column:    27,
						EndLine:   2,
						EndColumn: 28,
					},
				},
			},

			// ---- Comments between keyword and label suppress autofix ----
			{
				Code:   `A: while(true) { /*comment*/break A; }`,
				Output: []string{`A: while(true) { /*comment*/break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 35},
				},
			},
			{
				Code: `A: while(true) { break/**/ A; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code: `A: while(true) { continue /**/ A; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 32},
				},
			},
			{
				Code: `A: while(true) { break /**/A; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code: `A: while(true) { continue/**/A; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 30},
				},
			},
			{
				Code:   `A: while(true) { continue A/*comment*/; }`,
				Output: []string{`A: while(true) { continue/*comment*/; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},
			{
				Code:   "A: while(true) { break A//comment\n }",
				Output: []string{"A: while(true) { break//comment\n }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code:   "A: while(true) { break A/*comment*/\nfoo() }",
				Output: []string{"A: while(true) { break/*comment*/\nfoo() }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},

			// =================================================================
			// Unnecessary `continue` on every loop variant
			// =================================================================
			{
				Code:   `A: while (a) { continue A; }`,
				Output: []string{`A: while (a) { continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code:   `A: do { continue A; } while (x);`,
				Output: []string{`A: do { continue; } while (x);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code:   `A: for (;;) { continue A; }`,
				Output: []string{`A: for (;;) { continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code:   `A: for (const k in obj) { continue A; }`,
				Output: []string{`A: for (const k in obj) { continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},
			{
				Code:   `A: for (const x of ary) { continue A; }`,
				Output: []string{`A: for (const x of ary) { continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 36},
				},
			},

			// =================================================================
			// Same-name shadowing: inner `A` shadows outer `A`, and `break A`
			// lands on the inner breakable (direct label). Unnecessary.
			// =================================================================
			{
				Code:   `A: while (a) { A: while (b) { break A; } }`,
				Output: []string{`A: while (a) { A: while (b) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 37},
				},
			},
			{
				Code:   `A: { A: while (b) { break A; } }`,
				Output: []string{`A: { A: while (b) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},

			// =================================================================
			// Chained labels — the directly-labeling (innermost) label is
			// always redundant; outer labels in the chain are legitimate.
			// =================================================================
			{
				Code:   `A: B: while (true) { break B; }`,
				Output: []string{`A: B: while (true) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 28},
				},
			},
			{
				Code:   `A: B: C: while (true) { continue C; }`,
				Output: []string{`A: B: C: while (true) { continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},

			// =================================================================
			// Sequential labels at the same top-level block — each scope pops
			// cleanly between them so diagnoses are independent
			// =================================================================
			{
				Code:   `A: while (a) { break A; } B: while (b) { break B; }`,
				Output: []string{`A: while (a) { break; } B: while (b) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
					{MessageId: "unexpected", Line: 1, Column: 48},
				},
			},

			// =================================================================
			// Inner unnecessary, outer necessary, in the same statement pair
			// =================================================================
			{
				Code:   `A: while (a) { B: while (b) { break B; break A; } }`,
				Output: []string{`A: while (a) { B: while (b) { break; break A; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 37},
				},
			},

			// =================================================================
			// Multiple unnecessary labels inside a single breakable scope —
			// both reported with correct columns, autofix applied to both
			// =================================================================
			{
				Code:   `A: while (a) { break A; continue A; }`,
				Output: []string{`A: while (a) { break; continue; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
					{MessageId: "unexpected", Line: 1, Column: 34},
				},
			},

			// =================================================================
			// Real-world matrix search where the inner label is redundant
			// =================================================================
			{
				Code:   `outer: for (let i = 0; i < 10; i++) { inner: for (let j = 0; j < 10; j++) { if (i === j) { break inner; } } }`,
				Output: []string{`outer: for (let i = 0; i < 10; i++) { inner: for (let j = 0; j < 10; j++) { if (i === j) { break; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "This label 'inner' is unnecessary.",
						Line:      1,
						Column:    98,
					},
				},
			},

			// =================================================================
			// Directly-labeled breakable reached through a switch case —
			// `break B` ends the switch that B directly labels, which naked
			// `break` also does; `break A` exits A (labeled switch), which
			// is also what bare `break` in a case body does.
			// =================================================================
			{
				Code:   `A: while (a) { B: switch (b) { case 0: break B; case 1: break A; } }`,
				Output: []string{`A: while (a) { B: switch (b) { case 0: break; case 1: break A; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 46},
				},
			},
			{
				Code:   `A: switch (a) { case 0: B: { if (x) break B; break A; } }`,
				Output: []string{`A: switch (a) { case 0: B: { if (x) break B; break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 52},
				},
			},

			// =================================================================
			// Cross-labeled loops: `continue B` targets the inner while that
			// B directly labels → naked `continue` equivalent → unnecessary
			// =================================================================
			{
				Code:   `A: while (a) { B: while (b) { if (x) break A; else continue B; } }`,
				Output: []string{`A: while (a) { B: while (b) { if (x) break A; else continue; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 61},
				},
			},
			{
				Code:   `A: for (;;) { B: for (;;) { C: for (;;) { break A; continue B; break C; } } }`,
				Output: []string{`A: for (;;) { B: for (;;) { C: for (;;) { break A; continue B; break; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 70},
				},
			},

			// =================================================================
			// Labels inside function / arrow / method / async / generator
			// =================================================================
			{
				Code:   `function f() { A: while (a) { break A; } }`,
				Output: []string{`function f() { A: while (a) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 37},
				},
			},
			{
				Code:   `const f = () => { A: while (a) { break A; } };`,
				Output: []string{`const f = () => { A: while (a) { break; } };`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 40},
				},
			},
			{
				Code:   `class C { m() { A: while (a) { break A; } } }`,
				Output: []string{`class C { m() { A: while (a) { break; } } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},
			{
				Code:   `async function f() { A: while (a) { break A; } }`,
				Output: []string{`async function f() { A: while (a) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 43},
				},
			},
			{
				Code:   `function* g() { A: while (a) { break A; } }`,
				Output: []string{`function* g() { A: while (a) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 38},
				},
			},

			// =================================================================
			// TypeScript-specific
			// =================================================================
			{
				Code:   `function f<T>(arr: T[]) { A: while (arr.length) { break A; } }`,
				Output: []string{`function f<T>(arr: T[]) { A: while (arr.length) { break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 57},
				},
			},
			{
				Code:   `A: for (const x of (ary as number[])) { break A; }`,
				Output: []string{`A: for (const x of (ary as number[])) { break; }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 47},
				},
			},

			// =================================================================
			// Multi-line with comment between keyword and label — still no
			// autofix, and line/column correctly track the label position.
			// =================================================================
			{
				Code: "A: while (a) {\n  break /* note */ A;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 20},
				},
			},

			// =================================================================
			// Nested labeled switch — only inner is redundant
			// =================================================================
			{
				Code:   `A: switch (a) { case 0: B: switch (b) { case 0: break B; } }`,
				Output: []string{`A: switch (a) { case 0: B: switch (b) { case 0: break; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55},
				},
			},

			// =================================================================
			// Labeled do-while (unnecessary) — covers do-while specifically
			// =================================================================
			{
				Code:   `A: do { if (x) continue A; } while (y);`,
				Output: []string{`A: do { if (x) continue; } while (y);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
		},
	)
}
