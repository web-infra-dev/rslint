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
			{Code: "[foo([globalThis.foo]).bar] = values;"},
			{Code: "[foo({value: globalThis.foo}).bar] = values;"},
			{Code: "({[globalThis.foo]: value} = source);"},
			{Code: "[target = globalThis.foo] = values;"},
			{Code: "globalThis.foo.bar = value;"},
			{Code: "globalThis[Symbol.for(key)] = 1;"},
			{Code: "function f(Symbol) { globalThis[Symbol.iterator] = 1; }"},
			{Code: "function f(Symbol) { globalThis[Symbol.for('foo')] = 1; }"},
			{Code: "const key = unknown; globalThis[key] = 1;"},
			{Code: "const a = b; const b = a; globalThis[a] = 1;"},
			{Code: "let key = Symbol.iterator; key = unknown; globalThis[key] = 1;"},
			{Code: "globalThis[Symbol.iterator];"},
			{Code: "function f(globalThis) { globalThis.foo = 1; }"},
			{Code: "const window = {}; window.foo = 1;"},
			{Code: "delete window.foo;"},
			{Code: "globalThis[property] = value;"},
			{Code: "getGlobal().foo = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "((globalThis.foo!) as any)++;", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "((globalThis.foo as any) satisfies any) = value;", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[Symbol.for('foo')] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[Symbol.iterator] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[Symbol['iterator']] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[Symbol['for']('foo')] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[Symbol.for()] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "const key = Symbol.iterator; globalThis[key] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "const key = Symbol.for('foo'); globalThis[key]++;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "const a = Symbol.iterator; const b = a; globalThis[b] = 1;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "for ([globalThis.foo] of values) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "[globalThis.foo = value] = values;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "({target: [globalThis.foo]} = source);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "(globalThis.foo as any) = value;", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object."}}},
			{Code: "globalThis[\"foo\"]++;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			{Code: "for (self.value of values) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 6, EndLine: 1, EndColumn: 16}}},
			{Code: "({target: window.value} = source);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 11, EndLine: 1, EndColumn: 23}}},
		},
	)
}
