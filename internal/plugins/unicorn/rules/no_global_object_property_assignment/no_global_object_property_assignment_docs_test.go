package no_global_object_property_assignment_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_global_object_property_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalObjectPropertyAssignmentDocs(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_global_object_property_assignment.NoGlobalObjectPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: "export const foo = value;"},
			{Code: "const singleton = {\n\tfoo: value,\n};"},
			{Code: "globalThis.foo;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "globalThis.foo = value;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}}},
			{Code: "window.foo += 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}}},
			{Code: "self.foo ||= value;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
		},
	)
}
