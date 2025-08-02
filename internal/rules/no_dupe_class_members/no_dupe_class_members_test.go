package no_dupe_class_members

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoDupeClassMembersRule(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{Code: `
class A {
  foo() {}
  bar() {}
}`},
		{Code: `
class A {
  static foo() {}
  foo() {}
}`},
		{Code: `
class A {
  get foo() {}
  set foo(value) {}
}`},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		{
			Code: `
class A {
  foo() {}
  foo() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{Line: 4, Column: 3, MessageId: "unexpected"},
			},
		},
		{
			Code: `
class A {
  static foo() {}
  static foo() {}
}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{Line: 4, Column: 3, MessageId: "unexpected"},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDupeClassMembersRule, validTestCases, invalidTestCases)
}
