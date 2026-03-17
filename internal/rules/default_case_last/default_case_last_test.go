package default_case_last

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDefaultCaseLastRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultCaseLastRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (foo) {}`},
			{Code: `switch (foo) { case 1: bar(); break; }`},
			{Code: `switch (foo) { case 1: break; case 2: break; }`},
			{Code: `switch (foo) { default: bar(); break; }`},
			{Code: `switch (foo) { default: }`},
			{Code: `switch (foo) { case 1: break; default: break; }`},
			{Code: `switch (foo) { case 1: break; default: }`},
			{Code: `switch (foo) { case 1: default: break; }`},
			{Code: `switch (foo) { case 1: default: }`},
			{Code: `switch (foo) { case 1: break; case 2: break; default: break; }`},
			{Code: `switch (foo) { case 1: break; case 2: default: break; }`},
			{Code: `switch (foo) { case 1: case 2: default: }`},
			{Code: `switch (foo) { case 1: break; }`},
			{Code: `switch (foo) { case 1: case 2: break; }`},
			{Code: `switch (foo) { case 1: baz(); break; case 2: quux(); break; default: quuux(); break; }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `switch (foo) { default: bar(); break; case 1: baz(); break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: break; case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { case 1: break; default: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 31},
				},
			},
			{
				Code: `switch (foo) { case 1: default: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { case 1: default: case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { case 1: default: case 2: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 24},
				},
			},
			{
				Code: `switch (foo) { default: break; case 1: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
			{
				Code: `switch (foo) { default: case 1: case 2: }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "notLast", Line: 1, Column: 16},
				},
			},
		},
	)
}
