package default_case

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestDefaultCaseRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&DefaultCaseRule,
		[]rule_tester.ValidTestCase{
			{Code: `switch (a) { case 1: break; default: break; }`},
			{Code: `switch (a) { case 1: break; case 2: default: break; }`},
			{Code: "switch (a) {\n  case 1:\n    break;\n  // no default\n}"},
			{Code: `switch (a) { case 1: break; /* no default */ }`},
			{Code: "switch (a) {\n  case 1:\n    break;\n  // No Default\n}"},
			{Code: `switch (a) {}`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `switch (a) { case 1: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
			{
				Code: `switch (a) { case 1: break; case 2: break; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missingDefaultCase", Line: 1, Column: 1},
				},
			},
		},
	)
}
