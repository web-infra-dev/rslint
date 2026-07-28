package no_case_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCaseDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoCaseDeclarationsRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (a) { case 1: { let x = 1; break; } }`},
			{Code: `switch (a) { case 1: { const x = 1; break; } }`},
			{Code: `switch (a) { case 1: { function f() {} break; } }`},
			{Code: `switch (a) { case 1: { class C {} break; } }`},
			{Code: `switch (a) { case 1: var x = 1; break; }`},
			{Code: `switch (a) { default: var x = 1; break; }`},
			{Code: `switch (a) { case 1: break; }`},
			{Code: `switch (a) {}`},
			{Code: `switch (a) { case 1: if (a) { const x = 1; } break; }`},
			{Code: `switch (a) { case 1: type T = string; interface I { value: T } break; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `switch (a) { case 1: let x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: `switch (a) { case 1: const x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: `switch (a) { case 1: function f() {} break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: `switch (a) { case 1: class C {} break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: `switch (a) { default: let x = 1; break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 23},
				},
			},
			{
				Code: `switch (a) {
  case 1:
    let x = 1;
    const y = 2;
    break;
  case 2:
  default:
    class C {}
}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 3, Column: 5},
					{MessageId: "unexpected", Line: 4, Column: 5},
					{MessageId: "unexpected", Line: 8, Column: 5},
				},
			},
			{
				Code: `switch (a) { case 1: using resource = acquire(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
			{
				Code: `async function f() { switch (a) { case 1: await using resource = acquire(); break; } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected"},
				},
			},
		},
	)
}
