package init_declarations

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestInitDeclarationsRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &InitDeclarationsRule, []rule_tester.ValidTestCase{
		{Code: "var foo = null;"},
		{Code: "for (var i = 0; i < 1; i++) {}"},
		{Code: "for (var foo in []) {}"},
		{Code: "let a = true;", Options: []interface{}{"always"}},
		{Code: "var foo;", Options: []interface{}{"never"}},
		{Code: "const foo = 1;", Options: []interface{}{"never"}},
		{
			Code:    "for (var i = 0; i < 1; i++) {}",
			Options: []interface{}{"never", map[string]interface{}{"ignoreForLoopInit": true}},
		},
		{
			Code:    "for (var foo in []) {}",
			Options: []interface{}{"never", map[string]interface{}{"ignoreForLoopInit": true}},
		},
		{
			Code:    "const { a } = { a: 1 };",
			Options: []interface{}{"never"},
		},
		{
			Code:    "const [a] = [1];",
			Options: []interface{}{"never"},
		},
		{Code: "declare const foo: number;", Options: []interface{}{"always"}},
		{
			Code: `
declare namespace myLib {
  let numberOfGreetings: number;
}
      `,
			Options: []interface{}{"always"},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:    "var foo;",
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "initialized",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 8,
				},
			},
		},
		{
			Code: "let a;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "initialized",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 6,
				},
			},
		},
		{
			Code:    "var foo = true;",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notInitialized",
					Line:      1,
					Column:    5,
					EndLine:   1,
					EndColumn: 15,
				},
			},
		},
		{
			Code:    "for (var i = 0; i < 1; i++) {}",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notInitialized",
				},
			},
		},
		{
			Code:    "for (var foo in []) {}",
			Options: []interface{}{"never"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "notInitialized",
				},
			},
		},
		{
			Code: `
namespace myLib {
  let numberOfGreetings: number;
}
      `,
			Options: []interface{}{"always"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "initialized",
				},
			},
		},
	})
}
