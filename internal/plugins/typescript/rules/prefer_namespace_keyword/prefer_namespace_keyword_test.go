package prefer_namespace_keyword

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNamespaceKeywordRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNamespaceKeywordRule, []rule_tester.ValidTestCase{
		{Code: `declare module 'foo';`},
		{Code: `declare module 'foo' {}`},
		{Code: `namespace foo {}`},
		{Code: `declare namespace foo {}`},
		{Code: `declare global {}`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`namespace foo {}`},
		},
		{
			Code: `declare module foo {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
					Line:      1,
					Column:    1,
				},
			},
			Output: []string{`declare namespace foo {}`},
		},
		{
			Code: `
declare module foo {
  declare module bar {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "useNamespace",
				},
				{
					MessageId: "useNamespace",
				},
			},
			Output: []string{`
declare namespace foo {
  declare namespace bar {}
}`},
		},
	})
}
