// TestEolLastExtras locks in branches and edge shapes that the upstream test
// suite doesn't exercise. Each case carries an inline comment pointing at the
// specific branch / Dimension 4 row / Go-port-specific concern it covers, so
// future refactors can't silently regress them without breaking a named
// lock-in.
//
// Group layout:
//   1. Dimension 4 walk (mostly N/A — this is a whole-file rule)
//   2. Empty / single-terminator boundary
//   3. Mode-shape coverage (explicit 'always', bare-string CLI shape)
//   4. Fix-regex `(?:\r?\n)+$` coverage (3-plus runs, mixed terminators)
//   5. Whitespace / tab preservation under 'never'
//   6. Encoding edges (multi-byte UTF-8, surrogate pair, non-`\n` line
//      terminators like `\f` and U+2028)
//   7. Real-user syntactic shapes (function body, single-line comment, JSX,
//      shebang, multi-line file)
//   8. Message text lock-ins
package eol_last_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/eol_last"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEolLastExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&eol_last.EolLastRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: N/A — receiver / expression wrappers ----
			// N/A: this rule inspects only the final byte(s) of the source
			// buffer, not any AST node. Paren / non-null / as-cast /
			// satisfies / optional-chain wrappers anywhere inside the file
			// are invisible to it; the decision depends solely on
			// `endsWithLF`.
			//
			// ---- Dimension 4: N/A — access / key forms ----
			// N/A: rule never inspects member-access or computed-key shapes.
			//
			// ---- Dimension 4: N/A — declaration / container forms ----
			// N/A: rule never inspects function / class / arrow declarations.
			//
			// ---- Dimension 4: N/A — nesting / traversal boundaries ----
			// N/A: rule fires once per file at Run setup; there is no AST
			// traversal and therefore no boundary to bleed across.
			//
			// ---- Dimension 4: graceful degradation — empty file ----
			// Empty source must not crash and must produce zero diagnostics
			// under either mode. Upstream covers `''` only under default
			// 'always'; lock in the 'never' side here.
			{Code: ``, Options: optStr("never")},

			// ---- Locks in upstream create() arm: explicit 'always' ----
			// Upstream's valid suite tests only the default (omitted) options
			// path. Explicit `["always"]` must behave identically — lock in
			// both LF and CRLF terminators so the `parseMode` string
			// round-trip stays equivalent to the default.
			{Code: "var a = 1;\n", Options: optStr("always")},
			{Code: "var a = 1;\r\n", Options: optStr("always")},

			// ---- Encoding: file ending in multi-byte unicode under 'always' ----
			// Source ending with a multi-byte rune followed by `\n` must stay
			// valid under 'always'. Locks in that `strings.HasSuffix` on the
			// raw byte buffer correctly detects the trailing `\n` regardless
			// of what precedes it.
			{Code: "var a = '中文';\n"},

			// ---- Encoding: surrogate-pair (emoji) ending under 'always' ----
			// Astral characters (4-byte UTF-8, 2 UTF-16 code units) before
			// the final `\n` must not throw off the byte-level newline check.
			{Code: "var a = '😀';\n"},
		},
		[]rule_tester.InvalidTestCase{
			// =========================================================
			// 2. Empty / single-terminator boundary
			// =========================================================

			// Locks in upstream create() arm 2 with src.length == 1 (single
			// LF). Upstream's `[length-1, length]` becomes `[0, 1]`; the
			// diagnostic must report at line 1 col 1 → line 2 col 1, and
			// the fix wipes the whole file.
			{
				Code:    "\n",
				Output:  []string{``},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},
			// Same as above for a single CRLF. Locks in the
			// `endsWithCRLF ? 2 : 1` ternary on a 2-byte source.
			{
				Code:    "\r\n",
				Output:  []string{``},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1, EndLine: 2, EndColumn: 1},
				},
			},

			// =========================================================
			// 3. Mode-shape coverage
			// =========================================================

			// Bare-string option — the shape rslint's config loader
			// produces from CLI / `rslint.config.mjs` when the rule entry
			// is `['error', 'never']`. Upstream tests always pass options
			// as `['never']` (array form). The Go test framework lets us
			// pass the bare string directly; lock in that `parseMode`
			// accepts both shapes.
			{
				Code:    "var a = 1;\n",
				Output:  []string{"var a = 1;"},
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11, EndLine: 2, EndColumn: 1},
				},
			},

			// =========================================================
			// 4. Fix-regex `(?:\r?\n)+$` coverage
			// =========================================================
			// Upstream only tests doubled `\n\n` and `\r\n\r\n`; these
			// lock in 3-plus runs and mixed `\r\n` / `\n` terminators.

			// Three trailing LFs — all stripped in one fix.
			{
				Code:    "var a = 1;\n\n\n",
				Output:  []string{`var a = 1;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 1, EndLine: 4, EndColumn: 1},
				},
			},
			// Three trailing CRLFs — all stripped.
			{
				Code:    "var a = 1;\r\n\r\n\r\n",
				Output:  []string{`var a = 1;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 1, EndLine: 4, EndColumn: 1},
				},
			},
			// Mixed: LF after CRLF. Final terminator is a bare `\n`
			// (endsWithCRLF=false), so diagnostic span is `[n-1, n]`.
			// Fix-regex matches the whole `\r\n\n` run and strips it.
			{
				Code:    "var a = 1;\r\n\n",
				Output:  []string{`var a = 1;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
			// Mixed: CRLF after LF. Final terminator is `\r\n`
			// (endsWithCRLF=true), so diagnostic span is `[n-2, n]`.
			// Fix-regex strips the whole `\n\r\n` run.
			{
				Code:    "var a = 1;\n\r\n",
				Output:  []string{`var a = 1;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},

			// =========================================================
			// 5. Whitespace / tab preservation under 'never'
			// =========================================================

			// Real-user: trailing whitespace then newline under 'never'.
			// Upstream covers `\n   ` (LF then spaces, no final newline)
			// under 'always'; this is the inverse. The fix must remove
			// only the trailing `\n` (the spaces are not part of the
			// `(?:\r?\n)+` run). Guards against accidental adoption of
			// `utils.SkipTrailingWhitespace` which would also chew the
			// spaces.
			{
				Code:    "var a = 1;\n   \n",
				Output:  []string{"var a = 1;\n   "},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 4, EndLine: 3, EndColumn: 1},
				},
			},
			// Tabs before final `\n` under 'never' — fix preserves tabs;
			// only the final `\n` is removed by `(?:\r?\n)+$`.
			{
				Code:    "var a = 1;\n\t\t\n",
				Output:  []string{"var a = 1;\n\t\t"},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 3, EndLine: 3, EndColumn: 1},
				},
			},
			// Whitespace-only file under 'never' — diagnostic at
			// end-of-spaces; fix wipes only the `\n`.
			{
				Code:    "   \n",
				Output:  []string{"   "},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4, EndLine: 2, EndColumn: 1},
				},
			},

			// =========================================================
			// 6. Encoding edges
			// =========================================================

			// Multi-byte UTF-8 (CJK) at end, no LF, 'always' → 'missing'.
			// `中` is 3 bytes UTF-8 but 1 UTF-16 code unit; the column
			// must be UTF-16-based (12 in 1-based form), not byte-based
			// (14). Locks in that rslint's position renderer uses UTF-16
			// width for the column on a non-ASCII line — matching ESLint.
			{
				Code:   `var a = "中"`,
				Output: []string{"var a = \"中\"\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},
			// Surrogate pair (emoji) at end, no LF. `😀` is 4 bytes UTF-8
			// AND 2 UTF-16 code units; the column reflects the
			// surrogate-pair width (13 in 1-based form).
			{
				Code:   `var a = "😀"`,
				Output: []string{"var a = \"😀\"\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},
			// Form feed (`\f`, U+000C) at end — NOT an ECMA line
			// terminator. `endsWith('\n')` is false → 'always' fires
			// 'missing'. The `\f` is counted as one column on line 1.
			// Locks in that we don't accidentally treat `\f` as a
			// terminator (e.g., via an over-eager `IsWhitespaceLike`
			// check).
			{
				Code:   "var a = 1;\f",
				Output: []string{"var a = 1;\f\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
				},
			},
			// U+2028 LINE SEPARATOR at end — ECMA line terminator but
			// NOT `\n`. `endsWith('\n')` is false → 'always' fires.
			// Position is on the line AFTER the LS (line 2, col 1)
			// because tsgo's `ComputeECMALineStarts` recognizes U+2028
			// as a line break — matching ESLint's `getLocFromIndex`
			// behavior.
			{
				Code:   "var a = 1; ",
				Output: []string{"var a = 1; \n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 1},
				},
			},

			// =========================================================
			// 7. Real-user syntactic shapes
			// =========================================================

			// Real-user: bare empty-body function with no trailing newline.
			{
				Code:   "function f() {}",
				Output: []string{"function f() {}\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 16},
				},
			},
			// Real-user: single-line comment as the entire file, no LF.
			// Locks in that comment-only files go through the same
			// `endsWithLF` path (no statement-list special-casing).
			{
				Code:   "// only a comment",
				Output: []string{"// only a comment\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18},
				},
			},
			// Real-user: JSX self-closing element as last token, no LF.
			// Locks in that the rule fires for `.tsx` files too — the
			// whole-file check is language-agnostic.
			{
				Code:   "const x = <div/>",
				Output: []string{"const x = <div/>\n"},
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 17},
				},
			},
			// Real-user: shebang start + valid statement, no trailing
			// LF. Position must be on the LAST line (line 2), confirming
			// the line-map walks the body, not just the file head.
			{
				Code:   "#!/usr/bin/env node\nvar a = 1;",
				Output: []string{"#!/usr/bin/env node\nvar a = 1;\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 11},
				},
			},
			// Real-user: multi-line function body under 'never'. Locks
			// in position computation on a non-trivial 4-line file
			// (column on line 3, end position on line 4). Guards
			// against off-by-one regressions in line-start scan when
			// there are several intermediate `\n`s.
			{
				Code:    "function f() {\n  return 1;\n}\n",
				Output:  []string{"function f() {\n  return 1;\n}"},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 2, EndLine: 4, EndColumn: 1},
				},
			},

			// =========================================================
			// 8. Message text lock-ins
			// =========================================================
			// Lock in the exact upstream message strings so a copy-paste
			// refactor of the Description fields can't silently drift.
			{
				Code:   "var a = 1;",
				Output: []string{"var a = 1;\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "missing",
						Message:   "Newline required at end of file but not found.",
						Line:      1,
						Column:    11,
					},
				},
			},
			{
				Code:    "var a = 1;\n",
				Output:  []string{`var a = 1;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Message:   "Newline not allowed at end of file.",
						Line:      1, Column: 11, EndLine: 2, EndColumn: 1,
					},
				},
			},
		},
	)
}
