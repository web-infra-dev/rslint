package no_global_object_property_assignment_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_global_object_property_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalObjectPropertyAssignmentExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_global_object_property_assignment.NoGlobalObjectPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: "function f(globalThis) { globalThis.foo = 1; }"},
			{Code: "const window = {}; window.foo = 1;"},
			{Code: "delete window.foo;"},
			{Code: "globalThis[property] = value;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "globalThis[\"foo\"]++;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			{Code: "for (self.value of values) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 6, EndLine: 1, EndColumn: 16}}},
			{Code: "({target: window.value} = source);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23}}},
		},
	)
}
