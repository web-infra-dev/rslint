// TestNoUnusedVarsNestedScopesAndBindings covers ownership boundaries that are
// easy to misclassify with ancestor-only checks: nested declarations inside
// parameter/catch expressions, static-block variable scopes, object-rest
// siblings, recursive initializers, inline-global shadows, and JSX tag names.
package no_unused_vars

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnusedVarsNestedScopesAndBindings(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnusedVarsRule,
		[]rule_tester.ValidTestCase{
			// Only direct properties beside an object rest are ignored.
			{Code: `const { value: direct, ...rest } = source; console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},
			{Code: `let direct; let rest; ({ value: direct, ...rest } = source); console.log(rest);`, Options: map[string]interface{}{"ignoreRestSiblings": true}},

			// A static block is its own variable scope, so an outer value read
			// there can be observed independently of the outer declaration.
			{Code: `let x = 0; class C { static { x = x + 1; } } new C();`},
			{Code: `let x = 0; namespace N { x = x + 1; } consume(N);`},

			// A callback passed through the RHS can execute later and stores its read.
			{Code: `let x; x = consume(() => x);`},

			// Local shadows do not consume an inline global, but a real global read does.
			{Code: `/*global foo*/ function f(foo) { return foo; } consume(foo); f(1);`},
			{Code: `/*global foo*/ consume(foo);`},
			{Code: `/*global Foo*/ type Alias = Foo; consume({} as Alias);`},

			// Every binding introduced by an exported destructuring declaration is exported.
			{Code: `export const { nested: { value }, list: [item] } = source;`},

			// Core no-unused-vars follows the TypeScript scope manager and accepts
			// a value/type declaration consumed from a type position.
			{Code: `class Foo {} let value: Foo; consume(value);`},

			// Capitalized JSX tags are component references.
			{Code: `const Component = () => null; const view = <Component />; consume(view);`, Tsx: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `function f(a = (() => { const nested = 1; return 0; })()) {} f();`,
				Options: map[string]interface{}{"args": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 31, 37, `function f(a = (() => {  return 0; })()) {} f();`),
				},
			},
			{
				Code:    `try {} catch (error) { const nested = 1; console.log(error); }`,
				Options: map[string]interface{}{"caughtErrors": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 30, 36, `try {} catch (error) {  console.log(error); }`),
				},
			},
			{
				Code:    `const [head, ...tail] = source; console.log(tail);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("head", true, 1, 8, 12, `const [, ...tail] = source; console.log(tail);`),
				},
			},
			{
				Code:    `const { value: [nested], ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 17, 23, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `const { value: { nested }, ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 18, 24, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `const { value = 1, ...rest } = source; console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("value", true, 1, 9, 14, `const {  ...rest } = source; console.log(rest);`),
				},
			},
			{
				Code:    `let value; let rest; ({ value = 1, ...rest } = source); console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("value", true, 1, 25, 30, ""),
				},
			},
			{
				Code:    `let nested; let rest; ({ value: [nested], ...rest } = source); console.log(rest);`,
				Options: map[string]interface{}{"ignoreRestSiblings": true},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("nested", true, 1, 34, 40, ""),
				},
			},
			extraUnusedCase(`let x; class C { static { x = 1; } } new C();`, "x", true, 1, 5, 6, ""),
			extraUnusedCase(`let x; namespace N { x = 1; } consume(N);`, "x", true, 1, 5, 6, ""),
			{
				Code: `const f = (function () { return f(); });`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", true, 1, 7, 8, ""),
				},
			},
			{
				Code: `const f = ((() => f()));`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", true, 1, 7, 8, ""),
				},
			},
			{
				Code:    `function f(a = f) {}`,
				Options: map[string]interface{}{"args": "none"},
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedErrorWithSuggestion("f", false, 1, 10, 11, ""),
				},
			},
			{
				Code: `/*global foo*/ function f(foo) { return foo; } f(1);`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo*/ { const foo = 1; consume(foo); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo += 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo = foo + 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `/*global foo:writable*/ foo++;`,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("foo", false, 1, 10, 13, ""),
				},
			},
			{
				Code: `const div = 1; const view = <div />; consume(view);`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("div", true, 1, 7, 10, ` const view = <div />; consume(view);`),
				},
			},
			{
				Code: "import React from \"react\";\nconst view = <div />;\nconsume(view);",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					extraUnusedError("React", false, 1, 8, 13, "import \"react\";\nconst view = <div />;\nconsume(view);"),
				},
			},
		},
	)
}
