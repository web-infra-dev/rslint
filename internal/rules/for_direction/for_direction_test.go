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
			{Code: `for(var i = MIN; i <= MAX; i-=false){}`},
			{Code: `for(var i = MIN; i <= MAX; i-=0/0){}`},

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
			{Code: `for(var i = 0n; i < l; i-=-1n){}`},
			{Code: `for(var i = 10n; i >= 0n; i+=0n){}`},

			// Static expressions supported by ESLint's getStaticValue
			{Code: `for(var i = 0; i < 10; i+=+5e-7){}`},
			{Code: `for(var i = 0; i < MAX; i -= ~2);`},
			{Code: `for(var i = 0, n = -1; i < MAX; i += -n);`},
			{Code: `var n = -2; n = 2; for(var i = 0; i < 10; i += n);`},
			{Code: `var step = -1; { let step = 1; for(var i = 0; i < 10; i += step){} }`},
			{Code: `const a = b, b = a; for(var i = 0; i < 10; i += a){}`},
			{Code: `for(var i = 10; i > 0; i += (1 === true)){}`},
			{Code: `for(var i = 10n; i > 0n; i += (1n / 2n)){}`},
			{Code: `for(var i = 10n; i > 0n; i += (2n ** -1n)){}`},
			{Code: `for(var i = 10n; i > 0n; i += (9007199254740992n === 9007199254740993n)){}`},

			// Only direct condition identifiers are counters.
			{Code: `for(var i = 0; i + 1 < 10; i--){}`},

			// Sequence expressions with exactly one unambiguous counter update
			{Code: `for(var i = 0; i < 10; (i++, j++)){}`},
			{Code: `for(var i = 0; i < 10; (j++, i++)){}`},
			{Code: `for(var i = 10; i > 0; (j++, i--, k++)){}`},
			{Code: `for(var i = 0; i < 10; (j++, k++)){}`},

			// Multiple or ambiguous counter modifications are ignored.
			{Code: `for(var i = 10; i < 20; (i--, i++)){}`},
			{Code: `for(var i = 10; i < 20; (i--, i += 2)){}`},
			{Code: `for(var i = 10; i < 20; (i--, i += STEP_SIZE)){}`},
			{Code: `for(var i = 10; i < 20; (i--, (i = 5, j++))){}`},

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
					{MessageId: "incorrectDirection", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
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

			// Static RHS values, including effectively constant variables
			{
				Code: `for(var i = MIN; i <= MAX; i-=true){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0n; i > l; i+=1n){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; i-=+5e-7){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < MAX; i += (2 - 3));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `var n = -2; for(var i = 0; i < 10; i += n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 13, EndLine: 1, EndColumn: 43},
				},
			},
			{
				Code: `const step = -1; for(var i = 0; i < 10; i += step){} for(var j = 0; j < 10; j += step){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection"},
					{MessageId: "incorrectDirection"},
				},
			},
			{
				Code: `for(var i = 10; i > 0; i += (1 !== true)){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},
			{
				Code: `for(var i = 0; i < 10; i += (-1 as number)){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// The matching counter may be on the right even when both sides are identifiers.
			{
				Code: `for(var i = 0, limit = 10; limit > i; i--){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

			// Sequence expressions report only when exactly one counter mutation is wrong.
			{
				Code: `for(var i = 0; i < 10; (i--, j++)){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `for(var i = 10; i > 0; (j++, (i++, k++))){}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "incorrectDirection", Line: 1, Column: 1},
				},
			},

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
