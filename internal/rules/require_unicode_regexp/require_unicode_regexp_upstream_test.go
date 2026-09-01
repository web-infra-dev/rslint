// TestRequireUnicodeRegexpUpstream migrates the full valid/invalid suite from
// ESLint v10.8.1 tests/lib/rules/require-unicode-regexp.js 1:1. Upstream's
// RuleTester default is languageOptions.ecmaVersion 2015; cases that rely on
// that default rather than overriding it are migrated without an explicit
// ECMAVersion, since rslint's zero-value default ("latest") behaves
// identically here — every case in this suite that is sensitive to the u/v
// availability boundary (ecmaVersion <= 5 or <= 2023) sets ecmaVersion
// explicitly upstream too. The one case that relies on `sourceType: commonjs`
// only to make `global` a declared identifier is migrated with an equivalent
// `Globals: {"global": ...}` entry instead, since rslint does not model
// sourceType-driven implicit globals — see the "Scope: rule semantics, not
// framework parity" note in the port-rule skill. rslint-specific lock-ins
// live in require_unicode_regexp_extras_test.go.
package require_unicode_regexp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRequireUnicodeRegexpUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RequireUnicodeRegexpRule,
		[]rule_tester.ValidTestCase{
			{Code: "/foo/u"},
			{Code: "/foo/gimuy"},
			{Code: "RegExp('', 'u')"},
			{Code: "RegExp('', `u`)"},
			{Code: "new RegExp('', 'u')"},
			{Code: "RegExp('', 'gimuy')"},
			{Code: "RegExp('', `gimuy`)"},
			{Code: "RegExp(...patternAndFlags)"},
			{Code: "new RegExp('', 'gimuy')"},
			{Code: "const flags = 'u'; new RegExp('', flags)"},
			{Code: "const flags = 'g'; new RegExp('', flags + 'u')"},
			{Code: "const flags = 'gimu'; new RegExp('foo', flags[3])"},
			{Code: "new RegExp('', flags)"},
			{Code: "function f(flags) { return new RegExp('', flags) }"},
			{Code: "function f(RegExp) { return new RegExp('foo') }"},
			{Code: "function f(patternAndFlags) { return new RegExp(...patternAndFlags) }"},
			{Code: "new globalThis.RegExp('foo')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}},
			{Code: "new globalThis.RegExp('foo')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2017}},
			{Code: "new globalThis.RegExp('foo', 'u')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: "globalThis.RegExp('foo', 'u')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: "const flags = 'u'; new globalThis.RegExp('', flags)", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: "const flags = 'g'; new globalThis.RegExp('', flags + 'u')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: "const flags = 'gimu'; new globalThis.RegExp('foo', flags[3])", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020}},
			{Code: "class C { #RegExp; foo() { new globalThis.#RegExp('foo') } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "/foo/u", Options: []any{map[string]any{"requireFlag": "u"}}},
			{Code: "new RegExp('foo', 'u')", Options: []any{map[string]any{"requireFlag": "u"}}},

			// require the `v` flag.
			{Code: "/foo/v", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "/foo/gimvy", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "RegExp('', 'v')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "RegExp('', `v`)", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "new RegExp('', 'v')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "RegExp('', 'gimvy')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "RegExp('', `gimvy`)", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "new RegExp('', 'gimvy')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "const flags = 'v'; new RegExp('', flags)", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "const flags = 'g'; new RegExp('', flags + 'v')", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "const flags = 'gimv'; new RegExp('foo', flags[3])", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "/foo/v", Options: []any{map[string]any{"requireFlag": "v"}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
			{Code: "new RegExp('foo', 'v')", Options: []any{map[string]any{"requireFlag": "v"}}, LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "/\\a/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: "/foo/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/foo/u"}},
					},
				},
			},
			{
				Code: "/foo/gimy",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 10,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/foo/gimyu"}},
					},
				},
			},
			{
				Code: "RegExp()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code: "RegExp('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 14,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `RegExp('foo', "u")`}},
					},
				},
			},
			{
				Code: "RegExp('\\\\a')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "RegExp('foo', '')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "RegExp('foo', 'u')"}},
					},
				},
			},
			{
				Code: "RegExp('foo', 'gimy')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "RegExp('foo', 'gimyu')"}},
					},
				},
			},
			{
				Code: "RegExp('foo', `gimy`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "RegExp('foo', `gimyu`)"}},
					},
				},
			},
			{
				Code: "new RegExp('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp('foo', "u")`}},
					},
				},
			},
			{
				Code:            "new RegExp('foo',)",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2017},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp('foo', "u",)`}},
					},
				},
			},
			{
				Code: "new RegExp('foo', false)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code: "new RegExp('foo', 1)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code: "new RegExp('foo', '')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 22,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', 'u')"}},
					},
				},
			},
			{
				Code: "new RegExp('foo', 'gimy')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 26,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', 'gimyu')"}},
					},
				},
			},
			{
				Code: "new RegExp(('foo'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 20,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp(('foo'), "u")`}},
					},
				},
			},
			{
				Code: "new RegExp(('unrelated', 'foo'))",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new RegExp(('unrelated', 'foo'), "u")`}},
					},
				},
			},
			{
				Code: "const flags = 'gi'; new RegExp('foo', flags)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 21, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code: "const flags = 'gi'; new RegExp('foo', ('unrelated', flags))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 21, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code: "let flags; new RegExp('foo', flags = 'g')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 12, EndLine: 1, EndColumn: 42},
				},
			},
			{
				Code: "const flags = `gi`; new RegExp(`foo`, (`unrelated`, flags))",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 21, EndLine: 1, EndColumn: 60},
				},
			},
			{
				Code: "const flags = 'gimu'; new RegExp('foo', flags[0])",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireUFlag", Line: 1, Column: 23, EndLine: 1, EndColumn: 50},
				},
			},
			{
				Code:    "new window.RegExp('foo')",
				Globals: map[string]any{"window": "readonly"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new window.RegExp('foo', "u")`}},
					},
				},
			},
			{
				// Migrated from upstream's `languageOptions: { sourceType: "commonjs" }`
				// case, which relies on sourceType alone to declare `global` — see the
				// file header note.
				Code:    "new global.RegExp('foo')",
				Globals: map[string]any{"global": "writable"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new global.RegExp('foo', "u")`}},
					},
				},
			},
			{
				Code:            "new globalThis.RegExp('foo')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: `new globalThis.RegExp('foo', "u")`}},
					},
				},
			},
			{
				Code:            "/foo/",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/foo/v"}},
					},
				},
			},
			{
				Code:            "/foo/u",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "/foo/v"}},
					},
				},
			},
			{
				Code:            "/foo/u",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
				},
			},
			{
				Code:            "/[[a]/u",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 8},
				},
			},
			{
				Code:            "new RegExp('foo', 'u')",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "new RegExp('foo', 'v')"}},
					},
				},
			},
			{
				Code:            "new RegExp('[[a]', 'u')",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:            "new RegExp(\"foo\", \"\\u0067\")",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "new RegExp(\"foo\", \"\\u0067v\")"}},
					},
				},
			},
			{
				Code:            "new RegExp(\"foo\", `\\u0067`)",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "new RegExp(\"foo\", `\\u0067v`)"}},
					},
				},
			},
			{
				Code:            "new RegExp(\"foo\", \"\\u0075\")",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:            "new RegExp(\"foo\", `\\u0075`)",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:            "const regularFlags = \"sm\"; new RegExp(\"foo\", `${regularFlags}g`)",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 28, EndLine: 1, EndColumn: 65,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: "const regularFlags = \"sm\"; new RegExp(\"foo\", `${regularFlags}gv`)"}},
					},
				},
			},
			{
				Code:            "const regularFlags = \"smu\"; new RegExp(\"foo\", `${regularFlags}g`)",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "requireVFlag", Line: 1, Column: 29, EndLine: 1, EndColumn: 66},
				},
			},
			{
				Code:            "/foo/v",
				Options:         []any{map[string]any{"requireFlag": "u"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 7,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "/foo/u"}},
					},
				},
			},
			{
				Code:            "new RegExp('foo')",
				Options:         []any{map[string]any{"requireFlag": "v"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireVFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 18,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addVFlag", Output: `new RegExp('foo', "v")`}},
					},
				},
			},
			{
				Code:            "new RegExp('foo', 'v')",
				Options:         []any{map[string]any{"requireFlag": "u"}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2024},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "requireUFlag", Line: 1, Column: 1, EndLine: 1, EndColumn: 23,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "addUFlag", Output: "new RegExp('foo', 'u')"}},
					},
				},
			},
		},
	)
}
