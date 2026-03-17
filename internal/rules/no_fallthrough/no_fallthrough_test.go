package no_fallthrough

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoFallthroughRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoFallthroughRule,
		[]rule_tester.ValidTestCase{
			// Break exits the case
			{Code: `switch(foo) { case 0: a(); break; case 1: b(); }`},
			// Empty case (no statements), allowed
			{Code: `switch(foo) { case 0: case 1: a(); break; }`},
			// Comment suppresses warning
			{Code: `switch(foo) { case 0: a(); /* falls through */ case 1: b(); }`},
			// Return exits the case
			{Code: `function foo() { switch(bar) { case 0: a(); return; case 1: b(); } }`},
			// Throw exits the case
			{Code: `switch(foo) { case 0: a(); throw e; case 1: b(); }`},
			// Continue exits the case
			{Code: `while(a) { switch(foo) { case 0: a(); continue; case 1: b(); } }`},
			// "fall through" also works (case-insensitive)
			{Code: `switch(foo) { case 0: a(); // Fall Through` + "\n" + `case 1: b(); }`},
			// Last case doesn't need break
			{Code: `switch(foo) { case 0: a(); }`},
			// Multiple empty cases before a case with break
			{Code: `switch(foo) { case 0: case 1: case 2: a(); break; }`},
			// If/else both branches terminate
			{Code: `switch(foo) { case 0: if (a) { break; } else { break; } case 1: b(); }`},
			// Block with break at the end
			{Code: `switch(foo) { case 0: { a(); break; } case 1: b(); }`},
			// Default with break
			{Code: `switch(foo) { case 0: a(); break; default: b(); break; case 1: c(); }`},
		},
		[]rule_tester.InvalidTestCase{
			// Fallthrough from case to case
			{
				Code: "switch(foo) { case 0: a();\ncase 1: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 2, Column: 1},
				},
			},
			// Fallthrough from case to default
			{
				Code: "switch(foo) { case 0: a();\ndefault: b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "default", Line: 2, Column: 1},
				},
			},
			// Multiline fallthrough
			{
				Code: "switch(foo) { case 0:\n  a();\ncase 1:\n  b(); }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "case", Line: 3, Column: 1},
				},
			},
		},
	)
}
