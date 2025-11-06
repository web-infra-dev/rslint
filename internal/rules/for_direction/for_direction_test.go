package for_direction

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestForDirectionRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&ForDirectionRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			// Correct increment with <
			{Code: `for(var i = 0; i < 10; i++){}`},
			{Code: `for(var i = 0; i < 10; ++i){}`},
			{Code: `for(var i = 0; i < 10; i+=1){}`},
			{Code: `for(var i = 0; i < 10; i-=-1){}`},
			{Code: `for(var i = 0; i < 10; i-=(-1)){}`},

			// Correct increment with <=
			{Code: `for(var i = 0; i <= 10; i++){}`},
			{Code: `for(var i = 0; i <= 10; ++i){}`},
			{Code: `for(var i = 0; i <= 10; i+=1){}`},
			{Code: `for(var i = 0; i <= 10; i-=-1){}`},
			{Code: `for(var i = 0; i <= 10; i-=(-1)){}`},

			// Correct decrement with >
			{Code: `for(var i = 10; i > 0; i--){}`},
			{Code: `for(var i = 10; i > 0; --i){}`},
			{Code: `for(var i = 10; i > 0; i-=1){}`},
			{Code: `for(var i = 10; i > 0; i+=-1){}`},
			{Code: `for(var i = 10; i > 0; i+=(-1)){}`},

			// Correct decrement with >=
			{Code: `for(var i = 10; i >= 0; i--){}`},
			{Code: `for(var i = 10; i >= 0; --i){}`},
			{Code: `for(var i = 10; i >= 0; i-=1){}`},
			{Code: `for(var i = 10; i >= 0; i+=-1){}`},
			{Code: `for(var i = 10; i >= 0; i+=(-1)){}`},

			// Reversed comparison - counter on right side
			{Code: `for(var i = 0; 10 > i; i++){}`},
			{Code: `for(var i = 0; 10 >= i; i++){}`},
			{Code: `for(var i = 10; 0 < i; i--){}`},
			{Code: `for(var i = 10; 0 <= i; i--){}`},

			// Unknown direction - dynamic values
			{Code: `for(var i = 0; i < 10; i+=x){}`},
			{Code: `for(var i = 0; i < 10; i-=x){}`},
			{Code: `for(var i = 10; i > 0; i+=x){}`},
			{Code: `for(var i = 10; i > 0; i-=x){}`},
			{Code: `for(var i = MIN; i <= MAX; i++){}`},
			{Code: `for(var i = MIN; i <= MAX; i+=true){}`},
			{Code: `for(var i = MIN; i <= MAX; i-=true){}`},

			// Neutral increment (no effect)
			{Code: `for(var i = 10; i >= 0; i-=0){}`},
			{Code: `for(var i = 10; i >= 0; i+=0){}`},
			{Code: `for(var i = 0; i < 10; i-=0){}`},
			{Code: `for(var i = 0; i < 10; i+=0){}`},

			// No update clause
			{Code: `for(var i = 10; i > 0;){}`},
			{Code: `for(var i = 0; i < 10;){}`},

			// Different variable in update
			{Code: `for(var i = 10; i > 0; j++){}`},
			{Code: `for(var i = 10; i > 0; j--){}`},
			{Code: `for(var i = 0; i < 10; j++){}`},
			{Code: `for(var i = 0; i < 10; j--){}`},

			// Non-comparison operators in test
			{Code: `for(var i = 0; i !== 10; i++){}`},
			{Code: `for(var i = 0; i === 10; i++){}`},
			{Code: `for(var i = 0; i != 10; i++){}`},
			{Code: `for(var i = 0; i == 10; i++){}`},

			// BigInt literals (if supported)
			{Code: `for(var i = 0n; i < 10n; i++){}`},
			{Code: `for(var i = 0n; i > l; i-=1n){}`},

			// TypeScript syntax with type annotations
			{Code: `for(let i: number = 0; i < 10; i++){}`},
			{Code: `for(let i: number = 10; i >= 0; i--){}`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			// Wrong direction with <
			{
				Code: `for(var i = 0; i < 10; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; --i){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; i-=1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; i+=-1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Wrong direction with <=
			{
				Code: `for(var i = 0; i <= 10; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i <= 10; --i){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i <= 10; i-=1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i <= 10; i+=-1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Wrong direction with >
			{
				Code: `for(var i = 10; i > 0; i++){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i > 0; ++i){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i > 0; i+=1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i > 0; i-=-1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Wrong direction with >=
			{
				Code: `for(var i = 10; i >= 0; i++){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i >= 0; ++i){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i >= 0; i+=1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; i >= 0; i-=-1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Reversed comparison with wrong direction
			{
				Code: `for(var i = 0; 10 > i; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; 10 >= i; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; 0 < i; i++){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 10; 0 <= i; i++){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Static negative value in compound assignment
			{
				Code: `for(var i = 0; i < 10; i+=-1){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; i+=(-1)){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// TODO: With const variable - requires full scope-based constant resolution
			// which is not yet implemented. This test is skipped for now.
			// {
			// 	Code: `const n = -2; for(var i = 0; i < 10; i+=n){}`,
			// 	Errors: []rule_tester.InvalidTestCaseError{
			// 		{MessageId: "incorrectDirection", Line: 1, Column: 15},
			// 	},
			// },

			// TypeScript syntax with wrong direction
			{
				Code: `for(let i: number = 0; i < 10; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(let i: number = 10; i >= 0; i++){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
		},
	)
}
