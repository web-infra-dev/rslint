package no_obj_calls

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoObjCallsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoObjCallsRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `var x = Math.random();`},
			{Code: `var x = JSON.parse(foo);`},
			{Code: `Reflect.get(foo, 'x');`},
			{Code: `new Intl.Segmenter();`},
			{Code: `var x = Math;`},
			{Code: `var x = Math.PI;`},
			{Code: `var x = foo.Math();`},
			{Code: `var x = new foo.Math();`},
			{Code: `JSON.parse(foo)`},
			{Code: `new JSON.parse`},
			// globalThis property access (not calling the global itself)
			{Code: `var x = new globalThis.Math.foo;`},
			{Code: `new globalThis.Object()`},
			// Shadowed variable — should not be flagged
			{Code: `function f() { var Math = 1; Math(); }`},
			{Code: `function f(JSON: any) { JSON(); }`},
			{Code: `function f() { var globalThis = { Math: () => {} }; globalThis.Math(); }`},
			// A write to the global binding makes ReferenceTracker skip that
			// global entirely. Writing one of its properties does not.
			{Code: `Math(); Math = replacement;`},
			{Code: `Math += replacement; Math();`},
			{Code: `Math++; Math();`},
			{Code: `for (Math of values) {} Math();`},
			{Code: `({ Math } = value); Math();`},
			// Dynamic global-object properties and array patterns are not
			// paths represented by this rule's trace map.
			{Code: `globalThis[name]();`},
			{Code: `let value; [value] = [JSON]; value();`},
			{Code: `const { ...value } = globalThis; value();`},
			{Code: `const root = globalThis; root();`},
			{Code: `let value = (JSON, fallback); value();`},
			// Cycles which have no tracked source must terminate silently.
			{Code: `let a = b; let b = a; a();`},
			// A closer lexical binding must not inherit an outer alias.
			{Code: `let value; { let value; value = JSON; } value();`},
			{Code: `const { JSON: value } = { JSON: () => {} }; value();`},
			{Code: `function f(JSON = JSON) { JSON(); }`},
			{Code: `namespace Math {} Math();`},
			{Code: `enum JSON {} JSON();`},
			// Explicitly disabling a global removes the direct reference, but
			// does not remove the same-named property from globalThis.
			{Code: `Temporal();`, Globals: map[string]bool{"Temporal": false}},
			{Code: `const value = Temporal; value();`, Globals: map[string]bool{"Temporal": false}},
			// Property values are not lexical aliases followed by ESLint's
			// ReferenceTracker.
			{Code: `const obj = { foo: JSON }; obj.foo();`},
			{Code: `namespace ns { export const foo = JSON; } ns.foo();`},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `var x = JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `new Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			// globalThis access
			{
				Code: `var x = globalThis.Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = new globalThis.Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = globalThis.JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = globalThis.Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `globalThis.Atomics();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.Intl();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			// globalThis with optional chaining
			{
				Code: `var x = globalThis?.Reflect();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			{
				Code: `var x = (globalThis?.Reflect)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 9},
				},
			},
			// multiple errors in one expression
			{
				Code: `Math( JSON() );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
					{MessageId: "unexpectedCall", Line: 1, Column: 7},
				},
			},
			{
				Code: `globalThis.Math( globalThis.JSON() );`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
					{MessageId: "unexpectedCall", Line: 1, Column: 18},
				},
			},
			// indirect references via variable assignment
			{
				Code: `var foo = JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 17},
				},
			},
			{
				Code: `var foo = Math; new foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 17},
				},
			},
			{
				Code: `var foo = bar ? baz : JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 29},
				},
			},
			{
				Code: `var foo = globalThis.JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 28},
				},
			},
			// indirect via logical operators
			{
				Code: `var foo = undefined || JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 30},
				},
			},
			{
				Code: `var foo = undefined ?? Reflect; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 33},
				},
			},
			// TS type assertions as pass-through in initializer
			{
				Code: `var foo = JSON as any; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 24},
				},
			},
			{
				Code: `var foo = JSON satisfies any; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 31},
				},
			},
			{
				Code: `var foo = <any>JSON; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 22},
				},
			},
			{
				Code: `var foo = JSON!; foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 18},
				},
			},
			// comma operator
			{
				Code: `var foo = (0, JSON); foo();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 22},
				},
			},
			// multi-hop indirect references
			{
				Code: `var a = JSON; var b = a; b();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 26},
				},
			},
			// direct call with TS assertion as callee
			{
				Code: `(JSON as any)();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 1},
				},
			},
			// Static computed global-object access.
			{
				Code: `globalThis["Math"]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: "globalThis[`JSON`]();",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis["Ma" + "th"]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'null' is reference to 'Math', which is not a function.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `globalThis[true ? "Math" : "JSON"]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'null' is reference to 'Math', which is not a function.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `globalThis["Math" as const]();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'null' is reference to 'Math', which is not a function.",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Temporal is non-callable in the current ESLint rule.
			{
				Code: `Temporal();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			{
				Code:    `globalThis.Temporal();`,
				Globals: map[string]bool{"Temporal": false},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall", Line: 1, Column: 1},
				},
			},
			// Assignment aliases, including compound assignment and object
			// destructuring assignment.
			{
				Code: `let value; value = JSON; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'value' is reference to 'JSON', which is not a function.",
						Line:      1,
						Column:    26,
					},
				},
			},
			{
				Code: `let value; value += JSON; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 27},
				},
			},
			{
				Code: `let value; ({ Math: value } = globalThis); new value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'value' is reference to 'Math', which is not a function.",
						Line:      1,
						Column:    44,
					},
				},
			},
			{
				Code: `const { ["JS" + "ON"]: value } = globalThis; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'value' is reference to 'JSON', which is not a function.",
					},
				},
			},
			{
				Code: `let value; ({ ["Ma" + "th"]: value } = globalThis); value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'value' is reference to 'Math', which is not a function.",
					},
				},
			},
			// Alias a global object, then destructure a tracked property from
			// that alias.
			{
				Code: `const root = globalThis; const { Reflect: value } = root; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'value' is reference to 'Reflect', which is not a function.",
						Line:      1,
						Column:    59,
					},
				},
			},
			// AssignmentPattern/default-parameter propagation.
			{
				Code: `function f(value = JSON) { value(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 28},
				},
			},
			// A body-level var does not shadow a global used by a parameter
			// default.
			{
				Code: `function f(value = JSON) { var JSON; value(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall"},
				},
			},
			{
				Code: `function f({ value = JSON }) { value(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall"},
				},
			},
			{
				Code: `let value; ({ other: value = JSON } = object); value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall"},
				},
			},
			// ReferenceTracker is intentionally flow-insensitive: later writes
			// do not erase an alias established by an initializer.
			{
				Code: `let value = JSON; value = () => {}; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 37},
				},
			},
			// References are resolved by binder identity, including use before
			// declaration and closer shadowing.
			{
				Code: `value(); var value = JSON;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 1},
				},
			},
			{
				Code: `const value = JSON; function f() { const value = () => {}; value(); } value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 71},
				},
			},
			// A cycle reachable from a tracked source must terminate while
			// preserving the reachable call.
			{
				Code: `let a = JSON; let b = a; a = b; a();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 33},
				},
			},
			{
				Code: `let a, b; a = b = JSON; a(); b();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall"},
					{MessageId: "unexpectedRefCall"},
				},
			},
			// Multiple tracked writes reach the same call independently, which
			// matches ReferenceTracker's generator semantics.
			{
				Code: `let value; value = JSON; value = JSON; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall", Line: 1, Column: 40},
					{MessageId: "unexpectedRefCall", Line: 1, Column: 40},
				},
			},
			// A modified watched global can still be an alias reached from a
			// different, unmodified watched global.
			{
				Code: `Math = JSON; Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpectedRefCall",
						Message:   "'Math' is reference to 'JSON', which is not a function.",
						Line:      1,
						Column:    14,
					},
				},
			},
			// Type-only declarations do not shadow the corresponding runtime
			// global, while value-space namespace/enum declarations do.
			{
				Code: `interface Math {} Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall"},
				},
			},
			{
				Code: `type JSON = {}; JSON();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall"},
				},
			},
			// Global-object aliases retain the global object's trace-map shape.
			{
				Code: `const root = globalThis; root.Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall"},
				},
			},
			{
				Code: `const value = JSON<string>; value();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedRefCall"},
				},
			},
			{
				Code: `Math.member = replacement; Math();`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedCall"},
				},
			},
		},
	)
}

func TestSourceMayUseNonCallableGlobal(t *testing.T) {
	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "ordinary method call", code: `service.method()`, want: false},
		{name: "global identifier", code: `JSON.parse("{}")`, want: true},
		{name: "escaped global identifier", code: `M\u0061th()`, want: true},
		{name: "computed global object access", code: `globalThis["Math"]()`, want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
				FileName: "/source.ts",
				Path:     "/source.ts",
			}, testCase.code, core.ScriptKindTS)
			if got := sourceMayUseNonCallableGlobal(sourceFile); got != testCase.want {
				t.Fatalf("sourceMayUseNonCallableGlobal() = %v, want %v", got, testCase.want)
			}
		})
	}
}
