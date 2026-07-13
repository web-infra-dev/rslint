package prefer_regex_literals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferRegexLiteralsExtras locks in branches and edge shapes that the upstream test suite doesn't exercise. Each case carries an inline comment pointing at the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future refactors can't silently regress them without breaking a named lock-in.
func TestPreferRegexLiteralsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferRegexLiteralsRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: dynamic arguments remain allowed ----
			{Code: "const suffix = 'b'; new RegExp('a' + suffix);"},
			{Code: "const flags = 'g'; RegExp('a', flags);"},

			// ---- Real-user: eslint/eslint#13843 dynamic template and concatenation stay valid ----
			{Code: "const value = 'x'; new RegExp(`foo${value}`, 'gm');"},
			{Code: "const value = 'x'; new RegExp('foo' + value, 'gm');"},

			// ---- Dimension 4: access/key forms that are not static RegExp references ----
			{Code: "const key = 'RegExp'; globalThis[key]('a');"},
			{Code: "const raw = 'raw'; new RegExp(String[raw]`a`);"},

			// ---- Dimension 4: shadowed globals are not built-ins ----
			{Code: "const RegExp = function() {}; RegExp('a');"},
			{Code: "const String = { raw() { return 'a'; } }; new RegExp(String.raw`a`);"},
			{Code: "function f(globalThis) { globalThis.RegExp('a'); }"},

			// ---- Dimension 4: invalid argument count stays outside this rule ----
			{Code: "new RegExp('a', 'g', 'extra');"},

			// ---- Options: direct map shape with default-equivalent false ----
			{
				Code:    "new RegExp(/a/);",
				Options: map[string]interface{}{"disallowRedundantWrapping": false},
			},

			// N/A: declaration/container forms do not affect this expression-only rule.
			// N/A: object/property declaration key forms are not inspected by this rule.
			// N/A: class/function nesting boundaries do not create rule-owned state.
			// N/A: body-absent TS declarations are not RegExp call expressions.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: parenthesized callee ----
			{
				Code: "((RegExp))('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Message:   "Use a regular expression literal instead of the 'RegExp' constructor.",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},

			// ---- Dimension 4: TS assertion wrappers around callee and arguments ----
			{
				Code: "(RegExp as any)('a' as string, 'g' satisfies string);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/g;",
					}},
				}},
			},
			{
				Code: "(RegExp!)('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},

			// ---- Dimension 4: optional call chain still references global RegExp ----
			{
				Code: "RegExp?.('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},
			{
				Code: "globalThis?.RegExp?.('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},

			// ---- Dimension 4: global object property and element access ----
			{
				Code: "new (globalThis.RegExp)('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},
			{
				Code: "new ((globalThis as any).RegExp)('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},
			{
				Code: "new window['RegExp']('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},
			{
				Code: "globalThis[`RegExp`]('a');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},

			// Locks in upstream isStringRawTaggedStaticTemplateLiteral arm: element-access String.raw.
			{
				Code: "new RegExp(String['raw']`\\\\d`, String['raw']`g`);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/\\\\d/g;",
					}},
				}},
			},

			// ---- Real-user: eslint/eslint#16504 disallowRedundantWrapping reports regex literal args ----
			{
				Code:    "new RegExp((/a/) as RegExp);",
				Options: map[string]interface{}{"disallowRedundantWrapping": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRedundantRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "/a/;",
					}},
				}},
			},
			{
				Code:    "new RegExp(/a/i, 'g');",
				Options: map[string]interface{}{"disallowRedundantWrapping": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRedundantRegExpWithFlags",
					Message:   "Use regular expression literal with flags instead of the 'RegExp' constructor.",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceWithLiteralAndFlags", Output: "/a/g;"},
						{MessageId: "replaceWithIntendedLiteralAndFlags", Output: "/a/ig;"},
					},
				}},
			},

			// Locks in upstream canFixTo() comment arm: report without suggestions.
			{
				Code: "new RegExp('a' /* comment */);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
				}},
			},
			{
				Code: "new RegExp('https://example.com');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    `/https:\/\/example.com/;`,
					}},
				}},
			},
			{
				Code: "new RegExp(String.raw`// not a comment`);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    `/\/\/ not a comment/;`,
					}},
				}},
			},
			// The tsgo scanner validates the actual replacement literal; invalid
			// patterns/flags still report, but don't offer an unsafe suggestion.
			{
				Code: `new RegExp('\\p{NotAProperty}', 'u');`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
				}},
			},
			{
				Code: "new RegExp('a', '-');",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    1,
				}},
			},

			// Locks in upstream getSafeOutput() arm: prevent slash/keyword token fusion.
			{
				Code: "a/RegExp('foo')in b",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "unexpectedRegExp",
					Line:      1,
					Column:    3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "replaceWithLiteral",
						Output:    "a/ /foo/ in b",
					}},
				}},
			},
		},
	)
}
