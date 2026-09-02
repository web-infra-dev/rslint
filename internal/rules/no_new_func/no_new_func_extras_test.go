package no_new_func

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoNewFuncRuleExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoNewFuncRule,
		[]rule_tester.ValidTestCase{
			// Value references and non-matching member shapes are not calls to the constructor.
			{Code: `var x = [Function]`},
			{Code: `var x = Function`},
			{Code: `typeof Function`},
			{Code: `Function.hasOwnProperty("call")`},
			{Code: `Function.prototype`},
			{Code: `var x = Function.call`},
			{Code: `var x = Function.apply`},
			{Code: "Function`code`"},
			{Code: `Reflect.construct(Function, ["code"])`},
			{Code: `globalThis.Function("code")`},
			{Code: `new Function.call(null, "code")`},
			{Code: `Function.call.call(null, null, "code")`},
			{Code: `Function["ca" + "ll"](null, "code")`},

			// ESTree preserves TypeScript assertions as distinct expressions.
			{Code: `(Function as any)("code")`},
			{Code: `(<any>Function)("code")`},
			{Code: `Function!("code")`},
			{Code: `(Function satisfies any)("code")`},
			{Code: `new (Function as any)("code")`},
			{Code: `(Function as any).call(null, "code")`},
			{Code: `Function!.call(null, "code")`},
			{Code: `(Function.call as any)(null, "code")`},
			{Code: `Function.call!(null, "code")`},
			{Code: `(Function.call satisfies any)(null, "code")`},
			{Code: `Function["call" as const](null, "code")`},
			{Code: `Function[("call" as const)](null, "code")`},

			// Type-only imports do define a scope-manager variable; value namespaces also shadow.
			{Code: "import type { Function } from \"source\";\nFunction(\"code\")"},
			{Code: "namespace Function {}\nFunction(\"code\")"},
			{Code: "declare namespace Function {}\nFunction(\"code\")"},
			{Code: "enum Function {}\nFunction(\"code\")"},

			// Local value bindings shadow the global through every relevant scope.
			{Code: `function test() { var x = new Function("code"); var Function = function() {}; }`},
			{Code: `function test() { if (true) { var Function = 42; } new Function(); }`},
			{Code: `function test() { for (var Function = 0; Function < 1; Function++) {} new Function(); }`},
			{Code: `function test() { for (var Function in {}) {} new Function(); }`},
			{Code: `function test() { for (var Function of []) {} new Function(); }`},
			{Code: `function test() { switch (0) { case 0: var Function = 1; } new Function(); }`},
			{Code: `function test() { let Function = class {}; return new Function(); }`},
			{Code: `function test() { const Function = class {}; return Function(); }`},
			{Code: `function test(Function) { return new Function(); }`},
			{Code: `function test({ Function }) { return new Function(); }`},
			{Code: `function test([Function]) { return new Function(); }`},
			{Code: `function test(...Function) { return new Function(); }`},
			{Code: `function test(Function = class {}) { return new Function(); }`},
			{Code: `var fn = (Function) => Function();`},
			{Code: `function* gen(Function) { yield new Function(); }`},
			{Code: `async function af(Function) { return new Function(); }`},
			{Code: `try {} catch (Function) { new Function(); }`},
			{Code: `function test() { var Function = class {}; function inner() { return new Function(); } }`},
			{Code: `function test() { var Function = class {}; var fn = () => new Function(); }`},
			{Code: `var obj = { m(Function) { return new Function(); } };`},
			{Code: `class C { m(Function) { return new Function(); } }`},
			{Code: `class C { constructor(Function) { this.x = new Function(); } }`},
			{Code: `function test() { for (let Function in {}) { new Function(); } }`},
			{Code: `function test() { for (let Function of []) { new Function(); } }`},
			{Code: `function test(Function) { return Function.call(null, "code"); }`},
			{Code: `function test() { var Function = class {}; Function.apply(null, ["code"]); }`},
			{Code: `function test(Function = Function("code")) {}`},

			// A class declaration is bound in its outer scope; member decorators see the class scope.
			{Code: `@dec(Function("code")) class Function {}`},
			{Code: `class Function { @dec(Function("code")) method() {} }`},
			{Code: `class Function { method(@dec(Function("code")) value: any) {} }`},

			// A top-level definition changes the global variable's defs in non-module source goals.
			{Code: "interface Function {}\nnew Function(\"code\")", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "type Function = {}\nFunction(\"code\")", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "interface Function {}\nfunction nested() { Function(\"code\") }", LanguageOptions: rule.LanguageOptions{SourceType: "script"}},
			{Code: "interface Function {}\nnew Function(\"code\")", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},

			// Removing Function from the effective globals removes the rule's target variable.
			{Code: `new Function("code");`, Globals: map[string]any{"Function": "off"}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `new Function()`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noFunctionConstructor",
					Message:   "The Function constructor is eval.",
					Line:      1,
					Column:    1,
					EndLine:   1,
					EndColumn: 15,
				}},
			},
			{
				Code: "Function(\r\n  \"😀\"\r\n)",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "noFunctionConstructor",
					Line:      1,
					Column:    1,
					EndLine:   3,
					EndColumn: 2,
				}},
			},
			{
				Code:   `(Function)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `((Function))("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `new (Function)("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function<string>("code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `var a = Function["apply"](null, ["b", "c", "return b+c"]);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `var a = Function["bind"](null, "b", "c", "return b+c");`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   "var a = Function[`call`](null, \"code\")",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 9}},
			},
			{
				Code:   `Function?.call(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.apply(null, ["code"])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.bind(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function.call?.(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function?.["call"](null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function["call"]?.(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `Function[("call")](null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `(Function).call(null, "code")`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `(Function).apply(null, ["code"])`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:   `new (Function("code"))`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 6}},
			},
			{
				Code: `var a = new Function("a"); var b = Function("b");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 9},
					{MessageId: "noFunctionConstructor", Line: 1, Column: 36},
				},
			},
			{
				Code: `var x = true ? new Function("a") : Function("b");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noFunctionConstructor", Line: 1, Column: 16},
					{MessageId: "noFunctionConstructor", Line: 1, Column: 36},
				},
			},
			{
				Code:   `function f() { { let Function = class {}; } var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 53}},
			},
			{
				Code:   `function f() { var x = new Function("code"); (function() { var Function = 1; })(); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 24}},
			},
			{
				Code:   `function f() { var fn = (Function) => Function; var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 57}},
			},
			{
				Code:   `function f() { try {} catch (Function) {} var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 51}},
			},
			{
				Code:   `function f() { for (let Function of []) {} var x = new Function("code"); }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 52}},
			},
			{
				Code:    `new Function("code");`,
				Globals: map[string]any{"Function": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},
			{
				Code:     `new Function("code");`,
				TSConfig: "tsconfig.noLib.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 1}},
			},

			// Type-only declarations in module scope do not define the global variable.
			{
				Code:            "interface Function {}\nnew Function(\"code\")",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 2, Column: 1}},
			},
			{
				Code:            "type Function = {}\nFunction(\"code\")",
				LanguageOptions: rule.LanguageOptions{SourceType: "module"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 2, Column: 1}},
			},

			// Class names and parameters are not visible while their decorators evaluate.
			{
				Code:   `const C = @dec(Function("code")) class Function {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 16}},
			},
			{
				Code:   `class C { m(@dec(Function("code")) Function: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 1, Column: 18}},
			},
			{
				Code:   `class C { m(Function: any, @dec(Function("code")) value: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `class C { m<Function>(@dec(Function("code")) value: any) {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:   `function test(value = Function("code")) { var Function = 1; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor"}},
			},
			{
				Code:            "function test() { interface Function {}\nFunction(\"code\") }",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 2, Column: 1}},
			},
			{
				Code:            "if (true) { interface Function {} }\nFunction(\"code\")",
				LanguageOptions: rule.LanguageOptions{SourceType: "script"},
				Errors:          []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 2, Column: 1}},
			},

			// Ambient global augmentations extend, rather than shadow, the builtin.
			{
				Code:   "export {};\ndeclare global {\n\tnamespace Function {}\n}\nnew Function(\"code\");",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 5, Column: 1}},
			},
			{
				Code:   "export {};\ndeclare global {\n\tvar Function: any;\n}\nnew Function(\"code\");",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "noFunctionConstructor", Line: 5, Column: 1}},
			},
		},
	)
}
