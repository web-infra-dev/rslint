// TestCurlyExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
// Cases are grouped by `// ==== <area> ====` separators: Dimension-4 edge
// shapes, deep nesting & `consistent` chains, real-user / expression / TS-only
// bodies, statement-kind bodies & needsSemicolon resolution, and autofix
// boundaries. Everything past the upstream floor was differentially validated
// against ESLint (9.39.4 / v10.5.0) as oracle.
//
// Dimension 4 walk (rows that don't apply to curly, with reasons):
//   - N/A receiver/expression wrappers ((X).y, X!.y, (X as T).y, X?.y): curly
//     classifies the *statement* kind of a control-flow body, never an inner
//     expression/receiver, so paren/non-null/as/optional-chain wrappers can't
//     change which statement is matched.
//   - N/A access/key forms (dotted vs computed vs private): curly never inspects
//     member access or property keys.
//   - N/A three-way name/key equivalence classes: curly compares nothing by name.
package curly

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCurlyExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&CurlyRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration/container forms (function variants) ----
			// async / generator / async-generator function declarations are
			// FunctionDeclaration → braces required even under "multi".
			{Code: "if (a) { async function f() {} }", Options: "multi"},
			{Code: "if (a) { function* g() {} }", Options: "multi"},
			{Code: "if (a) { async function* h() {} }", Options: "multi"},
			// ---- Dimension 4: declaration/container forms (class variants) ----
			{Code: "if (a) { class C {} }", Options: "multi"},
			{Code: "if (a) { abstract class C {} }", Options: "multi"},

			// ---- Dimension 4: TS-only declarations (keep braces; unbraced is a syntax error) ----
			// Locks in the isLexicalDeclaration TS extension over ESLint's set.
			{Code: "if (a) { enum E { X } }", Options: "multi"},
			{Code: "if (a) { const enum E { X } }", Options: "multi"},
			{Code: "if (a) { namespace N { f(); } }", Options: "multi"},
			{Code: "if (a) { module M { f(); } }", Options: "multi"},
			{Code: "if (a) { interface I {} }", Options: "multi"},
			{Code: "if (a) { type T = number; }", Options: "multi"},
			// Same, under multi-or-nest (single-line block still keeps braces).
			{Code: "if (a) { enum E { X } }", Options: "multi-or-nest"},
			{Code: "if (a) { type T = number; }", Options: "multi-or-nest"},

			// ---- Dimension 4: graceful degradation (empty bodies) ----
			// Empty block has 0 statements (!= 1) → braces kept by every mode.
			{Code: "if (a) {}", Options: "multi"},
			{Code: "while (a) {}", Options: "multi"},
			{Code: "for (;;) {}", Options: "multi"},
			{Code: "if (a) {}"},
			// EmptyStatement body under multi / multi-or-nest is a one-liner → no braces wanted.
			{Code: "if (a);", Options: "multi"},
			{Code: "while (a);", Options: "multi-or-nest"},

			// ---- Real-user/tsgo: type annotation on a let keeps it lexical ----
			{Code: "if (a) { let x: number = 1; }", Options: "multi"},
			{Code: "if (a) { const x: Foo = bar; }", Options: "multi-or-nest"},

			// ---- Real-user/tsgo: optional chain / as-cast bodies are plain statements ----
			// (one-liners, so "multi-or-nest" wants them brace-less, and they are)
			{Code: "if (a) foo?.bar();", Options: "multi-or-nest"},
			{Code: "if (a) x = y as Foo;", Options: "multi"},

			// ---- Dimension 4: TS import-equals body keeps braces (unbraced is a syntax error) ----
			{Code: "if (a) { import X = N.M; }", Options: "multi"},
			// Locks in upstream hasUnsafeIf() WithStatement arm: a single-level
			// `with` wrapping an else-less `if`, followed by `else`, keeps braces.
			{Code: "if (a) { with (o) if (b) foo(); } else bar();", Options: "multi"},
			// ---- Dimension 4: parenthesized one-liner body, no braces wanted ----
			{Code: "if (a) (b());", Options: "multi-or-nest"},

			// ---- Graceful degradation: rslint does not schema-validate options
			// (no native rule does), so invalid combos degrade instead of erroring.
			// `["consistent"]` alone has no multi* mode → behaves like "all". ----
			{Code: "if (a) { b() }", Options: []interface{}{"consistent"}},
			// Unknown 2nd option is ignored → behaves like plain "multi".
			{Code: "if (a) b()", Options: []interface{}{"multi", "foo"}},

			// ==== deep nesting & consistent chains ====
			// else-if chains, all one-liners, under "multi" → no braces anywhere
			{Code: "if (a) foo(); else if (b) bar(); else baz();", Options: "multi"},
			// 4-branch chain: first is multi-statement (keeps), rest one-liners
			{Code: "if (a) { x(); y(); } else if (b) z(); else if (c) w(); else v();", Options: "multi"},
			// consistent + multi, all three branches braceless → consistent already
			{Code: "if (a) foo(); else if (b) bar(); else baz();", Options: []interface{}{"multi", "consistent"}},
			// consistent + multi-or-nest, nested removable everywhere
			{Code: "if (a) { if (b) c(); } else { d(); }", Options: []interface{}{"multi-or-nest", "consistent"}},
			// consistent + multi-line, multiline only in the condition → no braces forced
			{Code: "if (a &&\n b) foo()", Options: []interface{}{"multi-line", "consistent"}},

			// ==== real-user shapes & expression / TS-only bodies ====
			// ---- Real-user issue #12972: leading-comment forcing applies ONLY to block
			// bodies, not bare bodies → a comment before a bare one-liner reports nothing ----
			{Code: "if (foo)\n    // some comment\n    bar();", Options: "multi-or-nest"},
			{Code: "if (foo)\n    // some comment\n    bar();", Options: "multi"},
			// same comment INSIDE a single-stmt block under multi-or-nest → keep braces
			{Code: "if (foo) {\n    // some comment\n    bar();\n}", Options: "multi-or-nest"},
			// ---- Real-user issue #13280: real nested for-of/if/for-of/if-else, braces necessary ----
			{Code: "function find() {\n  for (const [lineId, line] of lines.entries())\n    if (regexp === false) {\n      for (const reg of regionRegexps)\n        if (testLine(line, reg, regionName)) {\n          start = lineId + 1;\n          regexp = reg;\n          break;\n        }\n    } else if (testLine(line, regexp, regionName, true))\n      return { start, end: lineId, regexp };\n}", Options: "multi"},
			// return one-liner body under multi
			{Code: "function f() { if (a) return; }", Options: "multi"},

			// ---- TS-only valid (declarations keep braces; one-liner TS exprs need none) ----
			{Code: "async function f() { for await (const x of y) z(); }", Options: "multi"},
			{Code: "if (a) { @dec class C {} }", Options: "multi"},
			{Code: "if (a) { @dec class C {} }", Options: "multi-or-nest"},
			{Code: "if (a) { declare const x: number; }", Options: "multi"},
			{Code: "if (a) { abstract class C { abstract m(): void; } }", Options: "multi"},
			{Code: "if (a)\n foo<number>();", Options: "multi-or-nest"},
			{Code: "if (a)\n x = y satisfies Foo;", Options: "multi-or-nest"},

			// ==== statement-kind bodies & needsSemicolon node resolution ====
			// do-while with a multi-statement body keeps its braces under "multi"
			{Code: "do { f(); g(); } while (x)", Options: "multi"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Real-user/tsgo: as-cast body, braces unnecessary under multi ----
			{
				Code:    "if (foo) { x = y as Foo; }",
				Output:  []string{"if (foo)  x = y as Foo; "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- Real-user/tsgo: optional-chain call body, braces unnecessary under multi ----
			{
				Code:    "if (foo) { a?.b(); }",
				Output:  []string{"if (foo)  a?.b(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- Dimension 4: EmptyStatement body under "all" wants braces ----
			{
				Code:   "if (a);",
				Output: []string{"if (a){;}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Message: "Expected { after 'if' condition.", Line: 1, Column: 7, EndLine: 1, EndColumn: 8},
				},
			},

			// ---- Dimension 4: same-kind nesting (both ifs flagged, fixed across passes) ----
			{
				Code:   "if (a) if (b) foo()",
				Output: []string{"if (a) {if (b) foo()}", "if (a) {if (b) {foo()}}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 20},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 15, EndLine: 1, EndColumn: 20},
				},
			},

			// Locks in needsSemicolon arm: next token starts with backtick → no fix.
			{
				Code:    "if (foo) { bar }\n`x`.length;",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// Locks in needsSemicolon arm: next token starts with `-` → no fix.
			{
				Code:    "if (foo) { bar }\n-baz;",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
			// Locks in needsSemicolon arm: last token is `--` → no fix.
			{
				Code:    "if (foo) { bar-- }\nbaz;",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},

			// ---- Message-text contract: every messageId × name the rule emits ----
			{
				Code:   "do foo(); while (bar)",
				Output: []string{"do {foo();} while (bar)"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Message: "Expected { after 'do'.", Line: 1, Column: 4, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:   "for (x in y) z();",
				Output: []string{"for (x in y) {z();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Message: "Expected { after 'for-in'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   "for (x of y) z();",
				Output: []string{"for (x of y) {z();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Message: "Expected { after 'for-of'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:   "for (;;) z();",
				Output: []string{"for (;;) {z();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Message: "Expected { after 'for' condition.", Line: 1, Column: 10, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:   "while (a) b();",
				Output: []string{"while (a) {b();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Message: "Expected { after 'while' condition.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code:   "if (a) {b()} else c();",
				Output: []string{"if (a) {b()} else {c();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Message: "Expected { after 'else'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "for (x in y) { z(); }",
				Output:  []string{"for (x in y)  z(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Message: "Unnecessary { after 'for-in'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 22},
				},
			},
			// Locks in needsSemicolon arm: last statement already ends with `;`
			// (tokenBefore is `;`) → fix proceeds even though `}` and `c` share a line.
			{
				Code:    "if (a) { b(); } c();",
				Output:  []string{"if (a)  b();  c();"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code:    "while (a) { b(); }",
				Output:  []string{"while (a)  b(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Message: "Unnecessary { after 'while' condition.", Line: 1, Column: 11, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "do { foo(); } while (bar)",
				Output:  []string{"do  foo();  while (bar)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Message: "Unnecessary { after 'do'.", Line: 1, Column: 4, EndLine: 1, EndColumn: 14},
				},
			},

			// ---- needsPrecedingSpace: do{...} glued, first inner token decides space ----
			// Identifier-part (numeric) → space inserted (matches canTokensBeAdjacent).
			{
				Code:    "do{5;}while(x)",
				Output:  []string{"do 5;while(x)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 7},
				},
			},
			// Punctuator `(` → no space.
			{
				Code:    "do{(1);}while(x)",
				Output:  []string{"do(1);while(x)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
				},
			},
			// Template backtick → no space.
			{
				Code:    "do{`x`;}while(x)",
				Output:  []string{"do`x`;while(x)"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- Dimension 4: labeled statement as the sole removable body ----
			{
				Code:    "if (a) { lbl: foo(); }",
				Output:  []string{"if (a)  lbl: foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- Dimension 4: parenthesized expression body ----
			{
				Code:    "if (a) { (b()); }",
				Output:  []string{"if (a)  (b()); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- Real-user/tsgo: `satisfies` and non-null `!` bodies are plain statements ----
			{
				Code:    "if (foo) { x = y satisfies Foo; }",
				Output:  []string{"if (foo)  x = y satisfies Foo; "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    "if (foo) { x = y!; }",
				Output:  []string{"if (foo)  x = y!; "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 10, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- Message-text contract: remaining messageId × name combos ----
			{
				Code:    "if (foo) { bar() }",
				Output:  []string{"if (foo)  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Message: "Unnecessary { after 'if' condition.", Line: 1, Column: 10, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "for (;;) { foo(); }",
				Output:  []string{"for (;;)  foo(); "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Message: "Unnecessary { after 'for' condition.", Line: 1, Column: 10, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    "if (foo) baz(); else { bar() }",
				Output:  []string{"if (foo) baz(); else  bar() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Message: "Unnecessary { after 'else'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "for (x of y) { z() }",
				Output:  []string{"for (x of y)  z() "},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfter", Message: "Unnecessary { after 'for-of'.", Line: 1, Column: 14, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- needsSemicolon × comments: the token scan skips comments, matching
			// ESLint's getTokenBefore/getTokenAfter ----
			// Comment after `}` on the same line as `bar` → still same-line ASI → no fix.
			{
				Code:    "if (a) { foo() } /* c */ bar()",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
				},
			},
			// Comment skipped to reach `[` → ASI hazard → no fix.
			{
				Code:    "if (a) { bar }\n/* c */[1, 2, 3].map(x)",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 15},
				},
			},
			// Multi-line comment pushes `bar` to the next line; `bar` is safe → fix proceeds.
			{
				Code:    "if (a) { foo() } /*\nc*/ bar()",
				Output:  []string{"if (a)  foo()  /*\nc*/ bar()"},
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
				},
			},
			// Comment between last statement and `}` — tokenBefore skips it to `)`.
			{
				Code:    "if (a) { foo() /* c */ } bar()",
				Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Dimension 4: multi-byte chars — report columns are UTF-16 units,
			// matching ESLint. An astral char (😀 = 2 UTF-16 units) widens endColumn. ----
			{
				Code:   `if (a) foo("😀")`,
				Output: []string{`if (a) {foo("😀")}`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 17},
				},
			},
			// BMP multi-byte (日本 = 1 UTF-16 unit each) in the condition does not shift the body column.
			{
				Code:   "if (日本) foo()",
				Output: []string{"if (日本) {foo()}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 9, EndLine: 1, EndColumn: 14},
				},
			},
			// ---- Graceful degradation (invalid option schema): no crash, predictable mode ----
			{
				Code:    "if (a) b()",
				Output:  []string{"if (a) {b()}"},
				Options: []interface{}{"consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
				},
			},

			// ==== deep nesting & consistent chains ====
			// ---- 3-level if/while/for under "all": braces added one level per pass ----
			{
				Code: "if (a) while (b) for (;;) c();",
				Output: []string{
					"if (a) {while (b) for (;;) c();}",
					"if (a) {while (b) {for (;;) c();}}",
					"if (a) {while (b) {for (;;) {c();}}}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 31},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 18, EndLine: 1, EndColumn: 31},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
				},
			},
			// ---- 4-level for nesting under "all" (4 passes) ----
			{
				Code: "for(;;) for(;;) for(;;) for(;;) x();",
				Output: []string{
					"for(;;) {for(;;) for(;;) for(;;) x();}",
					"for(;;) {for(;;) {for(;;) for(;;) x();}}",
					"for(;;) {for(;;) {for(;;) {for(;;) x();}}}",
					"for(;;) {for(;;) {for(;;) {for(;;) {x();}}}}",
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 9, EndLine: 1, EndColumn: 37},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 17, EndLine: 1, EndColumn: 37},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 25, EndLine: 1, EndColumn: 37},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
				},
			},
			// ---- 3-level nesting under multi-or-nest ----
			{
				Code:    "if (a)\n if (b)\n for (;;)\n c();",
				Output:  []string{"if (a)\n {if (b)\n for (;;)\n c();}", "if (a)\n {if (b)\n {for (;;)\n c();}}"},
				Options: "multi-or-nest",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 2, Column: 2, EndLine: 4, EndColumn: 6},
					{MessageId: "missingCurlyAfterCondition", Line: 3, Column: 2, EndLine: 4, EndColumn: 6},
				},
			},
			// ---- else-if chain, 3 branches, "all" (single pass) ----
			{
				Code:   "if (a) foo(); else if (b) bar(); else baz();",
				Output: []string{"if (a) {foo();} else if (b) {bar();} else {baz();}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 27, EndLine: 1, EndColumn: 33},
					{MessageId: "missingCurlyAfter", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
				},
			},
			// ---- consistent + multi: only 1st braced → all forced braceless ----
			{
				Code:    "if (a) { foo(); } else if (b) bar(); else baz();",
				Output:  []string{"if (a)  foo();  else if (b) bar(); else baz();"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- consistent + multi: all three one-liner branches braced → remove all ----
			{
				Code:    "if (a) { foo(); } else if (b) { bar(); } else { baz(); }",
				Output:  []string{"if (a)  foo();  else if (b)  bar();  else  baz(); "},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 18},
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 31, EndLine: 1, EndColumn: 41},
					{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 47, EndLine: 1, EndColumn: 57},
				},
			},
			// ---- consistent + multi-or-nest: only 2nd braced single-liner → remove ----
			{
				Code:    "if (a) foo(); else if (b) { bar(); } else baz();",
				Output:  []string{"if (a) foo(); else if (b)  bar();  else baz();"},
				Options: []interface{}{"multi-or-nest", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 27, EndLine: 1, EndColumn: 37},
				},
			},
			// ---- consistent + multi-line: one multiline branch forces braces on ALL ----
			{
				Code:    "if (a) foo(); else if (b) bar(); else\nbaz()",
				Output:  []string{"if (a) {foo();} else if (b) {bar();} else\n{baz()}"},
				Options: []interface{}{"multi-line", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 27, EndLine: 1, EndColumn: 33},
					{MessageId: "missingCurlyAfter", Line: 2, Column: 1, EndLine: 2, EndColumn: 6},
				},
			},
			// ---- consistent forcing ADD: else is lexical/multi-stmt → if branch forced to brace ----
			{
				Code:    "if (a) foo(); else { let x = 1; }",
				Output:  []string{"if (a) {foo();} else { let x = 1; }"},
				Options: []interface{}{"multi", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code:    "if (a) { if (b) foo(); } else bar();",
				Output:  []string{"if (a) { if (b) foo(); } else {bar();}"},
				Options: []interface{}{"multi-or-nest", "consistent"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfter", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
				},
			},

			// ==== real-user shapes & expression / TS-only bodies ====
			// ---- Real-user issue #13216: multi-line parenthesized return body wants braces (all) ----
			{
				Code:   "function C() { if (open) return (\n  foo\n); }",
				Output: []string{"function C() { if (open) {return (\n  foo\n);} }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 26, EndLine: 3, EndColumn: 3},
				},
			},
			// ---- Expression-statement body variety under "all" ----
			{Code: "if (a) b ? c() : d();", Output: []string{"if (a) {b ? c() : d();}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 22}}},
			{Code: "if (a) throw new Error('x');", Output: []string{"if (a) {throw new Error('x');}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},
			{Code: "if (a) b = 1, c = 2;", Output: []string{"if (a) {b = 1, c = 2;}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			{Code: "async function f() { if (a) await g(); }", Output: []string{"async function f() { if (a) {await g();} }"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 29, EndLine: 1, EndColumn: 39}}},
			{Code: "function* f() { if (a) yield x; }", Output: []string{"function* f() { if (a) {yield x;} }"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 24, EndLine: 1, EndColumn: 32}}},
			// ---- arrow-returning-object / IIFE block bodies removable under multi ----
			{Code: "if (a) { foo(() => ({ x: 1 })); }", Output: []string{"if (a)  foo(() => ({ x: 1 })); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 34}}},
			{Code: "if (a) { (function(){})(); }", Output: []string{"if (a)  (function(){})(); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},

			// ---- TS-only: for-await-of behaves like for-of ----
			{Code: "async function f() { for await (const x of y) z(); }",
				Output: []string{"async function f() { for await (const x of y) {z();} }"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfter", Line: 1, Column: 47, EndLine: 1, EndColumn: 51}}},
			{Code: "async function f() { for await (const x of y) { z(); } }",
				Output: []string{"async function f() { for await (const x of y)  z();  }"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 47, EndLine: 1, EndColumn: 55}}},
			// ---- TS-only: using / await using in for-of head ----
			{Code: "for (using x of y) { z(); }", Output: []string{"for (using x of y)  z(); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 20, EndLine: 1, EndColumn: 28}}},
			{Code: "async function f() { for (await using x of y) { z(); } }",
				Output: []string{"async function f() { for (await using x of y)  z();  }"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 47, EndLine: 1, EndColumn: 55}}},
			// ---- TS-only: generic call expression body behaves like a plain call ----
			{Code: "if (a) { foo<T>(); }", Output: []string{"if (a)  foo<T>(); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},

			// ==== statement-kind bodies & needsSemicolon node resolution ====
			// ---- switch-statement body ----
			{Code: "if (a) switch (b) { case 1: f(); }", Output: []string{"if (a) {switch (b) { case 1: f(); }}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 35}}},
			{Code: "if (a) { switch (b) { case 1: f(); } }", Output: []string{"if (a)  switch (b) { case 1: f(); } "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 39}}},
			// ---- try/catch and try/finally bodies ----
			{Code: "if (a) try { f(); } catch (e) {}", Output: []string{"if (a) {try { f(); } catch (e) {}}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 33}}},
			{Code: "if (a) try { f(); } finally { g(); }", Output: []string{"if (a) {try { f(); } finally { g(); }}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37}}},
			{Code: "if (a) { try { f(); } catch (e) {} }", Output: []string{"if (a)  try { f(); } catch (e) {} "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 37}}},
			// ---- bare block body — 2-pass removal (outer then inner) ----
			{Code: "if (a) { { x(); } }", Output: []string{"if (a)  { x(); } ", "if (a)   x();  "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// ---- do-while single-stmt body inside an if-block, removable ----
			{Code: "if (a) { do x(); while (y) }", Output: []string{"if (a)  do x(); while (y) "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 29}}},

			// ---- needsSemicolon lastBlockNode resolution (which tail makes removal unfixable) ----
			// switch tail → resolves to SwitchStatement (not Block) → same-line ASI → NO fix
			{Code: "if (a) { switch (b) {} } bar()", Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// switch tail then `;` → still NO fix
			{Code: "if (a) { switch (b) {} } ; bar()", Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 25}}},
			// try/catch tail → catch Block (parent CatchClause) → FIX proceeds
			{Code: "if (a) { try {} catch (e) {} } bar()", Output: []string{"if (a)  try {} catch (e) {}  bar()"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 31}}},
			// bare-block tail → BlockStatement → FIX proceeds
			{Code: "if (a) { { foo() } } bar()", Output: []string{"if (a)  { foo() }  bar()"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 21}}},
			// do-while tail → last token `)` not a block → same-line ASI → NO fix
			{Code: "if (a) { do {} while (x) } bar()", Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 27}}},

			// ---- for-in block whose single stmt is `var` (non-lexical) → removable ----
			{Code: "for (k in o) { var x = 1; }", Output: []string{"for (k in o)  var x = 1; "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfter", Line: 1, Column: 14, EndLine: 1, EndColumn: 28}}},
			// ---- empty-statement body of a for under "all" → wrap ----
			{Code: "for (;;);", Output: []string{"for (;;){;}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 9, EndLine: 1, EndColumn: 10}}},
			// ---- labeled `if` (label wraps the if; body still flagged) ----
			{Code: "lbl: if (a) b();", Output: []string{"lbl: if (a) {b();}"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 13, EndLine: 1, EndColumn: 17}}},

			// ==== autofix boundaries: comments / whitespace / multi-pass ====
			// ---- comment INSIDE the block preserved on brace removal ----
			{Code: "if (a) { /* c */ foo(); }", Output: []string{"if (a)  /* c */ foo(); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			{Code: "if (a) { foo(); /* c */ }", Output: []string{"if (a)  foo(); /* c */ "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 26}}},
			{Code: "if (a) { foo() /* trailing */ }", Output: []string{"if (a)  foo() /* trailing */ "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 32}}},
			{Code: "if (a) { // keep\n foo();\n}", Output: []string{"if (a)  // keep\n foo();\n"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 3, EndColumn: 2}}},
			// ---- comment BETWEEN condition and `{` preserved; report range starts at `{` ----
			{Code: "if (a) /*c*/ { foo(); }", Output: []string{"if (a) /*c*/  foo(); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 14, EndLine: 1, EndColumn: 24}}},
			// ---- multi-level parenthesized body ----
			{Code: "if (a) { ((b())); }", Output: []string{"if (a)  ((b())); "}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 20}}},
			// ---- arbitrary whitespace preserved verbatim on removal (blank lines / tab / CRLF) ----
			{Code: "if (a) {\n\n  foo();\n\n}", Output: []string{"if (a) \n\n  foo();\n\n"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 5, EndColumn: 2}}},
			{Code: "if (a) {\n\tfoo();\n}", Output: []string{"if (a) \n\tfoo();\n"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 3, EndColumn: 2}}},
			{Code: "if (a) {\r\n foo();\r\n}", Output: []string{"if (a) \r\n foo();\r\n"}, Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 3, EndColumn: 2}}},
			// ---- multi-pass add-braces (fix result still violates → 2nd pass) ----
			{Code: "while (a) if (b) c();", Output: []string{"while (a) {if (b) c();}", "while (a) {if (b) {c();}}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 11, EndLine: 1, EndColumn: 22},
					{MessageId: "missingCurlyAfterCondition", Line: 1, Column: 18, EndLine: 1, EndColumn: 22}}},
			// ---- `}` then comment then `else` → diagnostic but NO fix (else follows the block) ----
			{Code: "if (a) { foo() } /* c */ else b()", Options: "multi",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedCurlyAfterCondition", Line: 1, Column: 8, EndLine: 1, EndColumn: 17}}},
		},
	)
}
