// TestNoPlusplusUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-plusplus.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case. rslint-specific
// lock-in cases live in the no_plusplus_extras_test.go file.
package no_plusplus

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var allowForLoopAfterthoughts = []any{map[string]any{"allowForLoopAfterthoughts": true}}

func TestNoPlusplusUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoPlusplusRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `var foo = 0; foo=+1;`},

			// With "allowForLoopAfterthoughts" allowed
			{Code: `var foo = 0; foo=+1;`, Options: allowForLoopAfterthoughts},
			{Code: `for (i = 0; i < l; i++) { console.log(i); }`, Options: allowForLoopAfterthoughts},
			{Code: `for (var i = 0, j = i + 1; j < example.length; i++, j++) {}`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; i--, foo());`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; foo(), --i);`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; foo(), ++i, bar);`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; i++, (++j, k--));`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; foo(), (bar(), i++), baz());`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; (--i, j += 2), bar = j + 1);`, Options: allowForLoopAfterthoughts},
			{Code: `for (;; a, (i--, (b, ++j, c)), d);`, Options: allowForLoopAfterthoughts},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code: `var foo = 0; foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `var foo = 0; foo--;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '--' used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: `for (i = 0; i < l; i++) { console.log(i); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 20, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: `for (i = 0; i < l; foo, i++) { console.log(i); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 25, EndLine: 1, EndColumn: 28},
				},
			},

			// With "allowForLoopAfterthoughts" allowed
			{
				Code:    `var foo = 0; foo++;`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 14, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `for (i = 0; i < l; i++) { v++; }`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 27, EndLine: 1, EndColumn: 30},
				},
			},
			{
				// Locks in upstream isForStatementUpdate: node.parent.type must be
				// "ForStatement" AND parent.update === node — this i++ sits in the
				// init slot, not the update slot.
				Code:    `for (i++;;);`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 6, EndLine: 1, EndColumn: 9},
				},
			},
			{
				// Same as above, but the condition slot.
				Code:    `for (;--i;);`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '--' used.", Line: 1, Column: 7, EndLine: 1, EndColumn: 10},
				},
			},
			{
				// The update is in the loop body, not the afterthought slot.
				Code:    `for (;;) ++i;`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
				},
			},
			{
				// Locks in upstream isForLoopAfterthought base case: the update's
				// parent is an AssignmentExpression, not the ForStatement directly.
				Code:    `for (;; i = j++);`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 13, EndLine: 1, EndColumn: 16},
				},
			},
			{
				// The outer comma expression is the afterthought, but --j sits
				// inside a CallExpression argument, not directly in the comma chain.
				Code:    `for (;; i++, f(--j));`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '--' used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
				},
			},
			{
				// i++ sits inside a parenthesised comma expression that is itself
				// the right operand of a BinaryExpression (`+`), not the
				// afterthought slot.
				Code:    `for (;; foo + (i++, bar));`,
				Options: allowForLoopAfterthoughts,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedUnaryOp", Message: "Unary operator '++' used.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
				},
			},
		},
	)
}
