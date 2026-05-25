// cspell:ignore FFFD callees lookarounds
package no_control_regex

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestCollectControlChars(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		flags   string
		want    []string
	}{
		// ── Negative cases ──
		{"no control chars", `abc`, ``, nil},
		{"empty pattern", ``, ``, nil},
		{"escaped backslash then x hex", `\\x1f`, ``, nil}, // \\ + literal x1f
		{"raw char just above range", "\x20", ``, nil},
		{"raw DEL (above range)", "\x7f", ``, nil},

		// ── \xHH ──
		{"xHH at zero", `\x00`, ``, []string{`\x00`}},
		{"xHH at max", `\x1f`, ``, []string{`\x1f`}},
		{"xHH uppercase hex", `\x1F`, ``, []string{`\x1f`}},
		{"xHH mixed case", `\x1A`, ``, []string{`\x1a`}},
		{"xHH just above range", `\x20`, ``, nil},
		{"xHH ascii letter", `\x61`, ``, nil},
		{"xHH incomplete (1 digit)", `\x1`, ``, nil},
		{"xHH empty after x", `\x`, ``, nil},
		{"xHH non-hex first", `\xG0`, ``, nil},
		{"xHH non-hex second", `\x0G`, ``, nil},
		{"xHH multiple", `\x1f\x1e`, ``, []string{`\x1f`, `\x1e`}},
		{"xHH adjacent plus text", `a\x1fb\x00c`, ``, []string{`\x1f`, `\x00`}},
		{"xHH with quantifier", `\x1f+`, ``, []string{`\x1f`}},
		{"xHH inside alternation", `a|\x1f|b`, ``, []string{`\x1f`}},

		// ── \uHHHH ──
		{"uHHHH control", `\u001F`, ``, []string{`\x1f`}},
		{"uHHHH control zero", `\u0000`, ``, []string{`\x00`}},
		{"uHHHH non-control", `\u0020`, ``, nil},
		{"uHHHH above BMP edge", `\uFFFD`, ``, nil},
		{"uHHHH incomplete (3 digits)", `\u001`, ``, nil},
		{"uHHHH non-hex", `\u00GG`, ``, nil},
		{"uHHHH surrogate (not control)", `\uD83D`, ``, nil},

		// ── \u{H...} under u/v flag ──
		{"u-brace u: single", `\u{1f}`, `u`, []string{`\x1f`}},
		{"u-brace u: uppercase", `\u{1F}`, `u`, []string{`\x1f`}},
		{"u-brace u: leading zeros", `\u{0000001F}`, `u`, []string{`\x1f`}},
		{"u-brace u: zero", `\u{0}`, `u`, []string{`\x00`}},
		{"u-brace u: above range", `\u{20}`, `u`, nil},
		{"u-brace u: far above (BMP)", `\u{FFFF}`, `u`, nil},
		{"u-brace u: astral plane", `\u{10FFFF}`, `u`, nil},
		{"u-brace v: single", `\u{1f}`, `v`, []string{`\x1f`}},
		{"u-brace u: empty braces", `\u{}`, `u`, nil},
		{"u-brace u: non-hex", `\u{GG}`, `u`, nil},
		{"u-brace u: unclosed", `\u{1f`, `u`, nil},
		{"u-brace no flag: treated literally", `\u{1F}`, ``, nil},
		{"u-brace g flag only: treated literally", `\u{1F}`, `g`, nil},
		{"u-brace mixed flags with u", `\u{1F}`, `gui`, []string{`\x1f`}},
		{"u-brace u-flag multiple", `\u{1F}\u{1E}`, `u`, []string{`\x1f`, `\x1e`}},

		// ── Raw control chars in pattern ──
		{"raw zero", "\x00", ``, []string{`\x00`}},
		{"raw tab (0x09)", "\x09", ``, []string{`\x09`}},
		{"raw at range max", "\x1f", ``, []string{`\x1f`}},
		{"raw mixed", "a\x1fb\x1ec", ``, []string{`\x1f`, `\x1e`}},
		{"raw consecutive", "\x01\x02\x03", ``, []string{`\x01`, `\x02`, `\x03`}},

		// ── Symbolic escapes — all allowed ──
		{"symbolic \\t", `\t`, ``, nil},
		{"symbolic \\n", `\n`, ``, nil},
		{"symbolic \\r", `\r`, ``, nil},
		{"symbolic \\v", `\v`, ``, nil},
		{"symbolic \\f", `\f`, ``, nil},
		{"symbolic \\0", `\0`, ``, nil},
		{"symbolic \\b (word boundary)", `\b`, ``, nil},
		{"symbolic \\cI", `\cI`, ``, nil},
		{"symbolic \\cJ", `\cJ`, ``, nil},
		{"symbolic \\d", `\d`, ``, nil},
		{"symbolic \\w", `\w`, ``, nil},
		{"symbolic \\s", `\s`, ``, nil},
		{"mixed symbolic and control", `\t\x1f\n`, ``, []string{`\x1f`}},

		// ── Character classes ──
		{"class with xHH", `[\x1f]`, ``, []string{`\x1f`}},
		{"class negated with xHH", `[^\x1f]`, ``, []string{`\x1f`}},
		{"class range with controls", `[\x00-\x1f]`, ``, []string{`\x00`, `\x1f`}},
		{"class escaped bracket literal", `\[\x1f\]`, ``, []string{`\x1f`}},
		{"v-flag nested class control", `[[\u{1F}]]`, `v`, []string{`\x1f`}},
		{"v-flag set difference", `[\u{1F}--B]`, `v`, []string{`\x1f`}},

		// ── Groups, lookarounds ──
		{"named capture with control", `(?<a>\x1f)`, ``, []string{`\x1f`}},
		{"lookbehind with control", `(?<=\x1f)a`, ``, []string{`\x1f`}},
		{"lookahead with control", `a(?=\x1f)`, ``, []string{`\x1f`}},
		{"non-capturing group", `(?:\x1f)`, ``, []string{`\x1f`}},

		// ── Trailing backslash (malformed but shouldn't crash) ──
		{"trailing backslash", `abc\`, ``, nil},

		// ── Combined forms in same pattern ──
		{"mixed xHH + uHHHH + u-brace", `\x01\u0002\u{3}`, `u`, []string{`\x01`, `\x02`, `\x03`}},

		// ── Surrogate pairs (neither half is a control code point) ──
		{"surrogate pair as \\uHHHH\\uHHHH", `\uD83D\uDC7F`, ``, nil},
		{"surrogate pair under u flag", `\uD83D\uDC7F`, `u`, nil},

		// ── Legacy octal escapes — ESLint never reports these ──
		{"octal \\01", `\01`, ``, nil},
		{"octal \\07", `\07`, ``, nil},
		{"octal \\012", `\012`, ``, nil},
		{"octal preceded by content", `a\01b`, ``, nil},

		// ── Unicode property escape \p{...} — not a control escape ──
		{"\\p{Letter} under u", `\p{Letter}`, `u`, nil},
		{"\\P{Letter} under u", `\P{Letter}`, `u`, nil},
		{"\\p{Script=Latin} under u", `\p{Script=Latin}`, `u`, nil},
		{"\\p next to control xHH", `\p{Letter}\x1f`, `u`, []string{`\x1f`}},

		// ── Documented divergence from ESLint on syntactically-invalid patterns.
		// ESLint's collector inherits regexpp's stop-on-error behavior; this
		// scanner intentionally does not model that (see rule doc). The cases
		// below pin the current over-reporting behavior so changes are loud.
		{"malformed: \\u{...} invalid content before control", `\u{NOT_HEX}\x1f`, `u`, []string{`\x1f`}},
		{"malformed: unclosed [ still collects control inside", `[\x1f`, ``, []string{`\x1f`}},
		{"malformed: unclosed ( still collects control inside", `(\x1f`, ``, []string{`\x1f`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectControlChars(tt.pattern, tt.flags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectControlChars(%q, %q) = %v, want %v", tt.pattern, tt.flags, got, tt.want)
			}
		})
	}
}

// TestCollectControlChars_AllBoundary exercises every code point in 0x00..0x20
// across raw, \xHH, \uHHHH forms — the primary boundary for this rule.
func TestCollectControlChars_AllBoundary(t *testing.T) {
	for cp := range 0x21 {
		want := []string{fmt.Sprintf(`\x%02x`, cp)}
		if cp > 0x1f {
			want = nil
		}

		// \xHH
		if got := collectControlChars(fmt.Sprintf(`\x%02x`, cp), ``); !reflect.DeepEqual(got, want) {
			t.Errorf("\\x%02x: got %v want %v", cp, got, want)
		}
		// \uHHHH
		if got := collectControlChars(fmt.Sprintf(`\u%04x`, cp), ``); !reflect.DeepEqual(got, want) {
			t.Errorf("\\u%04x: got %v want %v", cp, got, want)
		}
		// \u{H} under u flag
		if got := collectControlChars(fmt.Sprintf(`\u{%x}`, cp), `u`); !reflect.DeepEqual(got, want) {
			t.Errorf("\\u{%x} (u flag): got %v want %v", cp, got, want)
		}
	}
	// Raw control characters in pattern (only 0x00..0x1f — 0x20 is a normal space).
	for cp := range 0x20 {
		want := []string{fmt.Sprintf(`\x%02x`, cp)}
		if got := collectControlChars(string(rune(cp)), ``); !reflect.DeepEqual(got, want) {
			t.Errorf("raw \\x%02x: got %v want %v", cp, got, want)
		}
	}
}

func TestNoControlRegexRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoControlRegexRule,
		[]rule_tester.ValidTestCase{
			// ── Baseline: no control chars ──
			{Code: `var regex = /x1f/`},
			{Code: `var regex = /\\x1f/`},
			{Code: `var regex = new RegExp('x1f')`},
			{Code: `var regex = RegExp('x1f')`},
			{Code: `new RegExp('[')`},
			{Code: `RegExp('[')`},
			{Code: `/\u{20}/u`},
			{Code: `/\u{1F}/`},
			{Code: `/\u{1F}/g`},
			{Code: `new RegExp("\\u{20}", "u")`},
			{Code: `new RegExp("\\u{1F}")`},
			{Code: `new RegExp("\\u{1F}", "g")`},
			{Code: `new RegExp("\\u{1F}", flags)`},
			{Code: `new RegExp("[\\q{\\u{20}}]", "v")`},
			{Code: `/[\u{20}--B]/v`},
			{Code: "new RegExp('\\x20')"},

			// ── Symbolic escapes — all allowed ──
			{Code: `/\t/`},
			{Code: `/\n/`},
			{Code: `/\r/`},
			{Code: `/\v/`},
			{Code: `/\f/`},
			{Code: `/\0/`},
			{Code: `/\b/`},  // word boundary
			{Code: `/\cI/`}, // control-I
			{Code: `/\cJ/`},

			// ── Non-RegExp callees — should not match ──
			{Code: `foo.RegExp('\x1f')`},
			{Code: `window.RegExp('\x1f')`},
			{Code: `this.RegExp('\x1f')`},
			{Code: `regexp('\x1f')`}, // lowercase, not RegExp
			{Code: `bar('\x1f')`},
			{Code: `new (function foo(){})('\x1f')`}, // callee isn't Identifier "RegExp"

			// ── Non-string first argument — constructor path skipped ──
			{Code: "RegExp(pattern)"},
			{Code: "RegExp(/x20/)"},      // inner regex has no controls
			{Code: "RegExp('a' + 'b')"},  // binary expression
			{Code: "RegExp(cond ? 'a' : 'b')"},
			{Code: "RegExp(123)"},
			{Code: "RegExp(null)"},
			{Code: "RegExp(undefined)"},
			{Code: "RegExp(`template`)"}, // template literal

			// ── No-args / zero-arg constructor ──
			{Code: "new RegExp"},
			{Code: "RegExp()"},
			{Code: "new RegExp()"},

			// ── Spread first argument — not a StringLiteral, skip ──
			{Code: "new RegExp(...args)"},
			{Code: "RegExp(...['\\x1f'])"},

			// ── Surrogate pair (non-control) ──
			{Code: `/\uD83D\uDC7F/`},
			{Code: `new RegExp("\\uD83D\\uDC7F")`},

			// ── Legacy octal in regex literal — ESLint does not report these ──
			// (Note: octal escapes are disallowed in TS string literals, so we
			// only exercise the regex-literal path here.)
			{Code: `/\01/`},
			{Code: `/\012/`},

			// ── \p{...} unicode property — not a control escape ──
			{Code: `/\p{Letter}/u`},
			{Code: `new RegExp("\\p{Letter}", "u")`},
		},
		[]rule_tester.InvalidTestCase{
			// ── Regex literals: \xHH ──
			{
				Code: `var regex = /\x1f/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /\x00/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /\x0C/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /\\\x1f\\x1e/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /\\\x1fFOO\\x00/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /FOO\\\x1fFOO\\x1f/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			// Regex literal: \uHHHH
			{
				Code: `/\u000C/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Regex literal: \u{H} under u flag
			{
				Code: `/\u{C}/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `/\u{1F}/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `/\u{1F}/gui`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Regex literal: u-flag combined with other escapes
			{
				Code: `/\u{1111}*\x1F/u`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Regex literal: v-flag set notation
			{
				Code: `/[\u{1F}--B]/v`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Regex literal: named capture + control
			{
				Code: `var regex = /(?<a>\x1f)/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `var regex = /(?<\u{1d49c}>.)\x1f/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},

			// ── Constructor: raw character strings ──
			{
				Code: "var regex = new RegExp('\\x1f\\x1e')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code: "var regex = new RegExp('\\x1fFOO\\x00')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code: "var regex = new RegExp('FOO\\x1fFOO\\x1f')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 24},
				},
			},
			{
				Code: "var regex = RegExp('\\x1f')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 20},
				},
			},
			// Constructor: \\uHHHH (escaped — interpreted by RegExp parser)
			{
				Code: `new RegExp("\\u001F", flags)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `new RegExp("\\u{1111}*\\x1F", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `new RegExp("\\u{1F}", "u")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `new RegExp("\\u{1F}", "gui")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `new RegExp("[\\q{\\u{1F}}]", "v")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},

			// ── Constructor: parenthesized callee ──
			{
				Code: `(RegExp)('\x1f')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code: `((RegExp))('\x1f')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12},
				},
			},
			{
				Code: `(new RegExp('\x1f'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},

			// ── Multiple expressions: report only the one with control chars ──
			{
				Code: `/\x11/; RegExp("foo", "uv");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// ── Nesting contexts ──
			// Inside function call arg
			{
				Code: `foo(/\x1f/)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},
			// Inside conditional
			{
				Code: `cond ? /\x1f/ : null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			// Inside array literal
			{
				Code: `[/\x1f/]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 2},
				},
			},
			// Inside object literal
			{
				Code: `({ re: /\x1f/ })`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8},
				},
			},
			// Inside arrow function body
			{
				Code: `const fn = () => /\x1f/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			// Nested RegExp: inner reports
			{
				Code: `RegExp(RegExp('\x1f'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// In template literal expression
			{
				Code: "`${/\\x1f/}`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4},
				},
			},
			// IIFE
			{
				Code: `(function() { /\x1f/; })();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15},
				},
			},
			// Class method body
			{
				Code: `class C { m() { /\x1f/; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17},
				},
			},
			// Default parameter
			{
				Code: `function f(x = /\x1f/) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 16},
				},
			},
			// For-loop init
			{
				Code: `for (let r = /\x1f/;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			// Return statement
			{
				Code: `function g() { return /\x1f/; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			// Logical expression
			{
				Code: `true && /\x1f/;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			// Multi-line code — regex on line 2
			{
				Code: "var x = 1;\nvar r = /\\x1f/;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 9},
				},
			},

			// ── Multiple control chars in one pattern — one diagnostic, joined list ──
			{
				Code: `/[\x00-\x1f]/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
