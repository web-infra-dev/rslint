package use_isnan

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUseIsNaNRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&UseIsNaNRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Simple assignments and non-comparison usage
			{Code: `var x = NaN;`},
			{Code: `isNaN(NaN) === true;`},
			{Code: `Number.isNaN(NaN) === true;`},
			{Code: `isNaN(123);`},
			{Code: `Number.isNaN(123);`},

			// Non-comparison binary operators
			{Code: `NaN + 1;`},
			{Code: `1 + NaN;`},

			// switch with NaN but enforceForSwitchCase disabled
			{
				Code:    `switch(NaN) { case foo: break; }`,
				Options: map[string]interface{}{"enforceForSwitchCase": false},
			},
			{
				Code:    `switch(foo) { case NaN: break; }`,
				Options: map[string]interface{}{"enforceForSwitchCase": false},
			},

			// indexOf/lastIndexOf without enforceForIndexOf (default false)
			{Code: `foo.indexOf(NaN)`},
			{Code: `foo.lastIndexOf(NaN)`},

			// indexOf/lastIndexOf with non-NaN argument and enforceForIndexOf enabled
			{
				Code:    `foo.indexOf(bar)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},
			{
				Code:    `foo.lastIndexOf(bar)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
			},

			// Regular switch (no NaN)
			{Code: `switch(foo) { case bar: break; }`},
			{Code: `switch(true) { case true: break; }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Binary comparisons with NaN
			{
				Code: `123 == NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `123 === NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `NaN === "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `NaN == "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `123 != NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `123 !== NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `NaN < "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `"abc" < NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `NaN > "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `"abc" >= NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `NaN <= "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},

			// Number.NaN comparisons
			{
				Code: `123 === Number.NaN;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `Number.NaN === "abc";`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "comparisonWithNaN", Line: 1, Column: 1},
				},
			},

			// switch(NaN) - default enforceForSwitchCase is true
			{
				Code: `switch(NaN) { case foo: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchNaN", Line: 1, Column: 1},
				},
			},
			{
				Code: `switch(Number.NaN) { case foo: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "switchNaN", Line: 1, Column: 1},
				},
			},

			// case NaN - default enforceForSwitchCase is true
			{
				Code: `switch(foo) { case NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "caseNaN", Line: 1, Column: 15},
				},
			},
			{
				Code: `switch(foo) { case Number.NaN: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "caseNaN", Line: 1, Column: 15},
				},
			},

			// indexOf/lastIndexOf with enforceForIndexOf enabled
			{
				Code:    `foo.indexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "indexOfNaN", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo.lastIndexOf(NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "indexOfNaN", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo.indexOf(Number.NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "indexOfNaN", Line: 1, Column: 1},
				},
			},
			{
				Code:    `foo.lastIndexOf(Number.NaN)`,
				Options: map[string]interface{}{"enforceForIndexOf": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "indexOfNaN", Line: 1, Column: 1},
				},
			},
		},
	)
}
