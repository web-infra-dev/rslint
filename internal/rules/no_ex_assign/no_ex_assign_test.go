package no_ex_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoExAssignRule,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: "try { } catch (e) { three = 2 + 1; }"},
			{Code: "try { } catch ({e}) { this.something = 2; }"},
			{Code: "function foo() { try { } catch (e) { return false; } }"},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: "try { } catch (e) { e = 10; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 21},
				},
			},
			{
				Code: "try { } catch (ex) { ex = 10; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: "try { } catch (ex) { [ex] = []; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 22},
				},
			},
			{
				Code: "try { } catch (ex) { ({x: ex = 0} = {}); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 27},
				},
			},
			{
				Code: "try { } catch ({message}) { message = 10; }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 29},
				},
			},
		},
	)
}
