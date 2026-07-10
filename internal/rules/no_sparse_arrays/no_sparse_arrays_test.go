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
			// Destructuring assignment targets are parsed as ArrayLiteralExpression
			// too, but omitted elements there are valid ES6 syntax for skipping
			// items, not sparse array literals.
			{Code: `[, suggestion] = await all();`},
			{Code: `[, , endLine, endChar] = o.range;`},
			{Code: `[, ref, authorName] = match;`},
			{Code: `[a, , b] = [1, 2, 3];`},
			{Code: `for ([, x] of y) {}`},
			{Code: `[[, a]] = b;`},
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
			{
				// The destructuring target `[a, b]` is not sparse, but the
				// RHS is a genuine sparse array literal and must still be flagged.
				Code: `[a, b] = [1, , 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 10},
				},
			},
		},
	)
}
