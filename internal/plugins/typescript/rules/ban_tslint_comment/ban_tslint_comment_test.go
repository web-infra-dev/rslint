package ban_tslint_comment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTslintCommentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTslintCommentRule, []rule_tester.ValidTestCase{
		{Code: "let a: readonly any[] = [];"},
		{Code: "let a = new Array();"},
		{Code: "// some other comment"},
		{Code: "// TODO: this is a comment that mentions tslint"},
		{Code: "/* another comment that mentions tslint */"},
	}, []rule_tester.InvalidTestCase{
		{
			Code: "/* tslint:disable */",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "/* tslint:enable */",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "/* tslint:disable:rule1 rule2 rule3... */",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "/* tslint:enable:rule1 rule2 rule3... */",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "// tslint:disable-next-line",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "someCode(); // tslint:disable-line",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 13},
			},
			Output: []string{"someCode();"},
		},
		{
			Code: "function f() {\n  return/* tslint:disable-line */foo;\n}",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 2, Column: 9},
			},
			Output: []string{"function f() {\n  return foo;\n}"},
		},
		{
			Code: "// tslint:disable-next-line:rule1 rule2 rule3...",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{""},
		},
		{
			Code: "const whoa = doSomeStuff();\n// tslint:disable-line\nconsole.log(whoa);\n",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 2, Column: 1},
			},
			Output: []string{"const whoa = doSomeStuff();\nconsole.log(whoa);\n"},
		},
		{
			Code: "/* tslint:disable */ const x = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "commentDetected", Line: 1, Column: 1},
			},
			Output: []string{"const x = 1;"},
		},
	})
}
