package eqeqeq

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEqeqeqRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&EqeqeqRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Default "always" mode - strict equality is valid
			{Code: `a === b`},
			{Code: `a !== b`},
			{Code: `a === null`},
			{Code: `null !== a`},
			{Code: `a === b`, Options: "always"},

			// "smart" mode
			{Code: `typeof a == 'number'`, Options: "smart"},
			{Code: `'string' != typeof a`, Options: "smart"},
			{Code: `typeof a == typeof b`, Options: "smart"},
			{Code: `null == a`, Options: "smart"},
			{Code: `a != null`, Options: "smart"},
			{Code: `'hello' != 'world'`, Options: "smart"},
			{Code: `2 == 3`, Options: "smart"},
			{Code: `0 == 0`, Options: "smart"},
			{Code: `true == true`, Options: "smart"},
			{Code: `null == null`, Options: "smart"},
			{Code: `a === b`, Options: "smart"},

			// "allow-null" (same as ["always", {"null": "ignore"}])
			{Code: `a == null`, Options: "allow-null"},
			{Code: `null == a`, Options: "allow-null"},
			{Code: `a != null`, Options: "allow-null"},
			{Code: `null != a`, Options: "allow-null"},
			{Code: `a === b`, Options: "allow-null"},

			// "always" with null:"ignore"
			{Code: `a == null`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `null == a`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `a != null`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `null != a`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `a === b`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `a !== null`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},

			// "always" with null:"always"
			{Code: `a === null`, Options: []interface{}{"always", map[string]interface{}{"null": "always"}}},
			{Code: `a !== null`, Options: []interface{}{"always", map[string]interface{}{"null": "always"}}},
			{Code: `null === null`, Options: []interface{}{"always", map[string]interface{}{"null": "always"}}},
			{Code: `null !== null`, Options: []interface{}{"always", map[string]interface{}{"null": "always"}}},

			// "always" with null:"never" - loose null checks are valid
			{Code: `a == null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `null == a`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `a != null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `null == null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `null != null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},

			// Parenthesized null — must unwrap parens to detect null
			{Code: `(null) == a`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `a == (null)`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `((null)) == a`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `a != (null)`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `(null) != a`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},

			// "smart" mode — parenthesized null, typeof, same-type literals
			{Code: `(null) == a`, Options: "smart"},
			{Code: `a != (null)`, Options: "smart"},
			{Code: `(typeof a) == 'number'`, Options: "smart"},
			{Code: `('hello') == ('world')`, Options: "smart"},
			{Code: `(1) == (2)`, Options: "smart"},
			{Code: `(true) == (false)`, Options: "smart"},

			// Nesting contexts — strict equality in various positions
			{Code: `if (a === b) {}`},
			{Code: `a === b ? 1 : 0`},
			{Code: `while (a === b) {}`},
			{Code: `for (; a === b; ) {}`},
			{Code: `var f = () => a === b`},

			// TS-specific — as / non-null assertion
			{Code: `(a as any) === b`},
			{Code: `a! === b`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Default "always" mode - suggestions (not autofix) for non-typeof/non-literal
			{
				Code: `a == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b`},
					}},
				},
			},
			{
				Code: `a != b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a !== b`},
					}},
				},
			},
			// typeof - autofix
			{
				Code:   `typeof a == 'number'`,
				Output: []string{`typeof a === 'number'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code:   `'string' != typeof a`,
				Output: []string{`'string' !== typeof a`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			// same-type literals - autofix
			{
				Code:   `true == true`,
				Output: []string{`true === true`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:   `2 == 3`,
				Output: []string{`2 === 3`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code:   `'hello' != 'world'`,
				Output: []string{`'hello' !== 'world'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			// null comparison - suggestion
			{
				Code: `a == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === null`},
					}},
				},
			},
			{
				Code: `null != a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `null !== a`},
					}},
				},
			},

			// "smart" mode - non-exempted loose equality (suggestions)
			{
				Code:    `a == b`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b`},
					}},
				},
			},
			{
				Code:    `a != b`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a !== b`},
					}},
				},
			},
			// smart mode - cross-type literals (suggestions)
			{
				Code:    `true == 1`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `true === 1`},
					}},
				},
			},
			{
				Code:    `0 != '1'`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `0 !== '1'`},
					}},
				},
			},

			// "allow-null" - non-null loose equality (typeof gets autofix, others get suggestions)
			{
				Code:    `a == b`,
				Options: "allow-null",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b`},
					}},
				},
			},
			{
				Code:    `typeof a == 'number'`,
				Options: "allow-null",
				Output:  []string{`typeof a === 'number'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code:    `'hello' != 'world'`,
				Options: "allow-null",
				Output:  []string{`'hello' !== 'world'`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code:    `2 == 3`,
				Options: "allow-null",
				Output:  []string{`2 === 3`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code:    `true == true`,
				Options: "allow-null",
				Output:  []string{`true === true`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// "always" with null:"always" - null comparisons flagged (null==null gets autofix, others get suggestions)
			{
				Code:    `true == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `true === null`},
					}},
				},
			},
			{
				Code:    `true != null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `true !== null`},
					}},
				},
			},
			{
				Code:    `null == null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Output:  []string{`null === null`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    `null != null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "always"}},
				Output:  []string{`null !== null`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// "always" with null:"ignore" - non-null loose equality still flagged
			{
				Code:    `a == b`,
				Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b`},
					}},
				},
			},

			// "always" with null:"never" - strict null checks flagged (suggestions),
			// null===null gets autofix, non-null loose equality gets suggestion
			{
				Code:    `a === null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a == null`},
					}},
				},
			},
			{
				Code:    `null !== a`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `null != a`},
					}},
				},
			},
			{
				Code:    `null === null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Output:  []string{`null == null`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    `null !== null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Output:  []string{`null != null`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code:    `a == b`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b`},
					}},
				},
			},

			// Parenthesized expressions (suggestions)
			{
				Code: `(a) == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a) === b`},
					}},
				},
			},
			{
				Code: `a == (b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === (b)`},
					}},
				},
			},
			{
				Code: `(a) == (b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a) === (b)`},
					}},
				},
			},
			{
				Code: `(a == b) == (c)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a == b) === (c)`},
					}},
					{MessageId: "unexpected", Line: 1, Column: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a === b) == (c)`},
					}},
				},
			},

			// ═══ Parenthesized null — null:"never" must unwrap parens ═══
			{
				Code:    `(null) === a`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(null) == a`},
					}},
				},
			},
			{
				Code:    `a === (null)`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a == (null)`},
					}},
				},
			},
			{
				Code:    `((null)) === a`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `((null)) == a`},
					}},
				},
			},

			// ═══ Nesting contexts — ensure rule fires everywhere ═══
			{
				Code: `if (a == b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `if (a === b) {}`},
					}},
				},
			},
			{
				Code: `a == b ? 1 : 0`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a === b ? 1 : 0`},
					}},
				},
			},
			{
				Code: `while (a == b) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `while (a === b) {}`},
					}},
				},
			},
			{
				Code: `for (; a == b; ) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `for (; a === b; ) {}`},
					}},
				},
			},
			{
				Code: `var f = () => a == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `var f = () => a === b`},
					}},
				},
			},
			{
				Code: `var o = { key: a == b }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `var o = { key: a === b }`},
					}},
				},
			},
			{
				Code: `var a = [a == b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `var a = [a === b]`},
					}},
				},
			},
			{
				Code: `(a == b) && (c != d)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a === b) && (c != d)`},
					}},
					{MessageId: "unexpected", Line: 1, Column: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a == b) && (c !== d)`},
					}},
				},
			},
			{
				Code: `console.log(a == b)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `console.log(a === b)`},
					}},
				},
			},
			{
				Code: "class C { m() { return a == b; } }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: "class C { m() { return a === b; } }"},
					}},
				},
			},

			// ═══ Whitespace / multiline ═══
			{
				Code: "a==b",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 2, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: "a===b"},
					}},
				},
			},
			{
				Code: "a  ==  b",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: "a  ===  b"},
					}},
				},
			},
			{
				Code: "a\n==\nb",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: "a\n===\nb"},
					}},
				},
			},

			// ═══ Special operands ═══
			{
				Code: `undefined == a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `undefined === a`},
					}},
				},
			},
			{
				Code: `void 0 == a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 8, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `void 0 === a`},
					}},
				},
			},
			{
				Code: `NaN == NaN`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `NaN === NaN`},
					}},
				},
			},
			{
				Code: `this == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `this === null`},
					}},
				},
			},
			{
				Code: `foo() == bar()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 7, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `foo() === bar()`},
					}},
				},
			},

			// ═══ TS-specific: as / non-null assertion ═══
			{
				Code: `(a as any) == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a as any) === b`},
					}},
				},
			},
			{
				Code: `a! == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 4, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `a! === b`},
					}},
				},
			},

			// ═══ Smart mode — parenthesized typeof/literals still exempt ═══
			{
				Code:    `(a) == b`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "replaceOperator", Output: `(a) === b`},
					}},
				},
			},
		},
	)
}
