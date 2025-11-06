package ban_types

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTypesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BanTypesRule,
		[]rule_tester.ValidTestCase{
			// Valid: Using primitive types
			{Code: `let a: string;`},
			{Code: `let b: number;`},
			{Code: `let c: boolean;`},
			{Code: `let d: symbol;`},
			{Code: `let e: bigint;`},

			// Valid: Object literal types
			{Code: `let f: { x: number; y: number } = { x: 1, y: 1 };`},

			// Valid: With custom options disabling defaults
			{
				Code:    `let a: String;`,
				Options: map[string]interface{}{"extendDefaults": false},
			},
		},
		[]rule_tester.InvalidTestCase{
			// Invalid: Using wrapper types
			{
				Code: `let a: String;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
				Output: []string{`let a: string;`},
			},
			{
				Code: `let b: Number;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
				Output: []string{`let b: number;`},
			},
			{
				Code: `let c: Boolean;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
				Output: []string{`let c: boolean;`},
			},

			// Invalid: Function type
			{
				Code: `let f: Function;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
			},

			// Invalid: Object type
			{
				Code: `let o: Object;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
			},

			// Invalid: Custom banned type
			{
				Code: `let a: String;`,
				Options: map[string]interface{}{
					"extendDefaults": false,
					"types": map[string]interface{}{
						"String": map[string]interface{}{
							"message": "Use string instead.",
							"fixWith": "string",
						},
					},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "bannedTypeMessage",
						Line:      1,
						Column:    8,
					},
				},
				Output: []string{`let a: string;`},
			},
		},
	)
}
