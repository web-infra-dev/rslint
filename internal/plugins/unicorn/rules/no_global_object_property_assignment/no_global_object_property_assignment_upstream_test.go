package no_global_object_property_assignment_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_global_object_property_assignment"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoGlobalObjectPropertyAssignmentUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_global_object_property_assignment.NoGlobalObjectPropertyAssignmentRule,
		[]rule_tester.ValidTestCase{
			{Code: "globalThis.foo"},
			{Code: "window.foo()"},
			{Code: "globalThis = value"},
			{Code: "globalThis[property] = value"},
			{Code: "delete globalThis.foo"},
			{Code: "Object.assign(globalThis, {foo: value})"},
			{Code: "Reflect.set(globalThis, \"foo\", value)"},
			{Code: "function test(window) {\n\twindow.foo = 1;\n}"},
			{Code: "const global = {};\nglobal.foo = 1;"},
			{Code: "const globalThis = {};\nglobalThis.foo = 1;"},
			{Code: "const self = {};\nself.foo++;"},
			{Code: "const root = globalThis;\nroot.foo = 1;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "globalThis.foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}}},
			{Code: "window.foo += 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}}},
			{Code: "self.foo ||= value", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 9}}},
			{Code: "global.foo++", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 11}}},
			{Code: "globalThis[\"foo\"] = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			{Code: "({\n\tfoo: globalThis.foo,\n} = object);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 2, Column: 7, EndLine: 2, EndColumn: 21}}},
			{Code: "[globalThis.foo] = array", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 2, EndLine: 1, EndColumn: 16}}},
			{Code: "({...globalThis.foo} = object)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 6, EndLine: 1, EndColumn: 20}}},
			{Code: "[...globalThis.foo] = array", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 5, EndLine: 1, EndColumn: 19}}},
			{Code: "for (globalThis.foo of iterable) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 6, EndLine: 1, EndColumn: 20}}},
			{Code: "for (globalThis.foo in object) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 6, EndLine: 1, EndColumn: 20}}},
			{Code: "globalThis!.foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16}}, FileName: "file.ts"},
			{Code: "(globalThis as any).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 24}}, FileName: "file.ts"},
			{Code: "(<any>globalThis).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 22}}, FileName: "file.ts"},
			{Code: "(globalThis satisfies any).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 31}}, FileName: "file.ts"},
			{Code: "globalThis.foo! = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment", Message: "Do not assign properties on the global object.", Line: 1, Column: 1, EndLine: 1, EndColumn: 15}}, FileName: "file.ts"},
		},
	)
}
