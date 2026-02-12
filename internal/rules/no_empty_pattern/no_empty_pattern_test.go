package no_empty_pattern

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyPatternRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyPatternRule,
		[]rule_tester.ValidTestCase{
			{Code: `var {a} = obj;`},
			{Code: `var [a] = arr;`},
			{Code: `var {a = 1} = obj;`},
			{Code: `var [a = 1] = arr;`},
			{Code: `function foo({a}) {}`},
			{Code: `function foo([a]) {}`},
			{Code: `var {a: {b}} = obj;`},
			{Code: `var {a: [b]} = obj;`},
			// allowObjectPatternsAsParameters: direct parameter
			{
				Code:    `function foo({}) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
			},
			// allowObjectPatternsAsParameters: parameter with empty object default
			{
				Code:    `function foo({} = {}) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `var {} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},
			{
				Code: `var [] = arr;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
				},
			},
			{
				Code: `var {a: {}} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `var {a: []} = obj;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 9},
				},
			},
			{
				Code: `function foo({}) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			{
				Code: `function foo([]) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			// allowObjectPatternsAsParameters: non-empty default should still report
			{
				Code:    `function foo({} = {a: 1}) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
			// allowObjectPatternsAsParameters: non-object default should still report
			{
				Code:    `function foo({} = bar) {}`,
				Options: map[string]interface{}{"allowObjectPatternsAsParameters": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 14},
				},
			},
		},
	)
}
