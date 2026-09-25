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
			{Code: "globalThis.foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "window.foo += 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "self.foo ||= value", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "global.foo++", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "globalThis[\"foo\"] = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "({\n\tfoo: globalThis.foo,\n} = object);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "[globalThis.foo] = array", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "({...globalThis.foo} = object)", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "[...globalThis.foo] = array", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "for (globalThis.foo of iterable) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "for (globalThis.foo in object) {}", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}},
			{Code: "globalThis!.foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}, FileName: "file.ts"},
			{Code: "(globalThis as any).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}, FileName: "file.ts"},
			{Code: "(<any>globalThis).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}, FileName: "file.ts"},
			{Code: "(globalThis satisfies any).foo = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}, FileName: "file.ts"},
			{Code: "globalThis.foo! = 1", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "no-global-object-property-assignment"}}, FileName: "file.ts"},
		},
	)
}
