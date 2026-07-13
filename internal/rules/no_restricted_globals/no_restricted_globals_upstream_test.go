package no_restricted_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoRestrictedGlobalsUpstream migrates the full valid/invalid suite from
// upstream https://github.com/eslint/eslint/blob/main/tests/lib/rules/no-restricted-globals.js
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in the no_restricted_globals_extras_test.go file.
func TestNoRestrictedGlobalsUpstream(t *testing.T) {
	customMessage := "Use bar instead."

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- basic string / object option forms ----
			{Code: `foo`},
			{Code: `foo`, Options: []interface{}{"bar"}},
			{Code: `var foo = 1;`, Options: []interface{}{"foo"}},
			{Code: `event`, Options: []interface{}{"bar"}},
			{Code: `import foo from 'bar';`, Options: []interface{}{"foo"}},
			{Code: `function foo() {}`, Options: []interface{}{"foo"}},
			{Code: `function fn() { var foo; }`, Options: []interface{}{"foo"}},
			{Code: `foo.bar`, Options: []interface{}{"bar"}},
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"name": "bar", "message": "Use baz instead."}},
			},
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"bar"}}},
			},
			{
				Code:    `const foo = 1`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `event`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"bar"}}},
			},
			{
				Code:    `import foo from 'bar';`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `function foo() {}`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `function fn() { let foo; }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `foo.bar`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"bar"}}},
			},
			{
				Code: `foo`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{map[string]interface{}{"name": "bar", "message": "Use baz instead."}},
				}},
			},

			// ---- checkGlobalObject defaults to false: bare object access never checked ----
			{
				Code:    `window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code:    `globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
			},
			{
				Code: `myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals":       []interface{}{"foo"},
					"globalObjects": []interface{}{"myGlobal"},
				}},
			},

			// SKIP: these upstream cases are valid only because "window"/"self"/
			// "globalThis"/"myGlobal" are not recognized globals absent an ESLint
			// environment (languageOptions.globals) or a sufficient ecmaVersion.
			// rslint does not model ESLint's environment/global configuration, so
			// it always recognizes globalThis/self/window (and configured
			// globalObjects) as global-object roots when checkGlobalObject is
			// true — see the rule doc's "Differences from ESLint" section and the
			// corresponding invalid cases in the extras file.
			{
				Code:    `window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Skip:    true,
			},
			{
				Code:    `self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Skip:    true,
			},
			{
				Code:    `globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Skip:    true,
			},
			{
				Code: `myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals":           []interface{}{"foo"},
					"checkGlobalObject": true,
					"globalObjects":     []interface{}{"myGlobal"},
				}},
				Skip: true,
			},
			// "otherGlobal" is valid here for a structural reason (it's simply not
			// in globalObjects), not environment-gating — migrated normally.
			{
				Code: `otherGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals":           []interface{}{"foo"},
					"checkGlobalObject": true,
					"globalObjects":     []interface{}{"myGlobal"},
				}},
			},

			// ---- checkGlobalObject: the restricted name must be the final property, not an intermediate ----
			{
				Code: `foo.window.bar()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"bar"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `foo.self.bar()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"bar"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `foo.globalThis.bar()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"bar"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `foo.myGlobal.bar()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"bar"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
			},

			// ---- checkGlobalObject: a local shadowing declaration suppresses the report ----
			{
				Code: `let window; window.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `let self; self.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `let globalThis; globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true,
				}},
			},
			{
				Code: `let myGlobal; myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
			},

			// ---- TypeScript: type positions are never flagged ----
			{Code: `const foo: number = 1;`, Options: []interface{}{"foo"}},
			{Code: `function foo(): void {}`, Options: []interface{}{"foo"}},
			{Code: `function fn(): void { let foo; }`, Options: []interface{}{"foo"}},
			{
				Code: `
export default class Test {
	private status: string;
	getStatus() {
		return this.status;
	}
}`,
				Options: []interface{}{"status"},
			},
			{Code: `type Handler = (event: string) => any`, Options: []interface{}{"event"}},
			{Code: `let value: bigint`, Options: []interface{}{"bigint"}},
			{Code: `let value: boolean`, Options: []interface{}{"boolean"}},
			{Code: `let value: never`, Options: []interface{}{"never"}},
			{Code: `let value: null`, Options: []interface{}{"null"}},
			{Code: `let value: number`, Options: []interface{}{"number"}},
			{Code: `let value: object`, Options: []interface{}{"object"}},
			{Code: `let value: string`, Options: []interface{}{"string"}},
			{Code: `let value: symbol`, Options: []interface{}{"symbol"}},
			{Code: `let value: undefined`, Options: []interface{}{"undefined"}},
			{Code: `let value: unknown`, Options: []interface{}{"unknown"}},
			{Code: `let value: void`, Options: []interface{}{"void"}},
			{Code: `let value: []`, Options: []interface{}{"[]"}},
			{Code: `let value: {}`, Options: []interface{}{"{}"}},
			{Code: `let value: Test`, Options: []interface{}{"Test"}},
			{Code: `let value: Test[]`, Options: []interface{}{"Test"}},
			{Code: `let value: [Test]`, Options: []interface{}{"Test"}},
			{Code: `let b: { c: Test }`, Options: []interface{}{"Test"}},
			{Code: `function foo(param: Test) {}`, Options: []interface{}{"Test"}},
			{Code: `1 as Test`, Options: []interface{}{"Test"}},
			{Code: `class Derived implements Test {}`, Options: []interface{}{"Test"}},
			{Code: `class Derived implements Test1, Test2 {}`, Options: []interface{}{"Test1", "Test2"}},
			{Code: `interface Derived extends Test {}`, Options: []interface{}{"Test"}},
			{Code: `type Intersection = Test & {}`, Options: []interface{}{"Test"}},
			{Code: `type Union = Test | {}`, Options: []interface{}{"Test"}},
			{Code: `let value: NS.Test`, Options: []interface{}{"NS"}},
			{Code: `let value: NS.Test`, Options: []interface{}{"Test"}},
			{Code: `let value: NS.Test`, Options: []interface{}{"NS.Test"}},
			{Code: `let value: typeof Test`, Options: []interface{}{"Test"}},
			{Code: `let value: Type<Test>`, Options: []interface{}{"Type", "Test"}},
			{Code: `type Intersection = Test<any>`, Options: []interface{}{"Test", "any"}},
			{Code: `type Intersection = Test<A, B>`, Options: []interface{}{"Test", "A", "B"}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- basic string option ----
			{
				Code:    `foo`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			// SKIP: duplicate of the case above with languageOptions.globals set to
			// mark "foo" as an environment global; rslint doesn't model environment
			// globals and produces the identical diagnostic without it.
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{"foo"},
				Skip:    true,
			},
			{
				Code:    `event`,
				Options: []interface{}{"foo", "event"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{"foo"}, Skip: true}, // SKIP: duplicate, see above
			{
				Code:    `foo()`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{"foo"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- {name} object option ----
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"name": "foo"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{map[string]interface{}{"name": "foo"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			{Code: `function fn() { foo; }`, Options: []interface{}{map[string]interface{}{"name": "foo"}}, Skip: true}, // SKIP: duplicate (languageOptions.globals)
			{
				Code:    `event`,
				Options: []interface{}{"foo", map[string]interface{}{"name": "event"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{map[string]interface{}{"name": "foo"}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `foo()`,
				Options: []interface{}{map[string]interface{}{"name": "foo"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{map[string]interface{}{"name": "foo"}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- {name, message} custom message ----
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Message: "Unexpected use of 'foo'. Use bar instead.", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			{Code: `function fn() { foo; }`, Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `event`,
				Options: []interface{}{"foo", map[string]interface{}{"name": "event", "message": "Use local event parameter."}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Message: "Unexpected use of 'event'. Use local event parameter.", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `foo()`,
				Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- shadowed-by-default global name ----
			{
				Code:    `var foo = obj => hasOwnProperty(obj, 'name');`,
				Options: []interface{}{"hasOwnProperty"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 18, EndLine: 1, EndColumn: 32}},
			},

			// ---- {globals: [...]} object option (string entries) ----
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			{Code: `function fn() { foo; }`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `event`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo", "event"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- {globals: [{name}]} object option (object entries) ----
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			{Code: `function fn() { foo; }`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `event`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo", map[string]interface{}{"name": "event"}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo"}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},

			// ---- {globals: [{name, message}]} custom message ----
			{
				Code:    `foo`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `function fn() { foo; }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20}},
			},
			{Code: `function fn() { foo; }`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `event`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo", map[string]interface{}{"name": "event", "message": "Use local event parameter."}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 6}},
			},
			{Code: `foo`, Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}}, Skip: true}, // SKIP: duplicate
			{
				Code:    `foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `foo.bar()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{map[string]interface{}{"name": "foo", "message": customMessage}}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "customMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4}},
			},
			{
				Code:    `var foo = obj => hasOwnProperty(obj, 'name');`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"hasOwnProperty"}}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 18, EndLine: 1, EndColumn: 32}},
			},

			// ---- checkGlobalObject: dot access ----
			{
				Code:    `window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 11}},
			},
			{
				Code:    `self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 6, EndLine: 1, EndColumn: 9}},
			},
			{
				Code:    `window.window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 15, EndLine: 1, EndColumn: 18}},
			},
			{
				Code:    `self.self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 11, EndLine: 1, EndColumn: 14}},
			},
			{
				Code:    `globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 12, EndLine: 1, EndColumn: 15}},
			},
			{
				Code:    `globalThis.globalThis.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 23, EndLine: 1, EndColumn: 26}},
			},
			{
				Code: `myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 10, EndLine: 1, EndColumn: 13}},
			},
			{
				Code: `myGlobal.myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 19, EndLine: 1, EndColumn: 22}},
			},

			// ---- checkGlobalObject: bracket access ----
			{
				Code:    `window["foo"]`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 13}},
			},
			{
				Code:    `self["foo"]`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 6, EndLine: 1, EndColumn: 11}},
			},
			{
				Code:    `globalThis["foo"]`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 12, EndLine: 1, EndColumn: 17}},
			},
			{
				Code: `myGlobal["foo"]`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 10, EndLine: 1, EndColumn: 15}},
			},

			// ---- checkGlobalObject: optional chaining ----
			{
				Code:    `window?.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 9, EndLine: 1, EndColumn: 12}},
			},
			{
				Code:    `self?.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 7, EndLine: 1, EndColumn: 10}},
			},

			// ---- checkGlobalObject: multiple diagnostics per file ----
			{
				Code: `window.foo(); myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "defaultMessage", Line: 1, Column: 8, EndLine: 1, EndColumn: 11},
					{MessageId: "defaultMessage", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: `myGlobal.foo(); myOtherGlobal.bar()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo", "bar"}, "checkGlobalObject": true,
					"globalObjects": []interface{}{"myGlobal", "myOtherGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "defaultMessage", Line: 1, Column: 10, EndLine: 1, EndColumn: 13},
					{MessageId: "defaultMessage", Line: 1, Column: 31, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    `foo(); window.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
					{MessageId: "defaultMessage", Line: 1, Column: 15, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `foo(); self.foo()`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"foo"}, "checkGlobalObject": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
					{MessageId: "defaultMessage", Line: 1, Column: 13, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: `foo(); myGlobal.foo()`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"foo"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "defaultMessage", Line: 1, Column: 1, EndLine: 1, EndColumn: 4},
					{MessageId: "defaultMessage", Line: 1, Column: 17, EndLine: 1, EndColumn: 20},
				},
			},

			// ---- checkGlobalObject: local `event` shadows the parameter, only the global-object access is reported ----
			{
				Code:    `function onClick(event) { console.log(event); console.log(window.event); }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"event"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 66, EndLine: 1, EndColumn: 71}},
			},
			{
				Code:    `function onClick(event) { console.log(event); console.log(self.event); }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"event"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 64, EndLine: 1, EndColumn: 69}},
			},
			{
				Code:    `function onClick(event) { console.log(event); console.log(globalThis.event); }`,
				Options: []interface{}{map[string]interface{}{"globals": []interface{}{"event"}, "checkGlobalObject": true}},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 70, EndLine: 1, EndColumn: 75}},
			},
			{
				Code: `function onClick(event) { console.log(event); console.log(myGlobal.event); }`,
				Options: []interface{}{map[string]interface{}{
					"globals": []interface{}{"event"}, "checkGlobalObject": true, "globalObjects": []interface{}{"myGlobal"},
				}},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 68, EndLine: 1, EndColumn: 73}},
			},

			// ---- TypeScript: value reference reported, type annotation is not ----
			{
				Code:    `const x: Promise<any> = Promise.resolve();`,
				Options: []interface{}{"Promise"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "defaultMessage", Line: 1, Column: 25, EndLine: 1, EndColumn: 32}},
			},
		},
	)
}
