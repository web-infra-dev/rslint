package no_unexpected_multiline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnexpectedMultilineRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnexpectedMultilineRule,
		// Valid cases — ported from ESLint's `tests/lib/rules/no-unexpected-multiline.js`
		[]rule_tester.ValidTestCase{
			{Code: `(x || y).aFunction()`},
			{Code: `[a, b, c].forEach(doSomething)`},
			{Code: `var a = b;
(x || y).doSomething()`},
			{Code: `var a = b
;(x || y).doSomething()`},
			{Code: `var a = b
void (x || y).doSomething()`},
			{Code: `var a = b;
[1, 2, 3].forEach(console.log)`},
			{Code: `var a = b
void [1, 2, 3].forEach(console.log)`},
			{Code: `"abc\
(123)"`},
			{Code: `var a = (
(123)
)`},
			{Code: `f(
(x)
)`},
			{Code: `(
function () {}
)[1]`},
			{Code: "let x = function() {};\n   `hello`"},
			{Code: "let x = function() {}\nx `hello`"},
			{Code: "String.raw `Hi\n${2+3}!`;"},
			{Code: "x\n.y\nz `Valid Test Case`"},
			{Code: "f(x\n)`Valid Test Case`"},
			{Code: "x.\ny `Valid Test Case`"},
			{Code: "(x\n)`Valid Test Case`"},
			{Code: `
                foo
                / bar /2
            `},
			{Code: `
                foo
                / bar / mgy
            `},
			{Code: `
                foo
                / bar /
                gym
            `},
			{Code: `
                foo
                / bar
                / ygm
            `},
			{Code: `
                foo
                / bar /GYM
            `},
			{Code: `
                foo
                / bar / baz
            `},
			{Code: `foo /bar/g`},
			{Code: `
                foo
                /denominator/
                2
            `},
			{Code: `
                foo
                / /abc/
            `},
			{Code: `
                5 / (5
                / 5)
            `},

			// https://github.com/eslint/eslint/issues/11650 — TypeScript generic
			// type-argument forms.
			{Code: `
                tag<generic>` + "`" + `
                    multiline
                ` + "`" + `;
            `},
			{Code: `
                tag<
                  generic
                >` + "`" + `
                    multiline
                ` + "`" + `;
            `},
			{Code: `
                tag<
                  generic
                >` + "`multiline`" + `;
            `},

			// Optional chaining — ESLint skips when the link itself carries
			// `?.`, so all four forms are valid regardless of newlines.
			{Code: "var a = b\n  ?.(x || y).doSomething()"},
			{Code: "var a = b\n  ?.[a, b, c].forEach(doSomething)"},
			{Code: "var a = b?.\n  (x || y).doSomething()"},
			{Code: "var a = b?.\n  [a, b, c].forEach(doSomething)"},

			// Class fields where ASI separates the members rather than chaining
			// them into a single MemberExpression.
			{Code: "class C { field1\n[field2]; }"},
			{Code: "class C { field1\n*gen() {} }"},
			// Arrow function initializer doesn't connect to the next member.
			{Code: "class C { field1 = () => {}\n[field2]; }"},
			{Code: "class C { field1 = () => {}\n*gen() {} }"},

			// tsgo-specific edges (locking in behavior beyond the ESLint suite):
			// Element access on the same line as the receiver — common case.
			{Code: `b[c]`},
			// Tagged template attached on the same line — common case.
			{Code: "tag`hello`"},
			// Optional chain continuation on a new line, but the call/access
			// stays single-line — `?.` is on the inner link, not this one.
			{Code: "a?.b\n.c"},
			// Sequence-style division `(a, b) / c / d`: the inner `/` is on
			// a sequence's right operand, not a `/` itself, so the listener
			// shouldn't activate the regex-flag check.
			{Code: `(a, b) / c / d`},
			// Compound assignment `/=` — must NOT match the slash listener.
			{Code: `let x = 1; x /= 2`},
			// Locks in upstream "no-newline" branch: regex-flag match requires
			// the identifier to be IMMEDIATELY after the slash (no trivia).
			// Here there's a space, so the check must skip even though the
			// identifier text matches `[gimsuy]+`.
			{Code: "foo / bar / g"},
			// Locks in identifier-only branch: numeric token after slash isn't
			// an Identifier, so the check skips even when adjacent.
			{Code: "foo\n/ bar /2"},
			// Same-line division+regex-flags shouldn't trigger.
			{Code: `foo / bar /gym`},
			// Multi-level paren wrap above the inner division — exercises the
			// paren-skipping ancestor walk.
			{Code: "((foo / bar)) / 2"},
			// `(a/b)/c` form: ESTree drops the inner parens so the selector
			// matches, but `c` is numeric so no flag-identifier — same line
			// also keeps it valid even if it did match.
			{Code: `(foo / bar) / 5`},
			// Optional-chain continuation call (outer call has no `?.`) but
			// stays single-line.
			{Code: "a?.b()"},
			// TS non-null receiver, single line.
			{Code: `b![c]`},
			// TS `as` cast wrapper around the receiver.
			{Code: `(b as any)[c]`},
			// Empty-arg call on its own line — ESLint skips because the
			// `()` form is unambiguous.
			{Code: "b\n()"},
			// Optional-chain element access (root) — skipped regardless of
			// the newline before `[`.
			{Code: "a\n?.[b]"},
			// Chained member access where parens disambiguate but no
			// newline is present.
			{Code: `(x || y).aFunction()`},

			// === Real-world IIFE / module patterns ===
			// IIFE with ASI guard `;(function...)` — the rule must NOT
			// flag this idiom because the `;` separates statements.
			{Code: "var a = b;\n;(function() {})()"},
			// UMD-style IIFE on its own — single statement, no newline issue.
			{Code: "(function () {\n  return 1;\n})();"},
			// Method chain across lines, each call is single-line — common
			// builder pattern (`.then().catch()`).
			{Code: "fetch(url)\n.then(r => r.json())\n.catch(handle)"},
			// jQuery-style chain spread across lines.
			{Code: "$(selector)\n  .find('.x')\n  .each(fn)"},

			// === Real-world generic / type-argument patterns ===
			// Generic call with type arguments — same line, common TS case.
			{Code: "fn<string>(x)"},
			// Generic call wrapped in parens, then property access on same line.
			{Code: "(fn<string>(x)).y"},
			// Multi-line type argument list followed by call args on the
			// same line as `>` — common in styled-components style code.
			{Code: "fn<\n  Props,\n  State\n>(x)"},
			// Generic-typed function reference with the call args on a new
			// line — ESLint compares the callee end (`fn`) to the next token
			// (`<`), which is on the SAME line, so no report.
			{Code: "fn<string>\n(x)"},
			// Type assertion wrapping a function call as receiver, single line.
			{Code: "(getX() as Foo).y"},

			// === Numeric / regex disambiguation in the wild ===
			// Pure regex literal at start of statement — not a division.
			{Code: "/foo/g.test(x)"},
			// Multiplication that visually resembles a regex pattern.
			{Code: "a * b / c * d"},
			// Modulo not affected by the rule.
			{Code: "a\n% b"},
			// `**` exponentiation, not `/`.
			{Code: "a\n** b"},
			// Bracket notation as a numeric index, single line.
			{Code: "arr[0]"},
			// Nested computed access, single line.
			{Code: "obj['a']['b']"},

			// === Template literal real-world patterns ===
			// Multi-line template literal content — newlines INSIDE the
			// template, not before the backtick — must NOT report.
			{Code: "tag`first\nsecond\nthird`"},
			// Tagged template where Tag is a generic-typed call (TS).
			{Code: "fn<T>`x`"},
			// Plain template literal at statement position, no tag — outside
			// the rule's scope (no TaggedTemplateExpression).
			{Code: "`hello`"},
			{Code: "let x = 1;\n`hello`"},

			// === Optional chaining real cases ===
			// Optional element access continuation across lines — only the
			// root has `?.`, but it's still on the same chain so ESLint's
			// `optional: false` would fire... but the property access here
			// stays single-line, so no break.
			{Code: "a?.b[c]"},
			// Optional call followed by template on next line — handled in
			// invalid section (the `?.` only skips the CallExpression
			// listener, not the TaggedTemplate listener).

			// === Receiver wrappers that must not bleed past boundaries ===
			// Class member with computed name on the SAME line as the field
			// keyword — must not be flagged.
			{Code: "class C { [a]; }"},
			// Class with static block — internal newlines don't reach the
			// rule's listeners.
			{Code: "class C {\n  static { console.log(1); }\n}"},

			// === Async / generator function bodies ===
			// Async function with newline between body and tagged template
			// — but the template is INSIDE the function body, not after.
			{Code: "async function f() {\n  return tag`x`;\n}"},
			// Generator function call with arg on same line.
			{Code: "function* g() { yield 1; }\ng().next()"},

			// === Cases that look like ASI hazards but aren't ===
			// `return` on its own line ends the statement (ASI inserts
			// semicolon BEFORE the next-line content).
			{Code: "function f() {\n  return\n  (1).toString()\n}"},
			// `throw` similarly — ASI inserts semicolon.
			{Code: "function f() {\n  throw new Error()\n}"},

			// === No leading expression — rule shouldn't fire ===
			// Standalone array literal as expression statement.
			{Code: "[1, 2, 3]"},
			// Standalone parenthesized expression statement.
			{Code: "(1 + 2)"},
			// Standalone tagged template at statement start.
			{Code: "tag`x`"},

			// === Multi-byte / Unicode content ===
			// Unicode identifier as receiver, then computed access on same
			// line — must scan correctly.
			{Code: `日本[c]`},
			// Multi-byte content inside string before the rule's listeners
			// fire on a separate expression.
			{Code: "var a = '日本';\n[1].forEach(f)"},
			// Emoji inside template literal content.
			{Code: "tag`hi 🎉`"},
			// Unicode identifier immediately after the second `/` — must NOT
			// match `^[gimsuy]+$` (ASCII regex), so no division report. This
			// stress-tests the UTF-8 decoding in scanIdentifier.
			{Code: "foo\n/bar/日本"},
			// Unicode identifier as numerator with division across lines —
			// the rule's checks are byte-position based but should treat the
			// multi-byte name as a single identifier source-text-wise.
			{Code: "日本 /bar/g"},
		},
		// Invalid cases — ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: "var a = b\n(x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			{
				Code: "var a = (a || b)\n(x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			{
				Code: "var a = (a || b)\n(x).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			{
				Code: "var a = b\n[a, b, c].forEach(doSomething)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			{
				Code: "var a = b\n    (x || y).doSomething()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 5, EndLine: 2, EndColumn: 6,
				}},
			},
			{
				Code: "var a = b\n  [a, b, c].forEach(doSomething)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 3, EndLine: 2, EndColumn: 4,
				}},
			},
			{
				Code: "let x = function() {}\n `hello`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      2, Column: 2, EndLine: 2, EndColumn: 3,
				}},
			},
			{
				Code: "let x = function() {}\nx\n`hello`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      3, Column: 1, EndLine: 3, EndColumn: 2,
				}},
			},
			{
				Code: "x\n.y\nz\n`Invalid Test Case`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      4, Column: 1, EndLine: 4, EndColumn: 2,
				}},
			},
			{
				Code: `
                foo
                / bar /gym
            `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      3, Column: 17, EndLine: 3, EndColumn: 18,
				}},
			},
			{
				Code: `
                foo
                / bar /g
            `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      3, Column: 17, EndLine: 3, EndColumn: 18,
				}},
			},
			{
				Code: `
                foo
                / bar /g.test(baz)
            `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      3, Column: 17, EndLine: 3, EndColumn: 18,
				}},
			},
			{
				Code: `
                foo
                /bar/gimuygimuygimuy.test(baz)
            `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      3, Column: 17, EndLine: 3, EndColumn: 18,
				}},
			},
			{
				Code: `
                foo
                /bar/s.test(baz)
            `,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      3, Column: 17, EndLine: 3, EndColumn: 18,
				}},
			},

			// https://github.com/eslint/eslint/issues/11650 — TS generics with
			// a block comment between `>` and the template.
			{
				Code: "const x = aaaa<\n  test\n>/*\ntest\n*/`foo`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      5, Column: 3, EndLine: 5, EndColumn: 4,
				}},
			},

			// Class fields where the parser keeps the initializer chained
			// across the newline.
			{
				Code: "class C { field1 = obj\n[field2]; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			{
				Code: "class C { field1 = function() {}\n[field2]; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},

			// Extra coverage beyond upstream — locks in tsgo-specific shape
			// behavior the ESLint suite doesn't exercise.

			// Paren-wrapped numerator: ESTree strips parens so the inner `/`
			// matches the selector; tsgo preserves them, so the
			// WalkUpParenthesizedExpressions skip is required for parity.
			// `(foo)\n/ bar /gym` — first `/` on the next line.
			{
				Code: "(foo)\n/ bar /gym",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Multi-line MemberExpression where the receiver itself spans
			// lines — break is between the `)` of the receiver and `[`.
			{
				Code: "(a\n||b)\n[c]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      3, Column: 1, EndLine: 3, EndColumn: 2,
				}},
			},
			// TS non-null assertion in front of the computed access — break
			// happens between `!` and `[`.
			{
				Code: "b!\n[c]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// TS `as` cast as receiver of a call.
			{
				Code: "(b as any)\n(c)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Tagged template where Tag is a CallExpression — both
			// "function" and "taggedTemplate" listeners can fire, but here
			// the call args are on the same line, so only the template
			// break should trigger.
			{
				Code: "tag(x)\n`hello`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Template with substitutions still triggers when separated.
			{
				Code: "tag\n`hi${x}!`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Optional-chain CONTINUATION call (outer call lacks `?.`)
			// across a line break — should still flag because ESLint's
			// `node.optional` is false on continuations.
			{
				Code: "a?.b\n(c)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Block comment between numerator and `/` should still flag —
			// trivia doesn't change the line of the slash.
			{
				Code: "foo /* x */\n/ bar /g",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},

			// === Real-world ASI traps ===
			// Module-pattern IIFE typo: missing `;` between two var decls.
			{
				Code: "var a = foo\n(function () {})()",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// `const X = arr` followed by destructure-looking line — common
			// developer mistake.
			{
				Code: "const x = arr\n[0]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Return-value mistakenly continues onto next line via call.
			{
				Code: "function f() {\n  return getThing\n  (x)\n}",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      3, Column: 3, EndLine: 3, EndColumn: 4,
				}},
			},

			// === Deeply nested receivers ===
			// PropertyAccess chain as receiver of a computed access across
			// a line break.
			{
				Code: "obj.a.b.c\n[d]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Call result as receiver of a tagged template across lines.
			{
				Code: "fn().g()\n`x`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Element access as receiver of a call across lines.
			{
				Code: "obj[k]\n(x)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},

			// === TS-only constructs as the receiver ===
			// `satisfies` clause on the receiver of a call.
			{
				Code: "(x satisfies Foo)\n(y)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "function",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// `as const` cast as receiver of an element access.
			{
				Code: "(arr as const)\n[0]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// === Division flag exhaustive coverage ===
			// All six valid regex flags (`gimsuy`) spelled out immediately
			// after the second slash — must match `^[gimsuy]+$` and report
			// at the FIRST slash on line 2.
			{
				Code: "foo\n/bar/gimsuy",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Single-flag minimal repro: first `/` on line 2 col 1.
			{
				Code: "x\n/y/i",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "division",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},

			// === Optional call as Tag of a tagged template ===
			// `fn?.(x)\n\`hello\`` — the call's `?.` opts only the
			// CallExpression listener out; the TaggedTemplate listener still
			// fires on the break between `)` and the backtick.
			{
				Code: "fn?.(x)\n`hello`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Tagged template after a call-with-args, all separated by
			// newlines — only the LAST break (between `)` and backtick)
			// triggers; the `(x)` args are immediately after `tag`.
			{
				Code: "tag(x)\n\n`hi`",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "taggedTemplate",
					Line:      3, Column: 1, EndLine: 3, EndColumn: 2,
				}},
			},
			// Multi-line break between receiver and `[` — many blank lines
			// in between, all should be one diagnostic at the `[`.
			{
				Code: "var a = b\n\n\n[c]",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      4, Column: 1, EndLine: 4, EndColumn: 2,
				}},
			},

			// === Class field continuations beyond upstream ===
			// Class field initializer that's a chained call across lines.
			{
				Code: "class C { field = obj.method()\n[k]; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
			// Class field initializer is a tagged template — the field's
			// initializer crosses into a `[k]` access.
			{
				Code: "class C { field = tag`x`\n[k]; }",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "property",
					Line:      2, Column: 1, EndLine: 2, EndColumn: 2,
				}},
			},
		},
	)
}
