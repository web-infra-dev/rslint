package dot_notation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestDotNotationRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		// Base rule
		{Code: "a.b;"},
		{Code: "a.b.c;"},
		{Code: "a['12'];"},
		{Code: "a[b];"},
		{Code: "a[0];"},
		{
			Code: "a.b.c;",
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: "a.arguments;",
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: "a.let;",
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: "a['while'];",
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: "a['true'];",
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: "a.true;",
			Options: map[string]interface{}{
				"allowKeywords": true,
			},
		},
		{
			Code: "a.null;",
			Options: map[string]interface{}{
				"allowKeywords": true,
			},
		},
		{
			Code: "a['snake_case'];",
			Options: map[string]interface{}{
				"allowPattern": "^[a-z]+(_[a-z]+)+$",
			},
		},
		{
			Code: "a['lots_of_snake_case'];",
			Options: map[string]interface{}{
				"allowPattern": "^[a-z]+(_[a-z]+)+$",
			},
		},
		{Code: "a[`time${range}`];"},
		{Code: "a[`time range`];"},
		{Code: "a.true;"},
		{Code: "a.null;"},
		{Code: "a[undefined];"},
		{Code: "a[void 0];"},
		{Code: "a[b()];"},
		// TypeScript specific
		{
			Code: `
class X {
  private priv_prop = 123;
}

const x = new X();
x['priv_prop'] = 123;
						`,
			Options: map[string]interface{}{
				"allowPrivateClassPropertyAccess": true,
			},
		},
		{
			Code: `
class X {
  protected protected_prop = 123;
}

const x = new X();
x['protected_prop'] = 123;
						`,
			Options: map[string]interface{}{
				"allowProtectedClassPropertyAccess": true,
			},
		},
		{
			Code: `
class X {
  prop: string;
  [key: string]: number;
}

const x = new X();
x['hello'] = 3;
						`,
			Options: map[string]interface{}{
				"allowIndexSignaturePropertyAccess": true,
			},
		},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: "a['true'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.true;"},
		},
		{
			Code: "a['b'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.b;"},
		},
		{
			Code: "a.b['c'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.b.c;"},
		},
		{
			Code: "a[null];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.null;"},
		},
		{
			Code: "a[true];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.true;"},
		},
		{
			Code: "a[false];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a.false;"},
		},
		{
			Code: "a['_dangle'];",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{"a._dangle;"},
			Options: map[string]interface{}{
				"allowPattern": "^[a-z]+(_[a-z]+)+$",
			},
		},
		{
			Code: "foo.while;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useBrackets",
				},
			},
			Output: []string{`foo["while"];`},
			Options: map[string]interface{}{
				"allowKeywords": false,
			},
		},
		{
			Code: `
class X {
  private priv_prop = 123;
}

const x = new X();
x['priv_prop'] = 123;
						`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{`
class X {
  private priv_prop = 123;
}

const x = new X();
x.priv_prop = 123;
						`},
			Options: map[string]interface{}{
				"allowPrivateClassPropertyAccess": false,
			},
		},
		{
			Code: `
class X {
  protected protected_prop = 123;
}

const x = new X();
x['protected_prop'] = 123;
						`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useDot",
				},
			},
			Output: []string{`
class X {
  protected protected_prop = 123;
}

const x = new X();
x.protected_prop = 123;
						`},
			Options: map[string]interface{}{
				"allowProtectedClassPropertyAccess": false,
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &DotNotationRule, validTestCases, invalidTestCases)
}
