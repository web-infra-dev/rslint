package no_restricted_exports

// TestNoRestrictedExportsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-restricted-exports.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases
// live in no_restricted_exports_extras_test.go.

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func namedMsg(name string) string {
	return fmt.Sprintf("'%s' is restricted from being used as an exported name.", name)
}

const defaultMsg = "Exporting 'default' is restricted."

func TestNoRestrictedExportsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRestrictedExportsRule,
		[]rule_tester.ValidTestCase{
			// ---- nothing configured ----
			{Code: `export var a;`},
			{Code: `export function a() {}`},
			{Code: `export class A {}`},
			{Code: `var a; export { a };`},
			{Code: `var b; export { b as a };`},
			{Code: `export { a } from 'foo';`},
			{Code: `export { b as a } from 'foo';`},
			{Code: `export var a;`, Options: []any{map[string]any{}}},
			{Code: `export function a() {}`, Options: []any{map[string]any{}}},
			{Code: `export class A {}`, Options: []any{map[string]any{}}},
			{Code: `var a; export { a };`, Options: []any{map[string]any{}}},
			{Code: `var b; export { b as a };`, Options: []any{map[string]any{}}},
			{Code: `export { a } from 'foo';`, Options: []any{map[string]any{}}},
			{Code: `export { b as a } from 'foo';`, Options: []any{map[string]any{}}},
			{Code: `export var a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `export function a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `export class A {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `var a; export { a };`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `var b; export { b as a };`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `export { a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},
			{Code: `export { b as a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{}}}},

			// ---- not a restricted name ----
			{Code: `export var a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export let a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export const a = 1;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export function a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export function *a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export async function a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export async function *a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export class A {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `var a; export { a };`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `var b; export { b as a };`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export { a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export { b as a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"x"}}}},
			{Code: `export { '' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"undefined"}}}},
			{Code: `export { '' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{" "}}}},
			{Code: `export { ' ' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{""}}}},
			{Code: `export { ' a', 'a ' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- does not mistakenly disallow non-exported names that appear in named export declarations ----
			{Code: `export var b = a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export let [b = a] = [];`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export const [b] = [a];`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export var { a: b } = {};`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export let { b = a } = {};`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export const { c: b = a } = {};`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export function b(a) {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export class A { a(){} }`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export class A extends B {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"B"}}}},
			{Code: `var a; export { a as b };`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `var a; export { a as 'a ' };`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export { a as b } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export { a as 'a ' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export { 'a' as 'a ' } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- does not check source in re-export declarations ----
			{Code: `export { b } from 'a';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export * as b from 'a';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- does not check non-export declarations ----
			{Code: `var a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `let a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `const a = 1;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `function a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `class A {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"A"}}}},
			{Code: `import a from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `import { a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `import { b as a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `var setSomething; export { setSomething };`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "^get"}}},
			{Code: `var foo, bar; export { foo, bar };`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "^(?!foo)(?!bar).+$"}}},
			{Code: `var foobar; export default foobar;`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "bar$"}}},
			{Code: `var foobar; export default foobar;`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `export default 'default';`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `var foobar; export { foobar as default };`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `var foobar; export { foobar as 'default' };`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `export { default } from 'mod';`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `export { default as default } from 'mod';`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `export { foobar as default } from 'mod';`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},
			{Code: `export * as default from 'mod';`, Options: []any{map[string]any{"restrictedNamedExportsPattern": "default"}}},

			// ---- does not check re-export all declarations ----
			{Code: `export * from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export * from 'a';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},

			// ---- does not mistakenly disallow identifiers in export default declarations (a default export will export "default" name) ----
			{Code: `export default a;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export default function a() {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export default class A {}`, Options: []any{map[string]any{"restrictedNamedExports": []any{"A"}}}},
			{Code: `export default (function a() {});`, Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}}},
			{Code: `export default (class A {});`, Options: []any{map[string]any{"restrictedNamedExports": []any{"A"}}}},

			// ---- by design, restricted name "default" does not apply to default export declarations, although they do export the "default" name. ----
			{Code: `export default 1;`, Options: []any{map[string]any{"restrictedNamedExports": []any{"default"}}}},

			// ---- "default" does not disallow re-exporting a renamed default export from another module ----
			{Code: `export { default as a } from 'foo';`, Options: []any{map[string]any{"restrictedNamedExports": []any{"default"}}}},

			// ---- restrictDefaultExports.direct option ----
			{Code: `export default foo;`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": false}}}},
			{Code: `export default 42;`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": false}}}},
			{Code: `export default function foo() {}`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": false}}}},

			// ---- restrictDefaultExports.named option ----
			{Code: "const foo = 123;\nexport { foo as default };", Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"named": false}}}},

			// ---- restrictDefaultExports.defaultFrom option ----
			{Code: `export { default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": false}}}},
			{Code: `export { default as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": false}}}},
			{Code: `export { foo as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}}},
			{Code: `export { default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"named": true, "defaultFrom": false}}}},
			{Code: `export { 'default' } from 'mod'; `, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": false}}}},

			// ---- restrictDefaultExports.namedFrom option ----
			{Code: `export { foo as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": false}}}},
			{Code: `export { default as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": true}}}},
			{Code: `export { default as default } from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": false}}}},
			{Code: `export { 'default' } from 'mod'; `, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": false, "namedFrom": true}}}},

			// ---- restrictDefaultExports.namespaceFrom option ----
			{Code: `export * as default from 'mod';`, Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namespaceFrom": false}}}},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:    `export function someFunction() {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"someFunction"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("someFunction"), Line: 1, Column: 17},
				},
			},

			// ---- basic tests ----
			{
				Code:    `export var a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export var a = 1;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export let a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export let a = 1;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export const a = 1;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `export function a() {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export function *a() {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 18},
				},
			},
			{
				Code:    `export async function a() {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 23},
				},
			},
			{
				Code:    `export async function *a() {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 24},
				},
			},
			{
				Code:    `export class A {}`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"A"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("A"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `let a; export { a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export { a }; var a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 10},
				},
			},
			{
				Code:    `let b; export { b as a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `export { a } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 10},
				},
			},
			{
				Code:    `export { b as a } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},

			// ---- string literals ----
			{
				Code:    `let a; export { a as 'a' };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `let a; export { a as 'b' };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `let a; export { a as ' b ' };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{" b "}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg(" b "), Line: 1, Column: 22},
				},
			},
			{
				Code:    `let a; export { a as '👍' };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"👍"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("👍"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `export { 'a' } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 10},
				},
			},
			{
				Code:    `export { '' } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{""}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg(""), Line: 1, Column: 10},
				},
			},
			{
				Code:    `export { ' ' } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{" "}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg(" "), Line: 1, Column: 10},
				},
			},
			{
				Code:    `export { b as 'a' } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export { b as '\u0061' } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export * as 'a' from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 13},
				},
			},

			// ---- destructuring ----
			{
				Code:    `export var [a] = [];`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 13},
				},
			},
			{
				Code:    `export let { a } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `export const { b: a } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 19},
				},
			},
			{
				Code:    `export var [{ a }] = [];`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export let { b: { c: a = d } = e } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 22},
				},
			},

			// ---- reports the correct identifier node in the case of a redeclaration. Note: functions cannot be redeclared in a module. ----
			{
				Code:    `var a; export var a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 19},
				},
			},
			{
				Code:    `export var a; var a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},

			// ---- reports the correct identifier node when the same identifier appears elsewhere in the declaration ----
			{
				Code:    `export var a = a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export let b = a, a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 19},
				},
			},
			{
				Code:    `export const a = 1, b = a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `export var [a] = a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 13},
				},
			},
			{
				Code:    `export let { a: a } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export const { a: b, b: a } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 25},
				},
			},
			{
				Code:    `export var { b: a, a: b } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export let a, { a: b } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export const { a: b } = {}, a = 1;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 29},
				},
			},
			{
				Code:    `export var [a = a] = [];`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 13},
				},
			},
			{
				Code:    `export var { a: a = a } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export let { a } = { a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `export function a(a) {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 17},
				},
			},
			{
				Code:    `export class A { A(){} };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"A"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("A"), Line: 1, Column: 14},
				},
			},
			{
				Code:    `var a; export { a as a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `let a, b; export { a as b, b as a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 33},
				},
			},
			{
				Code:    `const a = 1, b = 2; export { b as a, a as b };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 35},
				},
			},
			{
				Code:    `var a; export { a as b, a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 25},
				},
			},
			{
				Code:    `export { a as a } from 'a';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export { a as b, b as a } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 23},
				},
			},
			{
				Code:    `export { b as a, a as b } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export * as a from 'a';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 13},
				},
			},

			// Note: duplicate identifiers in the same export declaration are a 'duplicate export' syntax error. Example: export var a, a;

			// ---- invalid and valid or multiple invalid in the same declaration ----
			{
				Code:    `export var a, b;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
				},
			},
			{
				Code:    `export let b, a;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export const b = 1, a = 2;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 14},
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 21},
				},
			},
			{
				Code:    `export var a, b, c;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "c"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 12},
					{MessageId: "restrictedNamed", Message: namedMsg("c"), Line: 1, Column: 18},
				},
			},
			{
				Code:    `export let { a, b, c } = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b", "c"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 17},
					{MessageId: "restrictedNamed", Message: namedMsg("c"), Line: 1, Column: 20},
				},
			},
			{
				Code:    `export const [a, b, c, d] = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b", "c"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 18},
					{MessageId: "restrictedNamed", Message: namedMsg("c"), Line: 1, Column: 21},
				},
			},
			{
				Code: `export var { a, x: b, c, d, e: y } = {}, e, f = {};`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{
					"foo", "a", "b", "bar", "d", "e", "baz",
				}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 20},
					{MessageId: "restrictedNamed", Message: namedMsg("d"), Line: 1, Column: 26},
					{MessageId: "restrictedNamed", Message: namedMsg("e"), Line: 1, Column: 42},
				},
			},
			{
				Code:    `var a, b; export { a, b };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 20},
				},
			},
			{
				Code:    `let a, b; export { b, a };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 23},
				},
			},
			{
				Code:    `const a = 1, b = 1; export { a, b };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 30},
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 33},
				},
			},
			{
				Code:    `export { a, b, c }; var a, b, c;`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "c"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 10},
					{MessageId: "restrictedNamed", Message: namedMsg("c"), Line: 1, Column: 16},
				},
			},
			{
				Code:    `export { b as a, b } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
				},
			},
			{
				Code:    `export { b as a, b } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 18},
				},
			},
			{
				Code:    `export { b as a, b } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"a", "b"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 15},
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 18},
				},
			},
			{
				Code: `export { a, b, c, d, x as e, f, g } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{
					"foo", "b", "bar", "d", "e", "f", "baz",
				}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("b"), Line: 1, Column: 13},
					{MessageId: "restrictedNamed", Message: namedMsg("d"), Line: 1, Column: 19},
					{MessageId: "restrictedNamed", Message: namedMsg("e"), Line: 1, Column: 27},
					{MessageId: "restrictedNamed", Message: namedMsg("f"), Line: 1, Column: 30},
				},
			},

			// ---- restrictedNamedExportsPattern ----
			{
				Code:    `var getSomething; export { getSomething };`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "get*"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("getSomething"), Line: 1, Column: 28},
				},
			},
			{
				Code:    `var getSomethingFromUser; export { getSomethingFromUser };`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "User$"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("getSomethingFromUser"), Line: 1, Column: 36},
				},
			},
			{
				Code:    `var foo, ab, xy; export { foo, ab, xy };`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "(b|y)$"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("ab"), Line: 1, Column: 32},
					{MessageId: "restrictedNamed", Message: namedMsg("xy"), Line: 1, Column: 36},
				},
			},
			{
				Code:    `var foo; export { foo as ab };`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "(b|y)$"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("ab"), Line: 1, Column: 26},
				},
			},
			{
				Code:    `var privateUserEmail; export { privateUserEmail };`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "^privateUser"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("privateUserEmail"), Line: 1, Column: 32},
				},
			},
			{
				Code:    `export const a = 1;`,
				Options: []any{map[string]any{"restrictedNamedExportsPattern": "^(?!foo)(?!bar).+$"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("a"), Line: 1, Column: 14},
				},
			},

			// ---- reports "default" in named export declarations (when configured) ----
			{
				Code:    `var a; export { a as default };`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"default"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("default"), Line: 1, Column: 22},
				},
			},
			{
				Code:    `export { default } from 'foo';`,
				Options: []any{map[string]any{"restrictedNamedExports": []any{"default"}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedNamed", Message: namedMsg("default"), Line: 1, Column: 10},
				},
			},

			// ---- restrictDefaultExports.direct option ----
			{
				Code:    `export default foo;`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code:    `export default 42;`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code:    `export default function foo() {}`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"direct": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},
			{
				Code: `export default foo;`,
				Options: []any{map[string]any{
					"restrictedNamedExports": []any{"bar"},
					"restrictDefaultExports": map[string]any{"direct": true},
				}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 1},
				},
			},

			// ---- restrictDefaultExports.named option ----
			{
				Code:    "const foo = 123;\nexport { foo as default };",
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"named": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 2, Column: 17},
				},
			},

			// ---- restrictDefaultExports.defaultFrom option ----
			{
				Code:    `export { default } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 10},
				},
			},
			{
				Code:    `export { default as default } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 21},
				},
			},
			{
				Code:    `export { 'default' } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"defaultFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 10},
				},
			},

			// ---- restrictDefaultExports.namedFrom option ----
			{
				Code:    `export { foo as default } from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namedFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 17},
				},
			},

			// ---- restrictDefaultExports.namespaceFrom option ----
			{
				Code:    `export * as default from 'mod';`,
				Options: []any{map[string]any{"restrictDefaultExports": map[string]any{"namespaceFrom": true}}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "restrictedDefault", Message: defaultMsg, Line: 1, Column: 13},
				},
			},
		},
	)
}
