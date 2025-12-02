package no_sparse_arrays

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSparseArraysRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSparseArraysRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: `var a = [ 1, 2, ]`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var a = [,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 9},
				},
			},
			{
				Code: `var a = [ 1,, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 9},
				},
			},
			// This test case is commented out because it produces a TypeScript compilation error:
			// error creating TS program for /tsconfig.json: found 5 syntactic errors. [Invalid character. [\r\n\t/* comment */,\n// comment\n ,]; Invalid character. [\r\n\t/* comment */,\n// comment\n ,]; Invalid character. [\r\n\t/* comment */,\n// comment\n ,]; Invalid character. [\r\n\t/* comment */,\n// comment\n ,]; ']' expected. [\r\n\t/* comment */,\n// comment\n ,];]
			//{
			//	Code: `[\r\n\t/* comment */,\n// comment\n ,];`,
			//	Errors: []rule_tester.InvalidTestCaseError{
			//		{MessageId: "unexpectedSparseArray", Line: 1, Column: 9},
			//	},
			//},
			{
				Code: `[(( [a,] )),,,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 1},
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 1},
				},
			},
			{
				Code: `[,(( [a,] )),,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 1},
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 1},
				},
			},
		},
	)
}
