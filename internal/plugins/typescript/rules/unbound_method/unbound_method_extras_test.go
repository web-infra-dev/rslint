package unbound_method

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestUnboundMethodComputedMembers(t *testing.T) {
	const declaration = "class Service { method() {} safe(this: void) {} bound = () => {}; static method() {} }\nconst service = new Service();\n"
	var invalid []rule_tester.InvalidTestCase
	for _, expression := range []string{
		`service['method']`, "service[`method`]", `service[('method')]`,
		`service?.['method']`, `Service['method']`,
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code: declaration + expression + ";",
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "unboundWithoutThisAnnotation",
				Message:   "Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.",
				Line:      3, Column: 1, EndLine: 3, EndColumn: len(expression) + 1,
			}},
		})
	}
	for _, code := range []string{
		"const service = { method() {} }; const key = 'method';\nservice[key];",
		"const service = { method() {}, other() {} }; declare const key: 'method' | 'other';\nservice[key];",
		"const service = { method() {} };\nservice[`${'method'}`];",
		"const service = { 1() {} };\nservice[1];",
		"const service = { method() {} }; let method;\n({ method } = service);",
		"class A { method() {} } class B { method = 1; } declare const value: A | B;\nvalue.method;",
		"class A { method() {} } class B { method = 1; } declare const value: A | B;\nvalue['method'];",
		"class A { method() {} } class B { method: () => void; } declare const value: A & B;\nvalue.method;",
		"class A { method() {} } class B { method: () => void; } declare const value: A & B;\nvalue['method'];",
		"declare const value: Window;\nvalue.blur;",
		"declare const value: Window;\nvalue['blur'];",
	} {
		invalid = append(invalid, rule_tester.InvalidTestCase{
			Code:   "export {};\n" + code,
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3}},
		})
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnboundMethodRule,
		[]rule_tester.ValidTestCase{
			{Code: declaration + `service['method']();`},
			{Code: declaration + `service?.['method']?.();`},
			{Code: declaration + `if (service['method']) {}`},
			{Code: declaration + `service['safe']; service['bound'];`},
			{Code: declaration + `service['method'].bind(service);`},
			{Code: declaration + `Service['method'];`, Options: map[string]any{"ignoreStatic": true}},
			{Code: declaration + `declare const key: string; service[key];`},
			{Code: `const service = { method() {} }; declare const key: 'method' & { readonly brand: unique symbol }; service[key];`},
			{Code: `const method = Math.floor; const { parseInt } = Number;`},
			{Code: `const service = { method() {} }; let method; ({ ['method']: method } = service);`},
			{Code: `const service = { method() {} }; let method; ({ 'method': method } = service);`},
		}, invalid)
}

func TestUnboundMethodExemptionKeepsDestructuring(t *testing.T) {
	adapted := rule.CreateRule(rule.Rule{
		Name: "unbound-method-exemption", Schema: UnboundMethodRule.Schema, RequiresTypeInfo: true,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return CreateListeners(ctx, options, func(*ast.Node) bool { return true })
		},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &adapted,
		[]rule_tester.ValidTestCase{{Code: `const service = { method() {} }; service.method; service['method'];`}},
		[]rule_tester.InvalidTestCase{{
			Code:   "const service = { method() {} };\nconst { method } = service;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 2, Column: 9, EndLine: 2, EndColumn: 15}},
		}})
}

func TestUnboundMethodIgnoreStaticDoesNotIgnoreFunctionFields(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnboundMethodRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{{
			Code:    `class Service { static method = function () {} } Service.method;`,
			Options: map[string]any{"ignoreStatic": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "unbound"}},
		}})
}
