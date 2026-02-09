package member_ordering

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestMemberOrderingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &MemberOrderingRule, []rule_tester.ValidTestCase{
		{
			Code: `class Foo {
  a: string;
  constructor() {}
  b(): void {}
}`,
		},
		{
			Code: `class Foo {
  [A: string]: any;
  [a: string]: any;
  static C: boolean;
  static d: boolean;
  b: any;
  B: any;
  get e(): string {}
  get E(): string {}
  private imPrivate() {}
  private ImPrivate() {}
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": map[string]interface{}{
						"order": "alphabetically-case-insensitive",
					},
				},
			},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `class Foo {
  method() {}
  field: string;
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": []interface{}{"field", "method"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectGroupOrder",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `interface Foo {
  b: string;
  a: string;
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes": "never",
						"order":       "alphabetically",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectOrder",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `interface Foo {
  b?: string;
  a: string;
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes":      "never",
						"optionalityOrder": "required-first",
						"order":            "as-written",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectRequiredMembersOrder",
					Line:      2,
					Column:    3,
				},
			},
		},
		{
			Code: `class Foo {
  method() {}
  static {}
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": []interface{}{"static-initialization", "method"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectGroupOrder",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `class Foo {
  a: string;
  readonly b: string;
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": []interface{}{"readonly-field", "field"},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectGroupOrder",
					Line:      3,
					Column:    3,
				},
			},
		},
		{
			Code: `interface Foo {
  a1: string;
  a10: string;
  a5: string;
}`,
			Options: []interface{}{
				map[string]interface{}{
					"default": map[string]interface{}{
						"memberTypes": "never",
						"order":       "natural",
					},
				},
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "incorrectOrder",
					Line:      4,
					Column:    3,
				},
			},
		},
	})
}
