// TestNoInlineCommentsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 4 walk (rows that don't apply to no-inline-comments, with
// reasons): the rule scans every comment in the file once via
// ctx.Comments.All() and decides purely from surrounding source text
// (preceding/following characters on the comment's own start/end line); it
// never inspects a comment's enclosing expression's AST shape except through
// the single JSX-exception check (ast.GetNodeAtPosition + KindJsxExpression),
// which upstream's own JSX test suite already exercises exhaustively
// (migrated in full in no_inline_comments_upstream_test.go).
//   - N/A receiver/expression wrappers on inspected inputs (parens, `!`,
//     `as`/`satisfies`, optional chain): the rule never inspects an
//     expression's shape or wrapper depth, only raw source text around a
//     comment's position.
//   - N/A access/key forms (identifier/string/numeric/private/computed key,
//     element access): the rule has no notion of a "key" at all.
//   - N/A declaration/container forms (class/function declaration vs
//     expression, async/generator variants): the rule fires identically
//     regardless of the enclosing declaration kind.
//   - N/A autofix boundaries: the rule has neither an autofix nor a
//     suggestion (ESLint's upstream rule is report-only).
//   - N/A three-way equivalence classes: the rule does not compare names or
//     keys.
package no_inline_comments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInlineCommentsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInlineCommentsRule,
		[]rule_tester.ValidTestCase{
			// ---- Locks in upstream isDirectiveComment() Block arm: ESLINT_DIRECTIVE_PATTERN's "eslint " (space, no dash) alternative ----
			// Upstream's own tests only exercise the "eslint-" (dash) form via
			// eslint-disable-line; the plain "eslint " config-comment form
			// (e.g. `/* eslint no-console: 0 */`) is a distinct regex
			// alternative that upstream never tests.
			{Code: "var a = 1; /* eslint no-console: 0 */"},

			// ---- This linter recognizes "rslint-" directives alongside
			// "eslint-" ones, so the directive-comment exemption has to treat
			// both prefixes alike in both comment kinds ----
			{Code: "var a = 1; // rslint-disable-line no-console"},
			{Code: "var a = 1; /* rslint-disable-line no-console */"},

			// ---- Locks in upstream's `.trim()` on the text around the comment: JavaScript trims U+FEFF, Go's strings.TrimSpace does not ----
			// A zero-width no-break space is the only code point where the
			// two trims disagree, and it is legal JavaScript whitespace, so
			// a comment preceded by one alone on its line has an empty
			// preamble and stays unreported.
			{Code: "var a = 1;\n\ufeff// comment\n"},

			// ---- Dimension 4: graceful degradation — comment as the last bytes of the file (no trailing newline) ----
			{Code: "// a solitary trailing comment, no newline after it"},

			// ---- Dimension 4: graceful degradation — shebang line is not a scanned comment ----
			// The shebang itself must never surface as a reportable comment;
			// only the real inline comment on the following line should.
			{Code: "#!/usr/bin/env node\nvar a = 1;\n// fine, on its own line"},

			// ---- Real-user: JSX fragment shorthand (`<>...</>`) ----
			// Fragments are extremely common in modern React code and were
			// not part of upstream's <div>-based JSX suite; the JSX
			// exception must still apply to a fragment's expression
			// container.
			{Code: "var a = (\n  <>\n    {/* comment */}\n    <h1>Some heading</h1>\n  </>\n)", Tsx: true},

			// ---- Locks in upstream `customIgnoreRegExp.test(node.value)`: matched against the raw, untrimmed comment value ----
			// The pattern anchors on the single leading space that only the
			// untrimmed comment value has; matching a trimmed value would
			// never see that space and this comment would stay reported.
			{
				Code:    "var a = 3; // note",
				Options: []any{map[string]any{"ignorePattern": "^ note$"}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Locks in upstream isDirectiveComment() Line arm: Line comments only ever match the narrower "eslint-" (dash) startsWith check ----
			// The Block-only "eslint " (space) alternative of
			// ESLINT_DIRECTIVE_PATTERN must NOT exempt a Line comment: this
			// is an asymmetry between the Line and Block arms that upstream
			// itself never tests (its only "eslint " coverage is
			// "eslint-disable-line", which already starts with "eslint-").
			{
				Code: "var a = 1; // eslint no-console: 0",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 12, EndLine: 1, EndColumn: 35},
				},
			},

			// ---- Locks in upstream isDirectiveComment() Line arm: "global "/"globals "/"exported " only exempt Block comments, never Line comments ----
			{
				Code: "foo(); // global foo",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 8, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "var foo; // exported foo",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 10, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Locks in upstream JSX exception: preamble === "{" / postamble === "}" requires the enclosing node to be an empty JsxExpression, not any brace-delimited node ----
			// A plain JS block statement reproduces the exact same
			// preamble ("" — empty, nothing before on the comment's own
			// line) / postamble ("}") shape the JSX exception looks for,
			// but its enclosing node is a Block, not a JsxExpression, so it
			// must still be reported.
			{
				Code: "if (a)\n{\n/* comment */}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 3, Column: 1, EndLine: 3, EndColumn: 14},
				},
			},

			// ---- Real-user: TypeScript type annotation with a trailing inline comment ----
			{
				Code: "const x: number /* px */ = 5;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 17, EndLine: 1, EndColumn: 25},
				},
			},

			// ---- Real-user: optional chaining with a trailing line comment ----
			{
				Code: "foo?.bar(); // note",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 13, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- Dimension 4: graceful degradation — comment inside a rest element in a binding pattern ----
			// No node-kind special-casing anywhere in the rule: a comment
			// sitting between a shorthand property and a RestElement is
			// scanned and reported exactly like any other inline comment.
			{
				Code: "var { a, /* c */ ...rest } = obj;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Line: 1, Column: 10, EndLine: 1, EndColumn: 17},
				},
			},
		},
	)
}
