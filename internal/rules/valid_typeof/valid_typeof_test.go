//nolint:misspell // cspell:ignore strnig undefned nunber fucntion
package valid_typeof

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidTypeofRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ValidTypeofRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// All valid typeof comparison values
			{Code: `typeof foo === "string"`},
			{Code: `typeof foo === "object"`},
			{Code: `typeof foo === "function"`},
			{Code: `typeof foo === "undefined"`},
			{Code: `typeof foo === "boolean"`},
			{Code: `typeof foo === "number"`},
			{Code: `typeof foo === "bigint"`},
			{Code: `typeof foo === "symbol"`},

			// Reversed operands
			{Code: `"string" === typeof foo`},
			{Code: `"object" === typeof foo`},

			// typeof compared to typeof (always valid)
			{Code: `typeof foo === typeof bar`},

			// Non-equality operators are not checked
			{Code: `typeof foo > "string"`},

			// Without requireStringLiterals, non-string comparisons are OK (except bare undefined)
			{Code: `typeof foo === baz`},
			{Code: `typeof foo === Object`},

			// Not a comparison (typeof in other contexts)
			{Code: `var x = typeof foo`},
			{Code: `typeof foo`},

			// With requireStringLiterals: valid cases still valid
			{Code: `typeof foo === "string"`, Options: map[string]interface{}{"requireStringLiterals": true}},
			{Code: `typeof foo === typeof bar`, Options: map[string]interface{}{"requireStringLiterals": true}},
			{Code: `"undefined" === typeof foo`, Options: map[string]interface{}{"requireStringLiterals": true}},

			// != and !== with valid strings
			{Code: `typeof foo !== "string"`},
			{Code: `typeof foo != "function"`},
			{Code: `typeof foo == "number"`},

			// Static template literals with valid values
			{Code: "typeof foo === `string`"},
			{Code: "typeof foo === `object`"},
			{Code: "typeof foo === `undefined`"},

			// Parenthesized expressions with valid values
			{Code: `(typeof foo) === "string"`},
			{Code: `typeof foo === ("string")`},
			{Code: `((typeof foo)) === "string"`},
			{Code: `(typeof foo) === ("string")`},

			// Locally shadowed undefined — not reported without requireStringLiterals
			{Code: `function f(undefined: string) { typeof foo === undefined }`},
			{Code: `{ const undefined = "test"; typeof foo === undefined }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Invalid typeof comparison value (misspelled)
			{
				Code: `typeof foo === "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Reversed operands with invalid value
			{
				Code: `"strnig" === typeof foo`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 1},
				},
			},
			// !== with invalid value
			{
				Code: `typeof foo !== "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// == with invalid value
			{
				Code: `typeof foo == "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 15},
				},
			},
			// != with invalid value
			{
				Code: `typeof foo != "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 15},
				},
			},
			// Bare undefined identifier without requireStringLiterals → invalidValue + suggestion
			{
				Code: `typeof foo === undefined`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestString", Output: `typeof foo === "undefined"`},
					}},
				},
			},
			// Bare undefined identifier with requireStringLiterals → notString + suggestion
			{
				Code:    `typeof foo === undefined`,
				Options: map[string]interface{}{"requireStringLiterals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString", Line: 1, Column: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "suggestString", Output: `typeof foo === "undefined"`},
					}},
				},
			},
			// Non-string identifier with requireStringLiterals → notString
			{
				Code:    `typeof foo === Object`,
				Options: map[string]interface{}{"requireStringLiterals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString", Line: 1, Column: 16},
				},
			},
			// Completely invalid string
			{
				Code: `typeof foo === "foobar"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Empty string
			{
				Code: `typeof foo === ""`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Static template literal with invalid value
			{
				Code: "typeof foo === `strnig`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Static template literal with completely invalid value
			{
				Code: "typeof foo === `foobar`",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Non-string literals: null, number, boolean, regex
			{
				Code: `typeof foo === null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			{
				Code: `typeof foo === 42`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			{
				Code: `typeof foo === true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			{
				Code: `typeof foo === false`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			{
				Code: `typeof foo === /regex/`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 16},
				},
			},
			// Shadowed undefined with requireStringLiterals → notString (no suggestion)
			{
				Code:    `function f(undefined: string) { typeof foo === undefined }`,
				Options: map[string]interface{}{"requireStringLiterals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notString", Line: 1, Column: 48},
				},
			},
			// Parenthesized expressions with invalid values
			{
				Code: `(typeof foo) === "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 18},
				},
			},
			{
				Code: `typeof foo === ("strnig")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 17},
				},
			},
			{
				Code: `((typeof foo)) === "strnig"`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 20},
				},
			},
			{
				Code: `typeof foo === (("strnig"))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 18},
				},
			},
			{
				Code: `(typeof foo) === ("strnig")`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "invalidValue", Line: 1, Column: 19},
				},
			},
		},
	)
}
