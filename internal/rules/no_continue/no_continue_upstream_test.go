// TestNoContinueUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/no-continue.js 1:1, plus the incorrect/correct
// examples shown on the rule's documentation page. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_continue_extras_test.go file.
package no_continue

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoContinueUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoContinueRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: `var sum = 0, i; for(i = 0; i < 10; i++){ if(i > 5) { sum += i; } }`},
			{Code: `var sum = 0, i = 0; while(i < 10) { if(i > 5) { sum += i; } i++; }`},
			// ---- Doc examples: correct ----
			{Code: `let sum = 0,
    i;

for(i = 0; i < 10; i++) {
    if(i < 5) {
       sum += i;
    }
}`},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code: `var sum = 0, i; for(i = 0; i < 10; i++){ if(i <= 5) { continue; } sum += i; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Message: "Unexpected use of continue statement.", Line: 1, Column: 55, EndLine: 1, EndColumn: 64},
				},
			},
			{
				Code: `var sum = 0, i; myLabel: for(i = 0; i < 10; i++){ if(i <= 5) { continue myLabel; } sum += i; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 64, EndLine: 1, EndColumn: 81},
				},
			},
			{
				Code: `var sum = 0, i = 0; while(i < 10) { if(i <= 5) { i++; continue; } sum += i; i++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 55, EndLine: 1, EndColumn: 64},
				},
			},
			{
				Code: `var sum = 0, i = 0; myLabel: while(i < 10) { if(i <= 5) { i++; continue myLabel; } sum += i; i++; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 64, EndLine: 1, EndColumn: 81},
				},
			},
			// ---- Doc examples: incorrect ----
			{
				Code: `let sum = 0,
    i;

for(i = 0; i < 10; i++) {
    if(i >= 5) {
        continue;
    }

    sum += i;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 6, Column: 9, EndLine: 6, EndColumn: 18},
				},
			},
			{
				Code: `let sum = 0,
    i;

labeledLoop: for(i = 0; i < 10; i++) {
    if(i >= 5) {
        continue labeledLoop;
    }

    sum += i;
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 6, Column: 9, EndLine: 6, EndColumn: 30},
				},
			},
		},
	)
}
