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
			{Code: `({ value: [, a] } = b);`},
			{Code: `[...[, a]] = b;`},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: `var a = [,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 10, EndLine: 1, EndColumn: 11},
				},
			},
			{
				Code: `var a = [ 1,, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
				},
			},
			{
				Code: "[\r\n\t/* comment */,\n// comment\n ,];",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 2, Column: 15, EndLine: 2, EndColumn: 16},
					{MessageId: "unexpectedSparseArray", Line: 4, Column: 2, EndLine: 4, EndColumn: 3},
				},
			},
			{
				Code: `[(( [a,] )),,,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				Code: `[,(( [a,] )),,];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				// The destructuring target `[a, b]` is not sparse, but the
				// RHS is a genuine sparse array literal and must still be flagged.
				Code: `[a, b] = [1, , 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
				},
			},
			{
				// An array literal used as a destructuring default remains a real
				// array expression, even though the surrounding array is a target.
				Code: `[value = [,]] = source;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code: `const emoji = "😀"; const values = [1,, 2];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedSparseArray", Line: 1, Column: 39, EndLine: 1, EndColumn: 40},
				},
			},
		},
	)
}
