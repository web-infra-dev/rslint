package no_warning_comments

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoWarningCommentsExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise. Each case carries an inline comment
// pointing at the specific branch / Dimension 4 row / tsgo AST quirk it
// covers, so future refactors can't silently regress them without breaking a
// named lock-in.
//
// Dimension 1 (AST node types), Dimension 2 (scoping & nesting), and
// Dimension 3 (autofix boundaries) do not apply to this rule: it inspects
// only raw comment text and comment kind (Line vs Block) via a single flat
// pass over ctx.Comments.All() — there is no expression/receiver, property
// key, declaration form, nested listener boundary, or autofix to exercise.
// Dimension 4 rows are walked explicitly below; every row not covered by an
// active case is marked N/A with a reason.
func TestNoWarningCommentsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoWarningCommentsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: graceful degradation (empty / near-empty comments) ----
			{Code: "//"},
			{Code: "/**/"},
			{Code: "/* */"},

			// N/A: Dimension 4 "receiver / expression wrappers", "access / key
			// forms", and "declaration / container forms" rows do not apply —
			// this rule never inspects an expression, property key, function, or
			// class; it only reads comment text.

			// N/A: Dimension 4 "nesting / traversal boundaries" does not apply —
			// checkComment runs once per entry of the flat ctx.Comments.All()
			// list; there is no nested listener to bleed across a boundary.

			// ---- location "start" requires the prefix to be pure whitespace/decoration: a leading punctuation character that is not configured decoration blocks the match ----
			{Code: "// !TODO", Options: map[string]any{"terms": []interface{}{"todo"}, "location": "start"}},

			// ---- terms: [] explicitly overrides the default term list to empty — nothing can ever match ----
			{Code: "// TODO FIXME XXX", Options: map[string]any{"terms": []interface{}{}}},

			// ---- Real-user: eslint/eslint#16103 — a JSDoc-style block comment's leading `*` is not implicitly treated as decoration; without configuring `decoration`, "start" location does not skip past it ----
			{Code: "/**\n * todo\n */"},

			// ---- shebang is never surfaced as a comment at all (tsgo's scanner consumes it before comment collection begins), so unlike upstream's explicit `token.type !== "Shebang"` filter, no filtering is needed on the rslint side; a shebang line containing warning-term-shaped text is still never inspected ----
			{Code: "#!/usr/bin/env node --todo-flag\n// ok"},

			// ---- Real-user: eslint/eslint#15775 — astUtils.isDirectiveComment's
			// block-comment pattern also recognizes "globals " (plural) and
			// "exported " prefixes, not just "global "/"eslint "; each combines
			// with the self-config-mention guard to skip a comment that documents
			// this rule's own configuration, even though the configured term
			// ("globals"/"exported") otherwise appears in the comment ----
			{Code: "/* globals no-warning-comments */", Options: map[string]any{"terms": []interface{}{"globals"}, "location": "anywhere"}},
			{Code: "/* exported no-warning-comments */", Options: map[string]any{"terms": []interface{}{"exported"}, "location": "anywhere"}},

			// ---- Locks in isDirectiveComment()'s Line-comment branch (trimmed
			// value starts with "eslint-") combined with the self-config guard —
			// upstream's own suite only exercises the Block-comment branch of this
			// `&&` condition (via the `/*eslint no-warning-comments: [...] */`
			// case); verified against real ESLint (eslint-disable-next-line
			// reports zero no-warning-comments diagnostics for this input) ----
			{Code: "// eslint-disable-next-line no-warning-comments", Options: map[string]any{"terms": []interface{}{"eslint"}, "location": "anywhere"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- decoration is documented as ignored when location is
			// "anywhere": convertToRegExp only threads escapedDecoration into the
			// pattern when location === "start", so this reports identically
			// whether decoration is configured or not (paired with the next case)
			// — the leading "**" isn't consumed by decoration here, it simply
			// doesn't block the \b-bounded "anywhere" match ----
			{
				Code:    "//**TODO",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "//**TODO",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere", "decoration": []interface{}{"*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Locks in upstream convertToRegExp() prefix branch: decoration
			// and the term's own leading characters share the same character
			// class, so the regex engine can consume decoration greedily and
			// still backtrack to find the term — a naive "skip decoration, then
			// match term" two-step scan would miss this ----
			{
				Code:    "//tttodo",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "start", "decoration": []interface{}{"t"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Locks in escapeRegExp()'s `-` handling: a decoration character
			// list containing "-" must not be misread as forming a character-class
			// range once embedded in `[<whitespace><decoration>]*` ----
			{
				Code:    "//-*todo",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "start", "decoration": []interface{}{"-", "*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Real-user: eslint/eslint#11471 — a term containing a hyphen
			// used to produce an invalid regular expression ("Invalid escape") at
			// rule-init time because upstream's escaping omitted `-`; verifies the
			// port's escapeRegExp escapes it and the resulting regexp both compiles
			// and matches ----
			{
				Code:    "// flow-typed version: 1.0",
				Options: map[string]any{"terms": []interface{}{"flow-typed version"}, "location": "start"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'flow-typed version' comment: 'flow-typed version: 1.0'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- Dimension 4: leading NBSP (U+00A0) counts as whitespace for the
			// "start" prefix, matching ECMAScript WhiteSpace/JS `\s` — RE2's
			// built-in `\s` is ASCII-only, so this exercises the rule's explicit
			// jsWhitespaceClass rather than RE2's default ----
			{
				Code:    "// TODO",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "start"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			// ---- Real-user: eslint/eslint#15775 — astUtils.isDirectiveComment's
			// block-comment pattern also recognizes "globals " (plural) and
			// "exported " prefixes, not just "global "/"eslint "; each must combine
			// with the self-config-mention guard to skip a comment that documents
			// this rule's own configuration — moved to the valid-side table below
			// since a correct skip means zero diagnostics.

			// ---- Locks in checkComment()'s `&&` short-circuit: a comment that
			// mentions "no-warning-comments" but is NOT itself a directive comment
			// (no eslint-/global/globals/exported prefix) must still be reported ----
			{
				Code:    "// see no-warning-comments docs for details: todo",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 50},
				},
			},
			// ---- Real-user: eslint/eslint#16103 — with `decoration: ["*"]`
			// configured, the JSDoc-style leading `*` on each line is skipped by
			// the "start" prefix, so a term after it now matches (contrast with
			// the valid-side case above, which uses default options) ----
			{
				Code:    "/**\n * todo\n */",
				Options: map[string]any{"decoration": []interface{}{"*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 3, EndColumn: 4},
				},
			},
			// ---- Locks in the CHAR_LIMIT truncation boundary exactly: a
			// trimmed comment of precisely 40 UTF-16 code units is NOT truncated ----
			{
				Code:    "//" + strings.Repeat("a", 35) + " TODO",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedComment",
						Message:   "Unexpected 'todo' comment: '" + strings.Repeat("a", 35) + " TODO'.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 43,
					},
				},
			},
			// ---- Locks in the CHAR_LIMIT truncation boundary exactly: 41 UTF-16
			// code units crosses the limit and truncates with a trailing "..." ----
			{
				Code:    "//" + strings.Repeat("a", 36) + " TODO",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedComment",
						Message:   "Unexpected 'todo' comment: '" + strings.Repeat("a", 36) + "...'.",
						Line:      1, Column: 1, EndLine: 1, EndColumn: 44,
					},
				},
			},
			// ---- Dimension 4: an astral-plane rune (outside the BMP, e.g. an
			// emoji) counts as 2 UTF-16 code units toward CHAR_LIMIT, matching JS
			// String#length — not 1, as a naive Go rune count would give ----
			{
				Code:    "// " + strings.Repeat("\U0001F600", 10) + " TODO more text here to overflow",
				Options: map[string]any{"terms": []interface{}{"todo"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedComment",
						Message:   "Unexpected 'todo' comment: '" + strings.Repeat("\U0001F600", 10) + " TODO more text here...'.",
						Line:      1, Column: 1,
					},
				},
			},
			// ---- Dimension 4: term boundary uses ASCII-only \w on both sides
			// (matching JS `\w`, unaffected by the `u` flag), so a term ending in
			// a non-ASCII letter is not word-boundary-checked against a directly
			// following ASCII letter — permissive matching, consistent with
			// upstream ----
			{
				Code:    "// caféx",
				Options: map[string]any{"terms": []interface{}{"café"}, "location": "anywhere"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			// ---- Contract: options: [{}] (an explicitly supplied empty options
			// object) must fall back to the same defaults as no options at all ----
			{
				Code:    "// fixme",
				Options: map[string]any{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedComment", Message: "Unexpected 'fixme' comment: 'fixme'.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
		},
	)
}
