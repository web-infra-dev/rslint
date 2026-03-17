package no_unreachable

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnreachableRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnreachableRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Function declaration after return is hoisted
			{Code: `function foo() { return bar(); function bar() { return 1; } }`},
			// Conditional return does not make subsequent code unreachable
			{Code: `function foo() { if (x) { return; } bar(); }`},
			// Normal code before return
			{Code: `function foo() { var x = 1; return x; }`},
			// No statements after break
			{Code: `while (true) { break; }`},
			// No statements after continue
			{Code: `while (true) { continue; }`},
			// No statements after return
			{Code: `function foo() { return; }`},
			// No statements after throw
			{Code: `function foo() { throw new Error(); }`},
			// var without initializer after return is allowed (hoisted)
			{Code: `function foo() { return; var x; }`},
			// Empty statement after return is allowed
			{Code: `function foo() { return; ; }`},
			// Switch with break in case
			{Code: `switch (x) { case 1: break; }`},
			// Multiple var declarations without initializers
			{Code: `function foo() { return; var x, y, z; }`},
			// Function declaration after throw
			{Code: `function foo() { throw new Error(); function bar() {} }`},
			// Normal if/else without full coverage
			{Code: `function foo() { if (x) { return; } bar(); }`},
			{Code: `function foo() { if (x) { return; } else { bar(); } baz(); }`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Unreachable after return
			{
				Code: `function foo() { return; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// Unreachable after throw
			{
				Code: `function foo() { throw new Error(); x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 37},
				},
			},
			// Unreachable after break
			{
				Code: `while (true) { break; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 23},
				},
			},
			// Unreachable after continue
			{
				Code: `while (true) { continue; x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// var with initializer after return IS reported
			{
				Code: `function foo() { return; var x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// Multiple unreachable statements
			{
				Code: `function foo() { return; x = 1; y = 2; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
					{MessageId: "unreachableCode", Line: 1, Column: 33},
				},
			},
			// Unreachable in switch case
			{
				Code: `switch (x) { case 1: return; foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 30},
				},
			},
			// Unreachable in default clause
			{
				Code: `switch (x) { default: return; foo(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 31},
				},
			},
			// let declaration after return is reported (not hoisted)
			{
				Code: `function foo() { return; let x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// const declaration after return is reported (not hoisted)
			{
				Code: `function foo() { return; const x = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// Class declaration after return is reported
			{
				Code: `function foo() { return; class Bar {} }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 1, Column: 26},
				},
			},
			// Multiline unreachable code
			{
				Code: "function foo() {\n  return;\n  x = 1;\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unreachableCode", Line: 3, Column: 3},
				},
			},
		},
	)
}
