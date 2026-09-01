package prefer_named_capture_group

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferNamedCaptureGroupUpstream migrates the full valid/invalid suite
// from upstream tests/lib/rules/prefer-named-capture-group.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the prefer_named_capture_group_extras_test.go file.
func TestPreferNamedCaptureGroupUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferNamedCaptureGroupRule,
		[]rule_tester.ValidTestCase{
			{Code: "/normal_regex/"},
			{Code: "/(?:[0-9]{4})/"},
			{Code: "/(?<year>[0-9]{4})/"},
			{Code: `/\u{1F680}/u`},
			{Code: "new RegExp()"},
			{Code: "new RegExp(foo)"},
			{Code: "new RegExp('')"},
			{Code: "new RegExp('(?<year>[0-9]{4})')"},
			{Code: "RegExp()"},
			{Code: "RegExp(foo)"},
			{Code: "RegExp('')"},
			{Code: "RegExp('(?<year>[0-9]{4})')"},
			{Code: "RegExp('(')"}, // invalid regexp should be ignored
			{Code: `RegExp('\\u{1F680}', 'u')`},
			{
				Code:            "new globalThis.RegExp('([0-9]{4})')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2018},
			},
			{
				Code:            "new globalThis.RegExp('([0-9]{4})')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 6},
			},
			{
				Code:            "new globalThis.RegExp('([0-9]{4})')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2017},
			},
			{
				Code:            "new globalThis.RegExp()",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code:            "new globalThis.RegExp(foo)",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code:            "globalThis.RegExp(foo)",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code: `
                var globalThis = bar;
                globalThis.RegExp(foo);
                `,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code: `
                function foo () {
                    var globalThis = bar;
                    new globalThis.RegExp(baz);
                }
                `,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},

			// ES2024
			{Code: "new RegExp('(?<c>[[A--B]])', 'v')"},

			// Checks that the 'v' flag is understood: without it `([\q])` would
			// be a valid (unnamed-capturing-group) regex; with 'v' it's a
			// SyntaxError, so the pattern is skipped entirely.
			{Code: `new RegExp('([\\q])', 'v')`},

			// ES2025
			{
				Code:            "/(?i:foo)bar/",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{Code: "new RegExp('(?i:foo)bar')"},
			{
				Code:            "/(?-i:foo)bar/",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2025},
			},
			{Code: "new RegExp('(?-i:foo)bar')"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "/([0-9]{4})/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>[0-9]{4})/"},
							{MessageId: "addNonCapture", Output: "/(?:[0-9]{4})/"},
						},
					},
				},
			},
			{
				Code: "new RegExp('([0-9]{4})')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp('(?<temp1>[0-9]{4})')"},
							{MessageId: "addNonCapture", Output: "new RegExp('(?:[0-9]{4})')"},
						},
					},
				},
			},
			{
				Code: "RegExp('([0-9]{4})')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "RegExp('(?<temp1>[0-9]{4})')"},
							{MessageId: "addNonCapture", Output: "RegExp('(?:[0-9]{4})')"},
						},
					},
				},
			},
			{
				Code: "new RegExp(`a(bc)d`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(bc)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp(`a(?<temp1>bc)d`)"},
							{MessageId: "addNonCapture", Output: "new RegExp(`a(?:bc)d`)"},
						},
					},
				},
			},
			{
				Code: "new RegExp('ሴ噸(?:a)(b)');",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp('ሴ噸(?:a)(?<temp1>b)');"},
							{MessageId: "addNonCapture", Output: "new RegExp('ሴ噸(?:a)(?:b)');"},
						},
					},
				},
			},
			{
				Code: "new RegExp('\\u1234\\u5678(?:a)(b)');",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 35,
					},
				},
			},
			{
				Code: "/([0-9]{4})-(\\w{5})/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>[0-9]{4})-(\\w{5})/"},
							{MessageId: "addNonCapture", Output: "/(?:[0-9]{4})-(\\w{5})/"},
						},
					},
					{
						MessageId: "required",
						Message:   "Capture group '(\\w{5})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/([0-9]{4})-(?<temp1>\\w{5})/"},
							{MessageId: "addNonCapture", Output: "/([0-9]{4})-(?:\\w{5})/"},
						},
					},
				},
			},
			{
				Code: "/([0-9]{4})-(5)/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>[0-9]{4})-(5)/"},
							{MessageId: "addNonCapture", Output: "/(?:[0-9]{4})-(5)/"},
						},
					},
					{
						MessageId: "required",
						Message:   "Capture group '(5)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 17,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/([0-9]{4})-(?<temp1>5)/"},
							{MessageId: "addNonCapture", Output: "/([0-9]{4})-(?:5)/"},
						},
					},
				},
			},
			{
				Code: "/(?<temp2>(a))/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp2>(?<temp3>a))/"},
							{MessageId: "addNonCapture", Output: "/(?<temp2>(?:a))/"},
						},
					},
				},
			},
			{
				Code: "/(?<temp2>(a)(?<temp5>b))/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 27,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp2>(?<temp6>a)(?<temp5>b))/"},
							{MessageId: "addNonCapture", Output: "/(?<temp2>(?:a)(?<temp5>b))/"},
						},
					},
				},
			},
			{
				Code: "/(?<temp1>[0-9]{4})-(\\w{5})/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(\\w{5})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 29,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>[0-9]{4})-(?<temp2>\\w{5})/"},
							{MessageId: "addNonCapture", Output: "/(?<temp1>[0-9]{4})-(?:\\w{5})/"},
						},
					},
				},
			},
			{
				Code: "/(?<temp1>[0-9]{4})-(5)/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(5)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>[0-9]{4})-(?<temp2>5)/"},
							{MessageId: "addNonCapture", Output: "/(?<temp1>[0-9]{4})-(?:5)/"},
						},
					},
				},
			},
			{
				Code: "/(?<temp1>a)(?<temp2>a)(a)(?<temp3>a)/",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 39,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "/(?<temp1>a)(?<temp2>a)(?<temp4>a)(?<temp3>a)/"},
							{MessageId: "addNonCapture", Output: "/(?<temp1>a)(?<temp2>a)(?:a)(?<temp3>a)/"},
						},
					},
				},
			},
			{
				Code: "new RegExp('(' + 'a)')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 23,
					},
				},
			},
			{
				Code: "new RegExp('a(bc)d' + 'e')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(bc)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 27,
					},
				},
			},
			{
				Code: `new RegExp("foo" + "(a)" + "(b)");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 34,
					},
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 34,
					},
				},
			},
			{
				Code: `new RegExp("foo" + "(?:a)" + "(b)");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 36,
					},
				},
			},
			{
				Code: "RegExp('(a)'+'')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 17,
					},
				},
			},
			{
				Code: "RegExp( '' + '(ab)')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(ab)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
					},
				},
			},
			{
				Code: "new RegExp(`(ab)${''}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(ab)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 24,
					},
				},
			},
			{
				Code: "new RegExp(`(a)\n`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndLine:   2,
						EndColumn: 3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp(`(?<temp1>a)\n`)"},
							{MessageId: "addNonCapture", Output: "new RegExp(`(?:a)\n`)"},
						},
					},
				},
			},
			{
				Code: "RegExp(`a(b\nc)d`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b\nc)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndLine:   2,
						EndColumn: 6,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "RegExp(`a(?<temp1>b\nc)d`)"},
							{MessageId: "addNonCapture", Output: "RegExp(`a(?:b\nc)d`)"},
						},
					},
				},
			},
			{
				Code: `new RegExp('a(b)\'')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 21,
					},
				},
			},
			{
				Code: `RegExp('(a)\\d')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(a)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 17,
					},
				},
			},
			{
				Code: "RegExp(`\\a(b)`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '(b)' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 16,
					},
				},
			},
			{
				Code:            "new globalThis.RegExp('([0-9]{4})')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 36,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new globalThis.RegExp('(?<temp1>[0-9]{4})')"},
							{MessageId: "addNonCapture", Output: "new globalThis.RegExp('(?:[0-9]{4})')"},
						},
					},
				},
			},
			{
				Code:            "globalThis.RegExp('([0-9]{4})')",
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 32,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "globalThis.RegExp('(?<temp1>[0-9]{4})')"},
							{MessageId: "addNonCapture", Output: "globalThis.RegExp('(?:[0-9]{4})')"},
						},
					},
				},
			},
			{
				Code: `
                function foo() { var globalThis = bar; }
                new globalThis.RegExp('([0-9]{4})');
            `,
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([0-9]{4})' should be converted to a named or non-capturing group.",
						Line:      3,
						Column:    17,
						EndColumn: 52,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: `
                function foo() { var globalThis = bar; }
                new globalThis.RegExp('(?<temp1>[0-9]{4})');
            `},
							{MessageId: "addNonCapture", Output: `
                function foo() { var globalThis = bar; }
                new globalThis.RegExp('(?:[0-9]{4})');
            `},
						},
					},
				},
			},

			// ES2024
			{
				Code: "new RegExp('([[A--B]])', 'v')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "required",
						Message:   "Capture group '([[A--B]])' should be converted to a named or non-capturing group.",
						Line:      1,
						Column:    1,
						EndColumn: 30,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "addGroupName", Output: "new RegExp('(?<temp1>[[A--B]])', 'v')"},
							{MessageId: "addNonCapture", Output: "new RegExp('(?:[[A--B]])', 'v')"},
						},
					},
				},
			},
		},
	)
}
