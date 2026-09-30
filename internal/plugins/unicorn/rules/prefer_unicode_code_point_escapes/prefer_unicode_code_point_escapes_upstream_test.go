// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/prefer-unicode-code-point-escapes.js
// Documentation examples: https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/docs/rules/prefer-unicode-code-point-escapes.md
package prefer_unicode_code_point_escapes_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_unicode_code_point_escapes"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// The five script-mode cases marked Skip are rejected by tsgo (TS1487/TS1488).
// TestPreferUnicodeCodePointEscapesLegacyScripts below checks their rule behavior.
func TestPreferUnicodeCodePointEscapesUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_unicode_code_point_escapes.PreferUnicodeCodePointEscapesRule,
		[]rule_tester.ValidTestCase{
			{Code: "const foo = '\\u{7A}'", FileName: "test.js"},
			{Code: "const foo = '\\u{1F4A9}'", FileName: "test.js"},
			{Code: "const foo = '\\n\\t\\r\\\\\\'\\\"'", FileName: "test.js"},
			{Code: "const foo = '\\0'", FileName: "test.js"},
			{Code: "const foo = '\\8\\9\\08'", FileName: "test.js", Skip: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "const foo = '\\\\u2661'", FileName: "test.js"},
			{Code: "const foo = '\\\\x7A'", FileName: "test.js"},
			{Code: "const foo = tag`\\u2661`", FileName: "test.js"},
			{Code: "const foo = tag`\\123`", FileName: "test.js"},
			{Code: "const foo = String.raw`\\u2661`", FileName: "test.js"},
			{Code: "const foo = /\\u{61}/u", FileName: "test.js"},
			{Code: "const foo = /\\u{61}/v", FileName: "test.js"},
			{Code: "const foo = /[\\uD83D\\uDCA9]/u", FileName: "test.js"},
			{Code: "const foo = /[\\uD83D\\uDCA9]/v", FileName: "test.js"},
			{Code: "const foo = /[[\\uD83D\\uDCA9]\\uD83D\\uDCA9]/v", FileName: "test.js"},
			{Code: "const foo = /\\u{XYZ}/", FileName: "test.js"},
			{Code: "const foo = /\\u{110000}/", FileName: "test.js"},
			{Code: "const foo = /\\u{}/", FileName: "test.js"},
			{Code: "const foo = /\\u{/", FileName: "test.js"},
			{Code: "const foo = /\\cK/", FileName: "test.js"},
			{Code: "const foo = new RegExp(\"\\\\u0061\")", FileName: "test.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const foo = '\\x7A'", FileName: "test.js", Output: []string{"const foo = '\\u{7A}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = \"\\x7A\"", FileName: "test.js", Output: []string{"const foo = \"\\u{7A}\""}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = '\\xa9'", FileName: "test.js", Output: []string{"const foo = '\\u{A9}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = '\\u2661'", FileName: "test.js", Output: []string{"const foo = '\\u{2661}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = '\\uD83D\\uDCA9'", FileName: "test.js", Output: []string{"const foo = '\\u{1F4A9}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 27}}},
			{Code: "const foo = '\\123'", FileName: "test.js", Skip: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const foo = '\\u{53}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = '\\00'", FileName: "test.js", Skip: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const foo = '\\u{0}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 18}}},
			{Code: "const foo = '\\1\\12\\123\\4\\45'", FileName: "test.js", Skip: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const foo = '\\u{1}\\u{A}\\u{53}\\u{4}\\u{25}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 29}}},
			{Code: "const foo = '\\400'", FileName: "test.js", Skip: true, LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const foo = '\\u{20}0'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = '\\x7A\\u2661\\uD83D\\uDCA9'", FileName: "test.js", Output: []string{"const foo = '\\u{7A}\\u{2661}\\u{1F4A9}'"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 37}}},
			{Code: "const foo = `\\x7A${bar}\\u2661`", FileName: "test.js", Output: []string{"const foo = `\\u{7A}${bar}\\u{2661}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31}}},
			{Code: "const foo = `\\\\\\x7A`", FileName: "test.js", Output: []string{"const foo = `\\\\\\u{7A}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = /\\x7A/u", FileName: "test.js", Output: []string{"const foo = /\\u{7A}/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20}}},
			{Code: "const foo = /\\u0061/v", FileName: "test.js", Output: []string{"const foo = /\\u{61}/v"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
			{Code: "const foo = /\\uD83D\\uDCA9/u", FileName: "test.js", Output: []string{"const foo = /\\u{1F4A9}/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 28}}},
			{Code: "const foo = /\\[\\uD83D\\uDCA9/u", FileName: "test.js", Output: []string{"const foo = /\\[\\u{1F4A9}/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 30}}},
			{Code: "const foo = /[\\x2D]/u", FileName: "test.js", Output: []string{"const foo = /[\\u{2D}]/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
			{Code: "const foo = /[\\cA]/u", FileName: "test.js", Output: []string{"const foo = /[\\u{1}]/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = /\\cA/u", FileName: "test.js", Output: []string{"const foo = /\\u{1}/u"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}}},
			{Code: "const foo = /\\cA/", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const foo = /\\u{1}/u"}}}}},
			{Code: "const foo = /\\u0061/", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const foo = /\\u{61}/u"}}}}},
			{Code: "const foo = /\\u{61}/", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const foo = /\\u{61}/u"}}}}},
			{Code: "const foo = /\\x7A/g", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const foo = /\\u{7A}/gu"}}}}},
			{Code: "const foo = /\\x61\\_/", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = /\\u{61}\\_/", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23}}},
		},
	)
}

func TestPreferUnicodeCodePointEscapesDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_unicode_code_point_escapes.PreferUnicodeCodePointEscapesRule,
		[]rule_tester.ValidTestCase{
			{Code: "const foo = '\\u{7A}';\nconst bar = '\\u{2661}';\nconst baz = '\\u{1F4A9}';", FileName: "test.js"},
			{Code: "const foo = `\\u{7A}${bar}\\u{2661}`;", FileName: "test.js"},
			{Code: "const foo = /\\u{61}/u;", FileName: "test.js"},
			{Code: "const foo = /\\u{61}/u;", FileName: "test.js"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const foo = '\\x7A';\nconst bar = '\\u2661';\nconst baz = '\\uD83D\\uDCA9';\n", FileName: "test.js", Output: []string{"const foo = '\\u{7A}';\nconst bar = '\\u{2661}';\nconst baz = '\\u{1F4A9}';\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 2, Column: 13, EndLine: 2, EndColumn: 21}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 3, Column: 13, EndLine: 3, EndColumn: 27}}},
			{Code: "const foo = `\\x7A${bar}\\u2661`;\n", FileName: "test.js", Output: []string{"const foo = `\\u{7A}${bar}\\u{2661}`;\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20}, {MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 23, EndLine: 1, EndColumn: 31}}},
			{Code: "const foo = /\\u0061/u;\n", FileName: "test.js", Output: []string{"const foo = /\\u{61}/u;\n"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
			{Code: "const foo = /\\u0061/;\n", FileName: "test.js", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-unicode-code-point-escapes", Message: "Prefer Unicode code point escapes.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-unicode-code-point-escapes/add-unicode-flag", Output: "const foo = /\\u{61}/u;\n"}}}}},
		},
	)
}

// Exercise the five retained script-mode cases directly on the parsed AST:
// RuleTester rejects tsgo's legacy-escape syntax diagnostics before rule dispatch.
func TestPreferUnicodeCodePointEscapesLegacyScripts(t *testing.T) {
	for _, test := range []struct{ code, output string }{
		{`const foo = '\8\9\08'`, `const foo = '\8\9\08'`},
		{`const foo = '\123'`, `const foo = '\u{53}'`},
		{`const foo = '\00'`, `const foo = '\u{0}'`},
		{`const foo = '\1\12\123\4\45'`, `const foo = '\u{1}\u{A}\u{53}\u{4}\u{25}'`},
		{`const foo = '\400'`, `const foo = '\u{20}0'`},
	} {
		t.Run(test.code, func(t *testing.T) {
			diagnostics := lintEscapesSource(test.code, rule.EditDemandAll)
			wantCount := 1
			if test.code == test.output {
				wantCount = 0
			}
			if len(diagnostics) != wantCount {
				t.Fatalf("got %d diagnostics, want %d", len(diagnostics), wantCount)
			}
			if wantCount != 0 {
				d := diagnostics[0]
				if d.Message.Id != "prefer-unicode-code-point-escapes" || d.Message.Description != "Prefer Unicode code point escapes." || d.Range.Pos() != 12 || d.Range.End() != len(test.code) || d.Suggestions != nil {
					t.Fatalf("unexpected diagnostic: %#v", d)
				}
			}
			output, _, fixed := linter.ApplyRuleFixes(test.code, diagnostics)
			if output != test.output || fixed != (wantCount != 0) {
				t.Fatalf("fixed = %t, output = %q; want %q", fixed, output, test.output)
			}
		})
	}
}
